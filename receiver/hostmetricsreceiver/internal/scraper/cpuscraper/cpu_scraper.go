// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package cpuscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper"

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/scraper"
	"go.opentelemetry.io/collector/scraper/scrapererror"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/featuregates"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/internal/metadata"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/internal/metadata_semconv"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/cpuscraper/ucal"
)

const (
	metricsLen = 2
	hzInAMHz   = 1_000_000
)

// cpuScraper for CPU Metrics
type cpuScraper struct {
	settings  scraper.Settings
	config    *Config
	mb        *metadata.MetricsBuilder
	mbSemconv *metadata_semconv.MetricsBuilder
	ucal      *ucal.CPUUtilizationCalculator

	// for mocking
	bootTime func(context.Context) (uint64, error)
	times    func(context.Context, bool) ([]cpu.TimesStat, error)
	now      func() time.Time
}

type cpuInfo struct {
	frequency float64
	processor uint
}

// newCPUScraper creates a set of CPU related metrics
func newCPUScraper(_ context.Context, settings scraper.Settings, cfg *Config) *cpuScraper {
	s := &cpuScraper{
		config:   cfg,
		settings: settings,
		ucal:     &ucal.CPUUtilizationCalculator{},
		bootTime: host.BootTimeWithContext,
		times:    cpu.TimesWithContext,
		now:      time.Now,
	}
	return s
}

func (s *cpuScraper) start(ctx context.Context, _ component.Host) error {
	bootTime, err := s.bootTime(ctx)
	if err != nil {
		return err
	}

	startTime := pcommon.Timestamp(bootTime * 1e9)

	// Initialize legacy builder if gate is enabled
	if featuregates.EmitOldMetrics.IsEnabled() {
		s.mb = metadata.NewMetricsBuilder(
			s.config.MetricsBuilderConfig,
			s.settings,
			metadata.WithStartTime(startTime),
		)
	}

	// Initialize semconv builder if gate is enabled
	if featuregates.EmitSemconvMetrics.IsEnabled() {
		s.mbSemconv = metadata_semconv.NewMetricsBuilder(
			metadata_semconv.MetricsBuilderConfig{Metrics: s.config.Semconv},
			s.settings,
			metadata_semconv.WithStartTime(startTime),
		)
	}
	return nil
}

func (s *cpuScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	now := pcommon.NewTimestampFromTime(s.now())
	cpuTimes, err := s.times(ctx, true /*percpu=*/)
	if err != nil {
		return pmetric.NewMetrics(), scrapererror.NewPartialScrapeError(err, metricsLen)
	}

	// Record legacy metrics
	if s.mb != nil {
		for _, cpuTime := range cpuTimes {
			s.recordCPUTimeStateDataPoints(now, cpuTime)
		}
	}

	err = s.ucal.CalculateAndRecord(now, cpuTimes, s.recordCPUUtilization)
	if err != nil {
		return pmetric.NewMetrics(), scrapererror.NewPartialScrapeError(err, metricsLen)
	}

	if s.config.Metrics.SystemCPUPhysicalCount.Enabled {
		numCPU, err := cpu.Counts(false)
		if err != nil {
			return pmetric.NewMetrics(), scrapererror.NewPartialScrapeError(err, metricsLen)
		}
		s.mb.RecordSystemCPUPhysicalCountDataPoint(now, int64(numCPU))
	}

	if s.config.Metrics.SystemCPULogicalCount.Enabled {
		numCPU, err := cpu.Counts(true)
		if err != nil {
			return pmetric.NewMetrics(), scrapererror.NewPartialScrapeError(err, metricsLen)
		}
		s.mb.RecordSystemCPULogicalCountDataPoint(now, int64(numCPU))
	}

	if s.config.Metrics.SystemCPUFrequency.Enabled {
		cpuInfos, err := s.getCPUInfo()
		if err != nil {
			return pmetric.NewMetrics(), scrapererror.NewPartialScrapeError(err, metricsLen)
		}
		for _, cInfo := range cpuInfos {
			s.mb.RecordSystemCPUFrequencyDataPoint(now, cInfo.frequency*hzInAMHz, fmt.Sprintf("cpu%d", cInfo.processor))
		}
	}

	if s.mbSemconv != nil && s.config.Semconv.SystemCPUFrequency.Enabled {
		cpuInfos, err := s.getCPUInfo()
		if err != nil {
			return pmetric.NewMetrics(), scrapererror.NewPartialScrapeError(err, metricsLen)
		}
		for _, cInfo := range cpuInfos {
			s.mbSemconv.RecordSystemCPUFrequencyDataPoint(
				now,
				cInfo.frequency*hzInAMHz,
				fmt.Sprintf("cpu%d", cInfo.processor),
				fmt.Sprintf("test attribute, CPU freq: %v", cInfo.frequency),
			)
		}
	}

	return s.emitMetrics(), nil
}

func (s *cpuScraper) emitMetrics() pmetric.Metrics {
	var metrics pmetric.Metrics

	if s.mb != nil {
		metrics = s.mb.Emit()
	}

	if s.mbSemconv != nil {
		semconvMetrics := s.mbSemconv.Emit()
		if metrics.ResourceMetrics().Len() == 0 {
			metrics = semconvMetrics
		} else {
			// Merge semconv metrics into the result
			semconvMetrics.ResourceMetrics().MoveAndAppendTo(metrics.ResourceMetrics())
		}
	}

	return metrics
}
