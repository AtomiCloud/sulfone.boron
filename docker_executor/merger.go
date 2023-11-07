package docker_executor

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	// Decode the response body
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		if err == io.EOF { // Empty response is not necessarily an error
			return responseBody, nil
		}
		return responseBody, fmt.Errorf("error decoding response: %w", err)
	}

	return responseBody, nil
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
	writeDirChan := make(chan string, len(processors))
	semaphore := make(chan int, m.ParallelismLimit)
	for _, processor := range processors {
		go func(p CyanProcessorReq) {
			semaphore <- 0
			filePath, err := uuid.NewUUID()
			if err != nil {
				fmt.Printf("Error unique write-path: %v", err)
				errChan <- err
				writeDirChan <- ""
				return
			}
			fmt.Println("🔭 Checking Processor references...")
			pp, err := m.RegistryClient.convertProcessor(p)
			if err != nil {
				fmt.Printf("Error converting processor %s: %v", p.Name, err)
				errChan <- err
				writeDirChan <- ""
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
				writeDirChan <- ""
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
				writeDirChan <- ""
				return
			}
			fmt.Println("🎉 Processor", pp.Id, "completed")
			errChan <- nil
			writeDirChan <- res.OutputDir
			<-semaphore
		}(processor)
	}

	var errs []error
	var writeDirs []string
	for i := 0; i < len(processors); i++ {
		writeDor := <-writeDirChan
		writeDirs = append(writeDirs, writeDor)

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
			p, err := m.RegistryClient.convertPlugin(c)
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
func (m Merger) MergeFiles(fromDirs []string, mergeDir string) error {
	for _, dir := range fromDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Get the relative path by subtracting the "dir" from "path"
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}

			// Create a corresponding path in the mergeDir
			destPath := filepath.Join(mergeDir, relPath)

			if info.IsDir() {
				// Create the directory in mergeDir
				return os.MkdirAll(destPath, info.Mode())
			} else {
				// If it's a file, copy the file, overwriting if necessary
				return copyFile(path, destPath)
			}
		})
		if err != nil {
			return err
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
