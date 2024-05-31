package main

import (
	"fmt"
	"os"

	"kdeps/resolver"

	"github.com/charmbracelet/log"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	command  string
	resources []string
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
	logger := log.Default()
	logger.Helper()

	initConfig() // Load configuration at the start

	rootCmd := &cobra.Command{
		Use:   "kdeps",
		Short: "A resource dependency resolver",
	}

	resolver := resolver.NewDependencyResolver(afero.NewOsFs(), logger)

	resourceFiles := viper.GetStringSlice("resource_files")
	for _, file := range resourceFiles {
		resolver.LoadResourceEntries(file) // Load resource entries from each file
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "depends [resource_names...]",
		Short: "List dependencies of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleDependsCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "rdepends [resource_names...]",
		Short: "List reverse dependencies of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleRDependsCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "show [resource_names...]",
		Short: "Show details of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleShowCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "search [resource_names...]",
		Short: "Search for the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleSearchCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "category [resource_names...]",
		Short: "List categories of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleCategoryCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "tree [resource_names...]",
		Short: "Show dependency tree of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleTreeCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "tree-list [resource_names...]",
		Short: "Show dependency tree list of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleTreeListCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "index",
		Short: "List all resource entries",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleIndexCommand()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "run [resource_names...]",
		Short: "Run the commands for the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			resolver.HandleRunCommand(args)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
