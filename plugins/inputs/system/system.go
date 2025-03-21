//go:generate ../../../tools/readme_config_includer/generator
package system

import (
	_ "embed"
	"os"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type SystemStats struct {
	Log telegraf.Logger
}

func (*SystemStats) SampleConfig() string {
	return sampleConfig
}

func (s *SystemStats) Gather(acc telegraf.Accumulator) error {
	loadavg, err := load.Avg()
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		return err
	}

	numCPUs, err := cpu.Counts(true)
	if err != nil {
		return err
	}

	fields := map[string]interface{}{
		"load1":  loadavg.Load1,
		"load5":  loadavg.Load5,
		"load15": loadavg.Load15,
		"n_cpus": numCPUs,
	}

	users, err := host.Users()
	if err == nil {
		fields["n_users"] = len(users)
		fields["n_unique_users"] = findUniqueUsers(users)
	} else if os.IsNotExist(err) {
		s.Log.Debugf("Reading users: %s", err.Error())
	} else if os.IsPermission(err) {
		s.Log.Debug(err.Error())
	}

	now := time.Now()
	acc.AddGauge("system", fields, nil, now)

	uptime, err := host.Uptime()
	if err != nil {
		return err
	}

	acc.AddCounter("system", map[string]interface{}{
		"uptime": uptime,
	}, nil, now)

	return nil
}

func findUniqueUsers(userStats []host.UserStat) int {
	uniqueUsers := make(map[string]bool)
	for _, userstat := range userStats {
		if _, ok := uniqueUsers[userstat.User]; !ok {
			uniqueUsers[userstat.User] = true
		}
	}

	return len(uniqueUsers)
}

func init() {
	inputs.Add("system", func() telegraf.Input {
		return &SystemStats{}
	})
}
