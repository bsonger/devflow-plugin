package model

type (
	Internet    string
	ReleaseType string
)

const (
	Internal Internet = "internal"
	External Internet = "external"

	Normal    ReleaseType = "normal"
	Canary    ReleaseType = "canary"
	BlueGreen ReleaseType = "blue-green"
)

type Release struct {
	Type     ReleaseType `yaml:"type"`
	Replica  *int32      `json:"replica"`
	Internet Internet    `json:"internet"`
	App      `yaml:"app"`
	Image    `yaml:"image"`
	Configs  []Config `yaml:"configs"`
	Service  []Port   `yaml:"service"`
	Env      string   `yaml:"env"`
}

type App struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
}

type Config struct {
	Name      string            `yaml:"name"`
	MountPath string            `yaml:"mountPath"`
	FilesPath map[string]string `yaml:"filesPath"`
}

type Port struct {
	Name       string `yaml:"name"`
	Port       int    `yaml:"port"`
	TargetPort int    `yaml:"targetPort"`
}
