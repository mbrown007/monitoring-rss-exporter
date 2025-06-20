package collectors

import (
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"
)

type enhancedAvayaParser struct{}

func (enhancedAvayaParser) ServiceInfo(item *gofeed.Item) (string, string) {
	serviceName := extractAvayaService(item)
	region := extractAvayaRegion(item)
	return serviceName, region
}

func (enhancedAvayaParser) IncidentKey(item *gofeed.Item) string {
	// Prefer GUID (Avaya uses incident URLs as GUIDs)
	if item.GUID != "" {
		return item.GUID
	}
	if item.Link != "" {
		return item.Link
	}
	return strings.TrimSpace(item.Title)
}

func extractAvayaService(item *gofeed.Item) string {
	content := strings.ToLower(item.Title + " " + item.Description + " " + item.Content)
	
	// Avaya service patterns - ordered by specificity (most specific first)
	servicePatterns := []struct {
		pattern string
		service string
	}{
		// Specific service patterns first
		{"preview dialing|predictive dialing", "Preview Dialing"},
		{"axp preview dialing", "Preview Dialing"},
		
		// Core Avaya Platforms  
		{"avaya experience platform", "Avaya Experience Platform"},
		{"avaya enterprise cloud|aec", "Avaya Enterprise Cloud"},
		{"avaya cloud office|aco", "Avaya Cloud Office"},
		{"communications apis|cpaas", "Communications APIs"},
		{"avaya api gateway", "Avaya API Gateway"},
		
		// Contact Center & Communication Services
		{"contact center|call center", "Contact Center"},
		{"outbound dialing|dialer", "Outbound Dialing"},
		{"ivr|interactive voice response", "Interactive Voice Response"},
		{"voice services|voice platform", "Voice Services"},
		{"video conferencing|video calls", "Video Conferencing"},
		{"messaging platform|instant messaging", "Messaging Platform"},
		{"chat services|web chat", "Chat Services"},
		{"sms gateway|text messaging", "SMS Gateway"},
		{"email services|email platform", "Email Services"},
		
		// Telephony & Network Services
		{"pbx|private branch exchange", "PBX Services"},
		{"sip trunk|sip services", "SIP Trunking"},
		{"voip|voice over ip", "VoIP Services"},
		{"telephony platform|phone system", "Telephony Platform"},
		{"call routing|call management", "Call Routing"},
		{"call recording|voice recording", "Call Recording"},
		{"call analytics|voice analytics", "Call Analytics"},
		{"network connectivity|network services", "Network Services"},
		{"bandwidth management", "Bandwidth Management"},
		{"qos|quality of service", "Quality of Service"},
		
		// Collaboration & Productivity
		{"collaboration platform|team collaboration", "Collaboration Platform"},
		{"unified communications|uc", "Unified Communications"},
		{"presence services|user presence", "Presence Services"},
		{"calendar integration|scheduling", "Calendar Integration"},
		{"file sharing|document sharing", "File Sharing"},
		{"screen sharing|desktop sharing", "Screen Sharing"},
		{"whiteboard|interactive whiteboard", "Whiteboard Services"},
		
		// Analytics & Reporting
		{"analytics platform|reporting platform", "Analytics Platform"},
		{"real-time analytics|live analytics", "Real-time Analytics"},
		{"historical reporting|call reports", "Historical Reporting"},
		{"dashboard services|executive dashboard", "Dashboard Services"},
		{"workforce analytics|agent analytics", "Workforce Analytics"},
		{"customer analytics|interaction analytics", "Customer Analytics"},
		{"speech analytics|voice analytics", "Speech Analytics"},
		
		// Integration & APIs
		{"crm integration|customer relationship", "CRM Integration"},
		{"salesforce integration|sfdc", "Salesforce Integration"},
		{"microsoft teams integration|teams", "Microsoft Teams Integration"},
		{"api services|rest api", "API Services"},
		{"webhook services|event notifications", "Webhook Services"},
		{"single sign-on|sso", "Single Sign-On"},
		{"directory services|ldap", "Directory Services"},
		
		// Infrastructure & Operations
		{"database services|data storage", "Database Services"},
		{"backup services|data backup", "Backup Services"},
		{"security services|authentication", "Security Services"},
		{"monitoring services|system monitoring", "Monitoring Services"},
		{"load balancing|traffic management", "Load Balancing"},
		{"cdn|content delivery", "Content Delivery Network"},
		{"dns services|domain name", "DNS Services"},
		
		// Mobile & Edge Services
		{"mobile app|mobile platform", "Mobile Platform"},
		{"mobile push|push notifications", "Push Notifications"},
		{"edge computing|edge services", "Edge Computing"},
		{"iot platform|internet of things", "IoT Platform"},
		
		// Administrative & Management
		{"admin portal|administration", "Admin Portal"},
		{"user management|account management", "User Management"},
		{"provisioning services|auto-provisioning", "Provisioning Services"},
		{"billing services|usage tracking", "Billing Services"},
		{"license management|seat management", "License Management"},
		
		// Environment-specific patterns
		{"prod-na|production north america", "Production North America"},
		{"prod-eu|production europe", "Production Europe"},
		{"prod-ase|production asia", "Production Asia"},
		{"prod-anz|production australia", "Production Australia"},
		{"staging environment|test environment", "Staging Environment"},
		{"development environment|dev environment", "Development Environment"},
		
		// Generic service patterns
		{"web portal|customer portal", "Web Portal"},
		{"api endpoint|api gateway", "API Gateway"},
		{"authentication service|auth service", "Authentication Service"},
		{"notification service|alert service", "Notification Service"},
		{"storage service|data service", "Storage Service"},
		{"compute service|processing service", "Compute Service"},
	}
	
	for _, sp := range servicePatterns {
		if matched, _ := regexp.MatchString(sp.pattern, content); matched {
			return sp.service
		}
	}
	
	// Extract from title format: "Service Name - Description"
	title := item.Title
	if idx := strings.Index(title, " - "); idx != -1 {
		servicePart := strings.TrimSpace(title[:idx])
		return formatAvayaServiceName(servicePart)
	}
	
	// Check for Avaya-specific abbreviations (after pattern matching to avoid conflicts)
	if strings.Contains(content, "aec") && !strings.Contains(content, "preview") {
		return "Avaya Enterprise Cloud"
	}
	if strings.Contains(content, "aco") && !strings.Contains(content, "preview") {
		return "Avaya Cloud Office"
	}
	
	return "Avaya Cloud Platform"
}

func extractAvayaRegion(item *gofeed.Item) string {
	content := strings.ToLower(item.Title + " " + item.Description + " " + item.Content)
	
	// Avaya region patterns - ordered by specificity
	regionPatterns := []struct {
		pattern string
		region  string
	}{
		// Global/Multi-region patterns first
		{"multiple regions|multi-region", "Multi-Region"},
		{"all regions|globally|worldwide", "Global"},
		
		// Primary Avaya Regions - check specific patterns first
		{"prod-na|production north america|north america|na region", "North America"},
		{"prod-sa|production south america|south america|sa region", "South America"}, 
		{"prod-eu|production europe|europe region|eu region", "Europe"},
		{"prod-uk|production uk|united kingdom|uk region", "United Kingdom"},
		{"prod-ase|production asia|asia pacific|apac|ase region", "Asia Pacific"},
		{"prod-anz|production australia|australia|anz region", "Australia & New Zealand"},
		{"prod-ca|production canada|canada region|ca region", "Canada"},
		{"prod-jp|production japan|japan region|jp region", "Japan"},
		{"prod-in|production india|india region|in region", "India"},
		
		// Environment indicators
		{"staging|test environment", "Staging"},
		{"development|dev environment", "Development"},
		
		// Geographic patterns
		{"americas region|americas", "Americas"},
		{"emea region|europe middle east africa", "EMEA"},
		{"apac region|asia pacific region", "Asia Pacific"},
		
		// Country-specific patterns
		{"united states|usa|us region", "United States"},
		{"canada|canadian region", "Canada"},
		{"brazil|brazilian region", "Brazil"},
		{"mexico|mexican region", "Mexico"},
		{"argentina|argentinian region", "Argentina"},
		{"colombia|colombian region", "Colombia"},
		
		{"france|french region", "France"},
		{"germany|german region", "Germany"},
		{"spain|spanish region", "Spain"},
		{"italy|italian region", "Italy"},
		{"netherlands|dutch region", "Netherlands"},
		{"poland|polish region", "Poland"},
		{"sweden|swedish region", "Sweden"},
		{"norway|norwegian region", "Norway"},
		{"denmark|danish region", "Denmark"},
		{"finland|finnish region", "Finland"},
		
		{"china|chinese region", "China"},
		{"singapore|singaporean region", "Singapore"},
		{"hong kong|hk region", "Hong Kong"},
		{"taiwan|taiwanese region", "Taiwan"},
		{"south korea|korean region", "South Korea"},
		{"philippines|philippine region", "Philippines"},
		{"indonesia|indonesian region", "Indonesia"},
		{"malaysia|malaysian region", "Malaysia"},
		{"thailand|thai region", "Thailand"},
		{"vietnam|vietnamese region", "Vietnam"},
		
		{"australia|australian region", "Australia"},
		{"new zealand|nz region", "New Zealand"},
		
		{"south africa|south african region", "South Africa"},
		{"nigeria|nigerian region", "Nigeria"},
		{"kenya|kenyan region", "Kenya"},
		{"egypt|egyptian region", "Egypt"},
		
		{"israel|israeli region", "Israel"},
		{"uae|united arab emirates", "United Arab Emirates"},
		{"saudi arabia|saudi region", "Saudi Arabia"},
		{"qatar|qatari region", "Qatar"},
		{"kuwait|kuwaiti region", "Kuwait"},
	}
	
	for _, rp := range regionPatterns {
		if matched, _ := regexp.MatchString(rp.pattern, content); matched {
			return rp.region
		}
	}
	
	return ""
}

func formatAvayaServiceName(service string) string {
	service = strings.TrimSpace(service)
	if service == "" {
		return "Avaya Cloud Platform"
	}
	
	// Handle common Avaya service name formats
	lowerService := strings.ToLower(service)
	
	// Direct mappings for Avaya services
	serviceMap := map[string]string{
		"axp":                           "Avaya Experience Platform",
		"aec":                           "Avaya Enterprise Cloud", 
		"aco":                           "Avaya Cloud Office",
		"cpaas":                         "Communications APIs",
		"avaya experience platform":    "Avaya Experience Platform",
		"avaya enterprise cloud":       "Avaya Enterprise Cloud",
		"avaya cloud office":           "Avaya Cloud Office", 
		"communications apis":           "Communications APIs",
		"avaya api gateway":             "Avaya API Gateway",
		"contact center":                "Contact Center",
		"preview dialing":               "Preview Dialing",
		"voice services":                "Voice Services",
		"unified communications":        "Unified Communications",
		"collaboration platform":       "Collaboration Platform",
		"analytics platform":           "Analytics Platform",
		"telephony platform":           "Telephony Platform",
	}
	
	if mapped, exists := serviceMap[lowerService]; exists {
		return mapped
	}
	
	// If already properly formatted, return as-is
	if strings.HasPrefix(service, "Avaya ") {
		return service
	}
	
	// Capitalize and format service name
	words := strings.Fields(strings.ToLower(service))
	for i, word := range words {
		switch word {
		case "axp":
			words[i] = "AXP"
		case "aec":
			words[i] = "AEC"
		case "aco":
			words[i] = "ACO"
		case "api", "apis":
			words[i] = "API"
		case "ivr":
			words[i] = "IVR"
		case "pbx":
			words[i] = "PBX"
		case "sip":
			words[i] = "SIP"
		case "voip":
			words[i] = "VoIP"
		case "sms":
			words[i] = "SMS"
		case "crm":
			words[i] = "CRM"
		case "sso":
			words[i] = "SSO"
		case "ldap":
			words[i] = "LDAP"
		case "cdn":
			words[i] = "CDN"
		case "dns":
			words[i] = "DNS"
		case "iot":
			words[i] = "IoT"
		case "uc":
			words[i] = "UC"
		case "qos":
			words[i] = "QoS"
		default:
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			}
		}
	}
	
	formatted := strings.Join(words, " ")
	
	// Add Avaya prefix for non-Avaya branded services
	if !strings.HasPrefix(formatted, "Avaya ") && !containsAvayaBranding(formatted) {
		// Don't add Avaya prefix for platform-generic terms
		if !isPlatformGeneric(formatted) {
			formatted = "Avaya " + formatted
		}
	}
	
	return formatted
}

func containsAvayaBranding(service string) bool {
	// Check if service already contains Avaya branding or is generic enough
	lowerService := strings.ToLower(service)
	brandedTerms := []string{
		"avaya", "axp", "aec", "aco", "communications apis",
		"contact center", "unified communications", "telephony platform",
	}
	
	for _, term := range brandedTerms {
		if strings.Contains(lowerService, term) {
			return true
		}
	}
	
	return false
}

func isPlatformGeneric(service string) bool {
	// Platform-generic terms that shouldn't get Avaya prefix
	genericTerms := []string{
		"telephony platform", "collaboration platform", "analytics platform",
		"unified communications", "contact center",
	}
	
	lowerService := strings.ToLower(service)
	for _, term := range genericTerms {
		if strings.Contains(lowerService, term) {
			return true
		}
	}
	
	return false
}