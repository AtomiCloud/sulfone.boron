package docker_executor

import (
	"errors"
	"fmt"
	"strings"
)

func StripDash(id string) string {
	return strings.ReplaceAll(id, "-", "")
}

func InsertDash(uuid string) string {
	if len(uuid) != 32 {
		return uuid
	}

	return uuid[:8] + "-" + uuid[8:12] + "-" + uuid[12:16] + "-" + uuid[16:20] + "-" + uuid[20:]
}

type DockerImageReference struct {
	Reference string
	Tag       string
}

func DockerImageToString(image DockerImageReference) string {
	return fmt.Sprintf("%s:%s", image.Reference, image.Tag)
}

func DockerImageToStruct(imageString string) (DockerImageReference, error) {
	parts := strings.Split(imageString, ":")
	if len(parts) != 2 {
		fmt.Println("Invalid image string format")
		return DockerImageReference{}, errors.New("invalid image string format")
	}
	return DockerImageReference{
		Reference: parts[0],
		Tag:       parts[1],
	}, nil
}

type DockerContainerReference struct {
	CyanId    string
	CyanType  string
	SessionId string
}

func DockerContainerToString(container DockerContainerReference) string {

	templateVersionId := StripDash(container.CyanId)

	if container.SessionId == "" {
		return "cyan-" + container.CyanType + "-" + templateVersionId
	}
	return "cyan-" + container.CyanType + "-" + templateVersionId + "-" + container.SessionId
}

func DockerContainerNameToStruct(name string) (DockerContainerReference, error) {
	if strings.HasPrefix(name, "cyan-") {
		parts := strings.Split(name, "-")
		sessionId := ""
		if len(parts) == 4 {
			sessionId = parts[3]
		}
		return DockerContainerReference{
			CyanType:  parts[1],
			CyanId:    InsertDash(parts[2]),
			SessionId: sessionId,
		}, nil
	}
	return DockerContainerReference{}, errors.New("invalid container name: " + name)
}

type DockerVolumeReference struct {
	CyanId    string `json:"cyan_id"`
	SessionId string `json:"session_id"`
}

func DockerVolumeToString(volume DockerVolumeReference) string {
	templateVersionId := StripDash(volume.CyanId)
	if volume.SessionId == "" {
		return "cyan-" + templateVersionId
	}
	return "cyan-" + templateVersionId + "-" + volume.SessionId
}

func DockerVolumeNameToStruct(realName string) (DockerVolumeReference, error) {
	if strings.HasPrefix(realName, "cyan-") {
		parts := strings.Split(realName, "-")
		sessionId := ""
		if len(parts) == 3 {
			sessionId = parts[2]
		}
		return DockerVolumeReference{
			CyanId:    InsertDash(parts[1]),
			SessionId: sessionId,
		}, nil
	}
	return DockerVolumeReference{}, errors.New("invalid volume realName")
}
