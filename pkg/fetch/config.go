package fetch

import (
	"errors"
	"path"

	"github.com/shteou/helm-dependency-fetch/pkg/helm"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v1"
	"helm.sh/helm/v3/pkg/helmpath"
)

// ParseRepositories loads the helm repositories config file,
// if it exists.
func (f *HelmDependencyFetch) parseRepositories() (*helm.Repositories, error) {
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
