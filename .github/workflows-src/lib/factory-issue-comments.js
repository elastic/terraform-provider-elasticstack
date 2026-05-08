/**
 * Shared comment helpers for research-factory issue intake workflows.
 */

/**
 * Fetches human-authored comments for an issue, paginated, with bot filtering and a hard cap.
 *
 * @param {{ github: object, owner: string, repo: string, issueNumber: number }} params
 * @returns {Promise<{ comments: Array<{author: string, createdAt: string, body: string}>, truncated: boolean }>}
 */
async function factoryFetchIssueComments({ github, owner, repo, issueNumber }) {
  const MAX_COMMENTS = 200;
  const allComments = await github.paginate(github.rest.issues.listComments, {
    owner,
    repo,
    issue_number: issueNumber,
    per_page: 100,
  });

  const humanComments = [];
  let truncated = false;
  for (const comment of allComments) {
    if (comment.user?.login?.endsWith('[bot]')) {
      continue;
    }
    if (humanComments.length >= MAX_COMMENTS) {
      truncated = true;
      break;
    }
    humanComments.push({
      author: comment.user?.login ?? '',
      createdAt: comment.created_at ?? '',
      body: comment.body ?? '',
    });
  }

  return {
    comments: humanComments,
    truncated,
  };
}

const COMMENT_CONTEXT_BUDGET = 50_000;
/** Overhead reserved for truncation markers appended after the loop. */
const COMMENT_CONTEXT_MARKER_OVERHEAD = 200;

/**
 * Serializes captured issue comments into a deterministic markdown string for agent prompts.
 *
 * @param {{ comments: Array<{author: string, createdAt: string, body: string}>, truncated: boolean }} params
 * @returns {string}
 */
function serializeIssueComments({ comments, truncated }) {
  if (!Array.isArray(comments) || comments.length === 0) {
    return '';
  }

  const bodyBudget = COMMENT_CONTEXT_BUDGET - COMMENT_CONTEXT_MARKER_OVERHEAD;
  let result = '';
  let includedCount = 0;

  for (const comment of comments) {
    const header = `**@${comment.author || ''}** (${comment.createdAt || ''}):\n\n`;
    const body = comment.body || '';
    const footer = '\n\n---\n';
    const available = bodyBudget - result.length;

    if (available <= 0) {
      break;
    }

    const frameLength = header.length + footer.length;
    if (frameLength > available) {
      break;
    }
    const fullBlock = header + body + footer;
    if (fullBlock.length <= available) {
      result += fullBlock;
    } else {
      // Truncate this comment's body so the output stays within budget
      const truncatedBody = body.slice(0, available - frameLength);
      result += header + truncatedBody + footer;
    }
    includedCount++;
  }

  const remaining = comments.length - includedCount;
  if (remaining > 0) {
    result += `[... ${remaining} more comments truncated for context budget]\n`;
  }

  if (truncated) {
    result += '[... comment history truncated at 200 comments]\n';
  }

  return result;
}

if (typeof module !== 'undefined') {
  module.exports = {
    factoryFetchIssueComments,
    serializeIssueComments,
    COMMENT_CONTEXT_BUDGET,
  };
}
