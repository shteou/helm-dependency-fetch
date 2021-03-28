package fetch

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/shteou/helm-dependency-fetch/pkg/helm"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

type MockGetter struct {
	Response *http.Response
	Error    error
}

func (m MockGetter) Get(url string) (*http.Response, error) {
	return m.Response, m.Error
}

func newHelmDependencyFetchTest(fs afero.Fs, getter Getter) *HelmDependencyFetch {
	hdf := HelmDependencyFetch{
		Fs:      fs,
		Get:     getter,
		Indexes: map[string]helm.Index{},
	}
	return &hdf
}

func copyTestData(t *testing.T, fs afero.Fs, src string, target string) {
	bytes, err := ioutil.ReadFile(src)
	assert.Nil(t, err, fmt.Sprintf("Unable to read %s", src))

	err = afero.WriteFile(fs, target, bytes, 0644)
	assert.Nil(t, err, fmt.Sprintf("Unable to write %s", target))
}

func TestParseDependenciesV2(t *testing.T) {
	fs := afero.NewMemMapFs()
	copyTestData(t, fs, "test_data/v2chart/Chart.yaml", "Chart.yaml")
	hdf := newHelmDependencyFetchTest(fs, MockGetter{})

	// When
	deps, err := hdf.ParseDependencies()

	// Then
	assert.NoError(t, err, "Failed to parse ependencies from v2 Chart.yaml")
	assert.Equal(t, 1, len(*deps), "Expected a single dependency to be parsed")
}

func TestParseDependenciesV1(t *testing.T) {
	fs := afero.NewMemMapFs()
	copyTestData(t, fs, "test_data/v1chart/Chart.yaml", "Chart.yaml")
	copyTestData(t, fs, "test_data/v1chart/requirements.yaml", "requirements.yaml")
	hdf := newHelmDependencyFetchTest(fs, MockGetter{})

	// When
	deps, err := hdf.ParseDependencies()

	// Then
	assert.NoError(t, err, "Failed to parse ependencies from v1 Chart.yaml")
	assert.Equal(t, 1, len(*deps), "Expected a single dependency to be parsed")
}

func TestCreateChartsDirectory(t *testing.T) {
	fs := afero.NewMemMapFs()
	hdf := newHelmDependencyFetchTest(fs, MockGetter{})

	// When
	err := hdf.CreateChartsDirectory()

	// Then
	assert.NoError(t, err, "Failed to call CreateChartsDirectory")

	stat, err := fs.Stat("charts")
	assert.NoError(t, err, "Failed to check existence of charts directory")
	assert.True(t, stat.IsDir(), "charts should be a directory")
}

func TestCreateChartsDirectory_AlreadyExists(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.Mkdir("charts", 0777)
	hdf := newHelmDependencyFetchTest(fs, MockGetter{})

	// When
	err := hdf.CreateChartsDirectory()

	// Then
	assert.NoError(t, err, "Failed to call CreateChartsDirectory")

	stat, err := fs.Stat("charts")
	assert.NoError(t, err, "Failed to check existence of charts directory")
	assert.True(t, stat.IsDir(), "charts should be a directory")
}

func TestFetchVersion(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.Mkdir("charts", 0777)
	mockResponse := MockGetter{Response: &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte("hello world")))}}
	hdf := newHelmDependencyFetchTest(fs, mockResponse)
	hdf.Indexes["http://localhost:8080"] = helm.Index{
		Entries: map[string][]helm.Entry{
			"foo": {{
				Name:    "foo",
				Version: "0.1.0",
			}},
		},
	}

	// When
	err := hdf.FetchVersion(helm.Dependency{Name: "foo", Repository: "http://localhost:8080", Version: ">= 0.1.0"})

	// Then
	assert.NoError(t, err, "Failed to fetch chart version")
	stat, err := fs.Stat("charts/foo-0.1.0.tgz")
	assert.NoError(t, err, "Failed to check existence of downloaded chart")
	assert.Greater(t, stat.Size(), int64(10), "Resulting chart package should be more than a few bytes in size")
}
