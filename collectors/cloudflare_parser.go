package collectors

import (
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"
)

type enhancedCloudflareParser struct{}

func (enhancedCloudflareParser) ServiceInfo(item *gofeed.Item) (string, string) {
	serviceName := extractCloudflareService(item)
	region := extractCloudflareRegion(item)
	return serviceName, region
}

func (enhancedCloudflareParser) IncidentKey(item *gofeed.Item) string {
	// Prefer GUID (Cloudflare uses tag format)
	if item.GUID != "" {
		return item.GUID
	}
	if item.Link != "" {
		return item.Link
	}
	return strings.TrimSpace(item.Title)
}

func extractCloudflareService(item *gofeed.Item) string {
	content := strings.ToLower(item.Title + " " + item.Description + " " + item.Content)
	
	// Cloudflare service patterns - ordered by specificity
	servicePatterns := []struct {
		pattern string
		service string
	}{
		// Core Cloudflare Services
		{"dns service|dns resolution|dns outage", "DNS"},
		{"cdn performance|cdn issue|cdn outage|content delivery", "CDN"},
		{"waf blocking|waf issue|web application firewall", "WAF"},
		{"ddos protection|ddos mitigation|attack mitigation", "DDoS Protection"},
		{"api rate limiting|api gateway|api service", "API Gateway"},
		{"ssl certificate|ssl issue|certificate", "SSL/TLS"},
		{"load balancing|load balancer", "Load Balancing"},
		{"workers|edge computing|serverless", "Cloudflare Workers"},
		{"stream|video streaming", "Cloudflare Stream"},
		{"images|image optimization", "Cloudflare Images"},
		{"pages|static sites", "Cloudflare Pages"},
		{"access|zero trust|identity", "Cloudflare Access"},
		{"gateway|secure web gateway", "Cloudflare Gateway"},
		{"tunnel|argo tunnel", "Cloudflare Tunnel"},
		{"spectrum|tcp proxy", "Cloudflare Spectrum"},
		{"analytics|insights|reporting", "Analytics"},
		{"bot management|bot protection", "Bot Management"},
		{"rate limiting|rate protection", "Rate Limiting"},
		
		// Infrastructure & Network
		{"edge server|edge node", "Edge Servers"},
		{"network connectivity|network issue", "Network"},
		{"routing|traffic routing", "Traffic Routing"},
		{"caching|cache performance", "Caching"},
		{"bandwidth|data transfer", "Bandwidth"},
		
		// Datacenter Operations
		{"datacenter maintenance|data center|scheduled maintenance", "Datacenter Maintenance"},
		{"power maintenance|electrical maintenance", "Power Systems"},
		{"network maintenance|infrastructure maintenance", "Network Maintenance"},
		{"hardware maintenance|server maintenance", "Hardware Maintenance"},
		{"cooling maintenance|hvac maintenance", "Cooling Systems"},
		
		// Regional Services
		{"performance degradation|performance issue", "Performance"},
		{"connectivity issue|connection problem", "Connectivity"},
		{"service degradation|service issue", "Service Degradation"},
		{"latency issue|high latency", "Latency"},
		{"packet loss|network loss", "Packet Loss"},
	}
	
	for _, sp := range servicePatterns {
		if matched, _ := regexp.MatchString(sp.pattern, content); matched {
			return sp.service
		}
	}
	
	// Check for datacenter code pattern (3 letters followed by parentheses)
	if matched, _ := regexp.MatchString(`[A-Z]{3}\s*\([^)]+\)`, item.Title); matched {
		return "Datacenter Maintenance"
	}
	
	return "Cloudflare Services"
}

func extractCloudflareRegion(item *gofeed.Item) string {
	content := strings.ToLower(item.Title + " " + item.Description + " " + item.Content)
	
	// Extract datacenter code and location from title format: "XNH (Nasiriyah) on 2025-07-03"
	datacenterRegex := regexp.MustCompile(`([A-Z]{3})\s*\(([^)]+)\)`)
	if matches := datacenterRegex.FindStringSubmatch(item.Title); len(matches) >= 3 {
		code := matches[1]
		location := matches[2]
		return code + " (" + location + ")"
	}
	
	// Cloudflare region patterns - ordered by specificity
	regionPatterns := []struct {
		pattern string
		region  string
	}{
		// Global/Multi-region patterns first
		{"global|worldwide|all regions", "Global"},
		{"multiple regions|multi-region", "Multi-Region"},
		
		// Continental regions
		{"north america|na region", "North America"},
		{"south america|sa region", "South America"},
		{"europe|european|eu region", "Europe"},
		{"asia pacific|apac|asia-pacific", "Asia Pacific"},
		{"middle east|me region", "Middle East"},
		{"africa|african region", "Africa"},
		{"oceania|oceanic region", "Oceania"},
		
		// Specific regions
		{"european datacenters|europe datacenters", "European Datacenters"},
		{"asian datacenters|asia datacenters", "Asian Datacenters"},
		{"american datacenters|americas datacenters", "American Datacenters"},
		{"african datacenters", "African Datacenters"},
		
		// Countries and major regions
		{"united states|usa|us region", "United States"},
		{"canada|canadian region", "Canada"},
		{"brazil|brazilian region", "Brazil"},
		{"mexico|mexican region", "Mexico"},
		
		{"united kingdom|uk region|britain", "United Kingdom"},
		{"france|french region", "France"},
		{"germany|german region", "Germany"},
		{"netherlands|dutch region", "Netherlands"},
		{"spain|spanish region", "Spain"},
		{"italy|italian region", "Italy"},
		{"poland|polish region", "Poland"},
		{"russia|russian region", "Russia"},
		
		{"china|chinese region", "China"},
		{"japan|japanese region", "Japan"},
		{"south korea|korean region", "Korea"},
		{"india|indian region", "India"},
		{"singapore|singaporean region", "Singapore"},
		{"australia|australian region", "Australia"},
		{"new zealand|nz region", "New Zealand"},
		
		{"south africa|south african region", "South Africa"},
		{"egypt|egyptian region", "Egypt"},
		{"nigeria|nigerian region", "Nigeria"},
		{"kenya|kenyan region", "Kenya"},
		
		// Major cities (common Cloudflare datacenter locations)
		{"london|lon", "London"},
		{"paris|cdg", "Paris"},
		{"frankfurt|fra", "Frankfurt"},
		{"amsterdam|ams", "Amsterdam"},
		{"madrid|mad", "Madrid"},
		{"milan|mxp", "Milan"},
		{"stockholm|arn", "Stockholm"},
		{"warsaw|waw", "Warsaw"},
		
		{"new york|nyc|ewr", "New York"},
		{"los angeles|lax", "Los Angeles"},
		{"chicago|ord", "Chicago"},
		{"dallas|dfw", "Dallas"},
		{"atlanta|atl", "Atlanta"},
		{"miami|mia", "Miami"},
		{"seattle|sea", "Seattle"},
		{"san francisco|sfo", "San Francisco"},
		{"toronto|yyz", "Toronto"},
		{"vancouver|yvr", "Vancouver"},
		
		{"tokyo|nrt|hnd", "Tokyo"},
		{"osaka|kix", "Osaka"},
		{"seoul|icn", "Seoul"},
		{"hong kong|hkg", "Hong Kong"},
		{"singapore|sin", "Singapore"},
		{"sydney|syd", "Sydney"},
		{"melbourne|mel", "Melbourne"},
		{"mumbai|bom", "Mumbai"},
		{"bangalore|blr", "Bangalore"},
		{"delhi|del", "Delhi"},
		
		{"cairo|cai", "Cairo"},
		{"johannesburg|jnb", "Johannesburg"},
		{"casablanca|cmn", "Casablanca"},
		{"lagos|los", "Lagos"},
		{"nairobi|nbo", "Nairobi"},
		
		{"dublin|dub", "Dublin"},
		{"zurich|zur", "Zurich"},
		{"vienna|vie", "Vienna"},
		{"prague|prg", "Prague"},
		{"budapest|bud", "Budapest"},
		{"moscow|svo", "Moscow"},
		{"istanbul|ist", "Istanbul"},
		{"tel aviv|tlv", "Tel Aviv"},
		{"riyadh|ruh", "Riyadh"},
		{"dubai|dxb", "Dubai"},
		
		// Special patterns
		{"staging|test environment", "Staging"},
		{"development|dev environment", "Development"},
	}
	
	for _, rp := range regionPatterns {
		if matched, _ := regexp.MatchString(rp.pattern, content); matched {
			return rp.region
		}
	}
	
	return ""
}