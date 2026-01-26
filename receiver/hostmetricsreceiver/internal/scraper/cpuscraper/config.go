// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cpuscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper"

import (
	"errors"

	"go.opentelemetry.io/collector/confmap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/featuregates"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/internal/metadata_semconv"
)

// Config relating to CPU Metric Scraper.
type Config struct {
	metadata.MetricsBuilderConfig `mapstructure:",squash"`
	Semconv                       metadata_semconv.MetricsBuilderConfig `mapstructure:"semconv"`
}

// Unmarshal handles backward compatibility and validation
func (c *Config) Unmarshal(conf *confmap.Conf) error {
	if err := conf.Unmarshal(c); err != nil {
		return err
	}

	// Validation: if both gates are off, that's an error
	if !featuregates.EmitOldMetrics.IsEnabled() && !featuregates.EmitSemconvMetrics.IsEnabled() {
		return errors.New("at least one of emitOldMetrics or emitSemconvMetrics must be enabled")
	}

	return nil
}
