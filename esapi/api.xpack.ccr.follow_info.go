// Code generated from specification version 8-0-0-SNAPSHOT: DO NOT EDIT

package esapi

import (
	"context"
	"strings"
)

func newCCRFollowInfoFunc(t Transport) CCRFollowInfo {
	return func(o ...func(*CCRFollowInfoRequest)) (*Response, error) {
		var r = CCRFollowInfoRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/current/ccr-get-follow-info.html.
//
type CCRFollowInfo func(o ...func(*CCRFollowInfoRequest)) (*Response, error)

// CCRFollowInfoRequest configures the Ccr  Follow Info API request.
//
type CCRFollowInfoRequest struct {
	Index []string

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r CCRFollowInfoRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "GET"

	path.Grow(1 + len(strings.Join(r.Index, ",")) + 1 + len("_ccr") + 1 + len("info"))
	if len(r.Index) > 0 {
		path.WriteString("/")
		path.WriteString(strings.Join(r.Index, ","))
	}
	path.WriteString("/")
	path.WriteString("_ccr")
	path.WriteString("/")
	path.WriteString("info")

	params = make(map[string]string)

	if r.Pretty {
		params["pretty"] = "true"
	}

	if r.Human {
		params["human"] = "true"
	}

	if r.ErrorTrace {
		params["error_trace"] = "true"
	}

	if len(r.FilterPath) > 0 {
		params["filter_path"] = strings.Join(r.FilterPath, ",")
	}

	req, _ := newRequest(method, path.String(), nil)

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	res, err := transport.Perform(req)
	if err != nil {
		return nil, err
	}

	response := Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	return &response, nil
}

// WithContext sets the request context.
//
func (f CCRFollowInfo) WithContext(v context.Context) func(*CCRFollowInfoRequest) {
	return func(r *CCRFollowInfoRequest) {
		r.ctx = v
	}
}

// WithIndex - a list of index patterns; use `_all` to perform the operation on all indices.
//
func (f CCRFollowInfo) WithIndex(v ...string) func(*CCRFollowInfoRequest) {
	return func(r *CCRFollowInfoRequest) {
		r.Index = v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f CCRFollowInfo) WithPretty() func(*CCRFollowInfoRequest) {
	return func(r *CCRFollowInfoRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f CCRFollowInfo) WithHuman() func(*CCRFollowInfoRequest) {
	return func(r *CCRFollowInfoRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f CCRFollowInfo) WithErrorTrace() func(*CCRFollowInfoRequest) {
	return func(r *CCRFollowInfoRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f CCRFollowInfo) WithFilterPath(v ...string) func(*CCRFollowInfoRequest) {
	return func(r *CCRFollowInfoRequest) {
		r.FilterPath = v
	}
}