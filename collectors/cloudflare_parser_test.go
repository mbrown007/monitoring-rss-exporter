package collectors

import (
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func TestEnhancedCloudflareParser_ServiceInfo(t *testing.T) {
	parser := enhancedCloudflareParser{}
	
	tests := []struct {
		name            string
		title           string
		description     string
		guid            string
		expectedService string
		expectedRegion  string
	}{
		{
			name:            "DNS service outage global",
			title:           "DNS Service Outage - Global",
			description:     "We are investigating reports of DNS resolution failures affecting multiple regions.",
			expectedService: "DNS",
			expectedRegion:  "Global",
		},
		{
			name:            "CDN performance European datacenters",
			title:           "CDN Performance Issues - European Datacenters",
			description:     "Performance degradation in our European CDN nodes.",
			expectedService: "CDN",
			expectedRegion:  "Europe",
		},
		{
			name:            "Datacenter maintenance with location",
			title:           "LAX (Los Angeles) on 2025-07-03",
			description:     "Scheduled maintenance in LAX datacenter on 2025-07-03 between 03:00 and 05:30 UTC.",
			expectedService: "Datacenter Maintenance",
			expectedRegion:  "LAX (Los Angeles)",
		},
		{
			name:            "WAF blocking Asia Pacific",
			title:           "WAF Blocking Issues - Asia Pacific",
			description:     "Web Application Firewall experiencing blocking issues in Asia Pacific region.",
			expectedService: "WAF",
			expectedRegion:  "Asia Pacific",
		},
		{
			name:            "API rate limiting North America",
			title:           "API Rate Limiting - North America",
			description:     "API Gateway rate limiting issues affecting customers in North America.",
			expectedService: "API Gateway",
			expectedRegion:  "North America",
		},
		{
			name:            "DDoS protection global",
			title:           "DDoS Protection Service Degradation",
			description:     "DDoS mitigation services experiencing degradation globally.",
			expectedService: "DDoS Protection",
			expectedRegion:  "Global",
		},
		{
			name:            "SSL certificate issues",
			title:           "SSL Certificate Provisioning Delays",
			description:     "SSL certificate provisioning is experiencing delays in multiple regions.",
			expectedService: "SSL/TLS",
			expectedRegion:  "Multi-Region",
		},
		{
			name:            "Workers edge computing",
			title:           "Cloudflare Workers Performance Issues",
			description:     "Workers experiencing performance issues in European datacenters.",
			expectedService: "Cloudflare Workers",
			expectedRegion:  "Europe",
		},
		{
			name:            "Frankfurt datacenter maintenance",
			title:           "FRA (Frankfurt) on 2025-07-05",
			description:     "Scheduled maintenance in Frankfurt datacenter.",
			expectedService: "Datacenter Maintenance",
			expectedRegion:  "FRA (Frankfurt)",
		},
		{
			name:            "Generic service issue",
			title:           "Platform Connectivity Issue",
			description:     "General connectivity issues affecting Cloudflare services.",
			expectedService: "Connectivity",
			expectedRegion:  "",
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

func TestEnhancedCloudflareParser_IncidentKey(t *testing.T) {
	parser := enhancedCloudflareParser{}
	
	tests := []struct {
		name        string
		guid        string
		link        string
		title       string
		expectedKey string
	}{
		{
			name:        "GUID tag format",
			guid:        "tag:www.cloudflarestatus.com,2005:Incident/25460322",
			expectedKey: "tag:www.cloudflarestatus.com,2005:Incident/25460322",
		},
		{
			name:        "Different incident ID",
			guid:        "tag:www.cloudflarestatus.com,2005:Incident/25460400",
			expectedKey: "tag:www.cloudflarestatus.com,2005:Incident/25460400",
		},
		{
			name:        "No GUID, use link",
			link:        "https://www.cloudflarestatus.com/incidents/dns-outage-global",
			title:       "DNS Service Outage",
			expectedKey: "https://www.cloudflarestatus.com/incidents/dns-outage-global",
		},
		{
			name:        "No GUID or link, use title",
			title:       "CDN Performance Issues - European Datacenters",
			expectedKey: "CDN Performance Issues - European Datacenters",
		},
		{
			name:        "Empty GUID, fallback to link",
			guid:        "",
			link:        "https://www.cloudflarestatus.com/incidents/cdn-performance",
			title:       "CDN Issue",
			expectedKey: "https://www.cloudflarestatus.com/incidents/cdn-performance",
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

func TestExtractCloudflareService(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		content  string
		expected string
	}{
		{
			name:     "DNS service outage",
			content:  "DNS service outage affecting resolution",
			expected: "DNS",
		},
		{
			name:     "CDN performance issue",
			content:  "CDN performance degradation reported",
			expected: "CDN",
		},
		{
			name:     "WAF blocking issue",
			content:  "WAF blocking legitimate requests",
			expected: "WAF",
		},
		{
			name:     "DDoS protection service",
			content:  "DDoS protection experiencing issues",
			expected: "DDoS Protection",
		},
		{
			name:     "API rate limiting",
			content:  "API rate limiting causing failures",
			expected: "API Gateway",
		},
		{
			name:     "SSL certificate issue",
			content:  "SSL certificate provisioning delays",
			expected: "SSL/TLS",
		},
		{
			name:     "Load balancing issue",
			content:  "Load balancing service degradation",
			expected: "Load Balancing",
		},
		{
			name:     "Cloudflare Workers",
			content:  "Workers experiencing high latency",
			expected: "Cloudflare Workers",
		},
		{
			name:     "Analytics platform",
			content:  "Analytics dashboard unavailable",
			expected: "Analytics",
		},
		{
			name:     "Bot management",
			content:  "Bot management rules not applying",
			expected: "Bot Management",
		},
		{
			name:     "Datacenter maintenance by title",
			title:    "LAX (Los Angeles) on 2025-07-03",
			expected: "Datacenter Maintenance",
		},
		{
			name:     "Frankfurt datacenter",
			title:    "FRA (Frankfurt) on 2025-07-05",
			expected: "Datacenter Maintenance",
		},
		{
			name:     "Network connectivity",
			content:  "Network connectivity issues reported",
			expected: "Network",
		},
		{
			name:     "Edge servers",
			content:  "Edge server performance degradation",
			expected: "Edge Servers",
		},
		{
			name:     "No recognizable service",
			content:  "Platform wide issues",
			expected: "Cloudflare Services",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:       tt.title,
				Description: tt.content,
			}
			
			service := extractCloudflareService(item)
			assert.Equal(t, tt.expected, service)
		})
	}
}

func TestExtractCloudflareRegion(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		content  string
		expected string
	}{
		{
			name:     "LAX datacenter",
			title:    "LAX (Los Angeles) on 2025-07-03",
			expected: "LAX (Los Angeles)",
		},
		{
			name:     "Frankfurt datacenter",
			title:    "FRA (Frankfurt) on 2025-07-05",
			expected: "FRA (Frankfurt)",
		},
		{
			name:     "London datacenter",
			title:    "LON (London) on 2025-07-10",
			expected: "LON (London)",
		},
		{
			name:     "Global issue",
			content:  "Issues affecting services globally",
			expected: "Global",
		},
		{
			name:     "European datacenters",
			content:  "Performance issues in European datacenters",
			expected: "Europe",
		},
		{
			name:     "Asia Pacific region",
			content:  "Services affected in Asia Pacific region",
			expected: "Asia Pacific",
		},
		{
			name:     "North America region",
			content:  "Issues reported in North America",
			expected: "North America",
		},
		{
			name:     "Multiple regions",
			content:  "Multiple regions experiencing issues",
			expected: "Multi-Region",
		},
		{
			name:     "United States",
			content:  "Services affected in United States",
			expected: "United States",
		},
		{
			name:     "United Kingdom",
			content:  "UK region experiencing issues",
			expected: "United Kingdom",
		},
		{
			name:     "Germany",
			content:  "German region affected",
			expected: "Germany",
		},
		{
			name:     "France",
			content:  "French region experiencing issues",
			expected: "France",
		},
		{
			name:     "Japan",
			content:  "Japanese region affected",
			expected: "Japan",
		},
		{
			name:     "Singapore",
			content:  "Singapore region issues",
			expected: "Singapore",
		},
		{
			name:     "Australia",
			content:  "Australian region affected",
			expected: "Australia",
		},
		{
			name:     "No region mentioned",
			content:  "Service experiencing general issues",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:       tt.title,
				Description: tt.content,
			}
			
			region := extractCloudflareRegion(item)
			assert.Equal(t, tt.expected, region)
		})
	}
}

func TestCloudflareParserSelection(t *testing.T) {
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
			name:     "Cloudflare status provider",
			provider: "cloudflare-status",
			service:  "status",
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
			default:
				scraperType = "unknown"
			}
			assert.Equal(t, tt.expected, scraperType)
		})
	}
}