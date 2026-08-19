package nbu

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alexg/my-estate/internal/domain"
	"github.com/alexg/my-estate/internal/httpx"
)

const Name = "nbu"

type Client struct {
	HTTP *httpx.Client
	Now  func() time.Time
}

func New(httpClient *httpx.Client) *Client {
	if httpClient == nil {
		httpClient = httpx.New()
	}
	return &Client{HTTP: httpClient, Now: time.Now}
}

type rateJSON struct {
	CC     string  `json:"cc"`
	Rate   float64 `json:"rate"`
	ExDate string  `json:"exchangedate"`
}

func (c *Client) FetchDate(ctx context.Context, day time.Time) ([]domain.FxRate, error) {
	day = day.UTC()
	url := fmt.Sprintf("https://bank.gov.ua/NBUStatService/v1/statdirectory/exchange?date=%s&json", day.Format("20060102"))
	body, status, err := c.HTTP.Get(ctx, url)
	if err != nil {
		return nil, err
	}
	if status != 200 {
		return nil, fmt.Errorf("nbu status %d", status)
	}
	return ParseRates(body, day)
}

func ParseRates(body []byte, fallback time.Time) ([]domain.FxRate, error) {
	var rows []rateJSON
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, err
	}
	out := make([]domain.FxRate, 0, 2)
	for _, r := range rows {
		ccy := strings.ToUpper(r.CC)
		if ccy != "USD" && ccy != "EUR" {
			continue
		}
		d := fallback
		if t, err := time.Parse("02.01.2006", r.ExDate); err == nil {
			d = t
		}
		out = append(out, domain.FxRate{
			RateDate: d,
			CCY:      ccy,
			RateUAH:  r.Rate,
			Source:   Name,
		})
	}
	return out, nil
}

func Convert(original float64, ccy string, usdToUAH float64) (uah, usd float64) {
	switch strings.ToUpper(ccy) {
	case "UAH", "грн", "HRN":
		uah = original
		if usdToUAH > 0 {
			usd = original / usdToUAH
		}
	case "USD", "US$":
		usd = original
		uah = original * usdToUAH
	default:
		uah = original
		if usdToUAH > 0 {
			usd = original / usdToUAH
		}
	}
	return uah, usd
}
