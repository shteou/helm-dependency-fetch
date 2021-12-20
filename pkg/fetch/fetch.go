package fetch

import (
	"github.com/shteou/helm-dependency-fetch/pkg/getters"
	"github.com/shteou/helm-dependency-fetch/pkg/helm"
	"github.com/spf13/afero"
)

type HelmDependencyFetch struct {
	Fs      afero.Fs
	Get     getters.Getter
	Indexes map[string]helm.Index
}

func NewHelmDependencyFetch() *HelmDependencyFetch {
	hdf := HelmDependencyFetch{
		Fs:      afero.NewOsFs(),
		Get:     getters.NetworkGetter{},
		Indexes: map[string]helm.Index{},
	}
	return &hdf
}
