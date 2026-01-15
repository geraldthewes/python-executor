package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/geraldthewes/python-executor/pkg/client"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	serverURL string
	timeout   int
	memoryMB  int
	diskMB    int
	cpuShares int
	network   bool
	image     string
	async     bool
	quiet     bool
	verbose   bool

	// run command flags
	files            []string
	entrypoint       string
	requirementsFile string
	envVars          []string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "python-executor",
		Short: "Remote Python code execution CLI",
		Long: `Execute Python code remotely in isolated containers.

Documentation:     https://github.com/geraldthewes/python-executor/blob/main/README.md
Configuration:     https://github.com/geraldthewes/python-executor/blob/main/docs/configuration.md`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", getEnv("PYEXEC_SERVER", "http://localhost:8080"), "Server URL")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 0, "Execution timeout (seconds)")
	rootCmd.PersistentFlags().IntVar(&memoryMB, "memory", 0, "Memory limit (MB)")
	rootCmd.PersistentFlags().IntVar(&diskMB, "disk", 0, "Disk limit (MB)")
	rootCmd.PersistentFlags().IntVar(&cpuShares, "cpu", 0, "CPU shares")
	rootCmd.PersistentFlags().BoolVar(&network, "network", false, "Allow network access")
	rootCmd.PersistentFlags().StringVar(&image, "image", "", "Docker image")
	rootCmd.PersistentFlags().BoolVar(&async, "async", false, "Submit async")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Quiet output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")

	// Commands
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(submitCmd())
	rootCmd.AddCommand(followCmd())
	rootCmd.AddCommand(killCmd())
	rootCmd.AddCommand(versionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [file|directory|tar] [-- script-args...]",
		Short: "Execute code synchronously",
		Long: `Execute Python code and wait for result.

Arguments after -- are passed to the Python script.
Example: python-executor run script.py -- arg1 arg2`,
		RunE: runExecution,
	}

	cmd.Flags().StringSliceVar(&files, "file", nil, "File to include (can be specified multiple times)")
	cmd.Flags().StringVar(&entrypoint, "entrypoint", "", "Entrypoint script")
	cmd.Flags().StringVar(&requirementsFile, "requirements", "", "Path to requirements.txt file")
	cmd.Flags().StringArrayVarP(&envVars, "env", "e", nil, "Environment variable (VAR or VAR=value)")

	return cmd
}

func submitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit [file|directory|tar] [-- script-args...]",
		Short: "Submit code asynchronously",
		Long: `Submit code for execution and return immediately.

Arguments after -- are passed to the Python script.
Example: python-executor submit script.py -- arg1 arg2`,
		RunE: submitExecution,
	}

	cmd.Flags().StringSliceVar(&files, "file", nil, "File to include (can be specified multiple times)")
	cmd.Flags().StringVar(&entrypoint, "entrypoint", "", "Entrypoint script")
	cmd.Flags().StringVar(&requirementsFile, "requirements", "", "Path to requirements.txt file")
	cmd.Flags().StringArrayVarP(&envVars, "env", "e", nil, "Environment variable (VAR or VAR=value)")

	return cmd
}

func followCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "follow <execution-id>",
		Short: "Follow an async execution",
		Long:  `Poll execution until complete and show result`,
		Args:  cobra.ExactArgs(1),
		RunE:  followExecution,
	}
}

func killCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kill <execution-id>",
		Short: "Kill a running execution",
		Args:  cobra.ExactArgs(1),
		RunE:  killExecution,
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("python-executor v1.0.0")
		},
	}
}

func runExecution(cmd *cobra.Command, args []string) error {
	// Separate positional args from script args
	positionalArgs, scriptArgs := splitArgsAtDash(cmd, args)

	tarData, meta, err := prepareExecution(positionalArgs, scriptArgs)
	if err != nil {
		return err
	}

	c := client.New(serverURL)
	ctx := context.Background()

	if async {
		execID, err := c.ExecuteAsync(ctx, tarData, meta)
		if err != nil {
			return err
		}
		fmt.Println(execID)
		return nil
	}

	result, err := c.ExecuteSync(ctx, tarData, meta)
	if err != nil {
		return err
	}

	printResult(result)
	os.Exit(result.ExitCode)
	return nil
}

func submitExecution(cmd *cobra.Command, args []string) error {
	// Separate positional args from script args
	positionalArgs, scriptArgs := splitArgsAtDash(cmd, args)

	tarData, meta, err := prepareExecution(positionalArgs, scriptArgs)
	if err != nil {
		return err
	}

	c := client.New(serverURL)
	ctx := context.Background()

	execID, err := c.ExecuteAsync(ctx, tarData, meta)
	if err != nil {
		return err
	}

	fmt.Println(execID)
	return nil
}

func followExecution(cmd *cobra.Command, args []string) error {
	execID := args[0]

	c := client.New(serverURL)
	ctx := context.Background()

	if !quiet {
		fmt.Fprintf(os.Stderr, "Following execution %s...\n", execID)
	}

	result, err := c.WaitForCompletion(ctx, execID, 2*time.Second)
	if err != nil {
		return err
	}

	printResult(result)
	os.Exit(result.ExitCode)
	return nil
}

func killExecution(cmd *cobra.Command, args []string) error {
	execID := args[0]

	c := client.New(serverURL)
	ctx := context.Background()

	if err := c.KillExecution(ctx, execID); err != nil {
		return err
	}

	if !quiet {
		fmt.Println("Execution killed")
	}

	return nil
}

// splitArgsAtDash separates positional args from script args at the -- separator
func splitArgsAtDash(cmd *cobra.Command, args []string) ([]string, []string) {
	dashIdx := cmd.ArgsLenAtDash()
	if dashIdx == -1 {
		return args, nil
	}
	return args[:dashIdx], args[dashIdx:]
}

// resolveEnvVars processes --env flags, resolving VAR to VAR=value from environment
func resolveEnvVars(envFlags []string) ([]string, error) {
	result := make([]string, 0, len(envFlags))

	for _, env := range envFlags {
		if strings.Contains(env, "=") {
			// Explicit value: VAR=value
			result = append(result, env)
		} else {
			// Forward from environment: VAR
			value, exists := os.LookupEnv(env)
			if !exists {
				return nil, fmt.Errorf("environment variable %q not set", env)
			}
			result = append(result, fmt.Sprintf("%s=%s", env, value))
		}
	}

	return result, nil
}

// prepareExecution creates tar and metadata from inputs
func prepareExecution(args []string, scriptArgs []string) ([]byte, *client.Metadata, error) {
	var tarData []byte
	var err error

	// Priority 1: --file flags
	if len(files) > 0 {
		tarData, err = client.TarFromFiles(files)
		if err != nil {
			return nil, nil, fmt.Errorf("creating tar from files: %w", err)
		}
	} else if len(args) == 1 {
		// Check what kind of argument it is
		arg := args[0]

		if strings.HasSuffix(arg, ".tar") {
			// Priority 2: Explicit tar file
			tarData, err = os.ReadFile(arg)
			if err != nil {
				return nil, nil, fmt.Errorf("reading tar file: %w", err)
			}
		} else {
			info, err := os.Stat(arg)
			if err != nil {
				return nil, nil, fmt.Errorf("stat %s: %w", arg, err)
			}

			if info.IsDir() {
				// Priority 3: Directory
				tarData, err = client.TarFromDirectory(arg)
				if err != nil {
					return nil, nil, fmt.Errorf("creating tar from directory: %w", err)
				}
			} else {
				// Priority 4: Single file
				tarData, err = client.TarFromFiles([]string{arg})
				if err != nil {
					return nil, nil, fmt.Errorf("creating tar from file: %w", err)
				}
			}
		}
	} else if len(args) == 0 {
		// Priority 5: Stdin
		stdinData, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, nil, fmt.Errorf("reading stdin: %w", err)
		}

		// Validate stdin is not empty
		if len(stdinData) == 0 {
			return nil, nil, fmt.Errorf("no input provided: either specify a file/directory argument or pipe code via stdin")
		}

		tarData, err = client.TarFromReader(strings.NewReader(string(stdinData)), "main.py")
		if err != nil {
			return nil, nil, fmt.Errorf("creating tar from stdin: %w", err)
		}
	} else {
		return nil, nil, fmt.Errorf("invalid arguments")
	}

	// Detect entrypoint if not specified
	if entrypoint == "" {
		entrypoint, err = client.DetectEntrypoint(tarData)
		if err != nil {
			return nil, nil, fmt.Errorf("detecting entrypoint: %w", err)
		}
	}

	// Resolve environment variables
	resolvedEnvVars, err := resolveEnvVars(envVars)
	if err != nil {
		return nil, nil, err
	}

	// Build metadata
	meta := &client.Metadata{
		Entrypoint:  entrypoint,
		DockerImage: image,
		EnvVars:     resolvedEnvVars,
		ScriptArgs:  scriptArgs,
		Config: &client.ExecutionConfig{
			TimeoutSeconds:  timeout,
			NetworkDisabled: !network,
			MemoryMB:        memoryMB,
			DiskMB:          diskMB,
			CPUShares:       cpuShares,
		},
	}

	// Read requirements file if specified
	if requirementsFile != "" {
		reqData, err := os.ReadFile(requirementsFile)
		if err != nil {
			return nil, nil, fmt.Errorf("reading requirements file: %w", err)
		}
		meta.RequirementsTxt = string(reqData)

		// Enable network access for pip install
		if !network {
			network = true
			meta.Config.NetworkDisabled = false
			if !quiet {
				fmt.Fprintln(os.Stderr, "Network access enabled for package installation")
			}
		}
	}

	return tarData, meta, nil
}

func printResult(result *client.ExecutionResult) {
	if quiet {
		if result.ExitCode == 0 {
			fmt.Print(result.Stdout)
		}
		return
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "Execution ID: %s\n", result.ExecutionID)
		fmt.Fprintf(os.Stderr, "Status: %s\n", result.Status)
		if result.DurationMs > 0 {
			fmt.Fprintf(os.Stderr, "Duration: %dms\n", result.DurationMs)
		}
		fmt.Fprintf(os.Stderr, "---\n")
	}

	if result.Stdout != "" {
		fmt.Print(result.Stdout)
	}

	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	if result.Error != "" {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.Error)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
