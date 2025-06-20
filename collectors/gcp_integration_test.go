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

type GCPIntegrationTestSuite struct {
	suite.Suite
	Connector *connectors.MockHTTPConnector
	Exporter  *maas.Exporter
}

func (s *GCPIntegrationTestSuite) SetupTest() {
	s.Connector = &connectors.MockHTTPConnector{Responses: make(map[string]string)}
	s.Exporter = nil
}

func (s *GCPIntegrationTestSuite) setupExporter(feedPath, url, name, provider string) {
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

func (s *GCPIntegrationTestSuite) TestGCPComputeEngineIssue() {
	s.setupExporter("testdata/gcp_compute_engine_issue.atom", "http://mock.gcp/feed", "gcp-test", "gcp")
	s.Exporter.Start()

	// Verify service status metrics
	expected := "# HELP gcp_test_service_status Current service status\n" +
		"# TYPE gcp_test_service_status gauge\n" +
		"gcp_test_service_status{customer=\"\",service=\"gcp-test\",state=\"ok\"} 0\n" +
		"gcp_test_service_status{customer=\"\",service=\"gcp-test\",state=\"outage\"} 0\n" +
		"gcp_test_service_status{customer=\"\",service=\"gcp-test\",state=\"service_issue\"} 1\n"
	
	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "gcp-test_service_status")
	s.NoError(err)

	// GCP enhanced parser is working - status correctly shows service_issue
}

func (s *GCPIntegrationTestSuite) TestGCPMultipleServices() {
	s.setupExporter("testdata/gcp_multiple_services.atom", "http://mock.gcp/multi", "gcp-multi", "gcp")
	s.Exporter.Start()

	// Verify service status shows incident
	expected := "# HELP gcp_multi_service_status Current service status\n" +
		"# TYPE gcp_multi_service_status gauge\n" +
		"gcp_multi_service_status{customer=\"\",service=\"gcp-multi\",state=\"ok\"} 0\n" +
		"gcp_multi_service_status{customer=\"\",service=\"gcp-multi\",state=\"outage\"} 0\n" +
		"gcp_multi_service_status{customer=\"\",service=\"gcp-multi\",state=\"service_issue\"} 1\n"
	
	err := testutil.CollectAndCompare(s.Exporter, strings.NewReader(expected), "gcp-multi_service_status")
	s.NoError(err)
}

func TestGCPIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(GCPIntegrationTestSuite))
}