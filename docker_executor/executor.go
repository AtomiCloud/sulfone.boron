package docker_executor

import (
	"fmt"
	"net/http"
	"time"
)

type Executor struct {
	Docker   DockerClient
	Template TemplateVersionRes
}

type MergerReq struct {
	MergerId string `json:"merger_id"`
}

func (e Executor) missingPluginsImages(images []DockerImageReference) []DockerImageReference {
	var missing []DockerImageReference
	for _, plugin := range e.Template.Plugins {
		c := DockerImageReference{
			Reference: plugin.DockerReference,
			Sha:       plugin.DockerSHA,
		}
		found := false
		for _, i := range images {
			if i.Reference == plugin.DockerReference && i.Sha == plugin.DockerSHA {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, c)
		}
	}
	return missing
}

func (e Executor) missingProcessorImages(images []DockerImageReference) []DockerImageReference {
	var missing []DockerImageReference
	for _, processor := range e.Template.Processors {
		c := DockerImageReference{
			Reference: processor.DockerReference,
			Sha:       processor.DockerSHA,
		}
		found := false
		for _, i := range images {
			if i.Reference == processor.DockerReference && i.Sha == processor.DockerSHA {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, c)
		}
	}
	return missing
}

func (e Executor) startMerger(session string, temVolRef, workVolRef DockerVolumeReference, mr MergerReq) error {

	i, er := e.Docker.GetCoordinatorImage()
	if er != nil {
		fmt.Println("🚨 Error getting coordinator image", er)
		return er
	}

	c := DockerContainerReference{
		CyanId:    mr.MergerId,
		CyanType:  "merger",
		SessionId: session,
	}
	fmt.Println("🚀 Starting merger [inner]", c.CyanId)
	err := e.Docker.CreateContainerWithReadWriteVolume(c, temVolRef, workVolRef, i)
	if err != nil {
		fmt.Println("🚨 Error starting merger [inner]", c.CyanId, err)
		return err
	} else {
		fmt.Println("✅ Successfully started merger [inner]", c.CyanId)
	}
	fmt.Println("🕠 Waiting for merger [inner]", c.CyanId, "to be ready...")
	ep := "http://" + DockerContainerToString(c) + ":9000"
	err = e.statusCheck(ep, 60)
	if err != nil {
		fmt.Println("🚨 Error waiting for merger [inner]", c.CyanId, err)
	} else {
		fmt.Println("✅ Merger [inner]", c.CyanId, "is ready")
	}
	return nil
}

func (e Executor) startProcessors(session string, temVolRef, workVolRef DockerVolumeReference) []error {

	processors := e.Template.Processors

	errChan := make(chan error, len(processors))
	semaphore := make(chan int, e.Docker.ParallelismLimit)

	for _, processor := range processors {
		semaphore <- 0
		i := DockerImageReference{
			Reference: processor.DockerReference,
			Sha:       processor.DockerSHA,
		}
		c := DockerContainerReference{
			CyanId:    processor.ID,
			CyanType:  "processor",
			SessionId: session,
		}
		go func(container DockerContainerReference, image DockerImageReference) {
			fmt.Println("🚀 Starting processor", container.CyanId)
			err := e.Docker.CreateContainerWithReadWriteVolume(container, temVolRef, workVolRef, image)
			if err != nil {
				fmt.Println("🚨 Error starting processor [inner]", container.CyanId, err)
				errChan <- err
				<-semaphore
				return
			} else {
				fmt.Println("✅ Successfully started processor [inner]", container.CyanId)
			}
			fmt.Println("🕠 Waiting for processor [inner]", container.CyanId, "to be ready...")
			ep := "http://" + DockerContainerToString(container) + ":5551"
			err = e.statusCheck(ep, 60)
			if err != nil {
				fmt.Println("🚨 Error waiting for processor [inner]", container.CyanId, err)
			} else {
				fmt.Println("✅ Processor [inner]", container.CyanId, "is ready")
			}
			errChan <- err
			<-semaphore
		}(c, i)
	}

	var allErr []error

	for i := 0; i < len(processors); i++ {
		err := <-errChan
		if err != nil {
			allErr = append(allErr, err)
		}
	}

	for i := 0; i < cap(semaphore); i++ {
		semaphore <- 0
	}

	// close channels
	close(errChan)
	return allErr
}

func (e Executor) startPlugins(session string, temVolRef, workVolRef DockerVolumeReference) []error {
	plugins := e.Template.Plugins

	errChan := make(chan error, len(plugins))
	semaphore := make(chan int, e.Docker.ParallelismLimit)

	for _, plugin := range plugins {
		semaphore <- 0
		i := DockerImageReference{
			Reference: plugin.DockerReference,
			Sha:       plugin.DockerSHA,
		}
		c := DockerContainerReference{
			CyanId:    plugin.ID,
			CyanType:  "plugin",
			SessionId: session,
		}
		go func(container DockerContainerReference, image DockerImageReference) {
			fmt.Println("🚀 Starting plugin [inner]", container.CyanId)
			err := e.Docker.CreateContainerWithReadWriteVolume(container, temVolRef, workVolRef, image)
			if err != nil {
				fmt.Println("🚨 Error starting plugin [inner]", container.CyanId, err)
				errChan <- err
				<-semaphore
				return
			} else {
				fmt.Println("✅ Successfully started plugin [inner]", container.CyanId)
			}
			fmt.Println("🕠 Waiting for plugin [inner]", container.CyanId, "to be ready...")
			ep := "http://" + DockerContainerToString(container) + ":5552"
			err = e.statusCheck(ep, 60)
			if err != nil {
				fmt.Println("🚨 Error waiting for plugin [inner]", container.CyanId, err)
			} else {
				fmt.Println("✅ Plugin [inner]", container.CyanId, "is ready")
			}
			errChan <- err
			<-semaphore
		}(c, i)
	}

	var allErr []error

	for i := 0; i < len(plugins); i++ {
		err := <-errChan
		if err != nil {
			allErr = append(allErr, err)
		}
	}

	for i := 0; i < cap(semaphore); i++ {
		semaphore <- 0
	}

	// close channels
	close(errChan)
	return allErr
}

func (e Executor) listContainersVolumes() (
	[]DockerContainerReference,
	[]DockerVolumeReference,
	error) {
	containerChan := make(chan []DockerContainerReference)
	volumeChan := make(chan []DockerVolumeReference)
	errChan := make(chan error)
	d := e.Docker

	go func() {
		fmt.Println("🔍 Looking for containers to clean...")
		runningContainers, stoppedContainers, err := d.ListContainer()
		if err != nil {
			fmt.Println("🚨 Error looking for containers", err)
			containerChan <- nil
			errChan <- err
		} else {
			containerRefs := append(runningContainers, stoppedContainers...)
			fmt.Println("✅ Successfully retrieved containers, found", len(containerRefs), "containers")
			containerChan <- containerRefs
			errChan <- nil
		}
	}()

	go func() {
		fmt.Println("🔍 Looking for volumes to clean...")
		volumes, err := d.ListVolumes()
		if err != nil {
			fmt.Println("🚨 Error looking for volumes", err)
			errChan <- err
			volumeChan <- nil
		} else {
			fmt.Println("✅ Successfully retrieved volumes, found", len(volumes), "volumes")
			volumeChan <- volumes
			errChan <- nil
		}
	}()

	containerRefs := <-containerChan
	volumeRefs := <-volumeChan

	e1 := <-errChan
	e2 := <-errChan

	if e1 != nil {
		return nil, nil, e1
	}
	if e2 != nil {
		return nil, nil, e2
	}
	return containerRefs, volumeRefs, nil

}

func (e Executor) statusCheck(endpoint string, maxAttempts int) error {

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
			break // Escape from the loop if status code is 200
		} else {
			fmt.Println("🚨 Request failed! Status code:", resp.StatusCode)
		}
		if i == maxAttempts {
			fmt.Println("🚨 Reached maximum attempts of", maxAttempts)
			return fmt.Errorf("reached maximum attempts of %d", maxAttempts)
		} else {
			fmt.Println("⌛ Waiting for 1 second before next attempt...")
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}

func (e Executor) Clean(session string) []error {
	containers, volumes, err := e.listContainersVolumes()
	if err != nil {
		return []error{err}
	}

	var sessionContainer []DockerContainerReference
	var sessionVolume []DockerVolumeReference

	for _, container := range containers {
		if container.SessionId == session {
			sessionContainer = append(sessionContainer, container)
		}
	}
	for _, volume := range volumes {
		if volume.SessionId == session {
			sessionVolume = append(sessionVolume, volume)
		}
	}

	errs := e.Docker.RemoveAllContainers(sessionContainer)
	if len(errs) > 0 {
		return errs
	}
	errs = e.Docker.RemoveAllVolumes(sessionVolume)
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (e Executor) Start(session string, readVolRef, writeVolRef DockerVolumeReference, req MergerReq) []error {

	errChan := make(chan []error)

	// start processors
	go func() {
		fmt.Println("🚀 Starting processors...")
		errs := e.startProcessors(session, readVolRef, writeVolRef)
		if len(errs) > 0 {
			fmt.Println("🚨 Error starting processors", errs)
			errChan <- errs
		} else {
			fmt.Println("✅ Successfully started processors")
			errChan <- nil
		}
	}()

	// start plugins
	go func() {
		fmt.Println("🚀 Starting plugins...")
		errs := e.startPlugins(session, readVolRef, writeVolRef)
		if len(errs) > 0 {
			fmt.Println("🚨 Error starting plugins", errs)
			errChan <- errs
		} else {
			fmt.Println("✅ Successfully started plugins")
			errChan <- nil
		}
	}()

	// start merger
	go func() {
		fmt.Println("🚀 Starting merger...")
		err := e.startMerger(session, readVolRef, writeVolRef, req)
		if err != nil {
			fmt.Println("🚨 Error starting merger", err)
			errChan <- []error{err}
		} else {
			fmt.Println("✅ Successfully started merger")
			errChan <- nil
		}
	}()

	e1 := <-errChan
	e2 := <-errChan
	e3 := <-errChan

	allErrs := append(e1, e2...)
	allErrs = append(allErrs, e3...)

	if len(allErrs) > 0 {
		return allErrs
	}
	return nil

}

func (e Executor) Warm(session string) (string, DockerVolumeReference, []error) {
	fmt.Println("🔑 Starting a new session:", session)
	fmt.Println("🔍 Looking for images...")
	images, err := e.Docker.ListImages()
	if err != nil {
		fmt.Println("🚨 Error looking for images", err)
		return session, DockerVolumeReference{}, []error{err}
	} else {
		fmt.Println("✅ Successfully retrieved images, found", len(images), "images")
	}

	missingPluginImages := e.missingPluginsImages(images)
	missingProcessorImages := e.missingProcessorImages(images)
	missingImages := append(missingPluginImages, missingProcessorImages...)

	errChan := make(chan []error)
	volRefChan := make(chan DockerVolumeReference)

	// pull missing image
	go func() {
		if len(missingImages) > 0 {
			fmt.Println("📥 Pulling missing images...")
			errs := e.Docker.PullImages(missingImages)
			if len(errs) > 0 {
				fmt.Println("🚨 Error pulling images", errs)
				errChan <- errs
			} else {
				fmt.Println("✅ Successfully pulled images")
				errChan <- nil
			}
		} else {
			fmt.Println("✅ No missing images")
			errChan <- nil
		}
	}()

	// start session volume
	go func() {
		fmt.Println("📦 Creating session volume...")
		v := DockerVolumeReference{
			CyanId:    e.Template.Principal.ID,
			SessionId: session,
		}
		err = e.Docker.CreateVolume(v)
		if err != nil {
			fmt.Println("🚨 Error creating session volume", err)
			errChan <- []error{err}
		} else {
			fmt.Println("✅ Successfully created session volume")
			errChan <- nil
		}
		volRefChan <- v
	}()

	e1 := <-errChan
	e2 := <-errChan
	volRef := <-volRefChan

	if len(e1) > 0 {
		return session, volRef, e1
	}
	if len(e2) > 0 {
		return session, volRef, e2
	}

	return session, volRef, nil
}
