package collectors

import (
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"
)

// Enhanced GCP parser with service and region extraction capabilities
type enhancedGCPParser struct{}

// ServiceInfo extracts GCP service name and region from feed items
func (enhancedGCPParser) ServiceInfo(item *gofeed.Item) (string, string) {
	serviceName := extractGCPService(item)
	region := extractGCPRegion(item)
	return serviceName, region
}

// IncidentKey returns a stable identifier for GCP incidents
func (enhancedGCPParser) IncidentKey(item *gofeed.Item) string {
	// Use incident URL when available (most reliable)
	if strings.Contains(item.Link, "status.cloud.google.com/incidents/") {
		return item.Link
	}
	if item.GUID != "" {
		return item.GUID
	}
	return item.Title
}

// extractGCPService attempts to identify the primary GCP service affected
func extractGCPService(item *gofeed.Item) string {
	content := strings.ToLower(item.Title + " " + item.Description + " " + item.Content)
	
	// Check for generic "multiple products" indicators first (highest priority)
	if strings.Contains(content, "multiple") && (strings.Contains(content, "products") || strings.Contains(content, "services")) {
		return "multiple-services"
	}
	
	// Define GCP service patterns (ordered by specificity - most specific first)
	servicePatterns := []struct {
		name     string
		patterns []string
	}{
		// Specific product names first
		{"compute-engine", []string{"google compute engine", "compute engine"}},
		{"kubernetes-engine", []string{"google kubernetes engine", "kubernetes engine", "gke"}},
		{"cloud-storage", []string{"cloud storage", "google cloud storage"}},
		{"bigquery", []string{"bigquery"}},
		{"cloud-run", []string{"cloud run"}},
		{"app-engine", []string{"app engine", "google app engine"}},
		{"cloud-functions", []string{"cloud functions"}},
		{"persistent-disk", []string{"persistent disk"}},
		{"cloud-sql", []string{"cloud sql"}},
		{"cloud-spanner", []string{"cloud spanner", "spanner"}},
		{"cloud-bigtable", []string{"cloud bigtable", "bigtable"}},
		{"cloud-dataflow", []string{"cloud dataflow", "dataflow"}},
		{"cloud-dataproc", []string{"cloud dataproc", "dataproc"}},
		{"cloud-pubsub", []string{"cloud pub/sub", "pubsub", "pub/sub"}},
		{"cloud-datastore", []string{"cloud datastore", "datastore"}},
		{"vertex-ai", []string{"vertex ai"}},
		{"automl", []string{"automl"}},
		{"cloud-vision", []string{"cloud vision api", "vision api"}},
		{"cloud-speech", []string{"cloud speech api", "speech api"}},
		{"cloud-translation", []string{"cloud translation api", "translation api"}},
		{"cloud-trace", []string{"cloud trace", "trace"}},
		{"error-reporting", []string{"error reporting"}},
		{"cloud-iam", []string{"cloud iam", "iam"}},
		{"cloud-kms", []string{"cloud kms", "key management"}},
		{"security-command-center", []string{"security command center"}},
		{"cloud-vpn", []string{"cloud vpn", "vpn"}},
		{"cloud-load-balancing", []string{"cloud load balancing", "load balancing"}},
		{"cloud-cdn", []string{"cloud cdn", "cdn"}},
		{"cloud-dns", []string{"cloud dns"}},
		// Generic terms last (least specific)
		{"cloud-networking", []string{"cloud networking"}},
		{"cloud-monitoring", []string{"cloud monitoring"}}, 
		{"cloud-logging", []string{"cloud logging"}},
	}
	
	// Check for specific service mentions
	for _, service := range servicePatterns {
		for _, pattern := range service.patterns {
			if strings.Contains(content, pattern) {
				return service.name
			}
		}
	}
	
	return ""
}

// extractGCPRegion attempts to identify the affected GCP region(s)
func extractGCPRegion(item *gofeed.Item) string {
	content := strings.ToLower(item.Title + " " + item.Description + " " + item.Content)
	
	// GCP region patterns
	regionPatterns := []string{
		// Americas
		`us-central[1-4]`, `us-east[1-4]`, `us-west[1-4]`, `us-south1`,
		`northamerica-northeast[1-2]`, `southamerica-east1`,
		
		// Europe  
		`europe-west[1-9]`, `europe-north1`, `europe-central2`,
		
		// Asia Pacific
		`asia-southeast[1-2]`, `asia-northeast[1-3]`, `asia-south1`, `asia-east[1-2]`,
		`australia-southeast[1-2]`,
		
		// Multi-regions
		`\bus\b`, `\beu\b`, `\basia\b`,
		
		// Region descriptions (backup patterns)
		`iowa`, `oregon`, `virginia`, `london`, `frankfurt`, `singapore`, `tokyo`,
	}
	
	// Compile regex for region detection
	regionRegex := regexp.MustCompile(strings.Join(regionPatterns, "|"))
	
	// Find region matches
	matches := regionRegex.FindAllString(content, -1)
	if len(matches) > 0 {
		// Return first match, or "multiple" if multiple regions
		if len(matches) > 1 {
			// Check if all matches are the same
			first := matches[0]
			allSame := true
			for _, match := range matches[1:] {
				if match != first {
					allSame = false
					break
				}
			}
			if !allSame {
				return "multiple-regions"
			}
		}
		return matches[0]
	}
	
	// Check for global/worldwide indicators
	if strings.Contains(content, "global") || strings.Contains(content, "worldwide") || strings.Contains(content, "all regions") {
		return "global"
	}
	
	return ""
}

// extractGCPStatus provides GCP-specific status detection
func extractGCPStatus(item *gofeed.Item) (service string, state string, active bool) {
	content := strings.ToUpper(item.Title + " " + item.Description + " " + item.Content)
	
	switch {
	case strings.Contains(content, "RESOLVED"):
		state = "resolved"
	case strings.Contains(content, "SERVICE_OUTAGE") || strings.Contains(content, "OUTAGE"):
		state = "outage"
	case strings.Contains(content, "SERVICE_DISRUPTION") || 
		 strings.Contains(content, "SERVICE IMPACT") ||
		 strings.Contains(content, "EXPERIENCING") ||
		 strings.Contains(content, "ELEVATED ERRORS"):
		state = "service_issue"
	case strings.Contains(content, "INVESTIGATING") || strings.Contains(content, "MONITORING"):
		state = "service_issue"
	}
	
	if state == "" {
		return
	}
	
	service = strings.TrimSpace(item.Title)
	active = state != "resolved"
	return
}