package helpers

import (
	"crypto/sha1"
	"encoding/base64"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

func ReadAndUnmarshal(filename string, out interface{}) (err error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "error loading file.")
	}

	err = yaml.Unmarshal(b, out)
	if err != nil {
		return errors.Wrap(err, "error unmarshalling YAML.")
	}

	return nil
}

func MarshalAndWrite(filename string, in interface{}) error {
	data, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	file.Write(data)

	return nil
}

func OneLineHash(str string) string {
	hasher := sha1.New()
	hasher.Write([]byte(str))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}
