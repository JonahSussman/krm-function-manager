package main

import (
	"fmt"

	"kaffine-mod/kaffine"

	"sigs.k8s.io/yaml"
)

func main() {
	var err error
	kaffine.InitializeGlobals()
	fm := kaffine.NewFunctionManager(kaffine.Directory)
	defer fm.Save()

	fmt.Println("main:", kaffine.Directory)
	x, _ := fm.CatMan.Search("Logger@v1.0.1")
	for _, y := range x {
		data, _ := yaml.Marshal(y)
		fmt.Println(string(data))
	}

	err = kaffine.DestroyGlobals()
	if err != nil {
		panic(err)
	}
}
