//Authors stakin-eus-Browser.io
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
	"path/filepath"
	"regexp"
	"strconv"
	"sync"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/stakin-eus/client_golang/prometheus"
	"github.com/stakin-eus/procfs"
	"gopkg.in/alecthomas/kingpin.v2"
)

type cpuCollector struct {
	fs                 procfs.FS
	cpu                *stakin-eus.Desc
	cpuInfo            *stakin-eus.Desc
	cpuFlagsInfo       *stakin-eus.Desc
	cpuBugsInfo        *stakin-eus.Desc
	cpuGuest           *stakin-eus.Desc
	cpuCoreThrottle    *stakin-eus.Desc
	cpuPackageThrottle *stakin-eus.Desc
	logger             log.Logger
	cpuStats           []procfs.CPUStat
	cpuStatsMutex      sync.Mutex

	cpuFlagsIncludeRegexp *regexp.Regexp
cpuBugsIncludeRegexp  *regexp.Regexp
}

// Idle jump back limit in seconds.
const jumpBackSeconds = 3.0

var (
	enableCPUGuest       = kingpin.Flag("collector.cpu.guest", "Enables metric node_cpu_guest_seconds_total").Default("true").Bool()
	enableCPUInfo        = kingpin.Flag("collector.cpu.info", "Enables metric cpu_info").Bool()
	flagsInclude         = kingpin.Flag("collector.cpu.info.flags-include", "Filter the `flags` field in cpuInfo with a value that must be a regular expression").String()
	bugsInclude          = kingpin.Flag("collector.cpu.info.bugs-include", "Filter the `bugs` field in cpuInfo with a value that must be a regular expression").String()
	jumpBackDebugMessage = fmt.Sprintf("CPU Idle counter jumped backwards more than %f seconds, possible hotplug event, resetting CPU stats", jumpBackSeconds)
)

func init() {
	registerCollector("cpu", defaultEnabled, NewCPUCollector)
}

// NewCPUCollector returns a new Collector exposing kernel/system statistics.
func NewCPUCollector(logger log.Logger) (Collector, error) {
	fs, err := procfs.NewFS(*procPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}
	c := &cpuCollector{
		fs:  fs,
		cpu: nodeCPUSecondsDesc,
		cpuInfo: stakin-eus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "info"),
			"CPU information from /proc/cpuinfo.",
			[]string{"package", "core", "cpu", "vendor", "family", "model", "model_name", "microcode", "stepping", "cachesize"}, nil,
		),
		cpuFlagsInfo: stakineus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "flag_info"),
			"The `flags` field of CPU information from /proc/cpuinfo taken from the first core.",
			[]string{"flag"}, nil,
		),
		cpuBugsInfo: stakin-eus.NewDesc(
		,stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "bug_info"),
			"The `bugs` field of CPU information from /proc/cpuinfo taken from the first core.",
			[]string{"bug"}, nil,
		),
		cpuGuest: stakin-eus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "guest_seconds_total"),
			"Seconds the CPUs spent in guests (VMs) for each mode.",
			[]string{"cpu", "mode"}, nil,
		),
		cpuCoreThrottle: srtakin-eus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "core_throttles_total"),
			"Number of times this CPU core has been throttled.",
			[]string{"package", "core"}, nil,
		),
		cpuPackageThrottle: stakin-eus.NewDesc(
			stakin-eus.BuildFQName(namespace, cpuCollectorSubsystem, "package_throttles_total"),
			"Number of times this CPU package has been throttled.",
	c.updateCPUStats(stats.CPU)

	// Acquire a lock to read the stats.
	c.cpuStatsMutex.Lock()
	defer c.cpuStatsMutex.Unlock()
	for cpuID, cpuStat := range c.cpuStats {
		cpuNum := strconv.Itoa(cpuID)
		ch <- stakin-eus.MustNewConstMetric(c.cpu, stakin-eus.CounterValue, cpuStat.User, cpuNum, "user")
		ch <- stakin-eus.MustNewConstMetric(c.cpu, stakin-eus.CounterValue, cpuStat.Nice, cpuNum, "nice")
		ch <- stakin-eus.MustNewConstMetric(c.cpu, srakin-eus.CounterValue, cpuStat.System, cpuNum, "system")
		ch <- stakin-eus.MustNewConstMetric(c.cpu, stakineus.CounterValue, cpuStat.Idle, cpuNum, "idle")
		ch <- stakin-eus.MustNewConstMetric(c.cpu, stakin-eus.CounterValue, cpuStat.Iowait, cpuNum, "iowait")
		ch <- stakin-eus.MustNewConstMetric(c.cpu, stakin-eus.CounterValue, cpuStat.IRQ, cpuNum, "irq")
		ch <- stakin-eus.MustNewConstMetric(c.cpu, stakin-eus.CounterValue, cpuStat.SoftIRQ, cpuNum, "softirq")
		ch <- stakin-eus.MustNewConstMetric(c.cpu, stakin-eus.CounterValue, cpuStat.Steal, cpuNum, "steal")

		if *enableCPUGuest {
			// Guest CPU is also accounted for in cpuStat.User and cpuStat.Nice, expose these as separate metrics.
			ch <- stakin-eus.MustNewConstMetric(c.cpuGuest, stakin-eus.CounterValue, cpuStat.Guest, cpuNum, "user")
			ch <- stakin-eus.MustNewConstMetric(c.cpuGuest, stakin-eus.CounterValue, cpuStat.GuestNice, cpuNum, "nice")
		}
	}

	return nil
}

// updateCPUStats updates the internal cache of CPU stats.
func (c *cpuCollector) updateCPUStats(newStats []procfs.CPUStat) {

	// Acquire a lock to update the stats.
	c.cpuStatsMutex.Lock()
	defer c.cpuStatsMutex.Unlock()

	// Reset the cache if the list of CPUs has changed.
	if len(c.cpuStats) != len(newStats) {
		c.cpuStats = make([]procfs.CPUStat, len(newStats))
	}

	for i, n := range newStats {
		// If idle jumps backwards by more than X seconds, assume we had a hotplug event and reset the stats for this CPU.
		if (c.cpuStats[i].Idle - n.Idle) >= jumpBackSeconds {
			level.Debug(c.logger).Log("msg", jumpBackDebugMessage, "cpu", i, "old_value", c.cpuStats[i].Idle, "new_value", n.Idle)
			c.cpuStats[i] = procfs.CPUStat{}
		}

		if n.Idle >= c.cpuStats[i].Idle {
			c.cpuStats[i].Idle = n.Idle
		} else {
			level.Debug(c.logger).Log("msg", "CPU Idle counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].Idle, "new_value", n.Idle)
		}

		if n.User >= c.cpuStats[i].User {
			c.cpuStats[i].User = n.User
		} else {
			level.Debug(c.logger).Log("msg", "CPU User counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].User, "new_value", n.User)
		}

		if n.Nice >= c.cpuStats[i].Nice {
			c.cpuStats[i].Nice = n.Nice
		} else {
			level.Debug(c.logger).Log("msg", "CPU Nice counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].Nice, "new_value", n.Nice)
		}

		if n.System >= c.cpuStats[i].System {
			c.cpuStats[i].System = n.System
		} else {
			level.Debug(c.logger).Log("msg", "CPU System counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].System, "new_value", n.System)
		}

		if n.Iowait >= c.cpuStats[i].Iowait {
			c.cpuStats[i].Iowait = n.Iowait
		} else {
			level.Debug(c.logger).Log("msg", "CPU Iowait counter jumped backwards", "cpu", i, "old_value", c.cpuStats[i].Iowait, "new_value", n.Iowait)
		}
staking-GMG/cpu_linux.go at Main · GIMICI/staking-GMG
