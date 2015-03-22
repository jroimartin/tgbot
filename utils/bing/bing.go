// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bing

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// response represents the set of results returned by Bing.
type response struct {
	D struct {
		Results []Result
		Next    string `json:"__next"`
	}
}

// A Result represents one of the results returned by Bing.
type Result struct {
	MetaData    ResultMetadata `json:"__metadata"`
	Id          string
	Title       string
	Description string
	DisplayUrl  string
	MediaUrl    string
	Url         string
}

// ResultMetadata is the metadata linked to each result.
type ResultMetadata struct {
	Uri  string
	Type string
}

// A Client defines the parameters needed to perform a search using the
// Bing API.
type Client struct {
	key    string
	client *http.Client

	// Limit defines the maximum number of result pages to be retrieved.
	Limit int
}

// NewClient returns a new Client. The parameter key allows to specify
// the API key.
func NewClient(key string) Client {
	c := Client{}
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	c.client = &http.Client{Transport: tr}
	c.key = key
	c.Limit = 1
	return c
}

// A Kind defines the search type (web, images, video or news).
type Kind int

const (
	Web Kind = iota
	Image
	Video
	News
)

// String returns an string with the ASCII representation of the Kind object.
func (k Kind) String() string {
	switch k {
	case Web:
		return "Web"
	case Image:
		return "Image"
	case Video:
		return "Video"
	case News:
		return "News"
	}
	return fmt.Sprintf("Kind(%d)", k)
}

// Query sends a new query to Bing and returns the results.
func (c Client) Query(k Kind, q string) ([]Result, error) {
	uri := "https://api.datamarket.azure.com/Bing/Search/v1/" +
		k.String() + "?Query='" + q + "'&Adult='Off'&$format=json"

	results := []Result{}
	for i := 0; i < c.Limit; i++ {
		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth("", c.key)
		resp, err := c.client.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, errors.New("returned status is not OK")
		}

		br := response{}
		if err := json.Unmarshal(body, &br); err != nil {
			return nil, err
		}
		results = append(results, br.D.Results...)

		if br.D.Next != "" {
			uri = br.D.Next + "&$format=json"
		} else {
			break
		}
	}
	return results, nil
}
