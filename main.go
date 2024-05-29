package main

import (
	"fmt"
	"os"

	"kdeps/resolver"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	command  string
	packages []string
)

func initConfig() {
	viper.SetConfigName("kdeps")
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

	rootCmd := &cobra.Command{
		Use:   "kdeps",
		Short: "A package dependency resolver",
	}

	resolver := resolver.NewDependencyResolver(afero.NewOsFs())

	packageFiles := viper.GetStringSlice("package_files")
	for _, file := range packageFiles {
		resolver.LoadPackageEntries(file) // Load package entries from each file
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "depends [package_names...]",
		Short: "List dependencies of the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleDependsCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "rdepends [package_names...]",
		Short: "List reverse dependencies of the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleRDependsCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "show [package_names...]",
		Short: "Show details of the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleShowCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "search [package_names...]",
		Short: "Search for the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleSearchCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "category [package_names...]",
		Short: "List categories of the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleCategoryCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "tree [package_names...]",
		Short: "Show dependency tree of the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleTreeCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "tree-list [package_names...]",
		Short: "Show dependency tree list of the given packages",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleTreeListCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "index",
		Short: "List all package entries",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleIndexCommand()
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
