// Copyright 2015 The Prometheus Authors
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

// A simple example exposing fictional RPC latencies with different types of
// random distributions (uniform, normal, and exponential) as Prometheus
// metrics.

package ranges

import (
	"time"

	"github.com/oscarzhao/client_golang/api/prometheus" // need to update when pull request
	prometheusModel "github.com/prometheus/common/model"
	"golang.org/x/net/context"
)

// QueryRange executes a range vector query to Prometheus Server
func QueryRange(addr, token, queryString string, start, end time.Time) (prometheusModel.Value, error) {
	// create configuration
	c := prometheus.Config{
		Address: addr,
		Token:   token,
	}
	pClient, err := prometheus.New(c)
	if err != nil {
		return nil, err
	}

	// create a query api client
	queryAPI := prometheus.NewQueryAPI(pClient)

	rawResults, err := queryAPI.QueryRange(context.WithValue(context.Background(), prometheus.ContextBearerTokenKey, token), queryString, prometheus.Range{
		Start: start,
		End:   end,
		Step:  time.Minute,
	})

	if err != nil {
		return nil, err
	}

	return rawResults, nil
}
