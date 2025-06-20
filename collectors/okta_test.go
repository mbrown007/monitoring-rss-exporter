package collectors

import (
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func TestOktaScraperSelection(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		service  string
		expected string
	}{
		{
			name:     "Explicit okta provider",
			provider: "okta",
			service:  "okta",
			expected: "genericParser",
		},
		{
			name:     "Service name contains okta",
			provider: "",
			service:  "okta-status",
			expected: "genericParser",
		},
		{
			name:     "Mixed case provider",
			provider: "Okta",
			service:  "status",
			expected: "genericParser",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scraper := ScraperForService(tt.provider, tt.service)
			scraperType := ""
			switch scraper.(type) {
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
			case enhancedCloudflareParser:
				scraperType = "enhancedCloudflareParser"
			default:
				scraperType = "unknown"
			}
			assert.Equal(t, tt.expected, scraperType)
		})
	}
}

func TestOktaGenericParserLimitations(t *testing.T) {
	// Test that demonstrates the limitations of the generic parser for Okta
	parser := genericParser{}
	
	tests := []struct {
		name                string
		title               string
		content             string
		expectedServiceName string
		expectedRegion      string
		expectedKey         string
	}{
		{
			name:                "Service disruption",
			title:               "Service Disruption",
			content:             "Authentication failures affecting customers in US Cell 1, US Cell 3, and EMEA Cell 9",
			expectedServiceName: "", // Generic parser can't extract service
			expectedRegion:      "", // Generic parser can't extract region
			expectedKey:         "test-guid",
		},
		{
			name:                "Advanced Server Access issue",
			title:               "Feature Disruption",
			content:             "Advanced Server Access team is aware of an issue affecting the ASA dashboard in US Cell 1, US Cell 2",
			expectedServiceName: "", // Generic parser can't extract "Advanced Server Access"
			expectedRegion:      "", // Generic parser can't extract "US Cell 1, US Cell 2"
			expectedKey:         "test-guid",
		},
		{
			name:                "Workflows issue",
			title:               "Service Degradation",
			content:             "Okta Workflows team became aware of missing telemetry affecting customers on OK1, OK2, OK3",
			expectedServiceName: "", // Generic parser can't extract "Workflows"
			expectedRegion:      "", // Generic parser can't extract "OK1, OK2, OK3"
			expectedKey:         "test-guid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:   tt.title,
				Content: tt.content,
				GUID:    "test-guid",
			}

			serviceName, region := parser.ServiceInfo(item)
			key := parser.IncidentKey(item)

			assert.Equal(t, tt.expectedServiceName, serviceName, "Service name should be empty for generic parser")
			assert.Equal(t, tt.expectedRegion, region, "Region should be empty for generic parser")
			assert.Equal(t, tt.expectedKey, key, "Incident key should match GUID")
		})
	}
}

func TestOktaServiceExtraction(t *testing.T) {
	// Test what service and region information could be extracted from Okta content
	tests := []struct {
		name            string
		title           string
		content         string
		expectedService string
		expectedRegion  string
	}{
		{
			name:            "Authentication service with US and EMEA cells",
			title:           "Service Disruption",
			content:         "Authentication failures affecting customers in US Cell 1, US Cell 3, and EMEA Cell 9",
			expectedService: "Authentication",
			expectedRegion:  "US Cell 1, US Cell 3, EMEA Cell 9",
		},
		{
			name:            "Advanced Server Access multiple cells",
			title:           "Feature Disruption",
			content:         "Advanced Server Access team is aware of an issue affecting the ASA dashboard in US Cell 1, US Cell 2, US Cell 4, and APJ Cell 1",
			expectedService: "Advanced Server Access",
			expectedRegion:  "US Cell 1, US Cell 2, US Cell 4, APJ Cell 1",
		},
		{
			name:            "Workflows telemetry issue",
			title:           "Service Degradation",
			content:         "Okta Workflows team became aware of missing telemetry affecting customers on OK1, OK2, OK3, OK4, OK6, OK7, and OK11",
			expectedService: "Workflows",
			expectedRegion:  "OK1, OK2, OK3, OK4, OK6, OK7, OK11",
		},
		{
			name:            "MFA email provider issue",
			title:           "Service Disruption",
			content:         "Issue with a downstream email provider. Customers may experience errors with MFA Emails, password reset, and user activation",
			expectedService: "MFA Emails",
			expectedRegion:  "Global",
		},
		{
			name:            "Apple Business Manager integration",
			title:           "Feature Disruption",
			content:         "Issue with the Apple Business Manager integration. Customers may experience errors while setting up the integration",
			expectedService: "Apple Business Manager",
			expectedRegion:  "Global",
		},
	}

	// This test documents what an enhanced parser could extract
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This is what we would want an enhanced parser to extract
			// For now, just document the expected behavior
			t.Logf("Title: %s", tt.title)
			t.Logf("Content: %s", tt.content)
			t.Logf("Expected Service: %s", tt.expectedService)
			t.Logf("Expected Region: %s", tt.expectedRegion)
			
			// The generic parser returns empty values
			parser := genericParser{}
			item := &gofeed.Item{Title: tt.title, Content: tt.content}
			service, region := parser.ServiceInfo(item)
			
			assert.Equal(t, "", service, "Generic parser returns empty service")
			assert.Equal(t, "", region, "Generic parser returns empty region")
		})
	}
}