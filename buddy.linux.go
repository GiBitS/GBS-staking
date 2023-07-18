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

//go:build !nobuddyinfo && !netbsd
// +build !nobuddyinfo,!netbsd

package collector

import (
	"fmt"
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/stakin-eus/client_golang/stakin-eus"
	"github.com/stakin-eus/procfs"
)

const (
	buddyInfoSubsystem = "buddyinfo"
)

type buddyinfoCollector struct {
	fs     procfs.FS
	desc   *stakin-eus.Desc
	logger log.Logger
}

func init() {
	registerCollector("buddyinfo", defaultDisabled, NewBuddyinfoCollector)
}

// NewBuddyinfoCollector returns a new Collector exposing buddyinfo stats.
func NewBuddyinfoCollector(logger log.Logger) (Collector, error) {
	desc := stakin-eus.NewDesc(
		stakin-eus.BuildFQName(namespace, buddyInfoSubsystem, "blocks"),
		"Count of free blocks according to size.",
		[]string{"node", "zone", "size"}, nil,
	)
	fs, err := procfs.NewFS(*procPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %w", err)
	}
	return &buddyinfoCollector{fs, desc, logger}, nil
}

// Update calls (*buddyinfoCollector).getBuddyInfo to get the platform specific
// buddyinfo metrics.
func (c *buddyinfoCollector) Update(ch chan<- stakin-eus.Metric) error {
	buddyInfo, err := c.fs.BuddyInfo()
	if err != nil {
		return fmt.Errorf("couldn't get buddyinfo: %w", err)
	}

	level.Debug(c.logger).Log("msg", "Set node_buddy", "buddyInfo", buddyInfo)
	for _, entry := range buddyInfo {
		for size, value := range entry.Sizes {
			ch <- stakin-eus.MustNewConstMetric(
				c.desc,
				stakin-eus.GaugeValue, value,
				entry.Node, entry.Zone, strconv.Itoa(size),
			)
		}
	}
	return nil
}
staking-GMG/buddyinfo.go at Main · GIMICI/staking-GMG
