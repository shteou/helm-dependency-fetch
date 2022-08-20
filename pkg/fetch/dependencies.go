package fetch

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os/exec"
	"sort"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/shteou/helm-dependency-fetch/pkg/helm"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v1"
)

func (f *HelmDependencyFetch) FetchVersion(ctx context.Context, dependency helm.Dependency) error {
	if !strings.HasPrefix(dependency.Repository, "file://") {
		repos, err := f.parseRepositories()
		if err != nil {
			return err
		}

		username, password := getCredsForRepository(repos, dependency.Repository)

		index, err := f.getIndex(ctx, dependency.Repository, username, password)
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
		return f.fetchURLChart(ctx, chartUrlString, dependency.Name, version.String(), username, password)
	}

	return f.fetchFileChart(dependency.Repository)
}

func (f *HelmDependencyFetch) fetchIndex(ctx context.Context, repo string, username string, password string) (*helm.Index, error) {
	index := helm.Index{}

	fmt.Printf("Fetching index from %s\n", repo)

	resp, err := f.Get.Get(ctx, fmt.Sprintf("%s/index.yaml", strings.TrimSuffix(repo, "/")), username, password)

  if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	defer resp.Body.Close()

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

func (f *HelmDependencyFetch) fetchURLChart(ctx context.Context, url string, name string, version string, username string, password string) error {
	fmt.Printf("\tFetching chart: %s\n", url)
	resp, err := f.Get.Get(ctx, url, username, password)
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
