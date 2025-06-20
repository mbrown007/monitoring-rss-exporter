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

type GenesysIntegrationTestSuite struct {
	suite.Suite
	Connector *connectors.MockHTTPConnector
	Exporter  *maas.Exporter
}

func (s *GenesysIntegrationTestSuite) SetupTest() {
	s.Connector = &connectors.MockHTTPConnector{Responses: make(map[string]string)}
	s.Exporter = nil
}

func (s *GenesysIntegrationTestSuite) setupExporter(feedPath, url, name, provider string) {
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

func (s *GenesysIntegrationTestSuite) TestGenysysTTSIssue() {
	s.setupExporter("testdata/genesys_tts_issue.atom", "http://mock.genesys/tts", "genesys-test", "genesyscloud")
	s.Exporter.Start()

	// Verify service status metrics
	expected := "# HELP genesys_test_service_status Current service status\n" +
		"# TYPE genesys_test_service_status gauge\n" +
		"genesys_test_service_status{customer=\"\",service=\"genesys-test\",state=\"ok\"} 0\n" +
		"genesys_test_service_status{customer=\"\",service=\"genesys-test\",state=\"outage\"} 0\n" +
		"genesys_test_service_status{customer=\"\",service=\"genesys-test\",state=\"service_issue\"} 1\n"
	
	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "genesys-test_service_status")
	s.NoError(err)
}

func (s *GenesysIntegrationTestSuite) TestGenesysWhatsAppMultiRegion() {
	s.setupExporter("testdata/genesys_whatsapp_multi_region.atom", "http://mock.genesys/whatsapp", "genesys-whatsapp", "genesyscloud")
	s.Exporter.Start()

	// Verify service status shows incident
	expected := "# HELP genesys_whatsapp_service_status Current service status\n" +
		"# TYPE genesys_whatsapp_service_status gauge\n" +
		"genesys_whatsapp_service_status{customer=\"\",service=\"genesys-whatsapp\",state=\"ok\"} 0\n" +
		"genesys_whatsapp_service_status{customer=\"\",service=\"genesys-whatsapp\",state=\"outage\"} 0\n" +
		"genesys_whatsapp_service_status{customer=\"\",service=\"genesys-whatsapp\",state=\"service_issue\"} 1\n"
	
	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "genesys-whatsapp_service_status")
	s.NoError(err)
}

func (s *GenesysIntegrationTestSuite) TestGenesysAnalyticsResolved() {
	s.setupExporter("testdata/genesys_analytics_resolved.atom", "http://mock.genesys/analytics", "genesys-analytics", "genesyscloud")
	s.Exporter.Start()

	// Verify service status shows resolved (ok)
	expected := "# HELP genesys_analytics_service_status Current service status\n" +
		"# TYPE genesys_analytics_service_status gauge\n" +
		"genesys_analytics_service_status{customer=\"\",service=\"genesys-analytics\",state=\"ok\"} 1\n" +
		"genesys_analytics_service_status{customer=\"\",service=\"genesys-analytics\",state=\"outage\"} 0\n" +
		"genesys_analytics_service_status{customer=\"\",service=\"genesys-analytics\",state=\"service_issue\"} 0\n"
	
	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "genesys-analytics_service_status")
	s.NoError(err)

	// Verify no active incident metric
	serviceIssueMetrics := testutil.CollectAndCount(s.Exporter, "genesys-analytics_service_issue_info")
	s.Equal(0, serviceIssueMetrics, "Should have no active incident metrics for resolved incident")
}

func TestGenesysIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(GenesysIntegrationTestSuite))
}