package collectors

import (
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func TestEnhancedGCPParser_ServiceInfo(t *testing.T) {
	parser := enhancedGCPParser{}

	tests := []struct {
		name            string
		title           string
		description     string
		content         string
		expectedService string
		expectedRegion  string
	}{
		{
			name:            "Compute Engine with specific region",
			title:           "UPDATE: Google Compute Engine experiencing elevated errors",
			description:     "Users in us-central1 may experience issues",
			expectedService: "compute-engine",
			expectedRegion:  "us-central1",
		},
		{
			name:            "Multiple services and regions",
			title:           "RESOLVED: Multiple GCP products are experiencing Service issues",
			description:     "Issues affecting Cloud Storage and BigQuery in us-west1 and europe-west1",
			expectedService: "multiple-services",
			expectedRegion:  "multiple-regions",
		},
		{
			name:            "BigQuery global issue",
			title:           "INVESTIGATING: BigQuery performance degradation",
			description:     "Global impact across all regions",
			expectedService: "bigquery",
			expectedRegion:  "global",
		},
		{
			name:            "Cloud Storage regional",
			title:           "UPDATE: Cloud Storage elevated latency",
			description:     "Impact limited to europe-west3 region",
			expectedService: "cloud-storage",
			expectedRegion:  "europe-west3",
		},
		{
			name:            "Kubernetes Engine",
			title:           "UPDATE: GKE cluster creation failures",
			description:     "Google Kubernetes Engine experiencing issues in asia-southeast1",
			expectedService: "kubernetes-engine",
			expectedRegion:  "asia-southeast1",
		},
		{
			name:            "No clear service or region",
			title:           "RESOLVED: Network connectivity restored",
			description:     "All services back to normal operation",
			expectedService: "",
			expectedRegion:  "",
		},
		{
			name:            "Cloud Run with multiple regions",
			title:           "INVESTIGATING: Cloud Run deployment failures",
			description:     "Affecting us-central1, us-east1, and us-west1 regions",
			expectedService: "cloud-run",
			expectedRegion:  "multiple-regions",
		},
		{
			name:            "Vertex AI",
			title:           "UPDATE: Vertex AI Online Prediction elevated latency",
			description:     "Users may experience delays in model predictions",
			expectedService: "vertex-ai",
			expectedRegion:  "",
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

func TestEnhancedGCPParser_IncidentKey(t *testing.T) {
	parser := enhancedGCPParser{}

	tests := []struct {
		name        string
		link        string
		guid        string
		title       string
		expectedKey string
	}{
		{
			name:        "Incident URL available",
			link:        "https://status.cloud.google.com/incidents/abc123",
			guid:        "tag:status.cloud.google.com,2025:feed:abc123.def456",
			title:       "RESOLVED: Multiple GCP products issue",
			expectedKey: "https://status.cloud.google.com/incidents/abc123",
		},
		{
			name:        "No incident URL, use GUID",
			link:        "https://status.cloud.google.com/",
			guid:        "tag:status.cloud.google.com,2025:feed:xyz789.abc123",
			title:       "UPDATE: Service disruption ongoing",
			expectedKey: "tag:status.cloud.google.com,2025:feed:xyz789.abc123",
		},
		{
			name:        "No GUID, use title",
			link:        "https://status.cloud.google.com/",
			guid:        "",
			title:       "INVESTIGATING: Cloud Storage issues",
			expectedKey: "INVESTIGATING: Cloud Storage issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Link:  tt.link,
				GUID:  tt.guid,
				Title: tt.title,
			}

			key := parser.IncidentKey(item)
			assert.Equal(t, tt.expectedKey, key)
		})
	}
}

func TestExtractGCPService(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedService string
	}{
		{
			name:            "Google Compute Engine",
			content:         "Google Compute Engine experiencing issues",
			expectedService: "compute-engine",
		},
		{
			name:            "BigQuery mention",
			content:         "BigQuery queries are timing out",
			expectedService: "bigquery",
		},
		{
			name:            "Multiple services indicator",
			content:         "Multiple GCP products are experiencing service disruption",
			expectedService: "multiple-services",
		},
		{
			name:            "Cloud Storage",
			content:         "Cloud Storage upload failures reported",
			expectedService: "cloud-storage",
		},
		{
			name:            "GKE abbreviation",
			content:         "GKE cluster provisioning delayed",
			expectedService: "kubernetes-engine",
		},
		{
			name:            "No recognizable service",
			content:         "General platform maintenance completed",
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
			service := extractGCPService(item)
			assert.Equal(t, tt.expectedService, service)
		})
	}
}

func TestExtractGCPRegion(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		expectedRegion string
	}{
		{
			name:           "Single US region",
			content:        "Issues affecting us-central1 region",
			expectedRegion: "us-central1",
		},
		{
			name:           "Multiple regions",
			content:        "Problems in us-west1 and europe-west1",
			expectedRegion: "multiple-regions",
		},
		{
			name:           "Global impact",
			content:        "Global service disruption affecting all regions",
			expectedRegion: "global",
		},
		{
			name:           "Asia Pacific region",
			content:        "Service restored in asia-southeast1",
			expectedRegion: "asia-southeast1",
		},
		{
			name:           "No region mentioned",
			content:        "Service performance improved",
			expectedRegion: "",
		},
		{
			name:           "European region",
			content:        "Users in europe-north1 experiencing delays",
			expectedRegion: "europe-north1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:       tt.content,
				Description: "",
				Content:     "",
			}
			region := extractGCPRegion(item)
			assert.Equal(t, tt.expectedRegion, region)
		})
	}
}