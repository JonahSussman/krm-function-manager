package kaffine

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"example.com/kaffine/helpers"
	"github.com/ghodss/yaml"
	"k8s.io/utils/strings/slices"
)

var LocalConfig *KConfig

type KConfig struct {
	Directory        string                         `json:"-"`
	CatalogData      map[string]*KRMFunctionCatalog `json:"-"`
	InstalledCatalog *KRMFunctionCatalog            `json:"-"`
	ModifiedCatalog  map[string]bool                `json:"-"`

	Catalogs []string `json:"catalogs"`
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

	// Create the new KConfig struct
	c = new(KConfig)
	c.Directory = directory
	c.CatalogData = map[string]*KRMFunctionCatalog{}
	c.ModifiedCatalog = map[string]bool{}
	c.InstalledCatalog = &KRMFunctionCatalog{}

	if !helpers.FileExists(directory + "installed.yaml") {
		c.InstalledCatalog, err = NewKRMFunctionCatalog("Kaffine Output Catalog")
	} else {
		err = helpers.ReadAndUnmarshal(directory+"installed.yaml", c.InstalledCatalog)
	}

	if err != nil {
		return nil, err
	}

	err = helpers.ReadAndUnmarshal(directory+"config.yaml", c)
	if err != nil {
		return nil, err
	}

	// Ensure only unique values in catalogs array
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

	return c, nil
}

func (c *KConfig) Save() error {
	// Sort on hash
	sort.Slice(c.Catalogs, func(i, j int) bool {
		return helpers.OneLineHash(c.Catalogs[i]) < helpers.OneLineHash(c.Catalogs[j])
	})

	// Ensure uniqueness
	for i := 0; i < len(c.Catalogs)-1; i++ {
		if c.Catalogs[i] == c.Catalogs[i+1] {
			c.Catalogs = append(c.Catalogs[:i], c.Catalogs[i+1:]...)
			i--
		}
	}

	// Get all cached catalogs
	catalogFileInfo, err := ioutil.ReadDir(c.Directory + "catalogs")
	if err != nil {
		return err
	}
	sort.Slice(catalogFileInfo, func(i, j int) bool {
		return catalogFileInfo[i].Name() < catalogFileInfo[j].Name()
	})

	i, j := 0, 0

	cacheModifiedCat := func() {
		uri := c.Catalogs[i]
		dst := c.Directory + "catalogs/" + helpers.OneLineHash(uri) + ".yaml"

		helpers.MarshalAndWrite(dst, c.CatalogData[uri])
	}

	removeOldCat := func() {
		os.Remove(c.Directory + "catalogs/" + catalogFileInfo[j].Name())
	}

	// Two pointers
	for i < len(c.Catalogs) && j < len(catalogFileInfo) {
		n := catalogFileInfo[j].Name()
		n = n[:len(n)-len(filepath.Ext(n))]

		h := helpers.OneLineHash(c.Catalogs[i])

		if h == n {
			if c.ModifiedCatalog[c.Catalogs[i]] {
				cacheModifiedCat()
			}

			i++
			j++

			continue
		} else if h < n {
			if c.ModifiedCatalog[c.Catalogs[i]] {
				cacheModifiedCat()
			}

			i++
		} else { // h > n
			removeOldCat()
			j++
		}
	}
	for i < len(c.Catalogs) {
		if c.ModifiedCatalog[c.Catalogs[i]] {
			cacheModifiedCat()
		}
		i++
	}
	for j < len(catalogFileInfo) {
		removeOldCat()
		j++
	}

	err = helpers.MarshalAndWrite(c.Directory+"config.yaml", c)
	if err != nil {
		return err
	}

	err = helpers.MarshalAndWrite(c.Directory+"installed.yaml", c.InstalledCatalog)
	if err != nil {
		return err
	}

	return nil
}

func (c *KConfig) GetCatalog(uri string) (*KRMFunctionCatalog, error) {
	// Already loaded into memory
	if cat, ok := c.CatalogData[uri]; ok {
		return cat, nil
	}

	// Cached in catalogs directory
	catalogFileInfo, err := ioutil.ReadDir(c.Directory + "catalogs")
	if err != nil {
		return nil, err
	}
	hashedFilename := helpers.OneLineHash(uri) + ".yaml"
	for _, x := range catalogFileInfo {
		if x.Name() != hashedFilename {
			continue
		}

		cat := KRMFunctionCatalog{}
		err := helpers.ReadAndUnmarshal(c.Directory+"catalogs/"+x.Name(), &cat)
		if err != nil {
			return nil, err
		}
		c.CatalogData[uri] = &cat
		return &cat, nil
	}

	// Not yet downloaded
	if slices.Contains(c.Catalogs, uri) {
		cat, err := FetchCatalog(uri)
		if err != nil {
			return nil, err
		}
		c.CatalogData[uri] = cat
		c.ModifiedCatalog[uri] = true
		return cat, nil
	}

	// TODO: Check parent

	// Does not exist
	return nil, errors.New("uri not in managed catalogs")
}

// Fetch catalog from URI
func FetchCatalog(uri string) (*KRMFunctionCatalog, error) {
	// fmt.Fprintf(os.Stderr, "Fetching: %s", uri)

	u, e := url.ParseRequestURI(uri)
	if e != nil {
		return nil, e
	}

	var data []byte
	var err error

	if u.Scheme == "file" {
		data, err = os.ReadFile(u.Path)
	} else {
		var resp *http.Response
		resp, err = http.Get(uri)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
	}

	if err != nil {
		return nil, err
	}

	catalog := KRMFunctionCatalog{}

	err = yaml.Unmarshal(data, &catalog)
	if err != nil {
		return nil, err
	}

	// err = validate.AgainstSchema(CatalogSchema, catalog, strfmt.Default)
	// if err != nil {
	// 	return err
	// }

	return &catalog, nil
}

func (c *KConfig) AddCatalog(uri string) (*KRMFunctionCatalog, error) {
	cat, err := FetchCatalog(uri)
	if err != nil {
		return nil, err
	}
	c.Catalogs = append(c.Catalogs, uri)
	c.CatalogData[uri] = cat
	c.ModifiedCatalog[uri] = true
	return cat, nil
}

func (c *KConfig) RemoveCatalog(uri string) error {
	for i, x := range c.Catalogs {
		if uri != x {
			continue
		}

		c.Catalogs = append(c.Catalogs[:i], c.Catalogs[i+1:]...)
		delete(c.CatalogData, uri)
		// c.ModifiedCatalog[uri] = true
		return nil
	}

	return errors.New("catalog not in list of managed catalogs")
}

// TODO: Add function name formatting <group>/<name>:<version>
func (c *KConfig) SearchForFunction(fname string) (*KRMFunctionCatalog, error) {
	cat, err := NewKRMFunctionCatalog("Search results for '" + fname + "'")
	if err != nil {
		return nil, err
	}

	lowerFname := strings.ToLower(fname)

	for _, searchCatName := range LocalConfig.Catalogs {
		searchCat, _ := LocalConfig.GetCatalog(searchCatName)
		for _, searchFunc := range searchCat.Spec.KrmFunctions {
			if strings.Contains(strings.ToLower(searchFunc.Names.Kind), lowerFname) {
				cat.Spec.KrmFunctions = append(cat.Spec.KrmFunctions, searchFunc)
			}
		}
	}

	return cat, nil
}

func (c *KConfig) AddFunction(fname string) (*KRMFunctionDefinitionSpec, error) {
	cat, err := c.SearchForFunction(fname)
	if err != nil {
		return nil, err
	}

	if len(cat.Spec.KrmFunctions) == 0 {
		return nil, errors.New("cannot add catalog '" + fname + "' - no results found")
	}

	if len(cat.Spec.KrmFunctions) > 1 {
		return nil, errors.New("cannot add catalog '" + fname + "' - more than 1 result found")
	}

	krmFunc := cat.Spec.KrmFunctions[0]

	for _, x := range LocalConfig.InstalledCatalog.Spec.KrmFunctions {
		if x.Names.Kind != krmFunc.Names.Kind {
			continue
		}

		if x.Group != krmFunc.Group {
			continue
		}

		return nil, errors.New("KRM Function with same group and name already installed")
	}

	krmFunc.Versions = []*KRMFunctionVersion{krmFunc.GetHighestVersion()}
	// FIXME: Make a deep copy!
	LocalConfig.InstalledCatalog.Spec.KrmFunctions = append(LocalConfig.InstalledCatalog.Spec.KrmFunctions, krmFunc)

	// fmt.Println(LocalConfig.InstalledCatalog.Spec.KrmFunctions)

	return krmFunc, nil
}

func (c *KConfig) RemoveFunction(fname string) (*KRMFunctionDefinitionSpec, error) {
	var funcs []*KRMFunctionDefinitionSpec
	var idxs []int

	lowerFname := strings.ToLower(fname)

	for i, searchFunc := range LocalConfig.InstalledCatalog.Spec.KrmFunctions {
		if strings.Contains(strings.ToLower(searchFunc.Names.Kind), lowerFname) {
			funcs = append(funcs, searchFunc)
			idxs = append(idxs, i)
		}
	}

	if len(funcs) == 0 {
		return nil, errors.New("cannot remove catalog '" + fname + "' - no results found")
	}

	if len(funcs) > 1 {
		return nil, errors.New("cannot remove catalog '" + fname + "' - more than 1 result found")
	}

	LocalConfig.InstalledCatalog.Spec.KrmFunctions = append(LocalConfig.InstalledCatalog.Spec.KrmFunctions[:idxs[0]], LocalConfig.InstalledCatalog.Spec.KrmFunctions[idxs[0]+1:]...)

	return funcs[0], nil
}

func (c *KConfig) UpdateFunctions() error {
	// Make a map with pointer values OR name string values
	m := map[string]*KRMFunctionDefinitionSpec{}
	for _, x := range LocalConfig.InstalledCatalog.Spec.KrmFunctions {
		m[x.Group+"/"+x.Names.Kind] = x
	}
	// Fetch new catalog
	for _, uri := range LocalConfig.Catalogs {
		cat, err := FetchCatalog(uri)
		if err != nil {
			return err
		}

		// Iterate through, check if any match up
		for _, krmFunc := range cat.Spec.KrmFunctions {
			if y := krmFunc.Group + "/" + krmFunc.Names.Kind; m[y] != nil {
				// Update those that match up
				m[y].Versions = []*KRMFunctionVersion{krmFunc.GetHighestVersion()}
			}
		}

		LocalConfig.CatalogData[uri] = cat
		LocalConfig.ModifiedCatalog[uri] = true
	}

	return nil
}
