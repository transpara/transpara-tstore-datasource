// CJS-compatible mock for @react-hookz/web (ESM-only, incompatible with Jest jsdom)
// Only useIsomorphicLayoutEffect is used by @grafana/data and @grafana/ui CJS builds.
const { useEffect } = require('react');

module.exports = { useIsomorphicLayoutEffect: useEffect };
