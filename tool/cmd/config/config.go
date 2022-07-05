package config

import (
	"errors"
	"fmt"

	"example.com/kaffine/kaffine"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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
			err := kaffine.LocalConfig.FetchCatalog(uri)
			if err != nil {
				return err
			}

			kaffine.LocalConfig.Catalogs = append(kaffine.LocalConfig.Catalogs, uri)

			fmt.Printf("Successfully added catalog\"%s\"", uri)

			return nil
		},
	}

	remCatalog := &cobra.Command{
		Use:   "remove-catalog [catalog uri]",
		Short: "Removes catalog to list of managed catalogs in Kaffine",
		RunE: func(cmd *cobra.Command, args []string) error {
			uri := args[len(args)-1]
			for i, x := range kaffine.LocalConfig.Catalogs {
				if x == uri {
					kaffine.LocalConfig.Catalogs = append(kaffine.LocalConfig.Catalogs[:i], kaffine.LocalConfig.Catalogs[i+1:]...)
					fmt.Printf("Successfully removed catalog\"%s\"", uri)

					return nil
				}
			}

			return errors.New("catalog with URI not found")
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
