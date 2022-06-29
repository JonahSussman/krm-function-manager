package main

import (
	"fmt"

	"example.com/kaffine/cmd"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/pkg/errors"
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

	if err := validate.Spec(document, strfmt.Default); err != nil {
		return nil, errors.Wrap(err, "Spec is invalid")
	}

	return
}

func main() {
	_, e1 := validateSpec("kep-2906.yaml")

	if e1 != nil {
		fmt.Println(e1)
	}

	cmd.Execute()
}
