package fetch

import (
	"context"
	"os"

	"github.com/shteou/helm-dependency-fetch/pkg/helm"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v1"
)

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

func (f *HelmDependencyFetch) CreateChartsDirectory() error {
	_, err := f.Fs.Stat("charts")
	if os.IsNotExist(err) {
		err = f.Fs.Mkdir("charts", 0755)
		return err
	}
	return err
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

func (f *HelmDependencyFetch) getIndex(ctx context.Context, repo string, username string, password string) (*helm.Index, error) {
	var index helm.Index

	if _, ok := f.Indexes[repo]; !ok {
		retrievedIndex, err := f.fetchIndex(ctx, repo, username, password)
		if err != nil {
			return nil, err
		}
		f.Indexes[repo] = *retrievedIndex
	}

	index = f.Indexes[repo]
	return &index, nil
}
