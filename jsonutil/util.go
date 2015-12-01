package jsonutil

import (
	"bytes"
	"encoding/json"
	"github.com/johnnylee/goutil/fileutil"
	"io/ioutil"
)

// Load into v the json file at the give path. The path elements, `pathElem`
// are expanded by `fileutil.ExpandPath`.
func Load(v interface{}, pathElem ...string) error {
	path, err := fileutil.ExpandPath(pathElem...)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, &v)
}

// Store the data in v into a give file path. The path elements `pathElem` are
// expanded by `fileutil.ExpandPath`.
func Store(v interface{}, pathElem ...string) error {
	path, err := fileutil.ExpandPath(pathElem...)
	if err != nil {
		return err
	}

	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	if err = json.Indent(&out, buf, "", "\t"); err != nil {
		return err
	}

	return ioutil.WriteFile(path, out.Bytes(), 0600)
}
