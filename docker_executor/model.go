package docker_executor

type CyanGlobReq struct {
	Glob    string   `json:"glob"`
	Exclude []string `json:"exclude"`
	Type    string   `json:"type"`
}

type CyanPluginReq struct {
	Name   string      `json:"name"`
	Config interface{} `json:"config"`
}

type CyanPlugin struct {
	Id        string      `json:"id"`
	Reference string      `json:"reference"`
	Username  string      `json:"username"`
	Name      string      `json:"name"`
	Version   int         `json:"version"`
	Config    interface{} `json:"config"`
}

type CyanProcessorReq struct {
	Name   string        `json:"name"`
	Config interface{}   `json:"config"`
	Files  []CyanGlobReq `json:"files"`
}

type CyanProcessor struct {
	Id        string        `json:"id"`
	Reference string        `json:"reference"`
	Username  string        `json:"username"`
	Name      string        `json:"name"`
	Version   int           `json:"version"`
	Config    interface{}   `json:"config"`
	Files     []CyanGlobReq `json:"files"`
}

type CyanReq struct {
	Processors []CyanProcessorReq `json:"processors"`
	Plugins    []CyanPluginReq    `json:"plugins"`
}

type MergeReq struct {
	FromDirs []string
	ToDir    string
	Template TemplateVersionRes `json:"template"`
}

type ZipReq struct {
	TargetDir string `json:"target_dir"`
}

type StandardResponse struct {
	Status string `json:"status"`
}

type BuildReq struct {
	Template TemplateVersionRes `json:"template"`
	Cyan     CyanReq            `json:"cyan"`
	MergerId string             `json:"merger_id"`
}

// IsoProcessorRes
/**
 * Isolated Processor Response, per processor
 */
type IsoProcessorRes struct {
	OutputDir string `json:"outputDir"`
}

// IsoPluginRes
/**
 * Isolated Plugin Response, per plugin
 */
type IsoPluginRes struct {
	OutputDir string `json:"outputDir"`
}

// IsoProcessorReq
/**
 * Isolated Processor Request, per processor
 */
type IsoProcessorReq struct {
	ReadDir  string        `json:"readDir"`
	WriteDir string        `json:"writeDir"`
	Globs    []CyanGlobReq `json:"globs"`
	Config   interface{}   `json:"config"`
}

// IsoPluginReq
/**
 * Isolated Plugin Request, per plugin
 */
type IsoPluginReq struct {
	Directory string      `json:"directory"`
	Config    interface{} `json:"config"`
}

// Registry responses

type RegistryPluginVersionPrincipalRes struct {
	Id              string `json:"id"`
	Version         int    `json:"version"`
	CreatedAt       string `json:"created_at"`
	Description     string `json:"description"`
	DockerReference string `json:"dockerReference"`
	DockerTag       string `json:"dockerTag"`
}

type RegistryPluginRes struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Project     string   `json:"project"`
	Source      string   `json:"source"`
	Email       string   `json:"email"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	Readme      string   `json:"readme"`
	UserId      string   `json:"user_id"`
}

type RegistryPluginVersionRes struct {
	Principal RegistryPluginVersionPrincipalRes `json:"principal"`
	Plugin    RegistryPluginRes                 `json:"plugin"`
}

type RegistryProcessorVersionPrincipalRes struct {
	Id              string `json:"id"`
	Version         int    `json:"version"`
	CreatedAt       string `json:"created_at"`
	Description     string `json:"description"`
	DockerReference string `json:"dockerReference"`
	DockerTag       string `json:"dockerTag"`
}

type RegistryProcessorRes struct {
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Project     string   `json:"project"`
	Source      string   `json:"source"`
	Email       string   `json:"email"`
	Tags        []string `json:"tags"`
	Description string   `json:"description"`
	Readme      string   `json:"readme"`
	UserId      string   `json:"user_id"`
}

type RegistryProcessorVersionRes struct {
	Principal RegistryProcessorVersionPrincipalRes `json:"principal"`
	Processor RegistryProcessorRes                 `json:"processor"`
}
