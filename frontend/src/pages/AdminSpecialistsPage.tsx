import { useState, useEffect } from "react";
import { 
  ShieldAlert, CheckCircle2, Download, 
  ExternalLink, Loader2, Clock 
} from "lucide-react";
import axios from "axios";

interface SpecialistApplication {
  application_id: string;
  user_id: string;
  nickname: string;
  email: string;
  status: "pending" | "approved" | "rejected";
  bio: string;
  resume_path: string;
  certificate_path?: string;
  linkedin_url?: string;
  portfolio_url?: string;
  rejection_reason?: string;
  reviewed_by?: string;
  reviewed_at?: string;
  created_at: string;
}

export default function AdminSpecialistsPage() {
  const [applications, setApplications] = useState<SpecialistApplication[]>([]);
  const [statusFilter, setStatusFilter] = useState<string>("pending");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [actionSuccess, setActionSuccess] = useState("");

  // Reject modal state
  const [rejectingAppId, setRejectingAppId] = useState<string | null>(null);
  const [rejectionReason, setRejectionReason] = useState("");
  const [submittingAction, setSubmittingAction] = useState(false);

  const token = localStorage.getItem("token");
  const userRole = localStorage.getItem("userRole");

  const fetchApplications = async () => {
    if (!token) return;
    try {
      setLoading(true);
      setError("");
      const params = statusFilter === "all" ? {} : { status: statusFilter };
      const res = await axios.get("/api/admin/applications", {
        headers: { Authorization: `Bearer ${token}` },
        params,
      });
      setApplications(res.data);
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to load specialist applications");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchApplications();
  }, [statusFilter]);

  const handleDownload = async (appId: string, fileType: "resume" | "certificate") => {
    try {
      const response = await axios.get(`/api/admin/applications/${appId}/files/${fileType}`, {
        headers: { Authorization: `Bearer ${token}` },
        responseType: "blob",
      });
      const blob = new Blob([response.data]);
      const downloadUrl = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = downloadUrl;
      link.setAttribute("download", `${fileType}_${appId}`);
      document.body.appendChild(link);
      link.click();
      link.remove();
    } catch (err) {
      alert("Failed to download document. File may not exist on server.");
    }
  };

  const handleApprove = async (appId: string) => {
    if (!confirm("Are you sure you want to approve this candidate as a Specialist?")) return;
    setSubmittingAction(true);
    setError("");
    setActionSuccess("");

    try {
      await axios.post(`/api/admin/applications/${appId}/approve`, {}, {
        headers: { Authorization: `Bearer ${token}` },
      });
      setActionSuccess("Application approved successfully! Candidate is now a Specialist.");
      fetchApplications();
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to approve application");
    } finally {
      setSubmittingAction(false);
    }
  };

  const handleRejectSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!rejectingAppId || !rejectionReason.trim()) return;
    setSubmittingAction(true);
    setError("");
    setActionSuccess("");

    try {
      await axios.post(`/api/admin/applications/${rejectingAppId}/reject`, 
        { reason: rejectionReason.trim() },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setActionSuccess("Application rejected with feedback.");
      setRejectingAppId(null);
      setRejectionReason("");
      fetchApplications();
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to reject application");
    } finally {
      setSubmittingAction(false);
    }
  };

  if (userRole !== "admin") {
    return (
      <div className="min-h-[50vh] flex flex-col items-center justify-center space-y-4">
        <ShieldAlert className="w-12 h-12 text-rose-500" />
        <h2 className="text-xl font-bold text-white">Access Denied</h2>
        <p className="text-xs text-graphite-400">
          This portal is strictly restricted to platform administrators.
        </p>
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto space-y-8 animate-fade-in">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-graphite-800 pb-6">
        <div>
          <div className="flex items-center gap-2">
            <ShieldAlert className="w-6 h-6 text-amber-400" />
            <h1 className="text-2xl font-bold text-white tracking-tight">Specialist Accreditation Portal</h1>
          </div>
          <p className="text-xs text-graphite-400 mt-1 font-light">
            Review candidate qualifications, inspect submitted resumes/certs, and approve or decline Specialist privileges.
          </p>
        </div>

        {/* Filter Tabs */}
        <div className="flex gap-1 p-1 bg-graphite-900 border border-graphite-800 rounded-lg self-start">
          {["pending", "approved", "rejected", "all"].map((tab) => (
            <button
              key={tab}
              onClick={() => setStatusFilter(tab)}
              className={`px-3 py-1.5 rounded text-xs font-medium uppercase tracking-wider transition-all ${
                statusFilter === tab
                  ? "bg-white text-black font-semibold shadow-sm"
                  : "text-graphite-400 hover:text-white"
              }`}
            >
              {tab}
            </button>
          ))}
        </div>
      </div>

      {error && (
        <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-xs font-mono">
          {error}
        </div>
      )}

      {actionSuccess && (
        <div className="p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-mono flex items-center gap-2">
          <CheckCircle2 size={16} />
          <span>{actionSuccess}</span>
        </div>
      )}

      {/* Applications List */}
      {loading ? (
        <div className="py-20 flex justify-center items-center">
          <Loader2 className="w-8 h-8 animate-spin text-white" />
        </div>
      ) : applications.length === 0 ? (
        <div className="py-16 text-center bg-graphite-900 border border-graphite-800 rounded-xl space-y-2">
          <Clock className="w-8 h-8 text-graphite-600 mx-auto" />
          <p className="text-sm text-graphite-300 font-medium">No applications found</p>
          <p className="text-xs text-graphite-500">There are no applications matching the "{statusFilter}" status filter.</p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-6">
          {applications.map((app) => (
            <div 
              key={app.application_id}
              className="bg-graphite-900 border border-graphite-800 rounded-xl p-6 space-y-4 hover:border-graphite-700 transition-colors"
            >
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-graphite-800/60 pb-4">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-lg bg-graphite-950 border border-graphite-800 flex items-center justify-center font-mono font-bold text-white text-sm">
                    {app.nickname ? app.nickname.substring(0, 2).toUpperCase() : "SP"}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-base font-bold text-white font-mono">{app.nickname}</span>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-mono font-bold uppercase tracking-wider ${
                        app.status === "approved"
                          ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/30"
                          : app.status === "rejected"
                          ? "bg-rose-500/10 text-rose-400 border border-rose-500/30"
                          : "bg-amber-500/10 text-amber-400 border border-amber-500/30"
                      }`}>
                        {app.status}
                      </span>
                    </div>
                    <span className="text-xs text-graphite-400 font-mono">{app.email}</span>
                  </div>
                </div>

                <div className="text-xs text-graphite-500 font-mono">
                  Applied on {new Date(app.created_at).toLocaleString()}
                </div>
              </div>

              {/* Bio Section */}
              <div>
                <span className="text-[11px] font-mono uppercase tracking-wider text-graphite-400 block mb-1">
                  Candidate Bio & Experience
                </span>
                <p className="text-xs text-graphite-200 bg-graphite-950 p-3 rounded border border-graphite-800 font-light leading-relaxed whitespace-pre-wrap">
                  {app.bio}
                </p>
              </div>

              {/* Documents & External Links */}
              <div className="flex flex-wrap items-center justify-between gap-4 pt-2">
                <div className="flex flex-wrap items-center gap-3">
                  <button
                    onClick={() => handleDownload(app.application_id, "resume")}
                    className="flex items-center gap-1.5 px-3 py-1.5 bg-graphite-950 border border-graphite-700 hover:border-white rounded text-xs text-graphite-200 transition-colors cursor-pointer"
                  >
                    <Download size={13} />
                    <span>Download Resume</span>
                  </button>

                  {app.certificate_path && (
                    <button
                      onClick={() => handleDownload(app.application_id, "certificate")}
                      className="flex items-center gap-1.5 px-3 py-1.5 bg-graphite-950 border border-graphite-700 hover:border-white rounded text-xs text-graphite-200 transition-colors cursor-pointer"
                    >
                      <Download size={13} />
                      <span>Download Certificate</span>
                    </button>
                  )}

                  {app.linkedin_url && (
                    <a
                      href={app.linkedin_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-1 text-xs text-graphite-400 hover:text-white transition-colors"
                    >
                      <ExternalLink size={12} />
                      <span>LinkedIn</span>
                    </a>
                  )}

                  {app.portfolio_url && (
                    <a
                      href={app.portfolio_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-1 text-xs text-graphite-400 hover:text-white transition-colors"
                    >
                      <ExternalLink size={12} />
                      <span>Portfolio / Credly</span>
                    </a>
                  )}
                </div>

                {/* Review Controls */}
                {app.status === "pending" && (
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => setRejectingAppId(app.application_id)}
                      disabled={submittingAction}
                      className="px-3 py-1.5 bg-rose-500/10 text-rose-400 hover:bg-rose-500/20 border border-rose-500/30 rounded text-xs font-semibold transition-all cursor-pointer"
                    >
                      Reject...
                    </button>
                    <button
                      onClick={() => handleApprove(app.application_id)}
                      disabled={submittingAction}
                      className="px-4 py-1.5 bg-white text-black hover:bg-white/90 rounded text-xs font-semibold transition-all flex items-center gap-1.5 cursor-pointer shadow"
                    >
                      <CheckCircle2 size={13} />
                      <span>Approve Specialist</span>
                    </button>
                  </div>
                )}
              </div>

              {/* Display Rejection Reason if Rejected */}
              {app.status === "rejected" && app.rejection_reason && (
                <div className="p-3 bg-rose-500/5 border border-rose-500/20 rounded text-xs text-rose-300">
                  <span className="font-semibold block mb-0.5">Rejection Feedback:</span>
                  <p className="font-mono text-[11px]">{app.rejection_reason}</p>
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {/* Reject Modal Dialog */}
      {rejectingAppId && (
        <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-graphite-900 border border-graphite-800 rounded-xl max-w-md w-full p-6 space-y-4 shadow-2xl animate-fade-in">
            <div className="flex items-center justify-between">
              <h3 className="text-base font-bold text-white">Decline Specialist Application</h3>
              <button
                onClick={() => setRejectingAppId(null)}
                className="text-graphite-400 hover:text-white"
              >
                ✕
              </button>
            </div>

            <p className="text-xs text-graphite-400 font-light">
              Please provide feedback explaining why the candidate was not approved. The applicant will be able to see this reason and re-apply.
            </p>

            <form onSubmit={handleRejectSubmit} className="space-y-4">
              <div>
                <label className="block text-xs font-medium text-graphite-300 mb-1">
                  Rejection Reason <span className="text-red-400">*</span>
                </label>
                <textarea
                  value={rejectionReason}
                  onChange={(e) => setRejectionReason(e.target.value)}
                  required
                  minLength={3}
                  rows={4}
                  className="w-full px-3 py-2 bg-graphite-950 border border-graphite-800 rounded text-xs text-white placeholder-graphite-600 focus:outline-none focus:border-white font-light"
                  placeholder="e.g. Please provide proof of CompTIA Security+ or at least 2 years SOC analyst experience..."
                />
              </div>

              <div className="flex justify-end gap-3">
                <button
                  type="button"
                  onClick={() => setRejectingAppId(null)}
                  className="px-3 py-1.5 border border-graphite-800 rounded text-xs text-graphite-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={submittingAction || !rejectionReason.trim()}
                  className="px-4 py-1.5 bg-rose-600 text-white font-semibold rounded text-xs hover:bg-rose-500 disabled:opacity-50"
                >
                  {submittingAction ? <Loader2 className="w-3.5 h-3.5 animate-spin" /> : "Confirm Rejection"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
