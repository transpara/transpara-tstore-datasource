// force timezone to UTC to allow tests to work regardless of local timezone
// generally used by snapshots, but can affect specific tests
process.env.TZ = 'UTC';

const { grafanaESModules, nodeModulesToTransform } = require('./.config/jest/utils');

// @react-hookz/web is ESM-only and added as a dep by @grafana/ui v13
const additionalESModules = [...grafanaESModules, '@react-hookz/web'];

const swcTransform = [
  '@swc/jest',
  {
    sourceMaps: 'inline',
    jsc: {
      parser: { syntax: 'typescript', tsx: true, decorators: false, dynamicImport: true },
    },
    module: { type: 'commonjs' },
  },
];

const baseConfig = require('./.config/jest.config');

module.exports = {
  // Jest configuration provided by Grafana scaffolding
  ...baseConfig,
  transformIgnorePatterns: [nodeModulesToTransform(additionalESModules)],
  transform: { '^.+\\.(t|j)sx?$': swcTransform },
  moduleNameMapper: {
    ...baseConfig.moduleNameMapper,
    // @react-hookz/web is ESM-only; map to CJS-compatible mock for Jest jsdom
    '@react-hookz/web': '<rootDir>/src/__mocks__/react-hookz-web.js',
  },
};
