/**
 * Dependency-injected changelog engine. Used from changelog-engine.js (Node)
 * and from changelog-engine-workflow.js (inlined into GitHub Actions scripts).
 */

/** Semver X.Y.Z without leading `v` */
const TARGET_VERSION_PATTERN = /^\d+\.\d+\.\d+$/;

const TAG_LIST_CMD =
  'git tag --list "v[0-9]*.[0-9]*.[0-9]*" --sort=-version:refname';

/**
 * @param {{
 *   parseSemverTags: Function,
 *   selectPreviousTag: Function,
 *   buildCompareRange: Function,
 *   rewriteChangelogSection: Function,
 *   renderChangelogSection: Function,
 *   execSyncDefault?: Function,
 * }} deps
 */
function createChangelogEngine(deps) {
  const {
    parseSemverTags,
    selectPreviousTag,
    buildCompareRange,
    rewriteChangelogSection,
    renderChangelogSection,
    execSyncDefault,
  } = deps;

  const fallbackExec =
    execSyncDefault ?? require('child_process').execSync;

  function validateModeAndTargetVersion(mode, targetVersion, core) {
    if (mode !== 'unreleased' && mode !== 'release') {
      const msg = `Invalid changelog mode: "${mode}". Must be 'unreleased' or 'release'.`;
      if (core && typeof core.setFailed === 'function') core.setFailed(msg);
      throw new Error(msg);
    }
    if (mode === 'release') {
      if (!targetVersion || !TARGET_VERSION_PATTERN.test(targetVersion)) {
        const msg =
          'Release mode requires targetVersion: a non-empty semver string X.Y.Z without a leading "v".';
        if (core && typeof core.setFailed === 'function') core.setFailed(msg);
        throw new Error(msg);
      }
    }
  }

  function formatAssemblyFailureMessage(errors) {
    const errorMessages = errors.map((e) => `  - ${e.reason}`).join('\n');
    return (
      `Changelog assembly failed. The following pull requests are missing a required ## Changelog section or Summary field:\n${errorMessages}\n\n` +
      'Each merged PR must either:\n' +
      "  1. Have a '## Changelog' section with 'Customer impact' and (when not 'none') a 'Summary' field, OR\n" +
      "  2. Be labeled 'no-changelog'"
    );
  }

  function parseCommitShas(raw = '') {
    return raw
      .split('\n')
      .map((sha) => sha.trim())
      .filter(Boolean);
  }

  function getLabelNames(pr) {
    return (pr.labels ?? []).map((label) => label.name);
  }

  function resolveChangelogCompareContext({ mode, targetVersion, exec, core }) {
    validateModeAndTargetVersion(mode, targetVersion, core);

    let tags = [];
    try {
      const tagsRaw = exec(TAG_LIST_CMD, {
        encoding: 'utf8',
        stdio: ['pipe', 'pipe', 'pipe'],
      }).trim();
      tags = parseSemverTags(tagsRaw);
    } catch (err) {
      if (core && typeof core.warning === 'function') {
        core.warning(`Failed to list git tags: ${err.message}`);
      }
    }

    const {
      previousTag,
      excludedTag,
      excludedCurrentTag,
    } = selectPreviousTag({ tags, mode, targetVersion });
    const compareRange = buildCompareRange(previousTag);
    const targetBranch =
      mode === 'unreleased' ? 'generated-changelog' : `prep-release-${targetVersion}`;

    return {
      previousTag,
      compareRange,
      targetBranch,
      excludedTag,
      excludedCurrentTag,
      tags,
    };
  }

  async function gatherMergedPRRecordsForRange({
    github,
    owner,
    repo,
    compareRange,
    exec,
    core,
  }) {
    const range = compareRange || 'HEAD';
    let commitSHAs = [];
    try {
      const raw = exec(`git log --format=%H ${range}`, {
        encoding: 'utf8',
        stdio: ['pipe', 'pipe', 'pipe'],
      }).trim();
      commitSHAs = parseCommitShas(raw);
      if (core && typeof core.info === 'function') {
        core.info(`Found ${commitSHAs.length} commit(s) in range ${range}`);
      }
    } catch (err) {
      if (core && typeof core.warning === 'function') {
        core.warning(`Failed to list commits in range: ${err.message}`);
      }
    }

    const mergedPullRequestsByNumber = new Map();

    for (const sha of commitSHAs) {
      try {
        const { data: prs } = await github.rest.repos.listPullRequestsAssociatedWithCommit({
          owner,
          repo,
          commit_sha: sha,
        });
        for (const pr of prs) {
          if (pr.state === 'closed' && pr.merged_at && !mergedPullRequestsByNumber.has(pr.number)) {
            mergedPullRequestsByNumber.set(pr.number, pr);
          }
        }
      } catch (err) {
        if (core && typeof core.warning === 'function') {
          core.warning(`Failed to list PRs for commit ${sha}: ${err.message}`);
        }
      }
    }

    const mergedPullRequests = Array.from(mergedPullRequestsByNumber.values());
    if (core && typeof core.info === 'function') {
      core.info(`Found ${mergedPullRequests.length} unique merged PR(s) in compare range`);
    }

    return mergedPullRequests.map((pr) => ({
      number: pr.number,
      title: pr.title,
      url: pr.html_url,
      merge_commit_sha: pr.merge_commit_sha,
      author: pr.user?.login ?? 'unknown',
      labels: getLabelNames(pr),
      body: pr.body || '',
    }));
  }

  function runChangelogRenderAndWrite({ core, prRecords, mode, targetVersion, changelogPath, fs }) {
    validateModeAndTargetVersion(mode, targetVersion, core);

    const result = renderChangelogSection(prRecords);

    if (!result.success) {
      const msg = formatAssemblyFailureMessage(result.errors);
      if (core && typeof core.setFailed === 'function') core.setFailed(msg);
      const err = new Error(msg);
      err.assemblyErrors = result.errors;
      throw err;
    }

    let currentChangelog = '';
    try {
      currentChangelog = fs.readFileSync(changelogPath, 'utf8');
    } catch (err) {
      if (core && typeof core.warning === 'function') {
        core.warning(`Could not read ${changelogPath}: ${err.message}. Will create a new file.`);
      }
    }

    const today = new Date().toISOString().split('T')[0];
    let sectionHeader;
    if (mode === 'release' && targetVersion) {
      sectionHeader = `## [${targetVersion}] - ${today}`;
    } else {
      sectionHeader = '## [Unreleased]';
    }

    const sectionBody = result.sectionBody;
    const newSectionContent = sectionBody
      ? `${sectionHeader}\n\n${sectionBody}`
      : `${sectionHeader}`;

    const updatedChangelog = rewriteChangelogSection(
      currentChangelog,
      newSectionContent,
      mode,
      targetVersion
    );

    try {
      fs.writeFileSync(changelogPath, updatedChangelog, 'utf8');
      if (core && typeof core.info === 'function') {
        core.info(`CHANGELOG.md updated with section: ${sectionHeader}`);
      }
    } catch (err) {
      const msg = `Failed to write ${changelogPath}: ${err.message}`;
      if (core && typeof core.setFailed === 'function') core.setFailed(msg);
      throw new Error(msg);
    }

    const hasPRs = prRecords.length > 0;
    // hasUserFacingChanges reflects whether the rendered section body has any
    // content the changelog should publish. This is broader than `included`
    // (PRs that contributed change bullets) because PRs marked
    // `Customer impact: none` can still contribute a `### Breaking changes`
    // block, which is user-facing and must be committed/pushed.
    const hasUserFacingChanges =
      typeof sectionBody === 'string' && sectionBody.trim().length > 0;

    return {
      sectionHeader,
      hasPRs,
      hasUserFacingChanges,
      included: result.included,
      excluded: result.excluded,
      errors: [],
    };
  }

  async function runChangelogEngine({
    github,
    core,
    mode,
    targetVersion = '',
    owner,
    repo,
    changelogPath = 'CHANGELOG.md',
    exec = fallbackExec,
    fs = require('fs'),
  }) {
    validateModeAndTargetVersion(mode, targetVersion, core);

    const {
      previousTag,
      compareRange,
      targetBranch,
      excludedTag,
      excludedCurrentTag,
      tags,
    } = resolveChangelogCompareContext({ mode, targetVersion, exec, core });

    if (mode === 'release' && core && typeof core.info === 'function') {
      core.info(`Release mode: version=${targetVersion}`);
    }
    if (excludedCurrentTag && core && typeof core.info === 'function') {
      core.info(`Excluded current release tag ${excludedTag} from previous tag candidates`);
    }
    if (previousTag && core && typeof core.info === 'function') {
      core.info(`Resolved previous tag: ${previousTag}`);
    } else if (tags.length === 0 && core && typeof core.warning === 'function') {
      core.warning('No semver release tags found; compare range will cover full history');
    }

    const prRecords = await gatherMergedPRRecordsForRange({
      github,
      owner,
      repo,
      compareRange,
      exec,
      core,
    });

    const renderOutcome = runChangelogRenderAndWrite({
      core,
      prRecords,
      mode,
      targetVersion,
      changelogPath,
      fs,
    });

    return {
      mode,
      targetVersion: mode === 'release' ? targetVersion : '',
      previousTag,
      compareRange,
      targetBranch,
      sectionHeader: renderOutcome.sectionHeader,
      hasPRs: renderOutcome.hasPRs,
      hasUserFacingChanges: renderOutcome.hasUserFacingChanges,
      included: renderOutcome.included,
      excluded: renderOutcome.excluded,
      errors: [],
    };
  }

  return {
    TARGET_VERSION_PATTERN,
    validateModeAndTargetVersion,
    resolveChangelogCompareContext,
    gatherMergedPRRecordsForRange,
    formatAssemblyFailureMessage,
    runChangelogRenderAndWrite,
    runChangelogEngine,
  };
}

if (typeof module !== 'undefined') {
  module.exports = {
    TARGET_VERSION_PATTERN,
    createChangelogEngine,
  };
}
