import { useState, useEffect } from "react";
import { Terminal, Shield, CheckCircle2, Activity, Clock, Server, ArrowRight } from "lucide-react";

interface PipelineStep {
  id: string;
  code: string;
  label: string;
  details: string;
}

const STEPS: PipelineStep[] = [
  {
    id: "profile",
    code: "01/04",
    label: "SKILL GAP & PROFICIENCY PARSER",
    details: "Analyzing domain performance metrics & CompTIA Security+ weak spots...",
  },
  {
    id: "mitre",
    code: "02/04",
    label: "MITRE ATT&CK RAG RETRIEVAL",
    details: "Querying vector DB for TTPs matching target domain gap...",
  },
  {
    id: "synthesis",
    code: "03/04",
    label: "ADVERSARIAL SCENARIO SYNTHESIS",
    details: "Generating realistic attack chain, playbook steps & evidence logs...",
  },
  {
    id: "sandbox",
    code: "04/04",
    label: "SIEM & TELEMETRY SANDBOX INJECTION",
    details: "Indexing Firewall, EDR & Auth logs into virtual investigation workspace...",
  },
];

const STREAM_LOGS = [
  "SYS_INIT: Initializing SOC Virtual Sandbox environment v2.4",
  "AUTH_CHECK: Token verified | Learner Profile: Active Tier 1",
  "EVAL_ENGINE: Calculating proficiency delta across 5 Security+ domains",
  "RAG_QUERY: Executing semantic search on vector DB (CompTIA_701_TTPs)",
  "MITRE_MAP: Matched technique candidate: T1059.001 (PowerShell Execution)",
  "MITRE_MAP: Matched technique candidate: T1078 (Valid Accounts)",
  "LLM_GEN: Requesting structured JSON attack payload generation...",
  "LOG_ENGINE: Synthesizing 250+ telemetry events across 3 device sources",
  "LOG_PARSER: Ingesting PaloAlto Firewall traffic events [Src: 192.168.1.105]",
  "LOG_PARSER: Ingesting CrowdStrike EDR process lineage tree",
  "LOG_PARSER: Ingesting Windows Security Event Logs (EventID 4624, 4625)",
  "PLAYBOOK_GEN: Constructing Incident Response decision options...",
  "EVIDENCE_INDEX: Tagging sanitized evidence log indices for scoring",
  "SANDBOX_READY: Finalizing virtual environment. Launching session...",
];

export default function ScenarioLoadingState() {
  const [elapsedSeconds, setElapsedSeconds] = useState(0);
  const [currentStepIndex, setCurrentStepIndex] = useState(0);
  const [logIndex, setLogIndex] = useState(0);
  const [logs, setLogs] = useState<string[]>([]);
  const [progress, setProgress] = useState(12);

  // Timer for elapsed time
  useEffect(() => {
    const timer = setInterval(() => {
      setElapsedSeconds((prev) => prev + 0.1);
    }, 100);
    return () => clearInterval(timer);
  }, []);

  // Step advancement timer
  useEffect(() => {
    const stepInterval = setInterval(() => {
      setCurrentStepIndex((prev) => {
        if (prev < STEPS.length - 1) {
          return prev + 1;
        }
        return prev;
      });
    }, 2800);

    return () => clearInterval(stepInterval);
  }, []);

  // Progress bar calculation
  useEffect(() => {
    const progressInterval = setInterval(() => {
      setProgress((prev) => {
        if (prev >= 94) return 94;
        const bump = Math.floor(Math.random() * 4) + 2;
        return Math.min(94, prev + bump);
      });
    }, 350);

    return () => clearInterval(progressInterval);
  }, []);

  // Stream logs timer
  useEffect(() => {
    const logInterval = setInterval(() => {
      if (logIndex < STREAM_LOGS.length) {
        const newLog = `[${new Date().toISOString().substring(11, 19)}] ${STREAM_LOGS[logIndex]}`;
        setLogs((prev) => [...prev.slice(-6), newLog]);
        setLogIndex((prev) => prev + 1);
      }
    }, 900);

    return () => clearInterval(logInterval);
  }, [logIndex]);

  const formattedTime = elapsedSeconds.toFixed(1);

  return (
    <div className="min-h-[75vh] flex items-center justify-center p-4 md:p-6 animate-fade-in">
      <div className="w-full max-w-4xl bg-graphite-950 border border-graphite-800 rounded-xl shadow-2xl overflow-hidden text-graphite-100 font-sans">
        
        {/* Top Minimalist Header */}
        <div className="px-6 py-4 bg-graphite-900/80 border-b border-graphite-800 flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="p-1.5 rounded-md bg-graphite-800 border border-graphite-700 text-emerald-400">
              <Terminal className="w-4 h-4" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="text-xs font-mono font-bold text-slate-200 tracking-wider">
                  SOC_PROVISIONER // THREAT_ENGINE
                </span>
                <span className="inline-block w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
              </div>
              <p className="text-[11px] text-graphite-400 font-mono">
                TASK_ID: SIM-RUN-{Math.floor(elapsedSeconds * 17 + 1024)}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-6 text-xs font-mono">
            <div className="flex items-center gap-2 text-graphite-400">
              <Clock className="w-3.5 h-3.5 text-slate-400" />
              <span>ELAPSED:</span>
              <span className="text-slate-100 font-semibold tabular-nums">{formattedTime}s</span>
            </div>
            <div className="flex items-center gap-2 text-graphite-400">
              <Activity className="w-3.5 h-3.5 text-emerald-400" />
              <span>STATUS:</span>
              <span className="text-emerald-400 font-semibold">SYNTHESIZING</span>
            </div>
          </div>
        </div>

        {/* Main Content Area */}
        <div className="p-6 md:p-8 space-y-8">
          
          {/* Headline & Progress Bar */}
          <div className="space-y-3">
            <div className="flex items-baseline justify-between">
              <div>
                <h2 className="text-lg md:text-xl font-semibold text-slate-100 tracking-tight">
                  Constructing Cyber Attack Scenario
                </h2>
                <p className="text-xs text-graphite-400 mt-1 font-mono">
                  Synthesizing threat vectors & injecting simulated SIEM telemetry...
                </p>
              </div>
              <div className="text-right">
                <span className="text-xl md:text-2xl font-mono font-bold text-slate-100 tabular-nums">
                  {progress}%
                </span>
              </div>
            </div>

            {/* High-Contrast Crisp Tracer Bar */}
            <div className="relative h-1.5 w-full bg-graphite-900 rounded-full overflow-hidden border border-graphite-800/60">
              <div
                className="h-full bg-emerald-400 rounded-full transition-all duration-300 relative"
                style={{ width: `${progress}%` }}
              >
                <div className="absolute right-0 top-0 bottom-0 w-3 bg-white blur-[2px] opacity-80" />
              </div>
            </div>
          </div>

          {/* Pipeline Stage Matrix */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {STEPS.map((step, index) => {
              const isCompleted = index < currentStepIndex;
              const isActive = index === currentStepIndex;
              const isWaiting = index > currentStepIndex;

              return (
                <div
                  key={step.id}
                  className={`p-4 rounded-lg border transition-all duration-200 ${
                    isActive
                      ? "bg-graphite-900 border-emerald-500/50 shadow-[0_0_15px_rgba(16,185,129,0.06)]"
                      : isCompleted
                      ? "bg-graphite-900/40 border-graphite-800/80 text-graphite-300"
                      : "bg-graphite-950 border-graphite-900 opacity-60"
                  }`}
                >
                  <div className="flex items-start justify-between gap-3 mb-1.5">
                    <div className="flex items-center gap-2">
                      <span className="text-[10px] font-mono text-graphite-400 px-1.5 py-0.5 rounded bg-graphite-800">
                        {step.code}
                      </span>
                      <span className="text-xs font-mono font-semibold tracking-wide text-slate-200">
                        {step.label}
                      </span>
                    </div>

                    <div>
                      {isCompleted && (
                        <span className="flex items-center gap-1 text-[11px] font-mono text-emerald-400">
                          <CheckCircle2 className="w-3.5 h-3.5" />
                          <span>DONE</span>
                        </span>
                      )}
                      {isActive && (
                        <span className="flex items-center gap-1 text-[11px] font-mono text-emerald-400 animate-pulse">
                          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400" />
                          <span>ACTIVE</span>
                        </span>
                      )}
                      {isWaiting && (
                        <span className="text-[11px] font-mono text-graphite-500">
                          WAITING
                        </span>
                      )}
                    </div>
                  </div>

                  <p className="text-xs text-graphite-400 font-mono leading-relaxed pl-0.5">
                    {step.details}
                  </p>
                </div>
              );
            })}
          </div>

          {/* Live Tactical Console Log Feed */}
          <div className="space-y-2">
            <div className="flex items-center justify-between text-xs font-mono text-graphite-400 px-1">
              <div className="flex items-center gap-2">
                <Server className="w-3.5 h-3.5 text-slate-400" />
                <span>TACTICAL TELEMETRY STREAM</span>
              </div>
              <span className="text-[10px] text-graphite-500">FORMAT: IEEE 802.3 SIEM</span>
            </div>

            <div className="bg-graphite-900 border border-graphite-800 rounded-lg p-4 font-mono text-xs text-slate-300 min-h-[140px] flex flex-col justify-end space-y-1.5 overflow-hidden">
              {logs.length === 0 ? (
                <div className="text-graphite-500 italic">Initializing console pipe...</div>
              ) : (
                logs.map((log, idx) => (
                  <div
                    key={idx}
                    className={`flex items-start gap-2 leading-tight transition-all duration-150 ${
                      idx === logs.length - 1 ? "text-emerald-300 font-medium" : "text-graphite-400"
                    }`}
                  >
                    <ArrowRight className="w-3 h-3 text-graphite-600 mt-0.5 shrink-0" />
                    <span className="break-all">{log}</span>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Minimalist System Footer */}
        <div className="px-6 py-3 bg-graphite-900/60 border-t border-graphite-800 flex flex-wrap items-center justify-between text-[11px] font-mono text-graphite-400 gap-4">
          <div className="flex items-center gap-4">
            <span>STANDARD: <strong className="text-slate-300">COMPTIA SEC+ 701</strong></span>
          </div>
          <div className="flex items-center gap-2">
            <Shield className="w-3.5 h-3.5 text-emerald-400" />
            <span>SANDBOX ISOLATED & ENCRYPTED</span>
          </div>
        </div>

      </div>
    </div>
  );
}
