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
  { file: 'g1-01-shell.png', mode: 'viewport-top', shell: true },
  { file: 'g1-02-markdown.png', mode: 'panel', index: 0 },
  { file: 'g1-03-metric1.png', mode: 'panel', index: 1 },
  { file: 'g1-04-metric2.png', mode: 'panel', index: 2 },
  { file: 'g1-05-line.png', mode: 'panel', index: 3 },
  { file: 'g1-06-bar.png', mode: 'panel', index: 4 },
  { file: 'g1-07-final.png', mode: 'viewport-full' },
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

async function findDashboardId(request) {
  const dashboardId = process.env.DASHBOARD_ID?.trim();
  if (dashboardId) {
    return dashboardId;
  }
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
  if (candidates.length === 1) {
    return candidates[0].id;
  }
  // Multiple matches: list API omits panels; fetch each and pick the fullest
  // so we screenshot the real (panel-populated) dashboard rather than a leftover shell.
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
  counts.sort((a, b) => b.count - a.count);
  return counts[0].id;
}

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

async function waitForDashboardShell(page) {
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(2000);
}

async function waitForDashboardPanels(page) {
  await page.waitForLoadState('networkidle');
  await page.waitForSelector('[data-test-subj="embeddablePanel"]', { timeout: 60_000 });
  await page.waitForTimeout(3000);
}

async function captureScreenshot(page, spec) {
  const outPath = path.join(OUT_DIR, spec.file);
  const minBytes = spec.shell ? 2000 : 5000;
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

  const shellShots = shots.filter((s) => s.shell);
  const panelShots = shots.filter((s) => !s.shell);
  let shellDashboardId;

  try {
    await loginIfNeeded(page);

    if (shellShots.length) {
      shellDashboardId = await createShellDashboard(page.request);
      const shellUrl = `${KIBANA_URL}/app/dashboards#/view/${shellDashboardId}?_g=(time:(from:now-7d,to:now))`;
      console.log(`Navigating to shell dashboard ${shellDashboardId}`);
      await page.goto(shellUrl, { waitUntil: 'domcontentloaded' });
      await waitForDashboardShell(page);
      for (const spec of shellShots) {
        await captureScreenshot(page, spec);
      }
    }

    if (panelShots.length) {
      const dashboardId = await findDashboardId(page.request);
      const dashUrl = `${KIBANA_URL}/app/dashboards#/view/${dashboardId}?_g=(time:(from:now-7d,to:now))`;
      console.log(`Navigating to dashboard ${dashboardId}`);
      await page.goto(dashUrl, { waitUntil: 'domcontentloaded' });
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

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
