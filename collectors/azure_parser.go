package collectors

import (
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"
)

type enhancedAzureParser struct{}

func (enhancedAzureParser) ServiceInfo(item *gofeed.Item) (string, string) {
	serviceName := extractAzureService(item)
	region := extractAzureRegion(item)
	return serviceName, region
}

func (enhancedAzureParser) IncidentKey(item *gofeed.Item) string {
	if item.GUID != "" {
		return normalizeAzureIncidentKey(item.GUID)
	}
	if item.Link != "" {
		return item.Link
	}
	return strings.TrimSpace(item.Title)
}

// parseAzureGUID extracts service name and region from an Azure GUID of the form
// "service-region_xyz". Unknown formats return empty strings.
func parseAzureGUID(guid string) (serviceName, region string) {
	if idx := strings.Index(guid, "#"); idx != -1 {
		guid = guid[idx+1:]
	}
	if idx := strings.IndexAny(guid, "_"); idx != -1 {
		guid = guid[:idx]
	}
	parts := strings.Split(guid, "-")
	if len(parts) >= 2 {
		serviceName = strings.ToLower(parts[0])
		region = strings.Join(parts[1:], "-")
	}
	return
}

func extractAzureService(item *gofeed.Item) string {
	content := strings.ToLower(item.Title + " " + item.Description + " " + item.Content)
	
	// Azure service patterns - ordered by specificity
	servicePatterns := []struct {
		pattern string
		service string
	}{
		// Compute Services
		{"virtual machine|vm scale sets|vmss", "Virtual Machines"},
		{"azure kubernetes service|aks", "Azure Kubernetes Service"},
		{"app service|web apps", "App Service"},
		{"azure functions|function apps", "Azure Functions"},
		{"container instances|aci", "Container Instances"},
		{"service fabric", "Service Fabric"},
		{"batch", "Azure Batch"},
		
		// Storage Services
		{"blob storage|block blob", "Blob Storage"},
		{"azure files|file storage", "Azure Files"},
		{"queue storage", "Queue Storage"},
		{"table storage", "Table Storage"},
		{"disk storage|managed disks", "Disk Storage"},
		{"data lake storage|adls", "Data Lake Storage"},
		{"storage account", "Storage Accounts"},
		
		// Database Services
		{"sql database|azure sql", "SQL Database"},
		{"cosmos db|cosmosdb", "Cosmos DB"},
		{"mysql|azure database for mysql", "Azure Database for MySQL"},
		{"postgresql|azure database for postgresql", "Azure Database for PostgreSQL"},
		{"redis cache", "Azure Cache for Redis"},
		{"sql managed instance", "SQL Managed Instance"},
		{"synapse analytics|azure synapse", "Azure Synapse Analytics"},
		
		// Networking Services
		{"application gateway", "Application Gateway"},
		{"load balancer", "Load Balancer"},
		{"traffic manager", "Traffic Manager"},
		{"virtual network|vnet", "Virtual Network"},
		{"vpn gateway", "VPN Gateway"},
		{"expressroute", "ExpressRoute"},
		{"azure firewall", "Azure Firewall"},
		{"front door", "Azure Front Door"},
		{"cdn|content delivery", "Azure CDN"},
		{"dns", "Azure DNS"},
		
		// AI & ML Services
		{"cognitive services", "Cognitive Services"},
		{"machine learning|azure ml", "Azure Machine Learning"},
		{"bot service|bot framework", "Bot Service"},
		{"form recognizer", "Form Recognizer"},
		{"computer vision", "Computer Vision"},
		{"speech service", "Speech Services"},
		
		// Analytics & Data Services
		{"event hubs", "Event Hubs"},
		{"service bus", "Service Bus"},
		{"data factory", "Data Factory"},
		{"stream analytics", "Stream Analytics"},
		{"hdinsight", "HDInsight"},
		{"databricks", "Azure Databricks"},
		
		// Security & Identity
		{"active directory|azure ad", "Azure Active Directory"},
		{"key vault", "Key Vault"},
		{"security center", "Security Center"},
		{"sentinel", "Azure Sentinel"},
		{"information protection", "Azure Information Protection"},
		
		// Management & Monitoring
		{"monitor|azure monitor", "Azure Monitor"},
		{"application insights", "Application Insights"},
		{"log analytics", "Log Analytics"},
		{"automation", "Azure Automation"},
		{"backup", "Azure Backup"},
		{"site recovery", "Azure Site Recovery"},
		{"resource manager|arm", "Azure Resource Manager"},
		
		// Integration Services
		{"logic apps", "Logic Apps"},
		{"api management", "API Management"},
		{"event grid", "Event Grid"},
		
		// IoT Services
		{"iot hub", "IoT Hub"},
		{"iot central", "IoT Central"},
		{"digital twins", "Azure Digital Twins"},
		
		// Media Services
		{"media services", "Media Services"},
		{"content delivery network", "Content Delivery Network"},
		
		// Generic fallbacks
		{"storage", "Storage"},
		{"compute", "Compute"},
		{"networking", "Networking"},
		{"database", "Database"},
	}
	
	for _, sp := range servicePatterns {
		if matched, _ := regexp.MatchString(sp.pattern, content); matched {
			return sp.service
		}
	}
	
	// Extract from GUID if available
	if item.GUID != "" {
		if svc, _ := parseAzureGUID(item.GUID); svc != "" {
			return formatAzureServiceName(svc)
		}
	}
	
	// Extract from title after colon
	title := strings.ToLower(item.Title)
	if idx := strings.Index(title, ":"); idx != -1 {
		title = strings.TrimSpace(title[idx+1:])
		parts := strings.Split(title, " - ")
		if len(parts) >= 1 {
			return formatAzureServiceName(strings.TrimSpace(parts[0]))
		}
	}
	
	return "Azure Platform"
}

func extractAzureRegion(item *gofeed.Item) string {
	content := strings.ToLower(item.Title + " " + item.Description + " " + item.Content)
	
	// Azure region patterns - ordered by specificity (most specific first)
	regionPatterns := []struct {
		pattern string
		region  string
	}{
		// Global/Multi-region indicators - check these first
		{"multiple regions|multi-region", "Multi-Region"},
		{"global|worldwide|all regions", "Global"},
		// North America
		{"east us 2|eastus2", "East US 2"},
		{"east us|eastus", "East US"},
		{"west us 3|westus3", "West US 3"},
		{"west us 2|westus2", "West US 2"},
		{"west us|westus", "West US"},
		{"central us|centralus", "Central US"},
		{"north central us|northcentralus", "North Central US"},
		{"south central us|southcentralus", "South Central US"},
		{"west central us|westcentralus", "West Central US"},
		{"canada central|canadacentral", "Canada Central"},
		{"canada east|canadaeast", "Canada East"},
		
		// Europe
		{"north europe|northeurope", "North Europe"},
		{"west europe|westeurope", "West Europe"},
		{"uk south|uksouth", "UK South"},
		{"uk west|ukwest", "UK West"},
		{"france central|francecentral", "France Central"},
		{"france south|francesouth", "France South"},
		{"germany west central|germanywestcentral", "Germany West Central"},
		{"germany north|germanynorth", "Germany North"},
		{"norway east|norwayeast", "Norway East"},
		{"norway west|norwaywest", "Norway West"},
		{"switzerland north|switzerlandnorth", "Switzerland North"},
		{"switzerland west|switzerlandwest", "Switzerland West"},
		
		// Asia Pacific
		{"southeast asia|southeastasia", "Southeast Asia"},
		{"east asia|eastasia", "East Asia"},
		{"australia east|australiaeast", "Australia East"},
		{"australia southeast|australiasoutheast", "Australia Southeast"},
		{"australia central|australiacentral", "Australia Central"},
		{"japan east|japaneast", "Japan East"},
		{"japan west|japanwest", "Japan West"},
		{"korea central|koreacentral", "Korea Central"},
		{"korea south|koreasouth", "Korea South"},
		{"central india|centralindia", "Central India"},
		{"south india|southindia", "South India"},
		{"west india|westindia", "West India"},
		
		// South America & Africa
		{"brazil south|brazilsouth", "Brazil South"},
		{"south africa north|southafricanorth", "South Africa North"},
		{"south africa west|southafricawest", "South Africa West"},
		
		// Middle East
		{"uae north|uaenorth", "UAE North"},
		{"uae central|uaecentral", "UAE Central"},
		
		// Government & Special
		{"us gov virginia|usgovvirginia", "US Gov Virginia"},
		{"us gov texas|usgovtexas", "US Gov Texas"},
		{"us gov arizona|usgovarizona", "US Gov Arizona"},
		{"china east|chinaeast", "China East"},
		{"china north|chinanorth", "China North"},
	}
	
	for _, rp := range regionPatterns {
		if matched, _ := regexp.MatchString(rp.pattern, content); matched {
			return rp.region
		}
	}
	
	// Extract from GUID if available
	if item.GUID != "" {
		if _, reg := parseAzureGUID(item.GUID); reg != "" {
			return formatAzureRegionName(reg)
		}
	}
	
	// Extract from title after service name
	title := strings.ToLower(item.Title)
	if idx := strings.Index(title, ":"); idx != -1 {
		title = strings.TrimSpace(title[idx+1:])
		parts := strings.Split(title, " - ")
		if len(parts) >= 2 {
			return formatAzureRegionName(strings.TrimSpace(parts[1]))
		}
	}
	
	return ""
}

func normalizeAzureIncidentKey(guid string) string {
	key := guid
	if idx := strings.Index(key, "#"); idx != -1 {
		key = key[idx+1:]
	}
	// Remove Azure-specific suffixes
	suffixes := []string{"_resolved", "_issue", "_investigating", "_update", "_monitoring"}
	for _, suffix := range suffixes {
		key = strings.TrimSuffix(key, suffix)
	}
	return key
}

func formatAzureServiceName(service string) string {
	service = strings.TrimSpace(service)
	if service == "" {
		return "Azure Platform"
	}
	
	// Capitalize first letter and format known abbreviations
	words := strings.Fields(strings.ToLower(service))
	for i, word := range words {
		switch word {
		case "vm", "vms":
			words[i] = "VM"
		case "sql":
			words[i] = "SQL"
		case "ai":
			words[i] = "AI"
		case "ml":
			words[i] = "ML"
		case "iot":
			words[i] = "IoT"
		case "cdn":
			words[i] = "CDN"
		case "api":
			words[i] = "API"
		case "vpn":
			words[i] = "VPN"
		case "dns":
			words[i] = "DNS"
		case "ad":
			words[i] = "AD"
		default:
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			}
		}
	}
	
	return strings.Join(words, " ")
}

func formatAzureRegionName(region string) string {
	region = strings.TrimSpace(region)
	if region == "" {
		return ""
	}
	
	// Handle common Azure region formats - most specific first
	replacements := map[string]string{
		"eastus2":       "East US 2",
		"eastus":        "East US",
		"westus3":       "West US 3", 
		"westus2":       "West US 2",
		"westus":        "West US",
		"centralus":     "Central US",
		"northeurope":   "North Europe",
		"westeurope":    "West Europe",
		"southeastasia": "Southeast Asia",
		"eastasia":      "East Asia",
	}
	
	lowerRegion := strings.ToLower(region)
	for old, new := range replacements {
		if lowerRegion == old {
			return new
		}
	}
	
	// Capitalize first letter of each word for other formats
	words := strings.Fields(strings.ToLower(region))
	for i, word := range words {
		if len(word) > 0 {
			// Special cases for abbreviations
			if word == "us" {
				words[i] = "US"
			} else if word == "uk" {
				words[i] = "UK"
			} else {
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			}
		}
	}
	
	return strings.Join(words, " ")
}