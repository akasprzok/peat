package prometheus

import (
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// MockClient is a mock implementation of the Client interface for testing.
type MockClient struct {
	QueryFunc      func(query string, timeout time.Duration) (v1.Warnings, model.Vector, error)
	QueryRangeFunc func(query string, start, end time.Time, step time.Duration, timeout time.Duration) (model.Matrix, v1.Warnings, error)
	SeriesFunc     func(query string, start, end time.Time, limit uint64, timeout time.Duration) ([]model.LabelSet, v1.Warnings, error)
}

func (m *MockClient) Query(query string, timeout time.Duration) (v1.Warnings, model.Vector, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(query, timeout)
	}
	return nil, nil, nil
}

func (m *MockClient) QueryRange(query string, start, end time.Time, step time.Duration, timeout time.Duration) (model.Matrix, v1.Warnings, error) {
	if m.QueryRangeFunc != nil {
		return m.QueryRangeFunc(query, start, end, step, timeout)
	}
	return nil, nil, nil
}

func (m *MockClient) Series(query string, start, end time.Time, limit uint64, timeout time.Duration) ([]model.LabelSet, v1.Warnings, error) {
	if m.SeriesFunc != nil {
		return m.SeriesFunc(query, start, end, limit, timeout)
	}
	return nil, nil, nil
}
