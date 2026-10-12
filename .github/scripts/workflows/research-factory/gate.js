const SCORE_THRESHOLD = 85;
const STABILITY_WINDOW = 2;
const MAX_ROUNDS = 5;

const READY = 'ready-for-change-factory';
const NEEDS_HUMAN = 'research-needs-human';
const CHECKLIST_ITEMS = ['grounded', 'mapped', 'compatible', 'versioned', 'testable', 'idiomatic'];
const CRITIC_STATUSES = ['ok', 'unavailable', 'error'];

const KEBAB_CASE = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const CONFIDENCES = ['high', 'medium', 'low'];
const SCOPES = ['small', 'medium', 'large', 'unknown'];
const REFERENCE_TYPES = ['elastic-docs', 'repo-path', 'issue', 'pr', 'external'];
const FENCE_LINE = /^ {0,3}(`{3,}|~{3,})([^`~\s]*)?(.*)$/;
const DETAILS_TAG = /<details(?:\s[^>]*)?>|<\/details\s*>/gi;
const SUMMARY_TEXT = /<summary(?:\s[^>]*)?>(.*?)<\/summary\s*>/i;
const OUTCOME_LINE = /^\*\*Outcome:\*\* .*$/m;
const QUALITY_GATE_HEADING = /^### Quality gate[^\S\n]*$/m;
const REFERENCES_HEADING = /^### References[^\S\n]*$/m;

function isObject(value) {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function fenceOf(line) {
  const m = FENCE_LINE.exec(line.replace(/\r?\n$/, ''));
  if (!m) {
    return null;
  }
  const info = `${m[2] || ''}${m[3] || ''}`;
  if (m[1][0] === '`' && info.includes('`')) {
    return null;
  }
  return { char: m[1][0], length: m[1].length, info: info.trim() };
}

// Locates the content range of the single json fence inside the single "Pipeline metadata"
// <details> envelope. Fence-aware: tags inside code fences are ignored.
function locateMetadata(body) {
  const text = String(body ?? '');
  const stack = [];
  const envelopes = [];
  let fence = null;

  for (const match of text.matchAll(/[^\n]*\n|[^\n]+/g)) {
    const line = match[0];
    const index = match.index;
    const parsed = fenceOf(line);

    if (fence) {
      if (parsed && parsed.char === fence.char && parsed.length >= fence.length && parsed.info === '') {
        if (fence.json && fence.envelope) {
          const end = text[index - 1] === '\n' ? index - 1 : index;
          fence.envelope.fences.push({ start: fence.start, end: Math.max(fence.start, text[end - 1] === '\r' ? end - 1 : end) });
        }
        fence = null;
      }
      continue;
    }
    if (parsed) {
      fence = {
        ...parsed,
        json: parsed.info.toLowerCase() === 'json',
        start: index + line.length,
        envelope: [...stack].reverse().find((entry) => entry.meta),
      };
      continue;
    }

    const summary = SUMMARY_TEXT.exec(line);
    for (const tag of line.matchAll(DETAILS_TAG)) {
      if (tag[0][1] === '/') {
        const entry = stack.pop();
        if (entry && entry.meta) {
          envelopes.push(entry);
        }
      } else {
        stack.push({ meta: false, fences: [], summarySeen: false });
      }
    }
    if (summary && stack.length > 0) {
      const top = stack[stack.length - 1];
      if (!top.summarySeen) {
        top.summarySeen = true;
        top.meta = /pipeline metadata/i.test(summary[1]);
      }
    }
  }

  if (envelopes.length === 0) {
    return { error: 'Pipeline metadata <details> envelope not found in comment' };
  }
  if (envelopes.length > 1) {
    return { error: 'multiple Pipeline metadata <details> envelopes found in comment' };
  }
  const fences = envelopes[0].fences;
  if (fences.length !== 1) {
    return { error: `Pipeline metadata envelope must contain exactly one json block, found ${fences.length}` };
  }
  return { start: fences[0].start, end: fences[0].end, text: text.slice(fences[0].start, fences[0].end) };
}

function extractMetadata(body) {
  const located = locateMetadata(body);
  if (located.error) {
    return { error: located.error };
  }
  try {
    return { metadata: JSON.parse(located.text) };
  } catch (err) {
    return { error: `Pipeline metadata JSON could not be parsed: ${err.message}` };
  }
}

function isValidReference(ref) {
  const location = isObject(ref) ? ref.url ?? ref.path : undefined;
  return (
    isObject(ref) &&
    REFERENCE_TYPES.includes(ref.type) &&
    typeof location === 'string' &&
    location.trim() !== ''
  );
}

function validateGate(meta) {
  if (!isObject(meta)) {
    return ['metadata is not a JSON object'];
  }
  const errors = [];
  if (meta.schema_version !== '1.1') {
    errors.push('schema_version must be "1.1"');
  }
  const rec = meta.recommendation;
  if (
    !isObject(rec) ||
    typeof rec.spine !== 'string' ||
    !KEBAB_CASE.test(rec.spine) ||
    !Number.isInteger(rec.approach_index) ||
    rec.approach_index < 0
  ) {
    errors.push('recommendation must have a kebab-case spine and a non-negative integer approach_index');
  } else if (rec.confidence !== undefined && !CONFIDENCES.includes(rec.confidence)) {
    errors.push(`recommendation.confidence must be one of ${CONFIDENCES.join(', ')}`);
  }
  if (meta.estimated_scope !== undefined && !SCOPES.includes(meta.estimated_scope)) {
    errors.push(`estimated_scope must be one of ${SCOPES.join(', ')}`);
  }
  if (
    meta.affected_capabilities !== undefined &&
    !(
      Array.isArray(meta.affected_capabilities) &&
      meta.affected_capabilities.every((c) => typeof c === 'string' && KEBAB_CASE.test(c))
    )
  ) {
    errors.push('affected_capabilities must be an array of kebab-case strings');
  }
  if (meta.references !== undefined && !(Array.isArray(meta.references) && meta.references.every(isValidReference))) {
    errors.push(`references items must have a type of ${REFERENCE_TYPES.join(', ')} and a non-empty url or path`);
  }
  if (meta.open_questions !== undefined) {
    const valid =
      Array.isArray(meta.open_questions) &&
      meta.open_questions.every(
        (q) => isObject(q) && typeof q.id === 'string' && typeof q.text === 'string' && typeof q.blocking === 'boolean',
      );
    if (!valid) {
      errors.push('open_questions items must have string id, string text, and boolean blocking');
    }
  }

  const g = meta.gate;
  if (!isObject(g)) {
    errors.push('gate object is required');
    return errors;
  }
  if (g.outcome !== READY && g.outcome !== NEEDS_HUMAN) {
    errors.push(`gate.outcome must be ${READY} or ${NEEDS_HUMAN}`);
  }
  if (!isObject(g.checklist) || !CHECKLIST_ITEMS.every((item) => typeof g.checklist[item] === 'boolean')) {
    errors.push(`gate.checklist must have boolean values for ${CHECKLIST_ITEMS.join(', ')}`);
  }
  const isScore = (n) => typeof n === 'number' && n >= 0 && n <= 100;
  if (g.score !== null && !isScore(g.score)) {
    errors.push('gate.score must be a number from 0 to 100 or null');
  }
  if (!Array.isArray(g.scores) || !g.scores.every(isScore)) {
    errors.push('gate.scores must be an array of numbers from 0 to 100');
  }
  if (typeof g.converged !== 'boolean') {
    errors.push('gate.converged must be a boolean');
  }
  if (!Number.isInteger(g.rounds) || g.rounds < 0 || g.rounds > MAX_ROUNDS) {
    errors.push(`gate.rounds must be an integer from 0 to ${MAX_ROUNDS}`);
  } else if (Array.isArray(g.scores) && g.rounds !== g.scores.length) {
    errors.push('gate.rounds must equal the length of gate.scores');
  }
  if (!Array.isArray(g.outstanding_feedback) || !g.outstanding_feedback.every((f) => typeof f === 'string')) {
    errors.push('gate.outstanding_feedback must be an array of strings');
  }
  if (typeof g.author_model !== 'string' || g.author_model.trim() === '') {
    errors.push('gate.author_model must be a non-empty string');
  }
  if (
    !isObject(g.critic) ||
    typeof g.critic.model !== 'string' ||
    g.critic.model.trim() === '' ||
    !CRITIC_STATUSES.includes(g.critic.status)
  ) {
    errors.push(`gate.critic must have a model and a status of ${CRITIC_STATUSES.join(', ')}`);
  }
  return errors;
}

function isConverged(scores, outstandingFeedback) {
  if (scores.length === 0) {
    return false;
  }
  const atThreshold = (s) => s >= SCORE_THRESHOLD;
  if (!atThreshold(scores[scores.length - 1])) {
    return false;
  }
  const window = scores.slice(-STABILITY_WINDOW);
  return (window.length === STABILITY_WINDOW && window.every(atThreshold)) || outstandingFeedback.length === 0;
}

function safeText(value, max = 200) {
  return String(value).replace(/[\r\n`]+/g, ' ').trim().slice(0, max);
}

function safeId(value) {
  return String(value).replace(/[^A-Za-z0-9_-]/g, '').slice(0, 40);
}

function deriveOutcome(meta, extraction = {}) {
  const result = deriveRawOutcome(meta, extraction);
  return { ...result, reasons: result.reasons.map((r) => safeText(r)) };
}

function deriveRawOutcome(meta, extraction) {
  if (meta === undefined) {
    return { label: NEEDS_HUMAN, reasons: [extraction.error || 'metadata missing'], overridden: true };
  }
  const errors = validateGate(meta);
  if (errors.length > 0) {
    const reported = isObject(meta) && isObject(meta.gate) ? meta.gate.outcome : undefined;
    return {
      label: NEEDS_HUMAN,
      reasons: errors.map((e) => `invalid metadata: ${e}`),
      overridden: reported !== NEEDS_HUMAN,
    };
  }

  const g = meta.gate;
  const reasons = [];
  for (const item of CHECKLIST_ITEMS) {
    if (!g.checklist[item]) {
      reasons.push(`checklist item ${item} failed`);
    }
  }
  if (g.rounds < 1) {
    reasons.push('no critique round completed');
  } else if (g.score !== g.scores[g.scores.length - 1]) {
    reasons.push('gate.score does not match the last value in gate.scores');
  }
  if (!isConverged(g.scores, g.outstanding_feedback)) {
    reasons.push(`not converged: needs a final score of at least ${SCORE_THRESHOLD} with ${STABILITY_WINDOW} consecutive qualifying rounds or no actionable feedback`);
  }
  if (g.critic.status !== 'ok') {
    reasons.push(`critic status is ${g.critic.status}`);
  }
  const blocking = (meta.open_questions || []).filter((q) => q.blocking);
  if (blocking.length > 0) {
    reasons.push(`blocking open question(s): ${blocking.map((q) => safeId(q.id)).join(', ')}`);
  }

  const label = reasons.length === 0 ? READY : NEEDS_HUMAN;
  return {
    label,
    reasons: reasons.length === 0 ? ['all gate conditions satisfied'] : reasons,
    overridden: g.outcome !== label,
    reported: g.outcome,
  };
}

function sectionOutcome(body) {
  const heading = QUALITY_GATE_HEADING.exec(body);
  if (!heading) {
    return undefined;
  }
  const rest = body.slice(heading.index + heading[0].length);
  const next = /^#{1,3} /m.exec(rest);
  const section = next ? rest.slice(0, next.index) : rest;
  return /^\*\*Outcome:\*\*[^\S\n]*`?([^`\s]*)`?/m.exec(section)?.[1];
}

function evaluateBody(body) {
  const extraction = extractMetadata(body);
  const result = deriveOutcome(extraction.metadata, extraction);
  const shown = sectionOutcome(String(body ?? ''));
  if (shown !== undefined && shown !== result.label && !result.overridden) {
    return { ...result, overridden: true, reported: safeText(shown, 60) };
  }
  return result;
}

function outcomeLine(result) {
  return `**Outcome:** \`${result.label}\` - ${result.reasons.join('; ')}`;
}

function overrideNote(result) {
  const reported = result.reported ? ` \`${result.reported}\`` : '';
  return `> Note: the reported outcome${reported} was overridden by the gate rule: ${result.reasons.join('; ')}.`;
}

function correctSection(body, result) {
  const lines = `${outcomeLine(result)}\n\n${overrideNote(result)}`;
  const heading = QUALITY_GATE_HEADING.exec(body);
  if (!heading) {
    const section = `### Quality gate\n\n${lines}\n\n`;
    const refs = REFERENCES_HEADING.exec(body);
    if (refs) {
      return body.slice(0, refs.index) + section + body.slice(refs.index);
    }
    return `${body.replace(/\s*$/, '')}\n\n${section}`;
  }

  const sectionStart = heading.index + heading[0].length;
  const rest = body.slice(sectionStart);
  const nextHeading = /^#{1,3} /m.exec(rest);
  const sectionEnd = nextHeading ? sectionStart + nextHeading.index : body.length;
  let section = body.slice(sectionStart, sectionEnd);
  if (OUTCOME_LINE.test(section)) {
    section = section.replace(OUTCOME_LINE, () => `${outcomeLine(result)}\n\n${overrideNote(result)}`);
  } else {
    section = `\n\n${lines}${section.startsWith('\n') ? section : `\n\n${section}`}`;
  }
  return body.slice(0, sectionStart) + section + body.slice(sectionEnd);
}

function rewriteReportedOutcome(body, label) {
  const located = locateMetadata(body);
  if (located.error) {
    return body;
  }
  let meta;
  try {
    meta = JSON.parse(located.text);
  } catch {
    return body;
  }
  if (!isObject(meta) || !isObject(meta.gate)) {
    return body;
  }
  meta.gate.outcome = label;
  return body.slice(0, located.start) + JSON.stringify(meta, null, 2) + body.slice(located.end);
}

function applyOverride(body, result) {
  if (!result.overridden) {
    return body;
  }
  return rewriteReportedOutcome(correctSection(body, result), result.label);
}

if (typeof module !== 'undefined') {
  module.exports = {
    SCORE_THRESHOLD,
    STABILITY_WINDOW,
    MAX_ROUNDS,
    READY,
    NEEDS_HUMAN,
    extractMetadata,
    validateGate,
    isConverged,
    deriveOutcome,
    evaluateBody,
    applyOverride,
  };
}
