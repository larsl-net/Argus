// Copyright [2023] [Argus]
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

package deployedver

import (
	"encoding/json"
	"fmt"

	opt "github.com/release-argus/Argus/service/options"
	svcstatus "github.com/release-argus/Argus/service/status"
	"github.com/release-argus/Argus/util"
)

// applyOverrides to the Lookup and return that new Lookup.
func (l *Lookup) applyOverrides(
	allowInvalidCerts *string,
	basicAuth *string,
	body *string,
	headers *string,
	json *string,
	method *string,
	regex *string,
	regexTemplate *string,
	semanticVersioning *string,
	url *string,
	serviceID *string,
	logFrom *util.LogFrom,
) (*Lookup, error) {
	// Use the provided overrides, or the defaults.
	// allow_invalid_certs
	useAllowInvalidCerts := l.AllowInvalidCerts
	if allowInvalidCerts != nil {
		useAllowInvalidCerts = util.StringToBoolPtr(*allowInvalidCerts)
	}
	// basic_auth
	useBasicAuth := basicAuthFromString(
		basicAuth,
		l.BasicAuth,
		logFrom)
	// body
	useBody := util.FirstNonNilPtr(body, l.Body)
	// headers
	useHeaders := headersFromString(
		headers,
		&l.Headers,
		logFrom)
	// json
	useJSON := util.PtrValueOrValue(json, l.JSON)
	// method
	useMethod := util.PtrValueOrValue(method, l.Method)
	// regex
	useRegex := util.PtrValueOrValue(regex, l.Regex)
	useRegexTemplate := util.PtrValueOrValue(regexTemplate, util.DefaultIfNil(l.RegexTemplate))
	// semantic_versioning
	var useSemanticVersioning *bool
	if semanticVersioning != nil {
		useSemanticVersioning = util.StringToBoolPtr(*semanticVersioning)
	}
	// url
	useURL := util.PtrValueOrValue(url, l.URL)

	// options
	options := opt.New(
		nil, "",
		useSemanticVersioning,
		l.Options.Defaults,
		l.Options.HardDefaults)

	// Create a new lookup with the overrides.
	lookup := New(
		useAllowInvalidCerts,
		useBasicAuth,
		useBody,
		useHeaders,
		useJSON,
		useMethod,
		options,
		useRegex,
		&useRegexTemplate,
		&svcstatus.Status{},
		useURL,
		l.Defaults,
		l.HardDefaults)
	if err := lookup.CheckValues(""); err != nil {
		jLog.Error(err, logFrom, true)
		return nil, fmt.Errorf("values failed validity check:\n%w", err)
	}
	lookup.Status.Init(
		0, 0, 0,
		serviceID,
		nil)
	return lookup, nil
}

// Refresh (query) the Lookup with the provided overrides,
// returning the version found with this query
func (l *Lookup) Refresh(
	allowInvalidCerts *string,
	basicAuth *string,
	body *string,
	headers *string,
	json *string,
	method *string,
	regex *string,
	regexTemplate *string,
	semanticVersioning *string,
	url *string,
) (version string, announceUpdate bool, err error) {
	serviceID := *l.Status.ServiceID
	logFrom := &util.LogFrom{Primary: "deployed_version/refresh", Secondary: serviceID}

	var lookup *Lookup
	lookup, err = l.applyOverrides(
		allowInvalidCerts,
		basicAuth,
		body,
		headers,
		json,
		method,
		regex,
		regexTemplate,
		semanticVersioning,
		url,
		&serviceID,
		logFrom)
	if err != nil {
		return
	}

	// Log the lookup being used if debug.
	if jLog.IsLevel("DEBUG") {
		jLog.Debug(
			fmt.Sprintf("Refreshing with:\n%v", lookup),
			logFrom, true)
	}

	// Whether overrides were provided or not, we can update the status if not.
	overrides := headers != nil ||
		l.Options.GetSemanticVersioning() != lookup.Options.GetSemanticVersioning() ||
		url != nil ||
		json != nil ||
		regex != nil ||
		regexTemplate != nil

	// Query the lookup.
	version, err = lookup.Query(!overrides, logFrom)
	if err != nil {
		return
	}

	// Update the deployed version if it has changed.
	if version != l.Status.DeployedVersion() &&
		// and no overrides that may change a successful query were provided
		!overrides {
		announceUpdate = true
		l.Status.SetDeployedVersion(version, true)
		l.Status.AnnounceUpdate()
	}

	return
}

func basicAuthFromString(jsonStr *string, previous *BasicAuth, logFrom *util.LogFrom) *BasicAuth {
	// jsonStr == nil when it hasn't been changed, so return the previous
	if jsonStr == nil {
		return previous
	}

	basicAuth := &BasicAuth{}
	err := json.Unmarshal([]byte(*jsonStr), &basicAuth)
	// Ignore the JSON if it failed to unmarshal
	if err != nil {
		jLog.Error(fmt.Sprintf("Failed converting JSON - %q\n%s", *jsonStr, util.ErrorToString(err)),
			logFrom, true)
		return previous
	}
	keys := util.GetKeysFromJSON(*jsonStr)

	// Had no previous, so can't use it as defaults
	if previous == nil {
		return basicAuth
	}

	// defaults
	if !util.Contains(keys, "username") {
		basicAuth.Username = previous.Username
	}
	if !util.Contains(keys, "password") {
		basicAuth.Password = previous.Password
	}

	return basicAuth
}

func headersFromString(jsonStr *string, previous *[]Header, logFrom *util.LogFrom) *[]Header {
	// jsonStr == nil when it hasn't been changed, so return the previous
	if jsonStr == nil {
		return previous
	}

	var headers []Header
	err := json.Unmarshal([]byte(*jsonStr), &headers)
	// Ignore the JSON if it failed to unmarshal
	if err != nil {
		jLog.Error(fmt.Sprintf("Failed converting JSON - %q\n%s", *jsonStr, util.ErrorToString(err)),
			logFrom, true)
		return previous
	}

	return &headers
}
