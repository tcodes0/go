package httpmisc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
)

// an http client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	userAgent  string
	timeout    time.Duration
}

// options for setting a client; some options are required.
type SetClientOptions struct {
	Client    *http.Client
	UserAgent string
	BaseURL   string
	APIKey    string
	Timeout   time.Duration
}

// initializes a client with options.
func (c *Client) Init(opts *SetClientOptions) error {
	opts = misc.Default(opts, &SetClientOptions{})

	if opts.UserAgent == "" || opts.BaseURL == "" || opts.APIKey == "" {
		return errors.New("user agent, base url, api key are required")
	}

	opts.Client = misc.Default(opts.Client, &http.Client{})
	opts.Timeout = misc.Default(opts.Timeout, misc.Seconds(15))
	c.httpClient = opts.Client

	dialer := &net.Dialer{Timeout: opts.Timeout}
	c.httpClient.Transport = &http.Transport{Dial: dialer.Dial}

	c.baseURL = opts.BaseURL
	c.apiKey = opts.APIKey
	c.userAgent = opts.UserAgent
	c.timeout = opts.Timeout

	return nil
}

// sends a request with a body and headers.
func (c Client) Request(ctx context.Context, method, resource string, body any, headers http.Header) (*http.Response, []byte, error) {
	if c.httpClient == nil {
		return nil, nil, errors.New("nil client")
	}

	tBody := reflect.TypeOf(body)

	if tBody.Kind() != reflect.Ptr || tBody.Elem().Kind() != reflect.Struct {
		return nil, nil, errors.New("body must be a struct pointer")
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	req, err := makeRequest(ctx, method, c.baseURL+resource, body)
	if err != nil {
		return nil, nil, err
	}

	req.Header = misc.Default(headers, http.Header{})
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("User-Agent", c.userAgent)

	logger := logging.FromContext(ctx)

	logger.Debug().Logf("headers %v", req.Header)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, misc.Wrap(err, "doing request")
	}
	defer res.Body.Close()

	logger.Debug().Logf("status %d", res.StatusCode)

	if res.StatusCode >= http.StatusMultipleChoices {
		return res, nil, fmt.Errorf("status code: %d", res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return res, nil, misc.Wrap(err, "reading response body")
	}

	logger.Debug().Logf("response %s", string(data))

	return res, data, nil
}

func makeRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	logger := logging.FromContext(ctx)

	logger.Debug().Logf("url %s", url)

	if body == nil {
		req, err := http.NewRequestWithContext(ctx, method, url, http.NoBody)
		if err != nil {
			return nil, misc.Wrap(err, "creating request without body")
		}

		return req, nil
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, misc.Wrap(err, "marshalling body")
	}

	logger.Debug().Logf("body %v", body)

	reader, writer := io.Pipe()
	go func() {
		//nolint:govet // scope ok
		var err error
		_, err = io.Copy(writer, bytes.NewReader(data))
		writer.CloseWithError(err)
	}()

	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, misc.Wrap(err, "creating request")
	}

	return req, nil
}

func (c Client) Get(ctx context.Context, resource string, body any, headers http.Header) (*http.Response, []byte, error) {
	return c.Request(ctx, http.MethodGet, resource, body, headers)
}

func (c Client) Post(ctx context.Context, resource string, body any, headers http.Header) (*http.Response, []byte, error) {
	return c.Request(ctx, http.MethodPost, resource, body, headers)
}

func (c Client) Put(ctx context.Context, resource string, body any, headers http.Header) (*http.Response, []byte, error) {
	return c.Request(ctx, http.MethodPut, resource, body, headers)
}

func (c Client) Patch(ctx context.Context, resource string, body any, headers http.Header) (*http.Response, []byte, error) {
	return c.Request(ctx, http.MethodPatch, resource, body, headers)
}

func (c Client) Delete(ctx context.Context, resource string, body any, headers http.Header) (*http.Response, []byte, error) {
	return c.Request(ctx, http.MethodDelete, resource, body, headers)
}
