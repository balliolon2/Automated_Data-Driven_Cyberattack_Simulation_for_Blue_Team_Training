import { useState, type FormEvent, useEffect } from "react";
import { Search, AlertCircle, History } from "lucide-react";
import axios from "axios";

interface KqlSearchBarProps {
  sessionId: string;
  onFilter: (filteredLogs: any[]) => void;
  logs: any[];
  searchQuery?: string;
  onSearchQueryChange?: (q: string) => void;
}

export default function KqlSearchBar({ sessionId, onFilter, logs, searchQuery, onSearchQueryChange }: KqlSearchBarProps) {
  const [query, setQuery] = useState(searchQuery || "");
  const [error, setError] = useState<string | null>(null);
  const [history, setHistory] = useState<string[]>([]);
  const [showHistory, setShowHistory] = useState(false);

  // Sync external searchQuery prop if provided
  useEffect(() => {
    if (searchQuery !== undefined) {
      setQuery(searchQuery);
      runFilter(searchQuery);
    }
  }, [searchQuery, logs]);

  // Helper to load history from local storage
  useEffect(() => {
    const saved = localStorage.getItem(`kql_history_${sessionId}`);
    if (saved) {
      setHistory(JSON.parse(saved));
    }
  }, [sessionId]);

  const saveToHistory = (q: string) => {
    if (!q.trim() || history.includes(q)) return;
    const newHistory = [q, ...history].slice(0, 10);
    setHistory(newHistory);
    localStorage.setItem(`kql_history_${sessionId}`, JSON.stringify(newHistory));
  };

  // Client-side KQL implementation
  const runFilter = (searchQueryStr: string) => {
    setError(null);
    const q = searchQueryStr.trim();
    if (onSearchQueryChange) {
      onSearchQueryChange(searchQueryStr);
    }
    
    if (!q) {
      onFilter(logs);
      return;
    }

    try {
      const filtered = logs.filter(log => evaluateQuery(log, q));
      onFilter(filtered);
    } catch (err: any) {
      setError(err.message || "Invalid KQL query syntax");
      onFilter([]); // show empty logs on error
    }
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    runFilter(query);

    // Call API to log this query for analysis
    try {
      const token = localStorage.getItem("token");
      await axios.post("/api/simulation/log-query", {
        session_id: sessionId,
        query: query
      }, {
        headers: { Authorization: `Bearer ${token}` }
      });
      saveToHistory(query);
    } catch (err) {
      console.error("Failed to log query to server:", err);
    }
  };

  const handleSelectHistory = (historyQuery: string) => {
    setQuery(historyQuery);
    setShowHistory(false);
    runFilter(historyQuery);
  };

  return (
    <div className="space-y-2 relative">
      <form onSubmit={handleSubmit} className="flex gap-2">
        <div className="relative flex-1">
          <input
            type="text"
            placeholder='e.g., source == "firewall" and severity == "high" or message contains "PowerShell"'
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onFocus={() => setShowHistory(true)}
            onBlur={() => setTimeout(() => setShowHistory(false), 200)}
            className="w-full bg-graphite-950 text-graphite-100 placeholder-graphite-500 font-mono text-sm px-4 py-3 pl-10 rounded-lg border border-graphite-800/80 focus:border-emerald-500/50 focus:ring-1 focus:ring-emerald-500/30 outline-none transition-all"
          />
          <Search className="absolute left-3.5 top-3.5 w-4 h-4 text-graphite-500" />
          
          {history.length > 0 && showHistory && (
            <div className="absolute top-full left-0 right-0 mt-1 bg-graphite-900 border border-graphite-800 rounded-lg shadow-2xl z-50 overflow-hidden divide-y divide-graphite-800/50 max-h-48 overflow-y-auto">
              <div className="px-3 py-1.5 text-[10px] font-mono text-graphite-500 flex items-center gap-1 bg-graphite-950/40">
                <History className="w-3 h-3" /> Query History
              </div>
              {history.map((h, i) => (
                <button
                  key={i}
                  type="button"
                  onMouseDown={() => handleSelectHistory(h)}
                  className="w-full text-left px-4 py-2 text-xs font-mono text-graphite-300 hover:bg-graphite-800 hover:text-white transition-colors truncate"
                >
                  {h}
                </button>
              ))}
            </div>
          )}
        </div>
        <button
          type="submit"
          className="px-5 py-3 bg-emerald-500 text-emerald-950 font-bold rounded-lg shadow-md hover:bg-emerald-400 transition-all flex items-center gap-2 shrink-0 text-sm hover:-translate-y-0.5 active:translate-y-0"
        >
          <Search className="w-4 h-4" /> Run Query
        </button>
      </form>

      {error && (
        <div className="flex items-center gap-2 text-red-400 text-xs font-mono bg-red-950/20 px-3 py-1.5 rounded-lg border border-red-500/10 animate-fade-in">
          <AlertCircle className="w-4 h-4 shrink-0" />
          <span>Invalid syntax: {error}</span>
        </div>
      )}
    </div>
  );
}

// =========================================================================
// LIGHTWEIGHT KQL PARSER ENGINE
// Supports:
// - Operators: ==, !=, contains, !contains, startswith, endswith, has, >, >=, <, <=
// - Logical operations: AND, OR, NOT (lower/uppercase case insensitive)
// - Grouping via operator precedence (AND has higher precedence than OR)
// =========================================================================

function evaluateQuery(log: any, queryStr: string): boolean {
  // Simple check for plain text search
  if (!queryStr.includes("==") && !queryStr.includes("!=") && !queryStr.includes("contains") && !queryStr.includes("startswith") && !queryStr.includes("endswith") && !queryStr.includes(">") && !queryStr.includes("<")) {
    const term = queryStr.replace(/"/g, "").toLowerCase();
    return JSON.stringify(log).toLowerCase().includes(term);
  }

  // Pre-tokenize: normalize whitespace, uppercase logical operators to standard
  // E.g. "source == 'firewall' and severity == 'high'" -> "source == 'firewall' AND severity == 'high'"
  let normalized = queryStr
    .replace(/\s+and\s+/gi, " AND ")
    .replace(/\s+or\s+/gi, " OR ")
    .replace(/\s+not\s+/gi, " NOT ")
    .trim();

  // Handle OR clauses (lowest precedence)
  const orClauses = splitRespectingQuotes(normalized, " OR ");
  if (orClauses.length > 1) {
    return orClauses.some(clause => evaluateQuery(log, clause));
  }

  // Handle AND clauses (higher precedence)
  const andClauses = splitRespectingQuotes(normalized, " AND ");
  if (andClauses.length > 1) {
    return andClauses.every(clause => evaluateQuery(log, clause));
  }

  // Handle negation
  if (normalized.startsWith("NOT ")) {
    return !evaluateQuery(log, normalized.substring(4));
  }

  // Evaluate single condition
  return evaluateCondition(log, normalized);
}

function splitRespectingQuotes(str: string, delimiter: string): string[] {
  const result: string[] = [];
  let inQuotes = false;
  let quoteChar = "";
  let currentToken = "";

  for (let i = 0; i < str.length; i++) {
    const char = str[i];
    if ((char === '"' || char === "'") && (i === 0 || str[i - 1] !== '\\')) {
      if (inQuotes && char === quoteChar) {
        inQuotes = false;
      } else if (!inQuotes) {
        inQuotes = true;
        quoteChar = char;
      }
    }

    currentToken += char;

    // Check if we are at delimiter boundary outside of quotes
    if (!inQuotes && str.startsWith(delimiter, i + 1)) {
      result.push(currentToken.trim());
      currentToken = "";
      i += delimiter.length - 1; // skip delimiter
    }
  }
  
  if (currentToken) {
    result.push(currentToken.trim());
  }
  return result;
}

function evaluateCondition(log: any, conditionStr: string): boolean {
  // Regex to extract: key, operator, and value (supporting quoted strings or words)
  const regex = /^([a-zA-Z_][a-zA-Z0-9_]*)\s*(==|!=|contains|!contains|startswith|endswith|>=|<=|>|<|has)\s*(.*)$/;
  const match = conditionStr.match(regex);

  if (!match) {
    throw new Error(`expected valid key-value criteria in clause "${conditionStr}"`);
  }

  const [, key, op, rawVal] = match;
  
  // Clean up value (strip quotes)
  let val = rawVal.trim();
  if ((val.startsWith('"') && val.endsWith('"')) || (val.startsWith("'") && val.endsWith("'"))) {
    val = val.substring(1, val.length - 1);
  }

  const logVal = log[key];
  if (logVal === undefined) {
    // If checking != or !contains and key doesn't exist, it evaluates to true
    if (op === "!=" || op === "!contains") return true;
    return false;
  }

  const strLogVal = String(logVal).toLowerCase();
  const strQueryVal = String(val).toLowerCase();

  switch (op) {
    case "==":
      return String(logVal) === val;
    case "!=":
      return String(logVal) !== val;
    case "contains":
      return strLogVal.includes(strQueryVal);
    case "!contains":
      return !strLogVal.includes(strQueryVal);
    case "startswith":
      return strLogVal.startsWith(strQueryVal);
    case "endswith":
      return strLogVal.endsWith(strQueryVal);
    case "has":
      // Word-level match
      const words = strLogVal.split(/[^a-zA-Z0-9]/);
      return words.includes(strQueryVal);
    case ">":
      return parseFloat(logVal) > parseFloat(val);
    case ">=":
      return parseFloat(logVal) >= parseFloat(val);
    case "<":
      return parseFloat(logVal) < parseFloat(val);
    case "<=":
      return parseFloat(logVal) <= parseFloat(val);
    default:
      return false;
  }
}
