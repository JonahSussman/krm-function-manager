package config

import (
	"fmt"

	"example.com/kaffine/kaffine"
	"github.com/spf13/cobra"
)

// support both local and global config
// global config determined by KAFFINE_GLOBAL_CONFIG env variable.
// If unset, defaults to ~/.kaffine/config
//
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

			kaffine.LocalConfig.ConfigYaml.Catalogs = append(kaffine.LocalConfig.ConfigYaml.Catalogs, args[len(args)-1])

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
