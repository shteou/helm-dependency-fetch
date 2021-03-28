package helm

type Chart struct {
	APIVersion   string       `yaml:"apiVersion"`
	Dependencies []Dependency `yaml:"dependencies,omitempty"`
}

type Dependency struct {
	Name       string `yaml:"name"`
	Repository string `yaml:"repository"`
	Version    string `yaml:"version"`
}

type Requirements struct {
	Dependencies []Dependency `yaml:"dependencies"`
}

type Entry struct {
	APIVersion  string   `yaml:"apiVersion,omitempty"`
	AppVersion  string   `yaml:"appVersion,omitempty"`
	Created     string   `yaml:"created,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Digest      string   `yaml:"digest,omitempty"`
	Name        string   `yaml:"name,omitempty"`
	Urls        []string `yaml:"urls,omitempty"`
	Version     string   `yaml:"version,omitempty"`
}

type Index struct {
	APIVersion string             `yaml:"apiVersion"`
	Entries    map[string][]Entry `yaml:"entries"`
	Generated  string             `yaml:"generated"`
	ServerInfo string             `yaml:"serverInfo"`
}
