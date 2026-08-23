import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render query editor in visual mode', async ({ panelEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await expect(page.getByText('Dataset').first()).toBeVisible();
  await expect(page.getByText('Lookups').first()).toBeVisible();
  await expect(page.getByText('Aggregation').first()).toBeVisible();
  await expect(page.getByText('Interval').first()).toBeVisible();
});

test('should show Raw mode editor when Raw is selected', async ({ panelEditPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await panelEditPage.datasource.set(ds.name);
  await page.getByRole('radio', { name: 'Raw' }).click({ force: true });
  await expect(page.getByText('Query JSON')).toBeVisible();
});
