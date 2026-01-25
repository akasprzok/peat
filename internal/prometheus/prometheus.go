package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/prometheus/prometheus/promql/parser"
)

type prometheusClient struct {
	v1api v1.API
}

type Client interface {
	Query(query string, timeout time.Duration) (v1.Warnings, model.Vector, error)
	QueryRange(query string, start, end time.Time, step time.Duration, timeout time.Duration) (model.Matrix, v1.Warnings, error)
	Series(query string, start, end time.Time, limit uint64, timeout time.Duration) ([]model.LabelSet, v1.Warnings, error)
	LabelNames(start, end time.Time, timeout time.Duration) ([]string, v1.Warnings, error)
	LabelValues(labelName string, start, end time.Time, timeout time.Duration) ([]string, v1.Warnings, error)
}

func NewClient(url string) (Client, error) {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		return nil, fmt.Errorf("creating prometheus client: %w", err)
	}
	v1api := v1.NewAPI(client)
	return &prometheusClient{v1api: v1api}, nil
}

func (c *prometheusClient) Query(query string, timeout time.Duration) (v1.Warnings, model.Vector, error) {
	var vector model.Vector
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result, warnings, err := c.v1api.Query(ctx, query, time.Now(), v1.WithTimeout(timeout))
	if err != nil {
		return warnings, vector, err
	}

	switch result.Type() {
	case model.ValVector:
		v := result.(model.Vector)
		return warnings, v, nil
	case model.ValNone, model.ValScalar, model.ValMatrix, model.ValString:
		return warnings, vector, fmt.Errorf("unexpected result type: %s", result.Type())
	default:
		return warnings, vector, fmt.Errorf("unknown result type: %s", result.Type())
	}
}

func (c *prometheusClient) QueryRange(query string, start, end time.Time, step time.Duration, timeout time.Duration) (model.Matrix, v1.Warnings, error) {
	var matrix model.Matrix
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result, warnings, err := c.v1api.QueryRange(ctx, query, v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}, v1.WithTimeout(timeout))
	if err != nil {
		return matrix, warnings, err
	}

	switch result.Type() {
	case model.ValMatrix:
		m := result.(model.Matrix)
		return m, warnings, nil
	case model.ValNone, model.ValScalar, model.ValVector, model.ValString:
		return matrix, warnings, fmt.Errorf("unexpected result type: %s", result.Type())
	default:
		return matrix, warnings, fmt.Errorf("unknown result type: %s", result.Type())
	}
}

func (c *prometheusClient) Series(query string, start, end time.Time, limit uint64, timeout time.Duration) ([]model.LabelSet, v1.Warnings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	series, warnings, err := c.v1api.Series(ctx, []string{query}, start, end, v1.WithTimeout(timeout), v1.WithLimit(limit))
	if err != nil {
		return series, warnings, err
	}
	return series, warnings, nil
}

func (c *prometheusClient) LabelNames(start, end time.Time, timeout time.Duration) ([]string, v1.Warnings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	labels, warnings, err := c.v1api.LabelNames(ctx, []string{}, start, end, v1.WithTimeout(timeout))
	if err != nil {
		return nil, warnings, err
	}
	return labels, warnings, nil
}

func (c *prometheusClient) LabelValues(labelName string, start, end time.Time, timeout time.Duration) ([]string, v1.Warnings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	values, warnings, err := c.v1api.LabelValues(ctx, labelName, []string{}, start, end, v1.WithTimeout(timeout))
	if err != nil {
		return nil, warnings, err
	}
	// Convert model.LabelValue slice to string slice
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = string(v)
	}
	return result, warnings, nil
}

func FormatQuery(query string) string {
	ast, err := parser.ParseExpr(query)
	if err != nil {
		return query
	}
	return ast.Pretty(0)
}
