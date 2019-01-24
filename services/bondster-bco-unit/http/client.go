// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

/*
// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request, body []byte) string {
	// Create return string
	var request []string
	// Add the request string
	url := fmt.Sprintf("%v %v %v", r.Method, r.URL, r.Proto)
	request = append(request, url)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	if body != nil {
		request = append(request, fmt.Sprintf("Body: %+v", string(body)))
	}

	// Return the request as a string
	return strings.Join(request, "\n")
}
*/

type Client struct {
	underlying *http.Client
}

func NewClient() Client {
	return Client{
		underlying: &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 5 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 5 * time.Second,
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify:       true,
					MinVersion:               tls.VersionTLS12,
					MaxVersion:               tls.VersionTLS12,
					PreferServerCipherSuites: false,
					CurvePreferences: []tls.CurveID{
						tls.CurveP521,
						tls.CurveP384,
						tls.CurveP256,
					},
					CipherSuites: CipherSuites,
				},
			},
		},
	}
}

func (client Client) Post(url string, body []byte, headers map[string]string) (contents []byte, code int, err error) {
	var (
		req  *http.Request
		resp *http.Response
	)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Runtime Error %v", r)
		}

		if err != nil && resp != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		} else if resp == nil && err != nil {
			err = fmt.Errorf("Runtime Error no response")
		}

		if err != nil {
			contents = nil
		} else {
			contents, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}()

	req, err = http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return
	}

	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	//log.Debug(formatRequest(req, body))

	resp, err = client.underlying.Do(req)
	if err != nil {
		return
	}

	code = resp.StatusCode
	return
}

func (client Client) Get(url string, headers map[string]string) (contents []byte, code int, err error) {
	var (
		req  *http.Request
		resp *http.Response
	)

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Runtime Error %v", r)
		}

		if err != nil && resp != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		} else if resp == nil && err != nil {
			err = fmt.Errorf("Runtime Error no response")
		}

		if err != nil {
			contents = nil
		} else {
			contents, err = ioutil.ReadAll(resp.Body)
			resp.Body.Close()
		}
	}()

	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.Header.Set("accept", "application/json")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	//log.Debug(formatRequest(req, nil))

	resp, err = client.underlying.Do(req)
	if err != nil {
		return
	}

	code = resp.StatusCode
	return
}
