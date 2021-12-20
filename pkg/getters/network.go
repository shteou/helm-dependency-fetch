package getters

import (
	"net/http"
	"time"
)

type Getter interface {
	Get(string, string, string) (*http.Response, error)
}

type NetworkGetter struct {
}

func (NetworkGetter) Get(url string, username string, password string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if username != "" {
		req.SetBasicAuth(username, password)
	}

	client := &http.Client{
		Timeout: time.Second * 60,
	}

	return client.Do(req)
}
