// Copyright 2015 The tgbot Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

// download downloads the file on the given URL to a
// temporary directory and return the path where it has
// been saved.
func download(dstDir string, targetURL string) (filePath string, err error) {
	res, err := http.Get(targetURL)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error: %v (%v)", res.Status, res.StatusCode)
	}

	// Parse URL to get its path's "base"
	u, err := url.Parse(targetURL)
	if err != nil {
		return "", err
	}
	f, err := openCreateFile(dstDir, path.Base(u.Path))
	if err != nil && err != errorNew {
		return "", err
	}
	defer f.Close()

	if err == errorNew {
		_, err = io.Copy(f, res.Body)
		if err != nil {
			return "", err
		}
	}
	return f.Name(), nil
}

// openCreateFile tries to open an existing file with the
// given name at the specified destination directory. If
// the file does not exist, it will create a new one. It is
// important to note that the error will be errorNew if it
// was necessary to create the file.
func openCreateFile(dstDir string, filename string) (*os.File, error) {
	var err, ferr error

	path := path.Join(dstDir, filename)

	var f *os.File
	f, err = os.Open(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if os.IsNotExist(err) {
		f, err = os.Create(path)
		if err != nil {
			return nil, err
		}
		ferr = errorNew
	}

	return f, ferr
}
