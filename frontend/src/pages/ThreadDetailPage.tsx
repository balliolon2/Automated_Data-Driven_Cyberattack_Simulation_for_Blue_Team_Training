import { useState, useEffect } from "react";
import { useParams, useNavigate, Link } from "react-router-dom";
import axios from "axios";
import {
  ArrowLeft,
  Pin,
  Eye,
  ThumbsUp,
  Award,
  ShieldAlert,
  Crosshair,
  Edit,
  Trash2,
  Share2,
  Calendar,
  Loader2,
} from "lucide-react";
import MarkdownView from "../components/MarkdownView";
import RoleBadge from "../components/RoleBadge";
import CommentSection from "../components/CommentSection";
import { cn } from "../lib/utils";

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

interface ThreadDetail {
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
  user_has_upvoted?: boolean;
  created_at: string;
  updated_at: string;
}

export default function ThreadDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const token = localStorage.getItem("token");
  const currentUserId = localStorage.getItem("userId");
  const currentUserRole = localStorage.getItem("userRole") || "learner";

  const [thread, setThread] = useState<ThreadDetail | null>(null);
  const [loading, setLoading] = useState(true);
  const [deleteLoading, setDeleteLoading] = useState(false);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [copied, setCopied] = useState(false);

  // Upvote state
  const [threadUpvoted, setThreadUpvoted] = useState(false);
  const [threadUpvoteCount, setThreadUpvoteCount] = useState(0);
  const [upvoteLoading, setUpvoteLoading] = useState(false);

  useEffect(() => {
    if (id) {
      fetchThread(id);
    }
  }, [id]);

  const fetchThread = async (threadId: string) => {
    setLoading(true);
    try {
      const headers: Record<string, string> = {};
      if (token) headers["Authorization"] = `Bearer ${token}`;

      const res = await axios.get(`/api/threads/${threadId}`, { headers });
      setThread(res.data);
      setThreadUpvoteCount(res.data.upvote_count || 0);
      setThreadUpvoted(Boolean(res.data.user_has_upvoted));
    } catch (err) {
      console.error("Failed to load thread:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleToggleThreadUpvote = async () => {
    if (!token) {
      navigate("/login");
      return;
    }
    if (!id || upvoteLoading) return;
    setUpvoteLoading(true);
    try {
      const res = await axios.post(
        `/api/threads/${id}/upvote`,
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setThreadUpvoted(Boolean(res.data.has_upvoted ?? res.data.upvoted));
      setThreadUpvoteCount(res.data.upvote_count);
    } catch (err) {
      console.error("Failed to toggle upvote:", err);
    } finally {
      setUpvoteLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!id || !token) return;
    setDeleteLoading(true);
    try {
      await axios.delete(`/api/threads/${id}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      navigate("/discussions");
    } catch (err) {
      console.error("Failed to delete thread:", err);
      alert("Failed to delete thread.");
    } finally {
      setDeleteLoading(false);
      setShowDeleteConfirm(false);
    }
  };

  const handleShare = () => {
    navigator.clipboard.writeText(window.location.href);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const isAuthorOrAdmin =
    Boolean(thread && currentUserId && thread.author.user_id === currentUserId) ||
    currentUserRole === "admin";

  if (loading) {
    return (
      <div className="py-24 text-center text-graphite-500 font-mono text-xs">
        <Loader2 className="w-5 h-5 animate-spin mx-auto mb-2 text-indigo-400" />
        Loading analysis post...
      </div>
    );
  }

  if (!thread) {
    return (
      <div className="max-w-xl mx-auto text-center py-20 space-y-4">
        <h2 className="text-xl font-bold text-white">Analysis Thread Not Found</h2>
        <p className="text-xs text-graphite-400">
          This post may have been removed or does not exist.
        </p>
        <button
          onClick={() => navigate("/discussions")}
          className="px-4 py-2 rounded-md bg-graphite-900 hover:bg-graphite-800 text-graphite-200 border border-graphite-800 font-semibold text-xs transition-colors cursor-pointer"
        >
          Back to Discussions
        </button>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-8 animate-fade-in pb-16">
      {/* Top Bar Navigation */}
      <div className="flex items-center justify-between border-b border-graphite-800 pb-4">
        <button
          onClick={() => navigate("/discussions")}
          className="inline-flex items-center gap-2 text-xs font-medium text-graphite-400 hover:text-white transition-colors cursor-pointer"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Back to Discussions</span>
        </button>

        <div className="flex items-center gap-2">
          <button
            onClick={handleToggleThreadUpvote}
            disabled={upvoteLoading}
            className={cn(
              "flex items-center gap-1.5 px-3 py-1.5 rounded-md border text-xs font-mono transition-all cursor-pointer",
              threadUpvoted
                ? "bg-emerald-950/60 text-emerald-300 border-emerald-500/50 shadow-sm"
                : "bg-graphite-900 hover:bg-graphite-800 text-graphite-300 hover:text-white border-graphite-800"
            )}
          >
            <ThumbsUp className={cn("w-3.5 h-3.5", threadUpvoted ? "fill-emerald-400 text-emerald-400" : "")} />
            <span>Helpful ({threadUpvoteCount})</span>
          </button>

          <button
            onClick={handleShare}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-800 text-graphite-300 hover:text-white hover:bg-graphite-900 text-xs transition-colors cursor-pointer"
          >
            <Share2 className="w-3.5 h-3.5" />
            <span>{copied ? "Copied Link!" : "Share"}</span>
          </button>

          {isAuthorOrAdmin && (
            <>
              <Link
                to={`/discussions/${thread.thread_id}/edit`}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-graphite-800 text-graphite-300 hover:text-white hover:bg-graphite-900 text-xs transition-colors"
              >
                <Edit className="w-3.5 h-3.5" />
                <span>Edit</span>
              </Link>

              <button
                onClick={() => setShowDeleteConfirm(true)}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-red-500/30 text-red-400 hover:bg-red-950/40 text-xs transition-colors cursor-pointer"
              >
                <Trash2 className="w-3.5 h-3.5" />
                <span>Delete</span>
              </button>
            </>
          )}
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      {showDeleteConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm animate-fade-in">
          <div className="max-w-sm w-full bg-graphite-950 border border-graphite-800 rounded-xl p-5 space-y-4 shadow-2xl">
            <h3 className="text-base font-bold text-white">Delete Analysis Post?</h3>
            <p className="text-xs text-graphite-400 leading-relaxed">
              Are you sure you want to delete this analysis thread? This action cannot be undone.
            </p>
            <div className="flex items-center justify-end gap-2 pt-2">
              <button
                onClick={() => setShowDeleteConfirm(false)}
                className="px-3.5 py-1.5 rounded border border-graphite-800 text-graphite-300 text-xs hover:bg-graphite-900"
              >
                Cancel
              </button>
              <button
                onClick={handleDelete}
                disabled={deleteLoading}
                className="px-4 py-1.5 rounded bg-red-600 hover:bg-red-500 text-white font-semibold text-xs transition-colors"
              >
                {deleteLoading ? "Deleting..." : "Confirm Delete"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Post Article Header */}
      <div className="space-y-4">
        {thread.is_pinned && (
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20 text-[10px] font-mono font-bold">
            <Pin className="w-3 h-3" /> PINNED BY SPECIALIST
          </span>
        )}

        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-white leading-snug">
          {thread.title}
        </h1>

        {/* Author & Meta Row */}
        <div className="flex flex-wrap items-center justify-between gap-4 py-3 border-y border-graphite-800/80 text-xs">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-full bg-graphite-900 border border-graphite-800 flex items-center justify-center">
              {thread.author.role === "admin" ? (
                <ShieldAlert className="w-4 h-4 text-amber-400" />
              ) : (
                <Award className="w-4 h-4 text-indigo-400" />
              )}
            </div>

            <div>
              <div className="flex items-center gap-1.5">
                <span className="font-mono font-medium text-white text-xs">
                  {thread.author.nickname}
                </span>
                <RoleBadge role={thread.author.role} />
              </div>
              <div className="flex items-center gap-1.5 text-graphite-500 text-[11px] font-mono mt-0.5">
                <Calendar className="w-3 h-3" />
                <span>Published {new Date(thread.created_at).toLocaleDateString()}</span>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-4 text-graphite-400 font-mono text-xs">
            <span className="flex items-center gap-1">
              <Eye className="w-3.5 h-3.5" />
              {thread.view_count} views
            </span>
            <button
              onClick={handleToggleThreadUpvote}
              disabled={upvoteLoading}
              className={cn(
                "flex items-center gap-1 px-2 py-0.5 rounded border text-xs transition-colors cursor-pointer",
                threadUpvoted
                  ? "bg-emerald-950/40 text-emerald-300 border-emerald-500/40 font-bold"
                  : "text-graphite-400 hover:text-white border-graphite-800 hover:bg-graphite-900"
              )}
            >
              <ThumbsUp className={cn("w-3.5 h-3.5", threadUpvoted ? "fill-emerald-400 text-emerald-400" : "")} />
              <span>{threadUpvoteCount} helpful</span>
            </button>
          </div>
        </div>
      </div>

      {/* Linked Scenario Banner (if applicable) */}
      {thread.scenario && (
        <div className="p-4 rounded-lg bg-indigo-950/20 border border-indigo-500/30 flex items-center justify-between gap-4">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-indigo-500/10 border border-indigo-500/30 flex items-center justify-center shrink-0">
              <Crosshair className="w-4 h-4 text-indigo-400" />
            </div>
            <div>
              <div className="text-[10px] font-mono text-indigo-400 uppercase font-semibold">
                LINKED SCENARIO DEBRIEF
              </div>
              <div className="text-sm font-bold text-white mt-0.5">{thread.scenario.title}</div>
            </div>
          </div>

          <span className="px-2 py-1 rounded bg-graphite-900 border border-graphite-800 text-[10px] font-mono text-graphite-300 shrink-0">
            {thread.scenario.domain_id}
          </span>
        </div>
      )}

      {/* Markdown Content */}
      <article className="py-2 text-graphite-200">
        <MarkdownView content={thread.content} />
      </article>

      {/* Tags */}
      {thread.tags && thread.tags.length > 0 && (
        <div className="pt-6 border-t border-graphite-800 flex flex-wrap items-center gap-2">
          <span className="text-xs text-graphite-500 font-mono">TAGS:</span>
          {thread.tags.map((tag) => (
            <Link
              key={tag}
              to={`/discussions?tag=${tag}`}
              className="px-2.5 py-1 rounded-md bg-graphite-900 hover:bg-graphite-800 border border-graphite-800 text-graphite-300 text-xs font-mono transition-colors"
            >
              #{tag}
            </Link>
          ))}
        </div>
      )}

      {/* Thread Discussion & Comments (Spec 3) */}
      <CommentSection
        threadId={thread.thread_id}
        threadAuthorId={thread.author.user_id}
      />
    </div>
  );
}
