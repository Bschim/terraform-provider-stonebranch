package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/OptionMetrics/terraform-provider-stonebranch/cmd/sb2tf/generator"
)

var (
	exportAll    bool
	exportFilter string
	noDeps       bool
	dryRun       bool
	withImports  bool

	exportCmd = &cobra.Command{
		Use:   "export [resource-type] [name]",
		Short: "Export resources to Terraform configuration",
		Long: `Export StoneBranch resources to Terraform HCL configuration files.

Examples:
  # Export a single resource
  sb2tf export task_unix my_task
  sb2tf export variable my_var

  # Export a workflow with all its dependencies (tasks, vertices, edges)
  sb2tf export task_workflow my_workflow

  # Export all resources of a type
  sb2tf export task_unix --all
  sb2tf export triggers --all

  # Export with filters (supports * and ? wildcards)
  sb2tf export tasks --all --filter "prod-*"
  sb2tf export tasks --all --filter "test_task_???"

  # Output to directory instead of stdout
  sb2tf export task_unix --all --output ./terraform/

  # Export without resolving dependencies
  sb2tf export task_workflow my_workflow --no-deps`,
		Args: cobra.MaximumNArgs(2),
		RunE: runExport,
	}
)

func init() {
	exportCmd.Flags().BoolVar(&exportAll, "all", false, "Export all resources of the specified type")
	exportCmd.Flags().StringVar(&exportFilter, "filter", "", "Filter resources by name pattern (use with --all)")
	exportCmd.Flags().BoolVar(&noDeps, "no-deps", false, "Skip exporting dependencies")
	exportCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be exported without writing files")
	exportCmd.Flags().BoolVar(&withImports, "with-imports", false, "Emit paired import {} blocks in imports.tf for each exported resource (only correct when applying against the same UAC instance you exported from)")
}

func runExport(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	ds := GetDataSource()
	output := GetOutput()

	// Create generator
	gen := generator.NewGenerator(ds, output, noDeps, withImports)

	// Handle different argument combinations
	if len(args) == 0 {
		if !exportAll {
			return fmt.Errorf("must specify a resource type and name, or use --all flag")
		}
		return exportAllResources(ctx, gen)
	}

	resourceType := args[0]

	// Handle category shortcuts
	switch resourceType {
	case "tasks":
		return exportCategory(ctx, gen, "Tasks")
	case "triggers":
		return exportCategory(ctx, gen, "Triggers")
	case "connections":
		return exportCategory(ctx, gen, "Connections")
	}

	// Check if this is a valid resource type
	rt := generator.GetResourceType(resourceType)
	if rt == nil {
		return fmt.Errorf("unknown resource type: %s\nRun 'sb2tf list' to see available types", resourceType)
	}

	// Export all of this type
	if exportAll {
		if dryRun {
			return dryRunExport(ctx, gen, resourceType)
		}
		if err := gen.ExportAll(ctx, resourceType, exportFilter); err != nil {
			return err
		}
		return gen.Finalize()
	}

	// Export single resource
	if len(args) < 2 {
		return fmt.Errorf("must specify resource name or use --all flag")
	}

	name := args[1]

	if dryRun {
		fmt.Printf("Would export %s/%s\n", resourceType, name)
		if !noDeps {
			fmt.Println("  (plus any dependencies)")
		}
		return nil
	}

	// Special handling for workflows
	if resourceType == "task_workflow" && !noDeps {
		if err := gen.ExportWorkflow(ctx, name); err != nil {
			return err
		}
	} else {
		if err := gen.ExportResource(ctx, resourceType, name); err != nil {
			return err
		}
	}

	if err := gen.Finalize(); err != nil {
		return err
	}

	// Print summary
	exported := gen.GetExportedResources()
	if len(exported) > 1 {
		fmt.Fprintf(os.Stderr, "\nExported %d resources\n", len(exported))
	}

	return nil
}

func exportAllResources(ctx context.Context, gen *generator.Generator) error {
	categories := generator.GetResourceCategories()

	for _, cat := range categories {
		// Skip workflow components - they're exported as part of workflows
		if cat.Name == "Workflow" {
			continue
		}

		for _, rt := range cat.Types {
			if dryRun {
				fmt.Printf("Would export all %s resources\n", rt.CLIName)
				continue
			}

			fmt.Fprintf(os.Stderr, "Exporting %s...\n", rt.CLIName)
			if err := gen.ExportAll(ctx, rt.CLIName, exportFilter); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to export %s: %v\n", rt.CLIName, err)
			}
		}
	}

	if !dryRun {
		return gen.Finalize()
	}
	return nil
}

func exportCategory(ctx context.Context, gen *generator.Generator, categoryName string) error {
	categories := generator.GetResourceCategories()

	for _, cat := range categories {
		if cat.Name != categoryName {
			continue
		}

		if !exportAll && exportFilter == "" {
			return fmt.Errorf("must specify --all or --filter when exporting category '%s'", categoryName)
		}

		for _, rt := range cat.Types {
			if dryRun {
				fmt.Printf("Would export all %s resources\n", rt.CLIName)
				continue
			}

			fmt.Fprintf(os.Stderr, "Exporting %s...\n", rt.CLIName)
			if err := gen.ExportAll(ctx, rt.CLIName, exportFilter); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to export %s: %v\n", rt.CLIName, err)
			}
		}

		if !dryRun {
			return gen.Finalize()
		}
		return nil
	}

	return fmt.Errorf("unknown category: %s", categoryName)
}

func dryRunExport(ctx context.Context, gen *generator.Generator, resourceType string) error {
	rt := generator.GetResourceType(resourceType)
	if rt == nil {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}

	items, err := gen.ListResources(ctx, rt, exportFilter)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		fmt.Printf("No %s resources found\n", resourceType)
		return nil
	}

	fmt.Printf("Would export %d %s resource(s):\n", len(items), resourceType)
	for _, item := range items {
		fmt.Printf("  - %s\n", item.Name)
	}

	return nil
}
