package main

type Index struct {
	ApiVersion string             `yaml:"apiVersion"`
	Entries    map[string][]Entry `yaml:"entries"`
	Generated  string             `yaml:"generated"`
}

type Entry struct {
	Annotations Annotation   `yaml:"annotations"`
	ApiVersion  string       `yaml:"apiVersion"`
	Created     string       `yaml:"created"`
	Description string       `yaml:"description"`
	Digest      string       `yaml:"digest"`
	Icon        string       `yaml:"icon"`
	KubeVersion string       `yaml:"kubeVersion"`
	Maintainers []Maintainer `yaml:"maintainers"`
	Name        string       `yaml:"name"`
	Urls        []string     `yaml:"urls"`
	Version     string       `yaml:"version"`
}

type Annotation struct {
	Purpose string `yaml:"purpose"`
}

type Maintainer struct {
	Email string `yaml:"email"`
	Name  string `yaml:"name"`
	Url   string `yaml:"url"`
}
