import { useState, useEffect } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import axios from "axios";
import {
  MessageSquare,
  FileText,
  Search,
  Filter,
  Eye,
  ThumbsUp,
  Pin,
  ArrowRight,
  Plus,
  Tag as TagIcon,
  Crosshair,
} from "lucide-react";

interface ThreadAuthor {
  user_id: string;
  nickname: string;
  role: string;
}

interface ThreadScenario {
  scenario_id: string;
  title: string;
  domain_id: string;
  mitre_technique?: string;
}

interface Thread {
  thread_id: string;
  author: ThreadAuthor;
  scenario?: ThreadScenario;
  title: string;
  content: string;
  tags: string[];
  upvote_count: number;
  view_count: number;
  is_pinned: boolean;
  is_locked: boolean;
  created_at: string;
  updated_at: string;
}

export default function DiscussionsFeedPage() {
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  const currentScenarioId = searchParams.get("scenario_id") || "";
  const currentTag = searchParams.get("tag") || "";
  const currentSort = searchParams.get("sort") || "recent";

  const userRole = localStorage.getItem("userRole") || "learner";
  const isSpecialistOrAdmin = userRole === "specialist" || userRole === "admin";

  const [threads, setThreads] = useState<Thread[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");

  useEffect(() => {
    fetchThreads();
  }, [currentScenarioId, currentTag, currentSort]);

  const fetchThreads = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (currentScenarioId) params.append("scenario_id", currentScenarioId);
      if (currentTag) params.append("tag", currentTag);
      if (currentSort) params.append("sort", currentSort);

      const res = await axios.get(`/api/threads?${params.toString()}`);
      setThreads(res.data.threads || []);
    } catch (err) {
      console.error("Failed to load threads:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleSortChange = (newSort: string) => {
    const params = new URLSearchParams(searchParams);
    params.set("sort", newSort);
    setSearchParams(params);
  };

  const handleTagClick = (tag: string) => {
    const params = new URLSearchParams(searchParams);
    if (params.get("tag") === tag) {
      params.delete("tag");
    } else {
      params.set("tag", tag);
    }
    setSearchParams(params);
  };

  const clearFilters = () => {
    setSearchParams({});
    setSearchQuery("");
  };

  const filteredThreads = threads.filter((t) => {
    if (!searchQuery) return true;
    const q = searchQuery.toLowerCase();
    return (
      t.title.toLowerCase().includes(q) ||
      t.content.toLowerCase().includes(q) ||
      t.author.nickname.toLowerCase().includes(q) ||
      (t.tags && t.tags.some((tag) => tag.toLowerCase().includes(q)))
    );
  });

  return (
    <div className="space-y-8 animate-fade-in max-w-5xl mx-auto">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-graphite-800 pb-6">
        <div>
          <div className="flex items-center gap-2.5 mb-1">
            <MessageSquare className="w-6 h-6 text-indigo-400" />
            <h1 className="text-2xl font-bold tracking-tight text-white">Discussions & Specialist Insights</h1>
          </div>
          <p className="text-xs text-graphite-400">
            Peer learning, scenario debriefs, and deep-dive attack post-mortems authored by Cyber Specialists.
          </p>
        </div>

        {isSpecialistOrAdmin && (
          <button
            onClick={() => navigate("/discussions/new")}
            className="flex items-center gap-2 px-4 py-2 rounded-md bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs shadow-sm transition-all cursor-pointer self-start sm:self-auto"
          >
            <Plus className="w-4 h-4" />
            <span>New Analysis Post</span>
          </button>
        )}
      </div>

      {/* Filter / Search Bar */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 p-4 rounded-lg bg-graphite-900/50 border border-graphite-800 text-xs">
        {/* Search */}
        <div className="relative">
          <Search className="w-3.5 h-3.5 text-graphite-500 absolute left-3 top-1/2 -translate-y-1/2" />
          <input
            type="text"
            placeholder="Search discussion topics, tags..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-9 pr-3 py-1.5 rounded bg-graphite-950 border border-graphite-800 text-graphite-200 placeholder-graphite-500 focus:outline-none focus:border-graphite-600 text-xs font-mono"
          />
        </div>

        {/* Sort */}
        <div className="flex items-center gap-2">
          <Filter className="w-3.5 h-3.5 text-graphite-500 shrink-0" />
          <select
            value={currentSort}
            onChange={(e) => handleSortChange(e.target.value)}
            className="w-full px-2.5 py-1.5 rounded bg-graphite-950 border border-graphite-800 text-graphite-200 focus:outline-none focus:border-graphite-600 text-xs"
          >
            <option value="recent">Sort by Most Recent</option>
            <option value="popular">Sort by Most Helpful (Popular)</option>
          </select>
        </div>

        {/* Active Filter Indicators */}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-1.5 overflow-x-auto">
            {currentTag && (
              <span className="px-2 py-1 rounded bg-indigo-950/40 border border-indigo-500/30 text-indigo-300 font-mono text-[11px] flex items-center gap-1">
                <TagIcon className="w-3 h-3" />
                {currentTag}
              </span>
            )}
            {currentScenarioId && (
              <span className="px-2 py-1 rounded bg-amber-950/40 border border-amber-500/30 text-amber-300 font-mono text-[11px] flex items-center gap-1">
                <Crosshair className="w-3 h-3" />
                Scenario Filtered
              </span>
            )}
          </div>

          {(currentTag || currentScenarioId || searchQuery) && (
            <button
              onClick={clearFilters}
              className="text-[11px] text-graphite-400 hover:text-white underline underline-offset-2 ml-auto cursor-pointer"
            >
              Reset filters
            </button>
          )}
        </div>
      </div>

      {/* Feed List */}
      {loading ? (
        <div className="py-20 text-center text-graphite-500 font-mono text-xs">
          Loading discussion threads...
        </div>
      ) : filteredThreads.length === 0 ? (
        <div className="py-16 text-center border border-dashed border-graphite-800 rounded-lg space-y-3">
          <FileText className="w-8 h-8 text-graphite-600 mx-auto" />
          <p className="text-sm font-semibold text-graphite-300">No discussion posts yet</p>
          <p className="text-xs text-graphite-500 max-w-sm mx-auto">
            {isSpecialistOrAdmin
              ? "Be the first to share an attack analysis or playbook debrief for the community!"
              : "Check back soon for expert analysis and mitigation breakdowns."}
          </p>
          {isSpecialistOrAdmin && (
            <button
              onClick={() => navigate("/discussions/new")}
              className="mt-2 px-4 py-2 rounded-md bg-white text-black font-semibold text-xs hover:bg-white/90 transition-all cursor-pointer"
            >
              Create First Post
            </button>
          )}
        </div>
      ) : (
        <div className="space-y-3">
          {filteredThreads.map((thread) => (
            <div
              key={thread.thread_id}
              onClick={() => navigate(`/discussions/${thread.thread_id}`)}
              className="group p-5 rounded-lg border border-graphite-800 hover:border-graphite-700 bg-graphite-950 hover:bg-graphite-900/40 transition-all cursor-pointer space-y-3"
            >
              {/* Top metadata */}
              <div className="flex flex-wrap items-center justify-between gap-2 text-xs">
                <div className="flex items-center gap-2">
                  {thread.is_pinned && (
                    <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20 text-[10px] font-mono font-bold">
                      <Pin className="w-3 h-3" /> PINNED
                    </span>
                  )}

                  {/* Author badge */}
                  <div className="flex items-center gap-1.5">
                    <span className="font-mono text-xs font-medium text-graphite-200">
                      {thread.author.nickname}
                    </span>
                    {thread.author.role === "specialist" && (
                      <span className="px-1.5 py-0.2 rounded text-[9px] font-mono font-bold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                        SPECIALIST
                      </span>
                    )}
                    {thread.author.role === "admin" && (
                      <span className="px-1.5 py-0.2 rounded text-[9px] font-mono font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20">
                        ADMIN
                      </span>
                    )}
                  </div>

                  <span className="text-graphite-600">•</span>
                  <span className="text-graphite-500 text-[11px] font-mono">
                    {new Date(thread.created_at).toLocaleDateString()}
                  </span>
                </div>

                {/* Linked Scenario Badge */}
                {thread.scenario && (
                  <div className="flex items-center gap-1.5 px-2 py-0.5 rounded bg-graphite-900 border border-graphite-800 text-[11px] text-graphite-300">
                    <Crosshair className="w-3 h-3 text-indigo-400" />
                    <span className="font-medium truncate max-w-xs">{thread.scenario.title}</span>
                    <span className="text-graphite-500 font-mono text-[9px]">{thread.scenario.domain_id}</span>
                  </div>
                )}
              </div>

              {/* Title & Preview */}
              <div>
                <h3 className="text-base font-bold text-white group-hover:text-indigo-300 transition-colors">
                  {thread.title}
                </h3>
                <p className="text-xs text-graphite-400 line-clamp-2 mt-1 leading-relaxed">
                  {thread.content.replace(/[#*`>|]/g, "").slice(0, 180)}...
                </p>
              </div>

              {/* Tags & Bottom Stats */}
              <div className="flex flex-wrap items-center justify-between gap-3 pt-2 border-t border-graphite-800/60 text-xs">
                {/* Tags */}
                <div className="flex flex-wrap gap-1.5">
                  {thread.tags &&
                    thread.tags.map((tag) => (
                      <span
                        key={tag}
                        onClick={(e) => {
                          e.stopPropagation();
                          handleTagClick(tag);
                        }}
                        className="px-2 py-0.5 rounded bg-graphite-900 hover:bg-graphite-800 border border-graphite-800 text-graphite-300 text-[10px] font-mono transition-colors"
                      >
                        #{tag}
                      </span>
                    ))}
                </div>

                {/* Counters */}
                <div className="flex items-center gap-4 text-graphite-400 font-mono text-xs">
                  <span className="flex items-center gap-1">
                    <ThumbsUp className="w-3.5 h-3.5" />
                    {thread.upvote_count}
                  </span>
                  <span className="flex items-center gap-1">
                    <Eye className="w-3.5 h-3.5" />
                    {thread.view_count}
                  </span>
                  <span className="text-indigo-400 font-medium group-hover:translate-x-0.5 transition-transform flex items-center gap-0.5">
                    Read <ArrowRight className="w-3 h-3" />
                  </span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
