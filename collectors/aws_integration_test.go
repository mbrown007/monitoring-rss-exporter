package collectors

import (
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/suite"

	"github.com/mbrown007/monitoring-rss-exporter/connectors"
	maas "github.com/mbrown007/monitoring-rss-exporter/monitoring-maas"
)

type AWSIntegrationTestSuite struct {
	suite.Suite
	Connector *connectors.MockHTTPConnector
	Exporter  *maas.Exporter
}

func (s *AWSIntegrationTestSuite) SetupTest() {
	s.Connector = &connectors.MockHTTPConnector{Responses: make(map[string]string)}
	s.Exporter = nil
}

func (s *AWSIntegrationTestSuite) setupExporter(feedPath, url, name, provider string) {
	data, err := os.ReadFile(feedPath)
	s.Require().NoError(err)
	s.Connector.Responses[url] = string(data)

	app := kingpin.New("test", "")
	cfg := maas.ServiceFeed{Name: name, URL: url, Provider: provider, Interval: 300}

	e, err := maas.NewExporter(app, s.Connector,
		maas.WithScheduledScrapers(NewFeedCollector(app, cfg)),
		maas.WithLabels(&maas.MockLabels{}),
		maas.WithArgs([]string{"--web.listen-port=0"}),
	)
	s.Require().NoError(err)
	s.Exporter = e
}

func (s *AWSIntegrationTestSuite) TestAWSEC2Outage() {
	s.setupExporter("testdata/aws_outage.rss", "http://mock.aws/ec2", "aws-test", "aws")
	s.Exporter.Start()

	// Verify service status metrics
	expected := "# HELP aws_test_service_status Current service status\n" +
		"# TYPE aws_test_service_status gauge\n" +
		"aws_test_service_status{customer=\"\",service=\"aws-test\",state=\"ok\"} 0\n" +
		"aws_test_service_status{customer=\"\",service=\"aws-test\",state=\"outage\"} 1\n" +
		"aws_test_service_status{customer=\"\",service=\"aws-test\",state=\"service_issue\"} 0\n"

	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "aws-test_service_status")
	s.NoError(err)

	// AWS enhanced parser is working - status correctly shows outage
}

func (s *AWSIntegrationTestSuite) TestAWSAthenaServiceIssue() {
	s.setupExporter("testdata/aws_athena_us_west_2_issue.rss", "http://mock.aws/athena", "aws-athena", "aws")
	s.Exporter.Start()

	// Verify service status shows service issue
	expected := "# HELP aws_athena_service_status Current service status\n" +
		"# TYPE aws_athena_service_status gauge\n" +
		"aws_athena_service_status{customer=\"\",service=\"aws-athena\",state=\"ok\"} 0\n" +
		"aws_athena_service_status{customer=\"\",service=\"aws-athena\",state=\"outage\"} 0\n" +
		"aws_athena_service_status{customer=\"\",service=\"aws-athena\",state=\"service_issue\"} 1\n"

	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "aws-athena_service_status")
	s.NoError(err)
}

func (s *AWSIntegrationTestSuite) TestAWSMultipleItems() {
	s.setupExporter("testdata/aws_multi_item.rss", "http://mock.aws/multi", "aws-multi", "aws")
	s.Exporter.Start()

	// The multi-item feed has both resolved and active incidents
	// The latest incident state should be reflected in the metrics
	expected := "# HELP aws_multi_service_status Current service status\n" +
		"# TYPE aws_multi_service_status gauge\n" +
		"aws_multi_service_status{customer=\"\",service=\"aws-multi\",state=\"ok\"} 1\n" +
		"aws_multi_service_status{customer=\"\",service=\"aws-multi\",state=\"outage\"} 0\n" +
		"aws_multi_service_status{customer=\"\",service=\"aws-multi\",state=\"service_issue\"} 0\n"

	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "aws-multi_service_status")
	s.NoError(err)
}

func (s *AWSIntegrationTestSuite) TestAWSLambdaGlobal() {
	// Create test data for AWS Lambda global issue
	lambdaTestData := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title><![CDATA[AWS Lambda Service Status]]></title>
    <link>https://status.aws.amazon.com/</link>
    <language>en-us</language>
    <lastBuildDate>Fri, 13 Jun 2025 14:22:00 PDT</lastBuildDate>
    <generator>AWS Service Health Dashboard RSS Generator</generator>
    <description><![CDATA[Receive the most recent update for events affecting AWS Lambda globally.]]></description>
    <ttl>5</ttl>

    <item>
      <title><![CDATA[Service degradation: Lambda execution delays - Global]]></title>
      <link>https://status.aws.amazon.com/</link>
      <pubDate>Fri, 13 Jun 2025 14:22:00 PDT</pubDate>
      <guid isPermaLink="false">https://status.aws.amazon.com/#lambda-global_investigating</guid>
      <description><![CDATA[We are investigating elevated error rates and execution delays affecting AWS Lambda functions globally across multiple regions.]]></description>
    </item>

  </channel>
</rss>`

	s.Connector.Responses["http://mock.aws/lambda"] = lambdaTestData
	
	app := kingpin.New("test", "")
	cfg := maas.ServiceFeed{Name: "aws-lambda", URL: "http://mock.aws/lambda", Provider: "aws", Interval: 300}

	e, err := maas.NewExporter(app, s.Connector,
		maas.WithScheduledScrapers(NewFeedCollector(app, cfg)),
		maas.WithLabels(&maas.MockLabels{}),
		maas.WithArgs([]string{"--web.listen-port=0"}),
	)
	s.Require().NoError(err)
	s.Exporter = e
	s.Exporter.Start()

	// Verify service status shows incident
	expected := "# HELP aws_lambda_service_status Current service status\n" +
		"# TYPE aws_lambda_service_status gauge\n" +
		"aws_lambda_service_status{customer=\"\",service=\"aws-lambda\",state=\"ok\"} 0\n" +
		"aws_lambda_service_status{customer=\"\",service=\"aws-lambda\",state=\"outage\"} 0\n" +
		"aws_lambda_service_status{customer=\"\",service=\"aws-lambda\",state=\"service_issue\"} 1\n"

	err = testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "aws-lambda_service_status")
	s.NoError(err)
}

func (s *AWSIntegrationTestSuite) TestAWSS3Resolved() {
	// Create test data for resolved AWS S3 issue
	s3ResolvedData := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title><![CDATA[Amazon S3 Service Status]]></title>
    <link>https://status.aws.amazon.com/</link>
    <language>en-us</language>
    <lastBuildDate>Fri, 13 Jun 2025 15:45:00 PDT</lastBuildDate>
    <generator>AWS Service Health Dashboard RSS Generator</generator>
    <description><![CDATA[Receive the most recent update for events affecting Amazon S3.]]></description>
    <ttl>5</ttl>

    <item>
      <title><![CDATA[RESOLVED: S3 API Error Rate Issues]]></title>
      <link>https://status.aws.amazon.com/</link>
      <pubDate>Fri, 13 Jun 2025 15:45:00 PDT</pubDate>
      <guid isPermaLink="false">https://status.aws.amazon.com/#s3-us-east-1_resolved</guid>
      <description><![CDATA[The elevated API error rates affecting Amazon S3 in US East (N. Virginia) have been resolved. All S3 operations are now functioning normally.]]></description>
    </item>

  </channel>
</rss>`

	s.Connector.Responses["http://mock.aws/s3"] = s3ResolvedData
	
	app := kingpin.New("test", "")
	cfg := maas.ServiceFeed{Name: "aws-s3", URL: "http://mock.aws/s3", Provider: "aws", Interval: 300}

	e, err := maas.NewExporter(app, s.Connector,
		maas.WithScheduledScrapers(NewFeedCollector(app, cfg)),
		maas.WithLabels(&maas.MockLabels{}),
		maas.WithArgs([]string{"--web.listen-port=0"}),
	)
	s.Require().NoError(err)
	s.Exporter = e
	s.Exporter.Start()

	// Verify service status shows resolved (ok)
	expected := "# HELP aws_s3_service_status Current service status\n" +
		"# TYPE aws_s3_service_status gauge\n" +
		"aws_s3_service_status{customer=\"\",service=\"aws-s3\",state=\"ok\"} 1\n" +
		"aws_s3_service_status{customer=\"\",service=\"aws-s3\",state=\"outage\"} 0\n" +
		"aws_s3_service_status{customer=\"\",service=\"aws-s3\",state=\"service_issue\"} 0\n"

	err = testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "aws-s3_service_status")
	s.NoError(err)

	// Verify no active incident metric
	serviceIssueMetrics := testutil.CollectAndCount(s.Exporter, "aws-s3_service_issue_info")
	s.Equal(0, serviceIssueMetrics, "Should have no active incident metrics for resolved incident")
}

func TestAWSIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AWSIntegrationTestSuite))
}