package search

import (
	"fmt"

	"example.com/kaffine/kaffine"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

func NewSearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search [name]",
		Short: "Searches the managed catalogs for a function with the specified name",
		RunE: func(cmd *cobra.Command, args []string) error {
			fname := args[len(args)-1]
			res, err := kaffine.LocalConfig.SearchForFunction(fname)
			if err != nil {
				return err
			}

			b, err := yaml.Marshal(res)
			if err != nil {
				return err
			}

			fmt.Println(string(b))

			return nil
		},
	}

	return cmd
}
