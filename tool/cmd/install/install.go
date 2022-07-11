package install

import (
	"fmt"

	"example.com/kaffine/kaffine"
	"github.com/spf13/cobra"
)

func NewInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [name]",
		Short: "Searches the managed catalogs for a function with the specified name, and installs it",
		RunE: func(cmd *cobra.Command, args []string) error {
			fname := args[len(args)-1]
			krmFunc, err := kaffine.LocalConfig.AddFunction(fname)
			if err != nil {
				return err
			}

			fmt.Println("Successfully added KRM Function '" + krmFunc.GetShortName() + "'")

			return nil
		},
	}

	return cmd
}
