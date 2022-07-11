package main

import (
	"log"

	"example.com/kaffine/cmd/config"
	"example.com/kaffine/cmd/install"
	"example.com/kaffine/cmd/list"
	"example.com/kaffine/cmd/remove"
	"example.com/kaffine/cmd/search"
	"example.com/kaffine/cmd/update"
	"example.com/kaffine/cmd/version"
	"example.com/kaffine/kaffine"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kaffine",
	Short: "Kaffine is a KRM Function Manager",
}

func main() {
	var err error

	kaffine.LocalConfig, err = kaffine.NewKConfig("./.kaffine/")
	defer kaffine.LocalConfig.Save()

	if err != nil {
		log.Fatalf("error loading config. %v\n", err)
	}

	if kaffine.LocalConfig == nil {
		log.Fatalf("somehow LocalConfig is nil!\n")
	}

	rootCmd.AddCommand(version.NewVersionCommand())
	rootCmd.AddCommand(config.NewConfigCommand())
	rootCmd.AddCommand(list.NewListCommand())
	rootCmd.AddCommand(search.NewSearchCommand())
	rootCmd.AddCommand(install.NewInstallCommand())
	rootCmd.AddCommand(remove.NewRemoveCommand())
	rootCmd.AddCommand(update.NewUpdateCommand())

	rootErr := rootCmd.Execute()
	if rootErr != nil {
		log.Fatalf("kaffine encountered an error. %v\n", rootErr)
	}
}
