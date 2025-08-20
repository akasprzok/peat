package prometheus

import (
	"context"
	"fmt"
	"os"
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
}

func NewClient(url string) Client {
	client, err := api.NewClient(api.Config{
		Address: url,
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}
	v1api := v1.NewAPI(client)
	return &prometheusClient{v1api: v1api}
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
		vector := result.(model.Vector)
		return warnings, vector, nil
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
		matrix := result.(model.Matrix)
		return matrix, warnings, nil
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

func FormatQuery(query string) string {
	ast, err := parser.ParseExpr(query)
	if err != nil {
		return query
	}
	return ast.Pretty(0)
}
