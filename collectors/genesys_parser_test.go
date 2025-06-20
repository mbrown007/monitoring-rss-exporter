package collectors

import (
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func TestGenesysParser_ServiceInfo(t *testing.T) {
	parser := genesysParser{}

	tests := []struct {
		name            string
		title           string
		description     string
		content         string
		expectedService string
		expectedRegion  string
	}{
		{
			name:            "TTS/STT with specific region",
			title:           "Elevated Error Rates: Text to Speech, Speech to Text, and Dialogflow ES/CX bot Integrations",
			description:     "Issues affecting Americas (US East) region",
			expectedService: "text-to-speech", // First match wins
			expectedRegion:  "Americas (US East)",
		},
		{
			name:            "WhatsApp integration multiple regions",
			title:           "WhatsApp Message Errors",
			description:     "Affecting Americas (US East) and Americas (Sao Paulo)",
			expectedService: "whatsapp-integration",
			expectedRegion:  "multiple-regions",
		},
		{
			name:            "Analytics delays",
			title:           "Analytics Delays in EMEA (Frankfurt)",
			description:     "Users experiencing delays in analytics reporting",
			expectedService: "analytics",
			expectedRegion:  "EMEA (Frankfurt)",
		},
		{
			name:            "Voice/call issues",
			title:           "Inbound Calls Issues",
			description:     "Problems with inbound calling in Asia Pacific (Singapore)",
			expectedService: "inbound-calls",
			expectedRegion:  "Asia Pacific (Singapore)",
		},
		{
			name:            "Platform-wide issue",
			title:           "Connectivity Issues",
			description:     "Global platform connectivity problems",
			expectedService: "connectivity",
			expectedRegion:  "global",
		},
		{
			name:            "IVR specific issue",
			title:           "Interactive Voice Response Errors",
			description:     "IVR system experiencing issues in US West",
			expectedService: "ivr",
			expectedRegion:  "US West",
		},
		{
			name:            "Call notification issue",
			title:           "Call Notification Issues in EMEA (London)",
			description:     "Users not receiving call notifications",
			expectedService: "call-notifications",
			expectedRegion:  "EMEA (London)",
		},
		{
			name:            "Elevated errors without specific service",
			title:           "Elevated Error Rates: Americas (US East)",
			description:     "General elevated error rates in region",
			expectedService: "elevated-errors",
			expectedRegion:  "Americas (US East)",
		},
		{
			name:            "No clear service or region",
			title:           "Maintenance Complete",
			description:     "Scheduled maintenance has been completed",
			expectedService: "",
			expectedRegion:  "",
		},
		{
			name:            "Workforce engagement issue",
			title:           "Workforce Engagement Issues",
			description:     "WEM experiencing problems in Americas (US East) and EMEA (Frankfurt)",
			expectedService: "workforce-engagement",
			expectedRegion:  "multiple-regions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:       tt.title,
				Description: tt.description,
				Content:     tt.content,
			}

			serviceName, region := parser.ServiceInfo(item)
			assert.Equal(t, tt.expectedService, serviceName, "Service name mismatch")
			assert.Equal(t, tt.expectedRegion, region, "Region mismatch")
		})
	}
}

func TestGenesysParser_IncidentKey(t *testing.T) {
	parser := genesysParser{}

	tests := []struct {
		name        string
		guid        string
		link        string
		title       string
		expectedKey string
	}{
		{
			name:        "GUID available",
			guid:        "genesys-incident-123",
			link:        "https://status.mypurecloud.com/incident/123",
			title:       "Service Issue",
			expectedKey: "genesys-incident-123",
		},
		{
			name:        "No GUID, use link",
			guid:        "",
			link:        "https://status.mypurecloud.com/incident/456",
			title:       "Another Issue",
			expectedKey: "https://status.mypurecloud.com/incident/456",
		},
		{
			name:        "No GUID or link, use title",
			guid:        "",
			link:        "",
			title:       "Elevated Error Rates: TTS Issues",
			expectedKey: "Elevated Error Rates: TTS Issues",
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

func TestExtractGenesysService(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedService string
	}{
		{
			name:            "Text to Speech",
			content:         "Text to Speech experiencing issues",
			expectedService: "text-to-speech",
		},
		{
			name:            "TTS abbreviation",
			content:         "TTS service degraded",
			expectedService: "text-to-speech",
		},
		{
			name:            "Speech to Text",
			content:         "Speech to Text errors reported",
			expectedService: "speech-to-text",
		},
		{
			name:            "STT abbreviation",
			content:         "STT processing delays",
			expectedService: "speech-to-text",
		},
		{
			name:            "Dialogflow integration",
			content:         "Dialogflow ES/CX bot integrations failing",
			expectedService: "dialogflow-integration",
		},
		{
			name:            "WhatsApp integration",
			content:         "WhatsApp message delivery issues",
			expectedService: "whatsapp-integration",
		},
		{
			name:            "Analytics",
			content:         "Analytics reporting delayed",
			expectedService: "analytics",
		},
		{
			name:            "Voice service",
			content:         "Voice quality issues reported",
			expectedService: "voice",
		},
		{
			name:            "Inbound calls",
			content:         "Inbound calls not connecting",
			expectedService: "inbound-calls",
		},
		{
			name:            "Outbound calls",
			content:         "Outbound dialing problems",
			expectedService: "outbound-calls",
		},
		{
			name:            "IVR issues",
			content:         "Interactive Voice Response not working",
			expectedService: "ivr",
		},
		{
			name:            "Elevated error rates generic",
			content:         "Elevated error rates detected",
			expectedService: "elevated-errors",
		},
		{
			name:            "No recognizable service",
			content:         "General system maintenance",
			expectedService: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:       tt.content,
				Description: "",
				Content:     "",
			}
			service := extractGenesysService(item)
			assert.Equal(t, tt.expectedService, service)
		})
	}
}

func TestExtractGenesysRegion(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedRegion string
	}{
		{
			name:           "Americas US East",
			content:        "Issues in Americas (US East)",
			expectedRegion: "Americas (US East)",
		},
		{
			name:           "Americas US West",
			content:        "Problems affecting Americas (US West)",
			expectedRegion: "Americas (US West)",
		},
		{
			name:           "EMEA Frankfurt",
			content:        "Service degradation in EMEA (Frankfurt)",
			expectedRegion: "EMEA (Frankfurt)",
		},
		{
			name:           "EMEA London",
			content:        "Users in EMEA (London) experiencing issues",
			expectedRegion: "EMEA (London)",
		},
		{
			name:           "Asia Pacific Singapore",
			content:        "Connectivity issues in Asia Pacific (Singapore)",
			expectedRegion: "Asia Pacific (Singapore)",
		},
		{
			name:           "Multiple regions with 'and'",
			content:        "Affecting Americas (US East) and Americas (Sao Paulo)",
			expectedRegion: "multiple-regions",
		},
		{
			name:           "Multiple different regions",
			content:        "Issues in EMEA (Frankfurt) and Asia Pacific (Tokyo)",
			expectedRegion: "multiple-regions",
		},
		{
			name:           "Global platform issue",
			content:        "Global platform disruption",
			expectedRegion: "global",
		},
		{
			name:           "All regions affected",
			content:        "Service issues affecting all regions",
			expectedRegion: "global",
		},
		{
			name:           "AWS region format",
			content:        "Problems in us-east-1 region",
			expectedRegion: "us-east-1",
		},
		{
			name:           "No region mentioned",
			content:        "Service restored successfully",
			expectedRegion: "",
		},
		{
			name:           "Simplified region name",
			content:        "Issues reported in Frankfurt",
			expectedRegion: "Frankfurt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:       tt.content,
				Description: "",
				Content:     "",
			}
			region := extractGenesysRegion(item)
			assert.Equal(t, tt.expectedRegion, region)
		})
	}
}

func TestExtractGenesysStatus(t *testing.T) {
	tests := []struct {
		name           string
		title          string
		description    string
		content        string
		expectedState  string
		expectedActive bool
	}{
		{
			name:           "HTML Resolved status",
			title:          "Service Issue",
			description:    "",
			content:        "<strong>Resolved</strong> - This incident has been resolved.",
			expectedState:  "resolved",
			expectedActive: false,
		},
		{
			name:           "HTML Investigating status",
			title:          "TTS Issues",
			description:    "",
			content:        "<strong>Investigating</strong> - We are investigating the issue.",
			expectedState:  "service_issue",
			expectedActive: true,
		},
		{
			name:           "HTML Update status", 
			title:          "Ongoing Issue",
			description:    "",
			content:        "<strong>Update</strong> - Continue monitoring the situation.",
			expectedState:  "service_issue",
			expectedActive: true,
		},
		{
			name:           "Text-based elevated errors",
			title:          "Elevated Error Rates: Analytics",
			description:    "Users experiencing elevated error rates",
			content:        "",
			expectedState:  "service_issue",
			expectedActive: true,
		},
		{
			name:           "Text-based outage",
			title:          "Major Outage: Voice Services",
			description:    "Complete service outage",
			content:        "",
			expectedState:  "outage",
			expectedActive: true,
		},
		{
			name:           "Text-based resolved",
			title:          "RESOLVED: Service issues",
			description:    "All services restored",
			content:        "",
			expectedState:  "resolved",
			expectedActive: false,
		},
		{
			name:           "No clear status",
			title:          "Maintenance Scheduled",
			description:    "Routine maintenance planned",
			content:        "",
			expectedState:  "",
			expectedActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:       tt.title,
				Description: tt.description,
				Content:     tt.content,
			}
			
			_, state, active := extractGenesysStatus(item)
			assert.Equal(t, tt.expectedState, state, "State mismatch")
			assert.Equal(t, tt.expectedActive, active, "Active status mismatch")
		})
	}
}