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
}

func NewClient(url string) Client {
	client, err := api.NewClient(api.Config{
		Address: "http://vmselect-multi-cluster:8481/select/13/prometheus",
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
		return warnings, vector, fmt.Errorf("Unknown result type: %s\n", result.Type())
	}
}
