package collectors

import (
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func TestEnhancedAzureParser_ServiceInfo(t *testing.T) {
	parser := enhancedAzureParser{}
	
	tests := []struct {
		name           string
		title          string
		description    string
		guid           string
		expectedService string
		expectedRegion  string
	}{
		{
			name:           "VM Scale Sets with specific region",
			title:          "Service issue: VM Scale Sets - East US",
			description:    "We are investigating issues with Virtual Machine Scale Sets in East US region.",
			expectedService: "Virtual Machines",
			expectedRegion:  "East US",
		},
		{
			name:           "Azure SQL Database multiple regions", 
			title:          "Service degradation: Azure SQL Database - Multiple Regions",
			description:    "Customers may experience connectivity issues with Azure SQL Database across multiple regions including West Europe and Southeast Asia.",
			expectedService: "SQL Database",
			expectedRegion:  "Multi-Region",
		},
		{
			name:           "Blob Storage with GUID",
			title:          "Storage issue in West US 2",
			description:    "Issues with blob storage operations.",
			guid:           "blobstorage-westus2_issue",
			expectedService: "Blob Storage",
			expectedRegion:  "West US 2",
		},
		{
			name:           "Azure Kubernetes Service",
			title:          "AKS cluster provisioning delays - North Europe",
			description:    "Azure Kubernetes Service cluster creation is experiencing delays in North Europe.",
			expectedService: "Azure Kubernetes Service", 
			expectedRegion:  "North Europe",
		},
		{
			name:           "App Service global issue",
			title:          "App Service deployment issues - Global",
			description:    "Web Apps deployment is affected globally across all regions.",
			expectedService: "App Service",
			expectedRegion:  "Global",
		},
		{
			name:           "Cosmos DB with specific service",
			title:          "Database connectivity: Cosmos DB - Japan East",
			description:    "Cosmos DB operations are experiencing elevated latency.",
			expectedService: "Cosmos DB",
			expectedRegion:  "Japan East",
		},
		{
			name:           "Azure Functions in Australia",
			title:          "Function execution delays - Australia Southeast",
			description:    "Azure Functions are experiencing execution delays in Australia Southeast region.",
			expectedService: "Azure Functions",
			expectedRegion:  "Australia Southeast",
		},
		{
			name:           "Generic Azure platform issue",
			title:          "Platform connectivity issue",
			description:    "General connectivity issues affecting Azure services.",
			expectedService: "Azure Platform",
			expectedRegion:  "",
		},
		{
			name:           "Azure Monitor in UK",
			title:          "Monitoring data collection delays - UK South",
			description:    "Azure Monitor is experiencing data collection delays in UK South.",
			expectedService: "Azure Monitor",
			expectedRegion:  "UK South",
		},
		{
			name:           "Active Directory authentication",
			title:          "Azure AD authentication delays - Central US",
			description:    "Azure Active Directory authentication is experiencing delays.",
			expectedService: "Azure Active Directory",
			expectedRegion:  "Central US",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:       tt.title,
				Description: tt.description,
				GUID:        tt.guid,
			}
			
			serviceName, region := parser.ServiceInfo(item)
			assert.Equal(t, tt.expectedService, serviceName, "Service name mismatch")
			assert.Equal(t, tt.expectedRegion, region, "Region mismatch")
		})
	}
}

func TestEnhancedAzureParser_IncidentKey(t *testing.T) {
	parser := enhancedAzureParser{}
	
	tests := []struct {
		name        string
		guid        string
		link        string
		title       string
		expectedKey string
	}{
		{
			name:        "GUID available",
			guid:        "storage-eastus_issue",
			expectedKey: "storage-eastus",
		},
		{
			name:        "GUID with hash",
			guid:        "#sqldb-westeurope_investigating", 
			expectedKey: "sqldb-westeurope",
		},
		{
			name:        "No GUID, use link",
			link:        "https://status.azure.com/incidents/abc123",
			title:       "Service Issue",
			expectedKey: "https://status.azure.com/incidents/abc123",
		},
		{
			name:        "No GUID or link, use title",
			title:       "Azure Storage connectivity issue",
			expectedKey: "Azure Storage connectivity issue",
		},
		{
			name:        "GUID with multiple suffixes",
			guid:        "appservice-northeurope_monitoring_update",
			expectedKey: "appservice-northeurope",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				GUID:  tt.guid,
				Link:  tt.link,
				Title: tt.title,
			}
			
			key := parser.IncidentKey(item)
			assert.Equal(t, tt.expectedKey, key)
		})
	}
}

func TestExtractAzureService(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "Virtual Machines",
			content:  "Virtual Machine scale sets are experiencing issues",
			expected: "Virtual Machines",
		},
		{
			name:     "SQL Database",
			content:  "Azure SQL Database connectivity problems",
			expected: "SQL Database",
		},
		{
			name:     "Blob Storage",
			content:  "Blob storage operations affected",
			expected: "Blob Storage",
		},
		{
			name:     "Azure Kubernetes Service",
			content:  "AKS cluster provisioning delays",
			expected: "Azure Kubernetes Service",
		},
		{
			name:     "Application Gateway", 
			content:  "Application Gateway routing issues",
			expected: "Application Gateway",
		},
		{
			name:     "Cognitive Services",
			content:  "Cognitive Services API delays",
			expected: "Cognitive Services",
		},
		{
			name:     "Event Hubs",
			content:  "Event Hubs message processing delays",
			expected: "Event Hubs",
		},
		{
			name:     "Key Vault",
			content:  "Key Vault secret retrieval issues",
			expected: "Key Vault",
		},
		{
			name:     "Generic storage",
			content:  "Storage account access issues",
			expected: "Storage Accounts",
		},
		{
			name:     "No recognizable service",
			content:  "Platform wide connectivity issues",
			expected: "Azure Platform",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:       tt.content,
				Description: "",
			}
			
			service := extractAzureService(item)
			assert.Equal(t, tt.expected, service)
		})
	}
}

func TestExtractAzureRegion(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "East US region",
			content:  "Issues affecting East US region",
			expected: "East US",
		},
		{
			name:     "West Europe", 
			content:  "Services in West Europe are affected",
			expected: "West Europe",
		},
		{
			name:     "Southeast Asia",
			content:  "Southeast Asia region experiencing delays",
			expected: "Southeast Asia",
		},
		{
			name:     "Australia East",
			content:  "Australia East region impact",
			expected: "Australia East",
		},
		{
			name:     "Multiple regions",
			content:  "Multiple regions affected including East US and West Europe",
			expected: "Multi-Region",
		},
		{
			name:     "Global impact",
			content:  "Global connectivity issues affecting all regions",
			expected: "Global",
		},
		{
			name:     "UK South region",
			content:  "UK South region services degraded",
			expected: "UK South",
		},
		{
			name:     "Central India",
			content:  "Central India experiencing service issues",
			expected: "Central India",
		},
		{
			name:     "No region mentioned",
			content:  "Service experiencing general issues",
			expected: "",
		},
		{
			name:     "Japan East",
			content:  "Japan East region affected by outage",
			expected: "Japan East",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:       tt.content,
				Description: "",
			}
			
			region := extractAzureRegion(item)
			assert.Equal(t, tt.expected, region)
		})
	}
}

func TestFormatAzureServiceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"storage", "Storage"},
		{"vm scale sets", "VM Scale Sets"},
		{"sql database", "SQL Database"},
		{"iot hub", "IoT Hub"},
		{"api management", "API Management"},
		{"", "Azure Platform"},
		{"cdn", "CDN"},
		{"vpn gateway", "VPN Gateway"},
		{"active directory", "Active Directory"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatAzureServiceName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatAzureRegionName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"eastus", "East US"},
		{"westus2", "West US 2"},
		{"northeurope", "North Europe"},
		{"southeastasia", "Southeast Asia"},
		{"east us", "East US"},
		{"", ""},
		{"global", "Global"},
		{"uk south", "UK South"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatAzureRegionName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}