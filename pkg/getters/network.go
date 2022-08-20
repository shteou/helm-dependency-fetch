package getters

import (
	"context"
	"net/http"
)

type Getter interface {
	Get(context.Context, string, string, string) (*http.Response, error)
}

type NetworkGetter struct {
}

func (NetworkGetter) Get(ctx context.Context, url string, username string, password string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if username != "" {
		req.SetBasicAuth(username, password)
	}

	client := NewHttpClient()
	return client.Do(req)
}
