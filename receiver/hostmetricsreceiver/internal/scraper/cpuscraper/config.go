// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cpuscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper"

import (
	// "errors"
	//
	// "go.opentelemetry.io/collector/confmap"
	//
	// "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/featuregates"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/internal/metadata_semconv"
)

// Config relating to CPU Metric Scraper.
type Config struct {
	metadata.MetricsBuilderConfig `mapstructure:",squash"`
	Semconv                       metadata_semconv.MetricsConfig `mapstructure:"metrics_semconv"`
}
