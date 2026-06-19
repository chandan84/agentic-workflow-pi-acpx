/**
 * PiSpec Flow — Approval Sidecar Server
 *
 * A minimal Node.js HTTP server (no external dependencies) that handles
 * human approval decisions for spec reviews.
 *
 * Routes:
 *   POST /api/approval             — record an approval decision
 *   GET  /api/approval/:change_id  — retrieve existing decision (or 404)
 *   GET  /health                   — liveness probe
 *
 * Storage:
 *   Writes JSON to /workspace/openspec/changes/{change_id}/human-approval.json
 *   The /workspace volume is a writable EFS mount (separate from the
 *   read-only /specs mount).
 */

'use strict';

const http = require('http');
const fs   = require('fs');
const path = require('path');

/* ── Configuration ──────────────────────────────────────────────────────── */
const PORT           = parseInt(process.env.PORT || '3001', 10);
const WORKSPACE_ROOT = process.env.WORKSPACE_ROOT || '/workspace';
const LOG_LEVEL      = (process.env.LOG_LEVEL || 'info').toLowerCase();

/* ── Logging ────────────────────────────────────────────────────────────── */
const LEVELS = { debug: 0, info: 1, warn: 2, error: 3 };

function log (level, msg, meta = {}) {
  if (LEVELS[level] < LEVELS[LOG_LEVEL]) return;
  const entry = {
    ts:    new Date().toISOString(),
    level,
    msg,
    ...meta
  };
  const output = JSON.stringify(entry);
  if (level === 'error' || level === 'warn') {
    process.stderr.write(output + '\n');
  } else {
    process.stdout.write(output + '\n');
  }
}

/* ── Helpers ────────────────────────────────────────────────────────────── */

/**
 * Validate that a change_id is safe to use as a filesystem path component.
 * Allows alphanumeric, hyphens, underscores, and dots only.
 */
function isValidChangeId (id) {
  return typeof id === 'string' &&
         id.length > 0 &&
         id.length <= 128 &&
         /^[a-zA-Z0-9_\-\.]+$/.test(id);
}

/**
 * Resolve the path to the approval JSON for a given change_id.
 * Always verified to remain under WORKSPACE_ROOT to prevent traversal.
 */
function approvalFilePath (changeId) {
  const resolved = path.resolve(
    WORKSPACE_ROOT,
    'openspec', 'changes', changeId, 'human-approval.json'
  );
  // Guard against path traversal
  const expectedPrefix = path.resolve(WORKSPACE_ROOT) + path.sep;
  if (!resolved.startsWith(expectedPrefix)) {
    throw new Error('Path traversal detected');
  }
  return resolved;
}

/** Read the full request body as a Buffer, then parse as JSON. */
function readJsonBody (req) {
  return new Promise((resolve, reject) => {
    const chunks = [];
    let size = 0;
    const MAX = 1024 * 64; // 64 KB safety cap

    req.on('data', chunk => {
      size += chunk.length;
      if (size > MAX) {
        reject(new Error('Request body too large'));
        req.destroy();
        return;
      }
      chunks.push(chunk);
    });

    req.on('end', () => {
      try {
        const raw = Buffer.concat(chunks).toString('utf8');
        resolve(JSON.parse(raw));
      } catch (e) {
        reject(new Error('Invalid JSON body'));
      }
    });

    req.on('error', reject);
  });
}

/** Send a JSON response. */
function sendJson (res, statusCode, body) {
  const payload = JSON.stringify(body);
  res.writeHead(statusCode, {
    'Content-Type':  'application/json; charset=utf-8',
    'Content-Length': Buffer.byteLength(payload),
    'X-Content-Type-Options': 'nosniff'
  });
  res.end(payload);
}

/** Send a plain-text error response. */
function sendError (res, statusCode, message) {
  log('warn', 'Returning error response', { statusCode, message });
  sendJson(res, statusCode, { error: message });
}

/* ── Route handlers ─────────────────────────────────────────────────────── */

/**
 * GET /health
 * Kubernetes liveness / readiness probe.
 */
function handleHealth (res) {
  res.writeHead(200, { 'Content-Type': 'text/plain' });
  res.end('ok\n');
}

/**
 * GET /api/approval/:change_id
 * Returns 200 + approval JSON if a decision exists, or 404.
 */
function handleGetApproval (res, changeId) {
  if (!isValidChangeId(changeId)) {
    return sendError(res, 400, 'Invalid change_id');
  }

  let filePath;
  try {
    filePath = approvalFilePath(changeId);
  } catch (e) {
    return sendError(res, 400, e.message);
  }

  fs.readFile(filePath, 'utf8', (err, data) => {
    if (err) {
      if (err.code === 'ENOENT') {
        return sendError(res, 404, 'No approval decision found for this change_id');
      }
      log('error', 'Failed to read approval file', { filePath, err: err.message });
      return sendError(res, 500, 'Failed to read approval file');
    }

    let parsed;
    try {
      parsed = JSON.parse(data);
    } catch (e) {
      log('error', 'Approval file contains invalid JSON', { filePath });
      return sendError(res, 500, 'Approval file is corrupt');
    }

    log('info', 'Approval status retrieved', { changeId, approved: parsed.approved });
    sendJson(res, 200, parsed);
  });
}

/**
 * POST /api/approval
 * Body (JSON):
 *   {
 *     change_id:      string   (required)
 *     approved:       boolean  (required)
 *     reviewer_email: string   (required)
 *     timestamp:      string   (ISO 8601, required)
 *     feedback?:      string   (optional, used when approved=false)
 *   }
 *
 * Writes to /workspace/openspec/changes/{change_id}/human-approval.json
 * Returns 200 + the stored approval object.
 * Returns 409 if a decision already exists (idempotent guard).
 */
async function handlePostApproval (req, res) {
  let body;
  try {
    body = await readJsonBody(req);
  } catch (e) {
    return sendError(res, 400, e.message);
  }

  // ── Validate required fields ───────────────────────────────────────────
  const { change_id, approved, reviewer_email, timestamp, feedback } = body;

  if (!isValidChangeId(change_id)) {
    return sendError(res, 400, 'Missing or invalid change_id');
  }

  if (typeof approved !== 'boolean') {
    return sendError(res, 400, '"approved" must be a boolean');
  }

  if (!reviewer_email || typeof reviewer_email !== 'string' ||
      !reviewer_email.includes('@') || reviewer_email.length > 320) {
    return sendError(res, 400, 'Invalid or missing reviewer_email');
  }

  if (!timestamp || typeof timestamp !== 'string' || isNaN(Date.parse(timestamp))) {
    return sendError(res, 400, 'Invalid or missing timestamp (must be ISO 8601)');
  }

  if (!approved && feedback !== undefined && typeof feedback !== 'string') {
    return sendError(res, 400, '"feedback" must be a string when provided');
  }

  // ── Resolve path ───────────────────────────────────────────────────────
  let filePath;
  try {
    filePath = approvalFilePath(change_id);
  } catch (e) {
    return sendError(res, 400, e.message);
  }

  // ── Idempotency: refuse to overwrite an existing decision ──────────────
  if (fs.existsSync(filePath)) {
    log('warn', 'Approval already exists, rejecting overwrite', { change_id });
    return sendError(res, 409, 'An approval decision already exists for this change_id');
  }

  // ── Build approval record ──────────────────────────────────────────────
  const record = {
    change_id,
    approved,
    reviewer_email: reviewer_email.trim().toLowerCase(),
    timestamp:      new Date(timestamp).toISOString(),
    ...(feedback !== undefined && feedback !== '' ? { feedback: feedback.trim() } : {})
  };

  // ── Ensure directory exists ────────────────────────────────────────────
  const dir = path.dirname(filePath);
  try {
    fs.mkdirSync(dir, { recursive: true });
  } catch (e) {
    log('error', 'Failed to create approval directory', { dir, err: e.message });
    return sendError(res, 500, 'Failed to create approval directory');
  }

  // ── Write atomically via temp file + rename ────────────────────────────
  const tmpPath = `${filePath}.tmp.${process.pid}`;
  const payload = JSON.stringify(record, null, 2) + '\n';

  try {
    fs.writeFileSync(tmpPath, payload, { encoding: 'utf8', flag: 'wx', mode: 0o644 });
    fs.renameSync(tmpPath, filePath);
  } catch (e) {
    // Clean up temp file if it exists
    try { fs.unlinkSync(tmpPath); } catch (_) { /* ignore */ }
    log('error', 'Failed to write approval file', { filePath, err: e.message });
    return sendError(res, 500, 'Failed to persist approval decision');
  }

  log('info', 'Approval decision recorded', {
    change_id,
    approved,
    reviewer_email: record.reviewer_email
  });

  sendJson(res, 200, record);
}

/* ── Request router ─────────────────────────────────────────────────────── */

async function router (req, res) {
  const method = req.method ? req.method.toUpperCase() : '';
  const url    = req.url    || '/';

  // Strip query string for routing
  const urlPath = url.split('?')[0].replace(/\/+$/, '') || '/';

  log('debug', 'Incoming request', { method, urlPath, remoteAddress: req.socket?.remoteAddress });

  // GET /health
  if (method === 'GET' && urlPath === '/health') {
    return handleHealth(res);
  }

  // GET /api/approval/:change_id
  const getMatch = urlPath.match(/^\/api\/approval\/(.+)$/);
  if (method === 'GET' && getMatch) {
    return handleGetApproval(res, decodeURIComponent(getMatch[1]));
  }

  // POST /api/approval
  if (method === 'POST' && urlPath === '/api/approval') {
    return handlePostApproval(req, res);
  }

  // OPTIONS preflight (nginx handles CORS but sidecar should also respond)
  if (method === 'OPTIONS') {
    res.writeHead(204, {
      'Access-Control-Allow-Origin':  '*',
      'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Accept',
      'Access-Control-Max-Age':       '3600',
      'Content-Length':               '0'
    });
    return res.end();
  }

  // 404 for everything else
  sendError(res, 404, `Route not found: ${method} ${urlPath}`);
}

/* ── Server bootstrap ───────────────────────────────────────────────────── */

const server = http.createServer(async (req, res) => {
  try {
    await router(req, res);
  } catch (err) {
    log('error', 'Unhandled error in request handler', { err: err.message, stack: err.stack });
    if (!res.headersSent) {
      sendError(res, 500, 'Internal server error');
    }
  }
});

server.listen(PORT, '127.0.0.1', () => {
  log('info', 'Approval sidecar listening', {
    port:          PORT,
    workspaceRoot: WORKSPACE_ROOT,
    nodeVersion:   process.version
  });
});

/* ── Graceful shutdown ──────────────────────────────────────────────────── */
function shutdown (signal) {
  log('info', `Received ${signal}, shutting down gracefully`);
  server.close(err => {
    if (err) {
      log('error', 'Error during server close', { err: err.message });
      process.exit(1);
    }
    log('info', 'Server closed cleanly');
    process.exit(0);
  });

  // Force-kill if graceful shutdown takes too long
  setTimeout(() => {
    log('error', 'Graceful shutdown timed out, forcing exit');
    process.exit(1);
  }, 10_000).unref();
}

process.on('SIGTERM', () => shutdown('SIGTERM'));
process.on('SIGINT',  () => shutdown('SIGINT'));

process.on('uncaughtException', err => {
  log('error', 'Uncaught exception', { err: err.message, stack: err.stack });
  process.exit(1);
});

process.on('unhandledRejection', (reason) => {
  log('error', 'Unhandled promise rejection', {
    reason: reason instanceof Error ? reason.message : String(reason)
  });
  process.exit(1);
});
