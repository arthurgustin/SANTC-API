package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

var (
	ErrServerBadRequest       = errors.New("server responded with bad request")
	ErrServerError            = errors.New("server responded server error")
	ErrServerUnexpectedStatus = errors.New("server responded with unexpected status")
)

type Client interface {
	AddImageApprovalRequest(ctx context.Context, approval PhotoRequestTransport) error
}

type DefaultClient struct {
	protocol, hostname string
}

func NewDefaultClient(protocol, hostname string) (Client, error) {
	return &DefaultClient{
		protocol: protocol,
		hostname: hostname,
	}, nil
}

func (c *DefaultClient) AddImageApprovalRequest(ctx context.Context, approval PhotoRequestTransport) error {
	requestBody, err := json.Marshal(approval)
	if err != nil {
		return errors.Wrap(err, "failed to json encode the request filter")
	}

	requestUrl := url.URL{Scheme: c.protocol, Host: c.hostname, Path: "/api/v1/children"+approval.ChildId+"/photos"}
	req, err := http.NewRequest(http.MethodPost, requestUrl.String(), bytes.NewReader(requestBody))
	if err != nil {
		return errors.Wrap(err, "failed to build associated data images search request")
	}

	_, err = c.performRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to perform request")
	}
	return nil
}

func (c *DefaultClient) performRequest(ctx context.Context, r *http.Request) (*http.Response, error) {
	r = r.WithContext(ctx)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute the http request")
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return resp, nil
	case resp.StatusCode >= 400 && resp.StatusCode < 500:
		err = ErrServerBadRequest
	case resp.StatusCode >= 500:
		err = ErrServerError
	default:
		err = ErrServerUnexpectedStatus
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	return nil, errors.Wrapf(err, "server responded with status code %v, body: %s", resp.StatusCode, b)
}