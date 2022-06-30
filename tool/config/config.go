package config

import (
	"fmt"

	"github.com/spf13/cobra"
)

type KaffineConfig struct {
	catalogs []string
}

func LoadConfig() error {

}

func NewConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Edit the configuration of Kaffine.",
	}

	addCatalog := &cobra.Command{
		Use:   "add-catalog",
		Short: "Adds catalog to list of managed catalogs in Kaffine",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("add-catalog!")

			return nil
		},
	}

	remCatalog := &cobra.Command{
		Use:   "remove-catalog",
		Short: "Removes catalog to list of managed catalogs in Kaffine",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("remove-catalog!")

			return nil
		},
	}

	listConfig := &cobra.Command{
		Use:   "list",
		Short: "Lists current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("list!")

			return nil
		},
	}

	cmd.AddCommand(addCatalog)
	cmd.AddCommand(remCatalog)
	cmd.AddCommand(listConfig)

	return cmd
}
