//include: ../../lib/kibana-spec-impact-gate.js

const { execFileSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const ws = process.env.GITHUB_WORKSPACE;
const sha = process.env.GITHUB_SHA || 'HEAD';

const defaultMem = '/tmp/gh-aw/repo-memory/kibana-spec-impact/memory/kibana-spec-impact/kibana-spec-impact.json';
let mem = process.env.KIBANA_SPEC_IMPACT_MEMORY;
if (!mem && fs.existsSync(defaultMem)) {
  mem = defaultMem;
} else if (!mem) {
  mem = path.join(process.env.RUNNER_TEMP, 'kibana-spec-impact-memory.json');
}

const goEnv = {
  ...process.env,
  TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL: 'true',
};

if (!fs.existsSync(mem)) {
  execFileSync('go', ['run', './scripts/kibana-spec-impact', 'memory-bootstrap', '--memory', mem], {
    cwd: ws,
    env: goEnv,
    stdio: 'inherit',
  });
}

const reportJson = execFileSync(
  'go',
  ['run', './scripts/kibana-spec-impact', 'report', '--repo', '.', '--memory', mem, '--target', sha],
  { cwd: ws, encoding: 'utf8', env: goEnv },
);

const r = JSON.parse(reportJson);
const g = kibanaSpecImpactGate(r);

core.setOutput('should_run', g.shouldRun ? 'true' : 'false');
core.setOutput('issue_cap', String(g.issueCap));
core.setOutput('high_confidence_count', String(g.highConfidenceCount));
core.setOutput('gate_reason', g.gate_reason);

fs.writeFileSync(path.join(ws, 'kibana-spec-impact-report.json'), JSON.stringify(r, null, 2) + '\n');
core.info(`kibana-spec-impact gate: ${g.gate_reason}`);
