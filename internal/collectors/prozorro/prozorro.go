package prozorro

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alexg/my-estate/internal/domain"
	"github.com/alexg/my-estate/internal/httpx"
)

const Name = "prozorro"

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
		URL:  "https://procedure.prozorro.sale/api/search",
	}
}

func (c *Collector) FetchLots(ctx context.Context) ([]domain.AuctionLot, error) {
	body, status, err := c.HTTP.Get(ctx, c.URL+"?limit=20")
	if err != nil {
		return nil, err
	}
	if status >= 400 {
		return nil, fmt.Errorf("prozorro status %d", status)
	}
	return ParseLots(body)
}

func ParseLots(body []byte) ([]domain.AuctionLot, error) {
	var generic any
	if err := json.Unmarshal(body, &generic); err != nil {
		return nil, err
	}
	items := flattenItems(generic)
	out := make([]domain.AuctionLot, 0, len(items))
	for _, it := range items {
		id := stringField(it, "_id", "id", "auctionId", "procedureId")
		if id == "" {
			continue
		}
		raw, _ := json.Marshal(it)
		lot := domain.AuctionLot{
			Source:       Name,
			SourceLotID:  id,
			SourceURL:    "https://prozorro.sale/" + id,
			Title:        stringField(it, "title", "title.uk_UA", "item.title"),
			Description:  stringField(it, "description", "description.uk_UA"),
			AuctionStage: stringField(it, "status", "status_id"),
			Organizer:    stringField(it, "sellingEntity.name", "owner"),
			RawPayload:   raw,
		}
		if p, ok := numField(it, "value.amount", "valueAmount", "startingPrice"); ok {
			lot.StartPrice = &p
		}
		if g, ok := numField(it, "guarantee.amount", "guaranteeAmount"); ok {
			lot.GuaranteeDeposit = &g
		}
		out = append(out, lot)
	}
	return out, nil
}

func flattenItems(v any) []map[string]any {
	switch t := v.(type) {
	case []any:
		out := make([]map[string]any, 0, len(t))
		for _, x := range t {
			if m, ok := x.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	case map[string]any:
		for _, key := range []string{"data", "items", "results"} {
			if inner, ok := t[key]; ok {
				return flattenItems(inner)
			}
		}
		return []map[string]any{t}
	default:
		return nil
	}
}

func stringField(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := lookup(m, k); ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func numField(m map[string]any, keys ...string) (float64, bool) {
	for _, k := range keys {
		if v, ok := lookup(m, k); ok {
			switch t := v.(type) {
			case float64:
				return t, true
			}
		}
	}
	return 0, false
}

func lookup(m map[string]any, path string) (any, bool) {
	cur := any(m)
	start := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '.' {
			key := path[start:i]
			mm, ok := cur.(map[string]any)
			if !ok {
				return nil, false
			}
			cur, ok = mm[key]
			if !ok {
				return nil, false
			}
			start = i + 1
		}
	}
	return cur, true
}
