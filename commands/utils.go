// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

// download downloads the given URL to the directory dstDir in a file with a
// random name and the extension ext. If ext is "", the file will be created
// with the same extension of the original file at the given url. It returns the
// path of the created file. It also creates a temporary directory if dstDir is
// "".
func download(dstDir, ext, targetURL string) (filePath string, err error) {
	res, err := http.Get(targetURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %v (%v)", res.Status, res.StatusCode)
	}

	if ext == "" {
		// Parse URL to get its extension
		u, err := url.Parse(targetURL)
		if err != nil {
			return "", err
		}
		ext = path.Ext(u.Path)
	}

	f, err := tempFile(dstDir, "", ext)
	if err != nil {
		return "", nil
	}
	defer f.Close()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}
