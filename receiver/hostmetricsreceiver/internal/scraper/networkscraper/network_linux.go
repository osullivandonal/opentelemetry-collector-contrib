// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:build linux

package networkscraper // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/networkscraper"

import (
	"context"
	"fmt"
	"time"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/hostmetricsreceiver/internal/scraper/networkscraper/internal/metadata"
	"github.com/prometheus/procfs"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

const (
	conntrackMetricsLen    = 2
	socketBufferMetricsLen = 4
)

var allTCPStates = []string{
	"CLOSE_WAIT",
	"CLOSE",
	"CLOSING",
	"DELETE",
	"ESTABLISHED",
	"FIN_WAIT_1",
	"FIN_WAIT_2",
	"LAST_ACK",
	"LISTEN",
	"SYN_SENT",
	"SYN_RECV",
	"TIME_WAIT",
}

func (s *networkScraper) recordNetworkConntrackMetrics(ctx context.Context) error {
	if !s.config.Metrics.SystemNetworkConntrackCount.Enabled && !s.config.Metrics.SystemNetworkConntrackMax.Enabled {
		return nil
	}
	now := pcommon.NewTimestampFromTime(time.Now())
	conntrack, err := s.conntrack(ctx)
	if err != nil {
		return fmt.Errorf("failed to read conntrack info: %w", err)
	}
	s.mb.RecordSystemNetworkConntrackCountDataPoint(now, conntrack[0].ConnTrackCount)
	s.mb.RecordSystemNetworkConntrackMaxDataPoint(now, conntrack[0].ConnTrackMax)
	return nil
}

func (s *networkScraper) recordNetworkSocketBufferMetrics(ctx context.Context) error {
	if !s.config.Metrics.SystemNetworkSocketBufferReceive.Enabled && !s.config.Metrics.SystemNetworkSocketBufferTransmit.Enabled {
		return nil
	}
	fs, err := procfs.NewDefaultFS()
	if err != nil {
		return err
	}

	now := pcommon.NewTimestampFromTime(time.Now())
	tcpSummary, err := fs.NetTCPSummary()
	if err == nil {
		s.mb.RecordSystemNetworkSocketBufferReceiveDataPoint(now, int64(tcpSummary.RxQueueLength), metadata.AttributeProtocolTcp)
		s.mb.RecordSystemNetworkSocketBufferTransmitDataPoint(now, int64(tcpSummary.TxQueueLength), metadata.AttributeProtocolTcp)
	}

	udpSummary, err := fs.NetUDPSummary()
	if err == nil {
		s.mb.RecordSystemNetworkSocketBufferReceiveDataPoint(now, int64(udpSummary.RxQueueLength), metadata.AttributeProtocolUdp)
		s.mb.RecordSystemNetworkSocketBufferTransmitDataPoint(now, int64(udpSummary.TxQueueLength), metadata.AttributeProtocolUdp)
	}
	return nil
}
