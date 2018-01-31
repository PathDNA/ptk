package ptk

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type HTTPClient struct {
	http.Client
	DefaultHeaders http.Header
}

func (c *HTTPClient) AllowInsecureTLS(v bool) (old bool) {
	tr, ok := c.Transport.(*http.Transport)
	if !ok {
		tr = &http.Transport{}
		c.Transport = tr
	}

	if tr.TLSClientConfig == nil {
		tr.TLSClientConfig = &tls.Config{}
	}

	old, tr.TLSClientConfig.InsecureSkipVerify = tr.TLSClientConfig.InsecureSkipVerify, v

	return
}

// RequestCtx is a smart wrapper to handle http requests.
// - ctx: is a context.Context in case you want more control over canceling the request.
// - method: http method (GET, PUT, POST, etc..), if empty it defaults to GET.
// - ct: request content-type.
// - url: the request's url.
// - reqData: data to pass to POST/PUT requests, if it's an `io.Reader`, a `[]byte` or a `string`,
//	it will be passed as-is, any other object will be encoded as JSON.
// - respData: data object to get the response or `nil`, can be , `io.Writer`, `func(io.Reader) error`
//	to read the body directly, `func(*http.Response) error` to process the actual response,
//	or a pointer to an object to decode a JSON body into.
func (c *HTTPClient) RequestCtx(ctx context.Context, method, ct, url string, reqData, respData interface{}) error {
	var r io.Reader

	switch in := reqData.(type) {
	case nil:
	case io.Reader:
		r = in
	case []byte:
		r = bytes.NewReader(in)
	case string:
		r = strings.NewReader(in)
	default:
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(reqData); err != nil {
			return err
		}
		r = &buf
		if ct == "" {
			ct = "application/json"
		}
	}

	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return err
	}

	if ctx != context.Background() {
		req = req.WithContext(ctx)
	}

	if ct != "" {
		req.Header.Add("Content-Type", ct)
	}

	if len(c.DefaultHeaders) > 0 {
		req.Header = c.DefaultHeaders
	}

	resp, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("%s error: %v", req.URL, err)
	}
	defer resp.Body.Close()

	switch out := respData.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(out, resp.Body)
	case func(r io.Reader) error:
		err = out(resp.Body)
	case func(r *http.Response) error:
		err = out(resp)
	default:
		err = json.NewDecoder(resp.Body).Decode(out)
	}

	return err
}

// Request is a wrapper for `RequestCtx(context.Background(), method, ct, url, reqData, respData)`
func (c *HTTPClient) Request(method, ct, url string, reqData, respData interface{}) error {
	return c.RequestCtx(context.Background(), method, ct, url, reqData, respData)
}

var DefaultClient HTTPClient

func RequestCtx(ctx context.Context, method, ct, url string, reqData, respData interface{}) error {
	return DefaultClient.RequestCtx(ctx, method, ct, url, reqData, respData)
}

func Request(method, ct, url string, reqData, respData interface{}) error {
	return DefaultClient.Request(method, ct, url, reqData, respData)
}
