package collectors

import (
	"strings"

	"github.com/mmcdole/gofeed"
)

// Scraper extracts provider-specific information from a feed item and
// also provides a deduplication key used to filter repeated entries.
type Scraper interface {
	// ServiceInfo returns the service name and region associated with the item.
	ServiceInfo(item *gofeed.Item) (serviceName, region string)
	// IncidentKey returns a stable identifier for the incident represented
	// by this item. Items with the same key will be deduplicated.
	IncidentKey(item *gofeed.Item) string
}

type genericParser struct{}

func (genericParser) ServiceInfo(item *gofeed.Item) (string, string) {
	return "", ""
}

func (genericParser) IncidentKey(item *gofeed.Item) string {
	if item.GUID != "" {
		return item.GUID
	}
	if item.Link != "" {
		return item.Link
	}
	return strings.TrimSpace(item.Title)
}

// ScraperForService selects a scraper based on the provider or service name.
func ScraperForService(provider, service string) Scraper {
	p := strings.ToLower(provider)
	switch p {
	case "aws":
		return enhancedAWSParser{}
	case "gcp":
		return enhancedGCPParser{}
	case "azure":
		return enhancedAzureParser{}
	case "genesyscloud", "genesys-cloud":
		return genesysParser{}
	case "avaya", "avayacloud", "avaya-cloud":
		return enhancedAvayaParser{}
	case "cloudflare", "cloudflare-status":
		return enhancedCloudflareParser{}
	case "":
		// fall back to service name when provider not set
	default:
		if p != "" {
			return genericParser{}
		}
	}

	svc := strings.ToLower(service)
	switch {
	case strings.Contains(svc, "aws"):
		return enhancedAWSParser{}
	case strings.Contains(svc, "gcp"):
		return enhancedGCPParser{}
	case strings.Contains(svc, "azure"):
		return enhancedAzureParser{}
	case strings.Contains(svc, "genesys"):
		return genesysParser{}
	case strings.Contains(svc, "avaya"):
		return enhancedAvayaParser{}
	case strings.Contains(svc, "cloudflare"):
		return enhancedCloudflareParser{}
	default:
		return genericParser{}
	}
}

