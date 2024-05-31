package main

import (
	"log"
	"os"

	"kdeps/resolver"

	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	command   string
	resources []string
)

func initConfig() {
	viper.SetConfigName("kdeps")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		resolver.PrintError("Error reading config file", err)
		os.Exit(1)
	}
}

func main() {
	logger := resolver.GetLogger()
	logger.Helper()

	initConfig() // Load configuration at the start

	rootCmd := &cobra.Command{
		Use:   "kdeps",
		Short: "A resource dependency resolver",
	}

	dependencyResolver, err := resolver.NewDependencyResolver(afero.NewOsFs(), logger)
	if err != nil {
		log.Fatalf("Failed to create dependency resolver: %v", err)
	}

	resourceFiles := viper.GetStringSlice("resource_files")
	for _, file := range resourceFiles {
		dependencyResolver.LoadResourceEntries(file) // Load resource entries from each file
	}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "depends [resource_names...]",
		Short: "List dependencies of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dependencyResolver.HandleDependsCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "rdepends [resource_names...]",
		Short: "List reverse dependencies of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dependencyResolver.HandleRDependsCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "show [resource_names...]",
		Short: "Show details of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dependencyResolver.HandleShowCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "search [resource_names...]",
		Short: "Search for the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dependencyResolver.HandleSearchCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "category [resource_names...]",
		Short: "List categories of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dependencyResolver.HandleCategoryCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "tree [resource_names...]",
		Short: "Show dependency tree of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dependencyResolver.HandleTreeCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "tree-list [resource_names...]",
		Short: "Show dependency tree list of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dependencyResolver.HandleTreeListCommand(args)
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "index",
		Short: "List all resource entries",
		Run: func(cmd *cobra.Command, args []string) {
			dependencyResolver.HandleIndexCommand()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "run [resource_names...]",
		Short: "Run the commands for the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dependencyResolver.HandleRunCommand(args)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		resolver.PrintMessage("%v\n", err)
		os.Exit(1)
	}
}
