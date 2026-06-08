"use client";

import { Copy, Play, RotateCcw } from "lucide-react";
import { FormEvent, useMemo, useState } from "react";

type ExecutePayload = {
  data?: unknown;
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

const defaultCode = `console.log("hello from the sandbox");
console.error("stderr example");

const result = [1, 2, 3].map((value) => value * 2);
console.log(JSON.stringify({ result }, null, 2));`;

function cleanBaseUrl(value: string) {
  return value.trim().replace(/\/+$/, "");
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
    asText(payload.stdout) ||
    asText(payload.out) ||
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
  const [vmUrl, setVmUrl] = useState(
    process.env.NEXT_PUBLIC_DEFAULT_VM_URL ?? "http://localhost:3000"
  );
  const [endpoint, setEndpoint] = useState("/api/execute/js");
  const [code, setCode] = useState(defaultCode);
  const [result, setResult] = useState<RunResult>({
    stdout: "",
    stderr: "",
    status: "idle",
    exitCode: ""
  });

  const requestUrl = useMemo(() => {
    const base = cleanBaseUrl(vmUrl);
    const path = endpoint.trim().startsWith("/") ? endpoint.trim() : `/${endpoint.trim()}`;
    return `${base}${path}`;
  }, [endpoint, vmUrl]);

  async function execute(nextCode = code) {
    if (!cleanBaseUrl(vmUrl)) {
      setResult({
        stdout: "",
        stderr: "Enter a VM URL before running code.",
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

    try {
      const response = await fetch(requestUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
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
      setResult({
        stdout: "",
        stderr: error instanceof Error ? error.message : String(error),
        status: "error",
        exitCode: ""
      });
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
          <h1>Execute JavaScript in a fresh VM</h1>
        </div>
        <div className={`status status-${result.status}`}>
          <span aria-hidden="true" />
          {result.status === "idle" ? "Ready" : result.status}
        </div>
      </section>

      <form className="workspace" onSubmit={runCode}>
        <section className="panel connection-panel" aria-label="Execution target">
          <label>
            <span>Runner URL</span>
            <input
              type="url"
              value={vmUrl}
              onChange={(event) => setVmUrl(event.target.value)}
              placeholder="https://runner.example.com"
            />
          </label>

          <label>
            <span>Execute endpoint</span>
            <input
              value={endpoint}
              onChange={(event) => setEndpoint(event.target.value)}
              placeholder="/api/execute/js"
            />
          </label>
        </section>

        <section className="panel editor-panel" aria-label="JavaScript editor">
          <div className="panel-header">
            <h2>Code</h2>
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
              {result.status === "running" ? "Running" : "Run code"}
            </button>
          </div>
        </section>

        <section className="panel output-panel" aria-label="Execution output">
          <div className="panel-header">
            <div>
              <h2>Output</h2>
              <p>{requestUrl}</p>
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
