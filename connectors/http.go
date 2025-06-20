package connectors

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/sirupsen/logrus"
)

// HTTPConnector implements maas.Connector for fetching RSS feeds.
type HTTPConnector struct {
	Logger *logrus.Entry
}

// NewHTTPConnector creates a new HTTP connector.
func NewHTTPConnector() *HTTPConnector {
	return &HTTPConnector{
		Logger: logrus.WithField("component", "http_connector"),
	}
}

// The query will be the URL string.
type HTTPQuery struct {
	URL string
}

// Connect implements the maas.Connector interface (no-op for HTTP)
func (c *HTTPConnector) Connect() error {
	return nil
}

// Flags implements the maas.Connector interface (no flags needed for HTTP)
func (c *HTTPConnector) Flags(a *kingpin.Application) {
	// No flags needed for HTTP connector
}

// Execute fetches the RSS feed.
func (c *HTTPConnector) Execute(query interface{}) (interface{}, error) {
	httpQuery := query.(HTTPQuery)
	// Reuse existing FetchFeedWithRetry logic
	return FetchFeedWithRetry(httpQuery.URL, c.Logger)
}
