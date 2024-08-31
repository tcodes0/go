// Copyright 2024 Raphael Thomazella. All rights reserved.
// Use of this source code is governed by the BSD-3-Clause
// license that can be found in the LICENSE file and online
// at https://opensource.org/license/BSD-3-clause.

package apigithub

import (
	"context"
	"io"
	"net/http"

	"github.com/tcodes0/go/logging"
	"github.com/tcodes0/go/misc"
)

var base = "https://api.github.com"

func Get(ctx context.Context, path string, header http.Header, client *http.Client) (*http.Response, error) {
	return req(ctx, http.MethodGet, path, http.NoBody, header, client)
}

func req(ctx context.Context, method, path string, body io.Reader, header http.Header, client *http.Client) (*http.Response, error) {
	logger := logging.FromContext(ctx)

	req, err := http.NewRequestWithContext(ctx, method, base+"/"+path, body)
	if err != nil {
		return nil, misc.Wrapfl(err)
	}

	req.Header = header

	if client == nil {
		client = &http.Client{}
	}

	logger.DebugData(map[string]any{"method": method, "url": base + "/" + path, "headers": header}, "request")

	resp, err := client.Do(req)
	if err != nil {
		return nil, misc.Wrapfl(err)
	}

	logger.DebugData(map[string]any{"status": resp.StatusCode}, "response")

	return resp, nil
}
