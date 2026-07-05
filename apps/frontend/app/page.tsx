"use client";

import { Copy, Play, RotateCcw } from "lucide-react";
import { FormEvent, useState } from "react";

type ExecutePayload = {
  data?: unknown;
  output?: unknown;
  stdout?: unknown;
  stderr?: unknown;
  out?: unknown;
  err?: unknown;
  error?: unknown;
  exitCode?: unknown;
};

type RunResult = {
  stdout: string;
  stderr: string;
  status: "idle" | "running" | "success" | "error";
  exitCode: string;
};

const defaultCode = `const jobs = [
  {
    job: "A",
    boxType: "Small mailer",
    jobStart: "2026-06-20T08:00:00",
    jobEnd: "2026-06-20T10:15:00",
    nextJobStart: "2026-06-20T10:42:00",
    boxesMade: 5000,
    goodBoxes: 4960,
    downtimeMinutes: 8
  },
  {
    job: "B",
    boxType: "Large shipper",
    jobStart: "2026-06-20T10:42:00",
    jobEnd: "2026-06-20T13:00:00",
    nextJobStart: "2026-06-20T13:20:00",
    boxesMade: 4600,
    goodBoxes: 4545,
    downtimeMinutes: 5
  },
  {
    job: "C",
    boxType: "Printed retail box",
    jobStart: "2026-06-20T13:20:00",
    jobEnd: "2026-06-20T15:10:00",
    nextJobStart: "2026-06-20T15:55:00",
    boxesMade: 3900,
    goodBoxes: 3870,
    downtimeMinutes: 15
  }
];

const minutesBetween = (start, end) =>
  Math.round((new Date(end) - new Date(start)) / 60000);

const rows = jobs.map((job) => {
  const productionMinutes = minutesBetween(job.jobStart, job.jobEnd);
  const changeoverMinutes = minutesBetween(job.jobEnd, job.nextJobStart);
  const qualityPercent = (job.goodBoxes / job.boxesMade) * 100;

  return {
    job: job.job,
    boxType: job.boxType,
    productionMinutes,
    changeoverMinutes,
    boxesPerHour: Math.round((job.goodBoxes / productionMinutes) * 60),
    cycleSecondsPerGoodBox: Number(((productionMinutes * 60) / job.goodBoxes).toFixed(2)),
    downtimeMinutes: job.downtimeMinutes,
    qualityPercent: Number(qualityPercent.toFixed(2))
  };
});

const chartData = rows.map((row) => ({
  label: row.job,
  changeoverMinutes: row.changeoverMinutes,
  boxesPerHour: row.boxesPerHour,
  downtimeMinutes: row.downtimeMinutes
}));

console.table(rows);
console.log("Chart data:");
console.log(JSON.stringify(chartData, null, 2));

const slowestChangeover = rows.reduce((slowest, row) =>
  row.changeoverMinutes > slowest.changeoverMinutes ? row : slowest
);

console.log(
  "Biggest setup delay:",
  slowestChangeover.job,
  slowestChangeover.boxType,
  slowestChangeover.changeoverMinutes + " minutes"
);`;

const runnerBaseUrl = process.env.NEXT_PUBLIC_RUNNER_BASE_URL ?? "";
const executeEndpoint = process.env.NEXT_PUBLIC_EXECUTE_ENDPOINT ?? "/api/execute";
const executionTimeoutMs = 120_000;

function cleanBaseUrl(value: string) {
  return value.trim().replace(/\/+$/, "");
}

function buildRequestUrl() {
  const base = cleanBaseUrl(runnerBaseUrl);
  if (!base) {
    return "";
  }

  const path = executeEndpoint.trim().startsWith("/")
    ? executeEndpoint.trim()
    : `/${executeEndpoint.trim()}`;

  return `${base}${path}`;
}

function asText(value: unknown) {
  if (typeof value === "string") {
    return value;
  }

  if (value == null) {
    return "";
  }

  return JSON.stringify(value, null, 2);
}

function normalizeResponse(payload: ExecutePayload): RunResult {
  const nested = typeof payload.data === "object" && payload.data !== null
    ? (payload.data as ExecutePayload)
    : undefined;

  const stdout =
    asText(payload.output) ||
    asText(payload.stdout) ||
    asText(payload.out) ||
    asText(nested?.output) ||
    asText(nested?.stdout) ||
    asText(nested?.out) ||
    asText(payload.data);

  const stderr =
    asText(payload.stderr) ||
    asText(payload.err) ||
    asText(payload.error) ||
    asText(nested?.stderr) ||
    asText(nested?.err) ||
    asText(nested?.error);

  const exitCode = payload.exitCode ?? nested?.exitCode ?? "";

  return {
    stdout,
    stderr,
    status: stderr ? "error" : "success",
    exitCode: asText(exitCode)
  };
}

export default function Home() {
  const [code, setCode] = useState(defaultCode);
  const requestUrl = buildRequestUrl();
  const [result, setResult] = useState<RunResult>({
    stdout: "",
    stderr: "",
    status: "idle",
    exitCode: ""
  });

  async function execute(nextCode = code) {
    if (!cleanBaseUrl(runnerBaseUrl)) {
      setResult({
        stdout: "",
        stderr: "Missing NEXT_PUBLIC_RUNNER_BASE_URL.",
        status: "error",
        exitCode: ""
      });
      return;
    }

    setResult({
      stdout: "",
      stderr: "",
      status: "running",
      exitCode: ""
    });

    const controller = new AbortController();
    const timeoutId = window.setTimeout(() => controller.abort(), executionTimeoutMs);

    try {
      const response = await fetch(requestUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        signal: controller.signal,
        body: JSON.stringify({ code: nextCode })
      });

      const text = await response.text();
      let payload: ExecutePayload = {};

      if (text) {
        try {
          payload = JSON.parse(text) as ExecutePayload;
        } catch {
          payload = { data: text };
        }
      }

      const normalized = normalizeResponse(payload);

      setResult({
        ...normalized,
        status: response.ok && !normalized.stderr ? "success" : "error",
        stderr: response.ok
          ? normalized.stderr
          : normalized.stderr || `Request failed with HTTP ${response.status}.`
      });
    } catch (error) {
      const message =
        error instanceof DOMException && error.name === "AbortError"
          ? "Backend execution timed out before a response was returned."
          : error instanceof Error
            ? error.message
            : String(error);

      setResult({
        stdout: "",
        stderr: message,
        status: "error",
        exitCode: ""
      });
    } finally {
      window.clearTimeout(timeoutId);
    }
  }

  function runCode(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void execute();
  }

  async function copyOutput() {
    const output = `STDOUT\n${result.stdout || "(empty)"}\n\nSTDERR\n${
      result.stderr || "(empty)"
    }`;

    await navigator.clipboard.writeText(output);
  }

  return (
    <main className="shell">
      <section className="topbar" aria-label="VM connection">
        <div>
          <p className="eyebrow">Sandbox JS Runner</p>
          <h1>Execute JavaScript in a Firecracker microVM</h1>
        </div>
        <div className={`status status-${result.status}`}>
          <span aria-hidden="true" />
          {result.status === "idle" ? "Ready" : result.status}
        </div>
      </section>

      <form className="workspace" onSubmit={runCode}>
        <section className="panel editor-panel" aria-label="JavaScript editor">
          <div className="panel-header">
            <div>
              <h2>Request code</h2>
              <p>Sent to the backend for Firecracker execution.</p>
            </div>
            <button className="icon-button" type="button" onClick={() => setCode(defaultCode)}>
              <RotateCcw size={18} aria-hidden="true" />
              <span className="sr-only">Reset code</span>
            </button>
          </div>
          <textarea
            value={code}
            onChange={(event) => setCode(event.target.value)}
            spellCheck={false}
            aria-label="JavaScript code"
          />
          <div className="actions">
            <button className="primary-button" type="submit" disabled={result.status === "running"}>
              <Play size={18} aria-hidden="true" />
              {result.status === "running" ? "Executing" : "Execute request"}
            </button>
          </div>
        </section>

        <section className="panel output-panel" aria-label="Execution output">
          <div className="panel-header">
            <div>
              <h2>Backend response</h2>
              <p>Output returned after the Firecracker VM finishes.</p>
            </div>
            <button className="icon-button" type="button" onClick={copyOutput}>
              <Copy size={18} aria-hidden="true" />
              <span className="sr-only">Copy output</span>
            </button>
          </div>

          {result.exitCode ? <div className="exit-code">Exit {result.exitCode}</div> : null}

          <div className="stream-grid">
            <article className="stream">
              <div className="stream-title">out</div>
              <pre>{result.stdout || "No stdout yet."}</pre>
            </article>
            <article className="stream stream-error">
              <div className="stream-title">err</div>
              <pre>{result.stderr || "No stderr yet."}</pre>
            </article>
          </div>
        </section>
      </form>
    </main>
  );
}
