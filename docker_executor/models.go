package docker_executor

type TemplateVersionRes struct {
	Principal  TemplateVersionPrincipalRes `json:"principal"`
	Template   TemplatePrincipalRes        `json:"template"`
	Plugins    []PluginRes                 `json:"plugins"`
	Processors []ProcessorRes              `json:"processors"`
}

type PluginRes struct {
	ID              string `json:"id"`
	Version         int64  `json:"version"`
	CreatedAt       string `json:"createdAt"`
	Description     string `json:"description"`
	DockerReference string `json:"dockerReference"`
	DockerTag       string `json:"dockerTag"`
}

type ProcessorRes struct {
	ID              string `json:"id"`
	Version         int64  `json:"version"`
	CreatedAt       string `json:"createdAt"`
	Description     string `json:"description"`
	DockerReference string `json:"dockerReference"`
	DockerTag       string `json:"dockerTag"`
}

type TemplateVersionPrincipalRes struct {
	ID                      string `json:"id"`
	Version                 int64  `json:"version"`
	CreatedAt               string `json:"createdAt"`
	Description             string `json:"description"`
	BlobDockerReference     string `json:"blobDockerReference"`
	BlobDockerTag           string `json:"blobDockerTag"`
	TemplateDockerReference string `json:"templateDockerReference"`
	TemplateDockerTag       string `json:"templateDockerTag"`
}

type TemplatePrincipalRes struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Project     string   `json:"project"`
	Source      string   `json:"source"`
	Email       string   `json:"email"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	Readme      string   `json:"readme"`
	UserID      string   `json:"userId"`
}
