import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryEditor } from './QueryEditor';
import { DEFAULT_QUERY } from '../types';

// Mock DataSource
const mockDs = {
  getDatasets: jest.fn().mockResolvedValue(['plant-a', 'plant-b']),
  getLookups: jest.fn().mockResolvedValue(['plant-a|sensor_id=123', 'plant-a|sensor_id=456']),
};

const defaultProps = {
  datasource: mockDs as any,
  query: { ...DEFAULT_QUERY, refId: 'A' } as any,
  onChange: jest.fn(),
  onRunQuery: jest.fn(),
};

beforeEach(() => {
  jest.clearAllMocks();
});

test('renders visual mode by default', async () => {
  render(<QueryEditor {...defaultProps} />);
  await waitFor(() => {
    expect(screen.getAllByText(/Dataset/i)[0]).toBeInTheDocument();
  });
});

test('toggle switches to raw mode', async () => {
  render(<QueryEditor {...defaultProps} />);
  await waitFor(() => screen.getAllByText(/Raw/i)[0]);
  const rawButton = screen.getAllByText(/Raw/i)[0];
  fireEvent.click(rawButton);
  expect(defaultProps.onChange).toHaveBeenCalledWith(
    expect.objectContaining({ queryType: 'raw' })
  );
});

test('visual to raw toggle serializes lookups to rawJson', async () => {
  const onChange = jest.fn();
  render(
    <QueryEditor
      {...defaultProps}
      onChange={onChange}
      query={{
        ...DEFAULT_QUERY,
        refId: 'A',
        queryType: 'visual',
        lookups: ['plant-a|sensor_id=123'],
      } as any}
    />
  );
  await waitFor(() => screen.getAllByText(/Raw/i)[0]);
  fireEvent.click(screen.getAllByText(/Raw/i)[0]);
  expect(onChange).toHaveBeenCalledWith(
    expect.objectContaining({
      queryType: 'raw',
      rawJson: expect.stringContaining('plant-a|sensor_id=123'),
    })
  );
});

test('raw mode renders a textarea', () => {
  render(
    <QueryEditor
      {...defaultProps}
      query={{ ...DEFAULT_QUERY, refId: 'A', queryType: 'raw' } as any}
    />
  );
  expect(screen.getByRole('textbox')).toBeInTheDocument();
});
