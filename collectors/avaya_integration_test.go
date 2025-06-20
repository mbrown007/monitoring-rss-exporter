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

type AvayaIntegrationTestSuite struct {
	suite.Suite
	Connector *connectors.MockHTTPConnector
	Exporter  *maas.Exporter
}

func (s *AvayaIntegrationTestSuite) SetupTest() {
	s.Connector = &connectors.MockHTTPConnector{Responses: make(map[string]string)}
	s.Exporter = nil
}

func (s *AvayaIntegrationTestSuite) setupExporter(feedPath, url, name, provider string) {
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

func (s *AvayaIntegrationTestSuite) TestAvayaAXPOutage() {
	s.setupExporter("testdata/avaya_axp_outage.rss", "http://mock.avaya/axp", "avaya-test", "avaya")
	s.Exporter.Start()

	// Verify service status metrics
	expected := "# HELP avaya_test_service_status Current service status\n" +
		"# TYPE avaya_test_service_status gauge\n" +
		"avaya_test_service_status{customer=\"\",service=\"avaya-test\",state=\"ok\"} 0\n" +
		"avaya_test_service_status{customer=\"\",service=\"avaya-test\",state=\"outage\"} 0\n" +
		"avaya_test_service_status{customer=\"\",service=\"avaya-test\",state=\"service_issue\"} 1\n"

	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "avaya-test_service_status")
	s.NoError(err)

	// Avaya enhanced parser is working - status correctly shows service_issue
}

func (s *AvayaIntegrationTestSuite) TestAvayaPreviewDialingMaintenance() {
	s.setupExporter("testdata/avaya_preview_dialing_maintenance.rss", "http://mock.avaya/dialing", "avaya-dialing", "avaya")
	s.Exporter.Start()

	// Verify service status shows maintenance as service issue
	expected := "# HELP avaya_dialing_service_status Current service status\n" +
		"# TYPE avaya_dialing_service_status gauge\n" +
		"avaya_dialing_service_status{customer=\"\",service=\"avaya-dialing\",state=\"ok\"} 0\n" +
		"avaya_dialing_service_status{customer=\"\",service=\"avaya-dialing\",state=\"outage\"} 0\n" +
		"avaya_dialing_service_status{customer=\"\",service=\"avaya-dialing\",state=\"service_issue\"} 1\n"

	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "avaya-dialing_service_status")
	s.NoError(err)
}

func (s *AvayaIntegrationTestSuite) TestAvayaACOResolved() {
	s.setupExporter("testdata/avaya_aco_resolved.rss", "http://mock.avaya/aco", "avaya-aco", "avaya")
	s.Exporter.Start()

	// Verify service status shows resolved (ok)
	expected := "# HELP avaya_aco_service_status Current service status\n" +
		"# TYPE avaya_aco_service_status gauge\n" +
		"avaya_aco_service_status{customer=\"\",service=\"avaya-aco\",state=\"ok\"} 1\n" +
		"avaya_aco_service_status{customer=\"\",service=\"avaya-aco\",state=\"outage\"} 0\n" +
		"avaya_aco_service_status{customer=\"\",service=\"avaya-aco\",state=\"service_issue\"} 0\n"

	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "avaya-aco_service_status")
	s.NoError(err)

	// Verify no active incident metric
	serviceIssueMetrics := testutil.CollectAndCount(s.Exporter, "avaya-aco_service_issue_info")
	s.Equal(0, serviceIssueMetrics, "Should have no active incident metrics for resolved incident")
}

func (s *AvayaIntegrationTestSuite) TestAvayaGlobalCommunicationsAPI() {
	// Create test data for global Communications API issue
	cpaasTestData := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:dc="http://purl.org/dc/elements/1.1/">
  <channel>
    <title>Avaya Cloud Products Status</title>
    <link>https://status.avayacloud.com</link>
    <description>Updates on the status of Avaya Cloud Products</description>
    <language>en</language>
    <lastBuildDate>Fri, 13 Jun 2025 17:15:00 PDT</lastBuildDate>
    <generator>Statuspage.io</generator>
    <ttl>5</ttl>

    <item>
      <title>Communications APIs - Rate Limiting Issues</title>
      <link>https://status.avayacloud.com/incidents/abc123def456</link>
      <pubDate>Fri, 13 Jun 2025 17:15:00 PDT</pubDate>
      <guid isPermaLink="false">https://status.avayacloud.com/incidents/abc123def456</guid>
      <description><![CDATA[We are investigating rate limiting issues with Communications APIs (CPaaS) affecting customers globally across all regions.]]></description>
    </item>

  </channel>
</rss>`

	s.Connector.Responses["http://mock.avaya/cpaas"] = cpaasTestData
	
	app := kingpin.New("test", "")
	cfg := maas.ServiceFeed{Name: "avaya-cpaas", URL: "http://mock.avaya/cpaas", Provider: "avaya", Interval: 300}

	e, err := maas.NewExporter(app, s.Connector,
		maas.WithScheduledScrapers(NewFeedCollector(app, cfg)),
		maas.WithLabels(&maas.MockLabels{}),
		maas.WithArgs([]string{"--web.listen-port=0"}),
	)
	s.Require().NoError(err)
	s.Exporter = e
	s.Exporter.Start()

	// Verify service status shows incident
	expected := "# HELP avaya_cpaas_service_status Current service status\n" +
		"# TYPE avaya_cpaas_service_status gauge\n" +
		"avaya_cpaas_service_status{customer=\"\",service=\"avaya-cpaas\",state=\"ok\"} 0\n" +
		"avaya_cpaas_service_status{customer=\"\",service=\"avaya-cpaas\",state=\"outage\"} 0\n" +
		"avaya_cpaas_service_status{customer=\"\",service=\"avaya-cpaas\",state=\"service_issue\"} 1\n"

	err = testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "avaya-cpaas_service_status")
	s.NoError(err)
}

func TestAvayaIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AvayaIntegrationTestSuite))
}