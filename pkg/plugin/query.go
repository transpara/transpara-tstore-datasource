package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// QueryModel represents a single query from the Grafana frontend.
type QueryModel struct {
	QueryType string   `json:"queryType"` // "visual" or "raw"
	Lookups   []string `json:"lookups"`
	AggType   string   `json:"aggType"`
	AggInt    string   `json:"aggInt"`
	Tz        string   `json:"tz"`
	RawJson   string   `json:"rawJson"`
}

// dataPoint is one time-value pair from tstore-interface.
type dataPoint struct {
	Ts string      `json:"ts"`
	V  interface{} `json:"v"`
}

// QueryData handles all panel queries.
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()
	for _, q := range req.Queries {
		response.Responses[q.RefID] = d.runQuery(ctx, q)
	}
	return response, nil
}

func (d *Datasource) runQuery(ctx context.Context, q backend.DataQuery) backend.DataResponse {
	var qm QueryModel
	if err := json.Unmarshal(q.JSON, &qm); err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("parsing query: %v", err))
	}
	if qm.Tz == "" {
		qm.Tz = "UTC"
	}

	var reqBody []byte
	var queryParams string

	if qm.QueryType == "raw" {
		reqBody = []byte(qm.RawJson)
		queryParams = fmt.Sprintf("start_time=%s&end_time=%s&tz=%s",
			q.TimeRange.From.UTC().Format(time.RFC3339),
			q.TimeRange.To.UTC().Format(time.RFC3339),
			qm.Tz,
		)
	} else {
		body, err := json.Marshal(qm.Lookups)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, err.Error())
		}
		reqBody = body

		params := fmt.Sprintf("start_time=%s&end_time=%s&tz=%s",
			q.TimeRange.From.UTC().Format(time.RFC3339),
			q.TimeRange.To.UTC().Format(time.RFC3339),
			qm.Tz,
		)
		if qm.AggType != "" && qm.AggType != "raw" {
			params += "&agg_type=" + qm.AggType
		}
		if qm.AggInt != "" {
			params += "&agg_int=" + qm.AggInt
		}
		queryParams = params
	}

	targetURL := d.settings.URL + "/api/v1/read/historical-data?" + queryParams
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewReader(reqBody))
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, err.Error())
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := d.doRequest(ctx, httpReq)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadGateway, err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return backend.ErrDataResponse(backend.StatusBadGateway,
			fmt.Sprintf("tstore returned %d: %s", resp.StatusCode, string(body)))
	}

	var result map[string][]dataPoint
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, fmt.Sprintf("decoding response: %v", err))
	}

	return toDataFrames(result)
}

// toDataFrames converts tstore response data into Grafana DataFrames.
// One frame per lookup key.
func toDataFrames(result map[string][]dataPoint) backend.DataResponse {
	var response backend.DataResponse

	for lookup, points := range result {
		times := make([]time.Time, 0, len(points))
		values := make([]*float64, 0, len(points))

		for _, p := range points {
			t, err := time.Parse(time.RFC3339, p.Ts)
			if err != nil {
				t, err = time.Parse("2006-01-02T15:04:05.999999999Z07:00", p.Ts)
				if err != nil {
					continue
				}
			}
			times = append(times, t)

			var fval *float64
			switch v := p.V.(type) {
			case float64:
				fval = &v
			case string:
				if v != "" {
					var f float64
					if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
						fval = &f
					}
				}
			}
			values = append(values, fval)
		}

		// Use the filter portion of the lookup as the frame name.
		frameName := lookup
		parts := strings.SplitN(lookup, "|", 3)
		if len(parts) >= 2 {
			frameName = parts[1]
		}

		frame := data.NewFrame(frameName,
			data.NewField("time", nil, times),
			data.NewField("value", nil, values),
		)
		frame.RefID = lookup
		response.Frames = append(response.Frames, frame)
	}

	return response
}
