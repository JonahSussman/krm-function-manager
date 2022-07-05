package main

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"

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

func oneLineHash(str string) string {
	hasher := sha1.New()
	hasher.Write([]byte(str))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

type KaffineConfig struct {
	Directory  string
	ConfigYaml KaffineConfigYaml
	Catalogs   map[string]KRMFunctionCatalog
	HashToCat  map[string]string
	CatToHash  map[string]string
}

type KaffineConfigYaml struct {
	Catalogs []string
}

// TODO: Make into actual type
type KRMFunctionCatalog map[string]interface{}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}

// Directory must end with '/' character
// TODO: Change to use Go filepath
func NewKaffineConfig(directory string) (c *KaffineConfig, err error) {
	// Create .kaffine directory
	err = os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Create .kaffine/catalogs directory
	err = os.MkdirAll(directory+"catalogs/", os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Create .kaffine/config.yaml if not exists
	if !fileExists(directory + "config.yaml") {
		err := os.WriteFile(directory+"config.yaml", defaults.ConfigYaml, 0644)
		if err != nil {
			return nil, err
		}
	}

	// Create .kaffine/installed.yaml if not exists
	if !fileExists(directory + "installed.yaml") {
		err := os.WriteFile(directory+"installed.yaml", defaults.BlankCatalog, 0644)
		if err != nil {
			return nil, err
		}
	}

	// Create the new KaffineConfig struct
	c = new(KaffineConfig)
	c.Directory = directory
	c.Catalogs = map[string]KRMFunctionCatalog{}
	c.CatToHash = map[string]string{}
	c.HashToCat = map[string]string{}
	err = readAndUnmarshal(directory+"config.yaml", &(c.ConfigYaml))
	if err != nil {
		return nil, err
	}

	// Load catalog into struct
	// TODO: Unnecessary slice, remove
	var hashedNames []string
	for _, x := range c.ConfigYaml.Catalogs {
		sha := oneLineHash(x)
		c.CatToHash[x] = sha
		c.HashToCat[sha] = x

		hashedNames = append(hashedNames, sha)
	}
	sort.Strings(hashedNames)

	catalogFileInfo, err := ioutil.ReadDir(directory + "catalogs")
	if err != nil {
		return nil, err
	}
	sort.Slice(catalogFileInfo, func(i, j int) bool {
		return catalogFileInfo[i].Name() < catalogFileInfo[j].Name()
	})

	i, j := 0, 0
	for i < len(hashedNames) && j < len(catalogFileInfo) {
		n := catalogFileInfo[j].Name()
		n = n[:len(n)-len(filepath.Ext(n))]

		fmt.Printf("Comparing %s and %s\n", hashedNames, n)

		if hashedNames[i] == n {
			i++
			j++
			continue
		} else if c.ConfigYaml.Catalogs[i] < n {
			// New catalog
			cat := c.HashToCat[hashedNames[i]]
			err := c.FetchCatalog(cat)
			if err != nil {
				fmt.Printf("Error fetching catalog %s.\n", cat)
				// fmt.Printf("Error fetching catalog %s. Removing from list.\n", cat)
				// delete(c.CatToHash, cat)
				// delete(c.HashToCat, hashedNames[i])
			}
			i++
		} else {
			// Old catalog still present
			os.Remove(directory + "catalogs/" + catalogFileInfo[j].Name())
			j++
		}
	}

	for i < len(hashedNames) {
		cat := c.HashToCat[hashedNames[i]]
		err := c.FetchCatalog(cat)
		if err != nil {
			fmt.Printf("Error fetching catalog %s.\n", cat)
			// fmt.Printf("Error fetching catalog %s. Removing from list.\n", cat)
			// delete(c.CatToHash, cat)
			// delete(c.HashToCat, hashedNames[i])
		}
		i++
	}

	for j < len(catalogFileInfo) {
		os.Remove(directory + "catalogs/" + catalogFileInfo[j].Name())
		j++
	}

	return c, nil
}

func (c *KaffineConfig) Save() error {
	data, err := yaml.Marshal(c.ConfigYaml)
	if err != nil {
		return err
	}

	file, err := os.Create(c.Directory + "config.yaml")
	if err != nil {
		return err
	}
	defer file.Close()

	file.Write(data)

	return nil
}

func (c *KaffineConfig) LoadCatalog(uri string) (err error) {
	shaStr := oneLineHash(uri)
	shapath := c.Directory + "catalogs/" + shaStr + ".yaml"

	if !fileExists(shapath) {
		return os.ErrNotExist
	}

	// TODO: Validate

	catalog := KRMFunctionCatalog{}

	err = readAndUnmarshal(shapath, catalog)
	if err != nil {
		return err
	}

	c.Catalogs[shaStr] = catalog
	return nil
}

func (c *KaffineConfig) FetchCatalog(uri string) error {
	fmt.Println("Fetching: ", uri)
	dst := c.Directory + "catalogs/" + oneLineHash(uri) + ".yaml"

	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return err
	}

	if u.Scheme == "file" {
		in, err := os.Open(u.Path)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.Create(dst)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		if err != nil {
			return err
		}
		return out.Close()
	}
	// fmt.Printf("url:%v\nscheme:%v host:%v Path:%v\n\n", u, u.Scheme, u.Host, u.Path)

	return nil
}

// func (c *KaffineConfig) LoadCatalogSHA() (err error) { }

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

	LocalConfig.Save()

	return

	rootCmd.AddCommand(version.NewVersionCommand())
	rootCmd.AddCommand(config.NewConfigCommand())

	rootErr := rootCmd.Execute()
	if rootErr != nil {
		log.Fatalf("kaffine encountered an error. %v\n", rootErr)
	}
}
