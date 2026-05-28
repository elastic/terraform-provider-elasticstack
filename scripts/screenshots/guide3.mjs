#!/usr/bin/env node
/**
 * Capture Guide 3 dashboard screenshots for kibana-dashboard-advanced.
 *
 * Prerequisites:
 *   - Kibana 9.4+ with sample logs installed
 *   - terraform apply in examples/guides/guide3-advanced/
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
 *   node scripts/screenshots/guide3.mjs
 */

import { chromium } from 'playwright';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, '../..');
const OUT_DIR = path.join(REPO_ROOT, 'templates/guides/images');
const DASHBOARD_TITLE = 'Advanced: Sections, ES|QL, and access control';

const KIBANA_URL = (process.env.KIBANA_URL ?? 'http://localhost:5601').replace(/\/$/, '');
const KIBANA_USER = process.env.KIBANA_USER ?? 'elastic';
const KIBANA_PASS = process.env.KIBANA_PASS ?? 'password';

const SCREENSHOTS = [
  { file: 'g3-01-full.png', mode: 'viewport-full' },
  { file: 'g3-04-heatmap.png', mode: 'panel', panelTitle: 'Requests by hour and response' },
  { file: 'g3-03-gauge.png', mode: 'panel', panelTitle: '95th percentile bytes', expandSection: 'Goal tracking' },
  { file: 'g3-02-collapsed.png', mode: 'viewport-full', collapseAllSections: true },
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
      const sections = body.data?.sections ?? body.sections ?? [];
      const panels = body.data?.panels ?? body.panels ?? [];
      return { id: d.id, count: sections.length + panels.length };
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

async function expandSection(page, title) {
  const selectors = [
    `[data-test-subj^="dashboardSectionHeader-"]:has-text("${title}")`,
    `.kbnDashboardSection__header:has-text("${title}")`,
    `button:has-text("${title}")`,
  ];

  for (const selector of selectors) {
    const header = page.locator(selector).first();
    if (await header.isVisible({ timeout: 3000 }).catch(() => false)) {
      await header.click();
      await page.waitForLoadState('networkidle');
      await page.waitForTimeout(2000);
      return;
    }
  }

  await page.getByText(title, { exact: true }).first().click();
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(2000);
}

async function collapseAllSections(page) {
  const sectionTitles = ['Activity heatmap', 'Goal tracking'];
  for (const title of sectionTitles) {
    const header = page.getByText(title, { exact: true }).first();
    if (await header.isVisible({ timeout: 5000 }).catch(() => false)) {
      await header.click();
      await page.waitForTimeout(500);
    }
  }
  await page.waitForLoadState('networkidle');
  await page.waitForTimeout(1500);
}

async function panelByTitle(page, title) {
  const panel = page.locator('[data-test-subj="embeddablePanel"]').filter({ hasText: title }).first();
  await panel.waitFor({ state: 'visible', timeout: 30_000 });
  return panel;
}

async function captureScreenshot(page, spec) {
  const outPath = path.join(OUT_DIR, spec.file);
  const minBytes = 5000;
  try {
    if (spec.mode === 'panel') {
      if (spec.expandSection) {
        await expandSection(page, spec.expandSection);
      }
      const panel = spec.panelTitle
        ? await panelByTitle(page, spec.panelTitle)
        : page.locator('[data-test-subj="embeddablePanel"]').nth(spec.index ?? 0);
      await panel.screenshot({ path: outPath });
    } else if (spec.mode === 'viewport-full') {
      if (spec.collapseAllSections) {
        await collapseAllSections(page);
      }
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

    const dashboardId = await findDashboardId(page.request, { preferFullest: true });
    const dashUrl = `${KIBANA_URL}/app/dashboards#/view/${dashboardId}?_g=(time:(from:now-7d,to:now))`;
    console.log(`Navigating to dashboard ${dashboardId}`);
    await page.goto(dashUrl, { waitUntil: 'domcontentloaded' });
    await waitForDashboardPanels(page);

    for (const spec of shots) {
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
