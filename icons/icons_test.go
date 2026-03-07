package icons

import (
	"strings"
	"testing"
)

func TestNewLibrary(t *testing.T) {
	lib := NewLibrary()
	icons := lib.All()

	if len(icons) == 0 {
		t.Error("Expected non-empty icon library")
	}

	// Check we have icons from all categories
	categories := lib.Categories()
	expectedCats := []Category{
		CategoryEssentials,
		CategoryDev,
		CategoryInfra,
		CategoryTech,
		CategorySocial,
		CategoryAWS,
		CategoryAzure,
		CategoryGCP,
	}

	for _, cat := range expectedCats {
		if categories[cat] == 0 {
			t.Errorf("Expected icons in category %s", cat)
		}
	}
}

func TestByCategory(t *testing.T) {
	lib := NewLibrary()

	tests := []struct {
		category Category
		minCount int
	}{
		{CategoryEssentials, 10},
		{CategoryDev, 20},
		{CategoryAWS, 30},
		{CategoryAzure, 20},
		{CategoryGCP, 20},
	}

	for _, tt := range tests {
		t.Run(string(tt.category), func(t *testing.T) {
			icons := lib.ByCategory(tt.category)
			if len(icons) < tt.minCount {
				t.Errorf("Expected at least %d icons in %s, got %d", tt.minCount, tt.category, len(icons))
			}

			// All icons should have the correct category
			for _, icon := range icons {
				if icon.Category != tt.category {
					t.Errorf("Icon %s has wrong category: got %s, want %s", icon.Name, icon.Category, tt.category)
				}
			}
		})
	}
}

func TestSearch(t *testing.T) {
	lib := NewLibrary()

	tests := []struct {
		query    string
		minCount int
		mustFind string
	}{
		{"kubernetes", 2, "kubernetes"},
		{"docker", 1, "docker"},
		{"database", 2, "database"},
		{"lambda", 1, "Lambda"},
		{"s3", 1, "S3"},
		{"serverless", 2, ""},
		{"nosql", 2, ""},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results := lib.Search(tt.query)
			if len(results) < tt.minCount {
				t.Errorf("Search(%q) returned %d results, want at least %d", tt.query, len(results), tt.minCount)
			}

			if tt.mustFind != "" {
				found := false
				for _, icon := range results {
					if strings.Contains(strings.ToLower(icon.Name), strings.ToLower(tt.mustFind)) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Search(%q) did not find icon containing %q", tt.query, tt.mustFind)
				}
			}
		})
	}
}

func TestSearchNoResults(t *testing.T) {
	lib := NewLibrary()
	results := lib.Search("xyznonexistent123")
	if len(results) != 0 {
		t.Errorf("Expected no results for nonexistent query, got %d", len(results))
	}
}

func TestIconURL(t *testing.T) {
	lib := NewLibrary()
	icons := lib.All()

	for _, icon := range icons {
		// All URLs should start with the base URL
		if !strings.HasPrefix(icon.URL, BaseURL) {
			t.Errorf("Icon %s has invalid URL: %s", icon.Name, icon.URL)
		}

		// All URLs should end with .svg
		if !strings.HasSuffix(icon.URL, ".svg") {
			t.Errorf("Icon %s URL does not end with .svg: %s", icon.Name, icon.URL)
		}
	}
}

func TestBuildURL(t *testing.T) {
	tests := []struct {
		category Category
		parts    []string
		want     string
	}{
		{CategoryDev, []string{"docker"}, "https://icons.terrastruct.com/dev/docker.svg"},
		{CategoryEssentials, []string{"365-user"}, "https://icons.terrastruct.com/essentials/365-user.svg"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := BuildURL(tt.category, tt.parts...)
			if got != tt.want {
				t.Errorf("BuildURL(%s, %v) = %q, want %q", tt.category, tt.parts, got, tt.want)
			}
		})
	}
}

func TestAllCategories(t *testing.T) {
	cats := AllCategories()
	if len(cats) != 8 {
		t.Errorf("Expected 8 categories, got %d", len(cats))
	}
}

func TestIconHasKeywords(t *testing.T) {
	lib := NewLibrary()

	// Check that popular icons have keywords
	ec2 := lib.Search("EC2")
	if len(ec2) == 0 {
		t.Fatal("Expected to find EC2 icon")
	}

	if len(ec2[0].Keywords) == 0 {
		t.Error("Expected EC2 icon to have keywords")
	}
}
