import { useState, useEffect } from "react";
import { 
  User, Shield, Award, Clock, AlertTriangle, CheckCircle2, 
  Upload, Edit2, Loader2, ArrowRight 
} from "lucide-react";
import axios from "axios";

interface SpecialistApplication {
  application_id: string;
  status: "pending" | "approved" | "rejected";
  bio: string;
  resume_path: string;
  certificate_path?: string;
  linkedin_url?: string;
  portfolio_url?: string;
  rejection_reason?: string;
  created_at: string;
  reviewed_at?: string;
}

interface UserProfile {
  user_id: string;
  email: string;
  nickname: string;
  role: "learner" | "specialist" | "admin";
  current_tier: number;
  latest_application?: SpecialistApplication;
}

export default function ProfilePage() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  // Edit nickname state
  const [isEditingNickname, setIsEditingNickname] = useState(false);
  const [newNickname, setNewNickname] = useState("");
  const [savingNickname, setSavingNickname] = useState(false);

  // Specialist application form state
  const [showApplyForm, setShowApplyForm] = useState(false);
  const [bio, setBio] = useState("");
  const [resumeFile, setResumeFile] = useState<File | null>(null);
  const [certFile, setCertFile] = useState<File | null>(null);
  const [linkedinUrl, setLinkedinUrl] = useState("");
  const [portfolioUrl, setPortfolioUrl] = useState("");
  const [submittingApp, setSubmittingApp] = useState(false);

  const token = localStorage.getItem("token");

  const fetchProfile = async () => {
    if (!token) return;
    try {
      setLoading(true);
      const res = await axios.get("/api/user/profile", {
        headers: { Authorization: `Bearer ${token}` }
      });
      setProfile(res.data);
      setNewNickname(res.data.nickname);
      
      // Update local storage
      localStorage.setItem("userNickname", res.data.nickname);
      localStorage.setItem("userRole", res.data.role);

      // If rejected application exists, prefill fields for easy re-apply
      if (res.data.latest_application?.status === "rejected") {
        setBio(res.data.latest_application.bio || "");
        setLinkedinUrl(res.data.latest_application.linkedin_url || "");
        setPortfolioUrl(res.data.latest_application.portfolio_url || "");
      }
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to load user profile");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchProfile();
  }, []);

  const handleUpdateNickname = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newNickname.trim()) return;
    setSavingNickname(true);
    setError("");
    setSuccess("");

    try {
      await axios.put(
        "/api/user/nickname", 
        { nickname: newNickname.trim() },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setSuccess("Nickname updated successfully!");
      setIsEditingNickname(false);
      fetchProfile();
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to update nickname");
    } finally {
      setSavingNickname(false);
    }
  };

  const handleApplySpecialist = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");

    if (!resumeFile) {
      setError("Please attach your Resume file (PDF or image).");
      return;
    }

    if (!bio.trim()) {
      setError("Please provide a professional bio summary.");
      return;
    }

    setSubmittingApp(true);
    try {
      const formData = new FormData();
      formData.append("bio", bio.trim());
      formData.append("resume", resumeFile);
      if (certFile) formData.append("certificate", certFile);
      if (linkedinUrl) formData.append("linkedin_url", linkedinUrl.trim());
      if (portfolioUrl) formData.append("portfolio_url", portfolioUrl.trim());

      await axios.post("/api/specialist/apply", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
          Authorization: `Bearer ${token}`
        }
      });

      setSuccess("Specialist application submitted successfully!");
      setShowApplyForm(false);
      fetchProfile();
    } catch (err: any) {
      setError(err.response?.data?.error || "Application submission failed");
    } finally {
      setSubmittingApp(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-[60vh] flex items-center justify-center">
        <Loader2 className="animate-spin text-white w-8 h-8" />
      </div>
    );
  }

  if (!profile) {
    return (
      <div className="p-8 text-center text-graphite-400">
        Profile unavailable. Please log in again.
      </div>
    );
  }

  const app = profile.latest_application;

  return (
    <div className="max-w-4xl mx-auto space-y-8 animate-fade-in">
      {/* Header Profile Banner */}
      <div className="bg-graphite-900 border border-graphite-800 rounded-xl p-6 sm:p-8 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-6">
        <div className="flex items-center gap-5">
          <div className="w-16 h-16 rounded-xl bg-graphite-950 border border-graphite-800 flex items-center justify-center text-white shadow-inner">
            {profile.role === "admin" ? (
              <Shield className="w-8 h-8 text-amber-400" />
            ) : profile.role === "specialist" ? (
              <Award className="w-8 h-8 text-indigo-400" />
            ) : (
              <User className="w-8 h-8 text-graphite-300" />
            )}
          </div>
          <div>
            <div className="flex items-center gap-3">
              {isEditingNickname ? (
                <form onSubmit={handleUpdateNickname} className="flex items-center gap-2">
                  <input
                    type="text"
                    value={newNickname}
                    onChange={(e) => setNewNickname(e.target.value)}
                    required
                    className="px-2.5 py-1 bg-graphite-950 border border-graphite-700 rounded text-sm text-white font-mono"
                  />
                  <button
                    type="submit"
                    disabled={savingNickname}
                    className="px-2.5 py-1 bg-white text-black text-xs font-semibold rounded hover:bg-white/90"
                  >
                    {savingNickname ? <Loader2 className="w-3 h-3 animate-spin" /> : "Save"}
                  </button>
                  <button
                    type="button"
                    onClick={() => setIsEditingNickname(false)}
                    className="px-2.5 py-1 text-xs text-graphite-400 hover:text-white"
                  >
                    Cancel
                  </button>
                </form>
              ) : (
                <div className="flex items-center gap-2">
                  <h1 className="text-xl font-bold text-white tracking-tight font-mono">
                    {profile.nickname}
                  </h1>
                  <button
                    onClick={() => setIsEditingNickname(true)}
                    className="text-graphite-400 hover:text-white transition-colors"
                    title="Edit Nickname"
                  >
                    <Edit2 size={14} />
                  </button>
                </div>
              )}
              
              <span className={`px-2 py-0.5 rounded text-[11px] font-mono font-semibold uppercase tracking-wide border ${
                profile.role === "admin"
                  ? "bg-amber-500/10 text-amber-400 border-amber-500/30"
                  : profile.role === "specialist"
                  ? "bg-indigo-500/10 text-indigo-400 border-indigo-500/30"
                  : "bg-graphite-800 text-graphite-300 border-graphite-700"
              }`}>
                {profile.role}
              </span>
            </div>
            
            <p className="text-xs text-graphite-400 font-mono mt-1">{profile.email}</p>
            <p className="text-xs text-graphite-500 mt-2">
              Current Competence Level: <span className="text-white font-semibold">Tier {profile.current_tier} Analyst</span>
            </p>
          </div>
        </div>
      </div>

      {error && (
        <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-xs font-mono">
          {error}
        </div>
      )}

      {success && (
        <div className="p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs font-mono flex items-center gap-2">
          <CheckCircle2 size={16} />
          <span>{success}</span>
        </div>
      )}

      {/* Role-Specific Workflows & Specialist Application Section */}
      <div className="bg-graphite-900 border border-graphite-800 rounded-xl p-6 sm:p-8 space-y-6">
        <h2 className="text-base font-semibold text-white tracking-tight flex items-center gap-2">
          <Award className="w-4 h-4 text-indigo-400" />
          Specialist Accreditation
        </h2>

        {/* Status A: User is already Specialist */}
        {profile.role === "specialist" && (
          <div className="p-5 rounded-lg bg-indigo-500/10 border border-indigo-500/20 space-y-2">
            <div className="flex items-center gap-2 text-indigo-400 font-medium text-sm">
              <CheckCircle2 size={18} />
              <span>Verified Security Specialist</span>
            </div>
            <p className="text-xs text-graphite-300 leading-relaxed font-light">
              Your credentials have been validated by administrators. You are authorized to review learner scenario investigations, inspect triage decisions, and author instructional Analysis Threads.
            </p>
          </div>
        )}

        {/* Status B: User is Admin */}
        {profile.role === "admin" && (
          <div className="p-5 rounded-lg bg-amber-500/10 border border-amber-500/20 space-y-2">
            <div className="flex items-center gap-2 text-amber-400 font-medium text-sm">
              <Shield size={18} />
              <span>Administrator Privileges</span>
            </div>
            <p className="text-xs text-graphite-300 leading-relaxed font-light">
              You possess global platform governance rights. You can manage user roles, audit candidate qualifications, and moderate all community discussions.
            </p>
          </div>
        )}

        {/* Status C: Learner with Active / Previous Applications */}
        {profile.role === "learner" && (
          <div className="space-y-6">
            {/* Case 1: Application is PENDING */}
            {app && app.status === "pending" && (
              <div className="p-5 rounded-lg bg-amber-500/5 border border-amber-500/20 space-y-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 text-amber-400 font-medium text-sm">
                    <Clock size={16} />
                    <span>Application Under Review</span>
                  </div>
                  <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-amber-500/20 text-amber-300">
                    PENDING
                  </span>
                </div>
                <p className="text-xs text-graphite-300 leading-relaxed font-light">
                  Your specialist application was submitted on {new Date(app.created_at).toLocaleDateString()}. Platform administrators are reviewing your submitted credentials. You can continue taking exams and practicing simulation scenarios as a Learner in the meantime.
                </p>
                <div className="bg-graphite-950 p-3 rounded border border-graphite-800 text-xs text-graphite-400 font-mono">
                  <span className="text-graphite-500">Submitted Bio:</span> {app.bio}
                </div>
              </div>
            )}

            {/* Case 2: Application is REJECTED */}
            {app && app.status === "rejected" && !showApplyForm && (
              <div className="p-5 rounded-lg bg-rose-500/10 border border-rose-500/20 space-y-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 text-rose-400 font-medium text-sm">
                    <AlertTriangle size={16} />
                    <span>Specialist Application Declined</span>
                  </div>
                  <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-rose-500/20 text-rose-300">
                    DECLINED
                  </span>
                </div>
                <div className="p-3 bg-graphite-950 rounded border border-rose-500/30 text-xs text-graphite-200">
                  <span className="text-rose-400 font-semibold block mb-1">Administrator Feedback:</span>
                  <p className="font-mono text-xs">{app.rejection_reason || "Qualifications did not meet current specialist criteria."}</p>
                </div>
                <p className="text-xs text-graphite-400 font-light">
                  You can revise your information, upload updated certifications or a revised resume, and re-apply immediately.
                </p>
                <button
                  type="button"
                  onClick={() => setShowApplyForm(true)}
                  className="px-4 py-2 bg-white text-black font-semibold rounded text-xs hover:bg-white/90 transition-all flex items-center gap-2 cursor-pointer"
                >
                  <span>Edit & Re-apply</span>
                  <ArrowRight size={14} />
                </button>
              </div>
            )}

            {/* Case 3: No Application yet */}
            {(!app || (app.status === "rejected" && showApplyForm)) && (
              <div className="space-y-4">
                {!showApplyForm && !app && (
                  <div className="p-5 rounded-lg bg-graphite-950 border border-graphite-800 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                    <div>
                      <h3 className="text-sm font-semibold text-white">Apply to Become a Specialist</h3>
                      <p className="text-xs text-graphite-400 mt-1 font-light">
                        Share your industry expertise, guide junior blue teamers, and author analysis breakdowns for simulations.
                      </p>
                    </div>
                    <button
                      onClick={() => setShowApplyForm(true)}
                      className="px-4 py-2 bg-white text-black font-semibold rounded text-xs hover:bg-white/90 transition-all whitespace-nowrap cursor-pointer"
                    >
                      Start Application
                    </button>
                  </div>
                )}

                {/* Application Form */}
                {showApplyForm && (
                  <form onSubmit={handleApplySpecialist} className="p-6 bg-graphite-950 border border-graphite-800 rounded-lg space-y-4 animate-fade-in">
                    <div className="flex justify-between items-center border-b border-graphite-800 pb-3">
                      <h3 className="text-sm font-semibold text-white">
                        {app?.status === "rejected" ? "Revise & Re-apply for Specialist" : "Specialist Accreditation Application"}
                      </h3>
                      <button
                        type="button"
                        onClick={() => setShowApplyForm(false)}
                        className="text-xs text-graphite-400 hover:text-white"
                      >
                        Cancel
                      </button>
                    </div>

                    <div>
                      <label className="block text-xs font-medium text-graphite-300 mb-1">
                        Professional Bio / Experience Summary <span className="text-red-400">*</span>
                      </label>
                      <textarea
                        value={bio}
                        onChange={(e) => setBio(e.target.value)}
                        required
                        rows={3}
                        className="w-full px-3 py-2 bg-graphite-900 border border-graphite-800 rounded text-xs text-white placeholder-graphite-600 focus:outline-none focus:border-white transition-all font-light"
                        placeholder="Detail your SOC Tier 2/3, Incident Response, or threat hunting experience..."
                      />
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div>
                        <label className="block text-xs font-medium text-graphite-300 mb-1">
                          Resume (PDF/Image) <span className="text-red-400">*</span>
                        </label>
                        <label className="flex flex-col items-center justify-center p-4 border border-dashed border-graphite-700 hover:border-white rounded bg-graphite-900 cursor-pointer transition-colors text-center">
                          <Upload size={18} className="text-graphite-400 mb-1" />
                          <span className="text-xs text-graphite-300 truncate max-w-full">
                            {resumeFile ? resumeFile.name : "Attach Resume (<10MB)"}
                          </span>
                          <input
                            type="file"
                            accept=".pdf,.png,.jpg,.jpeg"
                            onChange={(e) => e.target.files && setResumeFile(e.target.files[0])}
                            className="hidden"
                          />
                        </label>
                      </div>

                      <div>
                        <label className="block text-xs font-medium text-graphite-300 mb-1">
                          Certificate (Optional)
                        </label>
                        <label className="flex flex-col items-center justify-center p-4 border border-dashed border-graphite-700 hover:border-white rounded bg-graphite-900 cursor-pointer transition-colors text-center">
                          <Upload size={18} className="text-graphite-400 mb-1" />
                          <span className="text-xs text-graphite-300 truncate max-w-full">
                            {certFile ? certFile.name : "Attach Certificate"}
                          </span>
                          <input
                            type="file"
                            accept=".pdf,.png,.jpg,.jpeg"
                            onChange={(e) => e.target.files && setCertFile(e.target.files[0])}
                            className="hidden"
                          />
                        </label>
                      </div>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div>
                        <label className="block text-xs font-medium text-graphite-300 mb-1">
                          LinkedIn URL
                        </label>
                        <input
                          type="url"
                          value={linkedinUrl}
                          onChange={(e) => setLinkedinUrl(e.target.value)}
                          className="w-full px-3 py-2 bg-graphite-900 border border-graphite-800 rounded text-xs text-white placeholder-graphite-600 focus:outline-none focus:border-white transition-all font-light"
                          placeholder="https://linkedin.com/in/..."
                        />
                      </div>

                      <div>
                        <label className="block text-xs font-medium text-graphite-300 mb-1">
                          Portfolio / Credly Link
                        </label>
                        <input
                          type="url"
                          value={portfolioUrl}
                          onChange={(e) => setPortfolioUrl(e.target.value)}
                          className="w-full px-3 py-2 bg-graphite-900 border border-graphite-800 rounded text-xs text-white placeholder-graphite-600 focus:outline-none focus:border-white transition-all font-light"
                          placeholder="https://credly.com/..."
                        />
                      </div>
                    </div>

                    <div className="flex justify-end gap-3 pt-2">
                      <button
                        type="button"
                        onClick={() => setShowApplyForm(false)}
                        className="px-4 py-2 border border-graphite-800 text-graphite-300 rounded text-xs hover:bg-graphite-900"
                      >
                        Cancel
                      </button>
                      <button
                        type="submit"
                        disabled={submittingApp}
                        className="px-5 py-2 bg-white text-black font-semibold rounded text-xs hover:bg-white/90 flex items-center gap-2 disabled:opacity-50 cursor-pointer"
                      >
                        {submittingApp ? <Loader2 className="w-4 h-4 animate-spin" /> : "Submit Application"}
                      </button>
                    </div>
                  </form>
                )}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
