package collectors

import (
	"os"
	"strings"

	"github.com/alecthomas/kingpin/v2"
	maas "github.com/mbrown007/monitoring-rss-exporter/monitoring-maas"
	"gopkg.in/yaml.v3"
)

// Config defines exporter settings loaded from YAML configuration.
type Config struct {
	Services []maas.ServiceFeed `yaml:"services"`
}


// NewRssExporter constructs a maas exporter with feed scrapers based on config.
func NewRssExporter(c maas.Connector, options ...func(*maas.Exporter)) (*maas.Exporter, error) {
	app := kingpin.New("rss_exporter", "Exporter for RSS/Atom status feeds.").DefaultEnvars()
	app.Flag("config.file", "RSS exporter configuration file.").Default("config.yml").String()

	// Read and parse the config file using default or specified path
	configPath := "config.yml"
	// Look for --config.file in args to override default
	for i, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--config.file=") {
			configPath = strings.TrimPrefix(arg, "--config.file=")
		} else if arg == "--config.file" && i+1 < len(os.Args[1:]) {
			configPath = os.Args[i+2] // next argument
		}
	}

	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, err
	}

	// Create scrapers based on config
	scrapers := []*maas.ScheduledScraper{}
	for _, svc := range cfg.Services {
		if svc.Interval <= 0 {
			svc.Interval = 300
		}
		scrapers = append(scrapers, NewFeedCollector(app, svc))
	}

	// Create the exporter with scrapers
	options = append(options, maas.WithScheduledScrapers(scrapers...))
	return maas.NewExporter(app, c, options...)
}
