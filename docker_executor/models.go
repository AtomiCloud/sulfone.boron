package docker_executor

type TemplateVersionRes struct {
	Principal  TemplateVersionPrincipalRes   `json:"principal"`
	Template   TemplatePrincipalRes          `json:"template"`
	Plugins    []PluginRes                   `json:"plugins"`
	Processors []ProcessorRes                `json:"processors"`
	Templates  []TemplateVersionPrincipalRes `json:"templates"`
	Resolvers  []ResolverRes                 `json:"resolvers"`
}

type PropertyRes struct {
	BlobDockerReference     string `json:"blobDockerReference"`
	BlobDockerTag           string `json:"blobDockerTag"`
	TemplateDockerReference string `json:"templateDockerReference"`
	TemplateDockerTag       string `json:"templateDockerTag"`
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

type ResolverRes struct {
	ID              string      `json:"id"`
	Version         int64       `json:"version"`
	CreatedAt       string      `json:"createdAt"`
	Description     string      `json:"description"`
	DockerReference string      `json:"dockerReference"`
	DockerTag       string      `json:"dockerTag"`
	Config          interface{} `json:"config"`
	Files           []string    `json:"files"`
}

type TemplateVersionPrincipalRes struct {
	ID          string       `json:"id"`
	Version     int64        `json:"version"`
	CreatedAt   string       `json:"createdAt"`
	Description string       `json:"description"`
	Properties  *PropertyRes `json:"properties"`
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

// TryExecutorReq is the request body for POST /executor/try
type TryExecutorReq struct {
	SessionId       string                `json:"session_id" binding:"required"`
	LocalTemplateId string                `json:"local_template_id" binding:"required"`
	Source          string                `json:"source"` // "image" (default) or "path"
	ImageRef        *DockerImageReference `json:"image_ref"`
	Path            string                `json:"path"`
	Template        TemplateVersionRes    `json:"template" binding:"required"`
	MergerId        string                `json:"merger_id" binding:"required"`
}

// TryExecutorRes is the response for POST /executor/try
type TryExecutorRes struct {
	SessionId     string                `json:"session_id"`
	BlobVolume    DockerVolumeReference `json:"blob_volume"`
	SessionVolume DockerVolumeReference `json:"session_volume"`
}
