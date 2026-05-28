/**
 * Shared helpers for the per-guide Playwright screenshot scripts in this directory.
 * Each guide script supplies its own SCREENSHOTS spec, DASHBOARD_TITLE, and any
 * guide-specific UI interactions (filters, section expand/collapse, etc.).
 */

import { chromium } from 'playwright';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
export const REPO_ROOT = path.resolve(__dirname, '../..');
export const OUT_DIR = path.join(REPO_ROOT, 'templates/guides/images');

export const KIBANA_URL = (process.env.KIBANA_URL ?? 'http://localhost:5601').replace(/\/$/, '');
export const KIBANA_USER = process.env.KIBANA_USER ?? 'elastic';
export const KIBANA_PASS = process.env.KIBANA_PASS ?? 'password';

export const VIEWPORT = { width: 1440, height: 900 };

export function apiHeaders() {
  return {
    'kbn-xsrf': 'true',
    'x-elastic-internal-origin': 'kibana',
    Authorization: `Basic ${Buffer.from(`${KIBANA_USER}:${KIBANA_PASS}`).toString('base64')}`,
  };
}

export function screenshotsToCapture(specs) {
  const only = process.env.SCREENSHOT_ONLY?.split(',').map((s) => s.trim()).filter(Boolean);
  if (!only?.length) {
    return specs;
  }
  return specs.filter((spec) => only.includes(spec.file));
}

export async function loginIfNeeded(page) {
  await page.goto(`${KIBANA_URL}/login`, { waitUntil: 'domcontentloaded' });
  const usernameField = page.locator('[data-test-subj="loginUsername"]');
  const needsLogin = await usernameField
    .waitFor({ state: 'visible', timeout: 15_000 })
    .then(() => true)
    .catch(() => false);
  if (!needsLogin) {
    return;
  }
  await usernameField.fill(KIBANA_USER);
  await page.locator('[data-test-subj="loginPassword"]').fill(KIBANA_PASS);
  await page.locator('[data-test-subj="loginSubmit"]').click();
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 30_000 });
  await page.waitForLoadState('networkidle');
}

export async function waitForDashboardPanels(page) {
  await page.waitForLoadState('networkidle');
  await page.waitForSelector('[data-test-subj="embeddablePanel"]', { timeout: 60_000 });
  await page.waitForTimeout(3000);
}

const defaultPanelCount = (body) => {
  const panels = body.data?.panels ?? body.panels ?? [];
  return panels.length;
};

/**
 * Resolve a dashboard saved-object ID by title. When DASHBOARD_ID is set in the
 * environment, the value is validated against the candidate list and returned.
 * When multiple dashboards share `title`, `countFn(body)` (default: panel count)
 * picks the richest result so accidental shell/empty duplicates do not win.
 */
export async function findDashboardId(request, { title, countFn = defaultPanelCount } = {}) {
  if (!title) {
    throw new Error('findDashboardId: title is required');
  }
  const listRes = await request.get(`${KIBANA_URL}/api/dashboards?per_page=500`, { headers: apiHeaders() });
  if (!listRes.ok()) {
    throw new Error(`Dashboard list failed: HTTP ${listRes.status()} ${await listRes.text()}`);
  }
  const listBody = await listRes.json();
  const candidates = (listBody.dashboards ?? []).filter(
    (d) => d.title === title || d.data?.title === title,
  );
  if (!candidates.length) {
    throw new Error(`Dashboard not found with title "${title}". Run terraform apply first.`);
  }

  const requested = process.env.DASHBOARD_ID?.trim();
  if (requested) {
    const byId = candidates.find((d) => d.id === requested);
    if (!byId) {
      throw new Error(`Dashboard id "${requested}" not found for title "${title}".`);
    }
    return byId.id;
  }

  if (candidates.length === 1) {
    return candidates[0].id;
  }

  const counts = await Promise.all(
    candidates.map(async (d) => {
      const res = await request.get(`${KIBANA_URL}/api/dashboards/${d.id}`, { headers: apiHeaders() });
      if (!res.ok()) {
        return { id: d.id, count: 0 };
      }
      const body = await res.json();
      return { id: d.id, count: countFn(body) };
    }),
  );
  counts.sort((a, b) => b.count - a.count);
  return counts[0].id;
}

export function dashboardURL(dashboardId, { from = 'now-7d', to = 'now' } = {}) {
  return `${KIBANA_URL}/app/dashboards#/view/${dashboardId}?_g=(time:(from:${from},to:${to}))`;
}

export function assertScreenshotSize(filePath, { minBytes = 5000 } = {}) {
  const stat = fs.statSync(filePath);
  if (stat.size < minBytes) {
    throw new Error(`Screenshot too small (${stat.size} bytes) — panel may not have rendered`);
  }
  console.log(`OK ${path.basename(filePath)} (${stat.size} bytes)`);
}

/**
 * Open a fresh headless chromium with the shared viewport. Caller is responsible
 * for closing the browser (the returned object exposes `browser` for that).
 */
export async function openBrowser() {
  fs.mkdirSync(OUT_DIR, { recursive: true });
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: VIEWPORT,
    ignoreHTTPSErrors: true,
  });
  const page = await context.newPage();
  return { browser, context, page };
}

/**
 * Standard top-level wrapper: log + non-zero exit on failure.
 */
export function runMain(main) {
  main().catch((err) => {
    console.error(err);
    process.exit(1);
  });
}
