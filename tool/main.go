package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"example.com/kaffine/cmd/config"
	"example.com/kaffine/cmd/version"
	"example.com/kaffine/defaults"
	"gopkg.in/yaml.v2"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func readAndUnmarshal(filename string, out interface{}) (err error) {
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

type KaffineConfig struct {
	Directory  string
	ConfigYaml KaffineConfigYaml
	Catalogs   map[string]KRMFunctionCatalog
}

type KaffineConfigYaml struct {
	Catalogs []string
}

// Not type safe, good for prototyping
type KRMFunctionCatalog map[string]interface{}

func NewKaffineConfig(directory string) (c *KaffineConfig, err error) {
	// Create .kaffine directory
	err = os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Create .kaffine/catalogs directory
	err = os.MkdirAll(directory+"/catalogs", os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Create .kaffine/config.yaml if not exists
	if _, err := os.Stat(directory + "config.yaml"); errors.Is(err, os.ErrNotExist) {
		err := os.WriteFile(directory+"config.yaml", defaults.ConfigYaml, 0644)
		if err != nil {
			return nil, err
		}
	}

	// Create .kaffine/installed.yaml if not exists
	if _, err := os.Stat(directory + "installed.yaml"); errors.Is(err, os.ErrNotExist) {
		err := os.WriteFile(directory+"installed.yaml", defaults.BlankCatalog, 0644)
		if err != nil {
			return nil, err
		}
	}

	c = new(KaffineConfig)
	c.Directory = directory
	err = readAndUnmarshal(directory+"config.yaml", &(c.ConfigYaml))
	if err != nil {
		return nil, err
	}

	// Get sha of catalog
	for _, uri := range c.ConfigYaml.Catalogs {
		hasher := sha1.New()
		hasher.Write([]byte(uri))
		shaStr := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		fmt.Println(shaStr)
		// shaBytes := sha1.Sum([]byte(uri))
		// if _, err := os.Stat(directory + "catalogs/" + string(sha) + ".yaml"); errors.Is(err, os.ErrNotExist) {

		// } else {

		// }
	}

	return c, nil
}

var rootCmd = &cobra.Command{
	Use:   "kaffine",
	Short: "Kaffine is a KRM Function Manager",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello!")
	},
}

var LocalConfig *KaffineConfig

func main() {
	var err error

	catalog := map[string]interface{}{}

	err = readAndUnmarshal("../examples/catalogs/example-catalog.yaml", catalog)
	if err != nil {
		log.Fatalf("error loading catalog file. %v\n", err)
	}

	err = validate.AgainstSchema(defaults.CatalogSchema, catalog, strfmt.Default)
	if err != nil {
		log.Fatalf("error validating catalog. %v\n", err)
	}

	LocalConfig, err = NewKaffineConfig("./.kaffine/")
	if err != nil {
		log.Fatalf("error loading config. %v\n", err)
	}

	fmt.Println(LocalConfig)

	return

	rootCmd.AddCommand(version.NewVersionCommand())
	rootCmd.AddCommand(config.NewConfigCommand())

	rootErr := rootCmd.Execute()
	if rootErr != nil {
		log.Fatalf("kaffine encountered an error. %v\n", rootErr)
	}
}
