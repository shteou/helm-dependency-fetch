package main

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func fetchUrlChart(url string, name string, version string) {
	fmt.Printf("\tFetching chart: %s\n", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Failed to fetch chart")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Unable to fetch chart")
	}

	err = ioutil.WriteFile(fmt.Sprintf("charts/%s-%s.tgz", name, version), body, 0644)
	if err != nil {
		log.Fatal("Unable to write chart")
	}
}

func addFile(tw *tar.Writer, path string, base string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if stat, err := file.Stat(); err == nil {
		// now lets create the header as needed for this file within the tarball
		header := new(tar.Header)
		header.Name = strings.TrimLeft(strings.TrimLeft(path, base), "/")
		header.Size = stat.Size()
		header.Mode = int64(0644)
		header.ModTime = stat.ModTime()
		// write the header to the tarball archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// copy the file data to the tarball
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
	}
	return nil
}

func fetchFileChart(path string, name string, version string) (*os.File, error) {
	file, err := os.Create(fmt.Sprintf("charts/%s-%s.tar.gz", name, version))
	if err != nil {
		log.Fatalf("Error: %v+", err)
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	repoPath := strings.TrimLeft(path, "file://")

	err = filepath.Walk(repoPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				if err := addFile(tw, path, filepath.Base(filepath.Dir(repoPath))); err != nil {
					log.Fatalf("Error: %v+", err)
				}
			}
			return nil
		})

	return file, err
}

var indexes map[string]Index = map[string]Index{}

func fetchVersion(dependency Dependency) error {
	if !strings.HasPrefix(dependency.Repository, "file://") {
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
		fetchUrlChart(chart, dependency.Name, version.String())
	} else {

		fetchFileChart(dependency.Repository, dependency.Name, dependency.Version)
	}

	return nil
}

func main() {
	requirements, err := fetchRequirements()
	if err != nil {
		fmt.Printf("Error: %v+", err)
		os.Exit(1)
	}

	for _, dependency := range requirements.Dependencies {
		fmt.Printf("Fetching %s @ %s\n", dependency.Name, dependency.Version)
		err := fetchVersion(dependency)
		if err != nil {
			log.Fatalf("Error: %v+", err)
		}
	}
}
