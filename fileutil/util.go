package fileutil

import (
	"os"
	"os/user"
	"path/filepath"
)

// Join path elements and expand to an absolute path.  If the first element is
// `~`, then the user's home directory will be pre-pended.
func ExpandPath(elem ...string) (string, error) {
	path := filepath.Join(elem...)

	var err error

	if len(path) == 0 {
		return path, nil
	}

	if path[0] == '~' {
		usr, err := user.Current()
		if err != nil {
			return path, err
		}
		path = filepath.Join(usr.HomeDir, path[1:])
	}

	if path, err = filepath.Abs(path); err != nil {
		return path, err
	}

	return path, nil
}

// Return true if the path exists. The path elements are first expanded by
// `ExpandPath`. If there are any errors in the process, then the function will
// return false.
func FileExists(elem ...string) bool {
	path, err := ExpandPath(elem...)
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// Returns true is the path exists. The path elements are first expanded by
// `ExpandPath`.
func IsDir(elem ...string) bool {
	path, err := ExpandPath(elem...)
	if err == nil {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
