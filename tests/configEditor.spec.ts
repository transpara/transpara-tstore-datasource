import { test, expect } from '@grafana/plugin-e2e';

test('smoke: should render config editor fields', async ({ createDataSourceConfigPage, readProvisionedDataSource, page }) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  await createDataSourceConfigPage({ type: ds.type });
  await expect(page.getByLabel('tstore-interface URL')).toBeVisible();
  await expect(page.getByLabel('Keycloak Token URL')).toBeVisible();
  await expect(page.getByLabel('Client ID')).toBeVisible();
  await expect(page.getByLabel('Client Secret')).toBeVisible();
});

test('"Save & test" should be successful when configuration is valid', async ({
  createDataSourceConfigPage,
  readProvisionedDataSource,
  page,
}) => {
  const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
  const configPage = await createDataSourceConfigPage({ type: ds.type });
  await page.getByLabel('tstore-interface URL').fill(ds.jsonData.url ?? '');
  await page.getByLabel('Keycloak Token URL').fill(ds.jsonData.tokenUrl ?? '');
  await page.getByLabel('Client ID').fill(ds.jsonData.clientId ?? '');
  await expect(configPage.saveAndTest()).toBeOK();
});
