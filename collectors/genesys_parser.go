package collectors

import (
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"
)

// Genesys Cloud parser for extracting service and region information
type genesysParser struct{}

// ServiceInfo extracts Genesys Cloud service name and region from feed items
func (genesysParser) ServiceInfo(item *gofeed.Item) (string, string) {
	serviceName := extractGenesysService(item)
	region := extractGenesysRegion(item)
	return serviceName, region
}

// IncidentKey returns a stable identifier for Genesys Cloud incidents
func (genesysParser) IncidentKey(item *gofeed.Item) string {
	if item.GUID != "" {
		return item.GUID
	}
	if item.Link != "" {
		return item.Link
	}
	return strings.TrimSpace(item.Title)
}

// extractGenesysService attempts to identify the Genesys Cloud service affected
func extractGenesysService(item *gofeed.Item) string {
	content := strings.ToLower(item.Title + " " + item.Description + " " + item.Content)
	
	// Define Genesys Cloud service patterns (ordered by specificity)
	servicePatterns := []struct {
		name     string
		patterns []string
	}{
		// AI/Integration Services (most specific first)
		{"text-to-speech", []string{"text to speech", "tts"}},
		{"speech-to-text", []string{"speech to text", "stt"}},
		{"dialogflow-integration", []string{"dialogflow es/cx bot integrations", "dialogflow integration", "dialogflow"}},
		{"whatsapp-integration", []string{"whatsapp message", "whatsapp integration", "whatsapp"}},
		
		// Contact Center Core Services
		{"identity-access", []string{"identity & access management", "identity and access", "authentication", "login"}},
		{"inbound-calls", []string{"inbound calls", "inbound calling"}},
		{"outbound-calls", []string{"outbound calls", "outbound calling", "outbound dialing"}},
		{"ivr", []string{"ivr", "interactive voice response"}},
		{"acd-routing", []string{"acd routing", "automatic call distribution", "call routing"}},
		{"web-messaging", []string{"web messaging", "messaging"}},
		{"chat", []string{"chat"}},
		{"email", []string{"email"}},
		{"voice", []string{"voice", "soft phone", "softphone"}},
		{"workforce-engagement", []string{"workforce engagement", "wem"}},
		
		// Supporting Systems
		{"analytics", []string{"analytics", "reporting"}},
		{"recording", []string{"recording", "quality management"}},
		{"data-sync", []string{"data sync integrations", "data sync"}},
		{"directory", []string{"directory"}},
		{"documents", []string{"documents"}},
		{"fax", []string{"fax"}},
		{"video", []string{"video"}},
		{"co-browse", []string{"co-browse", "cobrowse"}},
		{"agent-copilot", []string{"agent copilot"}},
		
		// Technical Components
		{"call-notifications", []string{"call notification", "notifications"}},
		{"connectivity", []string{"connectivity"}},
		{"instances", []string{"instances", "instance launch"}},
		
		// Platform/Regional Services
		{"platform", []string{"platform"}},
		{"gmf", []string{"global media fabric", "gmf"}},
	}
	
	// Check for specific service mentions
	for _, service := range servicePatterns {
		for _, pattern := range service.patterns {
			if strings.Contains(content, pattern) {
				return service.name
			}
		}
	}
	
	// Check for "elevated error rates" pattern without specific service
	if strings.Contains(content, "elevated error rates") {
		return "elevated-errors"
	}
	
	return ""
}

// extractGenesysRegion attempts to identify the affected Genesys Cloud region(s)
func extractGenesysRegion(item *gofeed.Item) string {
	content := item.Title + " " + item.Description + " " + item.Content
	
	// Genesys Cloud regional patterns (matches their exact naming)
	regionPatterns := []string{
		// Americas regions
		`Americas \(US East\)`, `Americas \(US West\)`, `Americas \(Canada\)`, 
		`Americas \(Sao Paulo\)`, `Americas \(São Paulo\)`,
		
		// EMEA regions  
		`EMEA \(Frankfurt\)`, `EMEA \(Ireland\)`, `EMEA \(London\)`, `EMEA \(UAE\)`,
		
		// APAC regions
		`Asia Pacific \(Singapore\)`, `Asia Pacific \(Sydney\)`, 
		`Asia Pacific \(Tokyo\)`, `Asia Pacific \(Seoul\)`, `Asia Pacific \(Mumbai\)`,
		
		// Simplified regional patterns (backup)
		`US East`, `US West`, `Canada`, `Sao Paulo`, `São Paulo`,
		`Frankfurt`, `Ireland`, `London`, `UAE`,
		`Singapore`, `Sydney`, `Tokyo`, `Seoul`, `Mumbai`,
		
		// AWS region patterns (underlying infrastructure)
		`us-east-1`, `us-east-2`, `us-west-2`, `ca-central-1`, `sa-east-1`,
		`eu-central-1`, `eu-west-1`, `eu-west-2`, `me-central-1`,
		`ap-southeast-1`, `ap-southeast-2`, `ap-northeast-1`, `ap-northeast-2`, `ap-south-1`,
	}
	
	// Compile regex for region detection
	regionRegex := regexp.MustCompile(strings.Join(regionPatterns, "|"))
	
	// Find region matches
	matches := regionRegex.FindAllString(content, -1)
	if len(matches) > 0 {
		// Check for multiple different regions
		uniqueRegions := make(map[string]bool)
		for _, match := range matches {
			// Normalize region names for comparison
			normalized := strings.ToLower(strings.ReplaceAll(match, " ", "-"))
			uniqueRegions[normalized] = true
		}
		
		if len(uniqueRegions) > 1 {
			return "multiple-regions"
		}
		// Return the first match (most specific)
		return matches[0]
	}
	
	// Check for global/platform-wide indicators
	if strings.Contains(strings.ToLower(content), "global") || 
	   strings.Contains(strings.ToLower(content), "all regions") ||
	   strings.Contains(strings.ToLower(content), "platform") {
		return "global"
	}
	
	return ""
}

// extractGenesysStatus provides Genesys Cloud-specific status detection
func extractGenesysStatus(item *gofeed.Item) (service string, state string, active bool) {
	content := strings.ToUpper(item.Title + " " + item.Description + " " + item.Content)
	
	// Look for HTML status tags first (most reliable)
	if strings.Contains(content, "<STRONG>RESOLVED</STRONG>") {
		state = "resolved"
	} else if strings.Contains(content, "<STRONG>INVESTIGATING</STRONG>") {
		state = "service_issue"
	} else if strings.Contains(content, "<STRONG>UPDATE</STRONG>") {
		state = "service_issue"  
	} else if strings.Contains(content, "<STRONG>MONITORING</STRONG>") {
		state = "service_issue"
	} else {
		// Fallback to text-based detection
		switch {
		case strings.Contains(content, "RESOLVED"):
			state = "resolved"
		case strings.Contains(content, "OUTAGE") || strings.Contains(content, "MAJOR OUTAGE"):
			state = "outage"
		case strings.Contains(content, "ELEVATED ERROR RATES") ||
			 strings.Contains(content, "ERRORS") ||
			 strings.Contains(content, "ISSUES") ||
			 strings.Contains(content, "DEGRADED") ||
			 strings.Contains(content, "PARTIAL OUTAGE") ||
			 strings.Contains(content, "INVESTIGATING") ||
			 strings.Contains(content, "MONITORING"):
			state = "service_issue"
		}
	}
	
	if state == "" {
		return
	}
	
	service = strings.TrimSpace(item.Title)
	active = state != "resolved"
	return
}