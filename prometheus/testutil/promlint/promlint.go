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

// Package promlint provides a linter for Prometheus metrics.
package promlint

import (
	"errors"
	"io"
	"sort"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/prometheus/client_golang/prometheus/testutil/promlint/validations"
)

// A Linter is a Prometheus metrics linter.  It identifies issues with metric
// names, types, and metadata, and reports them to the caller.
type Linter struct {
	// The linter will read metrics in the Prometheus text format from r and
	// then lint it, _and_ it will lint the metrics provided directly as
	// MetricFamily proto messages in mfs. Note, however, that the current
	// constructor functions New and NewWithMetricFamilies only ever set one
	// of them.
	r   io.Reader
	mfs []*dto.MetricFamily

	customValidations []validations.Validation
}

// New creates a new Linter that reads an input stream of Prometheus metrics in
// the Prometheus text exposition format.
func New(r io.Reader) *Linter {
	return &Linter{
		r: r,
	}
}

// NewWithMetricFamilies creates a new Linter that reads from a slice of
// MetricFamily protobuf messages.
func NewWithMetricFamilies(mfs []*dto.MetricFamily) *Linter {
	return &Linter{
		mfs: mfs,
	}
}

// AddCustomValidations adds custom validations to the linter.
func (l *Linter) AddCustomValidations(vs ...validations.Validation) {
	if l.customValidations == nil {
		l.customValidations = make([]validations.Validation, 0, len(vs))
	}
	l.customValidations = append(l.customValidations, vs...)
}

// Lint performs a linting pass, returning a slice of Problems indicating any
// issues found in the metrics stream. The slice is sorted by metric name
// and issue description.
func (l *Linter) Lint() ([]validations.Problem, error) {
	var problems []validations.Problem

	if l.r != nil {
		d := expfmt.NewDecoder(l.r, expfmt.FmtText)

		mf := &dto.MetricFamily{}
		for {
			if err := d.Decode(mf); err != nil {
				if errors.Is(err, io.EOF) {
					break
				}

				return nil, err
			}

			problems = append(problems, l.lint(mf)...)
		}
	}
	for _, mf := range l.mfs {
		problems = append(problems, l.lint(mf)...)
	}

	// Ensure deterministic output.
	sort.SliceStable(problems, func(i, j int) bool {
		if problems[i].Metric == problems[j].Metric {
			return problems[i].Text < problems[j].Text
		}
		return problems[i].Metric < problems[j].Metric
	})

	return problems, nil
}

// lint is the entry point for linting a single metric.
func (l *Linter) lint(mf *dto.MetricFamily) []validations.Problem {
	var problems []validations.Problem

	for _, fn := range validations.DefaultValidations {
		problems = append(problems, fn(mf)...)
	}

	if l.customValidations != nil {
		for _, fn := range l.customValidations {
			problems = append(problems, fn(mf)...)
		}
	}

	// TODO(mdlayher): lint rules for specific metrics types.
	return problems
}
