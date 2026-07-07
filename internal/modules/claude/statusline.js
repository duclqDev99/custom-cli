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

// ---------- weekly / monthly usage FALLBACK (from Claude Code's stats cache) ----------
// Preferred source is the real subscription quota Claude Code passes on stdin as
// `rate_limits` (see below) — identical to what /usage shows. These helpers are a
// FALLBACK for when rate_limits is absent (before the first API response of a
// session, or API-key users): a local estimate from ~/.claude/stats-cache.json's
// per-day token totals. Windows roll and include today: W = 7 days, M = 30 days.
// The cache is recomputed periodically, so it can lag ~1–2 days.

// parseBudget reads a token budget with an optional k/M suffix ("20M", "500k").
function parseBudget(v) {
  if (!v) return 0;
  const m = String(v).trim().match(/^([\d.]+)\s*([kKmM]?)/);
  if (!m) return 0;
  let n = parseFloat(m[1]) || 0;
  if (/[kK]/.test(m[2])) n *= 1e3;
  else if (/[mM]/.test(m[2])) n *= 1e6;
  return n;
}

// ymdLocal formats a Date as a local YYYY-MM-DD (matching stats-cache keys).
function ymdLocal(dt) {
  const y = dt.getFullYear();
  const mo = String(dt.getMonth() + 1).padStart(2, '0');
  const d = String(dt.getDate()).padStart(2, '0');
  return `${y}-${mo}-${d}`;
}

// dailyTokenMap reads stats-cache.json → { 'YYYY-MM-DD': totalTokens }.
function dailyTokenMap() {
  try {
    const os = require('os');
    const path = require('path');
    const file = path.join(os.homedir(), '.claude', 'stats-cache.json');
    const data = JSON.parse(fs.readFileSync(file, 'utf8'));
    const rows = Array.isArray(data.dailyModelTokens) ? data.dailyModelTokens : [];
    const map = {};
    for (const r of rows) {
      if (!r || !r.date) continue;
      let sum = 0;
      const by = r.tokensByModel || {};
      for (const k in by) sum += Number(by[k]) || 0;
      map[r.date] = (map[r.date] || 0) + sum;
    }
    return map;
  } catch (_) { return null; }
}

// sumWindow totals the last `days` days of tokens (including today).
function sumWindow(map, days) {
  const now = new Date();
  let total = 0;
  for (let i = 0; i < days; i++) {
    const dt = new Date(now);
    dt.setDate(now.getDate() - i);
    total += map[ymdLocal(dt)] || 0;
  }
  return total;
}

// usageSeg renders one window: a bar + % when a budget is set, else the raw total.
function usageSeg(label, used, budget) {
  if (budget > 0) {
    const frac = used / budget;
    const pct = frac * 100;
    const col = pct >= 85 ? c.red : pct >= 65 ? c.yellow : c.green;
    return `${col}${label} ${bar(frac, 5)} ${pct.toFixed(pct < 10 ? 1 : 0)}%${c.reset}`;
  }
  return `${c.cyan}${label} ${fmt(used)}${c.reset}`;
}

// quotaSeg renders a REAL subscription limit window (5h session / 7d weekly) as a
// bar + % used, colored by load — the same numbers Claude Code's /usage shows.
function quotaSeg(label, usedPct) {
  const p = Math.max(0, Math.min(100, Number(usedPct) || 0));
  const col = p >= 85 ? c.red : p >= 65 ? c.yellow : c.green;
  return `${col}${label} ${bar(p / 100, 5)} ${p.toFixed(p < 10 ? 1 : 0)}%${c.reset}`;
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

// quota session (5h) + tuần (7d) — số THẬT từ Claude Code, giống hệt lệnh /usage.
// rate_limits chỉ có với tài khoản subscription sau lần gọi API đầu tiên của phiên.
const rl = input.rate_limits || {};
const has5h = rl.five_hour && typeof rl.five_hour.used_percentage === 'number';
const has7d = rl.seven_day && typeof rl.seven_day.used_percentage === 'number';
if (has5h || has7d) {
  if (has5h) parts.push(quotaSeg('5h', rl.five_hour.used_percentage));
  if (has7d) parts.push(quotaSeg('7d', rl.seven_day.used_percentage));
} else {
  // Fallback (chưa có rate_limits / dùng API key): ước lượng token tuần/tháng cục bộ
  // từ stats-cache. Đặt CLAUDE_WEEKLY_TOKEN_BUDGET / CLAUDE_MONTHLY_TOKEN_BUDGET để ra %.
  const daily = dailyTokenMap();
  if (daily) {
    parts.push(usageSeg('W', sumWindow(daily, 7), parseBudget(process.env.CLAUDE_WEEKLY_TOKEN_BUDGET)));
    parts.push(usageSeg('M', sumWindow(daily, 30), parseBudget(process.env.CLAUDE_MONTHLY_TOKEN_BUDGET)));
  }
}

// lines changed (nếu có)
if (linesAdded || linesRemoved) {
  parts.push(`${c.green}+${linesAdded}${c.reset} ${c.red}-${linesRemoved}${c.reset}`);
}

process.stdout.write(parts.join(sep));
