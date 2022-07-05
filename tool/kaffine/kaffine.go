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
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"gopkg.in/yaml.v2"
)

var LocalConfig *KConfig

type KConfig struct {
	Directory   string                        `yaml:"-"`
	CatalogData map[string]KRMFunctionCatalog `yaml:"-"`

	Catalogs []string
}

// TODO: Make into actual type
type KRMFunctionCatalog map[string]interface{}

// FIXME: Does not currently work...
func ValidateCatalog(catalog interface{}) error {
	return validate.AgainstSchema(CatalogSchema, catalog, strfmt.Default)
}

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
	err = helpers.ReadAndUnmarshal(directory+"config.yaml", c)
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
	sort.Strings(c.Catalogs)

	data, err := yaml.Marshal(c)
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
	fmt.Println("Fetching:", uri)
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

	catalog := map[string]interface{}{}

	err = yaml.Unmarshal(data, catalog)
	if err != nil {
		return err
	}

	err = validate.AgainstSchema(CatalogSchema, catalog, strfmt.Default)
	if err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	c.CatalogData[uri] = catalog
	out.Write(data)

	return nil
}

func (c *KConfig) LazySyncListToFolder() error {
	// Load catalog into struct
	sort.Slice(c.Catalogs, func(i, j int) bool {
		return helpers.OneLineHash(c.Catalogs[i]) < helpers.OneLineHash(c.Catalogs[j])
	})

	// FIXME: Inefficient, better way?
	for i := 0; i < len(c.Catalogs)-1; i++ {
		if c.Catalogs[i] == c.Catalogs[i+1] {
			c.Catalogs = append(c.Catalogs[:i], c.Catalogs[i+1:]...)
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
		err := c.FetchCatalog(c.Catalogs[i])
		if err != nil {
			fmt.Printf("%v. Error fetching catalog.\nRemoving \"%s\" from catalog list.\n", err, c.Catalogs[i])
			c.Catalogs = append(c.Catalogs[:i], c.Catalogs[i+1:]...)
		} else {
			i++
		}
	}

	removeOldCat := func() {
		os.Remove(c.Directory + "catalogs/" + catalogFileInfo[j].Name())
		j++
	}

	// Two pointers
	for i < len(c.Catalogs) && j < len(catalogFileInfo) {
		n := catalogFileInfo[j].Name()
		n = n[:len(n)-len(filepath.Ext(n))]

		h := helpers.OneLineHash(c.Catalogs[i])

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
	for i < len(c.Catalogs) {
		fetchOrDelete()
	}
	for j < len(catalogFileInfo) {
		removeOldCat()
	}

	return nil
}

func (c *KConfig) SearchForFunction(fname string) error {

}
