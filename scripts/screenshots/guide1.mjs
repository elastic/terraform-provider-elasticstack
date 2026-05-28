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
 *   SCREENSHOT_ONLY  optional comma-separated PNG filenames to capture (e.g. g1-01-shell.png)
 *   DASHBOARD_ID     optional dashboard id when multiple match the title
 *
 * Usage (from repo root):
 *   node scripts/screenshots/guide1.mjs
 */

import path from 'node:path';

import {
  KIBANA_URL,
  OUT_DIR,
  apiHeaders,
  assertScreenshotSize,
  dashboardURL,
  findDashboardId,
  loginIfNeeded,
  openBrowser,
  runMain,
  screenshotsToCapture,
  waitForDashboardPanels,
} from './lib.mjs';

const DASHBOARD_TITLE = 'Getting started: Web server logs';

const SCREENSHOTS = [
  { file: 'g1-01-shell.png', mode: 'viewport-top', shell: true },
  { file: 'g1-02-markdown.png', mode: 'panel', index: 0 },
  { file: 'g1-03-metric1.png', mode: 'panel', index: 1 },
  { file: 'g1-04-metric2.png', mode: 'panel', index: 2 },
  { file: 'g1-05-line.png', mode: 'panel', index: 3 },
  { file: 'g1-06-bar.png', mode: 'panel', index: 4 },
  { file: 'g1-07-final.png', mode: 'viewport-full' },
];

async function createShellDashboard(request) {
  const body = {
    title: DASHBOARD_TITLE,
    description: 'Temporary empty dashboard created for guide 1 shell screenshot. Safe to delete.',
    panels: [],
    time_range: { from: 'now-7d', to: 'now' },
    refresh_interval: { pause: true, value: 0 },
    query: { language: 'kql', expression: '' },
  };
  const res = await request.post(`${KIBANA_URL}/api/dashboards`, {
    headers: { ...apiHeaders(), 'content-type': 'application/json' },
    data: body,
  });
  if (!res.ok()) {
    throw new Error(`Failed to create shell dashboard: HTTP ${res.status()} ${await res.text()}`);
  }
  const json = await res.json();
  const id = json.id ?? json.data?.id;
  if (!id) {
    throw new Error(`Shell dashboard create returned no id: ${JSON.stringify(json)}`);
  }
  return id;
}

async function deleteDashboard(request, id) {
  try {
    await request.delete(`${KIBANA_URL}/api/dashboards/${id}`, { headers: apiHeaders() });
  } catch (err) {
    console.warn(`Warning: failed to delete shell dashboard ${id}: ${err.message}`);
  }
}

async function waitForDashboardShell(page) {
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(2000);
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
  assertScreenshotSize(outPath, { minBytes: spec.shell ? 2000 : 5000 });
}

async function main() {
  const shots = screenshotsToCapture(SCREENSHOTS);
  if (!shots.length) {
    throw new Error('No screenshots matched SCREENSHOT_ONLY filter');
  }

  const { browser, page } = await openBrowser();
  const shellShots = shots.filter((s) => s.shell);
  const panelShots = shots.filter((s) => !s.shell);
  let shellDashboardId;

  try {
    await loginIfNeeded(page);

    if (shellShots.length) {
      shellDashboardId = await createShellDashboard(page.request);
      console.log(`Navigating to shell dashboard ${shellDashboardId}`);
      await page.goto(dashboardURL(shellDashboardId), { waitUntil: 'domcontentloaded' });
      await waitForDashboardShell(page);
      for (const spec of shellShots) {
        await captureScreenshot(page, spec);
      }
    }

    if (panelShots.length) {
      const dashboardId = await findDashboardId(page.request, { title: DASHBOARD_TITLE });
      console.log(`Navigating to dashboard ${dashboardId}`);
      await page.goto(dashboardURL(dashboardId), { waitUntil: 'domcontentloaded' });
      await waitForDashboardPanels(page);
      for (const spec of panelShots) {
        await captureScreenshot(page, spec);
      }
    }
  } finally {
    if (shellDashboardId) {
      await deleteDashboard(page.request, shellDashboardId);
    }
    await browser.close();
  }
}

runMain(main);
