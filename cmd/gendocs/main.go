// cmd/gendocs generates CLI documentation using cobra/doc.
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra/doc"

	// Import the main package to access NewRootCmd
	// This requires NewRootCmd to be in an importable package
)

// We need to recreate the command structure here since we can't import main.
// This duplicates the command structure but ensures accurate documentation.

import (
	"github.com/spf13/cobra"
)

func main() {
	outputDir := "docs/_generated/cli"

	// Check if output directory was provided as argument
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Create the root command
	rootCmd := newRootCmd()

	// Generate markdown documentation
	if err := doc.GenMarkdownTree(rootCmd, outputDir); err != nil {
		log.Fatalf("Failed to generate docs: %v", err)
	}

	// Count generated files
	files, _ := filepath.Glob(filepath.Join(outputDir, "*.md"))
	log.Printf("Generated %d documentation files in %s", len(files), outputDir)
}

// newRootCmd creates the command tree for documentation generation.
// This mirrors the structure in cmd/python-executor/main.go.
func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "python-executor",
		Short: "Remote Python code execution CLI",
		Long: `Execute Python code remotely in isolated containers.

The python-executor CLI provides command-line access to execute Python code
on a remote server in sandboxed Docker containers.

Environment Variables:
  PYEXEC_SERVER    Server URL (default: http://localhost:8080)

Documentation:     https://github.com/geraldthewes/python-executor/blob/main/README.md
Configuration:     https://github.com/geraldthewes/python-executor/blob/main/docs/configuration.md`,
	}

	// Global flags (for documentation purposes)
	rootCmd.PersistentFlags().String("server", "http://localhost:8080", "Server URL (env: PYEXEC_SERVER)")
	rootCmd.PersistentFlags().Int("timeout", 0, "Execution timeout in seconds (0 = server default)")
	rootCmd.PersistentFlags().Int("memory", 0, "Memory limit in MB (0 = server default)")
	rootCmd.PersistentFlags().Int("disk", 0, "Disk limit in MB (0 = server default)")
	rootCmd.PersistentFlags().Int("cpu", 0, "CPU shares (0 = server default)")
	rootCmd.PersistentFlags().Bool("network", false, "Allow network access (required for pip install)")
	rootCmd.PersistentFlags().String("image", "", "Docker image to use")
	rootCmd.PersistentFlags().Bool("async", false, "Submit asynchronously and return execution ID")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Quiet mode: only output stdout on success")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose mode: show execution details")

	// Commands
	rootCmd.AddCommand(runCmd())
	rootCmd.AddCommand(submitCmd())
	rootCmd.AddCommand(followCmd())
	rootCmd.AddCommand(killCmd())
	rootCmd.AddCommand(versionCmd())

	return rootCmd
}

func runCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [file|directory|tar] [-- script-args...]",
		Short: "Execute code synchronously",
		Long: `Execute Python code and wait for the result.

Input can be provided via:
  - stdin:     echo 'print("hi")' | python-executor run
  - file:      python-executor run script.py
  - directory: python-executor run ./myproject/
  - tar:       python-executor run code.tar

Arguments after -- are passed to the Python script as sys.argv.

Examples:
  # Run code from stdin
  echo 'print("Hello")' | python-executor run

  # Run a Python file
  python-executor run script.py

  # Run a directory (uses main.py or __main__.py as entrypoint)
  python-executor run ./myproject/

  # Pass arguments to the script
  python-executor run script.py -- --verbose input.txt

  # Run with dependencies
  python-executor run --requirements requirements.txt script.py

  # Forward environment variables
  python-executor run -e API_KEY -e DEBUG=true script.py`,
		Run: func(cmd *cobra.Command, args []string) {},
	}

	cmd.Flags().StringSlice("file", nil, "Additional file to include (can be repeated)")
	cmd.Flags().String("entrypoint", "", "Override the entrypoint script (default: auto-detect)")
	cmd.Flags().String("requirements", "", "Path to requirements.txt (enables network)")
	cmd.Flags().StringArrayP("env", "e", nil, "Environment variable: VAR (from env) or VAR=value")

	return cmd
}

func submitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit [file|directory|tar] [-- script-args...]",
		Short: "Submit code asynchronously",
		Long: `Submit code for execution and return immediately with an execution ID.

Use this for long-running tasks. The execution ID can be used to:
  - Check status: python-executor follow <id>
  - Kill:         python-executor kill <id>

Examples:
  # Submit and get execution ID
  EXEC_ID=$(python-executor submit long_task.py)
  echo "Submitted: $EXEC_ID"

  # Later, follow the execution
  python-executor follow $EXEC_ID`,
		Run: func(cmd *cobra.Command, args []string) {},
	}

	cmd.Flags().StringSlice("file", nil, "Additional file to include (can be repeated)")
	cmd.Flags().String("entrypoint", "", "Override the entrypoint script (default: auto-detect)")
	cmd.Flags().String("requirements", "", "Path to requirements.txt (enables network)")
	cmd.Flags().StringArrayP("env", "e", nil, "Environment variable: VAR (from env) or VAR=value")

	return cmd
}

func followCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "follow <execution-id>",
		Short: "Follow an async execution",
		Long: `Poll an asynchronous execution until complete and display the result.

The command polls the server every 2 seconds until the execution finishes,
then prints stdout/stderr and exits with the script's exit code.

Example:
  # Submit and follow
  EXEC_ID=$(python-executor submit script.py)
  python-executor follow $EXEC_ID`,
		Run: func(cmd *cobra.Command, args []string) {},
	}
}

func killCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "kill <execution-id>",
		Short: "Kill a running execution",
		Long: `Terminate a running execution.

The Docker container running the Python code will be forcefully stopped.

Example:
  python-executor kill exe_550e8400-e29b-41d4-a716-446655440000`,
		Run: func(cmd *cobra.Command, args []string) {},
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display the version of the python-executor CLI.`,
		Run:   func(cmd *cobra.Command, args []string) {},
	}
}
