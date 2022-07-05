package main

import (
	"log"

	"example.com/kaffine/cmd/config"
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
	defer kaffine.LocalConfig.SaveToYaml()

	if err != nil {
		log.Fatalf("error loading config. %v\n", err)
	}

	rootCmd.AddCommand(version.NewVersionCommand())
	rootCmd.AddCommand(config.NewConfigCommand())

	rootCmd.Execute()
	// rootErr := rootCmd.Execute()
	// if rootErr != nil {
	// 	log.Fatalf("kaffine encountered an error. %v\n", rootErr)
	// }
}
