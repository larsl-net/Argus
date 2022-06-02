// Copyright [2022] [Argus]
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

package command

import (
	"fmt"
	"strings"

	"github.com/release-argus/Argus/notifiers/shoutrrr"
	"github.com/release-argus/Argus/utils"
	metrics "github.com/release-argus/Argus/web/metrics"
)

// Init the Command metrics.
func (c *Controller) Init(
	log *utils.JLog,
	serviceID *string,
	command *Slice,
	shoutrrrNotifiers *shoutrrr.Slice,
) {
	jLog = log
	if c == nil {
		return
	}
	parentID := *serviceID

	c.Command = command
	c.Failed = make(Fails, len(*command))

	c.ServiceID = &parentID
	c.initMetrics(parentID)

	// Command fail notifiers
	(*c).Notifiers = Notifiers{
		Shoutrrr: shoutrrrNotifiers,
	}
}

// initMetrics, giving them all a starting value.
func (c *Controller) initMetrics(serviceID string) {
	// ############
	// # Counters #
	// ############
	for i := range *c.Command {
		name := (*c.Command)[i].String()
		metrics.InitPrometheusCounterActions(metrics.CommandMetric, name, serviceID, "", "SUCCESS")
		metrics.InitPrometheusCounterActions(metrics.CommandMetric, name, serviceID, "", "FAIL")
	}

	// ##########
	// # Gauges #
	// ##########
	metrics.SetPrometheusGaugeWithID(metrics.AckWaiting, serviceID, float64(0))
}

// FormattedString will convert Command to a string in the format of '[ "arg0", "arg1" ]'
func (c *Command) FormattedString() string {
	return fmt.Sprintf("[ \"%s\" ]", strings.Join(*c, "\", \""))
}

// String will convert Command to a string in the format of 'arg0 arg1'
func (c *Command) String() string {
	return strings.Join(*c, " ")
}