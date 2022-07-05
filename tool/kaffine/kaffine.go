package kaffine

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"

	"example.com/kaffine/helpers"
	"gopkg.in/yaml.v2"
)

var LocalConfig *KConfig

type KConfig struct {
	Directory   string                        `yaml:"-"`
	CatalogData map[string]KRMFunctionCatalog `yaml:"-"`
	ConfigYaml  KConfigYaml
	// Catalogs []string
}

type KConfigYaml struct {
	Catalogs []string
}

// TODO: Make into actual type
type KRMFunctionCatalog map[string]interface{}

// Directory must end with '/' character
// TODO: Change to use Go filepath
func NewKConfig(directory string) (c *KConfig, err error) {
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
	if !helpers.FileExists(directory + "config.yaml") {
		err := os.WriteFile(directory+"config.yaml", DefaultConfigYaml, 0644)
		if err != nil {
			return nil, err
		}
	}

	// Create .kaffine/installed.yaml if not exists
	if !helpers.FileExists(directory + "installed.yaml") {
		err := os.WriteFile(directory+"installed.yaml", DefaultBlankCatalog, 0644)
		if err != nil {
			return nil, err
		}
	}

	// Create the new KConfig struct
	c = new(KConfig)
	c.Directory = directory
	c.CatalogData = map[string]KRMFunctionCatalog{}
	err = helpers.ReadAndUnmarshal(directory+"config.yaml", &(c.ConfigYaml))
	if err != nil {
		return nil, err
	}

	err = c.LazySyncListToFolder()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *KConfig) SaveToYaml() error {
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

// Fetch catalog from URI and put in .kaffine/catalogs/<sha-of-uri>.yaml
func (c *KConfig) FetchCatalog(uri string) error {
	fmt.Println("Fetching: ", uri)
	dst := c.Directory + "catalogs/" + helpers.OneLineHash(uri) + ".yaml"

	u, e := url.ParseRequestURI(uri)
	if e != nil {
		return e
	}

	var data []byte
	var err error

	if u.Scheme == "file" {
		data, err = os.ReadFile(u.Path)
		if err != nil {
			return err
		}
	} else {
		resp, err := http.Get(uri)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	}

	// TODO: Validate data

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	out.Write(data)

	return nil
}

func (c *KConfig) LazySyncListToFolder() error {
	// Load catalog into struct
	sort.Slice(c.ConfigYaml.Catalogs, func(i, j int) bool {
		return helpers.OneLineHash(c.ConfigYaml.Catalogs[i]) < helpers.OneLineHash(c.ConfigYaml.Catalogs[j])
	})

	// FIXME: Inefficient, better way?
	for i := 0; i < len(c.ConfigYaml.Catalogs)-1; i++ {
		if c.ConfigYaml.Catalogs[i] == c.ConfigYaml.Catalogs[i+1] {
			c.ConfigYaml.Catalogs = append(c.ConfigYaml.Catalogs[:i], c.ConfigYaml.Catalogs[i+1:]...)
			i--
		}
	}

	catalogFileInfo, err := ioutil.ReadDir(c.Directory + "catalogs")
	if err != nil {
		return err
	}
	sort.Slice(catalogFileInfo, func(i, j int) bool {
		return catalogFileInfo[i].Name() < catalogFileInfo[j].Name()
	})

	i, j := 0, 0

	fetchOrDelete := func() {
		err := c.FetchCatalog(c.ConfigYaml.Catalogs[i])
		if err != nil {
			fmt.Printf("Error fetching catalog \"%s\". Removing from list.\n", c.ConfigYaml.Catalogs[i])
			c.ConfigYaml.Catalogs = append(c.ConfigYaml.Catalogs[:i], c.ConfigYaml.Catalogs[i+1:]...)
		} else {
			i++
		}
	}

	removeOldCat := func() {
		os.Remove(c.Directory + "catalogs/" + catalogFileInfo[j].Name())
		j++
	}

	// Two pointers
	for i < len(c.ConfigYaml.Catalogs) && j < len(catalogFileInfo) {
		n := catalogFileInfo[j].Name()
		n = n[:len(n)-len(filepath.Ext(n))]

		h := helpers.OneLineHash(c.ConfigYaml.Catalogs[i])

		fmt.Printf("Comparing %s and %s\n", h, n)

		if h == n {
			i++
			j++
			continue
		} else if h < n {
			fetchOrDelete()
		} else {
			removeOldCat()
		}
	}
	for i < len(c.ConfigYaml.Catalogs) {
		fetchOrDelete()
	}
	for j < len(catalogFileInfo) {
		removeOldCat()
	}

	return nil
}
