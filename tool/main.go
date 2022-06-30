package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"example.com/kaffine/config"
	"example.com/kaffine/version"
	"gopkg.in/yaml.v2"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func validateSpec(fpath string) (document *loads.Document, err error) {
	document, err = loads.Spec(fpath)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to load spec")
	}

	document, err = document.Expanded(&spec.ExpandOptions{RelativeBase: fpath})
	if err != nil {
		return nil, errors.Wrap(err, "Failed to expand spec")
	}

	err = validate.Spec(document, strfmt.Default)
	if err != nil {
		return nil, errors.Wrap(err, "Spec is invalid")
	}

	return
}

var catalogSchema *spec.Schema

var rootCmd = &cobra.Command{
	Use:   "kaffine",
	Short: "Kaffine is a KRM Function Manager",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello!")
	},
}

func main() {
	catalogSchema = new(spec.Schema)

	schemaBytes, loadSchemaErr := ioutil.ReadFile("kep-2906.yaml")
	if loadSchemaErr != nil {
		log.Fatalf("error loading spec file. %v\n", loadSchemaErr)
	}

	unmarshalSchemaErr := yaml.Unmarshal(schemaBytes, &catalogSchema)
	if unmarshalSchemaErr != nil {
		log.Fatalf("error unmarshalling spec YAML. %v\n", unmarshalSchemaErr)
	}

	catalog := map[string]interface{}{}

	catalogBytes, loadCatalogErr := ioutil.ReadFile("../examples/catalogs/example-catalog.yaml")
	if loadCatalogErr != nil {
		log.Fatalf("error loading catalog file. %v\n", loadCatalogErr)
	}

	unmarshalCatalogErr := yaml.Unmarshal(catalogBytes, &catalog)
	if unmarshalCatalogErr != nil {
		log.Fatalf("error unmarshalling catalog YAML. %v\n", unmarshalCatalogErr)
	}

	validateErr := validate.AgainstSchema(catalogSchema, catalog, strfmt.Default)
	if validateErr != nil {
		log.Fatalf("error validating catalog. %v\n", validateErr)
	}
	/* else {
		fmt.Printf("%v", catalog)
	} */

	rootCmd.AddCommand(version.NewVersionCommand())
	rootCmd.AddCommand(config.NewConfigCommand())

	rootErr := rootCmd.Execute()
	if rootErr != nil {
		log.Fatalf("kaffine encountered an error. %v\n", rootErr)
	}
}
