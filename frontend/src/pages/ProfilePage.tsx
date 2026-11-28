import { useState, useEffect, useCallback } from "react";
import { Link } from "react-router-dom";
import {
  User, Shield, Award, Clock, AlertTriangle, CheckCircle2,
  Upload, Edit2, Loader2, ArrowRight, Globe, ClipboardCheck
} from "lucide-react";
import axios from "axios";
import QuestionnairePromptBanner from "../components/QuestionnairePromptBanner";
import type { QuestionnaireStatusResponse } from "../lib/questionnaire";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError } from "../i18n/format.ts";
import type { Locale } from "../i18n/types.ts";
import {
  formatDurationSeconds,
  getExamTypeBadgeClass,
  type ExamHistoryItem,
  type ExamHistoryResponse,
} from "../lib/examReview";

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
  preferred_language?: Locale;
  email_verified?: boolean;
  latest_application?: SpecialistApplication;
}

type ProfileErrorState =
  | { kind: "key"; key: string }
  | { kind: "api"; err: unknown; fallbackKey: string }
  | null;

export default function ProfilePage() {
  const { locale, setLocale, syncLocaleFromUser, t, formatDate } = useI18n();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [errorState, setErrorState] = useState<ProfileErrorState>(null);
  const [successKey, setSuccessKey] = useState<string | null>(null);

  // Edit nickname state
  const [isEditingNickname, setIsEditingNickname] = useState(false);
  const [newNickname, setNewNickname] = useState("");
  const [savingNickname, setSavingNickname] = useState(false);

  // Exam history & questionnaire state
  const [examHistory, setExamHistory] = useState<ExamHistoryItem[]>([]);
  const [questionnaireStatus, setQuestionnaireStatus] = useState<QuestionnaireStatusResponse | null>(null);

  // Specialist application form state
  const [showApplyForm, setShowApplyForm] = useState(false);
  const [bio, setBio] = useState("");
  const [resumeFile, setResumeFile] = useState<File | null>(null);
  const [certFile, setCertFile] = useState<File | null>(null);
  const [linkedinUrl, setLinkedinUrl] = useState("");
  const [portfolioUrl, setPortfolioUrl] = useState("");
  const [submittingApp, setSubmittingApp] = useState(false);

  const token = localStorage.getItem("token");

  const fetchProfile = useCallback(async () => {
    if (!token) {
      setLoading(false);
      return;
    }
    try {
      setLoading(true);
      const [res, historyRes, qStatusRes] = await Promise.allSettled([
        axios.get("/api/user/profile", {
          headers: { Authorization: `Bearer ${token}` }
        }),
        axios.get<ExamHistoryResponse>("/api/exams/history", {
          headers: { Authorization: `Bearer ${token}` }
        }),
        axios.get<QuestionnaireStatusResponse>("/api/evaluation/questionnaire/status", {
          headers: { Authorization: `Bearer ${token}` }
        })
      ]);
      if (historyRes.status === "fulfilled" && Array.isArray(historyRes.value.data?.sessions)) {
        setExamHistory(historyRes.value.data.sessions);
      }
      if (qStatusRes.status === "fulfilled" && qStatusRes.value.data) {
        setQuestionnaireStatus(qStatusRes.value.data);
      }
      if (res.status === "rejected") {
        throw res.reason;
      }
      const profileData = res.value.data;
      setProfile(profileData);
      setNewNickname(profileData.nickname);
      
      // Update local storage and sync language from DB
      localStorage.setItem("userNickname", profileData.nickname);
      localStorage.setItem("userRole", profileData.role);
      if (profileData.user_id) {
        localStorage.setItem("userId", profileData.user_id);
      }
      syncLocaleFromUser(profileData.preferred_language);

      // If rejected application exists, prefill fields for easy re-apply
      if (profileData.latest_application?.status === "rejected") {
        setBio(profileData.latest_application.bio || "");
        setLinkedinUrl(profileData.latest_application.linkedin_url || "");
        setPortfolioUrl(profileData.latest_application.portfolio_url || "");
      }
    } catch (err: unknown) {
      setErrorState({ kind: "api", err, fallbackKey: "profile.loadError" });
    } finally {
      setLoading(false);
    }
  }, [token, syncLocaleFromUser]);

  useEffect(() => {
    fetchProfile();
  }, [fetchProfile]);

  const handleLanguageSelect = (next: Locale) => {
    setErrorState(null);
    setLocale(next);
    setSuccessKey("profile.languageSaved");
  };

  const handleUpdateNickname = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newNickname.trim()) return;
    setSavingNickname(true);
    setErrorState(null);
    setSuccessKey(null);

    try {
      await axios.put(
        "/api/user/nickname", 
        { nickname: newNickname.trim() },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      setSuccessKey("profile.nicknameUpdated");
      setIsEditingNickname(false);
      fetchProfile();
    } catch (err: unknown) {
      setErrorState({ kind: "api", err, fallbackKey: "profile.nicknameUpdateError" });
    } finally {
      setSavingNickname(false);
    }
  };

  const handleApplySpecialist = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorState(null);
    setSuccessKey(null);

    if (!resumeFile) {
      setErrorState({ kind: "key", key: "errors.SPECIALIST_RESUME_REQUIRED" });
      return;
    }

    if (!bio.trim()) {
      setErrorState({ kind: "key", key: "errors.SPECIALIST_BIO_REQUIRED" });
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

      setSuccessKey("profile.applicationSubmitted");
      setShowApplyForm(false);
      fetchProfile();
    } catch (err: unknown) {
      setErrorState({ kind: "api", err, fallbackKey: "profile.applicationError" });
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
        {t("profile.unavailable")}
      </div>
    );
  }

  const app = profile.latest_application;
  const error = errorState
    ? errorState.kind === "key"
      ? t(errorState.key)
      : resolveApiError(errorState.err, t, t(errorState.fallbackKey))
    : "";
  const success = successKey ? t(successKey) : "";

  const roleLabel =
    profile.role === "admin"
      ? t("common.roles.admin")
      : profile.role === "specialist"
      ? t("common.roles.specialist")
      : t("common.roles.learner");

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
                    className="px-3 py-1.5 bg-white text-black text-sm font-semibold rounded hover:bg-white/90 cursor-pointer"
                  >
                    {savingNickname ? <Loader2 className="w-4 h-4 animate-spin" /> : t("common.save")}
                  </button>
                  <button
                    type="button"
                    onClick={() => setIsEditingNickname(false)}
                    className="px-3 py-1.5 text-sm text-graphite-400 hover:text-white cursor-pointer"
                  >
                    {t("common.cancel")}
                  </button>
                </form>
              ) : (
                <div className="flex items-center gap-2">
                  <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight font-mono">
                    {profile.nickname}
                  </h1>
                  <button
                    onClick={() => setIsEditingNickname(true)}
                    className="text-graphite-400 hover:text-white transition-colors cursor-pointer"
                    title={t("profile.editNicknameTitle")}
                  >
                    <Edit2 size={16} />
                  </button>
                </div>
              )}
              
              <span className={`px-2.5 py-0.5 rounded text-xs font-mono font-semibold uppercase tracking-wide border ${
                profile.role === "admin"
                  ? "bg-amber-500/10 text-amber-400 border-amber-500/30"
                  : profile.role === "specialist"
                  ? "bg-indigo-500/10 text-indigo-400 border-indigo-500/30"
                  : "bg-graphite-800 text-graphite-300 border-graphite-700"
              }`}>
                {roleLabel}
              </span>
            </div>
            
            <p className="text-sm text-graphite-300 font-mono mt-1">{profile.email}</p>
            <p className="text-sm text-graphite-400 mt-1.5">
              {t("profile.competencePrefix")} <span className="text-white font-semibold">{t("profile.tierBadge", { tier: profile.current_tier })}</span>
            </p>
          </div>
        </div>
      </div>

      {error && (
        <div className="p-4 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400 text-sm font-mono">
          {error}
        </div>
      )}

      {success && (
        <div className="p-4 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-sm font-mono flex items-center gap-2">
          <CheckCircle2 size={16} />
          <span>{success}</span>
        </div>
      )}

      {/* Language Preference Card */}
      <div className="bg-graphite-900 border border-graphite-800 rounded-xl p-6 sm:p-8 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div className="space-y-1.5">
          <h2 className="text-lg font-semibold text-white tracking-tight flex items-center gap-2">
            <Globe className="w-5 h-5 text-graphite-300" />
            {t("profile.languageCardTitle")}
          </h2>
          <p className="text-sm text-graphite-300 font-normal leading-relaxed max-w-xl">
            {t("profile.languageCardDesc")}
          </p>
        </div>

        <div
          role="group"
          aria-label={t("common.languageToggleAria")}
          className="inline-flex items-center rounded-md border border-graphite-800 bg-graphite-950 p-1 shrink-0"
        >
          <button
            type="button"
            onClick={() => handleLanguageSelect("en")}
            aria-pressed={locale === "en"}
            className={`px-3.5 py-2 rounded text-sm font-medium transition-colors cursor-pointer ${
              locale === "en"
                ? "bg-white text-black font-semibold"
                : "text-graphite-400 hover:text-white"
            }`}
          >
            {t("profile.languageEnglish")}
          </button>
          <button
            type="button"
            onClick={() => handleLanguageSelect("th")}
            aria-pressed={locale === "th"}
            className={`px-3.5 py-2 rounded text-sm font-medium transition-colors cursor-pointer ${
              locale === "th"
                ? "bg-white text-black font-semibold"
                : "text-graphite-400 hover:text-white"
            }`}
          >
            {t("profile.languageThai")}
          </button>
        </div>
      </div>

      {/* Exam History & Reviews Section */}
      <div className="bg-graphite-900 border border-graphite-800 rounded-xl p-6 sm:p-8 space-y-5">
        <div className="space-y-1.5 border-b border-graphite-800 pb-4">
          <h2 className="text-lg font-semibold text-white tracking-tight flex items-center gap-2">
            <ClipboardCheck className="w-5 h-5 text-emerald-400" />
            {t("profile.examHistory.title")}
          </h2>
          <p className="text-sm text-graphite-300 font-normal leading-relaxed">
            {t("profile.examHistory.subtitle")}
          </p>
        </div>

        {examHistory.some((item) => item.exam_type === "post") && questionnaireStatus && (
          <QuestionnairePromptBanner status={questionnaireStatus} variant="compact" />
        )}

        {examHistory.length === 0 ? (
          <div className="p-5 rounded-lg bg-graphite-950 border border-graphite-800 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div className="space-y-1">
              <h3 className="text-base font-semibold text-white">
                {t("profile.examHistory.emptyTitle")}
              </h3>
              <p className="text-sm text-graphite-300">
                {t("profile.examHistory.emptyDesc")}
              </p>
            </div>
            <Link
              to="/pre-test"
              className="px-4 py-2.5 bg-white text-black font-semibold rounded text-sm hover:bg-white/90 transition-all whitespace-nowrap inline-flex items-center gap-2"
            >
              <span>{t("profile.examHistory.startPreTest")}</span>
              <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
        ) : (
          <div className="space-y-3">
            {examHistory.map((item) => {
              const isPost = item.exam_type === "post";
              return (
                <div
                  key={item.session_id}
                  className="p-4 rounded-lg bg-graphite-950/70 border border-graphite-800 flex flex-col sm:flex-row sm:items-center justify-between gap-4"
                >
                  <div className="space-y-1.5">
                    <div className="flex flex-wrap items-center gap-2.5">
                      <span
                        className={`px-2.5 py-0.5 rounded text-xs font-mono font-semibold border ${getExamTypeBadgeClass(item.exam_type)}`}
                      >
                        {isPost
                          ? t("profile.examHistory.postTestBadge")
                          : t("profile.examHistory.preTestBadge")}
                      </span>
                      <span className="text-base font-mono font-bold text-white">
                        {item.score.toFixed(1)}%
                      </span>
                      <span className="text-sm font-mono text-graphite-300">
                        {t("profile.examHistory.scoreSummary", {
                          correct: item.correct_count,
                          total: item.total_questions,
                        })}
                      </span>
                    </div>

                    <div className="flex flex-wrap items-center gap-4 text-xs font-mono text-graphite-400">
                      {item.completed_at && <span>{formatDate(item.completed_at)}</span>}
                      <span>
                        {t("profile.examHistory.duration", {
                          duration: formatDurationSeconds(item.time_spent_seconds),
                        })}
                      </span>
                    </div>
                  </div>

                  <Link
                    to={`/exams/review/${encodeURIComponent(item.session_id)}`}
                    className="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-md border border-graphite-700 bg-graphite-900 hover:bg-graphite-800 text-sm font-medium text-white transition-colors whitespace-nowrap"
                  >
                    <span>{t("profile.examHistory.reviewAction")}</span>
                    <ArrowRight className="w-4 h-4" />
                  </Link>
                </div>
              );
            })}
          </div>
        )}
      </div>

      {/* Role-Specific Workflows & Specialist Application Section */}
      <div className="bg-graphite-900 border border-graphite-800 rounded-xl p-6 sm:p-8 space-y-6">
        <h2 className="text-lg font-semibold text-white tracking-tight flex items-center gap-2">
          <Award className="w-5 h-5 text-indigo-400" />
          {t("profile.accreditationTitle")}
        </h2>

        {/* Status A: User is already Specialist */}
        {profile.role === "specialist" && (
          <div className="p-5 rounded-lg bg-indigo-500/10 border border-indigo-500/20 space-y-2">
            <div className="flex items-center gap-2 text-indigo-400 font-medium text-base">
              <CheckCircle2 size={18} />
              <span>{t("profile.verifiedTitle")}</span>
            </div>
            <p className="text-sm sm:text-base text-graphite-300 leading-relaxed font-normal">
              {t("profile.verifiedDesc")}
            </p>
          </div>
        )}

        {/* Status B: User is Admin */}
        {profile.role === "admin" && (
          <div className="p-5 rounded-lg bg-amber-500/10 border border-amber-500/20 space-y-2">
            <div className="flex items-center gap-2 text-amber-400 font-medium text-base">
              <Shield size={18} />
              <span>{t("profile.adminTitle")}</span>
            </div>
            <p className="text-sm sm:text-base text-graphite-300 leading-relaxed font-normal">
              {t("profile.adminDesc")}
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
                  <div className="flex items-center gap-2 text-amber-400 font-medium text-base">
                    <Clock size={18} />
                    <span>{t("profile.pendingTitle")}</span>
                  </div>
                  <span className="px-2.5 py-0.5 rounded text-xs font-mono font-bold bg-amber-500/20 text-amber-300">
                    {t("profile.pendingBadge")}
                  </span>
                </div>
                <p className="text-sm sm:text-base text-graphite-300 leading-relaxed font-normal">
                  {t("profile.pendingBodyBeforeDate")}
                  {formatDate(app.created_at)}
                  {t("profile.pendingBodyAfterDate")}
                </p>
                <div className="bg-graphite-950 p-3.5 rounded border border-graphite-800 text-sm text-graphite-300 font-mono">
                  <span className="text-graphite-400">{t("profile.submittedBioLabel")}</span> {app.bio}
                </div>
              </div>
            )}

            {/* Case 2: Application is REJECTED */}
            {app && app.status === "rejected" && !showApplyForm && (
              <div className="p-5 rounded-lg bg-rose-500/10 border border-rose-500/20 space-y-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2 text-rose-400 font-medium text-base">
                    <AlertTriangle size={18} />
                    <span>{t("profile.rejectedTitle")}</span>
                  </div>
                  <span className="px-2.5 py-0.5 rounded text-xs font-mono font-bold bg-rose-500/20 text-rose-300">
                    {t("profile.rejectedBadge")}
                  </span>
                </div>
                <div className="p-3.5 bg-graphite-950 rounded border border-rose-500/30 text-sm text-graphite-200">
                  <span className="text-rose-400 font-semibold block mb-1">{t("profile.adminFeedbackLabel")}</span>
                  <p className="font-mono text-sm">{app.rejection_reason || t("profile.defaultRejectionReason")}</p>
                </div>
                <p className="text-sm text-graphite-300 font-normal">
                  {t("profile.rejectedHint")}
                </p>
                <button
                  type="button"
                  onClick={() => setShowApplyForm(true)}
                  className="px-4 py-2.5 bg-white text-black font-semibold rounded text-sm hover:bg-white/90 transition-all flex items-center gap-2 cursor-pointer"
                >
                  <span>{t("profile.editReapplyButton")}</span>
                  <ArrowRight size={16} />
                </button>
              </div>
            )}

            {/* Case 3: No Application yet */}
            {(!app || (app.status === "rejected" && showApplyForm)) && (
              <div className="space-y-4">
                {!showApplyForm && !app && (
                  <div className="p-5 rounded-lg bg-graphite-950 border border-graphite-800 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
                    <div>
                      <h3 className="text-base font-semibold text-white">{t("profile.applyTitle")}</h3>
                      <p className="text-sm text-graphite-300 mt-1 font-normal">
                        {t("profile.applyDesc")}
                      </p>
                    </div>
                    <button
                      onClick={() => setShowApplyForm(true)}
                      className="px-4 py-2.5 bg-white text-black font-semibold rounded text-sm hover:bg-white/90 transition-all whitespace-nowrap cursor-pointer"
                    >
                      {t("profile.startApplicationButton")}
                    </button>
                  </div>
                )}

                {/* Application Form */}
                {showApplyForm && (
                  <form onSubmit={handleApplySpecialist} className="p-6 bg-graphite-950 border border-graphite-800 rounded-lg space-y-4 animate-fade-in">
                    <div className="flex justify-between items-center border-b border-graphite-800 pb-3">
                      <h3 className="text-base font-semibold text-white">
                        {app?.status === "rejected" ? t("profile.reviseFormTitle") : t("profile.newFormTitle")}
                      </h3>
                      <button
                        type="button"
                        onClick={() => setShowApplyForm(false)}
                        className="text-sm text-graphite-400 hover:text-white cursor-pointer"
                      >
                        {t("common.cancel")}
                      </button>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-graphite-200 mb-1.5">
                        {t("profile.bioLabel")} <span className="text-red-400">*</span>
                      </label>
                      <textarea
                        value={bio}
                        onChange={(e) => setBio(e.target.value)}
                        required
                        rows={3}
                        className="w-full px-3.5 py-2.5 bg-graphite-900 border border-graphite-800 rounded text-sm text-white placeholder-graphite-500 focus:outline-none focus:border-white transition-all"
                        placeholder={t("profile.bioPlaceholder")}
                      />
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-graphite-200 mb-1.5">
                          {t("profile.resumeLabel")} <span className="text-red-400">*</span>
                        </label>
                        <label className="flex flex-col items-center justify-center p-4 border border-dashed border-graphite-700 hover:border-white rounded bg-graphite-900 cursor-pointer transition-colors text-center">
                          <Upload size={18} className="text-graphite-400 mb-1" />
                          <span className="text-sm text-graphite-300 truncate max-w-full">
                            {resumeFile ? resumeFile.name : t("profile.attachResume")}
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
                        <label className="block text-sm font-medium text-graphite-200 mb-1.5">
                          {t("profile.certLabel")}
                        </label>
                        <label className="flex flex-col items-center justify-center p-4 border border-dashed border-graphite-700 hover:border-white rounded bg-graphite-900 cursor-pointer transition-colors text-center">
                          <Upload size={18} className="text-graphite-400 mb-1" />
                          <span className="text-sm text-graphite-300 truncate max-w-full">
                            {certFile ? certFile.name : t("profile.attachCert")}
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
                        <label className="block text-sm font-medium text-graphite-200 mb-1.5">
                          {t("profile.linkedinLabel")}
                        </label>
                        <input
                          type="url"
                          value={linkedinUrl}
                          onChange={(e) => setLinkedinUrl(e.target.value)}
                          className="w-full px-3.5 py-2.5 bg-graphite-900 border border-graphite-800 rounded text-sm text-white placeholder-graphite-500 focus:outline-none focus:border-white transition-all"
                          placeholder={t("profile.linkedinPlaceholder")}
                        />
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-graphite-200 mb-1.5">
                          {t("profile.portfolioLabel")}
                        </label>
                        <input
                          type="url"
                          value={portfolioUrl}
                          onChange={(e) => setPortfolioUrl(e.target.value)}
                          className="w-full px-3.5 py-2.5 bg-graphite-900 border border-graphite-800 rounded text-sm text-white placeholder-graphite-500 focus:outline-none focus:border-white transition-all"
                          placeholder={t("profile.portfolioPlaceholder")}
                        />
                      </div>
                    </div>

                    <div className="flex justify-end gap-3 pt-2">
                      <button
                        type="button"
                        onClick={() => setShowApplyForm(false)}
                        className="px-4 py-2 border border-graphite-800 text-graphite-300 rounded text-sm hover:bg-graphite-900 cursor-pointer"
                      >
                        {t("common.cancel")}
                      </button>
                      <button
                        type="submit"
                        disabled={submittingApp}
                        className="px-5 py-2 bg-white text-black font-semibold rounded text-sm hover:bg-white/90 flex items-center gap-2 disabled:opacity-50 cursor-pointer"
                      >
                        {submittingApp ? <Loader2 className="w-4 h-4 animate-spin" /> : t("profile.submitApplication")}
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
