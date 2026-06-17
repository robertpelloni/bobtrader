package influx

import (
	"context"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type Client struct {
	client influxdb2.Client
	writeAPI api.WriteAPIBlocking
}

func NewClient(url, token, org, bucket string) *Client {
	client := influxdb2.NewClient(url, token)
	writeAPI := client.WriteAPIBlocking(org, bucket)
	return &Client{
		client:   client,
		writeAPI: writeAPI,
	}
}

func (c *Client) Close() {
	c.client.Close()
}

func (c *Client) WriteMetric(ctx context.Context, measurement string, tags map[string]string, fields map[string]interface{}) error {
	p := influxdb2.NewPoint(measurement, tags, fields, time.Now())
	return c.writeAPI.WritePoint(ctx, p)
}

func (c *Client) WriteValuation(ctx context.Context, accountID string, portfolioValue, realizedPnL, unrealizedPnL float64) error {
	tags := map[string]string{"account_id": accountID}
	fields := map[string]interface{}{
		"portfolio_value": portfolioValue,
		"realized_pnl":    realizedPnL,
		"unrealized_pnl":  unrealizedPnL,
	}
	return c.WriteMetric(ctx, "portfolio_valuation", tags, fields)
}
