// Copyright [2022] [Hymenaios]
//
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

package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/hymenaios-io/Hymenaios/utils"
	"github.com/hymenaios-io/Hymenaios/web/metrics"
)

// Track will call Track on all Services in this Slice.
func (s *Slice) Track(ordering *[]string) {
	for _, key := range *ordering {
		msg := fmt.Sprintf("Tracking %s at %s every %s", *(*s)[key].ID, (*s)[key].GetServiceURL(true), (*s)[key].GetInterval())
		jLog.Verbose(msg, utils.LogFrom{Primary: *(*s)[key].ID}, true)

		// Track this Service in a infinite loop goroutine.
		go (*s)[key].Track()

		// Space out the tracking of each Service.
		time.Sleep(time.Duration(2) * time.Second)
	}
}

// Track will track the Service data and then send Slack
// messages (Service.Slack) as well as WebHooks (Service.WebHook)
// when a new release is spottes. It sleeps for Service.Interval
// between each check.
func (s *Service) Track() {
	serviceInfo := s.GetServiceInfo()
	// Track forever.
	for {
		// If new release found by this query.
		newVersion, err := s.Query()

		if newVersion {
			// Get updated serviceInfo
			serviceInfo = s.GetServiceInfo()

			// Send the Gotify Message(s).
			//nolint:errcheck
			go s.Gotify.Send("", "", &serviceInfo)

			// Send the Slack Message(s).
			//nolint:errcheck
			go s.Slack.Send("", &serviceInfo)

			// WebHook(s)
			go s.HandleWebHooks(false)
		}

		// If it failed
		if err != nil {
			if strings.HasPrefix(err.Error(), "regex ") {
				metrics.SetPrometheusGaugeWithID(metrics.QueryLiveness, *s.ID, 2)
			} else if strings.HasSuffix(err.Error(), "semantic version") {
				metrics.SetPrometheusGaugeWithID(metrics.QueryLiveness, *s.ID, 3)
			} else if strings.HasPrefix(err.Error(), "queried version") {
				metrics.SetPrometheusGaugeWithID(metrics.QueryLiveness, *s.ID, 4)
			} else {
				metrics.IncreasePrometheusCounterWithIDAndResult(metrics.QueryMetric, *s.ID, "FAIL")
				metrics.SetPrometheusGaugeWithID(metrics.QueryLiveness, *s.ID, 0)
			}
		} else {
			metrics.IncreasePrometheusCounterWithIDAndResult(metrics.QueryMetric, *s.ID, "SUCCESS")
			metrics.SetPrometheusGaugeWithID(metrics.QueryLiveness, *s.ID, 1)
		}
		// Sleep interval between checks.
		time.Sleep(s.GetIntervalDuration())
	}
}