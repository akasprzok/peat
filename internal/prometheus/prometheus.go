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
	QueryRange(query string, start, end time.Time, step time.Duration) (v1.Warnings, model.Matrix, error)
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

func (c *prometheusClient) QueryRange(query string, start, end time.Time, step time.Duration) (v1.Warnings, model.Matrix, error) {
	var matrix model.Matrix
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := c.v1api.QueryRange(ctx, query, v1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}, v1.WithTimeout(5*time.Second))
	if err != nil {
		return warnings, matrix, err
	}

	switch result.Type() {
	case model.ValMatrix:
		matrix := result.(model.Matrix)
		return warnings, matrix, nil
	default:
		return warnings, matrix, fmt.Errorf("unknown result type: %s", result.Type())
	}
}
