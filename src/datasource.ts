import {
  DataQueryRequest,
  DataQueryResponse,
  DataSourceApi,
  DataSourceInstanceSettings,
  TestDataSourceResponse,
} from '@grafana/data';
import { getBackendSrv } from '@grafana/runtime';
import { MyDataSourceOptions, MyQuery } from './types';

export class DataSource extends DataSourceApi<MyQuery, MyDataSourceOptions> {
  baseUrl: string;

  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
    this.baseUrl = instanceSettings.url!;
  }

  // query delegates to the Go backend via Grafana's data proxy.
  async query(options: DataQueryRequest<MyQuery>): Promise<DataQueryResponse> {
    const { data } = await getBackendSrv().fetch<DataQueryResponse>({
      url: `${this.baseUrl}/query`,
      method: 'POST',
      data: options,
    }).toPromise();
    return data;
  }

  // testDatasource calls the Go backend's CheckHealth.
  async testDatasource(): Promise<TestDataSourceResponse> {
    try {
      await getBackendSrv().fetch({ url: `${this.baseUrl}/health`, method: 'GET' }).toPromise();
      return { status: 'success', message: 'Data source connected and auth verified.' };
    } catch (err: any) {
      return { status: 'error', message: err?.data?.message ?? err?.message ?? 'Connection failed' };
    }
  }

  // getDatasets fetches the list of dataset names via Go CallResource.
  async getDatasets(): Promise<string[]> {
    try {
      const response = await getBackendSrv().fetch<string[]>({
        url: `${this.baseUrl}/resources/datasets`,
        method: 'GET',
      }).toPromise();
      return response.data ?? [];
    } catch {
      return [];
    }
  }

  // getLookups fetches lookup strings for a given dataset.
  async getLookups(dataset: string): Promise<string[]> {
    try {
      const response = await getBackendSrv().fetch<string[]>({
        url: `${this.baseUrl}/resources/lookups?dataset=${encodeURIComponent(dataset)}`,
        method: 'GET',
      }).toPromise();
      return response.data ?? [];
    } catch {
      return [];
    }
  }
}
