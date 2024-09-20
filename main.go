package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/charmbracelet/log"
	"github.com/jjuliano/runner/pkg/kdepexec"
	"github.com/jjuliano/runner/pkg/resolver"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	params  string
)

func initConfig() {
	log.Info("Initializing configuration...")

	viper.SetConfigName("runner")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Workflow file 'runner.yaml' not found in the current directory.")
	}

	if params != "" {
		setRunnerParams(params)
	}
}

func setRunnerParams(params string) {
	for i, param := range strings.Split(params, ";") {
		envVar := fmt.Sprintf("RUNNER_PARAMS%d", i+1)
		if err := os.Setenv(envVar, strings.TrimSpace(param)); err != nil {
			log.Fatalf("Error setting %s: %v", envVar, err)
		}
	}
}

func initLogger() *log.Logger {
	logger := resolver.GetLogger()
	logger.Helper()
	return logger
}

func createWorkDir() string {
	tmpDir, err := os.MkdirTemp("", "runner_workdir")
	if err != nil {
		log.Fatalf("Failed to create work directory: %v", err)
	}
	return tmpDir
}

func writeEnvToFile(envFilePath string) error {
	envFile, err := os.Create(envFilePath)
	if err != nil {
		return fmt.Errorf("error creating env file: %w", err)
	}
	defer envFile.Close()

	if err = os.Setenv("RUNNER_ENV", envFilePath); err != nil {
		return err
	}

	for _, env := range os.Environ() {
		keyValue := strings.SplitN(env, "=", 2)
		key, value := keyValue[0], keyValue[1]

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
		Use:   "runner",
		Short: "A simple graph-based orchestrated runner",
	}
	rootCmd.PersistentFlags().StringVar(&params, "params", "", "extra parameters, semi-colon separated")

	addCommands(rootCmd, dr)

	return rootCmd
}

func addCommands(rootCmd *cobra.Command, dr *resolver.DependencyResolver) {
	commands := []struct {
		use       string
		shortDesc string
		handler   func(*resolver.DependencyResolver, []string) error
	}{
		{"depends", "List dependencies of the given resources", func(dr *resolver.DependencyResolver, args []string) error { return dr.HandleDependsCommand(args) }},
		{"rdepends", "List reverse dependencies of the given resources", func(dr *resolver.DependencyResolver, args []string) error { return dr.HandleRDependsCommand(args) }},
		{"show", "Show details of the given resources", func(dr *resolver.DependencyResolver, args []string) error { return dr.HandleShowCommand(args) }},
		{"search", "Search for the given resources", func(dr *resolver.DependencyResolver, args []string) error { return dr.HandleSearchCommand(args) }},
		{"category", "List categories of the given resources", func(dr *resolver.DependencyResolver, args []string) error { return dr.HandleCategoryCommand(args) }},
		{"tree", "Show dependency tree of the given resources", func(dr *resolver.DependencyResolver, args []string) error { return dr.HandleTreeCommand(args) }},
		{"tree-list", "Show dependency tree list of the given resources", func(dr *resolver.DependencyResolver, args []string) error { return dr.HandleTreeListCommand(args) }},
		{"index", "List all resource entries", func(dr *resolver.DependencyResolver, _ []string) error { return dr.HandleIndexCommand() }}, // Ignoring args here
		{"run", "Run the commands for the given resources", func(dr *resolver.DependencyResolver, args []string) error { return dr.HandleRunCommand(args) }},
	}

	for _, cmd := range commands {
		cmd := cmd // Capture the loop variable
		rootCmd.AddCommand(&cobra.Command{
			Use:   cmd.use,
			Short: cmd.shortDesc,
			RunE: func(c *cobra.Command, args []string) error {
				return cmd.handler(dr, args)
			},
		})
	}
}

func handleCommand(fn func([]string) error, args []string) {
	if err := fn(args); err != nil {
		resolver.LogErrorExit("Command execution failed", err)
	}
}

func main() {
	initConfig()

	logger := initLogger()

	workDir := createWorkDir()
	defer func() {
		if err := os.RemoveAll(workDir); err != nil {
			logger.Errorf("Failed to remove work directory: %v", err)
		}
	}()

	signalCleanup(logger, workDir)

	envFilePath := filepath.Join(workDir, ".runner_env")
	if err := writeEnvToFile(envFilePath); err != nil {
		logger.Fatalf("Failed to write environment to file: %v", err)
	}

	if err := resolver.SourceEnvFile(envFilePath); err != nil {
		logger.Fatalf("Failed to source environment file: %v", err)
	}

	session := createShellSession(logger)
	defer session.Close()

	dependencyResolver := createDependencyResolver(logger, workDir, session)

	loadResourceFiles(dependencyResolver)

	rootCmd := createRootCmd(dependencyResolver)
	if err := rootCmd.Execute(); err != nil {
		resolver.PrintMessage("%v\n", err)
		os.Exit(1)
	}
}

func signalCleanup(logger *log.Logger, workDir string) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		logger.Infof("Received signal: %v, cleaning up...", sig)
		os.Exit(0)
	}()
}

func createShellSession(logger *log.Logger) *kdepexec.ShellSession {
	session, err := kdepexec.NewShellSession()
	if err != nil {
		logger.Fatalf("Failed to create shell session: %v", err)
	}
	return session
}

func createDependencyResolver(logger *log.Logger, workDir string, session *kdepexec.ShellSession) *resolver.DependencyResolver {
	dr, err := resolver.NewGraphResolver(afero.NewOsFs(), logger, workDir, session)
	if err != nil {
		logger.Fatalf("Failed to create dependency resolver: %v", err)
	}
	return dr
}

func loadResourceFiles(dr *resolver.DependencyResolver) {
	resourceFiles := viper.GetStringSlice("workflows")
	if len(resourceFiles) == 0 {
		log.Fatalf("No workflows defined in the configuration file")
	}

	for _, file := range resourceFiles {
		if err := dr.LoadResourceEntries(file); err != nil {
			resolver.LogErrorExit(fmt.Sprintf("Error loading resource entries from %s", file), err)
		}
	}
}
