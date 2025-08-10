package prometheus

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type prometheusClient struct {
	v1api v1.API
}

type Client interface {
	Query(query string) (v1.Warnings, model.Vector, error)
	QueryRange(query string, start, end time.Time, step time.Duration) (model.Matrix, v1.Warnings, error)
	Series(query string, start, end time.Time, limit uint64) ([]model.LabelSet, v1.Warnings, error)
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

func (c *prometheusClient) Query(query string) (v1.Warnings, model.Vector, error) {
	var vector model.Vector
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := c.v1api.Query(ctx, query, time.Now(), v1.WithTimeout(5*time.Second))
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

func (c *prometheusClient) QueryRange(query string, start, end time.Time, step time.Duration) (model.Matrix, v1.Warnings, error) {
	var matrix model.Matrix
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	result, warnings, err := c.v1api.QueryRange(ctx, query, v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}, v1.WithTimeout(120*time.Second))
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

func (c *prometheusClient) Series(query string, start, end time.Time, limit uint64) ([]model.LabelSet, v1.Warnings, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	series, warnings, err := c.v1api.Series(ctx, []string{query}, start, end, v1.WithTimeout(60*time.Second), v1.WithLimit(limit))
	if err != nil {
		return series, warnings, err
	}
	return series, warnings, nil
}
