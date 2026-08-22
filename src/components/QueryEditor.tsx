import React, { useCallback, useEffect, useState } from 'react';
import { QueryEditorProps } from '@grafana/data';
import { InlineField, Select, MultiSelect, TextArea, RadioButtonGroup, Input } from '@grafana/ui';
import { DataSource } from '../datasource';
import { DEFAULT_QUERY, MyDataSourceOptions, MyQuery } from '../types';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;

const AGG_OPTIONS = [
  { label: 'avg', value: 'avg' },
  { label: 'raw', value: 'raw' },
  { label: 'min', value: 'min' },
  { label: 'max', value: 'max' },
  { label: 'sum', value: 'sum' },
  { label: 'count', value: 'count' },
  { label: 'median', value: 'median' },
  { label: 'twavg', value: 'twavg' },
];

const MODE_OPTIONS = [
  { label: 'Visual', value: 'visual' as const },
  { label: 'Raw', value: 'raw' as const },
];

export function QueryEditor({ query, onChange, onRunQuery, datasource }: Props) {
  const q = { ...DEFAULT_QUERY, ...query };

  const [datasets, setDatasets] = useState<Array<{ label: string; value: string }>>([]);
  const [lookupOptions, setLookupOptions] = useState<Array<{ label: string; value: string }>>([]);
  const [selectedDataset, setSelectedDataset] = useState<string>('');

  // Load datasets on mount
  useEffect(() => {
    datasource.getDatasets().then((ds) => {
      setDatasets(ds.map((d) => ({ label: d, value: d })));
    });
  }, [datasource]);

  // Load lookups when dataset changes
  useEffect(() => {
    if (!selectedDataset) {
      return;
    }
    datasource.getLookups(selectedDataset).then((ls) => {
      setLookupOptions(ls.map((l) => ({ label: l, value: l })));
    });
  }, [datasource, selectedDataset]);

  const onModeChange = useCallback(
    (mode: 'visual' | 'raw') => {
      if (mode === 'raw') {
        const rawBody = JSON.stringify(query.lookups ?? [], null, 2);
        onChange({ ...DEFAULT_QUERY, ...query, queryType: 'raw', rawJson: rawBody });
      } else {
        onChange({ ...DEFAULT_QUERY, ...query, queryType: 'visual' });
      }
    },
    [query, onChange]
  );

  if (q.queryType === 'raw') {
    return (
      <div>
        <InlineField label="Mode" labelWidth={10}>
          <RadioButtonGroup options={MODE_OPTIONS} value={q.queryType} onChange={onModeChange} />
        </InlineField>
        <InlineField label="Query JSON" labelWidth={10} grow>
          <TextArea
            rows={8}
            value={q.rawJson ?? ''}
            onChange={(e) => onChange({ ...q, rawJson: e.currentTarget.value })}
            onBlur={onRunQuery}
            placeholder={'["dataset|filter|avg|5m"]'}
          />
        </InlineField>
      </div>
    );
  }

  return (
    <div>
      <InlineField label="Mode" labelWidth={10}>
        <RadioButtonGroup options={MODE_OPTIONS} value={q.queryType ?? 'visual'} onChange={onModeChange} />
      </InlineField>

      <InlineField label="Dataset" labelWidth={12}>
        <Select
          width={30}
          options={datasets}
          value={selectedDataset || null}
          onChange={(v) => setSelectedDataset(v.value ?? '')}
          placeholder="Select dataset..."
        />
      </InlineField>

      <InlineField label="Lookups" labelWidth={12}>
        <MultiSelect
          width={60}
          options={selectedDataset ? lookupOptions : []}
          value={q.lookups ?? []}
          onChange={(vals) => onChange({ ...q, lookups: vals.map((v) => v.value as string) })}
          onBlur={onRunQuery}
          placeholder="Select lookups..."
          isDisabled={!selectedDataset}
        />
      </InlineField>

      <InlineField label="Aggregation" labelWidth={12}>
        <Select
          width={16}
          options={AGG_OPTIONS}
          value={q.aggType ?? 'avg'}
          onChange={(v) => onChange({ ...q, aggType: v.value ?? 'avg' })}
        />
      </InlineField>

      <InlineField label="Interval" labelWidth={12} tooltip="e.g. 5m, 1h. Leave empty to use Grafana's $__interval.">
        <Input
          width={12}
          value={q.aggInt ?? ''}
          onChange={(e) => onChange({ ...q, aggInt: e.currentTarget.value })}
          onBlur={onRunQuery}
          placeholder="auto"
        />
      </InlineField>
    </div>
  );
}
