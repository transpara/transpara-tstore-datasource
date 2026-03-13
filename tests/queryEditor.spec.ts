import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render query editor in visual mode', async ({ panelEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await expect(page.getByText('Dataset')).toBeVisible();
  await expect(page.getByText('Lookups')).toBeVisible();
  await expect(page.getByText('Aggregation')).toBeVisible();
  await expect(page.getByText('Interval')).toBeVisible();
});

test('should show Raw mode editor when Raw is selected', async ({ panelEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await page.getByRole('radio', { name: 'Raw' }).click();
  await expect(page.getByText('Query JSON')).toBeVisible();
});
