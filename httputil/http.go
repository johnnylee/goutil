package httputil

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/schema"
	"github.com/johnnylee/goutil/logutil"
	"github.com/julienschmidt/httprouter"
)

var log = logutil.New("httputil")

var MaxMemory int64 = 4194304

var schemaDecoder *schema.Decoder
var tokenHandler TokenHandler

func init() {
	schemaDecoder = schema.NewDecoder()
	schemaDecoder.IgnoreUnknownKeys(true)
	SetSessionKeys(randBytes(32), randBytes(32))
}

// Override the default schema decoder.
func SetSchemaDecoder(decoder *schema.Decoder) {
	schemaDecoder = decoder
}

// Both keys should be 32 bytes.
func SetSessionKeys(signingKey, cryptKey []byte) {
	var err error
	tokenHandler, err = NewTokenHandler(signingKey, cryptKey)
	if err != nil {
		panic(err)
	}
}

/*************
 * Arguments *
 *************/

// Decode form arguments. Arguments are first pulled from the URL (httprouter),
// then from the form.
func DecodeForm(
	r *http.Request, ps httprouter.Params, args interface{},
) error {
	if args == nil {
		return nil
	}

	if err := r.ParseMultipartForm(MaxMemory); err != nil {
		return err
	}

	// Read data from URL.
	data := map[string][]string{}
	for _, p := range ps {
		data[p.Key] = []string{p.Value}
	}

	if err := schemaDecoder.Decode(args, data); err != nil {
		log.Err(err, "When decoding URL arguments")
		return err
	}

	// Read args from form.
	if err := schemaDecoder.Decode(args, r.Form); err != nil {
		log.Err(err, "When decoding form")
		return err
	}

	return nil
}

func DecodeJSON(r *http.Request, args interface{}) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&args); err != nil {
		log.Err(err, "When decoding JSON request body")
		return err
	}
	return nil
}

/***************
 * File Upload *
 ***************/

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

/************
 * Sessions *
 ************/

func StoreSession(
	w http.ResponseWriter, key string, maxAge int, session interface{},
) error {
	value, err := tokenHandler.Encode(session)
	if err != nil {
		log.Err(err, "When encoding session cookie")
		return err
	}

	cookie := http.Cookie{
		Name:     key,
		Value:    string(value),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   maxAge,
	}

	http.SetCookie(w, &cookie)
	return nil
}

func LoadSession(r *http.Request, key string, session interface{}) error {
	cookie, err := r.Cookie(key)
	if err != nil {
		if err != http.ErrNoCookie {
			log.Err(err, "When reading cookie: %v", key)
		}
		return err
	}

	err = tokenHandler.Decode([]byte(cookie.Value), session)
	if err != nil {
		log.Err(err, "When decoding cookie")
	}

	return err
}

/*************
 * Responses *
 *************/

// Redirect with 302 status. This isn't the most modern way to do this, but it
// seems to be more robust in older browsers.
func Redirect(
	w http.ResponseWriter, r *http.Request, URL string, args ...interface{},
) {
	http.Redirect(w, r, fmt.Sprintf(URL, args...), http.StatusFound)
}

func RespondJSON(w http.ResponseWriter, obj interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	encoded, err := json.Marshal(obj)
	if err != nil {
		log.Err(err, "When encoding request response")
		encoded = []byte("An unknown error occured.")
	}

	_, _ = w.Write(encoded)
}
