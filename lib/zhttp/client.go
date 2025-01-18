package zhttp

import (
	"context"
	"errors"
	"io"
	"net/http"
)

type Requests struct {
	client *http.Client
	before func(r *http.Request)
	after  func(r *http.Response) error
	Error  error
}

func NewRequest() *Requests {
	return &Requests{
		client: http.DefaultClient,
		before: func(_ *http.Request) {},
		after: func(r *http.Response) error {
			if r.StatusCode != 200 {
				return errors.New(r.Status)
			}
			return nil
		},
		Error: nil,
	}
}

func BreakRetryError(err error) error {
	return &errorRetry{err}
}

type errorRetry struct{ e error }

func (r *errorRetry) Error() string { return r.e.Error() }

func (c *Requests) Client(client *http.Client) *Requests {
	c.client = client
	return c
}

func (c *Requests) Before(before func(r *http.Request)) *Requests {
	c.before = before
	return c
}

func (c *Requests) After(after func(r *http.Response) error) *Requests {
	c.after = after
	return c
}

func (c *Requests) Do(method, url string, body io.Reader) error {
	c.Error = c.do(method, url, body)
	return c.Error
}

func (c *Requests) DoWithRetry(ctx context.Context, method,
	url string, body io.Reader, opts ...Retry) error {
	for range NewRetryTimer(ctx, opts...) {
		if c.Error = c.do(method, url, body); c.Error == nil {
			break
		}
		var constraint *errorRetry
		if errors.As(c.Error, &constraint) {
			break
		}
	}
	return c.Error
}

func (c *Requests) do(method, url string, body io.Reader) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	if c.before != nil {
		c.before(req)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if c.after != nil {
		return c.after(resp)
	}
	return nil
}
