// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package utils

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// A bingResults represents the set of results returned by Bing.
type bingResults struct {
	D struct {
		Results []BingResult
		Next    string `json:"__next"`
	}
}

// A BingResult represents one of the results returned by Bing.
type BingResult struct {
	MetaData    BingResultMetadata `json:"__metadata"`
	Id          string
	Title       string
	Description string
	DisplayUrl  string
	MediaUrl    string
	Url         string
}

// BingResultMetadata is the metadata linked to each result.
type BingResultMetadata struct {
	Uri  string
	Type string
}

// A BingSearch defines the parameters needed to perform a search using the
// Bing API.
type BingSearch struct {
	key    string
	client *http.Client

	// Limit defines the maximum number of result pages to be retrieved.
	Limit int
}

// NewBingSearch returns a new BingSearch. The parameter key allows to specify
// the key needed by API to be used.
func NewBingSearch(key string) BingSearch {
	bs := BingSearch{}
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	bs.client = &http.Client{Transport: tr}
	bs.key = key
	bs.Limit = 1
	return bs
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
func (bs BingSearch) Query(k Kind, q string) ([]BingResult, error) {
	uri := "https://api.datamarket.azure.com/Bing/Search/v1/" +
		k.String() + "?Query='" + q + "'&$format=json"

	results := []BingResult{}
	for i := 0; i < bs.Limit; i++ {
		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth("", bs.key)
		resp, err := bs.client.Do(req)
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

		br := bingResults{}
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
