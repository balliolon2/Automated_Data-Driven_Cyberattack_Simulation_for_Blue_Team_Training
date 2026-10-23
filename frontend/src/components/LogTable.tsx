import { useState, useMemo } from "react";
import { ArrowUpDown, AlertCircle, Info, ShieldAlert } from "lucide-react";
import { cn } from "../lib/utils";

interface LogEntry {
  timestamp: string;
  source: string;
  severity: string;
  severity_level?: number;
  message: string;
  src_ip?: string;
  dst_ip?: string;
  user?: string;
  host?: string;
  [key: string]: any;
}

interface LogTableProps {
  logs: LogEntry[];
  selectedIndices?: number[];
  onToggle?: (logIndex: number) => void;
}

export default function LogTable({ logs, selectedIndices = [], onToggle }: LogTableProps) {
  const [sortField, setSortField] = useState<string>("timestamp");
  const [sortAsc, setSortAsc] = useState<boolean>(true);

  const selectedSet = useMemo(() => new Set(selectedIndices), [selectedIndices]);

  // Dynamic columns discovery based on logs present
  const columns = useMemo(() => {
    const defaultCols = ["timestamp", "source", "severity", "host", "message"];
    const allKeys = new Set<string>();
    logs.forEach(log => {
      Object.keys(log).forEach(k => {
        if (k !== "severity_level" && !k.startsWith("__")) allKeys.add(k);
      });
    });
    // Ensure default columns are placed first
    return defaultCols.filter(c => allKeys.has(c));
  }, [logs]);

  const handleSort = (field: string) => {
    if (sortField === field) {
      setSortAsc(!sortAsc);
    } else {
      setSortField(field);
      setSortAsc(true);
    }
  };

  const sortedLogs = useMemo(() => {
    return [...logs].sort((a, b) => {
      const valA = a[sortField] ?? "";
      const valB = b[sortField] ?? "";

      if (valA < valB) return sortAsc ? -1 : 1;
      if (valA > valB) return sortAsc ? 1 : -1;
      return 0;
    });
  }, [logs, sortField, sortAsc]);

  const getSeverityIcon = (level?: number, severity?: string) => {
    const sev = (severity ?? "").toLowerCase();
    const lvl = level ?? 1;

    if (lvl >= 4 || sev === "critical" || sev === "high") {
      return <ShieldAlert className="w-4 h-4 text-red-400 animate-pulse" />;
    }
    if (lvl >= 2 || sev === "warning" || sev === "medium") {
      return <AlertCircle className="w-4 h-4 text-amber-400" />;
    }
    return <Info className="w-4 h-4 text-slate-500" />;
  };

  const getSeverityRowClass = (level?: number, severity?: string) => {
    const sev = (severity ?? "").toLowerCase();
    const lvl = level ?? 1;

    if (lvl >= 4 || sev === "critical" || sev === "high") {
      return "bg-red-950/10 hover:bg-red-950/20 text-red-200/90 border-l-2 border-l-red-500/80";
    }
    if (lvl >= 2 || sev === "warning" || sev === "medium") {
      return "bg-amber-950/5 hover:bg-amber-950/15 text-amber-200/90 border-l-2 border-l-amber-500/60";
    }
    return "hover:bg-graphite-800/30 text-graphite-300 border-l-2 border-l-transparent";
  };

  return (
    <div className="flex flex-col h-full border border-graphite-800/80 rounded-xl overflow-hidden glass-panel">
      {/* Table Header Info */}
      <div className="px-5 py-3 bg-graphite-950/80 border-b border-graphite-800/80 flex items-center justify-between">
        <span className="text-xs font-mono text-graphite-400">
          Logs Found: <span className="text-graphite-200 font-bold">{logs.length} entries</span>
        </span>
        <span className="text-[10px] font-mono text-graphite-500">
          Click a row or press Space/Enter to select evidence · click headers to sort
        </span>
      </div>

      {/* Table Body Scroll Container */}
      <div className="flex-1 overflow-auto max-h-[500px]">
        <table className="w-full text-left border-collapse font-mono text-xs">
          <thead className="sticky top-0 bg-graphite-950/95 backdrop-blur-sm border-b border-graphite-800/80 z-20">
            <tr>
              <th className="w-10 px-3 py-3">
                <span className="sr-only">Select evidence</span>
              </th>
              {columns.map((col) => (
                <th
                  key={col}
                  onClick={() => handleSort(col)}
                  className="px-4 py-3 font-semibold text-graphite-400 uppercase tracking-wider cursor-pointer hover:bg-graphite-900/80 hover:text-graphite-200 transition-all select-none"
                >
                  <div className="flex items-center gap-1.5 capitalize">
                    {col.replace("_", " ")}
                    <ArrowUpDown className="w-3.5 h-3.5 opacity-60 group-hover:opacity-100" />
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-graphite-900/60">
            {sortedLogs.length === 0 ? (
              <tr>
                <td colSpan={columns.length + 1} className="px-4 py-8 text-center text-graphite-500 font-medium bg-graphite-900/10">
                  No log entries match the active filter query.
                </td>
              </tr>
            ) : (
              sortedLogs.map((log, idx) => {
                // Evidence identity must come from the hydrated original index.
                // Never fall back to the sorted position: after a sort or filter
                // that would toggle a different log.
                const logIndex = typeof log.__log_index === "number" ? log.__log_index : null;
                const selectable = logIndex !== null;
                const isSelected = logIndex !== null && selectedSet.has(logIndex);
                const labelIndex = logIndex ?? idx;
                const toggleSelection = () => {
                  if (logIndex === null) return;
                  onToggle?.(logIndex);
                };

                return (
                  <tr
                    key={log.__log_index ?? idx}
                    tabIndex={selectable ? 0 : -1}
                    aria-selected={selectable ? isSelected : undefined}
                    onClick={selectable ? toggleSelection : undefined}
                    onKeyDown={selectable ? (event) => {
                      if (event.key === " " || event.key === "Enter") {
                        event.preventDefault();
                        toggleSelection();
                      }
                    } : undefined}
                    className={cn(
                      "transition-colors duration-150 border-b border-graphite-900/40 outline-none focus-visible:ring-1 focus-visible:ring-inset focus-visible:ring-cyan-500/70",
                      selectable && "cursor-pointer",
                      isSelected
                        ? "bg-cyan-950/30 hover:bg-cyan-900/40 text-cyan-100 border-l-2 border-l-cyan-400"
                        : getSeverityRowClass(log.severity_level, log.severity)
                    )}
                  >
                    <td className="px-3 py-2.5 align-middle">
                      <input
                        type="checkbox"
                        checked={isSelected}
                        disabled={!selectable}
                        aria-label={`Select log ${labelIndex}`}
                        onChange={toggleSelection}
                        onClick={(event) => event.stopPropagation()}
                        onKeyDown={(event) => {
                          // Space is handled by the checkbox itself; stop it so
                          // the row handler does not toggle a second time. Enter
                          // bubbles to the row so it toggles exactly once.
                          if (event.key === " " || event.key === "Spacebar") {
                            event.stopPropagation();
                          }
                        }}
                        className="h-3.5 w-3.5 accent-cyan-400 cursor-pointer disabled:cursor-not-allowed"
                      />
                    </td>
                    {columns.map((col) => {
                      const value = log[col];
                      const isMessage = col === "message";
                      const isSeverity = col === "severity";

                      return (
                        <td
                          key={col}
                          className={cn(
                            "px-4 py-2.5 break-all max-w-sm whitespace-pre-wrap vertical-middle align-middle",
                            isMessage
                              ? cn("max-w-xl font-normal", isSelected ? "text-cyan-100" : "text-graphite-200")
                              : "font-medium"
                          )}
                        >
                          {isSeverity ? (
                            <div className="flex items-center gap-1.5 font-bold uppercase tracking-wide">
                              {getSeverityIcon(log.severity_level, log.severity)}
                              <span>{value}</span>
                            </div>
                          ) : (
                            String(value ?? "-")
                          )}
                        </td>
                      );
                    })}
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
