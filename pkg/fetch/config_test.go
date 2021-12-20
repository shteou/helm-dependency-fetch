package fetch

import (
	"testing"

	"github.com/shteou/helm-dependency-fetch/pkg/helm"
	"github.com/stretchr/testify/assert"
)

func TestCredsForRepository(t *testing.T) {
	repos := helm.Repositories{Repositories: []helm.Repository{{Password: "foo", Username: "bar", Url: "http://localhost:8080"}}}
	user, pass := getCredsForRepository(&repos, "http://localhost:8080")
	assert.Equal(t, "foo", pass)
	assert.Equal(t, "bar", user)
}
