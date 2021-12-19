package fetch

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/shteou/helm-dependency-fetch/pkg/helm"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v1"
	"helm.sh/helm/v3/pkg/helmpath"
)

type HelmDependencyFetch struct {
	Fs      afero.Fs
	Get     Getter
	Indexes map[string]helm.Index
}

type Getter interface {
	Get(string, string, string) (*http.Response, error)
}

type NetworkGetter struct {
}

func (NetworkGetter) Get(url string, username string, password string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if username != "" {
		req.SetBasicAuth(username, password)
	}

	client := &http.Client{
		Timeout: time.Second * 60,
	}

	return client.Do(req)
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

// ParseRepositories loads the helm repositories config file,
// if it exists.
func (f *HelmDependencyFetch) ParseRepositories() (*helm.Repositories, error) {
	base := helmpath.ConfigPath()

	data, err := afero.ReadFile(f.Fs, path.Join(base, "repositories.yaml"))

	if errors.Is(err, afero.ErrFileNotFound) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	repositories := helm.Repositories{}
	err = yaml.Unmarshal([]byte(data), &repositories)

	return &repositories, err
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

func getCredsForRepository(repos *helm.Repositories, targetRepository string) (string, string) {
	if repos == nil {
		return "", ""
	}

	for _, repo := range repos.Repositories {
		if repo.Url == targetRepository {
			return repo.Username, repo.Password
		}
	}

	return "", ""
}

func (f *HelmDependencyFetch) FetchVersion(dependency helm.Dependency) error {
	if !strings.HasPrefix(dependency.Repository, "file://") {
		repos, err := f.ParseRepositories()
		if err != nil {
			return err
		}

		username, password := getCredsForRepository(repos, dependency.Repository)

		index, err := f.getIndex(dependency.Repository, username, password)
		if err != nil {
			return err
		}

		version, entry, err := resolveSemver(dependency.Version, index.Entries[dependency.Name])
		if err != nil {
			return err
		}

		chartUrl, err := url.Parse(entry.Urls[0])
		if err != nil {
			return err
		}

		var chartUrlString string
		if chartUrl.IsAbs() {
			chartUrlString = entry.Urls[0]
		} else {
			chartUrlString = fmt.Sprintf("%s/%s", strings.TrimSuffix(dependency.Repository, "/"), entry.Urls[0])
		}
		return f.fetchURLChart(chartUrlString, dependency.Name, version.String(), username, password)
	}

	return f.fetchFileChart(dependency.Repository)
}

func (f *HelmDependencyFetch) fetchIndex(repo string, username string, password string) (*helm.Index, error) {
	index := helm.Index{}

	fmt.Printf("Fetching index from %s\n", repo)

	resp, err := f.Get.Get(fmt.Sprintf("%s/index.yaml", strings.TrimSuffix(repo, "/")), username, password)
	defer resp.Body.Close()

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Failed to retrieve index (status: %s)", resp.Status))
	}

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

func findEntryByVersion(version string, entries []helm.Entry) *helm.Entry {
	for _, entry := range entries {
		if entry.Version == version {
			return &entry
		}
	}

	return nil
}

func resolveSemver(version string, entries []helm.Entry) (*semver.Version, *helm.Entry, error) {
	c, err := semver.NewConstraint(version)
	if err != nil {
		return nil, nil, err
	}

	versions := []*semver.Version{}

	for _, entry := range entries {
		v, err := semver.NewVersion(entry.Version)
		if err != nil {
			return nil, nil, err
		}

		a, _ := c.Validate(v)
		if a {
			versions = append(versions, v)
		}
	}

	largest := largestSemver(versions)
	if largest == nil {
		return nil, nil, errors.New("couldn't find a semver to satisfy the constraint")
	}

	return largest, findEntryByVersion(largest.Original(), entries), nil
}

func (f *HelmDependencyFetch) fetchURLChart(url string, name string, version string, username string, password string) error {
	fmt.Printf("\tFetching chart: %s\n", url)
	resp, err := f.Get.Get(url, username, password)
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

func (f *HelmDependencyFetch) getIndex(repo string, username string, password string) (*helm.Index, error) {
	var index helm.Index

	if _, ok := f.Indexes[repo]; !ok {
		retrievedIndex, err := f.fetchIndex(repo, username, password)
		if err != nil {
			return nil, err
		}
		f.Indexes[repo] = *retrievedIndex
	}

	index = f.Indexes[repo]
	return &index, nil
}
