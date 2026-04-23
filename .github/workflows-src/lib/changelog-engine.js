const fs = require('fs');
const { execSync } = require('child_process');

const {
  parseSemverTags,
  buildReleaseContext,
} = require('./changelog-release-context.js');
const {
  parseCommitShas,
  buildEvidenceManifest,
} = require('./changelog-pr-evidence.js');
const { renderChangelogSection } = require('./changelog-renderer.js');

function listSemverTags({ exec = execSync } = {}) {
  const tagsRaw = exec('git tag --list "v[0-9]*.[0-9]*.[0-9]*" --sort=-version:refname', {
    encoding: 'utf8',
    stdio: ['pipe', 'pipe', 'pipe'],
  }).trim();

  return parseSemverTags(tagsRaw);
}

function resolveEngineContext({
  mode,
  targetVersion = '',
  targetBranch,
  tags,
  eventName,
  headBranch = '',
}) {
  if (mode && mode !== 'release' && mode !== 'unreleased') {
    throw new Error(`unsupported changelog mode: ${mode}`);
  }

  if (mode === 'release' && !targetVersion) {
    throw new Error('release mode requires targetVersion');
  }

  if (mode) {
    const previousTagResult = buildReleaseContext({
      eventName: mode === 'release' ? 'pull_request' : 'workflow_dispatch',
      headBranch: mode === 'release' ? `prep-release-${targetVersion}` : '',
      tags,
    });

    return {
      mode,
      targetVersion,
      targetBranch: targetBranch ?? (mode === 'release' ? `prep-release-${targetVersion}` : 'generated-changelog'),
      previousTag: previousTagResult.previousTag,
      excludedTag: previousTagResult.excludedTag,
      excludedCurrentTag: previousTagResult.excludedCurrentTag,
      compareRange: previousTagResult.compareRange,
    };
  }

  return buildReleaseContext({ eventName, headBranch, tags });
}

function listCommitShasInRange(compareRange, { exec = execSync } = {}) {
  const range = compareRange || 'HEAD';
  const raw = exec(`git log --format=%H ${range}`, {
    encoding: 'utf8',
    stdio: ['pipe', 'pipe', 'pipe'],
  }).trim();

  return parseCommitShas(raw);
}

function getLabelNames(pr) {
  return (pr.labels ?? []).map((label) => (typeof label === 'string' ? label : label.name));
}

async function resolveMergedPullRequests({ github, owner, repo, commitShas }) {
  const mergedPullRequestsByNumber = new Map();

  for (const sha of commitShas) {
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
  }

  return Array.from(mergedPullRequestsByNumber.values());
}

function buildRendererPullRequestRecord(pr) {
  return {
    number: pr.number,
    title: pr.title,
    url: pr.html_url,
    merge_commit_sha: pr.merge_commit_sha,
    author: pr.user?.login ?? 'unknown',
    labels: getLabelNames(pr),
    body: pr.body || '',
  };
}

function buildSectionContent({ mode, targetVersion, generatedAt, sectionBody }) {
  const sectionHeader = mode === 'release' && targetVersion
    ? `## [${targetVersion}] - ${String(generatedAt).split('T')[0]}`
    : '## [Unreleased]';

  return {
    sectionHeader,
    newSectionContent: sectionBody ? `${sectionHeader}\n\n${sectionBody}` : `${sectionHeader}`,
  };
}

function findSectionEnd(lines, startIndex) {
  for (let i = startIndex + 1; i < lines.length; i++) {
    if (/^## /.test(lines[i])) {
      return i;
    }
  }

  return lines.length;
}

function rewriteChangelogSection(content, newSectionContent, mode, targetVersion) {
  const lines = content.split('\n');
  let targetStart = -1;

  if (mode === 'unreleased') {
    targetStart = lines.findIndex((line) => /^## \[Unreleased\]/.test(line));
  } else {
    targetStart = lines.findIndex((line) => line.startsWith(`## [${targetVersion}]`));
  }

  if (targetStart === -1) {
    if (mode === 'release') {
      const unreleasedStart = lines.findIndex((line) => /^## \[Unreleased\]/.test(line));
      if (unreleasedStart !== -1) {
        const insertAfter = findSectionEnd(lines, unreleasedStart);
        const before = lines.slice(0, insertAfter);
        const after = lines.slice(insertAfter);
        return [...before, '', newSectionContent, ...after].join('\n');
      }

      return newSectionContent + '\n\n' + content;
    }

    return newSectionContent + '\n\n' + content;
  }

  const sectionEnd = findSectionEnd(lines, targetStart);
  const before = lines.slice(0, targetStart);
  const after = lines.slice(sectionEnd);

  while (before.length > 0 && before[before.length - 1] === '') {
    before.pop();
  }

  const parts = [...before];
  if (parts.length > 0) parts.push('');
  parts.push(newSectionContent);

  let afterStart = 0;
  while (afterStart < after.length && after[afterStart] === '') {
    afterStart++;
  }

  if (afterStart < after.length) {
    parts.push('');
    parts.push(...after.slice(afterStart));
  }

  return parts.join('\n');
}

async function runChangelogEngine({
  github,
  owner,
  repo,
  mode,
  targetVersion = '',
  targetBranch,
  changelogPath = 'CHANGELOG.md',
  generatedAt = new Date().toISOString(),
  fsImpl = fs,
  exec = execSync,
  tags,
}) {
  if (!github) {
    throw new Error('github client is required');
  }
  if (!owner || !repo) {
    throw new Error('owner and repo are required');
  }

  const resolvedTags = tags ?? listSemverTags({ exec });
  const releaseContext = resolveEngineContext({
    mode,
    targetVersion,
    targetBranch,
    tags: resolvedTags,
  });

  const commitShas = listCommitShasInRange(releaseContext.compareRange, { exec });
  const mergedPullRequests = await resolveMergedPullRequests({
    github,
    owner,
    repo,
    commitShas,
  });
  const prRecords = mergedPullRequests.map(buildRendererPullRequestRecord);

  const manifest = buildEvidenceManifest({
    mode: releaseContext.mode,
    targetVersion: releaseContext.targetVersion,
    previousTag: releaseContext.previousTag,
    compareRange: releaseContext.compareRange,
    evidence: prRecords,
    generatedAt,
  });

  const renderResult = renderChangelogSection(prRecords);
  if (!renderResult.success) {
    const error = new Error('Changelog assembly failed');
    error.renderErrors = renderResult.errors;
    throw error;
  }

  let currentChangelog = '';
  try {
    currentChangelog = fsImpl.readFileSync(changelogPath, 'utf8');
  } catch (err) {
    if (err.code !== 'ENOENT') {
      throw err;
    }
  }

  const { sectionHeader, newSectionContent } = buildSectionContent({
    mode: releaseContext.mode,
    targetVersion: releaseContext.targetVersion,
    generatedAt,
    sectionBody: renderResult.sectionBody,
  });

  const updatedChangelog = rewriteChangelogSection(
    currentChangelog,
    newSectionContent,
    releaseContext.mode,
    releaseContext.targetVersion
  );
  fsImpl.writeFileSync(changelogPath, updatedChangelog, 'utf8');

  return {
    mode: releaseContext.mode,
    targetVersion: releaseContext.targetVersion,
    targetBranch: releaseContext.targetBranch,
    previousTag: releaseContext.previousTag,
    compareRange: releaseContext.compareRange,
    excludedTag: releaseContext.excludedTag,
    excludedCurrentTag: releaseContext.excludedCurrentTag,
    changelogPath,
    sectionHeader,
    hasChanges: prRecords.length > 0,
    hasUserFacingChanges: renderResult.included.length > 0,
    manifest,
    pullRequests: prRecords,
    includedPullRequests: renderResult.included,
    excludedPullRequests: renderResult.excluded,
  };
}

module.exports = {
  buildRendererPullRequestRecord,
  buildSectionContent,
  findSectionEnd,
  getLabelNames,
  listCommitShasInRange,
  listSemverTags,
  resolveEngineContext,
  resolveMergedPullRequests,
  rewriteChangelogSection,
  runChangelogEngine,
};
