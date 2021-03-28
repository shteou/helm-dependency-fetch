package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v1"
)

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

type Context struct {
	fs afero.Fs
}

func (c *Context) fetchIndex(repo string) (*Index, error) {
	index := Index{}

	fmt.Printf("Fetching index from %s\n", repo)

	resp, err := http.Get(fmt.Sprintf("%s/index.yaml", strings.TrimSuffix(repo, "/")))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	err = yaml.Unmarshal([]byte(body), &index)
	if err != nil {
		return nil, err
	}

	return &index, nil
}

func (c *Context) fetchChart() (*Chart, error) {
	data, err := ioutil.ReadFile("Chart.yaml")
	if err != nil {
		return nil, err
	}

	chart := Chart{}
	err = yaml.Unmarshal([]byte(data), &chart)

	return &chart, err
}

func (c *Context) fetchRequirements() (*[]Dependency, error) {
	requirements := Requirements{}
	data, err := ioutil.ReadFile("requirements.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), &requirements)
	return &requirements.Dependencies, err
}

func (c *Context) parseDependencies() (*[]Dependency, error) {
	chart, err := c.fetchChart()
	if err != nil {
		return nil, err
	}

	if chart.APIVersion == "v1" {
		return c.fetchRequirements()
	}
	return &chart.Dependencies, nil
}

func largestSemver(versions []*semver.Version) *semver.Version {
	if len(versions) == 0 {
		return nil
	}
	sort.Sort(semver.Collection(versions))
	return versions[len(versions)-1]
}

func resolveSemver(version string, entries []Entry) (*semver.Version, error) {
	c, err := semver.NewConstraint(version)
	if err != nil {
		return nil, err
	}

	versions := []*semver.Version{}

	for _, entry := range entries {
		v, err := semver.NewVersion(entry.Version)
		if err != nil {
			return nil, err
		}

		a, _ := c.Validate(v)
		if a == true {
			versions = append(versions, v)
		}
	}

	largest := largestSemver(versions)
	if largest == nil {
		return nil, errors.New("Couldn't find a semver to satisfy the constraint")
	}

	return largest, nil
}

func (c *Context) fetchURLChart(url string, name string, version string) error {
	fmt.Printf("\tFetching chart: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(fmt.Sprintf("charts/%s-%s.tgz", name, version), body, 0644)
	return err
}

func (c *Context) fetchFileChart(path string) error {
	repoPath := strings.TrimPrefix(path, "file://")

	err := exec.Command("helm", "package", repoPath, "-d", "charts/").Run()

	return err
}

var indexes map[string]Index = map[string]Index{}

func (c *Context) getIndex(repo string) (*Index, error) {
	var index Index

	if _, ok := indexes[repo]; !ok {
		retrievedIndex, err := c.fetchIndex(repo)
		if err != nil {
			return nil, err
		}
		indexes[repo] = *retrievedIndex
	}

	index = indexes[repo]
	return &index, nil
}

func (c *Context) fetchVersion(dependency Dependency) error {
	if !strings.HasPrefix(dependency.Repository, "file://") {
		index, err := c.getIndex(dependency.Repository)
		if err != nil {
			return err
		}

		version, err := resolveSemver(dependency.Version, index.Entries[dependency.Name])
		if err != nil {
			return err
		}
		chart := fmt.Sprintf("%s/charts/%s-%s.tgz", strings.TrimSuffix(dependency.Repository, "/"), dependency.Name, version.Original())
		return c.fetchURLChart(chart, dependency.Name, version.String())
	}

	return c.fetchFileChart(dependency.Repository)
}

func (c *Context) createChartsDirectory() error {
	_, err := c.fs.Stat("charts")
	if os.IsNotExist(err) {
		err = c.fs.Mkdir("charts", 0755)
		return err
	}
	return err
}

func main() {
	c := Context{fs: afero.NewOsFs()}

	dependencies, err := c.parseDependencies()
	if err != nil {
		fmt.Printf("Error: %v+", err)
		os.Exit(1)
	}

	err = c.createChartsDirectory()
	if err != nil {
		log.Fatalf("Failed to manage charts directory %v+", err)
	}

	for _, dependency := range *dependencies {
		fmt.Printf("Fetching %s @ %s\n", dependency.Name, dependency.Version)
		err := c.fetchVersion(dependency)
		if err != nil {
			log.Fatalf("Error: %v+", err)
		}
	}
}
