import {
  DataQueryRequest,
  DataQueryResponse,
  DataSourceInstanceSettings,
} from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { MyDataSourceOptions, MyQuery } from './types';

export class DataSource extends DataSourceWithBackend<MyQuery, MyDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<MyDataSourceOptions>) {
    super(instanceSettings);
  }

  // query delegates to the Go backend via Grafana's backend plugin framework.
  query(options: DataQueryRequest<MyQuery>): Promise<DataQueryResponse> {
    return super.query(options);
  }

  // getDatasets fetches the list of dataset names via Go CallResource.
  async getDatasets(): Promise<string[]> {
    try {
      return await this.getResource('datasets');
    } catch {
      return [];
    }
  }

  // getLookups fetches lookup strings for a given dataset.
  async getLookups(dataset: string): Promise<string[]> {
    try {
      return await this.getResource('lookups', { dataset });
    } catch {
      return [];
    }
  }
}
