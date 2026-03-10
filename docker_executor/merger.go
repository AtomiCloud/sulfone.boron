package docker_executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Merger struct {
	ParallelismLimit int
	RegistryClient   RegistryClient
	Template         TemplateVersionRes
	SessionId        string
}

func copyFile(src, dst string) error {
	// Open the source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(sourceFile *os.File) {
		_ = sourceFile.Close()
	}(sourceFile)

	// Create the destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(destFile *os.File) {
		_ = destFile.Close()
	}(destFile)

	// Copy the contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Copy file permissions
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, info.Mode())
}

func PostJSON[Req any, Res any](url string, requestBody Req) (Res, error) {
	var responseBody Res

	// Marshal the request into JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return responseBody, fmt.Errorf("error marshaling request: %w", err)
	}

	// Perform the HTTP POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return responseBody, fmt.Errorf("error performing POST request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		// Return error with status code and body for caller to format
		return responseBody, fmt.Errorf("%d %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the response body
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		if err == io.EOF { // Empty response is not necessarily an error
			return responseBody, nil
		}
		return responseBody, fmt.Errorf("error decoding response: %w", err)
	}

	return responseBody, nil
}

// processorFile represents a file from a processor output
type processorFile struct {
	Path     string // Full path to the file
	Template string // Processor reference
	Layer    int    // Layer order (processor index)
	RelPath  string // Relative path within the processor output
}

// findMatchingResolver finds resolvers whose file patterns match the given path
func findMatchingResolver(path string, resolvers []ResolverRes) []ResolverRes {
	var matches []ResolverRes
	for _, resolver := range resolvers {
		for _, pattern := range resolver.Files {
			matched, err := doublestar.Match(pattern, path)
			if err == nil && matched {
				matches = append(matches, resolver)
				break // One pattern match per resolver is enough
			}
		}
	}
	return matches
}

// buildResolverFiles reads file contents and builds resolver file requests
func buildResolverFiles(relPath string, versions []processorFile) ([]ResolverFile, error) {
	var files []ResolverFile
	for _, version := range versions {
		content, err := os.ReadFile(version.Path)
		if err != nil {
			return nil, fmt.Errorf("Failed to read file '%s': %w", relPath, err)
		}
		files = append(files, ResolverFile{
			Path:    relPath,
			Content: string(content),
			Origin: ResolverOrigin{
				Template: version.Template,
				Layer:    version.Layer,
			},
		})
	}
	return files, nil
}

// callResolver makes an HTTP POST request to the resolver container
func callResolver(resolverID string, sessionID string, req ResolverRequest) (*ResolverResponse, error) {
	ref := DockerContainerReference{
		CyanId:    resolverID,
		CyanType:  CyanTypeResolver,
		SessionId: sessionID,
	}
	containerName := DockerContainerToString(ref)
	endpoint := fmt.Sprintf("http://%s:%d/api/resolve", containerName, ResolverPort)

	// Use PostJSON generic function
	response, err := PostJSON[ResolverRequest, ResolverResponse](endpoint, req)
	if err != nil {
		// Detect connection failures (container not running)
		if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") || strings.Contains(err.Error(), "dial tcp") {
			return nil, fmt.Errorf("Resolver container for '%s' not found or not running", resolverID)
		}
		// For other errors, format as "Resolver call failed for '{file-path}': {error}"
		// PostJSON returns "{code} {body}" for non-200 status codes
		return nil, fmt.Errorf("Resolver call failed for '%s': %s", req.Files[0].Path, err.Error())
	}

	return &response, nil
}

// getResolverIDs extracts resolver IDs from a slice of ResolverRes
func getResolverIDs(resolvers []ResolverRes) []string {
	var ids []string
	for _, r := range resolvers {
		ids = append(ids, r.ID)
	}
	return ids
}

func parseCyanReference(ref string) (string, string, *string, error) {
	var version *string = nil

	if strings.Contains(ref, ":") {
		split := strings.Split(ref, ":")
		if len(split) != 2 {
			return "", "", nil, fmt.Errorf("invalid reference: %s", ref)
		}
		ref = split[0]
		version = &split[1]
	}
	sects := strings.Split(ref, "/")
	if len(sects) != 2 {
		return "", "", nil, fmt.Errorf("invalid reference: %s", ref)
	}
	return sects[0], sects[1], version, nil
}

func (m Merger) execProcessors(processors []CyanProcessorReq) ([]string, []error) {
	errChan := make(chan error, len(processors))
	// Pre-allocate slice with empty placeholders to preserve processor order
	writeDirs := make([]string, len(processors))
	for i := range writeDirs {
		writeDirs[i] = ""
	}
	semaphore := make(chan int, m.ParallelismLimit)
	for idx, processor := range processors {
		go func(p CyanProcessorReq, processorIndex int) {
			semaphore <- 0
			filePath, err := uuid.NewUUID()
			if err != nil {
				fmt.Printf("Error unique write-path: %v", err)
				errChan <- err
				<-semaphore
				return
			}
			fmt.Println("🔭 Checking Processor references...")
			pp, err := m.RegistryClient.convertProcessor(p, m.Template.Processors)
			if err != nil {
				fmt.Printf("Error converting processor %s: %v", p.Name, err)
				errChan <- err
				<-semaphore
				return
			}
			exist := false
			for _, proc := range m.Template.Processors {
				if proc.ID == pp.Id {
					exist = true
					break
				}
			}
			if !exist {
				err = fmt.Errorf("processor %s does not exist in template %s", pp.Id, m.Template.Principal.ID)
				fmt.Printf("Processor %s does not exist in template %s", pp.Id, m.Template.Principal.ID)
				errChan <- err
				<-semaphore
				return
			}
			fmt.Println("✅ Processor references checked.")
			container := DockerContainerReference{
				CyanId:    pp.Id,
				CyanType:  "processor",
				SessionId: m.SessionId,
			}
			endpoint := fmt.Sprintf("http://%s:5551/api/process", DockerContainerToString(container))
			fmt.Println("🚀 Starting processor with ", endpoint)
			res, err := PostJSON[IsoProcessorReq, IsoProcessorRes](endpoint, IsoProcessorReq{
				ReadDir:  "/workspace/cyanprint",
				WriteDir: "/workspace/area/" + filePath.String(),
				Globs:    pp.Files,
				Config:   pp.Config,
			})
			if err != nil {
				fmt.Printf("Error starting processor %s: %v", pp.Id, err)
				errChan <- err
				<-semaphore
				return
			}
			fmt.Println("🎉 Processor", pp.Id, "completed")
			// Store result at the processor's original index to preserve order
			writeDirs[processorIndex] = res.OutputDir
			errChan <- nil
			<-semaphore
		}(processor, idx)
	}

	var errs []error
	for i := 0; i < len(processors); i++ {
		err := <-errChan
		if err != nil {
			errs = append(errs, err)
		}
	}
	return writeDirs, errs
}

func (m Merger) execPlugins(mergePath string, plugins []CyanPluginReq) []error {

	// async conversion
	errChan := make(chan error, len(plugins))
	pluginChan := make(chan CyanPlugin, len(plugins))

	for _, plugin := range plugins {
		go func(c CyanPluginReq) {
			fmt.Println("🔭 Checking Plugin references...")
			p, err := m.RegistryClient.convertPlugin(c, m.Template.Plugins)
			if err != nil {
				fmt.Printf("🚨 Error converting plugin %s: %v\n", c.Name, err)
				errChan <- err
				pluginChan <- CyanPlugin{}

				return
			}
			fmt.Println("✅ Plugin references checked.")
			pluginChan <- p
			errChan <- nil
		}(plugin)
	}

	var errs []error
	var convertedPlugins []CyanPlugin
	for i := 0; i < len(plugins); i++ {
		p := <-pluginChan
		convertedPlugins = append(convertedPlugins, p)
		err := <-errChan
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}

	for _, plugin := range convertedPlugins {

		container := DockerContainerReference{
			CyanId:    plugin.Id,
			CyanType:  "plugin",
			SessionId: m.SessionId,
		}
		endpoint := fmt.Sprintf("http://%s:5552/api/plug", DockerContainerToString(container))
		fmt.Println("🚀 Process plugin on ", endpoint)
		_, err := PostJSON[IsoPluginReq, IsoPluginRes](endpoint, IsoPluginReq{
			Directory: mergePath,
			Config:    plugin.Config,
		})
		if err != nil {
			fmt.Printf("🚨 Error processing plugin %s: %v\n", plugin.Id, err)
			return []error{err}
		}
		fmt.Println("🎉 Plugin", plugin.Id, "completed")
	}
	return nil
}

func (m Merger) merge(dirs []string, mergePath string, mergerId string) error {
	req := MergeReq{
		FromDirs: dirs,
		ToDir:    mergePath,
		Template: m.Template,
	}

	c := DockerContainerReference{
		CyanId:    mergerId,
		CyanType:  "merger",
		SessionId: m.SessionId,
	}

	ep := DockerContainerToString(c)
	fullEp := "http://" + ep + ":9000/merge/" + m.SessionId
	fmt.Println("🚀 Starting merger with ", fullEp)
	_, err := PostJSON[MergeReq, StandardResponse](fullEp, req)
	if err != nil {
		fmt.Printf("🚨 Error starting merger: %v\n", err)
		return err
	}
	fmt.Println("🎉 Merger completed")
	return nil
}

// MergeFiles used by merger container
// Detects conflicts and calls resolvers to intelligently merge conflicting files
func (m Merger) MergeFiles(fromDirs []string, mergeDir string) error {
	// Step 1: Collect all files from all processor outputs
	fileMap := make(map[string][]processorFile) // path -> list of versions

	for layer, dir := range fromDirs {
		err := filepath.Walk(dir, func(fullPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get the relative path by subtracting the "dir" from "fullPath"
			relPath, err := filepath.Rel(dir, fullPath)
			if err != nil {
				return err
			}

			// Skip directories - we'll handle them when copying files
			if info.IsDir() {
				return nil
			}

			// Add file to the map
			fileMap[relPath] = append(fileMap[relPath], processorFile{
				Path:     fullPath,
				Template: m.Template.Processors[layer].ID,
				Layer:    layer,
				RelPath:  relPath,
			})

			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk directory '%s': %w", dir, err)
		}
	}

	// Step 2: Identify conflicts and non-conflicts
	var conflicts []string
	var nonConflicts []string
	for path, versions := range fileMap {
		if len(versions) > 1 {
			conflicts = append(conflicts, path)
		} else {
			nonConflicts = append(nonConflicts, path)
		}
	}

	// Step 3: Handle each conflict
	for _, conflictPath := range conflicts {
		versions := fileMap[conflictPath]

		// Match resolvers using doublestar.Match()
		matchingResolvers := findMatchingResolver(conflictPath, m.Template.Resolvers)

		if len(matchingResolvers) == 0 {
			// LWW: use last version
			fmt.Printf("Conflict detected for '%s': no resolver match, using last writer wins (layer %d)\n", conflictPath, versions[len(versions)-1].Layer)
			lastVersion := versions[len(versions)-1]
			destPath := filepath.Join(mergeDir, conflictPath)
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for '%s': %w", conflictPath, err)
			}
			if err := copyFile(lastVersion.Path, destPath); err != nil {
				return fmt.Errorf("failed to copy file '%s': %w", conflictPath, err)
			}
		} else if len(matchingResolvers) == 1 {
			// Call resolver with all versions
			resolver := matchingResolvers[0]
			files, err := buildResolverFiles(conflictPath, versions)
			if err != nil {
				return err
			}

			request := ResolverRequest{
				Config: resolver.Config,
				Files:  files,
			}

			fmt.Printf("Calling resolver '%s' for conflict '%s' with %d versions\n", resolver.ID, conflictPath, len(files))
			response, err := callResolver(resolver.ID, m.SessionId, request)
			if err != nil {
				return err
			}

			// Verify resolver returned the expected path
			if response.Path != conflictPath {
				return fmt.Errorf("Resolver returned invalid path: expected '%s', got '%s'", conflictPath, response.Path)
			}

			// Write resolved content to merge directory
			destPath := filepath.Join(mergeDir, response.Path)
			if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory for '%s': %w", response.Path, err)
			}
			if err := os.WriteFile(destPath, []byte(response.Content), 0644); err != nil {
				return fmt.Errorf("failed to write resolved file '%s': %w", response.Path, err)
			}
			fmt.Printf("Successfully resolved conflict '%s' using resolver '%s'\n", conflictPath, resolver.ID)
		} else {
			// Multiple resolvers match - ERROR
			return fmt.Errorf("Multiple resolvers match conflicting file '%s': [%s]. Template resolver configuration may be misconfigured.", conflictPath, strings.Join(getResolverIDs(matchingResolvers), ", "))
		}
	}

	// Step 4: Copy non-conflicts
	for _, path := range nonConflicts {
		version := fileMap[path][0]
		destPath := filepath.Join(mergeDir, path)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for '%s': %w", path, err)
		}
		if err := copyFile(version.Path, destPath); err != nil {
			return fmt.Errorf("failed to copy file '%s': %w", path, err)
		}
	}

	return nil
}

// Merge used by coordinator container
func (m Merger) Merge(req BuildReq) (string, []error) {

	// exec all processors
	fmt.Println("⚙️ Executing processors...")
	dirs, errs := m.execProcessors(req.Cyan.Processors)
	if len(errs) > 0 {
		fmt.Println("🚨 Error executing processors: ", errs)
		return "", errs
	}
	fmt.Println("🎉 Processors completed.")

	// merge all processor outputs
	fmt.Println("🔀 Merging processor outputs...")
	mergeDir, err := uuid.NewUUID()
	if err != nil {
		return "", []error{err}
	}
	mergePath := "/workspace/area/" + mergeDir.String()
	err = m.merge(dirs, mergePath, req.MergerId)
	if err != nil {
		fmt.Println("🚨 Error merging processor outputs: ", err)
		return "", []error{err}
	}
	fmt.Println("🎉 Processor outputs merged.")

	// exec all plugins
	fmt.Println("⚙️ Executing plugins...")
	errs = m.execPlugins(mergePath, req.Cyan.Plugins)
	if len(errs) > 0 {
		fmt.Println("🚨 Error executing plugins: ", errs)
		return "", errs
	}
	fmt.Println("🎉 Plugins completed.")
	return mergePath, nil
}
