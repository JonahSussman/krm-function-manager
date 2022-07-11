package update

import (
	"fmt"

	"example.com/kaffine/kaffine"
	"github.com/spf13/cobra"
)

func NewUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Updates all functions to their latest versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := kaffine.LocalConfig.UpdateFunctions()
			if err != nil {
				return err
			}

			fmt.Println("Successfully updated catalogs and functions")

			return nil
		},
	}

	return cmd
}
