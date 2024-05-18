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
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/tcodes0/go/src/errutil"
)

var defaultHTTPTimeout = 15 * time.Second

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

type SetClientOptions struct {
	Client  *http.Client
	Timeout time.Duration
}

func (c *Client) SetClient(userAgent, baseURL, APIKey string, opts *SetClientOptions) error {
	if opts == nil {
		opts = &SetClientOptions{}
	}

	if opts.Client == nil {
		opts.Client = &http.Client{}
	}

	c.httpClient = opts.Client

	if opts.Timeout == 0 {
		opts.Timeout = defaultHTTPTimeout
	}

	dialer := &net.Dialer{
		Timeout: opts.Timeout,
	}

	c.httpClient.Transport = &http.Transport{
		Dial: dialer.Dial,
	}

	c.baseURL = baseURL
	c.apiKey = APIKey
	c.userAgent = userAgent

	return nil
}

func (c Client) Request(ctx context.Context, method, path string, body any, headers http.Header, debug bool) (*http.Response, []byte, error) {
	if c.httpClient == nil {
		return nil, nil, errors.New("run Configure first")
	}

	rType := reflect.TypeOf(body)

	if rType.Kind() != reflect.Ptr {
		return nil, nil, errors.New("body must be a pointer")
	}

	if rType.Elem().Kind() != reflect.Struct {
		return nil, nil, errors.New("body must be a struct pointer")
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		timeout := defaultHTTPTimeout

		if debug {
			timeout = 5 * time.Minute
		}

		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	var logger zerolog.Logger
	if ctxLogger := zerolog.Ctx(ctx); ctxLogger != nil {
		logger = *ctxLogger
	} else {
		logger = zerolog.Nop()
	}

	if !strings.HasPrefix(path, "/") {
		path = fmt.Sprintf("/%s", path)
	}

	URL := fmt.Sprintf("%s%s", c.baseURL, path)

	logger.Debug().Interface("body", body).Msg("request")

	var b []byte
	if body != nil {
		var err error
		b, err = json.Marshal(body)
		if err != nil {
			return nil, nil, errutil.Wrap(err, "marshalling body")
		}
	}

	pr, pw := io.Pipe()

	go func() {
		_, err := io.Copy(pw, bytes.NewReader(b))
		pw.CloseWithError(err)
	}()

	req, err := http.NewRequestWithContext(ctx, method, URL, pr)
	if err != nil {
		return nil, nil, errutil.Wrap(err, "creating request")
	}

	if headers != nil {
		req.Header = headers
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("User-Agent", c.userAgent)

	logger.Debug().Str("URL", URL).Msg("request")
	logger.Debug().Interface("Headers", req.Header).Msg("request")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, errutil.Wrap(err, "doing request")
	}
	defer res.Body.Close()

	logger.Debug().Int("status", res.StatusCode).Msg("request")

	if res.StatusCode >= http.StatusMultipleChoices {
		return res, nil, fmt.Errorf("status code: %d", res.StatusCode)
	}

	b, err = io.ReadAll(res.Body)
	if err != nil {
		return res, nil, errutil.Wrap(err, "reading response body")
	}

	logger.Debug().Bytes("response", b).Msg("request")

	return res, b, nil
}
