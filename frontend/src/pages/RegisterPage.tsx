import { useEffect, useRef, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Shield, ArrowRight, Loader2, Award, BookOpen, Upload, CheckCircle2, MailCheck } from "lucide-react";
import axios from "axios";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError } from "../i18n/format.ts";
import ResendVerificationPanel from "../components/ResendVerificationPanel.tsx";

type RegisterErrorState =
  | { kind: "key"; key: string }
  | { kind: "api"; err: unknown }
  | null;

export default function RegisterPage() {
  const navigate = useNavigate();
  const { locale, syncLocaleFromUser, t } = useI18n();
  const [accountType, setAccountType] = useState<"learner" | "specialist">("learner");
  const [email, setEmail] = useState("");
  const [nickname, setNickname] = useState("");
  const [password, setPassword] = useState("");
  
  // Specialist specific fields
  const [bio, setBio] = useState("");
  const [resumeFile, setResumeFile] = useState<File | null>(null);
  const [certFile, setCertFile] = useState<File | null>(null);
  const [linkedinUrl, setLinkedinUrl] = useState("");
  const [portfolioUrl, setPortfolioUrl] = useState("");

  const [errorState, setErrorState] = useState<RegisterErrorState>(null);
  const [loading, setLoading] = useState(false);
  const [successKey, setSuccessKey] = useState<string | null>(null);
  const [pendingVerifyInfo, setPendingVerifyInfo] = useState<{
    email: string;
    isSpecialist: boolean;
  } | null>(null);
  const navTimerRef = useRef<number | null>(null);

  useEffect(() => {
    return () => {
      if (navTimerRef.current !== null) {
        window.clearTimeout(navTimerRef.current);
      }
    };
  }, []);

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorState(null);
    setSuccessKey(null);
    setLoading(true);

    if (accountType === "specialist" && !resumeFile) {
      setErrorState({ kind: "key", key: "errors.SPECIALIST_RESUME_REQUIRED" });
      setLoading(false);
      return;
    }

    if (accountType === "specialist" && !bio.trim()) {
      setErrorState({ kind: "key", key: "errors.SPECIALIST_BIO_REQUIRED" });
      setLoading(false);
      return;
    }

    try {
      if (accountType === "specialist") {
        const formData = new FormData();
        formData.append("email", email.trim());
        formData.append("password", password);
        formData.append("nickname", nickname.trim());
        formData.append("preferred_language", locale);
        formData.append("bio", bio);
        if (resumeFile) formData.append("resume", resumeFile);
        if (certFile) formData.append("certificate", certFile);
        if (linkedinUrl) formData.append("linkedin_url", linkedinUrl);
        if (portfolioUrl) formData.append("portfolio_url", portfolioUrl);

        const specRes = await axios.post("/api/register-specialist", formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });

        if (specRes.data?.requires_verification) {
          setPendingVerifyInfo({
            email: specRes.data?.email || email.trim(),
            isSpecialist: true,
          });
          return;
        }

        const loginRes = await axios.post("/api/login", { email: email.trim(), password });
        const token = loginRes.data.token;
        localStorage.setItem("token", token);
        localStorage.setItem("userEmail", loginRes.data.email);
        localStorage.setItem("userNickname", loginRes.data.nickname);
        localStorage.setItem("userRole", loginRes.data.role);
        if (loginRes.data.user_id) {
          localStorage.setItem("userId", loginRes.data.user_id);
        }
        syncLocaleFromUser(loginRes.data.preferred_language);

        setSuccessKey("auth.register.specialistSuccess");
        navTimerRef.current = window.setTimeout(() => navigate("/profile"), 1500);
        return;
      }

      // Learner registration with current UI locale
      const regRes = await axios.post("/api/register", {
        email: email.trim(),
        password,
        nickname: nickname.trim(),
        preferred_language: locale,
      });

      if (regRes.data?.requires_verification) {
        setPendingVerifyInfo({
          email: regRes.data?.email || email.trim(),
          isSpecialist: false,
        });
      } else {
        setSuccessKey("auth.register.learnerSuccess");
        navTimerRef.current = window.setTimeout(() => navigate("/login"), 1000);
      }
    } catch (err: unknown) {
      setErrorState({ kind: "api", err });
    } finally {
      setLoading(false);
    }
  };

  const errorMessage = errorState
    ? errorState.kind === "key"
      ? t(errorState.key)
      : resolveApiError(errorState.err, t, t("auth.register.fallbackError"))
    : "";
  const successMessage = successKey ? t(successKey) : "";

  if (pendingVerifyInfo) {
    return (
      <div className="min-h-[80vh] flex items-center justify-center py-12">
        <div className="w-full max-w-md bg-graphite-900 border border-graphite-800 rounded-lg p-8 shadow-xl space-y-5">
          <div className="flex flex-col items-center text-center">
            <div className="w-12 h-12 rounded-full bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400 mb-4">
              <MailCheck size={24} />
            </div>
            <h2 className="text-xl font-bold text-white tracking-tight">
              {t("auth.verification.pendingTitle")}
            </h2>
            <p className="text-graphite-200 text-sm mt-2 leading-relaxed">
              {pendingVerifyInfo.isSpecialist
                ? t("auth.verification.pendingSpecialistBody", { email: pendingVerifyInfo.email })
                : t("auth.verification.pendingLearnerBody", { email: pendingVerifyInfo.email })}
            </p>
            <p className="text-xs sm:text-sm text-graphite-400 mt-2 leading-relaxed">
              {t("auth.verification.expiryHint")}
            </p>
          </div>

          <ResendVerificationPanel initialEmail={pendingVerifyInfo.email} />

          <Link
            to="/login"
            className="w-full bg-white text-black font-semibold py-2.5 px-4 rounded text-sm hover:bg-white/90 transition-all flex items-center justify-center gap-2"
          >
            <span>{t("auth.verification.goToLogin")}</span>
            <ArrowRight size={16} />
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-[80vh] flex items-center justify-center py-12">
      <div className="w-full max-w-lg bg-graphite-900 border border-graphite-800 rounded-lg p-8 shadow-xl">
        <div className="flex flex-col items-center mb-6">
          <div className="w-12 h-12 bg-graphite-950 text-white rounded border border-graphite-800 flex items-center justify-center mb-4">
            <Shield size={20} />
          </div>
          <h2 className="text-xl font-bold text-white tracking-tight">{t("auth.register.title")}</h2>
          <p className="text-graphite-300 text-sm mt-1.5 text-center font-normal">
            {t("auth.register.subtitle")}
          </p>
        </div>

        {/* Role Selector Tabs */}
        <div className="grid grid-cols-2 gap-2 p-1 bg-graphite-950 border border-graphite-800 rounded-md mb-6">
          <button
            type="button"
            onClick={() => { setAccountType("learner"); setErrorState(null); }}
            className={`flex items-center justify-center gap-2 py-2.5 px-3 rounded text-sm font-medium transition-all cursor-pointer ${
              accountType === "learner"
                ? "bg-white text-black font-semibold shadow-sm"
                : "text-graphite-400 hover:text-white"
            }`}
          >
            <BookOpen size={16} />
            <span>{t("auth.register.learnerTab")}</span>
          </button>
          <button
            type="button"
            onClick={() => { setAccountType("specialist"); setErrorState(null); }}
            className={`flex items-center justify-center gap-2 py-2.5 px-3 rounded text-sm font-medium transition-all cursor-pointer ${
              accountType === "specialist"
                ? "bg-white text-black font-semibold shadow-sm"
                : "text-graphite-400 hover:text-white"
            }`}
          >
            <Award size={16} />
            <span>{t("auth.register.specialistTab")}</span>
          </button>
        </div>

        {errorMessage && (
          <div className="mb-4 p-3 rounded bg-red-500/10 border border-red-500/20 text-red-400 text-sm font-mono">
            {errorMessage}
          </div>
        )}

        {successMessage && (
          <div className="mb-4 p-3 rounded bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-sm font-mono flex items-center gap-2">
            <CheckCircle2 size={16} />
            <span>{successMessage}</span>
          </div>
        )}

        <form onSubmit={handleRegister} className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-graphite-200 mb-1.5">
              {t("auth.register.nicknameLabel")} <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              required
              minLength={2}
              maxLength={30}
              className="w-full px-3.5 py-2.5 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-500 focus:outline-none focus:border-white transition-all"
              placeholder={t("auth.register.nicknamePlaceholder")}
            />
            <p className="text-xs text-graphite-400 mt-1.5">{t("auth.register.nicknameHelp")}</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-graphite-200 mb-1.5">
              {t("auth.register.emailLabel")} <span className="text-red-400">*</span>
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="w-full px-3.5 py-2.5 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-500 focus:outline-none focus:border-white transition-all"
              placeholder={t("auth.register.emailPlaceholder")}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-graphite-200 mb-1.5">
              {t("auth.register.passwordLabel")} <span className="text-red-400">*</span>
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={6}
              className="w-full px-3.5 py-2.5 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-500 focus:outline-none focus:border-white transition-all"
              placeholder={t("auth.register.passwordPlaceholder")}
            />
          </div>

          {/* Specialist Fields */}
          {accountType === "specialist" && (
            <div className="pt-3 border-t border-graphite-800 space-y-4 animate-fade-in">
              <div className="bg-graphite-950/60 p-3.5 rounded border border-graphite-800">
                <p className="text-sm text-graphite-200 font-medium mb-1">{t("auth.register.specialistCardTitle")}</p>
                <p className="text-xs sm:text-sm text-graphite-300 leading-relaxed">
                  {t("auth.register.specialistCardDesc")}
                </p>
              </div>

              <div>
                <label className="block text-sm font-medium text-graphite-200 mb-1.5">
                  {t("auth.register.bioLabel")} <span className="text-red-400">*</span>
                </label>
                <textarea
                  value={bio}
                  onChange={(e) => setBio(e.target.value)}
                  required={accountType === "specialist"}
                  rows={3}
                  className="w-full px-3.5 py-2.5 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-500 focus:outline-none focus:border-white transition-all"
                  placeholder={t("auth.register.bioPlaceholder")}
                />
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-graphite-200 mb-1.5">
                    {t("auth.register.resumeLabel")} <span className="text-red-400">*</span>
                  </label>
                  <label className="flex flex-col items-center justify-center p-3.5 border border-dashed border-graphite-700 hover:border-white rounded bg-graphite-950 cursor-pointer transition-colors text-center">
                    <Upload size={16} className="text-graphite-400 mb-1" />
                    <span className="text-xs text-graphite-300 truncate max-w-full">
                      {resumeFile ? resumeFile.name : t("auth.register.chooseResume")}
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
                    {t("auth.register.certLabel")}
                  </label>
                  <label className="flex flex-col items-center justify-center p-3.5 border border-dashed border-graphite-700 hover:border-white rounded bg-graphite-950 cursor-pointer transition-colors text-center">
                    <Upload size={16} className="text-graphite-400 mb-1" />
                    <span className="text-xs text-graphite-300 truncate max-w-full">
                      {certFile ? certFile.name : t("auth.register.chooseCert")}
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

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium text-graphite-200 mb-1.5">
                    {t("auth.register.linkedinLabel")}
                  </label>
                  <input
                    type="url"
                    value={linkedinUrl}
                    onChange={(e) => setLinkedinUrl(e.target.value)}
                    className="w-full px-3.5 py-2.5 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-500 focus:outline-none focus:border-white transition-all"
                    placeholder={t("auth.register.linkedinPlaceholder")}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-graphite-200 mb-1.5">
                    {t("auth.register.portfolioLabel")}
                  </label>
                  <input
                    type="url"
                    value={portfolioUrl}
                    onChange={(e) => setPortfolioUrl(e.target.value)}
                    className="w-full px-3.5 py-2.5 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-500 focus:outline-none focus:border-white transition-all"
                    placeholder={t("auth.register.portfolioPlaceholder")}
                  />
                </div>
              </div>
            </div>
          )}

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-white text-black font-semibold py-2.5 rounded text-sm hover:bg-white/90 transition-all flex justify-center items-center gap-2 mt-4 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
          >
            {loading ? <Loader2 className="animate-spin w-4 h-4" /> : accountType === "specialist" ? t("auth.register.submitSpecialist") : t("auth.register.submitLearner")}
            {!loading && <ArrowRight size={16} />}
          </button>
        </form>

        <div className="mt-6 text-center text-sm text-graphite-400">
          {t("auth.register.alreadyHaveAccount")}{" "}
          <Link to="/login" className="text-white hover:underline">
            {t("auth.register.signInLink")}
          </Link>
        </div>
      </div>
    </div>
  );
}
