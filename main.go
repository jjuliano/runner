package main

import (
	"fmt"
	"os"
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
	defer func(envFile *os.File) {
		err := envFile.Close()
		if err != nil {
			resolver.LogErrorExit("Error creating env file: ", err)
		}
	}(envFile)

	err = os.Setenv("KDEPS_ENV", envFilePath)
	if err != nil {
		return err
	}

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
			err := dr.HandleDependsCommand(args)
			if err != nil {
				resolver.LogErrorExit("Error handling depends command", err)
			}
		},
	}
}

func createRDependsCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "rdepends [resource_names...]",
		Short: "List reverse dependencies of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			err := dr.HandleRDependsCommand(args)
			if err != nil {
				resolver.LogErrorExit("Error handling depends command", err)
			}
		},
	}
}

func createShowCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "show [resource_names...]",
		Short: "Show details of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			err := dr.HandleShowCommand(args)
			if err != nil {
				resolver.LogErrorExit("Error handling show command", err)
			}
		},
	}
}

func createSearchCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "search [resource_names...]",
		Short: "Search for the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			err := dr.HandleSearchCommand(args)
			if err != nil {
				resolver.LogErrorExit("Error handling search command", err)
			}
		},
	}
}

func createCategoryCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "category [resource_names...]",
		Short: "List categories of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			err := dr.HandleCategoryCommand(args)
			if err != nil {
				resolver.LogErrorExit("Error handling category command", err)
			}
		},
	}
}

func createTreeCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "tree [resource_names...]",
		Short: "Show dependency tree of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			err := dr.HandleTreeCommand(args)
			if err != nil {
				resolver.LogErrorExit("Error handling tree command", err)
			}
		},
	}
}

func createTreeListCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "tree-list [resource_names...]",
		Short: "Show dependency tree list of the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			err := dr.HandleTreeListCommand(args)
			if err != nil {
				resolver.LogErrorExit("Error handling tree list command", err)
			}
		},
	}
}

func createIndexCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "index",
		Short: "List all resource entries",
		Run: func(cmd *cobra.Command, args []string) {
			err := dr.HandleIndexCommand()
			if err != nil {
				resolver.LogErrorExit("Error handling index command", err)
			}
		},
	}
}

func createRunCmd(dr *resolver.DependencyResolver) *cobra.Command {
	return &cobra.Command{
		Use:   "run [resource_names...]",
		Short: "Run the commands for the given resources",
		Run: func(cmd *cobra.Command, args []string) {
			err := dr.HandleRunCommand(args)
			if err != nil {
				resolver.LogErrorExit("Error handling run command", err)
			}
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
	defer func(session *kdepexec.ShellSession) {
		err := session.Close()
		if err != nil {
			resolver.LogErrorExit("Error closing session", err)
		}
	}(session)

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
