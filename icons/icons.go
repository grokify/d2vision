// Package icons provides access to D2's icon library hosted at icons.terrastruct.com.
package icons

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const BaseURL = "https://icons.terrastruct.com"

// Category represents an icon category.
type Category string

const (
	CategoryEssentials Category = "essentials"
	CategoryDev        Category = "dev"
	CategoryInfra      Category = "infra"
	CategoryTech       Category = "tech"
	CategorySocial     Category = "social"
	CategoryAWS        Category = "aws"
	CategoryAzure      Category = "azure"
	CategoryGCP        Category = "gcp"
)

// AllCategories returns all available icon categories.
func AllCategories() []Category {
	return []Category{
		CategoryEssentials,
		CategoryDev,
		CategoryInfra,
		CategoryTech,
		CategorySocial,
		CategoryAWS,
		CategoryAzure,
		CategoryGCP,
	}
}

// Icon represents a single icon in the library.
type Icon struct {
	Name        string   `json:"name" toon:"Name"`
	Category    Category `json:"category" toon:"Category"`
	Subcategory string   `json:"subcategory,omitempty" toon:"Subcategory"`
	URL         string   `json:"url" toon:"URL"`
	Keywords    []string `json:"keywords,omitempty" toon:"Keywords"`
}

// IconLibrary contains all available icons.
type IconLibrary struct {
	icons []Icon
}

// NewLibrary creates a new icon library with all built-in icons.
func NewLibrary() *IconLibrary {
	lib := &IconLibrary{}
	lib.icons = append(lib.icons, essentialsIcons()...)
	lib.icons = append(lib.icons, devIcons()...)
	lib.icons = append(lib.icons, infraIcons()...)
	lib.icons = append(lib.icons, techIcons()...)
	lib.icons = append(lib.icons, socialIcons()...)
	lib.icons = append(lib.icons, awsIcons()...)
	lib.icons = append(lib.icons, azureIcons()...)
	lib.icons = append(lib.icons, gcpIcons()...)
	return lib
}

// All returns all icons in the library.
func (l *IconLibrary) All() []Icon {
	return l.icons
}

// ByCategory returns icons filtered by category.
func (l *IconLibrary) ByCategory(cat Category) []Icon {
	var result []Icon
	for _, icon := range l.icons {
		if icon.Category == cat {
			result = append(result, icon)
		}
	}
	return result
}

// Search returns icons matching the search query.
// Searches name, category, subcategory, and keywords.
func (l *IconLibrary) Search(query string) []Icon {
	query = strings.ToLower(query)
	var result []Icon

	for _, icon := range l.icons {
		if l.iconMatches(icon, query) {
			result = append(result, icon)
		}
	}

	// Sort by relevance (name match first, then keyword match)
	sort.Slice(result, func(i, j int) bool {
		iNameMatch := strings.Contains(strings.ToLower(result[i].Name), query)
		jNameMatch := strings.Contains(strings.ToLower(result[j].Name), query)
		if iNameMatch != jNameMatch {
			return iNameMatch
		}
		return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
	})

	return result
}

func (l *IconLibrary) iconMatches(icon Icon, query string) bool {
	// Check name
	if strings.Contains(strings.ToLower(icon.Name), query) {
		return true
	}

	// Check category
	if strings.Contains(strings.ToLower(string(icon.Category)), query) {
		return true
	}

	// Check subcategory
	if strings.Contains(strings.ToLower(icon.Subcategory), query) {
		return true
	}

	// Check keywords
	for _, kw := range icon.Keywords {
		if strings.Contains(strings.ToLower(kw), query) {
			return true
		}
	}

	return false
}

// Categories returns a summary of icons by category.
func (l *IconLibrary) Categories() map[Category]int {
	counts := make(map[Category]int)
	for _, icon := range l.icons {
		counts[icon.Category]++
	}
	return counts
}

// BuildURL constructs an icon URL from category and path components.
func BuildURL(category Category, parts ...string) string {
	path := string(category)
	for _, part := range parts {
		path += "/" + url.PathEscape(part)
	}
	return fmt.Sprintf("%s/%s.svg", BaseURL, path)
}

// essentialsIcons returns common UI/UX icons.
func essentialsIcons() []Icon {
	icons := []Icon{
		{Name: "user", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "365-user"), Keywords: []string{"person", "account", "profile"}},
		{Name: "users", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "359-users"), Keywords: []string{"people", "team", "group"}},
		{Name: "worldwide", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "214-worldwide"), Keywords: []string{"globe", "internet", "world", "web"}},
		{Name: "network", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "092-network"), Keywords: []string{"connection", "nodes", "mesh"}},
		{Name: "server", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "112-server"), Keywords: []string{"host", "machine", "computer"}},
		{Name: "database", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "119-database"), Keywords: []string{"db", "storage", "data"}},
		{Name: "cloud", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "109-cloud"), Keywords: []string{"hosting", "saas", "iaas"}},
		{Name: "lock", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "139-lock"), Keywords: []string{"security", "secure", "protected"}},
		{Name: "unlock", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "140-unlock"), Keywords: []string{"open", "unsecured"}},
		{Name: "key", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "141-key"), Keywords: []string{"access", "credential", "auth"}},
		{Name: "settings", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "182-settings"), Keywords: []string{"config", "gear", "options"}},
		{Name: "search", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "003-search"), Keywords: []string{"find", "lookup", "magnify"}},
		{Name: "home", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "001-home"), Keywords: []string{"house", "main", "index"}},
		{Name: "mail", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "044-mail"), Keywords: []string{"email", "message", "envelope"}},
		{Name: "chat", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "047-chat"), Keywords: []string{"message", "conversation", "talk"}},
		{Name: "calendar", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "070-calendar"), Keywords: []string{"date", "schedule", "event"}},
		{Name: "clock", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "073-clock"), Keywords: []string{"time", "schedule", "timer"}},
		{Name: "document", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "053-document"), Keywords: []string{"file", "page", "paper"}},
		{Name: "folder", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "089-folder"), Keywords: []string{"directory", "files"}},
		{Name: "download", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "007-download"), Keywords: []string{"save", "get"}},
		{Name: "upload", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "006-upload"), Keywords: []string{"send", "push"}},
		{Name: "refresh", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "018-refresh"), Keywords: []string{"reload", "sync", "update"}},
		{Name: "link", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "Chain"), Keywords: []string{"chain", "url", "connect"}},
		{Name: "api", Category: CategoryEssentials, URL: BuildURL(CategoryEssentials, "152-api"), Keywords: []string{"interface", "endpoint", "rest"}},
	}
	return icons
}

// devIcons returns development tool icons.
func devIcons() []Icon {
	icons := []Icon{
		// Languages
		{Name: "go", Category: CategoryDev, URL: BuildURL(CategoryDev, "go"), Keywords: []string{"golang", "language"}},
		{Name: "python", Category: CategoryDev, URL: BuildURL(CategoryDev, "python"), Keywords: []string{"py", "language"}},
		{Name: "javascript", Category: CategoryDev, URL: BuildURL(CategoryDev, "javascript"), Keywords: []string{"js", "node", "language"}},
		{Name: "typescript", Category: CategoryDev, URL: BuildURL(CategoryDev, "typescript"), Keywords: []string{"ts", "language"}},
		{Name: "java", Category: CategoryDev, URL: BuildURL(CategoryDev, "java"), Keywords: []string{"jvm", "language"}},
		{Name: "rust", Category: CategoryDev, URL: BuildURL(CategoryDev, "rust"), Keywords: []string{"rs", "language"}},
		{Name: "ruby", Category: CategoryDev, URL: BuildURL(CategoryDev, "ruby"), Keywords: []string{"rb", "rails", "language"}},
		{Name: "php", Category: CategoryDev, URL: BuildURL(CategoryDev, "php"), Keywords: []string{"language"}},
		{Name: "csharp", Category: CategoryDev, URL: BuildURL(CategoryDev, "csharp"), Keywords: []string{"c#", "dotnet", "language"}},
		{Name: "swift", Category: CategoryDev, URL: BuildURL(CategoryDev, "swift"), Keywords: []string{"ios", "apple", "language"}},
		{Name: "kotlin", Category: CategoryDev, URL: BuildURL(CategoryDev, "kotlin"), Keywords: []string{"android", "jvm", "language"}},

		// Tools & Platforms
		{Name: "docker", Category: CategoryDev, URL: BuildURL(CategoryDev, "docker"), Keywords: []string{"container", "containerization"}},
		{Name: "kubernetes", Category: CategoryDev, URL: BuildURL(CategoryDev, "kubernetes"), Keywords: []string{"k8s", "orchestration", "container"}},
		{Name: "git", Category: CategoryDev, URL: BuildURL(CategoryDev, "git"), Keywords: []string{"version control", "vcs"}},
		{Name: "github", Category: CategoryDev, URL: BuildURL(CategoryDev, "github"), Keywords: []string{"git", "repo", "repository"}},
		{Name: "gitlab", Category: CategoryDev, URL: BuildURL(CategoryDev, "gitlab"), Keywords: []string{"git", "repo", "ci"}},
		{Name: "bitbucket", Category: CategoryDev, URL: BuildURL(CategoryDev, "bitbucket"), Keywords: []string{"git", "repo", "atlassian"}},
		{Name: "jenkins", Category: CategoryDev, URL: BuildURL(CategoryDev, "jenkins"), Keywords: []string{"ci", "cd", "pipeline"}},
		{Name: "terraform", Category: CategoryDev, URL: BuildURL(CategoryDev, "terraform"), Keywords: []string{"iac", "infrastructure"}},
		{Name: "ansible", Category: CategoryDev, URL: BuildURL(CategoryDev, "ansible"), Keywords: []string{"automation", "config"}},

		// Databases
		{Name: "postgresql", Category: CategoryDev, URL: BuildURL(CategoryDev, "postgresql"), Keywords: []string{"postgres", "db", "database", "sql"}},
		{Name: "mysql", Category: CategoryDev, URL: BuildURL(CategoryDev, "mysql"), Keywords: []string{"db", "database", "sql"}},
		{Name: "mongodb", Category: CategoryDev, URL: BuildURL(CategoryDev, "mongodb"), Keywords: []string{"mongo", "db", "database", "nosql"}},
		{Name: "redis", Category: CategoryDev, URL: BuildURL(CategoryDev, "redis"), Keywords: []string{"cache", "db", "database", "kv"}},
		{Name: "elasticsearch", Category: CategoryDev, URL: BuildURL(CategoryDev, "elasticsearch"), Keywords: []string{"elastic", "search", "db"}},

		// Web Servers & Frameworks
		{Name: "nginx", Category: CategoryDev, URL: BuildURL(CategoryDev, "nginx"), Keywords: []string{"web server", "proxy", "load balancer"}},
		{Name: "apache", Category: CategoryDev, URL: BuildURL(CategoryDev, "apache"), Keywords: []string{"httpd", "web server"}},
		{Name: "react", Category: CategoryDev, URL: BuildURL(CategoryDev, "react"), Keywords: []string{"frontend", "ui", "javascript"}},
		{Name: "vue", Category: CategoryDev, URL: BuildURL(CategoryDev, "vue"), Keywords: []string{"frontend", "ui", "javascript"}},
		{Name: "angular", Category: CategoryDev, URL: BuildURL(CategoryDev, "angular"), Keywords: []string{"frontend", "ui", "javascript"}},
		{Name: "nodejs", Category: CategoryDev, URL: BuildURL(CategoryDev, "nodejs"), Keywords: []string{"node", "javascript", "backend"}},

		// Messaging
		{Name: "kafka", Category: CategoryDev, URL: BuildURL(CategoryDev, "kafka"), Keywords: []string{"queue", "messaging", "streaming"}},
		{Name: "rabbitmq", Category: CategoryDev, URL: BuildURL(CategoryDev, "rabbitmq"), Keywords: []string{"queue", "messaging", "amqp"}},

		// Browsers
		{Name: "chrome", Category: CategoryDev, URL: BuildURL(CategoryDev, "chrome"), Keywords: []string{"browser", "google"}},
		{Name: "firefox", Category: CategoryDev, URL: BuildURL(CategoryDev, "firefox"), Keywords: []string{"browser", "mozilla"}},
		{Name: "safari", Category: CategoryDev, URL: BuildURL(CategoryDev, "safari"), Keywords: []string{"browser", "apple"}},
	}
	return icons
}

// infraIcons returns infrastructure icons.
func infraIcons() []Icon {
	icons := []Icon{
		{Name: "firewall", Category: CategoryInfra, URL: BuildURL(CategoryInfra, "014-firewall"), Keywords: []string{"security", "network"}},
		{Name: "router", Category: CategoryInfra, URL: BuildURL(CategoryInfra, "006-router"), Keywords: []string{"network", "gateway"}},
		{Name: "switch", Category: CategoryInfra, URL: BuildURL(CategoryInfra, "010-switch"), Keywords: []string{"network", "ethernet"}},
		{Name: "load-balancer", Category: CategoryInfra, URL: BuildURL(CategoryInfra, "015-load-balancing"), Keywords: []string{"lb", "traffic", "distribution"}},
		{Name: "dns", Category: CategoryInfra, URL: BuildURL(CategoryInfra, "020-dns"), Keywords: []string{"domain", "name server"}},
		{Name: "vpn", Category: CategoryInfra, URL: BuildURL(CategoryInfra, "021-vpn"), Keywords: []string{"tunnel", "private network"}},
		{Name: "backup", Category: CategoryInfra, URL: BuildURL(CategoryInfra, "003-backup"), Keywords: []string{"restore", "recovery"}},
		{Name: "monitoring", Category: CategoryInfra, URL: BuildURL(CategoryInfra, "018-monitoring"), Keywords: []string{"observability", "metrics"}},
		{Name: "ssl", Category: CategoryInfra, URL: BuildURL(CategoryInfra, "024-ssl"), Keywords: []string{"tls", "certificate", "https"}},
		{Name: "access-denied", Category: CategoryInfra, URL: BuildURL(CategoryInfra, "001-access-denied"), Keywords: []string{"blocked", "forbidden"}},
	}
	return icons
}

// techIcons returns technology/hardware icons.
func techIcons() []Icon {
	icons := []Icon{
		{Name: "laptop", Category: CategoryTech, URL: BuildURL(CategoryTech, "001-laptop"), Keywords: []string{"computer", "notebook", "client"}},
		{Name: "desktop", Category: CategoryTech, URL: BuildURL(CategoryTech, "002-desktop"), Keywords: []string{"computer", "workstation", "pc"}},
		{Name: "mobile", Category: CategoryTech, URL: BuildURL(CategoryTech, "003-mobile"), Keywords: []string{"phone", "smartphone", "device"}},
		{Name: "tablet", Category: CategoryTech, URL: BuildURL(CategoryTech, "004-tablet"), Keywords: []string{"ipad", "device"}},
		{Name: "server-rack", Category: CategoryTech, URL: BuildURL(CategoryTech, "005-server"), Keywords: []string{"datacenter", "hardware"}},
		{Name: "cpu", Category: CategoryTech, URL: BuildURL(CategoryTech, "010-cpu"), Keywords: []string{"processor", "chip"}},
		{Name: "memory", Category: CategoryTech, URL: BuildURL(CategoryTech, "011-memory"), Keywords: []string{"ram", "chip"}},
		{Name: "storage", Category: CategoryTech, URL: BuildURL(CategoryTech, "012-storage"), Keywords: []string{"disk", "ssd", "hdd"}},
		{Name: "printer", Category: CategoryTech, URL: BuildURL(CategoryTech, "020-printer"), Keywords: []string{"print", "output"}},
		{Name: "camera", Category: CategoryTech, URL: BuildURL(CategoryTech, "030-camera"), Keywords: []string{"webcam", "video"}},
	}
	return icons
}

// socialIcons returns social media icons.
func socialIcons() []Icon {
	icons := []Icon{
		{Name: "twitter", Category: CategorySocial, URL: BuildURL(CategorySocial, "013-twitter"), Keywords: []string{"x", "social media"}},
		{Name: "facebook", Category: CategorySocial, URL: BuildURL(CategorySocial, "006-facebook"), Keywords: []string{"meta", "social media"}},
		{Name: "linkedin", Category: CategorySocial, URL: BuildURL(CategorySocial, "030-linkedin"), Keywords: []string{"professional", "social media"}},
		{Name: "instagram", Category: CategorySocial, URL: BuildURL(CategorySocial, "034-instagram"), Keywords: []string{"photo", "social media"}},
		{Name: "youtube", Category: CategorySocial, URL: BuildURL(CategorySocial, "015-youtube"), Keywords: []string{"video", "social media"}},
		{Name: "slack", Category: CategorySocial, URL: BuildURL(CategorySocial, "054-slack"), Keywords: []string{"chat", "messaging"}},
		{Name: "discord", Category: CategorySocial, URL: BuildURL(CategorySocial, "055-discord"), Keywords: []string{"chat", "gaming"}},
		{Name: "github-social", Category: CategorySocial, URL: BuildURL(CategorySocial, "039-github"), Keywords: []string{"code", "developer"}},
	}
	return icons
}

// awsIcons returns AWS service icons.
func awsIcons() []Icon {
	icons := []Icon{
		// Compute
		{Name: "EC2", Category: CategoryAWS, Subcategory: "Compute", URL: awsURL("Compute", "Amazon-EC2"), Keywords: []string{"instance", "vm", "server"}},
		{Name: "Lambda", Category: CategoryAWS, Subcategory: "Compute", URL: awsURL("Compute", "AWS-Lambda"), Keywords: []string{"serverless", "function"}},
		{Name: "ECS", Category: CategoryAWS, Subcategory: "Compute", URL: awsURL("Compute", "Amazon-ECS"), Keywords: []string{"container", "docker"}},
		{Name: "EKS", Category: CategoryAWS, Subcategory: "Compute", URL: awsURL("Compute", "Amazon-EKS"), Keywords: []string{"kubernetes", "k8s", "container"}},
		{Name: "Fargate", Category: CategoryAWS, Subcategory: "Compute", URL: awsURL("Compute", "AWS-Fargate"), Keywords: []string{"serverless", "container"}},
		{Name: "Batch", Category: CategoryAWS, Subcategory: "Compute", URL: awsURL("Compute", "AWS-Batch"), Keywords: []string{"job", "processing"}},

		// Storage
		{Name: "S3", Category: CategoryAWS, Subcategory: "Storage", URL: awsURL("Storage", "Amazon-S3-Standard"), Keywords: []string{"object storage", "bucket"}},
		{Name: "EBS", Category: CategoryAWS, Subcategory: "Storage", URL: awsURL("Storage", "Amazon-EBS"), Keywords: []string{"block storage", "volume"}},
		{Name: "EFS", Category: CategoryAWS, Subcategory: "Storage", URL: awsURL("Storage", "Amazon-EFS"), Keywords: []string{"file storage", "nfs"}},
		{Name: "Glacier", Category: CategoryAWS, Subcategory: "Storage", URL: awsURL("Storage", "Amazon-S3-Glacier"), Keywords: []string{"archive", "backup"}},

		// Database
		{Name: "RDS", Category: CategoryAWS, Subcategory: "Database", URL: awsURL("Database", "Amazon-RDS"), Keywords: []string{"relational", "sql", "mysql", "postgres"}},
		{Name: "DynamoDB", Category: CategoryAWS, Subcategory: "Database", URL: awsURL("Database", "Amazon-DynamoDB"), Keywords: []string{"nosql", "key-value"}},
		{Name: "Aurora", Category: CategoryAWS, Subcategory: "Database", URL: awsURL("Database", "Amazon-Aurora"), Keywords: []string{"mysql", "postgres", "relational"}},
		{Name: "ElastiCache", Category: CategoryAWS, Subcategory: "Database", URL: awsURL("Database", "Amazon-ElastiCache"), Keywords: []string{"redis", "memcached", "cache"}},
		{Name: "Redshift", Category: CategoryAWS, Subcategory: "Database", URL: awsURL("Database", "Amazon-Redshift"), Keywords: []string{"data warehouse", "analytics"}},

		// Networking
		{Name: "VPC", Category: CategoryAWS, Subcategory: "Networking", URL: awsURL("Networking & Content Delivery", "Amazon-VPC"), Keywords: []string{"network", "virtual private cloud"}},
		{Name: "CloudFront", Category: CategoryAWS, Subcategory: "Networking", URL: awsURL("Networking & Content Delivery", "Amazon-CloudFront"), Keywords: []string{"cdn", "edge"}},
		{Name: "Route53", Category: CategoryAWS, Subcategory: "Networking", URL: awsURL("Networking & Content Delivery", "Amazon-Route-53"), Keywords: []string{"dns", "domain"}},
		{Name: "ELB", Category: CategoryAWS, Subcategory: "Networking", URL: awsURL("Networking & Content Delivery", "Elastic-Load-Balancing"), Keywords: []string{"load balancer", "alb", "nlb"}},
		{Name: "API-Gateway", Category: CategoryAWS, Subcategory: "Networking", URL: awsURL("Networking & Content Delivery", "Amazon-API-Gateway"), Keywords: []string{"api", "rest", "http"}},

		// Security
		{Name: "IAM", Category: CategoryAWS, Subcategory: "Security", URL: awsURL("Security, Identity, & Compliance", "AWS-IAM"), Keywords: []string{"identity", "access", "permissions"}},
		{Name: "Cognito", Category: CategoryAWS, Subcategory: "Security", URL: awsURL("Security, Identity, & Compliance", "Amazon-Cognito"), Keywords: []string{"auth", "authentication", "user pool"}},
		{Name: "KMS", Category: CategoryAWS, Subcategory: "Security", URL: awsURL("Security, Identity, & Compliance", "AWS-KMS"), Keywords: []string{"key", "encryption"}},
		{Name: "WAF", Category: CategoryAWS, Subcategory: "Security", URL: awsURL("Security, Identity, & Compliance", "AWS-WAF"), Keywords: []string{"firewall", "web application"}},
		{Name: "Secrets-Manager", Category: CategoryAWS, Subcategory: "Security", URL: awsURL("Security, Identity, & Compliance", "AWS-Secrets-Manager"), Keywords: []string{"secrets", "credentials"}},

		// Integration
		{Name: "SQS", Category: CategoryAWS, Subcategory: "Integration", URL: awsURL("Application Integration", "Amazon-SQS"), Keywords: []string{"queue", "messaging"}},
		{Name: "SNS", Category: CategoryAWS, Subcategory: "Integration", URL: awsURL("Application Integration", "Amazon-SNS"), Keywords: []string{"notification", "pub/sub"}},
		{Name: "EventBridge", Category: CategoryAWS, Subcategory: "Integration", URL: awsURL("Application Integration", "Amazon-EventBridge"), Keywords: []string{"events", "bus"}},
		{Name: "Step-Functions", Category: CategoryAWS, Subcategory: "Integration", URL: awsURL("Application Integration", "AWS-Step-Functions"), Keywords: []string{"workflow", "orchestration"}},

		// Analytics
		{Name: "Kinesis", Category: CategoryAWS, Subcategory: "Analytics", URL: awsURL("Analytics", "Amazon-Kinesis"), Keywords: []string{"streaming", "real-time"}},
		{Name: "Athena", Category: CategoryAWS, Subcategory: "Analytics", URL: awsURL("Analytics", "Amazon-Athena"), Keywords: []string{"query", "sql", "s3"}},
		{Name: "EMR", Category: CategoryAWS, Subcategory: "Analytics", URL: awsURL("Analytics", "Amazon-EMR"), Keywords: []string{"hadoop", "spark", "big data"}},
		{Name: "Glue", Category: CategoryAWS, Subcategory: "Analytics", URL: awsURL("Analytics", "AWS-Glue"), Keywords: []string{"etl", "catalog"}},

		// Management
		{Name: "CloudWatch", Category: CategoryAWS, Subcategory: "Management", URL: awsURL("Management & Governance", "Amazon-CloudWatch"), Keywords: []string{"monitoring", "logs", "metrics"}},
		{Name: "CloudFormation", Category: CategoryAWS, Subcategory: "Management", URL: awsURL("Management & Governance", "AWS-CloudFormation"), Keywords: []string{"iac", "infrastructure"}},
		{Name: "CloudTrail", Category: CategoryAWS, Subcategory: "Management", URL: awsURL("Management & Governance", "AWS-CloudTrail"), Keywords: []string{"audit", "logging"}},

		// Machine Learning
		{Name: "SageMaker", Category: CategoryAWS, Subcategory: "ML", URL: awsURL("Machine Learning", "Amazon-SageMaker"), Keywords: []string{"ml", "ai", "training"}},
		{Name: "Rekognition", Category: CategoryAWS, Subcategory: "ML", URL: awsURL("Machine Learning", "Amazon-Rekognition"), Keywords: []string{"image", "video", "ai"}},
		{Name: "Bedrock", Category: CategoryAWS, Subcategory: "ML", URL: awsURL("Machine Learning", "Amazon-Bedrock"), Keywords: []string{"llm", "generative ai"}},
	}
	return icons
}

func awsURL(category, name string) string {
	return fmt.Sprintf("%s/aws/%s/%s.svg", BaseURL, url.PathEscape(category), url.PathEscape(name))
}

// azureIcons returns Azure service icons.
func azureIcons() []Icon {
	icons := []Icon{
		// Compute
		{Name: "Virtual-Machines", Category: CategoryAzure, Subcategory: "Compute", URL: azureURL("Compute Service Color", "Virtual Machine"), Keywords: []string{"vm", "instance", "server"}},
		{Name: "App-Service", Category: CategoryAzure, Subcategory: "Compute", URL: azureURL("Compute Service Color", "App Services"), Keywords: []string{"web app", "paas"}},
		{Name: "Functions", Category: CategoryAzure, Subcategory: "Compute", URL: azureURL("Compute Service Color", "Function Apps"), Keywords: []string{"serverless", "lambda"}},
		{Name: "AKS", Category: CategoryAzure, Subcategory: "Compute", URL: azureURL("Container Service Color", "Kubernetes Services"), Keywords: []string{"kubernetes", "k8s"}},
		{Name: "Container-Instances", Category: CategoryAzure, Subcategory: "Compute", URL: azureURL("Container Service Color", "Container Instances"), Keywords: []string{"docker", "container"}},

		// Storage
		{Name: "Storage-Accounts", Category: CategoryAzure, Subcategory: "Storage", URL: azureURL("Storage Service Color", "Storage Accounts"), Keywords: []string{"blob", "file", "queue"}},
		{Name: "Blob-Storage", Category: CategoryAzure, Subcategory: "Storage", URL: azureURL("Storage Service Color", "Blob Storage"), Keywords: []string{"object", "s3"}},

		// Database
		{Name: "SQL-Database", Category: CategoryAzure, Subcategory: "Database", URL: azureURL("Databases Service Color", "SQL Database"), Keywords: []string{"sql server", "relational"}},
		{Name: "Cosmos-DB", Category: CategoryAzure, Subcategory: "Database", URL: azureURL("Databases Service Color", "Azure Cosmos DB"), Keywords: []string{"nosql", "global"}},
		{Name: "Cache-for-Redis", Category: CategoryAzure, Subcategory: "Database", URL: azureURL("Databases Service Color", "Cache Redis Product"), Keywords: []string{"redis", "cache"}},

		// Networking
		{Name: "Virtual-Networks", Category: CategoryAzure, Subcategory: "Networking", URL: azureURL("Networking Service Color", "Virtual Networks"), Keywords: []string{"vnet", "vpc"}},
		{Name: "Load-Balancer", Category: CategoryAzure, Subcategory: "Networking", URL: azureURL("Networking Service Color", "Load Balancers"), Keywords: []string{"lb", "traffic"}},
		{Name: "Application-Gateway", Category: CategoryAzure, Subcategory: "Networking", URL: azureURL("Networking Service Color", "Application Gateways"), Keywords: []string{"waf", "lb"}},
		{Name: "DNS-Zones", Category: CategoryAzure, Subcategory: "Networking", URL: azureURL("Networking Service Color", "DNS Zones"), Keywords: []string{"dns", "domain"}},
		{Name: "CDN", Category: CategoryAzure, Subcategory: "Networking", URL: azureURL("Networking Service Color", "CDN Profiles"), Keywords: []string{"content delivery", "edge"}},

		// Security
		{Name: "Key-Vault", Category: CategoryAzure, Subcategory: "Security", URL: azureURL("Security Service Color", "Key Vaults"), Keywords: []string{"secrets", "keys", "certificates"}},
		{Name: "Active-Directory", Category: CategoryAzure, Subcategory: "Security", URL: azureURL("Identity Service Color", "Azure Active Directory"), Keywords: []string{"ad", "identity", "auth"}},

		// Integration
		{Name: "Service-Bus", Category: CategoryAzure, Subcategory: "Integration", URL: azureURL("Integration Service Color", "Service Bus"), Keywords: []string{"queue", "messaging"}},
		{Name: "Event-Grid", Category: CategoryAzure, Subcategory: "Integration", URL: azureURL("Integration Service Color", "Event Grid Domains"), Keywords: []string{"events", "pub/sub"}},
		{Name: "Logic-Apps", Category: CategoryAzure, Subcategory: "Integration", URL: azureURL("Integration Service Color", "Logic Apps"), Keywords: []string{"workflow", "automation"}},
		{Name: "API-Management", Category: CategoryAzure, Subcategory: "Integration", URL: azureURL("Integration Service Color", "API Management Services"), Keywords: []string{"api gateway", "proxy"}},

		// DevOps
		{Name: "DevOps", Category: CategoryAzure, Subcategory: "DevOps", URL: azureURL("DevOps Service Color", "Azure DevOps"), Keywords: []string{"ci/cd", "pipelines"}},
		{Name: "Repos", Category: CategoryAzure, Subcategory: "DevOps", URL: azureURL("DevOps Service Color", "Azure Repos"), Keywords: []string{"git", "source control"}},

		// Monitoring
		{Name: "Monitor", Category: CategoryAzure, Subcategory: "Management", URL: azureURL("Management + Governance Service Color", "Monitor"), Keywords: []string{"metrics", "logs"}},
		{Name: "Application-Insights", Category: CategoryAzure, Subcategory: "Management", URL: azureURL("Management + Governance Service Color", "Application Insights"), Keywords: []string{"apm", "tracing"}},

		// AI/ML
		{Name: "Cognitive-Services", Category: CategoryAzure, Subcategory: "AI", URL: azureURL("AI + Machine Learning Service Color", "Cognitive Services"), Keywords: []string{"ai", "ml"}},
		{Name: "Machine-Learning", Category: CategoryAzure, Subcategory: "AI", URL: azureURL("AI + Machine Learning Service Color", "Machine Learning"), Keywords: []string{"ml", "training"}},
		{Name: "OpenAI", Category: CategoryAzure, Subcategory: "AI", URL: azureURL("AI + Machine Learning Service Color", "Azure OpenAI"), Keywords: []string{"gpt", "llm", "ai"}},
	}
	return icons
}

func azureURL(category, name string) string {
	return fmt.Sprintf("%s/azure/%s/%s.svg", BaseURL, url.PathEscape(category), url.PathEscape(name))
}

// gcpIcons returns GCP service icons.
func gcpIcons() []Icon {
	icons := []Icon{
		// Compute
		{Name: "Compute-Engine", Category: CategoryGCP, Subcategory: "Compute", URL: gcpURL("Compute", "Compute Engine"), Keywords: []string{"vm", "instance", "server"}},
		{Name: "Cloud-Run", Category: CategoryGCP, Subcategory: "Compute", URL: gcpURL("Compute", "Cloud Run"), Keywords: []string{"serverless", "container"}},
		{Name: "Cloud-Functions", Category: CategoryGCP, Subcategory: "Compute", URL: gcpURL("Compute", "Cloud Functions"), Keywords: []string{"serverless", "function"}},
		{Name: "GKE", Category: CategoryGCP, Subcategory: "Compute", URL: gcpURL("Compute", "Google Kubernetes Engine"), Keywords: []string{"kubernetes", "k8s"}},
		{Name: "App-Engine", Category: CategoryGCP, Subcategory: "Compute", URL: gcpURL("Compute", "App Engine"), Keywords: []string{"paas", "web app"}},

		// Storage
		{Name: "Cloud-Storage", Category: CategoryGCP, Subcategory: "Storage", URL: gcpURL("Storage", "Cloud Storage"), Keywords: []string{"bucket", "object", "s3"}},
		{Name: "Persistent-Disk", Category: CategoryGCP, Subcategory: "Storage", URL: gcpURL("Storage", "Persistent Disk"), Keywords: []string{"block", "volume"}},
		{Name: "Filestore", Category: CategoryGCP, Subcategory: "Storage", URL: gcpURL("Storage", "Filestore"), Keywords: []string{"nfs", "file"}},

		// Database
		{Name: "Cloud-SQL", Category: CategoryGCP, Subcategory: "Database", URL: gcpURL("Databases", "Cloud SQL"), Keywords: []string{"mysql", "postgres", "sql server"}},
		{Name: "Cloud-Spanner", Category: CategoryGCP, Subcategory: "Database", URL: gcpURL("Databases", "Cloud Spanner"), Keywords: []string{"global", "relational"}},
		{Name: "Firestore", Category: CategoryGCP, Subcategory: "Database", URL: gcpURL("Databases", "Firestore"), Keywords: []string{"nosql", "document"}},
		{Name: "Bigtable", Category: CategoryGCP, Subcategory: "Database", URL: gcpURL("Databases", "Bigtable"), Keywords: []string{"nosql", "wide column"}},
		{Name: "Memorystore", Category: CategoryGCP, Subcategory: "Database", URL: gcpURL("Databases", "Memorystore"), Keywords: []string{"redis", "cache"}},

		// Networking
		{Name: "VPC", Category: CategoryGCP, Subcategory: "Networking", URL: gcpURL("Networking", "Virtual Private Cloud"), Keywords: []string{"network", "vpc"}},
		{Name: "Cloud-Load-Balancing", Category: CategoryGCP, Subcategory: "Networking", URL: gcpURL("Networking", "Cloud Load Balancing"), Keywords: []string{"lb", "traffic"}},
		{Name: "Cloud-CDN", Category: CategoryGCP, Subcategory: "Networking", URL: gcpURL("Networking", "Cloud CDN"), Keywords: []string{"cdn", "edge"}},
		{Name: "Cloud-DNS", Category: CategoryGCP, Subcategory: "Networking", URL: gcpURL("Networking", "Cloud DNS"), Keywords: []string{"dns", "domain"}},
		{Name: "Cloud-Armor", Category: CategoryGCP, Subcategory: "Networking", URL: gcpURL("Networking", "Cloud Armor"), Keywords: []string{"waf", "ddos"}},

		// Security
		{Name: "IAM", Category: CategoryGCP, Subcategory: "Security", URL: gcpURL("Security", "Cloud IAM"), Keywords: []string{"identity", "access"}},
		{Name: "Secret-Manager", Category: CategoryGCP, Subcategory: "Security", URL: gcpURL("Security", "Secret Manager"), Keywords: []string{"secrets", "credentials"}},
		{Name: "KMS", Category: CategoryGCP, Subcategory: "Security", URL: gcpURL("Security", "Key Management Service"), Keywords: []string{"encryption", "keys"}},

		// Analytics
		{Name: "BigQuery", Category: CategoryGCP, Subcategory: "Analytics", URL: gcpURL("Data Analytics", "BigQuery"), Keywords: []string{"data warehouse", "sql", "analytics"}},
		{Name: "Dataflow", Category: CategoryGCP, Subcategory: "Analytics", URL: gcpURL("Data Analytics", "Dataflow"), Keywords: []string{"streaming", "batch", "apache beam"}},
		{Name: "Pub/Sub", Category: CategoryGCP, Subcategory: "Analytics", URL: gcpURL("Data Analytics", "PubSub"), Keywords: []string{"messaging", "queue", "events"}},
		{Name: "Dataproc", Category: CategoryGCP, Subcategory: "Analytics", URL: gcpURL("Data Analytics", "Dataproc"), Keywords: []string{"hadoop", "spark"}},

		// AI/ML
		{Name: "Vertex-AI", Category: CategoryGCP, Subcategory: "AI", URL: gcpURL("AI and Machine Learning", "Vertex AI"), Keywords: []string{"ml", "training", "prediction"}},
		{Name: "Vision-AI", Category: CategoryGCP, Subcategory: "AI", URL: gcpURL("AI and Machine Learning", "Vision API"), Keywords: []string{"image", "ocr"}},
		{Name: "Natural-Language", Category: CategoryGCP, Subcategory: "AI", URL: gcpURL("AI and Machine Learning", "Natural Language API"), Keywords: []string{"nlp", "text"}},

		// Management
		{Name: "Cloud-Logging", Category: CategoryGCP, Subcategory: "Management", URL: gcpURL("Operations", "Cloud Logging"), Keywords: []string{"logs", "stackdriver"}},
		{Name: "Cloud-Monitoring", Category: CategoryGCP, Subcategory: "Management", URL: gcpURL("Operations", "Cloud Monitoring"), Keywords: []string{"metrics", "stackdriver"}},
	}
	return icons
}

func gcpURL(category, name string) string {
	return fmt.Sprintf("%s/gcp/Products and services/%s/%s.svg", BaseURL, url.PathEscape(category), url.PathEscape(name))
}
