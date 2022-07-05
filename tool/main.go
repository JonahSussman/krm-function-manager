package main

import (
	"log"

	"example.com/kaffine/cmd/config"
	"example.com/kaffine/cmd/version"
	"example.com/kaffine/helpers"
	"example.com/kaffine/kaffine"

	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kaffine",
	Short: "Kaffine is a KRM Function Manager",
}

func main() {
	var err error

	catalog := map[string]interface{}{}

	err = helpers.ReadAndUnmarshal("../examples/catalogs/example-catalog.yaml", catalog)
	if err != nil {
		log.Fatalf("error loading catalog file. %v\n", err)
	}

	err = validate.AgainstSchema(kaffine.CatalogSchema, catalog, strfmt.Default)
	if err != nil {
		log.Fatalf("error validating catalog. %v\n", err)
	}

	kaffine.LocalConfig, err = kaffine.NewKConfig("./.kaffine/")
	defer kaffine.LocalConfig.SaveToYaml()

	if err != nil {
		log.Fatalf("error loading config. %v\n", err)
	}

	// fmt.Println(LocalConfig)

	// return

	rootCmd.AddCommand(version.NewVersionCommand())
	rootCmd.AddCommand(config.NewConfigCommand())

	rootErr := rootCmd.Execute()
	if rootErr != nil {
		log.Fatalf("kaffine encountered an error. %v\n", rootErr)
	}

	return
}
