package list

import (
	"fmt"

	"example.com/kaffine/kaffine"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Lists the current installed catalog of functions",
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := yaml.Marshal(kaffine.LocalConfig.InstalledCatalog)
			if err != nil {
				return err
			}

			fmt.Println(string(b))

			return nil
		},
	}

	return cmd
}
