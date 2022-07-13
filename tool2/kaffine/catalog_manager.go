package kaffine

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

type CatalogManager struct {
	Directory string
	Catalogs  map[string]FunctionCatalog
	Functions map[string]FunctionDefinition
}

func MakeCatalogManager(directory string) CatalogManager {
	cm := CatalogManager{}
	cm.Directory = filepath.Clean(filepath.Join(directory, "/catalogs"))
	cm.Catalogs = map[string]FunctionCatalog{}
	cm.Functions = map[string]FunctionDefinition{}

	os.MkdirAll(cm.Directory, os.ModePerm)

	return cm
}

func (cm *CatalogManager) Save() (err error) {
	os.RemoveAll(cm.Directory)
	os.MkdirAll(cm.Directory, os.ModePerm)

	for uri, cat := range cm.Catalogs {
		b, err := yaml.Marshal(cat)
		if err != nil {
			return err
		}

		hashName := SHA1(uri) + ".yaml"
		err = os.WriteFile(filepath.Join(cm.Directory, hashName), b, os.ModePerm)
		if err != nil {
			return err
		}
	}

	return
}

// Tries to look in cache first
func (cm *CatalogManager) AddCatalog(uri string) (err error) {
	// Already added
	if _, ok := cm.Catalogs[uri]; ok {
		return errors.New("catalog already present")
	}

	// Cached on the filesystem
	cat := FunctionCatalog{}

	cat, err = cm.GetCachedCatalog(uri)
	if err != nil {
		cat, err = cm.GetExternalCatalog(uri)

		// Fetch externally
		if err != nil {
			return err
		}
	}

	// Check for conflicting names
	for _, fn := range cat.Spec.KrmFunctions {
		if _, ok := cm.Functions[fn.GroupName()]; ok {
			return errors.New("attempting to add catalog that contains conflicting names")
		}
	}

	for _, fn := range cat.Spec.KrmFunctions {
		cm.Functions[fn.GroupName()] = fn
	}
	cm.Catalogs[uri] = cat

	return nil
}

func (cm *CatalogManager) GetCachedCatalog(uri string) (fc FunctionCatalog, err error) {
	catalogFileInfo, err := os.ReadDir(cm.Directory)
	if err != nil {
		return
	}

	hashedFilename := SHA1(uri) + ".yaml"
	ok := false
	for _, x := range catalogFileInfo {
		if x.Name() == hashedFilename {
			ok = true
			break
		}
	}

	if !ok {
		return fc, fmt.Errorf("cached catalog '%s' (hash '%s') not present in filesystem", uri, hashedFilename)
	}

	data, err := os.ReadFile(filepath.Join(cm.Directory, hashedFilename))
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &fc)
	if err != nil {
		return
	}

	return
}

func (cm *CatalogManager) GetExternalCatalog(uri string) (fc FunctionCatalog, err error) {
	u, err := url.ParseRequestURI(uri)
	if err != nil {
		return
	}

	var data []byte

	if u.Scheme == "file" {
		data, err = os.ReadFile(u.Path)
	} else {
		var resp *http.Response
		resp, err = http.Get(uri)
		if err != nil {
			return
		}
		defer resp.Body.Close()
		data, err = io.ReadAll(resp.Body)
	}

	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &fc)
	if err != nil {
		return
	}

	return
}

// Removes all traces
func (cm *CatalogManager) RemoveCatalog(uri string) (oldFc FunctionCatalog, err error) {
	if _, ok := cm.Catalogs[uri]; !ok {
		return oldFc, errors.New("catalog with uri not present")
	}

	for _, fn := range cm.Catalogs[uri].Spec.KrmFunctions {
		delete(cm.Functions, fn.GroupName())
	}

	oldFc = cm.Catalogs[uri]

	delete(cm.Catalogs, uri)

	return oldFc, nil
}

// Clobbers catalog
func (cm *CatalogManager) UpdateCatalog(uri string) (oldFc FunctionCatalog, err error) {
	oldFc, err = cm.RemoveCatalog(uri)
	if err != nil {
		return
	}

	err = cm.AddCatalog(uri)
	if err != nil {
		return
	}

	return
}

func (cm *CatalogManager) UpdateAllCatalogs(uri string) (oldFcs []FunctionCatalog, errs []error) {
	for uri, _ := range cm.Catalogs {
		fc, err := cm.UpdateCatalog(uri)
		oldFcs = append(oldFcs, fc)
		errs = append(errs, err)
	}

	return
}

// use .GroupName() function
func (cm *CatalogManager) Search(fname string) (fns []FunctionDefinition, err error) {
	group, name, version := ToGroupNameVersion(fname)
	groupName := name
	if group != "" {
		groupName = group + "/" + groupName
	}

	for queryName, queryDef := range cm.Functions {
		if !strings.Contains(queryName, groupName) {
			continue
		}
		if version != "" {
			var versions []FunctionVersion
			for _, queryVersion := range queryDef.Versions {
				if queryVersion.Name == version {
					versions = append(versions, queryVersion)
				}
			}

			if len(versions) == 0 {
				continue
			}

			queryDef.Versions = versions
		}
		fns = append(fns, queryDef)
	}

	return fns, nil
}

// INPUT MUST BE SORTED
func (cm *CatalogManager) SearchMultiple(fnames []string) (fnss [][]FunctionDefinition, errs []error) {
	return
}
