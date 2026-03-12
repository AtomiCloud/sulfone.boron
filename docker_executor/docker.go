package docker_executor

import (
	"context"
	"fmt"
	container "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	imageTypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	networkTypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"io"
	"os"
	"strings"
)

type DockerClient struct {
	Docker           *client.Client
	Context          context.Context
	ParallelismLimit int
}

const networkName = "cyanprint"

func (d *DockerClient) WaitContainer(ref DockerContainerReference) (int, error) {

	name := DockerContainerToString(ref)

	statusCh, errCh := d.Docker.ContainerWait(d.Context, name, container.WaitConditionNotRunning)
	select {
	case e := <-errCh:
		if e != nil {
			return -1, e
		}
	case status := <-statusCh:
		return int(status.StatusCode), nil
	}

	return 0, nil

}

func (d *DockerClient) ListImages() ([]DockerImageReference, error) {

	f := filters.NewArgs()
	f.Add("label", "cyanprint.dev=true")
	images, err := d.Docker.ImageList(d.Context, imageTypes.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return nil, err
	}
	var imageNames []DockerImageReference

	for _, image := range images {
		for _, tags := range image.RepoTags {
			s, err := DockerImageToStruct(tags)
			if err != nil {
				return nil, err
			}
			imageNames = append(imageNames, s)
		}
	}

	return imageNames, nil
}

func (d *DockerClient) PullImages(images []DockerImageReference) []error {

	errChan := make(chan error, len(images))
	semaphore := make(chan int, d.ParallelismLimit)

	for _, image := range images {
		semaphore <- 0
		go func(image DockerImageReference) {
			ref := DockerImageToString(image)
			fmt.Println("📥 Pulling image:", ref)
			reader, err := d.Docker.ImagePull(d.Context, ref, imageTypes.PullOptions{
				All: true,
			})
			if err != nil {
				fmt.Println("🚨Failed to pull image [Docker.ImagePull]:", ref)
			}
			defer func(reader io.ReadCloser) {
				_ = reader.Close()
			}(reader)
			_, err = io.Copy(os.Stdout, reader)
			if err != nil {
				fmt.Println("🚨Failed to pull image [io.Copy]:", ref)
			} else {
				fmt.Println("✅ Image pulled:", ref)

			}
			errChan <- err
			<-semaphore
		}(image)
	}

	var allErr []error

	for i := 0; i < len(images); i++ {
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

func (d *DockerClient) GetCoordinatorImage() (DockerImageReference, error) {
	f := filters.NewArgs()
	f.Add("label", "cyanprint.name=sulfone-boron")
	images, err := d.Docker.ImageList(d.Context, imageTypes.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return DockerImageReference{}, err
	}
	var latest imageTypes.Summary

	for _, image := range images {
		// Skip images without repo tags (dangling/untagged images)
		if len(image.RepoTags) == 0 {
			continue
		}
		if latest.Created < image.Created {
			latest = image
		}
	}

	for _, tag := range latest.RepoTags {
		s, e := DockerImageToStruct(tag)
		if e != nil {
			return DockerImageReference{}, e
		}
		return s, nil
	}
	return DockerImageReference{}, fmt.Errorf("no coordinator image found")
}

func (d *DockerClient) ListContainer() ([]DockerContainerReference, []DockerContainerReference, error) {

	f := filters.NewArgs()
	f.Add("label", "cyanprint.dev=true")
	containers, err := d.Docker.ContainerList(d.Context, container.ListOptions{
		All:     true,
		Filters: f,
	})
	if err != nil {
		return nil, nil, err
	}
	var cyanRunning []DockerContainerReference
	var cyanStopped []DockerContainerReference
	for _, con := range containers {
		for _, name := range con.Names {
			n := strings.TrimPrefix(name, "/")
			containerRef, err := DockerContainerNameToStruct(n)
			if err != nil {
				return nil, nil, err
			}
			if con.State == "running" {
				cyanRunning = append(cyanRunning, containerRef)
			} else {
				cyanStopped = append(cyanStopped, containerRef)
			}
		}
	}
	return cyanRunning, cyanStopped, nil
}

func (d *DockerClient) CreateContainer(cc DockerContainerReference, image DockerImageReference) error {

	name := DockerContainerToString(cc)
	imageName := DockerImageToString(image)
	c, err := d.Docker.ContainerCreate(d.Context, &container.Config{
		Image: imageName,
		Labels: map[string]string{
			"cyanprint.dev": "true",
		},
	}, &container.HostConfig{
		NetworkMode: networkName,
	}, nil, nil, name)
	if err != nil {
		return err
	}
	err = d.Docker.ContainerStart(d.Context, c.ID, container.StartOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerClient) CreateContainerWithVolume(cc DockerContainerReference, v DockerVolumeReference, image DockerImageReference) error {

	name := DockerContainerToString(cc)
	imageName := DockerImageToString(image)
	volName := DockerVolumeToString(v)

	c, err := d.Docker.ContainerCreate(d.Context, &container.Config{
		Image: imageName,
		Labels: map[string]string{
			"cyanprint.dev": "true",
		},
	}, &container.HostConfig{
		NetworkMode: networkName,
		Mounts: []mount.Mount{
			{
				Type:     "volume",
				Source:   volName,
				Target:   "/workspace/cyanprint",
				ReadOnly: false,
			},
		},
	}, nil, nil, name)
	if err != nil {
		return err
	}
	err = d.Docker.ContainerStart(d.Context, c.ID, container.StartOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerClient) CreateContainerWithCopyMount(
	cc DockerContainerReference,
	sourcePath string,
	targetVolume DockerVolumeReference,
) error {
	name := DockerContainerToString(cc)
	targetVolName := DockerVolumeToString(targetVolume)

	// Get self-image (coordinator)
	image, err := d.GetCoordinatorImage()
	if err != nil {
		return fmt.Errorf("failed to get coordinator image: %w", err)
	}
	imageName := DockerImageToString(image)

	// Create container with bind mount for source and volume mount for target
	c, err := d.Docker.ContainerCreate(d.Context, &container.Config{
		Image: imageName,
		Cmd:   []string{"cp", "-r", "/source/.", "/target/"},
		Labels: map[string]string{
			"cyanprint.dev": "true",
		},
	}, &container.HostConfig{
		NetworkMode: networkName,
		Mounts: []mount.Mount{
			{
				Type:     "bind",
				Source:   sourcePath,
				Target:   "/source",
				ReadOnly: true,
			},
			{
				Type:   "volume",
				Source: targetVolName,
				Target: "/target",
			},
		},
	}, nil, nil, name)
	if err != nil {
		return err
	}

	// Ensure container is removed on all exit paths
	defer func() {
		_ = d.Docker.ContainerRemove(d.Context, c.ID, container.RemoveOptions{
			Force: true,
		})
	}()

	// Start container
	err = d.Docker.ContainerStart(d.Context, c.ID, container.StartOptions{})
	if err != nil {
		return err
	}

	// Wait for completion and check exit status
	statusCh, errCh := d.Docker.ContainerWait(d.Context, c.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return fmt.Errorf("copy container failed with exit code %d", status.StatusCode)
		}
	}

	return nil
}

func (d *DockerClient) RemoveContainer(cc DockerContainerReference) error {
	name := DockerContainerToString(cc)
	err := d.Docker.ContainerRemove(d.Context, name, container.RemoveOptions{
		Force:         true,
		RemoveVolumes: false,
	})
	if err != nil {
		return err
	}
	return nil
}

type indexedError struct {
	index int
	err   error
}

func (d *DockerClient) RemoveAllContainers(containerRefs []DockerContainerReference) []error {

	errChan := make(chan indexedError, len(containerRefs))
	semaphore := make(chan int, d.ParallelismLimit)

	for i, containerRef := range containerRefs {
		semaphore <- 0
		go func(idx int, cc DockerContainerReference) {
			fmt.Println("🗑 Removing container:", DockerContainerToString(cc))
			err := d.RemoveContainer(cc)
			if err != nil {
				fmt.Println("🚨 Failed to remove container:", DockerContainerToString(cc))
			} else {
				fmt.Println("✅ Container removed:", DockerContainerToString(cc))
			}
			errChan <- indexedError{index: idx, err: err}
			<-semaphore
		}(i, containerRef)
	}

	allErr := make([]error, len(containerRefs))

	for i := 0; i < len(containerRefs); i++ {
		ie := <-errChan
		allErr[ie.index] = ie.err
	}

	for i := 0; i < cap(semaphore); i++ {
		semaphore <- 0
	}

	// close channels
	close(errChan)
	return allErr
}

func (d *DockerClient) RemoveVolume(vol DockerVolumeReference) error {
	name := DockerVolumeToString(vol)
	err := d.Docker.VolumeRemove(d.Context, name, true)
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerClient) RemoveAllVolumes(volRefs []DockerVolumeReference) []error {

	errChan := make(chan indexedError, len(volRefs))
	semaphore := make(chan int, d.ParallelismLimit)

	for i, volRef := range volRefs {
		semaphore <- 0
		go func(idx int, v DockerVolumeReference) {
			fmt.Println("🗑 Removing volume:", DockerVolumeToString(v))
			err := d.RemoveVolume(v)
			if err != nil {
				fmt.Println("🚨 Failed to remove volume:", DockerVolumeToString(v))
			} else {
				fmt.Println("✅ Volume removed:", DockerVolumeToString(v))
			}
			errChan <- indexedError{index: idx, err: err}
			<-semaphore
		}(i, volRef)
	}

	allErr := make([]error, len(volRefs))

	for i := 0; i < len(volRefs); i++ {
		ie := <-errChan
		allErr[ie.index] = ie.err
	}

	for i := 0; i < cap(semaphore); i++ {
		semaphore <- 0
	}

	// close channels
	close(errChan)
	return allErr
}

func (d *DockerClient) CreateContainerWithReadWriteVolume(cc DockerContainerReference, readVolume, writeVolume DockerVolumeReference, image DockerImageReference) error {

	name := DockerContainerToString(cc)
	imageName := DockerImageToString(image)

	readVolName := DockerVolumeToString(readVolume)
	writeVolName := DockerVolumeToString(writeVolume)

	c, err := d.Docker.ContainerCreate(d.Context, &container.Config{
		Image: imageName,
		Labels: map[string]string{
			"cyanprint.dev": "true",
		},
	}, &container.HostConfig{
		NetworkMode: networkName,
		Mounts: []mount.Mount{
			{
				Type:     "volume",
				Source:   readVolName,
				Target:   "/workspace/cyanprint",
				ReadOnly: true,
			},
			{
				Type:     "volume",
				Source:   writeVolName,
				Target:   "/workspace/area",
				ReadOnly: false,
			},
		},
	}, nil, nil, name)
	if err != nil {
		return err
	}
	err = d.Docker.ContainerStart(d.Context, c.ID, container.StartOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerClient) CyanPrintNetworkExist() (bool, error) {
	networks, err := d.Docker.NetworkList(d.Context, networkTypes.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, network := range networks {
		if network.Name == networkName {
			return true, nil
		}
	}
	return false, nil
}

func (d *DockerClient) CreateNetwork() error {
	_, err := d.Docker.NetworkCreate(d.Context, networkName, networkTypes.CreateOptions{
		Driver: "bridge",
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerClient) EnforceNetwork() error {

	fmt.Println("🔍 Check if network exists...")
	exist, err := d.CyanPrintNetworkExist()
	fmt.Println("🌀 Network exists: ", exist)
	if err != nil {
		return err
	}
	if !exist {
		fmt.Println("🌐 Creating network...")
		err = d.CreateNetwork()
		if err != nil {
			fmt.Println("🚨 Failed to create network")
			return err
		}
		fmt.Println("✅ Network created")
	}
	return nil
}

func (d *DockerClient) ListVolumes() ([]DockerVolumeReference, error) {

	f := filters.NewArgs()
	f.Add("label", "cyanprint.dev=true")
	volumes, err := d.Docker.VolumeList(d.Context, volume.ListOptions{
		Filters: f,
	})
	if err != nil {
		return nil, err
	}
	var volumeNames []DockerVolumeReference
	for _, vol := range volumes.Volumes {
		v, err := DockerVolumeNameToStruct(vol.Name)
		if err != nil {
			return nil, err
		}
		volumeNames = append(volumeNames, v)
	}
	return volumeNames, nil
}

func (d *DockerClient) CreateVolume(vol DockerVolumeReference) error {
	volName := DockerVolumeToString(vol)

	_, err := d.Docker.VolumeCreate(d.Context, volume.CreateOptions{
		Labels: map[string]string{
			"cyanprint.dev": "true",
		},
		Name: volName,
	})
	return err
}

// RemoveImage removes a Docker image by its reference
func (d *DockerClient) RemoveImage(imageRef DockerImageReference) error {
	imageName := DockerImageToString(imageRef)
	_, err := d.Docker.ImageRemove(d.Context, imageName, imageTypes.RemoveOptions{
		Force:         true,
		PruneChildren: false,
	})
	return err
}

// RemoveAllImages removes multiple Docker images in parallel
func (d *DockerClient) RemoveAllImages(imageRefs []DockerImageReference) []error {
	errChan := make(chan indexedError, len(imageRefs))
	semaphore := make(chan int, d.ParallelismLimit)

	for i, imageRef := range imageRefs {
		semaphore <- 0
		go func(idx int, img DockerImageReference) {
			fmt.Println("🗑 Removing image:", DockerImageToString(img))
			err := d.RemoveImage(img)
			if err != nil {
				fmt.Println("🚨 Failed to remove image:", DockerImageToString(img))
			} else {
				fmt.Println("✅ Image removed:", DockerImageToString(img))
			}
			errChan <- indexedError{index: idx, err: err}
			<-semaphore
		}(i, imageRef)
	}

	// Initialize slice with nil values to preserve order
	allErr := make([]error, len(imageRefs))

	for i := 0; i < len(imageRefs); i++ {
		ie := <-errChan
		allErr[ie.index] = ie.err
	}

	for i := 0; i < cap(semaphore); i++ {
		semaphore <- 0
	}

	close(errChan)
	return allErr
}

// Cleanup removes all Docker resources (containers, images, volumes) labeled cyanprint.dev=true
// It returns lists of successfully removed resources and the first error encountered (if any)
func (d *DockerClient) Cleanup() (containersRemoved []string, imagesRemoved []string, volumesRemoved []string, err error) {
	var firstError error

	// 1. Remove containers
	runningContainers, stoppedContainers, listErr := d.ListContainer()
	if listErr != nil {
		return nil, nil, nil, fmt.Errorf("failed to list containers: %w", listErr)
	}
	allContainers := append(runningContainers, stoppedContainers...)
	if len(allContainers) > 0 {
		containerErrors := d.RemoveAllContainers(allContainers)
		for i, c := range allContainers {
			if containerErrors[i] != nil {
				if firstError == nil {
					firstError = fmt.Errorf("failed to remove container %s: %w", DockerContainerToString(c), containerErrors[i])
				}
			} else {
				containersRemoved = append(containersRemoved, DockerContainerToString(c))
			}
		}
	}

	// 2. Remove images
	images, listErr := d.ListImages()
	if listErr != nil {
		return containersRemoved, nil, nil, fmt.Errorf("failed to list images: %w", listErr)
	}
	if len(images) > 0 {
		imageErrors := d.RemoveAllImages(images)
		for i, img := range images {
			if imageErrors[i] != nil {
				if firstError == nil {
					firstError = fmt.Errorf("failed to remove image %s: %w", DockerImageToString(img), imageErrors[i])
				}
			} else {
				imagesRemoved = append(imagesRemoved, DockerImageToString(img))
			}
		}
	}

	// 3. Remove volumes
	volumes, listErr := d.ListVolumes()
	if listErr != nil {
		return containersRemoved, imagesRemoved, nil, fmt.Errorf("failed to list volumes: %w", listErr)
	}
	if len(volumes) > 0 {
		volumeErrors := d.RemoveAllVolumes(volumes)
		for i, v := range volumes {
			if volumeErrors[i] != nil {
				if firstError == nil {
					firstError = fmt.Errorf("failed to remove volume %s: %w", DockerVolumeToString(v), volumeErrors[i])
				}
			} else {
				volumesRemoved = append(volumesRemoved, DockerVolumeToString(v))
			}
		}
	}

	return containersRemoved, imagesRemoved, volumesRemoved, firstError
}
