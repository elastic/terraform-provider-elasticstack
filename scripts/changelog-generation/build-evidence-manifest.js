'use strict';
/**
 * build-evidence-manifest.js
 *
 * Standalone deterministic helper that fetches merged PRs in a git compare range,
 * classifies each PR, and builds the changelog evidence manifest.
 *
 * Usage (CLI):
 *   GITHUB_TOKEN=... \
 *   GITHUB_REPOSITORY=owner/repo \
 *   node scripts/changelog-generation/build-evidence-manifest.js \
 *     [--previous-tag v0.14.3] \
 *     [--compare-range v0.14.3..HEAD] \
 *     [--mode unreleased|release] \
 *     [--target-version 0.14.4] \
 *     [--output path/to/evidence.json]
 *
 * Env vars (alternative to CLI args):
 *   PREVIOUS_TAG, COMPARE_RANGE, MODE, TARGET_VERSION, OUTPUT_PATH
 *
 * Exports: classifyPR, fetchPRsInRange, buildManifest (for unit testing)
 */

const { execSync } = require('node:child_process');
const https = require('node:https');
const fs = require('node:fs');
const path = require('node:path');

// ---------------------------------------------------------------------------
// Classification constants (mirrors gather-pr-evidence.inline.js)
// ---------------------------------------------------------------------------

const USER_FACING_LABELS = new Set([
  'enhancement',
  'bug',
  'feature',
  'breaking-change',
  'deprecation',
  'new-resource',
  'new-data-source',
]);

const INTERNAL_LABELS = new Set([
  'dependencies',
  'chore',
  'internal',
  'documentation',
  'ci',
  'test',
  'openspec',
]);

const PROVIDER_PATH_PREFIXES = ['internal/', 'pkg/', 'libs/', 'provider/', 'go.mod', 'go.sum'];

// ---------------------------------------------------------------------------
// Classification logic
// ---------------------------------------------------------------------------

/**
 * Classifies a PR as 'user-facing', 'internal', or 'uncertain'.
 *
 * @param {{ user?: { login?: string }, labels?: Array<{ name: string }> }} pr
 * @param {Array<{ filename: string }>} files
 * @returns {{ classification: string, inclusion_rationale: string|null, exclusion_rationale: string|null }}
 */
function classifyPR(pr, files) {
  const labels = (pr.labels ?? []).map((l) => l.name);

  const hasUserFacingLabel = labels.some((l) => USER_FACING_LABELS.has(l));
  const hasInternalLabel = labels.some((l) => INTERNAL_LABELS.has(l));

  const isAutomated =
    pr.user?.login === 'dependabot[bot]' ||
    pr.user?.login === 'dependabot' ||
    pr.user?.login === 'github-actions[bot]';

  const touchesProviderCode = (files ?? []).some((f) =>
    PROVIDER_PATH_PREFIXES.some((prefix) => f.filename.startsWith(prefix))
  );

  const openspecOnly =
    (files ?? []).length > 0 && (files ?? []).every((f) => f.filename.startsWith('openspec/'));

  let classification;
  let inclusion_rationale = null;
  let exclusion_rationale = null;

  if (isAutomated) {
    classification = 'internal';
    exclusion_rationale = `Automated PR by ${pr.user?.login}`;
  } else if (openspecOnly) {
    classification = 'internal';
    exclusion_rationale = 'Touches only openspec/ files — no provider code changes';
  } else if (hasUserFacingLabel) {
    classification = 'user-facing';
    inclusion_rationale = `Has user-facing label(s): ${labels.filter((l) => USER_FACING_LABELS.has(l)).join(', ')}`;
  } else if (hasInternalLabel && !touchesProviderCode) {
    classification = 'internal';
    exclusion_rationale = `Has internal label(s): ${labels.filter((l) => INTERNAL_LABELS.has(l)).join(', ')} and does not touch provider code`;
  } else if (touchesProviderCode) {
    classification = 'user-facing';
    inclusion_rationale = 'Touches provider implementation paths — presumed user-facing';
  } else {
    classification = 'uncertain';
    inclusion_rationale = 'Classification uncertain — agent to decide';
  }

  return { classification, inclusion_rationale, exclusion_rationale };
}

// ---------------------------------------------------------------------------
// GitHub API helpers
// ---------------------------------------------------------------------------

/**
 * Makes an authenticated GitHub API GET request and returns parsed JSON.
 *
 * @param {string} urlPath  — e.g. '/repos/owner/repo/commits/SHA/pulls'
 * @param {string} token    — GitHub personal access token or GITHUB_TOKEN
 * @returns {Promise<unknown>}
 */
function githubGet(urlPath, token) {
  return new Promise((resolve, reject) => {
    const options = {
      hostname: 'api.github.com',
      path: urlPath,
      method: 'GET',
      headers: {
        Authorization: `Bearer ${token}`,
        Accept: 'application/vnd.github+json',
        'X-GitHub-Api-Version': '2022-11-28',
        'User-Agent': 'terraform-provider-elasticstack/build-evidence-manifest',
      },
    };

    const req = https.request(options, (res) => {
      let body = '';
      res.on('data', (chunk) => { body += chunk; });
      res.on('end', () => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          try {
            resolve(JSON.parse(body));
          } catch (e) {
            reject(new Error(`JSON parse error (status ${res.statusCode}): ${e.message}`));
          }
        } else {
          reject(new Error(`GitHub API error ${res.statusCode} for ${urlPath}: ${body.slice(0, 200)}`));
        }
      });
    });

    req.on('error', reject);
    req.end();
  });
}

/**
 * Fetches all pages of a paginated GitHub API endpoint.
 *
 * @param {string} basePath  — e.g. '/repos/owner/repo/pulls/123/files'
 * @param {string} token
 * @returns {Promise<unknown[]>}
 */
async function githubGetAll(basePath, token) {
  const results = [];
  let page = 1;

  while (true) {
    const sep = basePath.includes('?') ? '&' : '?';
    const data = await githubGet(`${basePath}${sep}per_page=100&page=${page}`, token);
    if (!Array.isArray(data) || data.length === 0) break;
    results.push(...data);
    if (data.length < 100) break;
    page++;
  }

  return results;
}

// ---------------------------------------------------------------------------
// Core logic
// ---------------------------------------------------------------------------

/**
 * Returns commit SHAs in the given git compare range.
 *
 * @param {string} compareRange  — e.g. 'v0.14.3..HEAD' or 'HEAD'
 * @param {{ cwd?: string }} [opts]
 * @returns {string[]}
 */
function getCommitSHAs(compareRange, opts = {}) {
  const range = compareRange || 'HEAD';
  try {
    const raw = execSync(`git log --format=%H ${range}`, {
      encoding: 'utf8',
      stdio: ['pipe', 'pipe', 'pipe'],
      cwd: opts.cwd,
    }).trim();
    return raw ? raw.split('\n').map((s) => s.trim()).filter(Boolean) : [];
  } catch (err) {
    process.stderr.write(`Warning: failed to list commits in range "${range}": ${err.message}\n`);
    return [];
  }
}

/**
 * Fetches all merged PRs associated with the commits in the compare range.
 *
 * @param {object} params
 * @param {string} params.owner
 * @param {string} params.repo
 * @param {string} params.compareRange
 * @param {string} params.token
 * @param {{ cwd?: string }} [opts]
 * @returns {Promise<Map<number, object>>}
 */
async function fetchPRsInRange({ owner, repo, compareRange, token }, opts = {}) {
  const commitSHAs = getCommitSHAs(compareRange, opts);
  process.stderr.write(`Found ${commitSHAs.length} commit(s) in range ${compareRange}\n`);

  const prMap = new Map();

  for (const sha of commitSHAs) {
    try {
      const prs = await githubGetAll(
        `/repos/${owner}/${repo}/commits/${sha}/pulls`,
        token
      );
      for (const pr of prs) {
        if (pr.state === 'closed' && pr.merged_at && !prMap.has(pr.number)) {
          prMap.set(pr.number, pr);
        }
      }
    } catch (err) {
      process.stderr.write(`Warning: failed to list PRs for commit ${sha}: ${err.message}\n`);
    }
  }

  process.stderr.write(`Found ${prMap.size} unique merged PR(s) in compare range\n`);
  return prMap;
}

/**
 * Builds the full evidence manifest for the given parameters.
 *
 * @param {object} params
 * @param {string} params.owner
 * @param {string} params.repo
 * @param {string} params.token
 * @param {string} [params.previousTag]
 * @param {string} [params.compareRange]
 * @param {string} [params.mode]          — 'unreleased' | 'release'
 * @param {string} [params.targetVersion] — e.g. '0.14.4' (release mode only)
 * @param {{ cwd?: string }} [opts]
 * @returns {Promise<object>}
 */
async function buildManifest(params, opts = {}) {
  const {
    owner,
    repo,
    token,
    previousTag = '',
    compareRange = 'HEAD',
    mode = 'unreleased',
    targetVersion = '',
  } = params;

  const prMap = await fetchPRsInRange({ owner, repo, compareRange, token }, opts);

  const evidence = [];

  for (const [prNumber, pr] of prMap) {
    let files = [];
    try {
      files = await githubGetAll(
        `/repos/${owner}/${repo}/pulls/${prNumber}/files`,
        token
      );
    } catch (err) {
      process.stderr.write(`Warning: failed to list files for PR #${prNumber}: ${err.message}\n`);
    }

    const { classification, inclusion_rationale, exclusion_rationale } = classifyPR(pr, files);

    evidence.push({
      number: pr.number,
      title: pr.title,
      url: pr.html_url,
      merge_commit_sha: pr.merge_commit_sha,
      author: pr.user?.login ?? 'unknown',
      labels: (pr.labels ?? []).map((l) => l.name),
      touched_files: files.map((f) => f.filename),
      classification,
      inclusion_rationale,
      exclusion_rationale,
    });
  }

  const targetSection =
    mode === 'release'
      ? `## [${targetVersion}] - ${new Date().toISOString().split('T')[0]}`
      : '## [Unreleased]';

  const manifest = {
    generated_at: new Date().toISOString(),
    mode,
    target_section: targetSection,
    target_section_mode: mode,
    target_version: targetVersion,
    previous_tag: previousTag,
    compare_range: compareRange,
    pr_count: evidence.length,
    user_facing_count: evidence.filter((e) => e.classification === 'user-facing').length,
    internal_count: evidence.filter((e) => e.classification === 'internal').length,
    uncertain_count: evidence.filter((e) => e.classification === 'uncertain').length,
    pull_requests: evidence,
  };

  return manifest;
}

// ---------------------------------------------------------------------------
// CLI entry point
// ---------------------------------------------------------------------------

if (require.main === module) {
  (async () => {
    // Parse CLI args (simple --key value pairs)
    const args = process.argv.slice(2);
    const argMap = {};
    for (let i = 0; i < args.length; i += 2) {
      const key = args[i].replace(/^--/, '').replace(/-([a-z])/g, (_, c) => c.toUpperCase());
      argMap[key] = args[i + 1];
    }

    const token = process.env.GITHUB_TOKEN;
    if (!token) {
      process.stderr.write('Error: GITHUB_TOKEN environment variable is required\n');
      process.exit(1);
    }

    const repository = process.env.GITHUB_REPOSITORY || '';
    const [owner, repo] = repository.split('/');
    if (!owner || !repo) {
      process.stderr.write('Error: GITHUB_REPOSITORY must be set as "owner/repo"\n');
      process.exit(1);
    }

    const previousTag = argMap.previousTag || process.env.PREVIOUS_TAG || '';
    const compareRange = argMap.compareRange || process.env.COMPARE_RANGE || 'HEAD';
    const mode = argMap.mode || process.env.MODE || 'unreleased';
    const targetVersion = argMap.targetVersion || process.env.TARGET_VERSION || '';
    const outputPath = argMap.output || process.env.OUTPUT_PATH || '';

    try {
      const manifest = await buildManifest({
        owner,
        repo,
        token,
        previousTag,
        compareRange,
        mode,
        targetVersion,
      });

      const json = JSON.stringify(manifest, null, 2);

      if (outputPath) {
        const dir = path.dirname(outputPath);
        fs.mkdirSync(dir, { recursive: true });
        fs.writeFileSync(outputPath, json, 'utf8');
        process.stderr.write(`Evidence manifest written to ${outputPath} (${manifest.pr_count} PRs)\n`);
      } else {
        process.stdout.write(json + '\n');
      }
    } catch (err) {
      process.stderr.write(`Error building manifest: ${err.message}\n`);
      process.exit(1);
    }
  })();
}

module.exports = { classifyPR, getCommitSHAs, fetchPRsInRange, buildManifest };
