package docker_executor

import (
	"context"
	"fmt"
	imageTypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types"
	container "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
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

func (d *DockerClient) WaitContainer(ref DockerContainerReference) error {

	name := DockerContainerToString(ref)

	statusCh, errCh := d.Docker.ContainerWait(d.Context, name, container.WaitConditionNotRunning)
	select {
	case e := <-errCh:
		if e != nil {
			return e
		}
	case <-statusCh:
	}

	return nil

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

func (d *DockerClient) RemoveAllContainers(containerRefs []DockerContainerReference) []error {

	errChan := make(chan error, len(containerRefs))
	semaphore := make(chan int, d.ParallelismLimit)

	for _, containerRef := range containerRefs {
		semaphore <- 0
		go func(cc DockerContainerReference) {
			fmt.Println("🗑 Removing container:", DockerContainerToString(cc))
			err := d.RemoveContainer(cc)
			if err != nil {
				fmt.Println("🚨Failed to remove container:", DockerContainerToString(cc))
			} else {
				fmt.Println("✅ Container removed:", DockerContainerToString(cc))
			}
			errChan <- err
			<-semaphore
		}(containerRef)
	}

	var allErr []error

	for i := 0; i < len(containerRefs); i++ {
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

func (d *DockerClient) RemoveVolume(vol DockerVolumeReference) error {
	name := DockerVolumeToString(vol)
	err := d.Docker.VolumeRemove(d.Context, name, true)
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerClient) RemoveAllVolumes(volRefs []DockerVolumeReference) []error {

	errChan := make(chan error, len(volRefs))
	semaphore := make(chan int, d.ParallelismLimit)

	for _, volRef := range volRefs {
		semaphore <- 0
		go func(v DockerVolumeReference) {
			fmt.Println("🗑 Removing volume:", DockerVolumeToString(v))
			err := d.RemoveVolume(v)
			if err != nil {
				fmt.Println("🚨Failed to remove volume:", DockerVolumeToString(v))
			} else {
				fmt.Println("✅ Volume removed:", DockerVolumeToString(v))
			}
			errChan <- err
			<-semaphore
		}(volRef)
	}

	var allErr []error

	for i := 0; i < len(volRefs); i++ {
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
	networks, err := d.Docker.NetworkList(d.Context, types.NetworkListOptions{})
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
	_, err := d.Docker.NetworkCreate(d.Context, networkName, types.NetworkCreate{
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
