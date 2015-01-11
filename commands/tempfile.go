// The following code been extracted from Go's source
// code. Adding the suffix parameter to tempFile.

package commands

import (
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// Random number state.
// We generate random temporary file names so that there's a good
// chance the file doesn't exist yet - keeps the number of tries in
// tempFile to a minimum.
var rnd uint32
var rndmu sync.Mutex

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func nextSuffix() string {
	rndmu.Lock()
	r := rnd
	if r == 0 {
		r = reseed()
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rnd = r
	rndmu.Unlock()
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

// tempFile creates a new temporary file in the directory dir
// with a name beginning with prefix, opens the file for reading
// and writing, and returns the resulting *os.File.
// If dir is the empty string, tempFile uses the default directory
// for temporary files (see os.TempDir).
// Multiple programs calling tempFile simultaneously
// will not choose the same file.  The caller can use f.Name()
// to find the pathname of the file.  It is the caller's responsibility
// to remove the file when no longer needed.
func tempFile(dir, prefix, suffix string) (f *os.File, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	nconflict := 0
	for i := 0; i < 10000; i++ {
		name := filepath.Join(dir, prefix+nextSuffix()+suffix)
		f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if os.IsExist(err) {
			if nconflict++; nconflict > 10 {
				rnd = reseed()
			}
			continue
		}
		break
	}
	return
}
