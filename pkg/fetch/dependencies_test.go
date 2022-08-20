package fetch

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/shteou/helm-dependency-fetch/pkg/getters"
	"github.com/shteou/helm-dependency-fetch/pkg/helm"
	"github.com/spf13/afero"

	"github.com/stretchr/testify/assert"
)

type MockGetter struct {
	Response *http.Response
	Error    error
}

func (m MockGetter) Get(ctx context.Context, url string, username string, password string) (*http.Response, error) {
	return m.Response, m.Error
}

func newHelmDependencyFetchTest(fs afero.Fs, getter getters.Getter) *HelmDependencyFetch {
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
				Urls:    []string{"charts/foo-0.1.0-tgz"},
			}},
		},
	}

	// When
	err := hdf.FetchVersion(context.TODO(), helm.Dependency{Name: "foo", Repository: "http://localhost:8080", Version: ">= 0.1.0"})

	// Then
	assert.NoError(t, err, "Failed to fetch chart version")
	stat, err := fs.Stat("charts/foo-0.1.0.tgz")
	assert.NoError(t, err, "Failed to check existence of downloaded chart")
	assert.Greater(t, stat.Size(), int64(10), "Resulting chart package should be more than a few bytes in size")
}

func TestFetchVersionAbsoluteUrl(t *testing.T) {
	fs := afero.NewMemMapFs()
	fs.Mkdir("charts", 0777)
	mockResponse := MockGetter{Response: &http.Response{Body: ioutil.NopCloser(bytes.NewReader([]byte("hello world")))}}
	hdf := newHelmDependencyFetchTest(fs, mockResponse)
	hdf.Indexes["http://localhost:8080"] = helm.Index{
		Entries: map[string][]helm.Entry{
			"foo": {{
				Name:    "foo",
				Version: "0.1.0",
				Urls:    []string{"https://chart-repo.com/charts/foo-0.1.0-tgz"},
			}},
		},
	}

	// When
	err := hdf.FetchVersion(context.TODO(), helm.Dependency{Name: "foo", Repository: "http://localhost:8080", Version: ">= 0.1.0"})

	// Then
	assert.NoError(t, err, "Failed to fetch chart version")
	stat, err := fs.Stat("charts/foo-0.1.0.tgz")
	assert.NoError(t, err, "Failed to check existence of downloaded chart")
	assert.Greater(t, stat.Size(), int64(10), "Resulting chart package should be more than a few bytes in size")
}
