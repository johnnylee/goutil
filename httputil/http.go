package httputil

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/johnnylee/goutil/logutil"
)

var log = logutil.New("httputil")

var MaxMemory int64 = 1048576

func SaveFormFileToTmp(r *http.Request, key string) (string, string, error) {
	// Parse the form.
	if err := r.ParseMultipartForm(MaxMemory); err != nil {
		return "", "", err
	}

	// Create a temporary directory.
	tempDir, err := ioutil.TempDir("", "upload-")
	if err != nil {
		log.Err(err, "When creating a temporary directory")
		return "", "", err
	}

	src, header, err := r.FormFile(key)
	if err != nil {
		return tempDir, "", err
	}

	path := filepath.Join(tempDir, header.Filename)
	dst, err := os.Create(path)
	defer dst.Close()

	if err != nil {
		log.Err(err, "When creating output file: %v", path)
	}

	// Copy the file.
	if _, err = io.Copy(dst, src); err != nil {
		log.Err(err, "When copying to a temporary file")
	}

	return tempDir, path, err
}

// Save posted file to a temporary directory and file, returning the directory,
// to full path to each file, and an error.
//
// Use CleanUpTmpDir to delete the temporary directory created.
func SaveFormFilesToTmp(r *http.Request) (string, []string, error) {
	// Parse the form.
	if err := r.ParseMultipartForm(MaxMemory); err != nil {
		return "", nil, err
	}

	// Check for no files.
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return "", nil, nil
	}

	// Our output paths.
	paths := make([]string, 0)

	// Create a temporary directory.
	tempDir, err := ioutil.TempDir("", "upload-")
	if err != nil {
		return "", paths, err
	}

	// Loop through files and save to the temporary directory.
	for i := range r.MultipartForm.File {
		for j := range r.MultipartForm.File[i] {
			header := r.MultipartForm.File[i][j]
			if len(header.Filename) == 0 {
				continue
			}

			src, err := header.Open()
			if err != nil {
				return "", paths, err
			}

			path := filepath.Join(tempDir, header.Filename)
			paths = append(paths, path)

			dst, err := os.Create(path)
			defer dst.Close()

			if err != nil {
				return tempDir, paths, err
			}

			// Copy the file.
			if _, err = io.Copy(dst, src); err != nil {
				return tempDir, paths, err
			}
		}
	}

	return tempDir, paths, nil
}

// Delete the temporary directory created by the `SaveFormFilesToTmp` function.
func CleanUpTmpDir(tmpDir string) {
	if len(tmpDir) == 0 {
		return
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		log.Err(err, "When removing temporary directory %v", tmpDir)
	}
}
