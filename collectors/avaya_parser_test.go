package collectors

import (
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func TestEnhancedAvayaParser_ServiceInfo(t *testing.T) {
	parser := enhancedAvayaParser{}
	
	tests := []struct {
		name            string
		title           string
		description     string
		guid            string
		expectedService string
		expectedRegion  string
	}{
		{
			name:            "AXP service with North America region",
			title:           "Avaya Experience Platform - Service Outage",
			description:     "We are investigating a service outage affecting AXP in the North America region.",
			expectedService: "Avaya Experience Platform",
			expectedRegion:  "North America",
		},
		{
			name:            "Preview Dialing maintenance with multiple regions",
			title:           "AXP Preview Dialing Scheduled Maintenance",
			description:     "Scheduled maintenance for Preview Dialing across multiple regions including Prod-NA and Prod-EU.",
			expectedService: "Preview Dialing",
			expectedRegion:  "Multi-Region",
		},
		{
			name:            "Contact Center issue in Asia Pacific",
			title:           "Contact Center - Performance Degradation",
			description:     "Contact Center services are experiencing performance issues in the Asia Pacific region.",
			expectedService: "Contact Center",
			expectedRegion:  "Asia Pacific",
		},
		{
			name:            "Avaya Cloud Office global issue",
			title:           "Avaya Cloud Office - Authentication Issues",
			description:     "ACO authentication services are affected globally across all regions.",
			expectedService: "Avaya Cloud Office",
			expectedRegion:  "Global",
		},
		{
			name:            "Communications APIs UK issue",
			title:           "Communications APIs - Rate Limiting",
			description:     "CPaaS rate limiting issues affecting customers in the United Kingdom region.",
			expectedService: "Communications APIs",
			expectedRegion:  "United Kingdom",
		},
		{
			name:            "Voice Services in Australia",
			title:           "Voice Platform - Call Quality Issues",
			description:     "Voice services experiencing call quality degradation in Australia & New Zealand region.",
			expectedService: "Voice Services",
			expectedRegion:  "Australia & New Zealand",
		},
		{
			name:            "API Gateway Europe issue",
			title:           "Avaya API Gateway - Connection Timeouts",
			description:     "API Gateway experiencing connection timeouts in the Europe region.",
			expectedService: "Avaya API Gateway",
			expectedRegion:  "Europe",
		},
		{
			name:            "IVR service Canada",
			title:           "Interactive Voice Response - Menu Issues",
			description:     "IVR menu navigation issues reported in Canada region.",
			expectedService: "Interactive Voice Response",
			expectedRegion:  "Canada",
		},
		{
			name:            "Telephony platform Japan",
			title:           "PBX Services - Registration Failures",
			description:     "PBX registration failures affecting telephony platform in Japan region.",
			expectedService: "PBX Services",
			expectedRegion:  "Japan",
		},
		{
			name:            "Generic Avaya platform issue",
			title:           "Platform Connectivity Issue",
			description:     "General connectivity issues affecting Avaya cloud services.",
			expectedService: "Avaya Cloud Platform",
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

func TestEnhancedAvayaParser_IncidentKey(t *testing.T) {
	parser := enhancedAvayaParser{}
	
	tests := []struct {
		name        string
		guid        string
		link        string
		title       string
		expectedKey string
	}{
		{
			name:        "GUID URL format",
			guid:        "https://status.avayacloud.com/incidents/888hk0hhs4xc",
			expectedKey: "https://status.avayacloud.com/incidents/888hk0hhs4xc",
		},
		{
			name:        "Different incident ID",
			guid:        "https://status.avayacloud.com/incidents/xyz123abc456",
			expectedKey: "https://status.avayacloud.com/incidents/xyz123abc456",
		},
		{
			name:        "No GUID, use link",
			link:        "https://status.avayacloud.com/incidents/fallback",
			title:       "Service Issue",
			expectedKey: "https://status.avayacloud.com/incidents/fallback",
		},
		{
			name:        "No GUID or link, use title",
			title:       "Avaya Experience Platform Service Outage",
			expectedKey: "Avaya Experience Platform Service Outage",
		},
		{
			name:        "Empty GUID, fallback to link",
			guid:        "",
			link:        "https://status.avayacloud.com/",
			title:       "Platform Issue",
			expectedKey: "https://status.avayacloud.com/",
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

func TestExtractAvayaService(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		content  string
		expected string
	}{
		{
			name:     "Avaya Experience Platform",
			content:  "Avaya Experience Platform services experiencing issues",
			expected: "Avaya Experience Platform",
		},
		{
			name:     "Avaya Enterprise Cloud",
			content:  "AEC platform maintenance required",
			expected: "Avaya Enterprise Cloud",
		},
		{
			name:     "Avaya Cloud Office",
			content:  "ACO authentication problems",
			expected: "Avaya Cloud Office",
		},
		{
			name:     "Communications APIs",
			content:  "CPaaS rate limiting issues",
			expected: "Communications APIs",
		},
		{
			name:     "Contact Center services",
			content:  "Contact center platform degradation",
			expected: "Contact Center",
		},
		{
			name:     "Preview Dialing",
			content:  "Preview dialing campaign issues",
			expected: "Preview Dialing",
		},
		{
			name:     "Voice Services",
			content:  "Voice platform connectivity problems",
			expected: "Voice Services",
		},
		{
			name:     "Interactive Voice Response",
			content:  "IVR menu navigation failures",
			expected: "Interactive Voice Response",
		},
		{
			name:     "Video Conferencing",
			content:  "Video calls experiencing quality issues",
			expected: "Video Conferencing",
		},
		{
			name:     "Unified Communications",
			content:  "UC platform service disruption",
			expected: "Unified Communications",
		},
		{
			name:     "Title format extraction",
			title:    "Preview Dialing - Scheduled Maintenance",
			expected: "Preview Dialing",
		},
		{
			name:     "PBX Services",
			content:  "PBX registration failures reported",
			expected: "PBX Services",
		},
		{
			name:     "Call Analytics",
			content:  "Call analytics reporting delays",
			expected: "Call Analytics",
		},
		{
			name:     "No recognizable service",
			content:  "Platform wide connectivity issues",
			expected: "Avaya Cloud Platform",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:       tt.title,
				Description: tt.content,
			}
			
			service := extractAvayaService(item)
			assert.Equal(t, tt.expected, service)
		})
	}
}

func TestExtractAvayaRegion(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "North America region",
			content:  "Issues affecting services in North America",
			expected: "North America",
		},
		{
			name:     "Production North America",
			content:  "Prod-NA environment experiencing issues",
			expected: "North America",
		},
		{
			name:     "Europe region", 
			content:  "Services in Europe region are affected",
			expected: "Europe",
		},
		{
			name:     "Production Europe",
			content:  "Prod-EU maintenance scheduled",
			expected: "Europe",
		},
		{
			name:     "Asia Pacific region",
			content:  "APAC region services degraded",
			expected: "Asia Pacific",
		},
		{
			name:     "Production Asia",
			content:  "Prod-ASE region maintenance",
			expected: "Asia Pacific",
		},
		{
			name:     "United Kingdom",
			content:  "UK region specific issues",
			expected: "United Kingdom",
		},
		{
			name:     "Australia & New Zealand",
			content:  "ANZ region connectivity problems",
			expected: "Australia & New Zealand",
		},
		{
			name:     "Canada region",
			content:  "Canadian region service disruption",
			expected: "Canada",
		},
		{
			name:     "Japan region",
			content:  "Japan region experiencing delays",
			expected: "Japan",
		},
		{
			name:     "India region",
			content:  "India region performance issues",
			expected: "India",
		},
		{
			name:     "Global issue",
			content:  "Global connectivity issues affecting all regions",
			expected: "Global",
		},
		{
			name:     "Multiple regions",
			content:  "Multiple regions affected including NA and EU",
			expected: "Multi-Region",
		},
		{
			name:     "South America",
			content:  "South America region maintenance",
			expected: "South America",
		},
		{
			name:     "Staging environment",
			content:  "Staging environment deployment issues",
			expected: "Staging",
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
				Title:       tt.content,
				Description: "",
			}
			
			region := extractAvayaRegion(item)
			assert.Equal(t, tt.expected, region)
		})
	}
}

func TestFormatAvayaServiceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"axp", "Avaya Experience Platform"},
		{"aec", "Avaya Enterprise Cloud"},
		{"aco", "Avaya Cloud Office"},
		{"cpaas", "Communications APIs"},
		{"contact center", "Contact Center"},
		{"preview dialing", "Preview Dialing"},
		{"voice services", "Voice Services"},
		{"api gateway", "Avaya API Gateway"},
		{"ivr", "Avaya IVR"},
		{"pbx services", "Avaya PBX Services"},
		{"sip trunking", "Avaya SIP Trunking"},
		{"unified communications", "Unified Communications"},
		{"", "Avaya Cloud Platform"},
		{"unknown service", "Avaya Unknown Service"},
		{"Avaya Experience Platform", "Avaya Experience Platform"}, // Already formatted
		{"telephony platform", "Telephony Platform"},
		{"collaboration platform", "Collaboration Platform"},
		{"analytics platform", "Analytics Platform"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatAvayaServiceName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsAvayaBranding(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"Avaya Experience Platform", true},
		{"Contact Center", true},
		{"AXP Service", true},
		{"Communications APIs", true},
		{"Unified Communications", true},
		{"Generic Service", false},
		{"Unknown Platform", false},
		{"Voice Services", false},
		{"API Gateway", false},
		{"Telephony Platform", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := containsAvayaBranding(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}