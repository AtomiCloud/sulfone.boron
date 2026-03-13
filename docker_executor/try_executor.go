package docker_executor

import (
	"fmt"
	"net/http"
	"time"
)

// httpClient with timeout for health checks
var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

type TryExecutor struct {
	Docker  DockerClient
	Request TryExecutorReq
}

// TrySetup performs the full try setup flow
func (e *TryExecutor) TrySetup() (TryExecutorRes, []error) {
	// 0. Validate source type
	source := e.Request.Source
	if source == "" {
		source = "image"
	}
	if source != "image" && source != "path" {
		return TryExecutorRes{}, []error{fmt.Errorf("invalid source type: %s (must be 'image' or 'path')", source)}
	}

	// 1. Create blob volume (idempotent)
	blobVol, errs := e.createBlobVolume()
	if len(errs) > 0 {
		return TryExecutorRes{}, errs
	}

	// 2. Pull missing images
	errs = e.pullMissingImages()
	if len(errs) > 0 {
		return TryExecutorRes{}, errs
	}

	// 3. Warm resolvers
	errs = e.warmResolvers()
	if len(errs) > 0 {
		return TryExecutorRes{}, errs
	}

	return TryExecutorRes{
		BlobVolume: blobVol,
	}, nil
}

func (e *TryExecutor) createBlobVolume() (DockerVolumeReference, []error) {
	blobVol := DockerVolumeReference{
		CyanId:    e.Request.LocalTemplateId,
		SessionId: "",
	}

	// Check if exists
	volumes, err := e.Docker.ListVolumes()
	if err != nil {
		return blobVol, []error{err}
	}

	for _, v := range volumes {
		if v.CyanId == e.Request.LocalTemplateId && v.SessionId == "" {
			fmt.Println("✅ Blob volume already exists:", DockerVolumeToString(blobVol))
			return blobVol, nil
		}
	}

	// Create volume
	fmt.Println("📦 Creating blob volume:", DockerVolumeToString(blobVol))
	if err := e.Docker.CreateVolume(blobVol); err != nil {
		return blobVol, []error{err}
	}

	// Extract/populate blob based on source type
	source := e.Request.Source
	if source == "" {
		source = "image"
	}

	var populateErr error
	if source == "path" {
		populateErr = e.populateBlobFromPath(blobVol)
	} else {
		populateErr = e.populateBlobFromImage(blobVol)
	}

	// CLEANUP: Remove blob volume if population failed
	if populateErr != nil {
		_ = e.Docker.RemoveVolume(blobVol) // best effort cleanup
		return blobVol, []error{populateErr}
	}

	return blobVol, nil
}

func (e *TryExecutor) populateBlobFromPath(blobVol DockerVolumeReference) error {
	cc := DockerContainerReference{
		CyanId:    e.Request.LocalTemplateId,
		CyanType:  "copy-helper",
		SessionId: "",
	}
	fmt.Println("📂 Copying files from path:", e.Request.Path)
	return e.Docker.CreateContainerWithCopyMount(cc, e.Request.Path, blobVol)
}

func (e *TryExecutor) populateBlobFromImage(blobVol DockerVolumeReference) error {
	props := e.Request.Template.Principal.Properties
	if props == nil {
		return fmt.Errorf("template properties are required for blob extraction")
	}

	blobImage := DockerImageReference{
		Reference: props.BlobDockerReference,
		Tag:       props.BlobDockerTag,
	}

	cc := DockerContainerReference{
		CyanId:    e.Request.LocalTemplateId,
		CyanType:  "unzip",
		SessionId: "",
	}

	fmt.Println("📦 Extracting blob from image:", DockerImageToString(blobImage))

	// Create and start the unzip container with the blob image
	if err := e.Docker.CreateContainerWithVolume(cc, blobVol, blobImage); err != nil {
		return fmt.Errorf("failed to start unzip container: %w", err)
	}

	// Ensure container is removed on all exit paths
	defer func() {
		fmt.Println("🧹 Removing unzip container")
		_ = e.Docker.RemoveContainer(cc) // best effort cleanup
	}()

	// Wait for the blob extraction (tar) to complete and exit
	fmt.Println("⚙️ Waiting for blob extraction to complete...")
	exitCode, err := e.Docker.WaitContainer(cc)
	if err != nil {
		return fmt.Errorf("failed waiting for blob extraction: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("blob extraction failed with exit code %d", exitCode)
	}
	fmt.Println("✅ Blob extraction completed")

	return nil
}

func (e *TryExecutor) pullMissingImages() []error {
	images, err := e.Docker.ListImages()
	if err != nil {
		return []error{err}
	}

	var missing []DockerImageReference

	// Check processors
	for _, p := range e.Request.Template.Processors {
		ref := DockerImageReference{Reference: p.DockerReference, Tag: p.DockerTag}
		if !imageExistsInList(images, ref) {
			missing = append(missing, ref)
		}
	}

	// Check plugins
	for _, p := range e.Request.Template.Plugins {
		ref := DockerImageReference{Reference: p.DockerReference, Tag: p.DockerTag}
		if !imageExistsInList(images, ref) {
			missing = append(missing, ref)
		}
	}

	// Check resolvers
	for _, r := range e.Request.Template.Resolvers {
		ref := DockerImageReference{Reference: r.DockerReference, Tag: r.DockerTag}
		if !imageExistsInList(images, ref) {
			missing = append(missing, ref)
		}
	}

	if len(missing) > 0 {
		fmt.Println("📥 Pulling missing images...")
		return e.Docker.PullImages(missing)
	}

	return nil
}

func imageExistsInList(images []DockerImageReference, ref DockerImageReference) bool {
	for _, img := range images {
		if img.Reference == ref.Reference && img.Tag == ref.Tag {
			return true
		}
	}
	return false
}

func (e *TryExecutor) warmResolvers() []error {
	runningContainers, stoppedContainers, err := e.Docker.ListContainer()
	if err != nil {
		return []error{err}
	}

	images, err := e.Docker.ListImages()
	if err != nil {
		return []error{err}
	}

	// De-duplicate resolvers by ID before warming
	seenResolvers := make(map[string]bool)
	uniqueResolvers := make([]ResolverRes, 0, len(e.Request.Template.Resolvers))
	for _, resolver := range e.Request.Template.Resolvers {
		if !seenResolvers[resolver.ID] {
			seenResolvers[resolver.ID] = true
			uniqueResolvers = append(uniqueResolvers, resolver)
		}
	}

	var allErrs []error
	for _, resolver := range uniqueResolvers {
		// Check if container exists and is running
		missing, conRef := e.missingResolverContainer(resolver, runningContainers, stoppedContainers)

		if !missing {
			// Container already running - verify health before skipping
			fmt.Println("✅ Resolver container already running:", resolver.ID)
			ep := fmt.Sprintf("http://%s:%d/", DockerContainerToString(conRef), ResolverPort)
			if err := e.statusCheck(ep, 10); err != nil {
				allErrs = append(allErrs, fmt.Errorf("resolver %s health check failed: %w", resolver.ID, err))
			}
			continue
		}

		// Check if image exists, pull if missing
		imgMissing, imgRef := e.missingResolverImage(resolver, images)
		if imgMissing {
			fmt.Println("📥 Pulling resolver image:", DockerImageToString(imgRef))
			if errs := e.Docker.PullImages([]DockerImageReference{imgRef}); len(errs) > 0 {
				allErrs = append(allErrs, errs...)
				continue
			}
		}

		// Start container
		fmt.Println("🚀 Starting resolver container:", resolver.ID)
		if err := e.startResolverContainer(resolver, conRef); err != nil {
			allErrs = append(allErrs, err)
			continue
		}

		// Health check
		ep := fmt.Sprintf("http://%s:%d/", DockerContainerToString(conRef), ResolverPort)
		if err := e.statusCheck(ep, 60); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	return allErrs
}

func (e *TryExecutor) missingResolverContainer(resolver ResolverRes, runningContainers, stoppedContainers []DockerContainerReference) (bool, DockerContainerReference) {
	conRef := DockerContainerReference{
		CyanId:    resolver.ID,
		CyanType:  CyanTypeResolver,
		SessionId: "",
	}

	// Check if container is running
	for _, c := range runningContainers {
		if c.CyanType == CyanTypeResolver && c.CyanId == resolver.ID {
			return false, c
		}
	}

	// Check if container exists but is stopped - need to restart
	for _, c := range stoppedContainers {
		if c.CyanType == CyanTypeResolver && c.CyanId == resolver.ID {
			// Remove stopped container so it can be recreated
			fmt.Println("🧹 Removing stopped resolver container:", resolver.ID)
			_ = e.Docker.RemoveContainer(c) // best effort cleanup
			break
		}
	}

	return true, conRef
}

func (e *TryExecutor) missingResolverImage(resolver ResolverRes, images []DockerImageReference) (bool, DockerImageReference) {
	ref := DockerImageReference{Reference: resolver.DockerReference, Tag: resolver.DockerTag}
	for _, img := range images {
		if img.Reference == ref.Reference && img.Tag == ref.Tag {
			return false, ref
		}
	}
	return true, ref
}

func (e *TryExecutor) startResolverContainer(resolver ResolverRes, conRef DockerContainerReference) error {
	imgRef := DockerImageReference{Reference: resolver.DockerReference, Tag: resolver.DockerTag}
	return e.Docker.CreateContainer(conRef, imgRef)
}

func (e *TryExecutor) statusCheck(endpoint string, maxAttempts int) error {
	for i := 0; i < maxAttempts; i++ {
		fmt.Println("🏓 Ping endpoint:", endpoint, "Attempt:", i+1)
		resp, err := httpClient.Get(endpoint)
		if err != nil {
			fmt.Println("🚨 Error:", err)
			fmt.Println("⌛ Waiting for 1 second before next attempt...")
			time.Sleep(1 * time.Second)
			continue
		}
		if resp.StatusCode == http.StatusOK {
			fmt.Println("✅ Health Check successful, Status code:", resp.StatusCode)
			resp.Body.Close()
			return nil
		}
		fmt.Println("🚨 Request failed! Status code:", resp.StatusCode)
		resp.Body.Close()
		fmt.Println("⌛ Waiting for 1 second before next attempt...")
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("health check failed for %s after %d attempts", endpoint, maxAttempts)
}
