package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/OptionMetrics/terraform-provider-stonebranch/cmd/sb2tf/generator"
)

var (
	listFilter string

	listCmd = &cobra.Command{
		Use:   "list [resource-type]",
		Short: "List available resources",
		Long: `List resources from StoneBranch Universal Controller.

Without arguments, shows available resource types.
With a resource type, lists resources of that type.

Examples:
  sb2tf list                    # Show all resource types
  sb2tf list tasks              # List all tasks
  sb2tf list task_unix          # List Unix tasks only
  sb2tf list triggers           # List all triggers
  sb2tf list --filter "prod-*"  # Filter by name pattern`,
		Args: cobra.MaximumNArgs(1),
		RunE: runList,
	}
)

func init() {
	listCmd.Flags().StringVar(&listFilter, "filter", "", "Filter resources by name pattern (supports wildcards)")
}

func runList(cmd *cobra.Command, args []string) error {
	// No arguments: show available resource types
	if len(args) == 0 {
		return listResourceTypes()
	}

	resourceType := args[0]

	// Handle category shortcuts (tasks, triggers, connections, etc.)
	switch resourceType {
	case "tasks":
		return listAllTasks()
	case "triggers":
		return listAllTriggers()
	case "connections":
		return listAllConnections()
	default:
		return listSpecificType(resourceType)
	}
}

func listResourceTypes() error {
	fmt.Println("Available resource types:")
	fmt.Println()

	categories := generator.GetResourceCategories()
	for _, cat := range categories {
		fmt.Printf("  %s:\n", cat.Name)
		for _, rt := range cat.Types {
			fmt.Printf("    %-25s %s\n", rt.CLIName, rt.TerraformResource)
		}
		fmt.Println()
	}

	fmt.Println("Shortcuts:")
	fmt.Println("  tasks       - List all task types")
	fmt.Println("  triggers    - List all trigger types")
	fmt.Println("  connections - List all connection types")

	return nil
}

// listByCategory lists every resource across all resource types belonging
// to the given category, calling DataSource.List once per type and
// concatenating the results. When an item's Type field comes back empty,
// it defaults to defaultType(rt) - this preserves parity with the old
// combined-endpoint behavior where "type" was always populated.
func listByCategory(categoryName string, defaultType func(rt *generator.ResourceType) string) ([]generator.ResourceItem, error) {
	ctx := context.Background()
	ds := GetDataSource()

	categories := generator.GetResourceCategories()

	var all []generator.ResourceItem
	for _, cat := range categories {
		if cat.Name != categoryName {
			continue
		}

		for _, rt := range cat.Types {
			items, err := ds.List(ctx, rt, listFilter)
			if err != nil {
				return nil, fmt.Errorf("failed to list %s: %w", rt.CLIName, err)
			}
			for i := range items {
				if items[i].Type == "" {
					items[i].Type = defaultType(rt)
				}
			}
			all = append(all, items...)
		}
	}

	return all, nil
}

func listAllTasks() error {
	tasks, err := listByCategory("Tasks", func(rt *generator.ResourceType) string {
		return rt.APITypeValue
	})
	if err != nil {
		return fmt.Errorf("failed to list tasks: %w", err)
	}

	printResourceTable("Tasks", tasks, true)
	return nil
}

func listAllTriggers() error {
	triggers, err := listByCategory("Triggers", func(rt *generator.ResourceType) string {
		return rt.APITypeValue
	})
	if err != nil {
		return fmt.Errorf("failed to list triggers: %w", err)
	}

	printResourceTable("Triggers", triggers, true)
	return nil
}

func listAllConnections() error {
	allConnections, err := listByCategory("Connections", func(rt *generator.ResourceType) string {
		return rt.CLIName
	})
	if err != nil {
		return fmt.Errorf("failed to list connections: %w", err)
	}

	printResourceTable("Connections", allConnections, true)
	return nil
}

func listSpecificType(resourceType string) error {
	rt := generator.GetResourceType(resourceType)
	if rt == nil {
		return fmt.Errorf("unknown resource type: %s\nRun 'sb2tf list' to see available types", resourceType)
	}

	ctx := context.Background()
	ds := GetDataSource()

	items, err := ds.List(ctx, rt, listFilter)
	if err != nil {
		return fmt.Errorf("failed to list %s: %w", resourceType, err)
	}

	// Set the type for display if not returned by the data source
	for i := range items {
		if items[i].Type == "" {
			items[i].Type = resourceType
		}
	}

	printResourceTable(strings.Title(resourceType), items, rt.HasTypeField)
	return nil
}

func printResourceTable(title string, items []generator.ResourceItem, showType bool) {
	if len(items) == 0 {
		fmt.Printf("No %s found.\n", strings.ToLower(title))
		return
	}

	// Sort by name
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})

	fmt.Printf("%s (%d):\n", title, len(items))
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	if showType {
		fmt.Fprintln(w, "NAME\tTYPE\tSUMMARY")
		fmt.Fprintln(w, "----\t----\t-------")
		for _, item := range items {
			summary := truncate(item.Summary, 50)
			fmt.Fprintf(w, "%s\t%s\t%s\n", item.Name, item.Type, summary)
		}
	} else {
		fmt.Fprintln(w, "NAME\tSUMMARY")
		fmt.Fprintln(w, "----\t-------")
		for _, item := range items {
			summary := truncate(item.Summary, 60)
			fmt.Fprintf(w, "%s\t%s\n", item.Name, summary)
		}
	}
	w.Flush()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
