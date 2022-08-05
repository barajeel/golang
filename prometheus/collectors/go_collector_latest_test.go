// Copyright 2021 The Prometheus Authors
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

//go:build go1.17
// +build go1.17

package collectors

import (
	"encoding/json"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var baseMetrics = []string{
	"go_gc_duration_seconds",
	"go_goroutines",
	"go_info",
	"go_memstats_last_gc_time_seconds",
	"go_threads",
}

func TestGoCollectorMarshalling(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewGoCollector(
		WithGoCollectorRuntimeMetrics(GoRuntimeMetricsRule{
			Matcher: regexp.MustCompile("/.*"),
		}),
	))
	result, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}

	if _, err := json.Marshal(result); err != nil {
		t.Errorf("json marshalling shoud not fail, %v", err)
	}
}

func TestWithGoCollectorMemStatsMetricsDisabled(t *testing.T) {
	reg := prometheus.NewRegistry()
	reg.MustRegister(NewGoCollector(
		WithGoCollectorMemStatsMetricsDisabled(),
	))
	result, err := reg.Gather()
	if err != nil {
		t.Fatal(err)
	}

	got := []string{}
	for _, r := range result {
		got = append(got, r.GetName())
	}

	if !reflect.DeepEqual(got, baseMetrics) {
		t.Errorf("got %v, want %v", got, baseMetrics)
	}
}

func TestGoCollectorAllowList(t *testing.T) {
	for _, test := range []struct {
		name     string
		rules    []GoRuntimeMetricsRule
		expected []string
	}{
		{
			name:     "Without any rules",
			rules:    nil,
			expected: baseMetrics,
		},
		{
			name: "allow all",
			rules: []GoRuntimeMetricsRule{{
				Matcher: regexp.MustCompile("/.*"),
			}},
			expected: withBaseMetrics([]string{
				"go_gc_cycles_automatic_gc_cycles_total",
				"go_gc_cycles_forced_gc_cycles_total",
				"go_gc_cycles_total_gc_cycles_total",
				"go_gc_heap_allocs_by_size_bytes",
				"go_gc_heap_allocs_bytes_total",
				"go_gc_heap_allocs_objects_total",
				"go_gc_heap_frees_by_size_bytes",
				"go_gc_heap_frees_bytes_total",
				"go_gc_heap_frees_objects_total",
				"go_gc_heap_goal_bytes",
				"go_gc_heap_objects_objects",
				"go_gc_heap_tiny_allocs_objects_total",
				"go_gc_pauses_seconds",
				"go_memory_classes_heap_free_bytes",
				"go_memory_classes_heap_objects_bytes",
				"go_memory_classes_heap_released_bytes",
				"go_memory_classes_heap_stacks_bytes",
				"go_memory_classes_heap_unused_bytes",
				"go_memory_classes_metadata_mcache_free_bytes",
				"go_memory_classes_metadata_mcache_inuse_bytes",
				"go_memory_classes_metadata_mspan_free_bytes",
				"go_memory_classes_metadata_mspan_inuse_bytes",
				"go_memory_classes_metadata_other_bytes",
				"go_memory_classes_os_stacks_bytes",
				"go_memory_classes_other_bytes",
				"go_memory_classes_profiling_buckets_bytes",
				"go_memory_classes_total_bytes",
				"go_sched_goroutines_goroutines",
				"go_sched_latencies_seconds",
			}),
		},
		{
			name:  "allow GC",
			rules: []GoRuntimeMetricsRule{MetricsGC},
			expected: withBaseMetrics([]string{
				"go_gc_cycles_automatic_gc_cycles_total",
				"go_gc_cycles_forced_gc_cycles_total",
				"go_gc_cycles_total_gc_cycles_total",
				"go_gc_heap_allocs_by_size_bytes",
				"go_gc_heap_allocs_bytes_total",
				"go_gc_heap_allocs_objects_total",
				"go_gc_heap_frees_by_size_bytes",
				"go_gc_heap_frees_bytes_total",
				"go_gc_heap_frees_objects_total",
				"go_gc_heap_goal_bytes",
				"go_gc_heap_objects_objects",
				"go_gc_heap_tiny_allocs_objects_total",
				"go_gc_pauses_seconds",
			}),
		},
		{
			name:  "allow Memory",
			rules: []GoRuntimeMetricsRule{MetricsMemory},
			expected: withBaseMetrics([]string{
				"go_memory_classes_heap_free_bytes",
				"go_memory_classes_heap_objects_bytes",
				"go_memory_classes_heap_released_bytes",
				"go_memory_classes_heap_stacks_bytes",
				"go_memory_classes_heap_unused_bytes",
				"go_memory_classes_metadata_mcache_free_bytes",
				"go_memory_classes_metadata_mcache_inuse_bytes",
				"go_memory_classes_metadata_mspan_free_bytes",
				"go_memory_classes_metadata_mspan_inuse_bytes",
				"go_memory_classes_metadata_other_bytes",
				"go_memory_classes_os_stacks_bytes",
				"go_memory_classes_other_bytes",
				"go_memory_classes_profiling_buckets_bytes",
				"go_memory_classes_total_bytes",
			}),
		},
		{
			name:  "allow Scheduler",
			rules: []GoRuntimeMetricsRule{MetricsScheduler},
			expected: []string{
				"go_gc_duration_seconds",
				"go_goroutines",
				"go_info",
				"go_memstats_last_gc_time_seconds",
				"go_sched_goroutines_goroutines",
				"go_sched_latencies_seconds",
				"go_threads",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			reg.MustRegister(NewGoCollector(
				WithGoCollectorMemStatsMetricsDisabled(),
				WithGoCollectorRuntimeMetrics(test.rules...),
			))
			result, err := reg.Gather()
			if err != nil {
				t.Fatal(err)
			}

			got := []string{}
			for _, r := range result {
				got = append(got, r.GetName())
			}

			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("got %v, want %v", got, test.expected)
			}
		})
	}
}

func withBaseMetrics(metricNames []string) []string {
	metricNames = append(metricNames, baseMetrics...)
	sort.Strings(metricNames)
	return metricNames
}

func TestGoCollectorDenyList(t *testing.T) {
	for _, test := range []struct {
		name     string
		matchers []*regexp.Regexp
		expected []string
	}{
		{
			name:     "Without any matchers",
			matchers: nil,
			expected: baseMetrics,
		},
		{
			name:     "deny all",
			matchers: []*regexp.Regexp{regexp.MustCompile("/.*")},
			expected: baseMetrics,
		},
		{
			name: "deny gc and scheduler latency",
			matchers: []*regexp.Regexp{
				regexp.MustCompile("^/gc/.*"),
				regexp.MustCompile("^/sched/latencies:.*"),
			},
			expected: baseMetrics,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			reg := prometheus.NewRegistry()
			reg.MustRegister(NewGoCollector(
				WithGoCollectorMemStatsMetricsDisabled(),
				WithoutGoCollectorRuntimeMetrics(test.matchers...),
			))
			result, err := reg.Gather()
			if err != nil {
				t.Fatal(err)
			}

			got := []string{}
			for _, r := range result {
				got = append(got, r.GetName())
			}

			if !reflect.DeepEqual(got, test.expected) {
				t.Errorf("got %v, want %v", got, test.expected)
			}
		})
	}
}

func ExampleGoCollector() {
	reg := prometheus.NewRegistry()

	// Register the GoCollector with the default options. Only the base metrics will be enabled.
	reg.MustRegister(NewGoCollector())

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ExampleGoCollector_WithAdvancedGoMetrics() {
	reg := prometheus.NewRegistry()

	// Enable Go metrics with pre-defined rules. Or your custom rules.
	reg.MustRegister(
		NewGoCollector(
			WithGoCollectorMemStatsMetricsDisabled(),
			WithGoCollectorRuntimeMetrics(
				MetricsScheduler,
				MetricsMemory,
				GoRuntimeMetricsRule{
					Matcher: regexp.MustCompile("^/mycustomrule.*"),
				},
			),
			WithoutGoCollectorRuntimeMetrics(regexp.MustCompile("^/gc/.*")),
		))

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func ExampleGoCollector_DefaultRegister() {
	// Unregister the default GoCollector.
	prometheus.Unregister(NewGoCollector())

	// Register the default GoCollector with a custom config.
	prometheus.MustRegister(NewGoCollector(WithGoCollectorRuntimeMetrics(
		MetricsScheduler,
		MetricsGC,
		GoRuntimeMetricsRule{
			Matcher: regexp.MustCompile("^/mycustomrule.*"),
		},
	),
	))

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}
