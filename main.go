package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"gopkg.in/yaml.v1"
)

type Dependency struct {
	Name       string `yaml:"name"`
	Repository string `yaml:"repository"`
	Version    string `yaml:"version"`
}

type Requirements struct {
	Dependencies []Dependency `yaml:"dependencies"`
}

type Entry struct {
	ApiVersion  string   `yaml:"apiVersion,omitempty"`
	AppVersion  string   `yaml:"appVersion,omitempty"`
	Created     string   `yaml:"created,omitempty"`
	Description string   `yaml:"description,omitempty"`
	Digest      string   `yaml:"digest,omitempty"`
	Name        string   `yaml:"name,omitempty"`
	Urls        []string `yaml:"urls,omitempty"`
	Version     string   `yaml:"version,omitempty"`
}

type Index struct {
	ApiVersion string             `yaml:"apiVersion"`
	Entries    map[string][]Entry `yaml:"entries"`
	Generated  string             `yaml:"generated"`
	ServerInfo string             `yaml:"serverInfo"`
}

func fetchIndex(repo string) (*Index, error) {
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

func fetchRequirements() (*Requirements, error) {
	fmt.Println("Reading requirements.yaml")

	requirements := Requirements{}
	data, err := ioutil.ReadFile("requirements.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), &requirements)
	if err != nil {
		return nil, err
	}

	return &requirements, nil
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

var indexes map[string]Index = map[string]Index{}

func fetchVersion(dependency Dependency) error {

	if _, ok := indexes[dependency.Repository]; !ok {
		index, err := fetchIndex(dependency.Repository)
		if err != nil {
			return err
		}
		indexes[dependency.Repository] = *index
	}

	version, err := resolveSemver(dependency.Version, indexes[dependency.Repository].Entries[dependency.Name])

	if err != nil {
		return err
	}

	chart := fmt.Sprintf("%s/charts/%s-%s.tgz", strings.TrimSuffix(dependency.Repository, "/"), dependency.Name, version.Original())
	fmt.Printf("\tFetching chart: %s\n", chart)
	resp, err := http.Get(chart)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	fmt.Printf("Chart is %d bytes in size\n", len(body))

	return nil
}

func main() {

	fmt.Println("Fetching requirements.yaml dependencies")
	dependencies, err := fetchRequirements()
	if err != nil {
		fmt.Printf("Error: %v+", err)
		os.Exit(1)
	}

	fmt.Println("Read requirements.yaml dependencies")

	for _, dependency := range dependencies.Dependencies {
		fmt.Printf("Fetching %s @ %s\n", dependency.Name, dependency.Version)
		err := fetchVersion(dependency)
		if err != nil {
			fmt.Printf("Error: %v+", err)
			os.Exit(1)
		}
	}

	fmt.Println("Done")
}
