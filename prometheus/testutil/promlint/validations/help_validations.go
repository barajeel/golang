// Copyright 2020 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validations

import dto "github.com/prometheus/client_model/go"

// lintHelp detects issues related to the help text for a metric.
func lintHelp(mf *dto.MetricFamily) []Problem {
	var problems []Problem

	// Expect all metrics to have help text available.
	if mf.Help == nil {
		problems = append(problems, newProblem(mf, "no help text"))
	}

	return problems
}
