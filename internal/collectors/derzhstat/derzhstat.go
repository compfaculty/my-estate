package derzhstat

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alexg/my-estate/internal/domain"
	"github.com/alexg/my-estate/internal/httpx"
)

const Name = "derzhstat"

type Collector struct {
	HTTP *httpx.Client
	URL  string
}

func New(httpClient *httpx.Client) *Collector {
	if httpClient == nil {
		httpClient = httpx.New()
	}
	return &Collector{
		HTTP: httpClient,
		URL:  "https://stat.gov.ua/en/datasets/changes-housing-market-prices",
	}
}

func (c *Collector) Fetch(ctx context.Context) (domain.MarketStatistic, error) {
	body, status, err := c.HTTP.Get(ctx, c.URL)
	if err != nil {
		return domain.MarketStatistic{}, err
	}
	if status >= 400 {
		return domain.MarketStatistic{}, fmt.Errorf("derzhstat status %d", status)
	}
	period := time.Now().UTC().Format("2006-Q")
	q := (int(time.Now().UTC().Month())-1)/3 + 1
	period = fmt.Sprintf("%d-Q%d", time.Now().UTC().Year(), q)
	payload, _ := json.Marshal(map[string]any{
		"http_status": status,
		"bytes":       len(body),
		"url":         c.URL,
		"note":        "dataset landing captured; SDMX series parsed when JSON array present",
	})
	stat := domain.MarketStatistic{
		Source:  Name,
		Region:  "UA",
		Market:  "housing",
		Period:  period,
		Payload: payload,
	}
	var arr []map[string]any
	if json.Unmarshal(body, &arr) == nil && len(arr) > 0 {
		if v, ok := asFloat(arr[0]["value"]); ok {
			stat.IndexValue = &v
		}
	}
	return stat, nil
}

func asFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}
