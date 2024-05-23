package httpclient

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

	"github.com/tcodes0/go/src/errutil"
	"github.com/tcodes0/go/src/logging"
	"github.com/tcodes0/go/src/misc"
	"github.com/tcodes0/go/src/reflectutil"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	userAgent  string
	timeout    time.Duration
}

type SetClientOptions struct {
	Client  *http.Client
	Timeout time.Duration
}

func (c *Client) Init(userAgent, baseURL, apiKey string, opts *SetClientOptions) {
	opts = reflectutil.Default(opts, &SetClientOptions{})
	opts.Client = reflectutil.Default(opts.Client, &http.Client{})
	opts.Timeout = reflectutil.Default(opts.Timeout, misc.Seconds(15))

	c.httpClient = opts.Client

	dialer := &net.Dialer{Timeout: opts.Timeout}
	c.httpClient.Transport = &http.Transport{Dial: dialer.Dial}

	c.baseURL = baseURL
	c.apiKey = apiKey
	c.userAgent = userAgent
	c.timeout = opts.Timeout
}

func (c Client) Request(ctx context.Context, method, path string, body any, headers http.Header) (*http.Response, []byte, error) {
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

	req, err := getRequest(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, nil, err
	}

	req.Header = reflectutil.Default(headers, http.Header{})
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("User-Agent", c.userAgent)

	logger := logging.FromContext(ctx)

	logger.Debug().Logf("headers %v", req.Header)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, errutil.Wrap(err, "doing request")
	}
	defer res.Body.Close()

	logger.Debug().Logf("status %d", res.StatusCode)

	if res.StatusCode >= http.StatusMultipleChoices {
		return res, nil, fmt.Errorf("status code: %d", res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return res, nil, errutil.Wrap(err, "reading response body")
	}

	logger.Debug().Logf("response %s", string(data))

	return res, data, nil
}

func getRequest(ctx context.Context, method, url string, body any) (*http.Request, error) {
	logger := logging.FromContext(ctx)

	logger.Debug().Logf("url %s", url)

	if body == nil {
		req, err := http.NewRequestWithContext(ctx, method, url, http.NoBody)
		if err != nil {
			return nil, errutil.Wrap(err, "creating request without body")
		}

		return req, nil
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, errutil.Wrap(err, "marshalling body")
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
		return nil, errutil.Wrap(err, "creating request")
	}

	return req, nil
}
