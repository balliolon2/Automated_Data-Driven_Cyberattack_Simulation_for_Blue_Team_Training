import { useState, useEffect } from "react";
import axios from "axios";
import {
  MessageSquare,
  ThumbsUp,
  Pin,
  Reply,
  Edit2,
  Trash2,
  EyeOff,
  Send,
  Loader2,
  AlertCircle,
  X,
} from "lucide-react";
import MarkdownView from "./MarkdownView";
import RoleBadge from "./RoleBadge";
import { cn } from "../lib/utils";

interface CommentAuthor {
  user_id: string;
  nickname: string;
  role: string;
}

interface CommentItem {
  comment_id: string;
  thread_id: string;
  author: CommentAuthor;
  parent_comment_id?: string;
  content: string;
  upvote_count: number;
  is_pinned: boolean;
  is_hidden: boolean;
  has_upvoted: boolean;
  replies: CommentItem[];
  created_at: string;
  updated_at: string;
}

interface CommentSectionProps {
  threadId: string;
  threadAuthorId: string;
}

export default function CommentSection({ threadId, threadAuthorId }: CommentSectionProps) {
  const token = localStorage.getItem("token");
  const currentUserId = localStorage.getItem("userId");
  const currentUserRole = localStorage.getItem("userRole") || "learner";

  const [comments, setComments] = useState<CommentItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [newCommentContent, setNewCommentContent] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [errorMsg, setErrorMsg] = useState("");

  // Reply state
  const [replyTarget, setReplyTarget] = useState<{
    commentId: string;
    authorNickname: string;
  } | null>(null);
  const [replyContent, setReplyContent] = useState("");
  const [replySubmitting, setReplySubmitting] = useState(false);

  // Edit state
  const [editingCommentId, setEditingCommentId] = useState<string | null>(null);
  const [editContent, setEditContent] = useState("");
  const [editSubmitting, setEditSubmitting] = useState(false);

  useEffect(() => {
    fetchComments();
  }, [threadId]);

  useEffect(() => {
    if (!loading && window.location.hash) {
      setTimeout(() => {
        try {
          const el = document.querySelector(window.location.hash);
          if (el) {
            el.scrollIntoView({ behavior: "smooth", block: "center" });
          }
        } catch (_) {
          // ignore malformed selectors
        }
      }, 100);
    }
  }, [loading]);

  const fetchComments = async () => {
    setLoading(true);
    try {
      const headers: Record<string, string> = {};
      if (token) headers["Authorization"] = `Bearer ${token}`;

      const res = await axios.get(`/api/threads/${threadId}/comments`, { headers });
      setComments(res.data);
    } catch (err) {
      console.error("Failed to load comments:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateTopLevelComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!token) return;
    if (!newCommentContent.trim()) return;

    setSubmitting(true);
    setErrorMsg("");
    try {
      await axios.post(
        `/api/threads/${threadId}/comments`,
        { content: newCommentContent.trim() },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setNewCommentContent("");
      fetchComments();
    } catch (err: any) {
      console.error("Failed to post comment:", err);
      setErrorMsg(err.response?.data?.error || "Failed to post comment.");
    } finally {
      setSubmitting(false);
    }
  };

  const handleCreateReply = async (parentCommentId: string) => {
    if (!token || !replyContent.trim()) return;

    setReplySubmitting(true);
    try {
      await axios.post(
        `/api/threads/${threadId}/comments`,
        {
          content: replyContent.trim(),
          parent_comment_id: parentCommentId,
        },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setReplyContent("");
      setReplyTarget(null);
      fetchComments();
    } catch (err) {
      console.error("Failed to submit reply:", err);
    } finally {
      setReplySubmitting(false);
    }
  };

  const handleUpdateComment = async (commentId: string) => {
    if (!token || !editContent.trim()) return;

    setEditSubmitting(true);
    try {
      await axios.put(
        `/api/comments/${commentId}`,
        { content: editContent.trim() },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setEditingCommentId(null);
      setEditContent("");
      fetchComments();
    } catch (err) {
      console.error("Failed to update comment:", err);
    } finally {
      setEditSubmitting(false);
    }
  };

  const handleDeleteComment = async (commentId: string) => {
    if (!token || !window.confirm("Are you sure you want to delete this comment?")) return;

    try {
      await axios.delete(`/api/comments/${commentId}`, {
        headers: { Authorization: `Bearer ${token}` },
      });
      fetchComments();
    } catch (err) {
      console.error("Failed to delete comment:", err);
    }
  };

  const handleTogglePin = async (commentId: string) => {
    if (!token) return;
    try {
      await axios.post(
        `/api/comments/${commentId}/pin`,
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );
      fetchComments();
    } catch (err: any) {
      alert(err.response?.data?.error || "Failed to pin comment.");
    }
  };

  const handleToggleHide = async (commentId: string) => {
    if (!token) return;
    try {
      await axios.post(
        `/api/comments/${commentId}/hide`,
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );
      fetchComments();
    } catch (err) {
      console.error("Failed to toggle visibility:", err);
    }
  };

  const handleToggleUpvote = async (commentId: string) => {
    if (!token) return;
    try {
      const res = await axios.post(
        `/api/comments/${commentId}/upvote`,
        {},
        { headers: { Authorization: `Bearer ${token}` } }
      );

      // Update local state optimistically
      const updateList = (list: CommentItem[]): CommentItem[] =>
        list.map((c) => {
          if (c.comment_id === commentId) {
            return {
              ...c,
              upvote_count: res.data.upvote_count,
              has_upvoted: res.data.has_upvoted,
            };
          }
          if (c.replies && c.replies.length > 0) {
            return { ...c, replies: updateList(c.replies) };
          }
          return c;
        });

      setComments(updateList(comments));
    } catch (err) {
      console.error("Failed to upvote comment:", err);
    }
  };

  const canModerate =
    Boolean(currentUserId && threadAuthorId === currentUserId) || currentUserRole === "admin";

  return (
    <div className="space-y-6 pt-6 border-t border-graphite-800 animate-fade-in" id="comments">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <MessageSquare className="w-4 h-4 text-indigo-400" />
          <h3 className="text-base font-bold text-white">Community Discussion & Follow-ups</h3>
        </div>
        <span className="text-xs font-mono text-graphite-500">
          {comments.reduce((acc, curr) => acc + 1 + (curr.replies?.length || 0), 0)} comments
        </span>
      </div>

      {/* Top-Level Comment Input */}
      {token ? (
        <form onSubmit={handleCreateTopLevelComment} className="space-y-3">
          {errorMsg && (
            <div className="p-3 rounded-lg bg-red-950/30 border border-red-500/20 text-red-400 text-xs flex items-center gap-2">
              <AlertCircle className="w-3.5 h-3.5 shrink-0" />
              <span>{errorMsg}</span>
            </div>
          )}

          <div className="relative rounded-lg bg-graphite-950 border border-graphite-800 focus-within:border-graphite-600 transition-colors">
            <textarea
              rows={3}
              placeholder="Ask a technical question, suggest an alternative containment rule, or share insights..."
              value={newCommentContent}
              onChange={(e) => setNewCommentContent(e.target.value)}
              className="w-full p-3 bg-transparent text-xs text-graphite-200 placeholder-graphite-600 focus:outline-none resize-y"
            />
            <div className="flex items-center justify-between px-3 py-2 border-t border-graphite-800/80 text-xs">
              <span className="text-[10px] font-mono text-graphite-500">
                Markdown & code blocks supported
              </span>
              <button
                type="submit"
                disabled={submitting || !newCommentContent.trim()}
                className="flex items-center gap-1.5 px-3.5 py-1.5 rounded-md bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white font-semibold text-xs shadow-sm transition-all cursor-pointer"
              >
                {submitting ? (
                  <Loader2 className="w-3 h-3 animate-spin" />
                ) : (
                  <Send className="w-3 h-3" />
                )}
                <span>Comment</span>
              </button>
            </div>
          </div>
        </form>
      ) : (
        <div className="p-4 rounded-lg bg-graphite-950 border border-graphite-800 text-center space-y-2">
          <p className="text-xs text-graphite-400">Sign in to join the technical discussion.</p>
        </div>
      )}

      {/* Comments Listing */}
      {loading ? (
        <div className="py-12 text-center text-graphite-500 font-mono text-xs">
          Loading discussion comments...
        </div>
      ) : comments.length === 0 ? (
        <div className="py-10 text-center border border-dashed border-graphite-800/80 rounded-lg space-y-1 text-xs">
          <p className="font-medium text-graphite-300">No comments yet</p>
          <p className="text-graphite-500">Be the first to ask a question or leave feedback!</p>
        </div>
      ) : (
        <div className="space-y-4">
          {comments.map((rootComment) => (
            <div
              key={rootComment.comment_id}
              id={`comment-${rootComment.comment_id}`}
              className={cn(
                "p-4 rounded-lg border transition-all space-y-3",
                rootComment.is_pinned
                  ? "bg-amber-950/10 border-amber-500/30"
                  : "bg-graphite-950 border-graphite-800"
              )}
            >
              {/* Pinned Label */}
              {rootComment.is_pinned && (
                <div className="flex items-center gap-1 text-[10px] font-mono font-bold text-amber-400">
                  <Pin className="w-3 h-3" />
                  <span>PINNED BY AUTHOR</span>
                </div>
              )}

              {/* Comment Header */}
              <div className="flex items-center justify-between gap-2 text-xs">
                <div className="flex items-center gap-2">
                  <span className="font-mono font-medium text-white text-xs">
                    {rootComment.author.nickname}
                  </span>
                  <RoleBadge role={rootComment.author.role} />
                  <span className="text-graphite-600">•</span>
                  <span className="text-[11px] font-mono text-graphite-500">
                    {new Date(rootComment.created_at).toLocaleDateString()}
                  </span>
                </div>

                {/* Moderation Controls (Thread Author / Admin) */}
                <div className="flex items-center gap-1.5 text-xs">
                  {canModerate && (
                    <>
                      <button
                        onClick={() => handleTogglePin(rootComment.comment_id)}
                        title={rootComment.is_pinned ? "Unpin comment" : "Pin comment to top"}
                        className={cn(
                          "p-1 rounded hover:bg-graphite-800 text-xs transition-colors",
                          rootComment.is_pinned ? "text-amber-400" : "text-graphite-500 hover:text-white"
                        )}
                      >
                        <Pin className="w-3.5 h-3.5" />
                      </button>

                      <button
                        onClick={() => handleToggleHide(rootComment.comment_id)}
                        title={rootComment.is_hidden ? "Unhide comment" : "Hide comment as spam"}
                        className={cn(
                          "p-1 rounded hover:bg-graphite-800 text-xs transition-colors",
                          rootComment.is_hidden ? "text-red-400" : "text-graphite-500 hover:text-white"
                        )}
                      >
                        <EyeOff className="w-3.5 h-3.5" />
                      </button>
                    </>
                  )}

                  {(currentUserId === rootComment.author.user_id || currentUserRole === "admin") && (
                    <>
                      <button
                        onClick={() => {
                          setEditingCommentId(rootComment.comment_id);
                          setEditContent(rootComment.content);
                        }}
                        title="Edit comment"
                        className="p-1 rounded text-graphite-500 hover:text-white hover:bg-graphite-800 transition-colors"
                      >
                        <Edit2 className="w-3.5 h-3.5" />
                      </button>

                      <button
                        onClick={() => handleDeleteComment(rootComment.comment_id)}
                        title="Delete comment"
                        className="p-1 rounded text-graphite-500 hover:text-red-400 hover:bg-graphite-800 transition-colors"
                      >
                        <Trash2 className="w-3.5 h-3.5" />
                      </button>
                    </>
                  )}
                </div>
              </div>

              {/* Comment Content or Edit Box */}
              {editingCommentId === rootComment.comment_id ? (
                <div className="space-y-2 pt-1">
                  <textarea
                    rows={3}
                    value={editContent}
                    onChange={(e) => setEditContent(e.target.value)}
                    className="w-full p-2.5 rounded bg-graphite-900 border border-graphite-800 text-xs text-white focus:outline-none focus:border-graphite-600"
                  />
                  <div className="flex items-center justify-end gap-2">
                    <button
                      onClick={() => setEditingCommentId(null)}
                      className="px-2.5 py-1 rounded border border-graphite-800 text-graphite-300 text-xs hover:bg-graphite-800"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={() => handleUpdateComment(rootComment.comment_id)}
                      disabled={editSubmitting}
                      className="px-3 py-1 rounded bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs transition-colors"
                    >
                      {editSubmitting ? "Saving..." : "Save Changes"}
                    </button>
                  </div>
                </div>
              ) : (
                <div className="text-xs text-graphite-300 leading-relaxed">
                  <MarkdownView content={rootComment.content} />
                </div>
              )}

              {/* Comment Actions Bar */}
              <div className="flex items-center gap-3 pt-2 text-xs border-t border-graphite-800/60">
                {/* Helpful Upvote */}
                <button
                  onClick={() => handleToggleUpvote(rootComment.comment_id)}
                  className={cn(
                    "flex items-center gap-1 px-2 py-0.5 rounded font-mono text-[11px] transition-colors cursor-pointer",
                    rootComment.has_upvoted
                      ? "bg-emerald-950/40 text-emerald-400 border border-emerald-500/30 font-bold"
                      : "text-graphite-400 hover:text-white hover:bg-graphite-900"
                  )}
                >
                  <ThumbsUp className="w-3 h-3" />
                  <span>{rootComment.upvote_count}</span>
                </button>

                {/* Reply Toggle */}
                {token && (
                  <button
                    onClick={() => {
                      if (replyTarget?.commentId === rootComment.comment_id) {
                        setReplyTarget(null);
                      } else {
                        setReplyTarget({
                          commentId: rootComment.comment_id,
                          authorNickname: rootComment.author.nickname,
                        });
                        setReplyContent("");
                      }
                    }}
                    className="flex items-center gap-1 text-[11px] text-graphite-400 hover:text-indigo-400 transition-colors cursor-pointer"
                  >
                    <Reply className="w-3 h-3" />
                    <span>Reply</span>
                  </button>
                )}
              </div>

              {/* Nested Reply Box */}
              {replyTarget?.commentId === rootComment.comment_id && (
                <div className="mt-3 p-3 rounded-lg bg-graphite-900/50 border border-graphite-800 space-y-2">
                  <div className="flex items-center justify-between text-[11px] text-graphite-400">
                    <span>
                      Replying to <span className="text-white font-mono">@{replyTarget.authorNickname}</span>
                    </span>
                    <button
                      onClick={() => setReplyTarget(null)}
                      className="text-graphite-500 hover:text-white"
                    >
                      <X className="w-3 h-3" />
                    </button>
                  </div>
                  <textarea
                    rows={2}
                    placeholder="Write your reply..."
                    value={replyContent}
                    onChange={(e) => setReplyContent(e.target.value)}
                    className="w-full p-2 bg-graphite-950 border border-graphite-800 rounded text-xs text-white placeholder-graphite-600 focus:outline-none focus:border-graphite-600 resize-y"
                  />
                  <div className="flex items-center justify-end gap-2">
                    <button
                      onClick={() => setReplyTarget(null)}
                      className="px-2.5 py-1 rounded text-graphite-400 hover:text-white text-xs"
                    >
                      Cancel
                    </button>
                    <button
                      onClick={() => handleCreateReply(rootComment.comment_id)}
                      disabled={replySubmitting || !replyContent.trim()}
                      className="flex items-center gap-1 px-3 py-1 rounded bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs shadow-sm transition-colors cursor-pointer disabled:opacity-50"
                    >
                      {replySubmitting ? (
                        <Loader2 className="w-3 h-3 animate-spin" />
                      ) : (
                        <Send className="w-3 h-3" />
                      )}
                      <span>Send Reply</span>
                    </button>
                  </div>
                </div>
              )}

              {/* Level 2 Nested Replies */}
              {rootComment.replies && rootComment.replies.length > 0 && (
                <div className="mt-3 ml-4 sm:ml-6 pl-3 border-l-2 border-graphite-800 space-y-3">
                  {rootComment.replies.map((reply) => (
                    <div
                      key={reply.comment_id}
                      id={`comment-${reply.comment_id}`}
                      className="p-3 rounded-md bg-graphite-900/40 border border-graphite-800/80 space-y-2 text-xs"
                    >
                      <div className="flex items-center justify-between gap-2">
                        <div className="flex items-center gap-2">
                          <span className="font-mono font-medium text-white text-xs">
                            {reply.author.nickname}
                          </span>
                          <RoleBadge role={reply.author.role} />
                          <span className="text-graphite-600">•</span>
                          <span className="text-[10px] font-mono text-graphite-500">
                            {new Date(reply.created_at).toLocaleDateString()}
                          </span>
                        </div>

                        {/* Reply Author / Admin Controls */}
                        <div className="flex items-center gap-1.5">
                          {canModerate && (
                            <button
                              onClick={() => handleToggleHide(reply.comment_id)}
                              title={reply.is_hidden ? "Unhide reply" : "Hide reply as spam"}
                              className={cn(
                                "p-1 rounded hover:bg-graphite-800 text-xs transition-colors",
                                reply.is_hidden ? "text-red-400" : "text-graphite-500 hover:text-white"
                              )}
                            >
                              <EyeOff className="w-3 h-3" />
                            </button>
                          )}

                          {(currentUserId === reply.author.user_id || currentUserRole === "admin") && (
                            <>
                              <button
                                onClick={() => {
                                  setEditingCommentId(reply.comment_id);
                                  setEditContent(reply.content);
                                }}
                                title="Edit reply"
                                className="p-1 rounded text-graphite-500 hover:text-white hover:bg-graphite-800 transition-colors"
                              >
                                <Edit2 className="w-3 h-3" />
                              </button>

                              <button
                                onClick={() => handleDeleteComment(reply.comment_id)}
                                title="Delete reply"
                                className="p-1 rounded text-graphite-500 hover:text-red-400 hover:bg-graphite-800 transition-colors"
                              >
                                <Trash2 className="w-3 h-3" />
                              </button>
                            </>
                          )}
                        </div>
                      </div>

                      {/* Reply Content or Edit */}
                      {editingCommentId === reply.comment_id ? (
                        <div className="space-y-2 pt-1">
                          <textarea
                            rows={2}
                            value={editContent}
                            onChange={(e) => setEditContent(e.target.value)}
                            className="w-full p-2 rounded bg-graphite-950 border border-graphite-800 text-xs text-white focus:outline-none"
                          />
                          <div className="flex items-center justify-end gap-2">
                            <button
                              onClick={() => setEditingCommentId(null)}
                              className="px-2 py-0.5 rounded border border-graphite-800 text-graphite-400 text-xs"
                            >
                              Cancel
                            </button>
                            <button
                              onClick={() => handleUpdateComment(reply.comment_id)}
                              disabled={editSubmitting}
                              className="px-2.5 py-0.5 rounded bg-indigo-600 hover:bg-indigo-500 text-white font-semibold text-xs"
                            >
                              {editSubmitting ? "Saving..." : "Save"}
                            </button>
                          </div>
                        </div>
                      ) : (
                        <div className="text-xs text-graphite-300 leading-relaxed">
                          <MarkdownView content={reply.content} />
                        </div>
                      )}

                      {/* Reply Upvote & Sub-reply action */}
                      <div className="flex items-center gap-3 pt-1 text-xs">
                        <button
                          onClick={() => handleToggleUpvote(reply.comment_id)}
                          className={cn(
                            "flex items-center gap-1 px-1.5 py-0.5 rounded font-mono text-[10px] transition-colors cursor-pointer",
                            reply.has_upvoted
                              ? "bg-emerald-950/40 text-emerald-400 border border-emerald-500/30 font-bold"
                              : "text-graphite-400 hover:text-white"
                          )}
                        >
                          <ThumbsUp className="w-2.5 h-2.5" />
                          <span>{reply.upvote_count}</span>
                        </button>

                        {token && (
                          <button
                            onClick={() => {
                              setReplyTarget({
                                commentId: rootComment.comment_id, // 2-level normalization to root
                                authorNickname: reply.author.nickname,
                              });
                              setReplyContent(`@${reply.author.nickname} `);
                            }}
                            className="text-[10px] text-graphite-400 hover:text-indigo-400 transition-colors flex items-center gap-1 cursor-pointer"
                          >
                            <Reply className="w-2.5 h-2.5" />
                            <span>Reply</span>
                          </button>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
