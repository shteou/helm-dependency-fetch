package fetch

import (
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

func TestParseDependenciesV2(t *testing.T) {
	fs := afero.NewMemMapFs()
	chartBytes, err := ioutil.ReadFile("test_data/v2chart/Chart.yaml")
	assert.Nil(t, err, "Unable to read Chart.yaml")

	err = afero.WriteFile(fs, "Chart.yaml", chartBytes, 0644)
	assert.Nil(t, err, "Unable to write Chart.yaml")

	hdf := HelmDependencyFetch{Fs: fs}
	deps, err := hdf.ParseDependencies()

	assert.Nil(t, err, "Failed to parse ependencies from v2 Chart.yaml")
	assert.Equal(t, 1, len(*deps), "Expected a single dependency to be parsed")
}

func TestParseDependenciesV1(t *testing.T) {
	fs := afero.NewMemMapFs()
	chartBytes, err := ioutil.ReadFile("test_data/v1chart/Chart.yaml")
	assert.Nil(t, err, "Unable to read Chart.yaml")

	err = afero.WriteFile(fs, "Chart.yaml", chartBytes, 0644)
	assert.Nil(t, err, "Unable to write Chart.yaml")

	requirementsBytes, err := ioutil.ReadFile("test_data/v1chart/requirements.yaml")
	assert.Nil(t, err, "Unable to read requirements.yaml")

	err = afero.WriteFile(fs, "requirements.yaml", requirementsBytes, 0644)
	assert.Nil(t, err, "Unable to write requirements.yaml")

	hdf := HelmDependencyFetch{Fs: fs}
	deps, err := hdf.ParseDependencies()

	assert.Nil(t, err, "Failed to parse ependencies from v1 Chart.yaml")
	assert.Equal(t, 1, len(*deps), "Expected a single dependency to be parsed")
}
