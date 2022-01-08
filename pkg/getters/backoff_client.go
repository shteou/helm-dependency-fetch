package getters

import (
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/jpillora/backoff"
)

type client struct {
	httpClient *http.Client
}

func NewHttpClient() *client {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	httpClient := &http.Client{
		Timeout:   time.Second * 30,
		Transport: transport,
	}

	return &client{
		httpClient: httpClient,
	}
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	b := &backoff.Backoff{
		Min:    1.0 * time.Second,
		Max:    10.0 * time.Second,
		Factor: float64(1.1),
		Jitter: true,
	}

	for {
		select {
		case <-req.Context().Done():
			return nil, errors.New("timed out performing http request")
		case <-time.After(b.Duration()):
		}

		res, err := c.httpClient.Do(req)
		if err != nil {
			// Retry due to connection error
			continue
		}

		sc := res.StatusCode
		// TODO: verify if 408 is a valid retry status
		if sc == 500 || (sc >= 502 && sc <= 504) || sc == 522 || sc == 524 {
			// Retry due to temporary http error
			continue
		}

		return res, err
	}
}
