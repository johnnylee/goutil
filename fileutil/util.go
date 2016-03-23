package fileutil

import (
	"os"
	"os/user"
	"path/filepath"
)

// Join path elements and expand to an absolute path.  If the first element is
// `~`, then the user's home directory will be pre-pended. This function could
// panic if there is no working directory or there is no user home directory.
func ExpandPath(elem ...string) string {
	path := filepath.Join(elem...)

	var err error

	if len(path) == 0 {
		return path
	}

	if path[0] == '~' {
		usr, err := user.Current()
		if err != nil {
			panic(err)
		}
		path = filepath.Join(usr.HomeDir, path[1:])
	}

	if path, err = filepath.Abs(path); err != nil {
		panic(err)
	}

	return path
}

// Return true if the path exists. The path elements are first expanded by
// `ExpandPath`. If there are any errors in the process, then the function will
// return false.
func FileExists(elem ...string) bool {
	path := ExpandPath(elem...)
	_, err := os.Stat(path)
	return err == nil
}

// Returns true is the path exists. The path elements are first expanded by
// `ExpandPath`.
func IsDir(elem ...string) bool {
	path := ExpandPath(elem...)
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// Make all directories in the path. The path permissions are set to 0700 on
// components.
func MkdirAll(elem ...string) error {
	path := ExpandPath(elem...)
	return os.MkdirAll(path, 0700)
}
