package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"kdeps/resolver"

	"github.com/charmbracelet/log"
	"github.com/kdeps/plugins/kdepexec"
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
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		resolver.PrintError("Error reading config file", err)
		os.Exit(1)
	}
}

func initLogger() *log.Logger {
	logger := resolver.GetLogger()
	logger.Helper()
	return logger
}

func createWorkDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "kdeps_workdir")
	if err != nil {
		return "", err
	}
	return tmpDir, nil
}

func writeEnvToFile(envFilePath string) error {
	envFile, err := os.Create(envFilePath)
	if err != nil {
		return err
	}
	defer envFile.Close()

	os.Setenv("KDEPS_ENV", envFilePath)

	for _, env := range os.Environ() {
		// Split the environment variable into key and value
		parts := strings.SplitN(env, "=", 2)
		key := parts[0]
		value := parts[1]

		// Quote the value if it contains special characters or spaces
		if strings.ContainsAny(value, " \t\n\r\"'") {
			value = strconv.Quote(value)
		}

		// Write the environment variable to the file
		if _, err := envFile.WriteString(fmt.Sprintf("%s=%s\n", key, value)); err != nil {
			return err
		}
	}
	return nil
}

func sourceEnvFile(envFilePath string) error {
	cmd := exec.Command("sh", "-c", "set -a && source "+envFilePath+" && set +a")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createRootCmd(dr *resolver.DependencyResolver) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kdeps",
		Short: "A resource dependency resolver",
	}

	rootCmd.AddCommand(createDependsCmd(dr))
	rootCmd.AddCommand(createRDependsCmd(dr))
	rootCmd.AddCommand(createShowCmd(dr))
	rootCmd.AddCommand(createSearchCmd(dr))
	rootCmd.AddCommand(createCategoryCmd(dr))
	rootCmd.AddCommand(createTreeCmd(dr))
	rootCmd.AddCommand(createTreeListCmd(dr))
	rootCmd.AddCommand(createIndexCmd(dr))
	rootCmd.AddCommand(createRunCmd(dr))

	return rootCmd
}

func createDependsCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "depends [resource_names...]",
		Short: "List dependencies of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dr.HandleDependsCommand(args)
		},
	}
}

func createRDependsCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "rdepends [resource_names...]",
		Short: "List reverse dependencies of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dr.HandleRDependsCommand(args)
		},
	}
}

func createShowCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "show [resource_names...]",
		Short: "Show details of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dr.HandleShowCommand(args)
		},
	}
}

func createSearchCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "search [resource_names...]",
		Short: "Search for the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dr.HandleSearchCommand(args)
		},
	}
}

func createCategoryCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "category [resource_names...]",
		Short: "List categories of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dr.HandleCategoryCommand(args)
		},
	}
}

func createTreeCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "tree [resource_names...]",
		Short: "Show dependency tree of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dr.HandleTreeCommand(args)
		},
	}
}

func createTreeListCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "tree-list [resource_names...]",
		Short: "Show dependency tree list of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dr.HandleTreeListCommand(args)
		},
	}
}

func createIndexCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "List all resource entries",
		Run: func(cmd *cobra.Command, args []string) {
			dr.HandleIndexCommand()
		},
	}
}

func createRunCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "run [resource_names...]",
		Short: "Run the commands for the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			dr.HandleRunCommand(args)
		},
	}
}

func main() {
	initConfig()

	logger := initLogger()

	workDir, err := createWorkDir()
	if err != nil {
		logger.Fatalf("Failed to create work directory: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(workDir); err != nil {
			logger.Errorf("Failed to remove work directory: %v", err)
		} else {
			logger.Infof("Cleaned up work directory: %s", workDir)
		}
	}
	defer cleanup()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		logger.Infof("Received signal: %v, cleaning up...", sig)
		cleanup()
		os.Exit(0)
	}()

	envFilePath := filepath.Join(workDir, ".kdeps_env")
	if err := writeEnvToFile(envFilePath); err != nil {
		logger.Fatalf("Failed to write environment variables to file: %v", err)
	}

	if err := resolver.SourceEnvFile(envFilePath); err != nil {
		logger.Fatalf("Failed to source environment file: %v", err)
	}

	session, err := kdepexec.NewShellSession()
	if err != nil {
		logger.Fatalf("Failed to create shell session: %v", err)
	}
	defer session.Close()

	dependencyResolver, err := resolver.NewDependencyResolver(afero.NewOsFs(), logger, workDir, session)
	if err != nil {
		logger.Fatalf("Failed to create dependency resolver: %v", err)
	}

	resourceFiles := viper.GetStringSlice("resource_files")
	for _, file := range resourceFiles {
		if err := dependencyResolver.LoadResourceEntries(file); err != nil {
			resolver.PrintError("Error loading resource entries", err)
			os.Exit(1)
		}
	}

	rootCmd := createRootCmd(dependencyResolver)

	if err := rootCmd.Execute(); err != nil {
		resolver.PrintMessage("%v\n", err)
		os.Exit(1)
	}
}
