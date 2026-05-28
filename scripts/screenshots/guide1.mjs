#!/usr/bin/env node
/**
 * Capture Guide 1 dashboard screenshots for kibana-dashboard-getting-started.
 *
 * Prerequisites:
 *   - Kibana 9.4+ with sample logs installed
 *   - terraform apply in examples/guides/guide1-getting-started/
 *   - npm install && npx playwright install chromium (in scripts/screenshots/)
 *
 * Environment:
 *   KIBANA_URL  (default http://localhost:5601)
 *   KIBANA_USER (default elastic)
 *   KIBANA_PASS (default password)
 *
 * Usage (from repo root):
 *   node scripts/screenshots/guide1.mjs
 */

import { chromium } from 'playwright';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, '../..');
const OUT_DIR = path.join(REPO_ROOT, 'templates/guides/images');
const DASHBOARD_TITLE = 'Getting started: Web server logs';

const KIBANA_URL = (process.env.KIBANA_URL ?? 'http://localhost:5601').replace(/\/$/, '');
const KIBANA_USER = process.env.KIBANA_USER ?? 'elastic';
const KIBANA_PASS = process.env.KIBANA_PASS ?? 'password';

const SCREENSHOTS = [
  { file: 'g1-01-shell.png', mode: 'viewport-top' },
  { file: 'g1-02-markdown.png', mode: 'panel', index: 0 },
  { file: 'g1-03-metric1.png', mode: 'panel', index: 1 },
  { file: 'g1-04-metric2.png', mode: 'panel', index: 2 },
  { file: 'g1-05-line.png', mode: 'panel', index: 3 },
  { file: 'g1-06-bar.png', mode: 'panel', index: 4 },
  { file: 'g1-07-final.png', mode: 'viewport-full' },
  { file: 'g1-08-overview.png', mode: 'viewport-top' },
];

function apiHeaders() {
  return {
    'kbn-xsrf': 'true',
    'x-elastic-internal-origin': 'kibana',
    Authorization: `Basic ${Buffer.from(`${KIBANA_USER}:${KIBANA_PASS}`).toString('base64')}`,
  };
}

async function findDashboardId(request) {
  const listRes = await request.get(`${KIBANA_URL}/api/dashboards?per_page=500`, { headers: apiHeaders() });
  if (!listRes.ok()) {
    throw new Error(`Dashboard list failed: HTTP ${listRes.status()} ${await listRes.text()}`);
  }
  const listBody = await listRes.json();
  const match = (listBody.dashboards ?? []).find(
    (d) => d.title === DASHBOARD_TITLE || d.data?.title === DASHBOARD_TITLE,
  );
  if (!match?.id) {
    throw new Error(`Dashboard not found with title "${DASHBOARD_TITLE}". Run terraform apply first.`);
  }
  return match.id;
}

async function waitForDashboardPanels(page) {
  await page.waitForLoadState('networkidle');
  await page.waitForSelector('[data-test-subj="embeddablePanel"]', { timeout: 60_000 });
  await page.waitForTimeout(3000);
}

async function captureScreenshot(page, spec) {
  const outPath = path.join(OUT_DIR, spec.file);
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
    if (stat.size < 5000) {
      throw new Error(`Screenshot too small (${stat.size} bytes) — panel may not have rendered`);
    }
    console.log(`OK ${spec.file} (${stat.size} bytes)`);
  } catch (err) {
    console.error(`FAILED ${spec.file}: ${err.message}`);
    throw err;
  }
}

async function main() {
  fs.mkdirSync(OUT_DIR, { recursive: true });

  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    ignoreHTTPSErrors: true,
  });
  const page = await context.newPage();

  try {
    await page.goto(`${KIBANA_URL}/login`, { waitUntil: 'domcontentloaded' });
    await page.locator('[data-test-subj="loginUsername"]').fill(KIBANA_USER);
    await page.locator('[data-test-subj="loginPassword"]').fill(KIBANA_PASS);
    await page.locator('[data-test-subj="loginSubmit"]').click();
    await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 30_000 });

    const dashboardId = await findDashboardId(page.request);
    const dashUrl = `${KIBANA_URL}/app/dashboards#/view/${dashboardId}?_g=(time:(from:now-7d,to:now))`;
    console.log(`Navigating to dashboard ${dashboardId}`);
    await page.goto(dashUrl, { waitUntil: 'domcontentloaded' });
    await waitForDashboardPanels(page);

    for (const spec of SCREENSHOTS) {
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
