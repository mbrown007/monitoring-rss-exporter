package collectors

import (
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"
)

type enhancedAWSParser struct{}

func (enhancedAWSParser) ServiceInfo(item *gofeed.Item) (string, string) {
	serviceName := extractAWSService(item)
	region := extractAWSRegion(item)
	return serviceName, region
}

func (enhancedAWSParser) IncidentKey(item *gofeed.Item) string {
	key := item.GUID
	if key == "" {
		key = item.Link
	}
	if key == "" {
		key = item.Title
	}
	if idx := strings.Index(key, "#"); idx != -1 {
		key = key[idx+1:]
	}
	// Remove AWS-specific suffixes - process in order
	suffixes := []string{"_monitoring_update", "_resolved", "_issue", "_outage", "_investigating", "_monitoring", "_update"}
	for _, suffix := range suffixes {
		key = strings.TrimSuffix(key, suffix)
	}
	return key
}

// ParseAWSGUID extracts the AWS service name and region from a GUID string.
// GUIDs may appear in several formats, including:
//
//	https://status.aws.amazon.com/#service-region_12345
//	arn:aws:health:region::event/AWS_SERVICE_eventid
//
// Unknown formats return empty strings.
func ParseAWSGUID(guid string) (serviceName, region string) {
	if idx := strings.Index(guid, "#"); idx != -1 {
		guid = guid[idx+1:]
	}

	if strings.HasPrefix(guid, "arn:aws:health:") {
		// arn:aws:health:region::event/AWS_SERVICENAME_foo
		parts := strings.Split(guid, ":")
		if len(parts) >= 4 {
			region = parts[3]
		}
		if idx := strings.LastIndex(guid, "/"); idx != -1 {
			svc := guid[idx+1:]
			svc = strings.TrimPrefix(svc, "AWS_")
			svcParts := strings.SplitN(svc, "_", 2)
			serviceName = strings.ToLower(svcParts[0])
		}
		return
	}

	if idx := strings.IndexAny(guid, "_"); idx != -1 {
		guid = guid[:idx]
	}

	parts := strings.Split(guid, "-")
	if len(parts) < 2 {
		return "", ""
	}

	if len(parts) >= 3 {
		region = strings.Join(parts[len(parts)-3:], "-")
		serviceName = strings.Join(parts[:len(parts)-3], "-")
	} else {
		region = parts[len(parts)-1]
		serviceName = strings.Join(parts[:len(parts)-1], "-")
	}
	serviceName = strings.ToLower(serviceName)
	return
}

func extractAWSService(item *gofeed.Item) string {
	// First try GUID parsing for structured service extraction
	if item.GUID != "" {
		if svc, _ := ParseAWSGUID(item.GUID); svc != "" {
			return formatAWSServiceName(svc)
		}
	}
	
	// Content-based service detection
	content := strings.ToLower(item.Title + " " + item.Description + " " + item.Content)
	
	// AWS service patterns - ordered by specificity
	servicePatterns := []struct {
		pattern string
		service string
	}{
		// Compute Services
		{"amazon ec2|elastic compute cloud", "Amazon EC2"},
		{"aws lambda|lambda functions", "AWS Lambda"},
		{"elastic beanstalk", "AWS Elastic Beanstalk"},
		{"amazon ecs|elastic container service", "Amazon ECS"},
		{"amazon eks|elastic kubernetes service", "Amazon EKS"},
		{"aws fargate", "AWS Fargate"},
		{"aws batch", "AWS Batch"},
		{"amazon lightsail", "Amazon Lightsail"},
		{"aws app runner", "AWS App Runner"},
		
		// Storage Services
		{"amazon s3|simple storage service", "Amazon S3"},
		{"amazon ebs|elastic block store", "Amazon EBS"},
		{"amazon efs|elastic file system", "Amazon EFS"},
		{"amazon fsx", "Amazon FSx"},
		{"aws storage gateway", "AWS Storage Gateway"},
		{"aws backup", "AWS Backup"},
		{"amazon glacier", "Amazon S3 Glacier"},
		
		// Database Services
		{"amazon rds|relational database", "Amazon RDS"},
		{"amazon dynamodb", "Amazon DynamoDB"},
		{"amazon redshift", "Amazon Redshift"},
		{"amazon elasticache", "Amazon ElastiCache"},
		{"amazon documentdb", "Amazon DocumentDB"},
		{"amazon neptune", "Amazon Neptune"},
		{"amazon aurora", "Amazon Aurora"},
		{"amazon timestream", "Amazon Timestream"},
		{"amazon keyspaces", "Amazon Keyspaces"},
		
		// Networking Services
		{"amazon vpc|virtual private cloud", "Amazon VPC"},
		{"aws direct connect", "AWS Direct Connect"},
		{"amazon route 53|route53", "Amazon Route 53"},
		{"amazon cloudfront", "Amazon CloudFront"},
		{"elastic load balancing|application load balancer|network load balancer", "Elastic Load Balancing"},
		{"aws transit gateway", "AWS Transit Gateway"},
		{"aws vpn", "AWS VPN"},
		{"aws privatelink", "AWS PrivateLink"},
		{"amazon api gateway", "Amazon API Gateway"},
		
		// Analytics Services
		{"amazon emr|elastic mapreduce", "Amazon EMR"},
		{"amazon athena", "Amazon Athena"},
		{"aws glue", "AWS Glue"},
		{"amazon quicksight", "Amazon QuickSight"},
		{"amazon kinesis", "Amazon Kinesis"},
		{"aws data pipeline", "AWS Data Pipeline"},
		{"amazon elasticsearch|opensearch", "Amazon OpenSearch Service"},
		{"amazon cloudwatch", "Amazon CloudWatch"},
		
		// AI/ML Services
		{"amazon sagemaker", "Amazon SageMaker"},
		{"amazon rekognition", "Amazon Rekognition"},
		{"amazon comprehend", "Amazon Comprehend"},
		{"amazon translate", "Amazon Translate"},
		{"amazon polly", "Amazon Polly"},
		{"amazon transcribe", "Amazon Transcribe"},
		{"amazon lex", "Amazon Lex"},
		{"amazon textract", "Amazon Textract"},
		{"amazon bedrock", "Amazon Bedrock"},
		
		// Security & Identity
		{"aws iam|identity and access management", "AWS IAM"},
		{"aws kms|key management service", "AWS KMS"},
		{"aws secrets manager", "AWS Secrets Manager"},
		{"amazon cognito", "Amazon Cognito"},
		{"aws directory service", "AWS Directory Service"},
		{"aws certificate manager", "AWS Certificate Manager"},
		{"aws shield", "AWS Shield"},
		{"aws waf", "AWS WAF"},
		{"amazon guardduty", "Amazon GuardDuty"},
		{"amazon inspector", "Amazon Inspector"},
		{"aws security hub", "AWS Security Hub"},
		
		// Management Services
		{"aws cloudformation", "AWS CloudFormation"},
		{"aws cloudtrail", "AWS CloudTrail"},
		{"aws config", "AWS Config"},
		{"aws systems manager", "AWS Systems Manager"},
		{"aws organizations", "AWS Organizations"},
		{"aws control tower", "AWS Control Tower"},
		{"aws service catalog", "AWS Service Catalog"},
		{"aws trusted advisor", "AWS Trusted Advisor"},
		{"aws personal health dashboard", "AWS Personal Health Dashboard"},
		
		// Integration Services
		{"amazon sns|simple notification service", "Amazon SNS"},
		{"amazon sqs|simple queue service", "Amazon SQS"},
		{"amazon eventbridge", "Amazon EventBridge"},
		{"aws step functions", "AWS Step Functions"},
		{"amazon mq", "Amazon MQ"},
		{"aws app sync", "AWS AppSync"},
		
		// Developer Tools
		{"aws codecommit", "AWS CodeCommit"},
		{"aws codebuild", "AWS CodeBuild"},
		{"aws codedeploy", "AWS CodeDeploy"},
		{"aws codepipeline", "AWS CodePipeline"},
		{"aws codestar", "AWS CodeStar"},
		{"aws cloud9", "AWS Cloud9"},
		{"aws x-ray", "AWS X-Ray"},
		
		// IoT Services
		{"aws iot core", "AWS IoT Core"},
		{"aws iot device management", "AWS IoT Device Management"},
		{"aws iot analytics", "AWS IoT Analytics"},
		{"aws iot greengrass", "AWS IoT Greengrass"},
		
		// Media Services
		{"amazon elastic transcoder", "Amazon Elastic Transcoder"},
		{"aws elemental", "AWS Elemental"},
		{"amazon ivs", "Amazon IVS"},
		
		// Generic fallbacks based on common AWS terminology
		{"ec2", "Amazon EC2"},
		{"s3", "Amazon S3"},
		{"rds", "Amazon RDS"},
		{"lambda", "AWS Lambda"},
		{"cloudfront", "Amazon CloudFront"},
		{"route 53", "Amazon Route 53"},
		{"elb", "Elastic Load Balancing"},
		{"vpc", "Amazon VPC"},
		{"iam", "AWS IAM"},
		{"cloudwatch", "Amazon CloudWatch"},
	}
	
	for _, sp := range servicePatterns {
		if matched, _ := regexp.MatchString(sp.pattern, content); matched {
			return sp.service
		}
	}
	
	// Extract service from title if it starts with AWS service format
	title := item.Title
	if strings.Contains(strings.ToUpper(title), "AMAZON") || strings.Contains(strings.ToUpper(title), "AWS") {
		// Try to extract service name from titles like "Amazon EC2 (Oregon) Service Status"
		re := regexp.MustCompile(`(?i)(Amazon \w+|AWS \w+)`)
		if matches := re.FindStringSubmatch(title); len(matches) > 1 {
			return formatAWSServiceName(matches[1])
		}
	}
	
	return "AWS Platform"
}

func extractAWSRegion(item *gofeed.Item) string {
	// First try GUID parsing for structured region extraction
	if item.GUID != "" {
		if _, reg := ParseAWSGUID(item.GUID); reg != "" {
			return formatAWSRegionName(reg)
		}
	}
	
	content := strings.ToLower(item.Title + " " + item.Description + " " + item.Content)
	
	// AWS region patterns - ordered by specificity (most specific first)
	regionPatterns := []struct {
		pattern string
		region  string
	}{
		// Global services - check these first
		{"multiple regions|multi-region", "Multi-Region"},
		{"global|worldwide|all regions", "Global"},
		{"edge locations", "Edge Locations"},
		// Standard AWS regions - ordered by specificity
		{"us-east-1|virginia|n\\. virginia", "US East (N. Virginia)"},
		{"us east", "US East (N. Virginia)"},
		{"us-east-2|ohio", "US East (Ohio)"},
		{"us-west-1|california|n\\. california", "US West (N. California)"},
		{"us-west-2|oregon", "US West (Oregon)"},
		{"ca-central-1|canada central", "Canada (Central)"},
		{"ca-west-1|canada west", "Canada (West)"},
		
		// Europe
		{"eu-west-1|ireland", "Europe (Ireland)"},
		{"eu-west-2|london", "Europe (London)"},
		{"eu-west-3|paris", "Europe (Paris)"},
		{"eu-central-1|frankfurt", "Europe (Frankfurt)"},
		{"eu-central-2|zurich", "Europe (Zurich)"},
		{"eu-north-1|stockholm", "Europe (Stockholm)"},
		{"eu-south-1|milan", "Europe (Milan)"},
		{"eu-south-2|spain", "Europe (Spain)"},
		
		// Asia Pacific
		{"ap-southeast-1|singapore", "Asia Pacific (Singapore)"},
		{"ap-southeast-2|sydney", "Asia Pacific (Sydney)"},
		{"ap-southeast-3|jakarta", "Asia Pacific (Jakarta)"},
		{"ap-southeast-4|melbourne", "Asia Pacific (Melbourne)"},
		{"ap-northeast-1|tokyo", "Asia Pacific (Tokyo)"},
		{"ap-northeast-2|seoul", "Asia Pacific (Seoul)"},
		{"ap-northeast-3|osaka", "Asia Pacific (Osaka)"},
		{"ap-south-1|mumbai", "Asia Pacific (Mumbai)"},
		{"ap-south-2|hyderabad", "Asia Pacific (Hyderabad)"},
		{"ap-east-1|hong kong", "Asia Pacific (Hong Kong)"},
		
		// Middle East & Africa
		{"me-south-1|bahrain", "Middle East (Bahrain)"},
		{"me-central-1|uae", "Middle East (UAE)"},
		{"af-south-1|cape town", "Africa (Cape Town)"},
		
		// South America
		{"sa-east-1|sao paulo", "South America (São Paulo)"},
		
		// China
		{"cn-north-1|beijing", "China (Beijing)"},
		{"cn-northwest-1|ningxia", "China (Ningxia)"},
		
		// AWS GovCloud
		{"us-gov-west-1|govcloud west", "AWS GovCloud (US-West)"},
		{"us-gov-east-1|govcloud east", "AWS GovCloud (US-East)"},
	}
	
	for _, rp := range regionPatterns {
		if matched, _ := regexp.MatchString(rp.pattern, content); matched {
			return rp.region
		}
	}
	
	// Extract region from title patterns like "Amazon EC2 (Oregon)"
	re := regexp.MustCompile(`\(([^)]+)\)`)
	if matches := re.FindStringSubmatch(item.Title); len(matches) > 1 {
		regionName := strings.TrimSpace(matches[1])
		return formatAWSRegionName(regionName)
	}
	
	return ""
}

func formatAWSServiceName(service string) string {
	service = strings.TrimSpace(service)
	if service == "" {
		return "AWS Platform"
	}
	
	// Handle AWS service name formatting
	lowerService := strings.ToLower(service)
	
	// Direct mappings for common services
	serviceMap := map[string]string{
		"elastic load balancing": "Elastic Load Balancing",
		"ec2":                     "Amazon EC2",
		"s3":                      "Amazon S3",
		"rds":                     "Amazon RDS",
		"lambda":                  "AWS Lambda",
		"iam":                     "AWS IAM",
		"vpc":                     "Amazon VPC",
		"elb":                     "Elastic Load Balancing",
		"cloudfront":              "Amazon CloudFront",
		"route53":                 "Amazon Route 53",
		"cloudwatch":              "Amazon CloudWatch",
		"dynamodb":                "Amazon DynamoDB",
		"sns":                     "Amazon SNS",
		"sqs":                     "Amazon SQS",
		"athena":                  "Amazon Athena",
		"emr":                     "Amazon EMR",
		"redshift":                "Amazon Redshift",
		"elasticache":             "Amazon ElastiCache",
		"kinesis":                 "Amazon Kinesis",
		"ecs":                     "Amazon ECS",
		"eks":                     "Amazon EKS",
		"fargate":                 "AWS Fargate",
		"glue":                    "AWS Glue",
		"sagemaker":               "Amazon SageMaker",
		"cognito":                 "Amazon Cognito",
		"apigateway":              "Amazon API Gateway",
		"cloudformation":          "AWS CloudFormation",
		"cloudtrail":              "AWS CloudTrail",
		"config":                  "AWS Config",
		"kms":                     "AWS KMS",
		"secretsmanager":          "AWS Secrets Manager",
		"directconnect":           "AWS Direct Connect",
		"transitgateway":          "AWS Transit Gateway",
		"eventbridge":             "Amazon EventBridge",
		"stepfunctions":           "AWS Step Functions",
		"codecommit":              "AWS CodeCommit",
		"codebuild":               "AWS CodeBuild",
		"codedeploy":              "AWS CodeDeploy",
		"codepipeline":            "AWS CodePipeline",
		"xray":                    "AWS X-Ray",
		"systemsmanager":          "AWS Systems Manager",
		"organizations":           "AWS Organizations",
		"controltower":            "AWS Control Tower",
		"waf":                     "AWS WAF",
		"shield":                  "AWS Shield",
		"guardduty":               "Amazon GuardDuty",
		"inspector":               "Amazon Inspector",
		"securityhub":             "AWS Security Hub",
		"quicksight":              "Amazon QuickSight",
		"opensearch":              "Amazon OpenSearch Service",
		"elasticsearch":           "Amazon OpenSearch Service",
		"rekognition":             "Amazon Rekognition",
		"comprehend":              "Amazon Comprehend",
		"translate":               "Amazon Translate",
		"polly":                   "Amazon Polly",
		"transcribe":              "Amazon Transcribe",
		"lex":                     "Amazon Lex",
		"textract":                "Amazon Textract",
		"bedrock":                 "Amazon Bedrock",
	}
	
	if mapped, exists := serviceMap[lowerService]; exists {
		return mapped
	}
	
	// If already properly formatted, return as-is
	if strings.HasPrefix(service, "Amazon ") || strings.HasPrefix(service, "AWS ") {
		return service
	}
	
	// Capitalize first letter and handle common abbreviations
	words := strings.Fields(strings.ToLower(service))
	for i, word := range words {
		switch word {
		case "ec2", "s3", "rds", "vpc", "elb", "sns", "sqs", "ecs", "eks", "emr", "api", "iam", "kms", "waf", "cdn", "iot", "ml", "ai":
			words[i] = strings.ToUpper(word)
		case "aws":
			words[i] = "AWS"
		case "amazon":
			words[i] = "Amazon"
		default:
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			}
		}
	}
	
	formatted := strings.Join(words, " ")
	
	// Add appropriate prefix if not present
	if !strings.HasPrefix(formatted, "Amazon ") && !strings.HasPrefix(formatted, "AWS ") {
		// Determine if it should be Amazon or AWS service
		if strings.Contains(lowerService, "amazon") || isAmazonService(lowerService) {
			formatted = "Amazon " + formatted
		} else {
			formatted = "AWS " + formatted
		}
	}
	
	return formatted
}

func formatAWSRegionName(region string) string {
	region = strings.TrimSpace(region)
	if region == "" {
		return ""
	}
	
	// Direct mappings for AWS regions
	regionMap := map[string]string{
		"us-east-1":      "US East (N. Virginia)",
		"us-east-2":      "US East (Ohio)",
		"us-west-1":      "US West (N. California)",
		"us-west-2":      "US West (Oregon)",
		"ca-central-1":   "Canada (Central)",
		"ca-west-1":      "Canada (West)",
		"eu-west-1":      "Europe (Ireland)",
		"eu-west-2":      "Europe (London)",
		"eu-west-3":      "Europe (Paris)",
		"eu-central-1":   "Europe (Frankfurt)",
		"eu-central-2":   "Europe (Zurich)",
		"eu-north-1":     "Europe (Stockholm)",
		"eu-south-1":     "Europe (Milan)",
		"eu-south-2":     "Europe (Spain)",
		"ap-southeast-1": "Asia Pacific (Singapore)",
		"ap-southeast-2": "Asia Pacific (Sydney)",
		"ap-southeast-3": "Asia Pacific (Jakarta)",
		"ap-southeast-4": "Asia Pacific (Melbourne)",
		"ap-northeast-1": "Asia Pacific (Tokyo)",
		"ap-northeast-2": "Asia Pacific (Seoul)",
		"ap-northeast-3": "Asia Pacific (Osaka)",
		"ap-south-1":     "Asia Pacific (Mumbai)",
		"ap-south-2":     "Asia Pacific (Hyderabad)",
		"ap-east-1":      "Asia Pacific (Hong Kong)",
		"me-south-1":     "Middle East (Bahrain)",
		"me-central-1":   "Middle East (UAE)",
		"af-south-1":     "Africa (Cape Town)",
		"sa-east-1":      "South America (São Paulo)",
		"cn-north-1":     "China (Beijing)",
		"cn-northwest-1": "China (Ningxia)",
		"us-gov-west-1":  "AWS GovCloud (US-West)",
		"us-gov-east-1":  "AWS GovCloud (US-East)",
	}
	
	lowerRegion := strings.ToLower(region)
	if mapped, exists := regionMap[lowerRegion]; exists {
		return mapped
	}
	
	// Handle common region name formats
	regionAliases := map[string]string{
		"multiple regions": "Multi-Region",
		"virginia":       "US East (N. Virginia)",
		"ohio":           "US East (Ohio)",
		"california":     "US West (N. California)",
		"oregon":         "US West (Oregon)",
		"ireland":        "Europe (Ireland)",
		"london":         "Europe (London)",
		"paris":          "Europe (Paris)",
		"frankfurt":      "Europe (Frankfurt)",
		"stockholm":      "Europe (Stockholm)",
		"singapore":      "Asia Pacific (Singapore)",
		"sydney":         "Asia Pacific (Sydney)",
		"tokyo":          "Asia Pacific (Tokyo)",
		"seoul":          "Asia Pacific (Seoul)",
		"mumbai":         "Asia Pacific (Mumbai)",
		"global":         "Global",
		"worldwide":      "Global",
	}
	
	if mapped, exists := regionAliases[lowerRegion]; exists {
		return mapped
	}
	
	// Capitalize and return as-is for unknown regions
	words := strings.Fields(strings.ToLower(region))
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	
	return strings.Join(words, " ")
}

func isAmazonService(service string) bool {
	amazonServices := []string{
		"ec2", "s3", "rds", "vpc", "cloudfront", "route53", "dynamodb", "sns", "sqs",
		"redshift", "elasticache", "kinesis", "ecs", "eks", "emr", "athena", "quicksight",
		"sagemaker", "rekognition", "comprehend", "translate", "polly", "transcribe",
		"lex", "textract", "cognito", "api", "gateway", "opensearch", "elasticsearch",
		"documentdb", "neptune", "aurora", "timestream", "keyspaces", "lightsail",
		"workspaces", "connect", "chime", "workmail", "worklink", "workdocs",
		"guardduty", "inspector", "macie", "detective", "eventbridge", "mq",
		"managed", "streaming", "kafka", "fsx", "efs", "glacier", "storagegateway",
	}
	
	for _, amazonSvc := range amazonServices {
		if strings.Contains(service, amazonSvc) {
			return true
		}
	}
	
	return false
}