// Lint only what a change touches: run ESLint on the frontend files changed since
// <base> and fail only on errors reported on added or modified lines. Existing
// code has a large lint backlog; this keeps new code clean without blocking on it.
//
//   node scripts/lint-changed.mjs <base-sha>
//
// CI passes the PR base SHA, or the previous commit for a push.
import { execFileSync } from 'node:child_process';
import { existsSync } from 'node:fs';
import { ESLint } from 'eslint';

const LINTABLE = /\.(ts|tsx|js|jsx|mjs|cjs)$/;

function git(...args) {
  return execFileSync('git', args, { encoding: 'utf8', maxBuffer: 64 * 1024 * 1024 });
}

function resolveBase(arg) {
  const candidates = [arg, 'HEAD~1'].filter((ref) => ref && !/^0+$/.test(ref));
  for (const ref of candidates) {
    try {
      return git('rev-parse', '--verify', '--quiet', `${ref}^{commit}`).trim();
    } catch {
      // not available in this checkout; try the next candidate
    }
  }
  return null;
}

// Map of file -> Set of added/modified line numbers, from a zero-context diff.
// -C detects moves and copies, so code moved into a new file isn't counted as new.
function changedLines(base) {
  const diff = git('diff', '-C', '--relative', '--unified=0', '--diff-filter=ACMR', `${base}...HEAD`, '--', '.');
  const lines = new Map();
  let file = null;
  for (const line of diff.split('\n')) {
    if (line.startsWith('+++ ')) {
      const path = line.slice(4).replace(/^b\//, '');
      file = path === '/dev/null' || !LINTABLE.test(path) ? null : path;
      if (file && !lines.has(file)) lines.set(file, new Set());
    } else if (file && line.startsWith('@@')) {
      const m = /\+(\d+)(?:,(\d+))?/.exec(line);
      const start = Number(m[1]);
      const count = m[2] === undefined ? 1 : Number(m[2]);
      for (let n = start; n < start + count; n++) lines.get(file).add(n);
    }
  }
  return lines;
}

async function main() {
  const base = resolveBase(process.argv[2]);
  if (!base) {
    console.log('lint-changed: no base commit available, nothing to compare against.');
    return;
  }

  const changed = changedLines(base);
  const files = [...changed.keys()].filter((f) => existsSync(f));
  if (files.length === 0) {
    console.log('lint-changed: no changed frontend source files.');
    return;
  }

  const eslint = new ESLint();
  const lintable = [];
  for (const f of files) {
    if (!(await eslint.isPathIgnored(f))) lintable.push(f);
  }
  const results = await eslint.lintFiles(lintable);

  let errors = 0;
  let warnings = 0;
  for (const result of results) {
    const file = result.filePath.slice(process.cwd().length + 1);
    const touched = changed.get(file) ?? new Set();
    for (const msg of result.messages) {
      // Parse errors have no rule and may sit on unchanged lines; always report them
      if (msg.ruleId && !touched.has(msg.line)) continue;
      const level = msg.severity === 2 ? 'error' : 'warning';
      if (msg.severity === 2) errors++;
      else warnings++;
      console.log(`${file}:${msg.line}:${msg.column}  ${level}  ${msg.message}  ${msg.ruleId ?? ''}`);
    }
  }

  console.log(`lint-changed: ${lintable.length} file(s) checked against ${base.slice(0, 7)}; ` +
    `${errors} error(s), ${warnings} warning(s) on changed lines.`);
  if (errors > 0) process.exitCode = 1;
}

main().catch((err) => {
  console.error(err);
  process.exitCode = 1;
});
