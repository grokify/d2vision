package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/grokify/d2vision/format"
	"github.com/grokify/d2vision/icons"
	"github.com/spf13/cobra"
)

var (
	iconsFormat   string
	iconsCategory string
	iconsLimit    int
)

var iconsCmd = &cobra.Command{
	Use:   "icons",
	Short: "Browse and search D2 icon library",
	Long: `Browse and search icons from the D2 icon library at icons.terrastruct.com.

All icons are SVG (vector) format and can be used in D2 diagrams with the icon property.

Categories:
  essentials  - Common UI icons (user, network, database, etc.)
  dev         - Development tools (docker, kubernetes, git, languages)
  infra       - Infrastructure (firewall, router, load-balancer)
  tech        - Hardware (laptop, server, mobile)
  social      - Social media (twitter, github, slack)
  aws         - Amazon Web Services icons
  azure       - Microsoft Azure icons
  gcp         - Google Cloud Platform icons

Examples:
  # List all categories
  d2vision icons list

  # List icons in a category
  d2vision icons list --category aws

  # Search for icons
  d2vision icons search database

  # Search with JSON output
  d2vision icons search kubernetes --format json
`,
}

var iconsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available icons",
	Long: `List available icons from the D2 icon library.

Without flags, shows a summary of all categories.
With --category, lists all icons in that category.

Examples:
  # Show category summary
  d2vision icons list

  # List AWS icons
  d2vision icons list --category aws

  # List dev icons as JSON
  d2vision icons list --category dev --format json
`,
	RunE: runIconsList,
}

var iconsSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for icons",
	Long: `Search for icons by name, category, or keyword.

The search matches against icon names, categories, subcategories, and keywords.
Results are sorted by relevance (name matches first).

Examples:
  # Search for database icons
  d2vision icons search database

  # Search for container-related icons
  d2vision icons search container

  # Search with limit
  d2vision icons search storage --limit 5

  # JSON output for scripting
  d2vision icons search lambda --format json
`,
	Args: cobra.ExactArgs(1),
	RunE: runIconsSearch,
}

func init() {
	iconsCmd.AddCommand(iconsListCmd)
	iconsCmd.AddCommand(iconsSearchCmd)

	// List flags
	iconsListCmd.Flags().StringVarP(&iconsFormat, "format", "f", "text", "Output format: text, toon, json")
	iconsListCmd.Flags().StringVarP(&iconsCategory, "category", "c", "", "Filter by category")

	// Search flags
	iconsSearchCmd.Flags().StringVarP(&iconsFormat, "format", "f", "text", "Output format: text, toon, json")
	iconsSearchCmd.Flags().IntVarP(&iconsLimit, "limit", "n", 0, "Limit number of results (0 = no limit)")
}

func runIconsList(cmd *cobra.Command, args []string) error {
	lib := icons.NewLibrary()

	if iconsCategory == "" {
		// Show category summary
		return listCategories(lib)
	}

	// List icons in category
	cat := icons.Category(strings.ToLower(iconsCategory))
	iconList := lib.ByCategory(cat)

	if len(iconList) == 0 {
		return fmt.Errorf("unknown category: %s\n\nAvailable categories: %s",
			iconsCategory, strings.Join(categoryNames(), ", "))
	}

	return outputIcons(iconList, fmt.Sprintf("Icons in category '%s'", cat))
}

func runIconsSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
	lib := icons.NewLibrary()

	results := lib.Search(query)

	if iconsLimit > 0 && len(results) > iconsLimit {
		results = results[:iconsLimit]
	}

	if len(results) == 0 {
		fmt.Printf("No icons found matching '%s'\n", query)
		fmt.Println("\nTry searching for:")
		fmt.Println("  - Service names: ec2, lambda, kubernetes")
		fmt.Println("  - Categories: aws, azure, gcp, dev")
		fmt.Println("  - Keywords: database, container, serverless")
		return nil
	}

	return outputIcons(results, fmt.Sprintf("Search results for '%s'", query))
}

func listCategories(lib *icons.IconLibrary) error {
	counts := lib.Categories()

	if iconsFormat == "text" {
		fmt.Println("D2 Icon Library Categories")
		fmt.Println("==========================")
		fmt.Println()

		// Sort categories for consistent output
		cats := icons.AllCategories()
		total := 0

		for _, cat := range cats {
			count := counts[cat]
			total += count
			desc := categoryDescription(cat)
			fmt.Printf("  %-12s %3d icons  - %s\n", cat, count, desc)
		}

		fmt.Println()
		fmt.Printf("Total: %d icons\n", total)
		fmt.Println()
		fmt.Println("Use 'd2vision icons list --category <name>' to see icons in a category")
		fmt.Println("Use 'd2vision icons search <query>' to search for icons")
		return nil
	}

	// Structured output
	type CategorySummary struct {
		Name        string `json:"name" toon:"Name"`
		Count       int    `json:"count" toon:"Count"`
		Description string `json:"description" toon:"Description"`
	}

	var summaries []CategorySummary
	for _, cat := range icons.AllCategories() {
		summaries = append(summaries, CategorySummary{
			Name:        string(cat),
			Count:       counts[cat],
			Description: categoryDescription(cat),
		})
	}

	f, err := format.Parse(iconsFormat)
	if err != nil {
		return err
	}

	output, err := format.Marshal(summaries, f)
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

func outputIcons(iconList []icons.Icon, title string) error {
	if iconsFormat == "text" {
		fmt.Println(title)
		fmt.Println(strings.Repeat("=", len(title)))
		fmt.Println()

		// Group by subcategory if present
		bySubcat := make(map[string][]icons.Icon)
		var subcats []string

		for _, icon := range iconList {
			key := icon.Subcategory
			if key == "" {
				key = "_default"
			}
			if _, exists := bySubcat[key]; !exists {
				subcats = append(subcats, key)
			}
			bySubcat[key] = append(bySubcat[key], icon)
		}

		sort.Strings(subcats)

		for _, subcat := range subcats {
			if subcat != "_default" {
				fmt.Printf("## %s\n\n", subcat)
			}

			for _, icon := range bySubcat[subcat] {
				fmt.Printf("  %-25s %s\n", icon.Name, icon.URL)
			}
			fmt.Println()
		}

		fmt.Printf("Total: %d icons\n", len(iconList))
		fmt.Println()
		fmt.Println("Usage in D2:")
		fmt.Println("  node {")
		fmt.Println("    icon: <url>")
		fmt.Println("  }")
		return nil
	}

	// Structured output
	f, err := format.Parse(iconsFormat)
	if err != nil {
		return err
	}

	output, err := format.Marshal(iconList, f)
	if err != nil {
		return err
	}

	fmt.Println(string(output))
	return nil
}

func categoryNames() []string {
	cats := icons.AllCategories()
	names := make([]string, len(cats))
	for i, cat := range cats {
		names[i] = string(cat)
	}
	return names
}

func categoryDescription(cat icons.Category) string {
	descriptions := map[icons.Category]string{
		icons.CategoryEssentials: "Common UI icons (user, database, cloud, lock)",
		icons.CategoryDev:        "Development tools & languages (docker, kubernetes, go, python)",
		icons.CategoryInfra:      "Infrastructure (firewall, router, load-balancer, vpn)",
		icons.CategoryTech:       "Hardware & devices (laptop, server, mobile, cpu)",
		icons.CategorySocial:     "Social media & communication (twitter, github, slack)",
		icons.CategoryAWS:        "Amazon Web Services (EC2, S3, Lambda, RDS)",
		icons.CategoryAzure:      "Microsoft Azure (VMs, Functions, Cosmos DB)",
		icons.CategoryGCP:        "Google Cloud Platform (Compute, BigQuery, GKE)",
	}
	return descriptions[cat]
}
