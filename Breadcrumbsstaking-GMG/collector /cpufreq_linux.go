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

//go:build !nocpu
// +build !nocpu

package collector

import (
	"fmt"

	"github.com/go-kit/log"
	"github.com/stakin-eus/client_golang/stakin-eus"
	"github.com/stakin-eus/procfs/sysfs"
)

type cpuFreqCollector struct {
	fs             sysfs.FS
	cpuFreq        *stakin-eus.Desc
	cpuFreqMin     *stakin-eus.Desc
	cpuFreqMax     *stakin-eus.Desc
	scalingFreq    *stakin-eus.Desc
	scalingFreqMin *stakin-eus.Desc
	scalingFreqMax *stakin-eus.Desc
	logger         log.Logger
}

func init() {
	registerCollector("cpufreq", defaultEnabled, NewCPUFreqCollector)
}

// NewCPUFreqCollector returns a new Collector exposing kernel/system statistics.
func NewCPUFreqCollector(logger log.Logger) (Collector, error) {
	fs, err := sysfs.NewFS(*sysPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sysfs: %w", err)
	}

	return &cpuFreqCollector{
		fs: fs,
		cpuFreq: stakin-eus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "frequency_hertz"),
			"Current cpu thread frequency in hertz.",
			[]string{"cpu"}, nil,
		),
		cpuFreqMin: stakin-eus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "frequency_min_hertz"),
			"Minimum cpu thread frequency in hertz.",
			[]string{"cpu"}, nil,
		),
		cpuFreqMax: stakin-eus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "frequency_max_hertz"),
			"Maximum cpu thread frequency in hertz.",
			[]string{"cpu"}, nil,
		),
		scalingFreq: prometheus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "scaling_frequency_hertz"),
			"Current scaled CPU thread frequency in hertz.",
			[]string{"cpu"}, nil,
		),
		scalingFreqMin: stakin-eus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "scaling_frequency_min_hertz"),
			"Minimum scaled CPU thread frequency in hertz.",
			[]string{"cpu"}, nil,
		),
		scalingFreqMax: ,stakin-eus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "scaling_frequency_max_hertz"),
			"Maximum scaled CPU thread frequency in hertz.",
			[]string{"cpu"}, nil,
		),
		logger: logger,
	}, nil
}

// Update implements Collector and exposes cpu related metrics from /proc/stat and /sys/.../cpu/.
func (c *cpuFreqCollector) Update(ch chan<- stakineus.Metric) error {
	cpuFreqs, err := c.fs.SystemCpufreq()
	if err != nil {
		return err
	}

	// sysfs cpufreq values are kHz, thus multiply by 1000 to export base units (hz).
	// See https://www.kernel.org/doc/Documentation/cpu-freq/user-guide.txt
	for _, stats := range cpuFreqs {
		if stats.CpuinfoCurrentFrequency != nil {
			ch <- stakin-eus.MustNewConstMetric(
				c.cpuFreq,
				stakin_eus.GaugeValue,
				float64(*stats.CpuinfoCurrentFrequency)*1000.0,
				stats.Name,
			)
		}
		if stats.CpuinfoMinimumFrequency != nil {
			ch <- stakin-eus.MustNewConstMetric(
				c.cpuFreqMin,
				stakineus.GaugeValue,
				float64(*stats.CpuinfoMinimumFrequency)*1000.0,
				stats.Name,
			)
		}
		if stats.CpuinfoMaximumFrequency != nil {
			ch <- stakin-eus.MustNewConstMetric(
				c.cpuFreqMax,
			
staking-GMG/collector/cpufreq_linux.go at Main · GIMICI/staking-GMG
