package config

import (
	"fmt"

	"example.com/kaffine/kaffine"
	"github.com/ghodss/yaml"
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
		Use:   "add-catalog [catalog uri]",
		Short: "Adds catalog to list of managed catalogs in Kaffine",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uri := args[len(args)-1]
			_, err := kaffine.LocalConfig.AddCatalog(uri)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully added catalog\"%s\"", uri)

			return nil
		},
	}

	remCatalog := &cobra.Command{
		Use:   "remove-catalog [catalog uri]",
		Short: "Removes catalog to list of managed catalogs in Kaffine",
		RunE: func(cmd *cobra.Command, args []string) error {
			uri := args[len(args)-1]
			err := kaffine.LocalConfig.RemoveCatalog(uri)
			if err != nil {
				return err
			}

			fmt.Printf("Successfully removed catalog\"%s\"", uri)

			return nil
		},
	}

	listConfig := &cobra.Command{
		Use:   "list",
		Short: "Lists current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := yaml.Marshal(kaffine.LocalConfig)
			if err != nil {
				return err
			}

			fmt.Println(string(data))

			return nil
		},
	}

	cmd.AddCommand(addCatalog)
	cmd.AddCommand(remCatalog)
	cmd.AddCommand(listConfig)

	return cmd
}
