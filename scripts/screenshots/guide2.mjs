#!/usr/bin/env node
/**
 * Capture Guide 2 dashboard screenshots for kibana-dashboard-operations.
 *
 * Prerequisites:
 *   - Kibana 9.4+ with sample eCommerce data installed
 *   - terraform apply in examples/guides/guide2-operations/
 *   - npm install && npx playwright install chromium (in scripts/screenshots/)
 *
 * Environment:
 *   KIBANA_URL  (default http://localhost:5601)
 *   KIBANA_USER (default elastic)
 *   KIBANA_PASS (default password)
 *   SCREENSHOT_ONLY  optional comma-separated PNG filenames to capture
 *   DASHBOARD_ID     optional dashboard id when multiple match the title
 *
 * Usage (from repo root):
 *   node scripts/screenshots/guide2.mjs
 */

import path from 'node:path';

import {
  OUT_DIR,
  assertScreenshotSize,
  dashboardURL,
  findDashboardId,
  loginIfNeeded,
  openBrowser,
  runMain,
  screenshotsToCapture,
  waitForDashboardPanels,
} from './lib.mjs';

const DASHBOARD_TITLE = 'Operations: eCommerce monitoring';
const FILTER_OPTION = "Men's Clothing";

// Panel shots run before the category filter is applied; full-dashboard shots follow.
const SCREENSHOTS = [
  { file: 'g2-03-discover.png', mode: 'panel', index: 6 },
  { file: 'g2-04-table.png', mode: 'panel', index: 4 },
  { file: 'g2-01-full.png', mode: 'viewport-full' },
  { file: 'g2-02-filtered.png', mode: 'viewport-full', needsFilter: true },
];

async function applyCategoryFilter(page) {
  const controlButton = page.locator('[data-test-subj^="optionsList-control-"]').first();
  await controlButton.waitFor({ state: 'visible', timeout: 30_000 });
  await controlButton.click();

  const popover = page.locator('[data-test-subj="optionsListPopover"]');
  const popoverVisible = await popover.isVisible({ timeout: 5000 }).catch(() => false);

  const optionByTestSubj = page.locator(
    `[data-test-subj="optionsList-control-selection-${FILTER_OPTION}"]`,
  );
  if (await optionByTestSubj.isVisible({ timeout: 3000 }).catch(() => false)) {
    await optionByTestSubj.click();
  } else if (popoverVisible) {
    await popover.getByText(FILTER_OPTION, { exact: true }).click();
  } else {
    await page.getByText(FILTER_OPTION, { exact: true }).click();
  }

  await page.keyboard.press('Escape');
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(3000);
}

async function captureScreenshot(page, spec) {
  const outPath = path.join(OUT_DIR, spec.file);
  if (spec.mode === 'panel') {
    const panels = page.locator('[data-test-subj="embeddablePanel"]');
    const count = await panels.count();
    if (spec.index >= count) {
      throw new Error(`Panel index ${spec.index} out of range (${count} panels)`);
    }
    await panels.nth(spec.index).screenshot({ path: outPath });
  } else if (spec.mode === 'viewport-full') {
    await page.screenshot({ path: outPath, fullPage: true });
  } else {
    await page.screenshot({ path: outPath, fullPage: false });
  }
  assertScreenshotSize(outPath);
}

async function main() {
  const shots = screenshotsToCapture(SCREENSHOTS);
  if (!shots.length) {
    throw new Error('No screenshots matched SCREENSHOT_ONLY filter');
  }

  const { browser, page } = await openBrowser();
  try {
    await loginIfNeeded(page);

    const dashboardId = await findDashboardId(page.request, { title: DASHBOARD_TITLE });
    console.log(`Navigating to dashboard ${dashboardId}`);
    await page.goto(dashboardURL(dashboardId), { waitUntil: 'domcontentloaded' });
    await waitForDashboardPanels(page);

    for (const spec of shots) {
      if (spec.needsFilter) {
        await applyCategoryFilter(page);
      }
      await captureScreenshot(page, spec);
    }
  } finally {
    await browser.close();
  }
}

runMain(main);
