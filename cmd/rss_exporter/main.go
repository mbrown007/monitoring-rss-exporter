package main

import (
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"

	"github.com/mbrown007/monitoring-rss-exporter/collectors"
	"github.com/mbrown007/monitoring-rss-exporter/connectors"
)

func main() {
	// Instantiate the exporter with the new HTTP connector
	e, err := collectors.RssExporter(connectors.NewHTTPConnector())
	if err != nil {
		logrus.Fatal(err)
	}

	e.Start()
	e.Serve()

	sentry.Flush(5 * time.Second)
}
