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

const DASHBOARD_TITLE = 'Advanced: Sections, ES|QL, and access control';

// Section titles tracked here mirror the `sections[*].title` values in
// examples/guides/guide3-advanced/main.tf — keep them in sync when editing the example.
const SECTION_TITLES = ['Activity heatmap', 'Goal tracking'];

const SCREENSHOTS = [
  { file: 'g3-01-full.png', mode: 'viewport-full' },
  { file: 'g3-04-heatmap.png', mode: 'panel', panelTitle: 'Requests by hour and response' },
  { file: 'g3-03-gauge.png', mode: 'panel', panelTitle: '95th percentile bytes', expandSection: 'Goal tracking' },
  { file: 'g3-02-collapsed.png', mode: 'viewport-full', collapseAllSections: true },
];

// Sections may appear with different test-subj attributes across Kibana releases;
// try the stable selectors first, then fall back to a text match.
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
  for (const title of SECTION_TITLES) {
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

// Section panels factor into "richest dashboard" for title disambiguation here.
const countSectionsAndPanels = (body) => {
  const sections = body.data?.sections ?? body.sections ?? [];
  const panels = body.data?.panels ?? body.panels ?? [];
  return sections.length + panels.length;
};

async function captureScreenshot(page, spec) {
  const outPath = path.join(OUT_DIR, spec.file);
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

    const dashboardId = await findDashboardId(page.request, {
      title: DASHBOARD_TITLE,
      countFn: countSectionsAndPanels,
    });
    console.log(`Navigating to dashboard ${dashboardId}`);
    await page.goto(dashboardURL(dashboardId), { waitUntil: 'domcontentloaded' });
    await waitForDashboardPanels(page);

    for (const spec of shots) {
      await captureScreenshot(page, spec);
    }
  } finally {
    await browser.close();
  }
}

runMain(main);
