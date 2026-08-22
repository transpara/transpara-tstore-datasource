# Architecture

## System Context

The plugin sits between Grafana and the Transpara Platform. It has no database of its own — it translates Grafana queries into tstore-interface HTTP calls and maps the responses back into Grafana DataFrames.

```mermaid
graph TD
    User([User])
    Grafana[Grafana Server]
    Plugin[transpara-tstore-datasource\nGo backend process]
    TStore[tstore-interface\nFastAPI / Python]
    TSDB[(TimescaleDB\nPostgreSQL)]
    Keycloak[Keycloak\nOIDC / OAuth2]

    User -->|views dashboards| Grafana
    Grafana -->|gRPC plugin pipe| Plugin
    Plugin -->|client_credentials grant| Keycloak
    Keycloak -->|access_token| Plugin
    Plugin -->|Bearer token\nHTTP REST| TStore
    TStore -->|SQL| TSDB
    TStore -->|time-series JSON| Plugin
    Plugin -->|DataFrames| Grafana
    Grafana -->|renders panel| User
```

Keycloak realm: `transpara`. The plugin uses the **client credentials** OAuth2 flow (service account — no user login). Tokens are cached in memory per datasource instance and refreshed automatically on expiry or HTTP 401.

---

## Components

### Frontend (TypeScript / React)

```mermaid
graph LR
    module.ts --> DataSource
    module.ts --> ConfigEditor
    module.ts --> QueryEditor
    DataSource -->|CallResource| GoBackend[Go backend]
    QueryEditor -->|getDatasets| DataSource
    QueryEditor -->|getLookups| DataSource
```

| File | Role |
|---|---|
| `src/module.ts` | Webpack entry point — wires the three classes together |
| `src/datasource.ts` | Extends `DataSourceWithBackend`; thin wrapper — only exposes `getDatasets()` and `getLookups()` via `CallResource` |
| `src/components/ConfigEditor.tsx` | Datasource config form (URL, token URL, client ID, client secret) |
| `src/components/QueryEditor.tsx` | Panel query editor — visual/raw toggle, dataset/lookup dropdowns, aggregation controls |
| `src/types.ts` | TypeScript interfaces for `MyQuery`, `MyDataSourceOptions`, `MySecureJsonData` |

The frontend does **no direct HTTP calls to tstore-interface**. All network requests go through the Go backend via Grafana's plugin pipe.

### Backend (Go)

```mermaid
graph TD
    SDK[Grafana Plugin SDK]
    SDK -->|QueryData| query.go
    SDK -->|CheckHealth| datasource.go
    SDK -->|CallResource| datasource.go
    datasource.go -->|doRequest| auth.go
    auth.go -->|FetchToken| Keycloak
    datasource.go -->|HTTP| TStore
    query.go -->|doRequest via datasource| TStore
    query.go -->|toDataFrames| SDK
```

| File | Role |
|---|---|
| `pkg/main.go` | Binary entry point — registers plugin factory with Grafana SDK |
| `pkg/plugin/datasource.go` | `Datasource` struct; implements `CheckHealth`, `CallResource`; `doRequest` with 401 retry |
| `pkg/plugin/query.go` | `QueryData`, `runQuery`, `toDataFrames`, `formatDuration` |
| `pkg/plugin/auth.go` | `TokenCache` (mutex-protected), `FetchToken` (client credentials POST) |

---

## Data Flow: Panel Query

```mermaid
sequenceDiagram
    actor User
    participant GF as Grafana Frontend
    participant BE as Plugin Go Backend
    participant KC as Keycloak
    participant TS as tstore-interface

    User->>GF: selects time range, runs panel
    GF->>BE: QueryData (gRPC) — lookups[], aggType, aggInt, timeRange
    BE->>BE: TokenCache.GetToken()
    alt cache miss or expired
        BE->>KC: POST /token (client_credentials)
        KC-->>BE: access_token, expires_in
        BE->>BE: cache token (expires_in - 30s)
    end
    BE->>TS: POST /api/v1/read/trend-data?start_time=...&end_time=...&agg_type=...&agg_int=...&width=...
    note over BE,TS: body = JSON array of lookup strings
    alt HTTP 401
        BE->>BE: TokenCache.Reset()
        BE->>KC: POST /token (re-fetch)
        KC-->>BE: new access_token
        BE->>TS: retry POST /api/v1/read/trend-data
    end
    TS-->>BE: { "lookup_key": [{ "ts": "...", "v": "42.5" }, ...], ... }
    BE->>BE: toDataFrames() — one DataFrame per lookup key
    BE-->>GF: DataFrames
    GF-->>User: renders time-series panel
```

### Data Flow: Dropdown Population

```mermaid
sequenceDiagram
    participant QE as QueryEditor (React)
    participant DS as datasource.ts
    participant BE as Plugin Go Backend
    participant TS as tstore-interface

    QE->>DS: getDatasets()
    DS->>BE: CallResource("datasets")
    BE->>TS: GET /api/v1/dataset
    TS-->>BE: ["plant-a", "plant-b", ...]
    BE-->>DS: JSON array (pass-through)
    DS-->>QE: string[]

    QE->>DS: getLookups("plant-a")
    DS->>BE: CallResource("lookups?dataset=plant-a")
    BE->>TS: GET /api/v1/lookups/?filter=plant-a&limit=1000
    TS-->>BE: { "results": [...], "total_count": N }
    BE->>BE: unwrap → results[]
    BE-->>DS: JSON array
    DS-->>QE: string[]
```

---

## Auth: Token Cache State

```mermaid
stateDiagram-v2
    [*] --> Empty
    Empty --> Valid: FetchToken() succeeds\ncache token for (expires_in - 30s)
    Valid --> Valid: GetToken() — return cached
    Valid --> Empty: expiry reached\nor Reset() called on 401
    Empty --> Error: FetchToken() fails\n(network / bad credentials)
    Error --> Empty: next GetToken() retries
```

Token lifetime: whatever Keycloak issues minus a 30-second safety buffer. If a 401 comes back from tstore-interface mid-flight, the cache is cleared and the request retries exactly once with a fresh token. A second consecutive 401 is returned as an error.

---

## Lookup String Format

The lookup string is the central data model shared between the frontend, the Go backend, and tstore-interface.

```
dataset|filter|agg_type|agg_int
```

Examples:
```
plant-a|sensor_id=motor-1,unit=rpm|avg|5m
plant-a|sensor_id=motor-1,unit=rpm
```

| Segment | Description |
|---|---|
| `dataset` | The tstore dataset name (e.g. `plant-a`) |
| `filter` | Comma-separated label pairs identifying the time series |
| `agg_type` | Aggregation function (`avg`, `min`, `max`, `sum`, `count`, `median`, `twavg`, `raw`). In visual mode this is a separate query param, not embedded in the string. |
| `agg_int` | Aggregation interval (e.g. `5m`, `1h`). Also a separate query param in visual mode. |

In **visual mode**, `agg_type` and `agg_int` are sent as URL query parameters — the lookup strings in the body contain only the first two segments. The Go backend splices in the query-level values. In **raw mode**, the user supplies the full JSON array verbatim; the backend forwards it unchanged.

The Grafana panel legend shows the `filter` segment (index 1) as the series name, not the full string.

---

## tstore-interface API Surface Used

| Method | Path | Used by |
|---|---|---|
| `GET` | `/api/v1/up` | `CheckHealth` |
| `GET` | `/api/v1/dataset` | `CallResource("datasets")` |
| `GET` | `/api/v1/lookups/?filter=<dataset>&limit=1000` | `CallResource("lookups?dataset=<name>")` |
| `POST` | `/api/v1/read/trend-data?<params>` | `QueryData` |

All requests carry a `Bearer <token>` header. The tstore-interface base URL is configured per datasource instance and never hard-coded.
