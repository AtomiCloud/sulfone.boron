package docker_executor

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

type TemplateExecutor struct {
	Docker    DockerClient
	Template  TemplateVersionPrincipalRes
	Resolvers []ResolverRes
}

func (de TemplateExecutor) missingTemplateContainer(containers []DockerContainerReference) (bool, DockerContainerReference) {
	template := de.Template
	c := DockerContainerReference{
		CyanType:  "template",
		CyanId:    template.ID,
		SessionId: "",
	}
	// if exists, return false (not missing)
	for _, container := range containers {
		if container.CyanType == "template" && container.CyanId == template.ID {
			return false, c
		}
	}
	// if not exists, return true (missing), and the container that needs to be created
	return true, c
}

func (de TemplateExecutor) missingTemplateVolume(volumes []DockerVolumeReference) (bool, DockerVolumeReference) {
	template := de.Template
	v := DockerVolumeReference{
		CyanId:    template.ID,
		SessionId: "",
	}
	for _, volume := range volumes {
		if volume.CyanId == template.ID {
			return false, v
		}
	}
	return true, v
}

func (de TemplateExecutor) missingTemplateVolumeImage(images []DockerImageReference) (bool, DockerImageReference) {
	template := de.Template
	i := DockerImageReference{
		Reference: template.Properties.BlobDockerReference,
		Tag:       template.Properties.BlobDockerTag,
	}
	for _, image := range images {
		if strings.HasSuffix(i.Reference, image.Reference) && image.Tag == i.Tag {
			return false, i
		}
	}
	return true, i
}

func (de TemplateExecutor) missingTemplateImages(images []DockerImageReference) (bool, DockerImageReference) {
	template := de.Template
	i := DockerImageReference{
		Reference: template.Properties.TemplateDockerReference,
		Tag:       template.Properties.TemplateDockerTag,
	}
	for _, image := range images {
		if strings.HasSuffix(i.Reference, image.Reference) && image.Tag == i.Tag {
			return false, i
		}
	}
	return true, i
}

func (de TemplateExecutor) missingResolverContainer(resolver ResolverRes, containers []DockerContainerReference) (bool, DockerContainerReference) {
	c := DockerContainerReference{
		CyanType:  CyanTypeResolver,
		CyanId:    resolver.ID,
		SessionId: "",
	}
	for _, container := range containers {
		if container.CyanType == CyanTypeResolver && container.CyanId == resolver.ID {
			return false, c
		}
	}
	return true, c
}

func (de TemplateExecutor) missingResolverImages(resolver ResolverRes, images []DockerImageReference) (bool, DockerImageReference) {
	i := DockerImageReference{
		Reference: resolver.DockerReference,
		Tag:       resolver.DockerTag,
	}
	for _, image := range images {
		if strings.HasSuffix(i.Reference, image.Reference) && image.Tag == i.Tag {
			return false, i
		}
	}
	return true, i
}

func (de TemplateExecutor) listContainersVolumesImages() (
	[]DockerContainerReference,
	[]DockerImageReference,
	[]DockerVolumeReference,
	[]error) {
	containerChan := make(chan []DockerContainerReference)
	imageChan := make(chan []DockerImageReference)
	volumeChan := make(chan []DockerVolumeReference)
	errChan := make(chan []error)
	d := de.Docker

	go func() {
		fmt.Println("🔍 Looking for volumes...")
		volumes, err := d.ListVolumes()
		if err != nil {
			fmt.Println("🚨 Error looking for volumes", err)
			errChan <- []error{err}
		} else {
			fmt.Println("✅ Successfully retrieved volumes, found", len(volumes), "volumes")
			volumeChan <- volumes
		}
	}()

	go func() {
		fmt.Println("🔍 Looking for containers...")
		runningContainers, stoppedContainers, err := d.ListContainer()
		if err != nil {
			fmt.Println("🚨 Error looking for containers", err)
			errChan <- []error{err}
		} else {
			fmt.Println("🗑️ Removing stopped containers...")
			errs := d.RemoveAllContainers(stoppedContainers)
			if len(errs) > 0 {
				fmt.Println("🚨 Error removing containers", err)
				errChan <- errs
			} else {
				fmt.Println("✅ Successfully removed stopped containers, found", len(runningContainers), "running containers")
				containerChan <- runningContainers
			}
		}
	}()

	go func() {
		fmt.Println("🔍 Looking for images...")
		images, err := d.ListImages()
		if err != nil {
			fmt.Println("🚨 Error looking for images", err)
			errChan <- []error{err}
		} else {
			fmt.Println("✅ Successfully retrieved images, found", len(images), "images")
			imageChan <- images
		}
	}()

	select {
	case err := <-errChan:
		return nil, nil, nil, err
	case containers := <-containerChan:
		select {
		case err := <-errChan:
			return nil, nil, nil, err
		case images := <-imageChan:
			select {
			case err := <-errChan:
				return nil, nil, nil, err
			case volumes := <-volumeChan:
				return containers, images, volumes, nil
			}
		}
	}
}

func (de TemplateExecutor) startContainer(conRef DockerContainerReference) error {
	imageRef := DockerImageReference{
		Reference: de.Template.Properties.TemplateDockerReference,
		Tag:       de.Template.Properties.TemplateDockerTag,
	}
	err := de.Docker.CreateContainer(conRef, imageRef)
	if err != nil {

		return err
	}

	return nil
}

func (de TemplateExecutor) startResolverContainer(resolver ResolverRes, conRef DockerContainerReference) error {
	imageRef := DockerImageReference{
		Reference: resolver.DockerReference,
		Tag:       resolver.DockerTag,
	}
	err := de.Docker.CreateContainer(conRef, imageRef)
	if err != nil {
		return err
	}
	return nil
}

func (de TemplateExecutor) startVolume(volRef DockerVolumeReference) error {
	d := de.Docker
	fmt.Println("🚧 Creating volume", volRef.CyanId)
	err := d.CreateVolume(volRef)
	if err != nil {
		fmt.Println("🚨 Failed to create volume", volRef.CyanId)
		return err
	} else {
		fmt.Println("✅ Volume created", volRef.CyanId)
	}
	unzipImage := DockerImageReference{
		Reference: de.Template.Properties.BlobDockerReference,
		Tag:       de.Template.Properties.BlobDockerTag,
	}
	unzipContainer := DockerContainerReference{
		CyanId:    de.Template.ID,
		CyanType:  "volume",
		SessionId: "",
	}

	fmt.Println("🚧 Unzipping volume ", volRef.CyanId)
	err = d.CreateContainerWithVolume(unzipContainer, volRef, unzipImage)
	if err != nil {
		fmt.Println("🚨 Failed to start unzip container", volRef.CyanId)
		return err
	} else {
		fmt.Println("⚙️ Still unzipping...", volRef.CyanId)
	}
	err = d.WaitContainer(unzipContainer)
	if err != nil {
		fmt.Println("🚨 Failed to unzip volume", volRef.CyanId)
		return err
	} else {
		fmt.Println("✅ Volume unzipped", volRef.CyanId)
	}
	fmt.Println("🧹 Removing unzip container", volRef.CyanId)
	err = d.RemoveContainer(unzipContainer)
	if err != nil {
		fmt.Println("🚨 Failed to remove unzip container", volRef.CyanId)
		return err
	}
	fmt.Println("✅ Unzip container removed", volRef.CyanId)
	return nil
}

func (de TemplateExecutor) startContainerAndVolume(conMissing, volMissing bool, con DockerContainerReference, vol DockerVolumeReference) []error {

	errChan := make(chan error)
	var errs []error

	if conMissing {
		go func() {
			fmt.Println("🚀 Starting Template container")
			err := de.startContainer(con)
			if err != nil {
				fmt.Println("🚨 Failed to start Template container", err)
			} else {
				fmt.Println("✅ Template container started")
			}
			errChan <- err
		}()
	}

	if volMissing {
		go func() {
			fmt.Println("📦 Creating Volumes...")
			err := de.startVolume(vol)
			if err != nil {
				fmt.Println("🚨 Failed to create volume", err)
			} else {
				fmt.Println("✅ Volumes created")
			}
			errChan <- err
		}()
	}

	if conMissing {
		err := <-errChan
		if err != nil {
			errs = append(errs, err)
		}
	}
	if volMissing {
		err := <-errChan
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (de TemplateExecutor) statusCheck(endpoint string, maxAttempts int) error {

	for i := 0; i < maxAttempts; i++ {
		// Send GET request
		fmt.Println("🏓 Ping endpoint:", endpoint, "Attempt:", i+1)
		resp, err := http.Get(endpoint)
		if err != nil {
			fmt.Println("🚨 Error:", err)
			fmt.Println("⌛ Waiting for 1 second before next attempt...")
			time.Sleep(1 * time.Second)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			fmt.Println("✅ Health Check successful, Status code:", resp.StatusCode)
			resp.Body.Close()
			return nil // Success - escape from the loop
		} else {
			fmt.Println("🚨 Request failed! Status code:", resp.StatusCode)
			resp.Body.Close()
		}

		fmt.Println("⌛ Waiting for 1 second before next attempt...")
		time.Sleep(1 * time.Second)
	}

	// If we get here, we exhausted all attempts without success
	fmt.Println("🚨 Reached maximum attempts of", maxAttempts)
	return fmt.Errorf("reached maximum attempts of %d without successful health check for %s", maxAttempts, endpoint)
}

func (de TemplateExecutor) WarmTemplate() []error {

	d := de.Docker

	containerRefs, imageRefs, volumeRefs, errs := de.listContainersVolumesImages()

	if len(errs) > 0 {
		return errs
	}

	conMissing, container := de.missingTemplateContainer(containerRefs)
	imageMissing, image := de.missingTemplateImages(imageRefs)
	volumeImageMissing, volumeImage := de.missingTemplateVolumeImage(imageRefs)
	volumeMissing, volume := de.missingTemplateVolume(volumeRefs)

	if conMissing {
		fmt.Println("🚧 Template container is missing")
	} else {
		fmt.Println("✅ Template container exists")
	}
	if imageMissing {
		fmt.Println("🚧 Template image is missing")
	} else {
		fmt.Println("✅ Template image exists")
	}
	if volumeImageMissing {
		fmt.Println("🚧 Template volume image is missing")
	} else {
		fmt.Println("✅ Template volume image exists")
	}
	if volumeMissing {
		fmt.Println("🚧 Template volume is missing")
	} else {
		fmt.Println("✅ Template volume exists")
	}

	var images []DockerImageReference
	if imageMissing {
		images = append(images, image)
	}
	if volumeImageMissing {
		images = append(images, volumeImage)
	}
	errs = d.PullImages(images)
	if len(errs) > 0 {
		return errs
	}
	errs = de.startContainerAndVolume(conMissing, volumeMissing, container, volume)
	if len(errs) > 0 {
		return errs
	}

	realName := DockerContainerToString(container)

	fmt.Println("🔍 Checking if container is running...")

	err := de.statusCheck("http://"+realName+":5550/", 60)
	if err != nil {
		fmt.Println("🚨 Starting template container failed:", err)
		return []error{err}
	}
	fmt.Println("✅ Template container is running")

	// Warm resolvers
	for _, resolver := range de.Resolvers {
		resolverConMissing, resolverCon := de.missingResolverContainer(resolver, containerRefs)
		resolverImageMissing, resolverImage := de.missingResolverImages(resolver, imageRefs)

		if resolverConMissing {
			fmt.Println("🚧 Resolver container is missing:", resolver.ID)
		} else {
			fmt.Println("✅ Resolver container exists:", resolver.ID)
		}
		if resolverImageMissing {
			fmt.Println("🚧 Resolver image is missing:", resolver.ID)
		} else {
			fmt.Println("✅ Resolver image exists:", resolver.ID)
		}

		if resolverImageMissing {
			errs = d.PullImages([]DockerImageReference{resolverImage})
			if len(errs) > 0 {
				return errs
			}
		}

		if resolverConMissing {
			fmt.Println("🚀 Starting Resolver container:", resolver.ID)
			err := de.startResolverContainer(resolver, resolverCon)
			if err != nil {
				fmt.Println("🚨 Failed to start Resolver container:", resolver.ID, err)
				return []error{err}
			}
			fmt.Println("✅ Resolver container started:", resolver.ID)
		}

		resolverRealName := DockerContainerToString(resolverCon)
		fmt.Println("🔍 Checking if resolver container is running:", resolver.ID)
		err := de.statusCheck(fmt.Sprintf("http://%s:%d/", resolverRealName, ResolverPort), 60)
		if err != nil {
			fmt.Println("🚨 Starting resolver container failed:", resolver.ID, err)
			return []error{err}
		}
		fmt.Println("✅ Resolver container is running:", resolver.ID)
	}

	return nil

}
