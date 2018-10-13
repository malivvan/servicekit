package servicekit

import (
	"io/ioutil"
	"encoding/json"
	"path/filepath"
)

type Workdir string

func (w Workdir) Path(p ...string) string {
	if len(p) == 0 {
		return string(w)
	}
	return filepath.Join(append([]string{string(w)}, p...)...)
}

func Configure(path string, config interface{}, defaultConfig interface{}) error {
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		b, err := json.MarshalIndent(defaultConfig, "", "  ")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(path, b, 0600)
		if err != nil {
			return err
		}
	}
	configBytes, err = ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(configBytes, config)
}
