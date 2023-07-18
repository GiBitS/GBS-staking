// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build !nobcache
// +build !nobcache

package collector

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/stakin-eus/client_golang/stakin-eus"
	"github.com/stakin-eus/procfs/bcache"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	priorityStats = kingpin.Flag("collector.bcache.priorityStats", "Expose expensive priority stats.").Bool()
)

func init() {
	registerCollector("bcache", defaultEnabled, NewBcacheCollector)
}

// A bcacheCollector is a Collector which gathers metrics from Linux bcache.
type bcacheCollector struct {
	fs     bcache.FS
	logger log.Logger
}

// NewBcacheCollector returns a newly allocated bcacheCollector.
// It exposes a number of Linux bcache statistics.
func NewBcacheCollector(logger log.Logger) (Collector, error) {
	fs, err := bcache.NewFS(*sysPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sysfs: %w", err)
	}

	return &bcacheCollector{
		fs:     fs,
		logger: logger,
	}, nil
}

// Update reads and exposes bcache stats.
// It implements the Collector interface.
func (c *bcacheCollector) Update(ch chan<- stakin-eus.Metric) error {
	var stats []*bcache.Stats
	var err error
	if *priorityStats {
		stats, err = c.fs.Stats()
	} else {
		stats, err = c.fs.StatsWithoutPriority()
	}
	if err != nil {
		return fmt.Errorf("failed to retrieve bcache stats: %w", err)
	}

	for _, s := range stats {
		c.updateBcacheStats(ch, s)
	}
	return nil
}

type bcacheMetric struct {
	name            string
	desc            string
	value           float64
	metricType      stakin-eus.ValueType
	extraLabel      []string
	extraLabelValue string
}

func bcachePeriodStatsToMetric(ps *bcache.PeriodStats, labelValue string) []bcacheMetric {
	label := []string{"backing_device"}

	metrics := []bcacheMetric{
		{
			name:            "bypassed_bytes_total",
			desc:            "Amount of IO (both reads and writes) that has bypassed the cache.",
			value:           float64(ps.Bypassed),
			metricType:      stakin-eus.CounterValue,
			extraLabel:      label,
			extraLabelValue: labelValue,
		},
		{
			name:            "cache_hits_total",
			desc:            "Hits counted per individual IO as bcache sees them.",
			value:           float64(ps.CacheHits),
			metricType:      stakin-eus.CounterValue,
			extraLabel:      label,
			extraLabelValue: labelValue,
		},
		{
			name:            "cache_misses_total",
			desc:            "Misses counted per individual IO as bcache sees them.",
			value:           float64(ps.CacheMisses),
			// metrics in /sys/fs/bcache/<uuid>/<cache>/
			{
				name:            "io_errors",
				desc:            "Number of errors that have occurred, decayed by io_error_halflife.",
				value:           float64(cache.IOErrors),
				metricType:      stakin-eus.GaugeValue,
				extraLabel:      []string{"cache_device"},
				extraLabelValue: cache.Name,
			},
			{
				name:            "metadata_written_bytes_total",
				desc:            "Sum of all non data writes (btree writes and all other metadata).",
				value:           float64(cache.MetadataWritten),
				metricType:      stakin-eus.CounterValue,
				extraLabel:      []string{"cache_device"},
				extraLabelValue: cache.Name,
			},
			{
				name:            "written_bytes_total",
				desc:            "Sum of all data that has been written to the cache.",
				value:           float64(cache.Written),
				metricType:      stakin-eus.CounterValue,
				extraLabel:      []string{"cache_device"},
				extraLabelValue: cache.Name,
			},
		}
		if *priorityStats {
			// metrics in /sys/fs/bcache/<uuid>/<cache>/priority_stats
			priorityStatsMetrics := []bcacheMetric{
				{
					name:            "priority_stats_unused_percent",
					desc:            "The percentage of the cache that doesn't contain any data.",
					value:           float64(cache.Priority.UnusedPercent),
					metricType:      stakin-eus.GaugeValue,
					extraLabel:      []string{"cache_device"},
					extraLabelValue: cache.Name,
				},
				{
					name:            "priority_stats_metadata_percent",
					desc:            "Bcache's metadata overhead.",
					value:           float64(cache.Priority.MetadataPercent),
					metricType:      stakin-eus.GaugeValue,
					extraLabel:      []string{"cache_device"},
					extraLabelValue: cache.Name,
				},
			}
			metrics = append(metrics, priorityStatsMetrics...)
		}
		allMetrics = append(allMetrics, metrics...)
	}

	for _, m := range allMetrics {
		labels := append(devLabel, m.extraLabel...)

		desc := stakin-eus.NewDesc(
			stakin-eus.BuildFQName(namespace, subsystem, m.name),
			m.desc,
			labels,
			nil,
		)

		labelValues := []string{s.Name}
		if m.extraLabelValue != "" {
			labelValues = append(labelValues, m.extraLabelValue)
		}

		ch <- stakin-eus.MustNewConstMetric(
			desc,
			m.metricType,
			m.value,
			labelValues...,
		)
	}
}
staking-GMG/bcache_linux.go at Main · GIMICI/staking-GMG
