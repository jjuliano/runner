package main

import (
	"fmt"
	"os"

	"kdeps/resolver"

	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	initConfig() // Load configuration at the start

	var command string
	var packages []string

	for _, arg := range os.Args[1:] {
		switch arg {
		case "depends", "rdepends", "add", "update", "search", "category", "show", "tree", "tree-list":
			if command == "" {
				command = arg
			} else {
				packages = append(packages, arg)
			}
		default:
			packages = append(packages, arg)
		}
	}

	if command == "" {
		fmt.Println("Usage: kdeps [depends|rdepends|search|category|show|tree|tree-list] [package_names...]")
		return
	}

	resolver := resolver.NewDependencyResolver(afero.NewOsFs())
	resolver.LoadPackageEntries(viper.GetString("package_file")) // Use Viper to get the package file path

	switch command {
	case "show":
		resolver.HandleShowCommand(packages)
	case "depends":
		resolver.HandleDependsCommand(packages)
	case "rdepends":
		resolver.HandleRDependsCommand(packages)
	case "search":
		resolver.HandleSearchCommand(packages)
	case "category":
		resolver.HandleCategoryCommand(packages)
	case "tree":
		resolver.HandleTreeCommand(packages)
	case "tree-list":
		resolver.HandleTreeListCommand(packages)
	default:
		fmt.Println("Invalid command:", command)
	}
}
