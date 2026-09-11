// Package cli provides the command-line interface for sb2tf.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/OptionMetrics/terraform-provider-stonebranch/cmd/sb2tf/generator"
	"github.com/OptionMetrics/terraform-provider-stonebranch/internal/client"
)

var (
	// Global flags
	apiToken  string
	username  string
	password  string
	baseURL   string
	output    string
	sourceDir string

	// Shared client
	apiClient *client.Client

	// Shared local data source (built lazily, cached once)
	localDataSource *generator.LocalDataSource

	// Version (set from main)
	version = "dev"

	rootCmd = &cobra.Command{
		Use:   "sb2tf",
		Short: "Export StoneBranch resources to Terraform configuration",
		Long: `sb2tf reads resources from the StoneBranch Universal Controller API
and generates Terraform configuration files (.tf).

Use this tool to:
  - Bootstrap a new Terraform project from existing resources
  - Test that Terraform configs can reproduce existing infrastructure
  - Migrate manually-created resources to Infrastructure as Code

Authentication:
  Set STONEBRANCH_API_TOKEN environment variable or use --token flag for Bearer auth.
  Or set STONEBRANCH_USERNAME/STONEBRANCH_PASSWORD (or --username/--password) for HTTP Basic Auth.
  Set STONEBRANCH_BASE_URL environment variable or use --url flag.`,
		PersistentPreRunE: initClient,
		SilenceUsage:      true,
	}
)

// SetVersion sets the version string (called from main).
func SetVersion(v string) {
	version = v
	rootCmd.Version = v
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&apiToken, "token", "", "StoneBranch API token (env: STONEBRANCH_API_TOKEN)")
	rootCmd.PersistentFlags().StringVar(&username, "username", "", "StoneBranch Basic Auth username (env: STONEBRANCH_USERNAME)")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "StoneBranch Basic Auth password (env: STONEBRANCH_PASSWORD)")
	rootCmd.PersistentFlags().StringVar(&baseURL, "url", "", "StoneBranch base URL (env: STONEBRANCH_BASE_URL)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "", "Output directory (default: stdout)")
	rootCmd.PersistentFlags().StringVarP(&sourceDir, "source-dir", "s", "", "Read resources from a local JSON export directory instead of the API")

	// Add subcommands
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(exportCmd)
}

// initClient initializes the API client from flags or environment variables.
func initClient(cmd *cobra.Command, args []string) error {
	// Skip client init for help/version commands
	if cmd.Name() == "help" || cmd.Name() == "version" {
		return nil
	}

	// When reading from a local export directory, skip the API
	// token/URL requirements entirely - no client is needed.
	if sourceDir != "" {
		return nil
	}

	// Get token from flag or environment
	token := apiToken
	if token == "" {
		token = os.Getenv("STONEBRANCH_API_TOKEN")
	}

	// Get Basic Auth credentials from flags or environment
	user := username
	if user == "" {
		user = os.Getenv("STONEBRANCH_USERNAME")
	}
	pass := password
	if pass == "" {
		pass = os.Getenv("STONEBRANCH_PASSWORD")
	}

	if (user != "") != (pass != "") {
		return fmt.Errorf("both --username and --password (or STONEBRANCH_USERNAME/STONEBRANCH_PASSWORD) must be set together")
	}
	hasBasicAuth := user != "" && pass != ""

	if token == "" && !hasBasicAuth {
		return fmt.Errorf("authentication required: set STONEBRANCH_API_TOKEN (or --token), or STONEBRANCH_USERNAME/STONEBRANCH_PASSWORD (or --username/--password)")
	}
	if token != "" && hasBasicAuth {
		return fmt.Errorf("ambiguous authentication: set either STONEBRANCH_API_TOKEN (--token) or STONEBRANCH_USERNAME/STONEBRANCH_PASSWORD (--username/--password), not both")
	}

	// Get base URL from flag or environment
	url := baseURL
	if url == "" {
		url = os.Getenv("STONEBRANCH_BASE_URL")
	}
	if url == "" {
		fmt.Fprintln(os.Stderr, "Error: STONEBRANCH_BASE_URL environment variable or --base-url flag is required")
		os.Exit(1)
	}

	// Create client
	if hasBasicAuth {
		apiClient = client.NewBasicAuthClient(url, user, pass)
	} else {
		apiClient = client.NewClient(url, token)
	}
	return nil
}

// GetDataSource returns the DataSource to use for the current invocation:
// a LocalDataSource (built once and cached) when --source-dir/-s was
// provided, otherwise an APIDataSource wrapping the shared API client
// initialized by initClient.
func GetDataSource() generator.DataSource {
	if sourceDir != "" {
		if localDataSource == nil {
			localDataSource = generator.NewLocalDataSource(sourceDir)
		}
		return localDataSource
	}
	return generator.NewAPIDataSource(apiClient)
}

// GetOutput returns the output directory.
func GetOutput() string {
	return output
}
