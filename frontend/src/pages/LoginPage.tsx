import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Shield, ArrowRight, Loader2 } from "lucide-react";
import axios from "axios";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError } from "../i18n/format.ts";

export default function LoginPage() {
  const navigate = useNavigate();
  const { syncLocaleFromUser, t } = useI18n();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [errorObj, setErrorObj] = useState<unknown>(null);
  const [loading, setLoading] = useState(false);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorObj(null);
    setLoading(true);

    try {
      const response = await axios.post("/api/login", {
        email,
        password,
      });
      // Save user details to localStorage
      localStorage.setItem("token", response.data.token);
      localStorage.setItem("userEmail", response.data.email);
      localStorage.setItem("userNickname", response.data.nickname || response.data.email.split('@')[0]);
      localStorage.setItem("userRole", response.data.role || "learner");
      if (response.data.user_id) {
        localStorage.setItem("userId", response.data.user_id);
      }
      syncLocaleFromUser(response.data.preferred_language);
      
      if (response.data.role === "admin") {
        navigate("/admin/specialists");
      } else {
        navigate("/exam");
      }
    } catch (err: unknown) {
      setErrorObj(err);
    } finally {
      setLoading(false);
    }
  };

  const errorMessage = errorObj ? resolveApiError(errorObj, t, t("auth.login.fallbackError")) : "";

  return (
    <div className="min-h-[80vh] flex items-center justify-center">
      <div className="w-full max-w-md bg-graphite-900 border border-graphite-800 rounded-lg p-8">
        <div className="flex flex-col items-center mb-8">
          <div className="w-12 h-12 bg-graphite-950 text-white rounded border border-graphite-800 flex items-center justify-center mb-4">
            <Shield size={20} />
          </div>
          <h2 className="text-xl font-bold text-white tracking-tight">{t("auth.login.title")}</h2>
          <p className="text-graphite-400 text-xs mt-2 text-center font-light leading-relaxed">
            {t("auth.login.subtitle")}
          </p>
        </div>

        {errorMessage && (
          <div className="mb-4 p-3 rounded bg-red-500/10 border border-red-500/20 text-red-400 text-xs font-mono">
            {errorMessage}
          </div>
        )}

        <form onSubmit={handleLogin} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-graphite-300 mb-1">
              {t("auth.login.emailLabel")}
            </label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className="w-full px-3 py-2 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-600 focus:outline-none focus:border-white transition-all font-light"
              placeholder={t("auth.login.emailPlaceholder")}
            />
          </div>

          <div>
            <label className="block text-xs font-medium text-graphite-300 mb-1 flex justify-between">
              <span>{t("auth.login.passwordLabel")}</span>
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className="w-full px-3 py-2 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-600 focus:outline-none focus:border-white transition-all font-light"
              placeholder={t("auth.login.passwordPlaceholder")}
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-white text-black font-semibold py-2 rounded text-xs hover:bg-white/90 transition-all flex justify-center items-center gap-2 mt-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {loading ? <Loader2 className="animate-spin w-4 h-4" /> : t("auth.login.submitButton")}
            {!loading && <ArrowRight size={14} />}
          </button>
        </form>

        <div className="mt-6 text-center text-xs text-graphite-400">
          {t("auth.login.noAccount")}{" "}
          <Link to="/register" className="text-white hover:underline">
            {t("auth.login.registerLink")}
          </Link>
        </div>
      </div>
    </div>
  );
}
