import React, { ChangeEvent } from 'react';
import { InlineField, Input, SecretInput } from '@grafana/ui';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { MyDataSourceOptions, MySecureJsonData } from '../types';

type Props = DataSourcePluginOptionsEditorProps<MyDataSourceOptions, MySecureJsonData>;

export function ConfigEditor({ options, onOptionsChange }: Props) {
  const { jsonData, secureJsonFields, secureJsonData } = options;

  const onJsonDataChange = (key: keyof MyDataSourceOptions) => (e: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({ ...options, jsonData: { ...jsonData, [key]: e.target.value } });
  };

  const onSecretChange = (e: ChangeEvent<HTMLInputElement>) => {
    onOptionsChange({ ...options, secureJsonData: { ...secureJsonData, clientSecret: e.target.value } });
  };

  const onSecretReset = () => {
    onOptionsChange({
      ...options,
      secureJsonFields: { ...secureJsonFields, clientSecret: false },
      secureJsonData: { ...secureJsonData, clientSecret: '' },
    });
  };

  return (
    <>
      <InlineField label="tstore-interface URL" labelWidth={24} tooltip="e.g. https://tstore.internal:8080">
        <Input
          width={40}
          value={jsonData.url || ''}
          onChange={onJsonDataChange('url')}
          placeholder="https://tstore.internal:8080"
          aria-label="tstore-interface URL"
        />
      </InlineField>

      <InlineField label="Keycloak Token URL" labelWidth={24} tooltip="e.g. https://keycloak.internal/realms/transpara/protocol/openid-connect/token">
        <Input
          width={60}
          value={jsonData.tokenUrl || ''}
          onChange={onJsonDataChange('tokenUrl')}
          placeholder="https://keycloak.internal/realms/transpara/protocol/openid-connect/token"
          aria-label="Keycloak Token URL"
        />
      </InlineField>

      <InlineField label="Client ID" labelWidth={24}>
        <Input
          width={40}
          value={jsonData.clientId || ''}
          onChange={onJsonDataChange('clientId')}
          placeholder="tstore-grafana"
          aria-label="Client ID"
        />
      </InlineField>

      <InlineField label="Client Secret" labelWidth={24}>
        <SecretInput
          width={40}
          isConfigured={Boolean(secureJsonFields?.clientSecret)}
          value={secureJsonData?.clientSecret || ''}
          onReset={onSecretReset}
          onChange={onSecretChange}
          aria-label="Client Secret"
        />
      </InlineField>
    </>
  );
}
