package collectors

import (
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func TestCloudflareScraperSelection(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		service  string
		expected string
	}{
		{
			name:     "Explicit cloudflare provider",
			provider: "cloudflare",
			service:  "cloudflare",
			expected: "enhancedCloudflareParser",
		},
		{
			name:     "Service name contains cloudflare",
			provider: "",
			service:  "cloudflare-status",
			expected: "enhancedCloudflareParser",
		},
		{
			name:     "Mixed case provider",
			provider: "Cloudflare",
			service:  "status",
			expected: "enhancedCloudflareParser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraper := ScraperForService(tt.provider, tt.service)
			scraperType := ""
			switch scraper.(type) {
			case enhancedCloudflareParser:
				scraperType = "enhancedCloudflareParser"
			case genericParser:
				scraperType = "genericParser"
			case enhancedAWSParser:
				scraperType = "enhancedAWSParser"
			case enhancedGCPParser:
				scraperType = "enhancedGCPParser"
			case enhancedAzureParser:
				scraperType = "enhancedAzureParser"
			case genesysParser:
				scraperType = "genesysParser"
			case enhancedAvayaParser:
				scraperType = "enhancedAvayaParser"
			default:
				scraperType = "unknown"
			}
			assert.Equal(t, tt.expected, scraperType)
		})
	}
}

func TestCloudflareGenericParserLimitations(t *testing.T) {
	// Test that demonstrates the limitations of the generic parser for Cloudflare
	parser := genericParser{}
	
	tests := []struct {
		name                string
		title               string
		description         string
		guid                string
		expectedServiceName string
		expectedRegion      string
		expectedKey         string
	}{
		{
			name:                "Maintenance with datacenter code",
			title:               "LAX (Los Angeles) on 2025-07-03",
			description:         "Scheduled maintenance in LAX datacenter",
			guid:                "tag:www.cloudflarestatus.com,2005:Incident/25460322",
			expectedServiceName: "", // Generic parser can't extract service
			expectedRegion:      "", // Generic parser can't extract region
			expectedKey:         "tag:www.cloudflarestatus.com,2005:Incident/25460322",
		},
		{
			name:                "DNS service outage",
			title:               "DNS Service Outage - Global",
			description:         "DNS resolution failures affecting multiple regions",
			guid:                "tag:www.cloudflarestatus.com,2005:Incident/25460400",
			expectedServiceName: "", // Generic parser can't extract "DNS" service
			expectedRegion:      "", // Generic parser can't extract "Global" region
			expectedKey:         "tag:www.cloudflarestatus.com,2005:Incident/25460400",
		},
		{
			name:                "CDN performance issue",
			title:               "CDN Performance Issues - European Datacenters",
			description:         "Performance degradation in European CDN nodes",
			guid:                "tag:www.cloudflarestatus.com,2005:Incident/25460401",
			expectedServiceName: "", // Generic parser can't extract "CDN" service
			expectedRegion:      "", // Generic parser can't extract "European" region
			expectedKey:         "tag:www.cloudflarestatus.com,2005:Incident/25460401",
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
			key := parser.IncidentKey(item)

			assert.Equal(t, tt.expectedServiceName, serviceName, "Service name should be empty for generic parser")
			assert.Equal(t, tt.expectedRegion, region, "Region should be empty for generic parser")
			assert.Equal(t, tt.expectedKey, key, "Incident key should match GUID")
		})
	}
}

func TestCloudflareServiceExtraction(t *testing.T) {
	// Test what service and region information could be extracted from Cloudflare titles
	tests := []struct {
		name            string
		title           string
		expectedService string
		expectedRegion  string
	}{
		{
			name:            "Datacenter maintenance with location",
			title:           "LAX (Los Angeles) on 2025-07-03",
			expectedService: "Datacenter Maintenance",
			expectedRegion:  "LAX (Los Angeles)",
		},
		{
			name:            "DNS service outage global",
			title:           "DNS Service Outage - Global",
			expectedService: "DNS",
			expectedRegion:  "Global",
		},
		{
			name:            "CDN performance European datacenters",
			title:           "CDN Performance Issues - European Datacenters",
			expectedService: "CDN",
			expectedRegion:  "European Datacenters",
		},
		{
			name:            "WAF service Asia Pacific",
			title:           "WAF Blocking Issues - Asia Pacific",
			expectedService: "WAF",
			expectedRegion:  "Asia Pacific",
		},
		{
			name:            "API service North America",
			title:           "API Rate Limiting - North America",
			expectedService: "API",
			expectedRegion:  "North America",
		},
	}

	// This test documents what an enhanced parser could extract
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is what we would want an enhanced parser to extract
			// For now, just document the expected behavior
			t.Logf("Title: %s", tt.title)
			t.Logf("Expected Service: %s", tt.expectedService)
			t.Logf("Expected Region: %s", tt.expectedRegion)
			
			// The generic parser returns empty values
			parser := genericParser{}
			item := &gofeed.Item{Title: tt.title}
			service, region := parser.ServiceInfo(item)
			
			assert.Equal(t, "", service, "Generic parser returns empty service")
			assert.Equal(t, "", region, "Generic parser returns empty region")
		})
	}
}