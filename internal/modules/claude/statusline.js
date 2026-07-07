#!/usr/bin/env node
/**
 * Claude Code status line — hiển thị: model · context usage · token · chi phí ($)
 * Nhận JSON qua stdin, đọc thêm transcript để tính token/context.
 * Docs: https://docs.claude.com/en/docs/claude-code/statusline
 */
'use strict';
const fs = require('fs');

// ---------- helpers ----------
const c = {
  reset: '\x1b[0m', dim: '\x1b[2m', bold: '\x1b[1m',
  cyan: '\x1b[36m', green: '\x1b[32m', yellow: '\x1b[33m',
  red: '\x1b[31m', magenta: '\x1b[35m', gray: '\x1b[90m', white: '\x1b[37m',
};
const sep = ` ${c.gray}│${c.reset} `;

function fmt(n) {
  n = Number(n) || 0;
  if (n >= 1e6) return (n / 1e6).toFixed(n >= 1e7 ? 0 : 2).replace(/\.?0+$/, '') + 'M';
  if (n >= 1e3) return (n / 1e3).toFixed(n >= 1e5 ? 0 : 1).replace(/\.0$/, '') + 'k';
  return String(Math.round(n));
}

function bar(frac, width = 10) {
  frac = Math.max(0, Math.min(1, frac || 0));
  const filled = Math.round(frac * width);
  return '█'.repeat(filled) + '░'.repeat(width - filled);
}

// ---------- read stdin ----------
let raw = '';
try { raw = fs.readFileSync(0, 'utf8'); } catch (_) {}
let input = {};
try { input = JSON.parse(raw || '{}'); } catch (_) {}

const modelName = input.model?.display_name || input.model?.id || 'Claude';
const modelId = `${input.model?.id || ''} ${input.model?.display_name || ''}`;
const costUsd = input.cost?.total_cost_usd ?? 0;
const linesAdded = input.cost?.total_lines_added ?? 0;
const linesRemoved = input.cost?.total_lines_removed ?? 0;

// context window limit: model 1M -> 1,000,000; ngược lại 200,000. Override bằng env.
const is1M = /1m|1M context|\[1m\]/i.test(modelId);
const CONTEXT_LIMIT = process.env.CLAUDE_CONTEXT_LIMIT
  ? Number(process.env.CLAUDE_CONTEXT_LIMIT)
  : (is1M ? 1_000_000 : 200_000);

// ---------- parse transcript ----------
let ctxTokens = 0;      // token đang nằm trong context (message gần nhất)
let totalTokens = 0;    // tổng token đã xử lý cả session (gồm cache-read)
try {
  const p = input.transcript_path;
  if (p && fs.existsSync(p)) {
    const lines = fs.readFileSync(p, 'utf8').split('\n');
    let lastUsage = null;
    for (const line of lines) {
      if (!line) continue;
      let o;
      try { o = JSON.parse(line); } catch (_) { continue; }
      if (o.type !== 'assistant' || o.isSidechain) continue;
      const u = o.message?.usage;
      if (!u) continue;
      lastUsage = u;
      totalTokens += (u.input_tokens || 0) + (u.cache_creation_input_tokens || 0)
                   + (u.cache_read_input_tokens || 0) + (u.output_tokens || 0);
    }
    if (lastUsage) {
      ctxTokens = (lastUsage.input_tokens || 0)
                + (lastUsage.cache_creation_input_tokens || 0)
                + (lastUsage.cache_read_input_tokens || 0);
    }
  }
} catch (_) {}

if (input.exceeds_200k_tokens && ctxTokens < 200_000) ctxTokens = 200_001;

// ---------- build segments ----------
const parts = [];

// model
parts.push(`${c.cyan}${c.bold}⧉ ${modelName}${c.reset}`);

// context bar
const frac = ctxTokens / CONTEXT_LIMIT;
const pct = frac * 100;
const ctxColor = pct >= 85 ? c.red : pct >= 65 ? c.yellow : c.green;
parts.push(
  `${ctxColor}${bar(frac)}${c.reset} ` +
  `${ctxColor}${fmt(ctxTokens)}${c.gray}/${fmt(CONTEXT_LIMIT)}${c.reset} ` +
  `${c.dim}(${pct.toFixed(pct < 10 ? 1 : 0)}%)${c.reset}`
);

// tổng token session
parts.push(`${c.magenta}Σ ${fmt(totalTokens)} tok${c.reset}`);

// chi phí
const costColor = costUsd >= 5 ? c.red : costUsd >= 1 ? c.yellow : c.green;
parts.push(`${costColor}$${costUsd.toFixed(costUsd < 1 ? 3 : 2)}${c.reset}`);

// lines changed (nếu có)
if (linesAdded || linesRemoved) {
  parts.push(`${c.green}+${linesAdded}${c.reset} ${c.red}-${linesRemoved}${c.reset}`);
}

process.stdout.write(parts.join(sep));
