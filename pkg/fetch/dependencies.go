package fetch

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/shteou/helm-dependency-fetch/pkg/helm"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v1"
)

type HelmDependencyFetch struct {
	Fs      afero.Fs
	Get     Getter
	Indexes map[string]helm.Index
}

type Getter interface {
	Get(string) (*http.Response, error)
}

type NetworkGetter struct {
}

func (NetworkGetter) Get(url string) (*http.Response, error) {
	return http.Get(url)
}

func NewHelmDependencyFetch() *HelmDependencyFetch {
	hdf := HelmDependencyFetch{
		Fs:      afero.NewOsFs(),
		Get:     NetworkGetter{},
		Indexes: map[string]helm.Index{},
	}
	return &hdf
}

func (f *HelmDependencyFetch) CreateChartsDirectory() error {
	_, err := f.Fs.Stat("charts")
	if os.IsNotExist(err) {
		err = f.Fs.Mkdir("charts", 0755)
		return err
	}
	return err
}

func (f *HelmDependencyFetch) ParseDependencies() (*[]helm.Dependency, error) {
	chart, err := f.fetchChart()
	if err != nil {
		return nil, err
	}

	if chart.APIVersion == "v1" {
		return f.fetchRequirements()
	}
	return &chart.Dependencies, nil
}

func (f *HelmDependencyFetch) FetchVersion(dependency helm.Dependency) error {
	if !strings.HasPrefix(dependency.Repository, "file://") {
		index, err := f.getIndex(dependency.Repository)
		if err != nil {
			return err
		}

		version, err := resolveSemver(dependency.Version, index.Entries[dependency.Name])
		if err != nil {
			return err
		}
		// FIXME: Shouldd determine chart URL from URLs, relative + absolute
		chart := fmt.Sprintf("%s/charts/%s-%s.tgz", strings.TrimSuffix(dependency.Repository, "/"), dependency.Name, version.Original())
		return f.fetchURLChart(chart, dependency.Name, version.String())
	}

	return f.fetchFileChart(dependency.Repository)
}

func (f *HelmDependencyFetch) fetchIndex(repo string) (*helm.Index, error) {
	index := helm.Index{}

	fmt.Printf("Fetching index from %s\n", repo)

	resp, err := f.Get.Get(fmt.Sprintf("%s/index.yaml", strings.TrimSuffix(repo, "/")))

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(body), &index)
	if err != nil {
		return nil, err
	}

	return &index, nil
}

func (f *HelmDependencyFetch) fetchChart() (*helm.Chart, error) {
	data, err := afero.ReadFile(f.Fs, "Chart.yaml")
	if err != nil {
		return nil, err
	}

	chart := helm.Chart{}
	err = yaml.Unmarshal([]byte(data), &chart)

	return &chart, err
}

func (f *HelmDependencyFetch) fetchRequirements() (*[]helm.Dependency, error) {
	requirements := helm.Requirements{}
	data, err := afero.ReadFile(f.Fs, "requirements.yaml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal([]byte(data), &requirements)
	return &requirements.Dependencies, err
}

func largestSemver(versions []*semver.Version) *semver.Version {
	if len(versions) == 0 {
		return nil
	}
	sort.Sort(semver.Collection(versions))
	return versions[len(versions)-1]
}

func resolveSemver(version string, entries []helm.Entry) (*semver.Version, error) {
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
		if a {
			versions = append(versions, v)
		}
	}

	largest := largestSemver(versions)
	if largest == nil {
		return nil, errors.New("couldn't find a semver to satisfy the constraint")
	}

	return largest, nil
}

func (f *HelmDependencyFetch) fetchURLChart(url string, name string, version string) error {
	fmt.Printf("\tFetching chart: %s\n", url)
	resp, err := f.Get.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = afero.WriteFile(f.Fs, fmt.Sprintf("charts/%s-%s.tgz", name, version), body, 0644)
	return err
}

func (f *HelmDependencyFetch) fetchFileChart(path string) error {
	repoPath := strings.TrimPrefix(path, "file://")

	err := exec.Command("helm", "package", repoPath, "-d", "charts/").Run()

	return err
}

func (f *HelmDependencyFetch) getIndex(repo string) (*helm.Index, error) {
	var index helm.Index

	if _, ok := f.Indexes[repo]; !ok {
		retrievedIndex, err := f.fetchIndex(repo)
		if err != nil {
			return nil, err
		}
		f.Indexes[repo] = *retrievedIndex
	}

	index = f.Indexes[repo]
	return &index, nil
}
