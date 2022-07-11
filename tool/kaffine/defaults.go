package kaffine

import (
	_ "embed"
)

//go:embed config.yaml
var DefaultConfigYaml []byte

// var catalogSpecFS embed.FS
// //go:embed kep-2906.v3.yaml
// var CatalogSpec []byte
// var CatalogSchema *spec.Schema

//go:embed catalog.yaml
var DefaultBlankCatalog []byte

func init() {
	// var err error

	// spec validation
	// CatalogSpec, err = catalogSpecFS.ReadFile("kep-2906.yaml")
	// if err != nil {
	// 	panic(errors.Wrap(err, "error loading kep-2906 embedded file."))
	// }

	// var document *loads.Document
	// document, err = document.Expanded(&spec.ExpandOptions{RelativeBase: "kep-2906.yaml"})
	// if err != nil {
	// 	panic(errors.Wrap(err, "Failed to expand spec"))
	// }

	// err = validate.Spec(document, strfmt.Default)
	// if err != nil {
	// 	panic(errors.Wrap(err, "Spec is invalid"))
	// }

	// CatalogSchema = new(spec.Schema)
	// err = yaml.Unmarshal(CatalogSpec, CatalogSchema)
	// if err != nil {
	// 	panic(errors.Wrap(err, "error unmarshalling catalog YAML."))
	// }
}
