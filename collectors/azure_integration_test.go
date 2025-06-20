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

type AzureIntegrationTestSuite struct {
	suite.Suite
	Connector *connectors.MockHTTPConnector
	Exporter  *maas.Exporter
}

func (s *AzureIntegrationTestSuite) SetupTest() {
	s.Connector = &connectors.MockHTTPConnector{Responses: make(map[string]string)}
	s.Exporter = nil
}

func (s *AzureIntegrationTestSuite) setupExporter(feedPath, url, name, provider string) {
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

func (s *AzureIntegrationTestSuite) TestAzureStorageIssue() {
	s.setupExporter("testdata/azure_issue.rss", "http://mock.azure/storage", "azure-test", "azure")
	s.Exporter.Start()

	// Verify service status metrics
	expected := "# HELP azure_test_service_status Current service status\n" +
		"# TYPE azure_test_service_status gauge\n" +
		"azure_test_service_status{customer=\"\",service=\"azure-test\",state=\"ok\"} 0\n" +
		"azure_test_service_status{customer=\"\",service=\"azure-test\",state=\"outage\"} 0\n" +
		"azure_test_service_status{customer=\"\",service=\"azure-test\",state=\"service_issue\"} 1\n"

	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "azure-test_service_status")
	s.NoError(err)

	// Azure enhanced parser is working - status correctly shows service_issue
}

func (s *AzureIntegrationTestSuite) TestAzureVMScaleSetsGlobal() {
	// Create test data for VM Scale Sets global issue
	vmssTestData := `<?xml version="1.0" encoding="utf-8"?>
<rss version="2.0">
  <channel>
    <title>Azure Status</title>
    <item>
      <title>Service degradation: VM Scale Sets - Global</title>
      <link>https://status.azure.com/en-us/status</link>
      <pubDate>Fri, 13 Jun 2025 14:22:00 PDT</pubDate>
      <guid>vmss-global_investigating</guid>
      <description>We are investigating service degradation affecting Virtual Machine Scale Sets globally across multiple regions.</description>
    </item>
  </channel>
</rss>`

	s.Connector.Responses["http://mock.azure/vmss"] = vmssTestData
	
	app := kingpin.New("test", "")
	cfg := maas.ServiceFeed{Name: "azure-vmss", URL: "http://mock.azure/vmss", Provider: "azure", Interval: 300}

	e, err := maas.NewExporter(app, s.Connector,
		maas.WithScheduledScrapers(NewFeedCollector(app, cfg)),
		maas.WithLabels(&maas.MockLabels{}),
		maas.WithArgs([]string{"--web.listen-port=0"}),
	)
	s.Require().NoError(err)
	s.Exporter = e
	s.Exporter.Start()

	// Verify service status shows incident
	expected := "# HELP azure_vmss_service_status Current service status\n" +
		"# TYPE azure_vmss_service_status gauge\n" +
		"azure_vmss_service_status{customer=\"\",service=\"azure-vmss\",state=\"ok\"} 0\n" +
		"azure_vmss_service_status{customer=\"\",service=\"azure-vmss\",state=\"outage\"} 0\n" +
		"azure_vmss_service_status{customer=\"\",service=\"azure-vmss\",state=\"service_issue\"} 1\n"

	err = testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "azure-vmss_service_status")
	s.NoError(err)
}

func (s *AzureIntegrationTestSuite) TestAzureSQLResolved() {
	// Create test data for resolved Azure SQL issue
	sqlResolvedData := `<?xml version="1.0" encoding="utf-8"?>
<rss version="2.0">
  <channel>
    <title>Azure Status</title>
    <item>
      <title>Resolved: Azure SQL Database connectivity - West Europe</title>
      <link>https://status.azure.com/en-us/status</link>
      <pubDate>Fri, 13 Jun 2025 15:45:00 PDT</pubDate>
      <guid>sqldb-westeurope_resolved</guid>
      <description>The connectivity issues with Azure SQL Database in West Europe have been resolved. All services are now operating normally.</description>
    </item>
  </channel>
</rss>`

	s.Connector.Responses["http://mock.azure/sql"] = sqlResolvedData
	
	app := kingpin.New("test", "")
	cfg := maas.ServiceFeed{Name: "azure-sql", URL: "http://mock.azure/sql", Provider: "azure", Interval: 300}

	e, err := maas.NewExporter(app, s.Connector,
		maas.WithScheduledScrapers(NewFeedCollector(app, cfg)),
		maas.WithLabels(&maas.MockLabels{}),
		maas.WithArgs([]string{"--web.listen-port=0"}),
	)
	s.Require().NoError(err)
	s.Exporter = e
	s.Exporter.Start()

	// Verify service status shows resolved (ok)
	expected := "# HELP azure_sql_service_status Current service status\n" +
		"# TYPE azure_sql_service_status gauge\n" +
		"azure_sql_service_status{customer=\"\",service=\"azure-sql\",state=\"ok\"} 1\n" +
		"azure_sql_service_status{customer=\"\",service=\"azure-sql\",state=\"outage\"} 0\n" +
		"azure_sql_service_status{customer=\"\",service=\"azure-sql\",state=\"service_issue\"} 0\n"

	err = testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "azure-sql_service_status")
	s.NoError(err)

	// Verify no active incident metric
	serviceIssueMetrics := testutil.CollectAndCount(s.Exporter, "azure-sql_service_issue_info")
	s.Equal(0, serviceIssueMetrics, "Should have no active incident metrics for resolved incident")
}

func TestAzureIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AzureIntegrationTestSuite))
}