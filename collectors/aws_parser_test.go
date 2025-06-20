package collectors

import (
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func TestEnhancedAWSParser_ServiceInfo(t *testing.T) {
	parser := enhancedAWSParser{}
	
	tests := []struct {
		name            string
		title           string
		description     string
		guid            string
		expectedService string
		expectedRegion  string
	}{
		{
			name:            "EC2 outage with GUID",
			title:           "OUTAGE: Unable to Launch Instances",
			description:     "We are investigating connectivity issues that are preventing instance launches.",
			guid:            "https://status.aws.amazon.com/#ec2-us-west-2_outage",
			expectedService: "Amazon EC2",
			expectedRegion:  "US West (Oregon)",
		},
		{
			name:            "Athena service impact",
			title:           "Service impact: Increased Queue Processing Time",
			description:     "We are investigating increased processing times for queued queries in the US-WEST-2 Region.",
			guid:            "https://status.aws.amazon.com/#athena-us-west-2_1749832722",
			expectedService: "Amazon Athena",
			expectedRegion:  "US West (Oregon)",
		},
		{
			name:            "Lambda function delays",
			title:           "AWS Lambda function execution delays - US East",
			description:     "Lambda functions are experiencing execution delays in the US East region.",
			expectedService: "AWS Lambda",
			expectedRegion:  "US East (N. Virginia)",
		},
		{
			name:            "S3 global issue from title",
			title:           "Amazon S3 Service Degradation - Global",
			description:     "S3 operations are experiencing elevated error rates globally.",
			expectedService: "Amazon S3",
			expectedRegion:  "Global",
		},
		{
			name:            "RDS connectivity issue",
			title:           "Amazon RDS connectivity issue - Europe (Frankfurt)",
			description:     "RDS instances in Frankfurt are experiencing connectivity issues.",
			expectedService: "Amazon RDS",
			expectedRegion:  "Europe (Frankfurt)",
		},
		{
			name:            "DynamoDB from content",
			title:           "Database service degradation",
			description:     "Amazon DynamoDB operations are experiencing increased latency in ap-southeast-1.",
			expectedService: "Amazon DynamoDB",
			expectedRegion:  "Asia Pacific (Singapore)",
		},
		{
			name:            "CloudFront edge locations",
			title:           "CloudFront cache miss rate increase",
			description:     "Amazon CloudFront edge locations are experiencing increased cache miss rates.",
			expectedService: "Amazon CloudFront",
			expectedRegion:  "Edge Locations",
		},
		{
			name:            "ARN format GUID",
			title:           "EC2 instance launch failures",
			description:     "EC2 instances failing to launch in us-east-1.",
			guid:            "arn:aws:health:us-east-1::event/AWS_EC2_EXAMPLE",
			expectedService: "Amazon EC2",
			expectedRegion:  "US East (N. Virginia)",
		},
		{
			name:            "VPC networking issue",
			title:           "Amazon VPC (Oregon) Service Status",
			description:     "VPC networking experiencing intermittent connectivity issues.",
			expectedService: "Amazon VPC",
			expectedRegion:  "US West (Oregon)",
		},
		{
			name:            "Generic AWS platform issue",
			title:           "Platform connectivity issue",
			description:     "General connectivity issues affecting AWS services.",
			expectedService: "AWS Platform",
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

func TestEnhancedAWSParser_IncidentKey(t *testing.T) {
	parser := enhancedAWSParser{}
	
	tests := []struct {
		name        string
		guid        string
		link        string
		title       string
		expectedKey string
	}{
		{
			name:        "GUID with hash and suffix",
			guid:        "https://status.aws.amazon.com/#ec2-us-west-2_outage",
			expectedKey: "ec2-us-west-2",
		},
		{
			name:        "GUID with resolved suffix",
			guid:        "athena-us-west-2_resolved",
			expectedKey: "athena-us-west-2",
		},
		{
			name:        "GUID with investigating suffix",
			guid:        "#lambda-us-east-1_investigating",
			expectedKey: "lambda-us-east-1",
		},
		{
			name:        "No GUID, use link",
			link:        "https://status.aws.amazon.com/",
			title:       "Service Issue",
			expectedKey: "https://status.aws.amazon.com/",
		},
		{
			name:        "No GUID or link, use title",
			title:       "AWS Lambda connectivity issue",
			expectedKey: "AWS Lambda connectivity issue",
		},
		{
			name:        "Multiple suffixes",
			guid:        "s3-global_monitoring_update",
			expectedKey: "s3-global",
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

func TestExtractAWSService(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		content  string
		guid     string
		expected string
	}{
		{
			name:     "EC2 from title",
			title:    "Amazon EC2 connectivity issues",
			expected: "Amazon EC2",
		},
		{
			name:     "Lambda from content",
			content:  "AWS Lambda functions are experiencing delays",
			expected: "AWS Lambda",
		},
		{
			name:     "S3 abbreviation",
			content:  "S3 bucket operations affected",
			expected: "Amazon S3",
		},
		{
			name:     "RDS from description",
			content:  "Relational Database Service connectivity issues",
			expected: "Amazon RDS",
		},
		{
			name:     "CloudWatch monitoring",
			content:  "CloudWatch metrics collection delays",
			expected: "Amazon CloudWatch",
		},
		{
			name:     "DynamoDB",
			content:  "Amazon DynamoDB table operations experiencing latency",
			expected: "Amazon DynamoDB",
		},
		{
			name:     "Route 53 DNS",
			content:  "Route 53 DNS resolution issues",
			expected: "Amazon Route 53",
		},
		{
			name:     "Elastic Load Balancing",
			content:  "Application Load Balancer health checks failing",
			expected: "Elastic Load Balancing",
		},
		{
			name:     "SageMaker ML",
			content:  "Amazon SageMaker training jobs experiencing delays",
			expected: "Amazon SageMaker",
		},
		{
			name:     "From GUID",
			guid:     "athena-us-west-2_issue",
			expected: "Amazon Athena",
		},
		{
			name:     "Title format extraction",
			title:    "Amazon Redshift (Oregon) Service Status",
			expected: "Amazon Redshift",
		},
		{
			name:     "No recognizable service",
			content:  "Platform wide connectivity issues",
			expected: "AWS Platform",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &gofeed.Item{
				Title:       tt.title,
				Description: tt.content,
				GUID:        tt.guid,
			}
			
			service := extractAWSService(item)
			assert.Equal(t, tt.expected, service)
		})
	}
}

func TestExtractAWSRegion(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		content  string
		guid     string
		expected string
	}{
		{
			name:     "US West Oregon",
			content:  "Issues affecting services in Oregon region",
			expected: "US West (Oregon)",
		},
		{
			name:     "US East Virginia", 
			content:  "Services in Virginia are experiencing issues",
			expected: "US East (N. Virginia)",
		},
		{
			name:     "Europe Frankfurt",
			content:  "eu-central-1 region experiencing connectivity issues",
			expected: "Europe (Frankfurt)",
		},
		{
			name:     "Asia Pacific Singapore",
			content:  "ap-southeast-1 services affected",
			expected: "Asia Pacific (Singapore)",
		},
		{
			name:     "Global issue",
			content:  "Global connectivity issues affecting all regions",
			expected: "Global",
		},
		{
			name:     "Multiple regions",
			content:  "Multiple regions affected including us-east-1 and eu-west-1",
			expected: "Multi-Region",
		},
		{
			name:     "From GUID",
			guid:     "ec2-us-west-2_outage",
			expected: "US West (Oregon)",
		},
		{
			name:     "Title format extraction",
			title:    "Amazon EC2 (Oregon) Service Status",
			expected: "US West (Oregon)",
		},
		{
			name:     "Canada Central",
			content:  "ca-central-1 region experiencing issues",
			expected: "Canada (Central)",
		},
		{
			name:     "GovCloud",
			content:  "us-gov-west-1 region affected",
			expected: "AWS GovCloud (US-West)",
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
				Title:       tt.title,
				Description: tt.content,
				GUID:        tt.guid,
			}
			
			region := extractAWSRegion(item)
			assert.Equal(t, tt.expected, region)
		})
	}
}

func TestFormatAWSServiceName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ec2", "Amazon EC2"},
		{"lambda", "AWS Lambda"},
		{"s3", "Amazon S3"},
		{"rds", "Amazon RDS"},
		{"iam", "AWS IAM"},
		{"cloudfront", "Amazon CloudFront"},
		{"route53", "Amazon Route 53"},
		{"dynamodb", "Amazon DynamoDB"},
		{"sagemaker", "Amazon SageMaker"},
		{"elastic load balancing", "Elastic Load Balancing"},
		{"Amazon EC2", "Amazon EC2"}, // Already formatted
		{"AWS Lambda", "AWS Lambda"}, // Already formatted
		{"", "AWS Platform"},
		{"unknown service", "AWS Unknown Service"},
		{"api gateway", "Amazon API Gateway"},
		{"kms", "AWS KMS"},
		{"vpc", "Amazon VPC"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatAWSServiceName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatAWSRegionName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"us-east-1", "US East (N. Virginia)"},
		{"us-west-2", "US West (Oregon)"},
		{"eu-central-1", "Europe (Frankfurt)"},
		{"ap-southeast-1", "Asia Pacific (Singapore)"},
		{"ca-central-1", "Canada (Central)"},
		{"virginia", "US East (N. Virginia)"},
		{"oregon", "US West (Oregon)"},
		{"frankfurt", "Europe (Frankfurt)"},
		{"singapore", "Asia Pacific (Singapore)"},
		{"global", "Global"},
		{"", ""},
		{"unknown region", "Unknown Region"},
		{"us-gov-west-1", "AWS GovCloud (US-West)"},
		{"multiple regions", "Multi-Region"},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := formatAWSRegionName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsAmazonService(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"ec2", true},
		{"s3", true},
		{"rds", true},
		{"lambda", false},
		{"iam", false},
		{"vpc", true},
		{"cloudfront", true},
		{"route53", true},
		{"dynamodb", true},
		{"sagemaker", true},
		{"unknown", false},
		{"kms", false},
		{"waf", false},
		{"cognito", true},
		{"api", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isAmazonService(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}