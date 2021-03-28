package fetch

import (
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

func copyTestData(t *testing.T, fs afero.Fs, src string, target string) {
	bytes, err := ioutil.ReadFile(src)
	assert.Nil(t, err, fmt.Sprintf("Unable to read %s", src))

	err = afero.WriteFile(fs, target, bytes, 0644)
	assert.Nil(t, err, fmt.Sprintf("Unable to write %s", target))
}

func TestParseDependenciesV2(t *testing.T) {
	fs := afero.NewMemMapFs()
	copyTestData(t, fs, "test_data/v2chart/Chart.yaml", "Chart.yaml")

	hdf := HelmDependencyFetch{Fs: fs}
	deps, err := hdf.ParseDependencies()

	assert.Nil(t, err, "Failed to parse ependencies from v2 Chart.yaml")
	assert.Equal(t, 1, len(*deps), "Expected a single dependency to be parsed")
}

func TestParseDependenciesV1(t *testing.T) {
	fs := afero.NewMemMapFs()
	copyTestData(t, fs, "test_data/v1chart/Chart.yaml", "Chart.yaml")
	copyTestData(t, fs, "test_data/v1chart/requirements.yaml", "requirements.yaml")

	hdf := HelmDependencyFetch{Fs: fs}
	deps, err := hdf.ParseDependencies()

	assert.Nil(t, err, "Failed to parse ependencies from v1 Chart.yaml")
	assert.Equal(t, 1, len(*deps), "Expected a single dependency to be parsed")
}
