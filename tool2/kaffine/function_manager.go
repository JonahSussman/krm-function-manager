package kaffine

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/exp/maps"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type FunctionManager struct {
	Directory string

	CatMan *CatalogManager
	Cfg    *Config

	Installed map[string]FunctionDefinition
}

func NewFunctionManager(directory string) *FunctionManager {
	fm := FunctionManager{}

	fm.Directory = directory
	catman := MakeCatalogManager(directory)
	fm.CatMan = &catman
	cfg := MakeConfig(directory)
	fm.Cfg = &cfg

	for _, uri := range fm.Cfg.Catalogs {
		fm.CatMan.AddCatalog(uri)
	}
	fm.Cfg.Catalogs = maps.Keys(fm.CatMan.Catalogs)

	fm.Installed = map[string]FunctionDefinition{}
	for _, gnv := range fm.Cfg.Dependencies.KrmFunctions {
		// FIXME: Extremely inefficient!
		fm.Install(gnv)
		// group, name, version := ToGroupNameVersion(gnv)
		// fnDir := filepath.Join(directory, "functions", group)
		// fnFile := name + ".yaml"

		// if _, err := os.Stat(filepath.Join(fnDir, fnFile)); errors.Is(err, os.ErrNotExist) {
		// 	// TODO Error handling
		// 	// FIXME This is ugly
		// 	_, err := fm.Install(gnv)
		// 	if err != nil {
		// 		continue
		// 	}
		// 	continue
		// }

		// // Load from file
		// b, _ := os.ReadFile(filepath.Join(fnDir, fnFile))
		// fd := FunctionDefinition{}
		// yaml.Unmarshal(b, &fd)
		// if len(fd.Versions) != 1 {
		// 	continue
		// }
		// if version != "" {
		// 	if fd.Versions[0].Name != version {

		// 		continue
		// 	}
		// 	fd.Metadata.Annotations["kaffine.config/ignore-auto-updates"] = "true"
		// }

		// fm.Installed[fd.GroupName()] = fd
	}

	return &fm
}

func (fm *FunctionManager) Save() error {
	fm.Cfg.Dependencies.KrmFunctions = make([]string, 0)
	os.RemoveAll(filepath.Join(fm.Directory, "functions"))

	for _, groupName := range maps.Keys(fm.Installed) {
		fd, _ := fm.SaveFunction(groupName)
		name := fd.GroupName()
		if fd.Metadata != nil {
			if val, ok := fd.Metadata.Annotations[IgnoreAutoUpdates]; ok && val == "true" {
				name = name + "@" + fd.Versions[0].Name
			}
		}
		fm.Cfg.Dependencies.KrmFunctions = append(fm.Cfg.Dependencies.KrmFunctions, name)
	}

	if installedCatalog, err := fm.GenerateInstalledCatalog(); err != nil {
		return err
	} else {
		os.WriteFile(filepath.Join(fm.Directory, "installed.yaml"), installedCatalog, os.ModePerm)
	}

	if err := fm.Cfg.Save(); err != nil {
		return err
	}
	if err := fm.CatMan.Save(); err != nil {
		return err
	}

	return nil
}

func (fm *FunctionManager) SaveFunction(groupName string) (fd FunctionDefinition, error error) {
	fd, ok := fm.Installed[groupName]
	if !ok {
		return fd, fmt.Errorf("function '%s' not installed (check spelling?)", groupName)
	}

	fnDir := filepath.Join(fm.Directory, "functions", fd.Group)
	fmt.Println(fnDir)
	fnFile := fd.Names.Kind + ".yaml"

	os.MkdirAll(fnDir, os.ModePerm)
	b, err := yaml.Marshal(fd)
	if err != nil {
		return fd, err
	}
	os.WriteFile(filepath.Join(fnDir, fnFile), b, os.ModePerm)

	return fd, nil
}

func (fm *FunctionManager) Install(fname string) (result []FunctionDefinition, err error) {
	group, name, version := ToGroupNameVersion(fname)
	groupName := group + "/" + name
	result, err = fm.CatMan.Search(groupName)
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no functions with name '%s'", groupName)
	}
	if len(result) > 1 {
		return result, fmt.Errorf("more than one function found with search term '%s'", groupName)
	}

	var resultVersion FunctionVersion

	if version == "" {
		resultVersion = result[0].GetHighestVersion()
	} else {
		resultVersion, err = result[0].GetVersion(version)
		if err != nil {
			return result, err
		}
		if result[0].Metadata == nil {
			result[0].Metadata = &v1.ObjectMeta{Annotations: map[string]string{}}
		}
		result[0].Metadata.Annotations[IgnoreAutoUpdates] = "true"
	}

	result[0].Versions = []FunctionVersion{resultVersion}
	fm.Installed[result[0].GroupName()] = result[0]

	return
}

func (fm *FunctionManager) GenerateInstalledCatalog() (result []byte, err error) {
	fc := MakeFunctionCatalog("Kaffine Managed Functions")
	for _, fn := range fm.Installed {
		fc.Spec.KrmFunctions = append(fc.Spec.KrmFunctions, fn)
	}
	return yaml.Marshal(fc)
}
