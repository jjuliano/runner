package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/jjuliano/runner/resolver"

	"github.com/charmbracelet/log"
	"github.com/kdeps/plugins/kdepexec"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var params string

func initConfig() {
	if cfgFile != "" {
		if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
			fmt.Println("Config file does not exist:", cfgFile)
			os.Exit(1)
		}
		viper.SetConfigFile(cfgFile)
		// fmt.Println("Using config file:", cfgFile)
	} else {
		viper.SetConfigName("kdeps")
		viper.AddConfigPath(".")
		// fmt.Println("Using default config settings")
	}
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		resolver.PrintError("Error reading config file", err)
		os.Exit(1)
	}

	if params != "" {
		paramList := strings.Split(params, ";")
		for i, param := range paramList {
			param = strings.TrimSpace(param) // Trim spaces around each param
			envVar := fmt.Sprintf("RUNNER_PARAMS%d", i+1)
			if err := os.Setenv(envVar, param); err != nil {
				fmt.Printf("Error setting %s: %v\n", envVar, err)
				os.Exit(1)
			}
		}
	}

	// else {
	// 	fmt.Println("Successfully read config file:", viper.ConfigFileUsed())
	// }

	// Debugging line: Print the entire Viper configuration
	// fmt.Printf("Viper configuration: %+v\n", viper.AllSettings())

	// Additional debugging to check the presence of workflows
	// if !viper.IsSet("workflows") {
	// 	fmt.Println("Workflows configuration found")
	// 	fmt.Printf("Workflows: %+v\n", viper.Get("workflows"))
	// } else {
	// 	fmt.Println("Workflows configuration NOT found")
	// }
}

func initLogger() *log.Logger {
	logger := resolver.GetLogger()
	logger.Helper()
	return logger
}

func createWorkDir() (string, error) {
	tmpDir, err := os.MkdirTemp("", "runner_workdir")
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

	if err = os.Setenv("RUNNER_ENV", envFilePath); err != nil {
		return err
	}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		key := parts[0]
		value := parts[1]

		if strings.ContainsAny(value, " \t\n\r\"'") {
			value = strconv.Quote(value)
		}

		if _, err := envFile.WriteString(fmt.Sprintf("%s=%s\n", key, value)); err != nil {
			return err
		}
	}
	return nil
}

func createRootCmd(dr *resolver.DependencyResolver) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kdeps",
		Short: "A graph-based AI orchestrator",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			initConfig()
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is runner.yaml)")
	rootCmd.PersistentFlags().StringVar(&params, "params", "", "extra parameters, semi-colon separated")

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
				resolver.LogErrorExit("Error handling rdepends command", err)
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
	// fmt.Println("Created work directory:", workDir)

	cleanup := func() {
		if err := os.RemoveAll(workDir); err != nil {
			logger.Errorf("Failed to remove work directory: %v", err)
		} else {
			fmt.Println("Cleaned up work directory:", workDir)
		}
	}
	defer cleanup()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		logger.Infof("Received signal: %v, cleaning up...", sig)
		fmt.Println("Received signal:", sig)
		cleanup()
		os.Exit(0)
	}()

	envFilePath := filepath.Join(workDir, ".runner_env")
	if err := writeEnvToFile(envFilePath); err != nil {
		logger.Fatalf("Failed to write environment variables to file: %v", err)
	}
	// fmt.Println("Wrote environment variables to file:", envFilePath)

	if err := resolver.SourceEnvFile(envFilePath); err != nil {
		logger.Fatalf("Failed to source environment file: %v", err)
	}
	fmt.Println("Sourced environment file:", envFilePath)

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

	dependencyResolver, err := resolver.NewGraphResolver(afero.NewOsFs(), logger, workDir, session)
	if err != nil {
		logger.Fatalf("Failed to create dependency resolver: %v", err)
	}
	// fmt.Println("Created dependency resolver")

	resourceFiles := viper.GetStringSlice("workflows")
	// fmt.Println("Resource files from config:", resourceFiles)
	if len(resourceFiles) == 0 {
		fmt.Println("No workflows defined in the configuration file")
		os.Exit(1)
	}

	for _, file := range resourceFiles {
		fmt.Printf("Loading resource file: %s\n", file) // Debugging line
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
