import { DataQuery, DataSourceJsonData } from '@grafana/schema';

export interface MyQuery extends DataQuery {
  queryType: 'visual' | 'raw';
  // Visual mode
  lookups: string[];   // full lookup strings: "dataset|filter" or "dataset|filter|agg_type|agg_int"
  aggType: string;     // global agg_type override (empty = use per-lookup agg)
  aggInt: string;      // global agg_int override (empty = use per-lookup agg)
  tz: string;
  // Raw mode
  rawJson: string;     // JSON body sent directly to /api/v1/read/trend-data
}

export const DEFAULT_QUERY: Partial<MyQuery> = {
  queryType: 'visual',
  lookups: [],
  aggType: 'avg',
  aggInt: '',
  tz: 'UTC',
  rawJson: JSON.stringify([], null, 2),
};

export interface MyDataSourceOptions extends DataSourceJsonData {
  url: string;
  tokenUrl: string;
  clientId: string;
}

export interface MySecureJsonData {
  clientSecret?: string;
}
