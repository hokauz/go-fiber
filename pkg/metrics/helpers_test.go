package metrics

import (
	dto "github.com/prometheus/client_model/go"
)

func findCounterValue(mfs []*dto.MetricFamily, name string, labels map[string]string) float64 {
	m := findMetric(mfs, name, labels)
	if m == nil || m.Counter == nil {
		return 0
	}
	return m.GetCounter().GetValue()
}

func findHistogramCount(mfs []*dto.MetricFamily, name string, labels map[string]string) uint64 {
	m := findMetric(mfs, name, labels)
	if m == nil || m.Histogram == nil {
		return 0
	}
	return m.GetHistogram().GetSampleCount()
}

func findGaugeValue(mfs []*dto.MetricFamily, name string) float64 {
	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		for _, metric := range mf.GetMetric() {
			if len(metric.GetLabel()) == 0 && metric.Gauge != nil {
				return metric.GetGauge().GetValue()
			}
		}
	}
	return 0
}

func findGaugeValueWithLabels(mfs []*dto.MetricFamily, name string, labels map[string]string) float64 {
	m := findMetric(mfs, name, labels)
	if m == nil || m.Gauge == nil {
		return 0
	}
	return m.GetGauge().GetValue()
}

func findMetric(mfs []*dto.MetricFamily, name string, labels map[string]string) *dto.Metric {
	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		for _, metric := range mf.GetMetric() {
			if labelsMatch(metric.GetLabel(), labels) {
				return metric
			}
		}
	}
	return nil
}

func labelsMatch(got []*dto.LabelPair, expected map[string]string) bool {
	if len(got) != len(expected) {
		return false
	}

	lookup := make(map[string]string, len(got))
	for _, label := range got {
		lookup[label.GetName()] = label.GetValue()
	}
	for key, want := range expected {
		if lookup[key] != want {
			return false
		}
	}
	return true
}
