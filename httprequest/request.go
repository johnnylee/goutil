package httprequest

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/schema"
	"github.com/johnnylee/goutil/logutil"
)

var schemaDecoder *schema.Decoder

var log = logutil.New("httprequest")

func init() {
	log.Msg("Initializing schema decoder...")
	schemaDecoder = schema.NewDecoder()
	schemaDecoder.IgnoreUnknownKeys(true)
}

type Request struct {
	W http.ResponseWriter
	R *http.Request
}

func New(w http.ResponseWriter, r *http.Request) Request {
	return Request{w, r}
}

// Decode the form into the given struct pointer.
func (req Request) DecodeForm(dest interface{}) error {
	err := schemaDecoder.Decode(dest, req.R.Form)
	if err != nil {
		log.Err(err, "When decoding form")
	}

	return err
}

// Save posted file to a temporary directory and file, returning the directory,
// to full path to each file, and an error.
func (req Request) SaveFormFilesToTmp() (string, []string, error) {
	// Check for no files. 
	if req.R.MultipartForm == nil || req.R.MultipartForm.File == nil {
		log.Err(nil, "Form doesn't have any files")
		return "", nil, nil
	}

	paths := make([]string, 0)

	// Create a temporary directory.
	tempDir, err := ioutil.TempDir("", "upload-")
	if err != nil {
		log.Err(err, "When creating a temporary directory")
		return "", paths, err
	}

	for i := range req.R.MultipartForm.File {
		for j := range req.R.MultipartForm.File[i] {
			header := req.R.MultipartForm.File[i][j]
			if len(header.Filename) == 0 {
				continue
			}

			src, err := header.Open()
			if err != nil {
				log.Err(err, "When opening multipart file: %v",
					header.Filename)
				return "", paths, err
			}

			path := filepath.Join(tempDir, header.Filename)
			paths = append(paths, path)

			dst, err := os.Create(path)
			defer dst.Close()

			if err != nil {
				log.Err(err, "When creating output file: %v", path)
				return tempDir, paths, err
			}

			// Copy the file.
			if _, err = io.Copy(dst, src); err != nil {
				log.Err(err, "When copying to a temporary file")
				return tempDir, paths, err
			}
		}
	}

	return tempDir, paths, nil
}

// Delete the temporary directory created by the `SaveFormFileToTmp` function.
func (req Request) CleanUpTmpDir(tmpDir string) {
	if len(tmpDir) == 0 {
		return
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		log.Err(err, "When removing a temporary directory %v", tmpDir)
	}
}
