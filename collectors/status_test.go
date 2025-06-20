package collectors

import (
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func TestExtractServiceStatus(t *testing.T) {
	tests := []struct {
		name           string
		title          string
		description    string
		content        string
		expectedState  string
		expectedActive bool
	}{
		{
			name:           "Standard resolved format",
			title:          "RESOLVED: Service issue fixed",
			expectedState:  "resolved",
			expectedActive: false,
		},
		{
			name:           "Genesys resolved format",
			title:          "WhatsApp Message Errors",
			description:    "Resolved - This incident has been resolved.",
			expectedState:  "resolved",
			expectedActive: false,
		},
		{
			name:           "HTML resolved format",
			title:          "Service Issue",
			content:        "<strong>Resolved</strong> - Service is now working normally.",
			expectedState:  "resolved",
			expectedActive: false,
		},
		{
			name:           "Detailed resolved message",
			title:          "Platform Issue",
			description:    "This incident has been resolved and all services are operational.",
			expectedState:  "resolved",
			expectedActive: false,
		},
		{
			name:           "Colon resolved format",
			title:          "System Status",
			description:    "Resolved: All systems operational",
			expectedState:  "resolved",
			expectedActive: false,
		},
		{
			name:           "Service outage",
			title:          "Major outage affecting services",
			expectedState:  "outage",
			expectedActive: true,
		},
		{
			name:           "Service issue investigating",
			title:          "Investigating connectivity issues",
			expectedState:  "service_issue",
			expectedActive: true,
		},
		{
			name:           "Service issue elevated errors",
			title:          "Platform Errors",
			description:    "We are experiencing elevated errors on the platform",
			expectedState:  "service_issue",
			expectedActive: true,
		},
		{
			name:           "Service issue monitoring",
			title:          "Service Monitoring",
			content:        "We are monitoring the situation closely",
			expectedState:  "service_issue",
			expectedActive: true,
		},
		{
			name:           "Maintenance notification - no status",
			title:          "Scheduled maintenance window",
			description:    "Scheduled maintenance will occur tonight",
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

			service, state, active := extractServiceStatus(item)

			assert.Equal(t, tt.expectedState, state, "State mismatch")
			assert.Equal(t, tt.expectedActive, active, "Active status mismatch")
			
			if tt.expectedState != "" {
				assert.NotEmpty(t, service, "Service should not be empty when state is detected")
			}
		})
	}
}