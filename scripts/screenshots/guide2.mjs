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

import { chromium } from 'playwright';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, '../..');
const OUT_DIR = path.join(REPO_ROOT, 'templates/guides/images');
const DASHBOARD_TITLE = 'Operations: eCommerce monitoring';
const FILTER_OPTION = "Men's Clothing";

const KIBANA_URL = (process.env.KIBANA_URL ?? 'http://localhost:5601').replace(/\/$/, '');
const KIBANA_USER = process.env.KIBANA_USER ?? 'elastic';
const KIBANA_PASS = process.env.KIBANA_PASS ?? 'password';

const SCREENSHOTS = [
  { file: 'g2-01-full.png', mode: 'viewport-full', needsFilter: false },
  { file: 'g2-02-filtered.png', mode: 'viewport-full', needsFilter: true },
  { file: 'g2-03-discover.png', mode: 'panel', index: 6 },
  { file: 'g2-04-table.png', mode: 'panel', index: 4 },
];

function screenshotsToCapture() {
  const only = process.env.SCREENSHOT_ONLY?.split(',').map((s) => s.trim()).filter(Boolean);
  if (!only?.length) {
    return SCREENSHOTS;
  }
  return SCREENSHOTS.filter((spec) => only.includes(spec.file));
}

function apiHeaders() {
  return {
    'kbn-xsrf': 'true',
    'x-elastic-internal-origin': 'kibana',
    Authorization: `Basic ${Buffer.from(`${KIBANA_USER}:${KIBANA_PASS}`).toString('base64')}`,
  };
}

async function findDashboardId(request, { preferFullest = true } = {}) {
  const listRes = await request.get(`${KIBANA_URL}/api/dashboards?per_page=500`, { headers: apiHeaders() });
  if (!listRes.ok()) {
    throw new Error(`Dashboard list failed: HTTP ${listRes.status()} ${await listRes.text()}`);
  }
  const listBody = await listRes.json();
  const candidates = (listBody.dashboards ?? []).filter(
    (d) => d.title === DASHBOARD_TITLE || d.data?.title === DASHBOARD_TITLE,
  );
  if (!candidates.length) {
    throw new Error(`Dashboard not found with title "${DASHBOARD_TITLE}". Run terraform apply first.`);
  }

  const dashboardId = process.env.DASHBOARD_ID?.trim();
  if (dashboardId) {
    const byId = candidates.find((d) => d.id === dashboardId);
    if (!byId) {
      throw new Error(`Dashboard id "${dashboardId}" not found for title "${DASHBOARD_TITLE}".`);
    }
    return byId.id;
  }

  const counts = await Promise.all(
    candidates.map(async (d) => {
      const res = await request.get(`${KIBANA_URL}/api/dashboards/${d.id}`, { headers: apiHeaders() });
      if (!res.ok()) {
        return { id: d.id, count: 0 };
      }
      const body = await res.json();
      const panels = body.data?.panels ?? body.panels ?? [];
      return { id: d.id, count: panels.length };
    }),
  );
  counts.sort((a, b) => (preferFullest ? b.count - a.count : a.count - b.count));
  return counts[0].id;
}

async function loginIfNeeded(page) {
  await page.goto(`${KIBANA_URL}/login`, { waitUntil: 'domcontentloaded' });
  const usernameField = page.locator('[data-test-subj="loginUsername"]');
  const needsLogin = await usernameField
    .waitFor({ state: 'visible', timeout: 15_000 })
    .then(() => true)
    .catch(() => false);
  if (needsLogin) {
    await usernameField.fill(KIBANA_USER);
    await page.locator('[data-test-subj="loginPassword"]').fill(KIBANA_PASS);
    await page.locator('[data-test-subj="loginSubmit"]').click();
    await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 30_000 });
    await page.waitForLoadState('networkidle');
  }
}

async function waitForDashboardPanels(page) {
  await page.waitForLoadState('networkidle');
  await page.waitForSelector('[data-test-subj="embeddablePanel"]', { timeout: 60_000 });
  await page.waitForTimeout(3000);
}

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
  const minBytes = 5000;
  try {
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

    const stat = fs.statSync(outPath);
    if (stat.size < minBytes) {
      throw new Error(`Screenshot too small (${stat.size} bytes) — panel may not have rendered`);
    }
    console.log(`OK ${spec.file} (${stat.size} bytes)`);
  } catch (err) {
    console.error(`FAILED ${spec.file}: ${err.message}`);
    throw err;
  }
}

async function main() {
  const shots = screenshotsToCapture();
  if (!shots.length) {
    throw new Error('No screenshots matched SCREENSHOT_ONLY filter');
  }

  fs.mkdirSync(OUT_DIR, { recursive: true });

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    ignoreHTTPSErrors: true,
  });
  const page = await context.newPage();

  try {
    await loginIfNeeded(page);

    const preferFullest = true;
    const dashboardId = await findDashboardId(page.request, { preferFullest });
    const dashUrl = `${KIBANA_URL}/app/dashboards#/view/${dashboardId}?_g=(time:(from:now-7d,to:now))`;
    console.log(`Navigating to dashboard ${dashboardId}`);
    await page.goto(dashUrl, { waitUntil: 'domcontentloaded' });
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

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
