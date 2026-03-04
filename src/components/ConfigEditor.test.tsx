import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { ConfigEditor } from './ConfigEditor';

const defaultOptions = {
  id: 1,
  uid: 'test',
  orgId: 1,
  name: 'Test',
  type: 'transpara-tstore-datasource',
  typeName: 'TStore Datasource',
  typeLogoUrl: '',
  access: 'proxy' as const,
  url: '',
  user: '',
  basicAuth: false,
  basicAuthUser: '',
  isDefault: false,
  withCredentials: false,
  readOnly: false,
  jsonData: { url: '', tokenUrl: '', clientId: '' },
  secureJsonData: {},
  secureJsonFields: {},
};

const defaultProps = {
  options: defaultOptions as any,
  onOptionsChange: jest.fn(),
};

beforeEach(() => {
  jest.clearAllMocks();
});

test('renders all four fields', () => {
  render(<ConfigEditor {...defaultProps} />);
  expect(screen.getByPlaceholderText('https://tstore.internal:8080')).toBeInTheDocument();
  expect(screen.getByPlaceholderText('https://keycloak.internal/realms/transpara/protocol/openid-connect/token')).toBeInTheDocument();
  expect(screen.getByPlaceholderText('tstore-grafana')).toBeInTheDocument();
  // SecretInput renders a password input when not yet configured
  expect(document.querySelector('input[type="password"]')).toBeInTheDocument();
});

test('onOptionsChange called with correct url update', () => {
  const onChange = jest.fn();
  render(<ConfigEditor {...defaultProps} onOptionsChange={onChange} />);
  fireEvent.change(screen.getByPlaceholderText('https://tstore.internal:8080'), {
    target: { value: 'http://localhost:8080' },
  });
  expect(onChange).toHaveBeenCalledWith(
    expect.objectContaining({
      jsonData: expect.objectContaining({ url: 'http://localhost:8080' }),
    })
  );
});

test('onSecretReset sets secureJsonFields.clientSecret to false', () => {
  const onChange = jest.fn();
  render(
    <ConfigEditor
      {...defaultProps}
      onOptionsChange={onChange}
      options={{ ...defaultOptions, secureJsonFields: { clientSecret: true } } as any}
    />
  );
  // When isConfigured=true, SecretInput shows a Reset button
  fireEvent.click(screen.getByText(/Reset/i));
  expect(onChange).toHaveBeenCalledWith(
    expect.objectContaining({
      secureJsonFields: expect.objectContaining({ clientSecret: false }),
      secureJsonData: expect.objectContaining({ clientSecret: '' }),
    })
  );
});

test('onSecretChange preserves existing secureJsonData fields', () => {
  const onChange = jest.fn();
  render(
    <ConfigEditor
      {...defaultProps}
      onOptionsChange={onChange}
      options={{
        ...defaultOptions,
        secureJsonData: { clientSecret: '', someOtherField: 'preserved' },
      } as any}
    />
  );
  fireEvent.change(document.querySelector('input[type="password"]')!, {
    target: { value: 'new-secret' },
  });
  expect(onChange).toHaveBeenCalledWith(
    expect.objectContaining({
      secureJsonData: expect.objectContaining({
        clientSecret: 'new-secret',
        someOtherField: 'preserved',
      }),
    })
  );
});
