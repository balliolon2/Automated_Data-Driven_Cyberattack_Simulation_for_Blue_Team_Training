import { useEffect, useState } from "react";
import { Mail, Loader2, CheckCircle2 } from "lucide-react";
import axios from "axios";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError } from "../i18n/format.ts";

interface ResendVerificationPanelProps {
  initialEmail?: string;
  allowEmailEdit?: boolean;
}

export default function ResendVerificationPanel({
  initialEmail = "",
  allowEmailEdit = false,
}: ResendVerificationPanelProps) {
  const { t } = useI18n();
  const [targetEmail, setTargetEmail] = useState(initialEmail);
  const [loading, setLoading] = useState(false);
  const [cooldown, setCooldown] = useState(0);
  const [sentMessage, setSentMessage] = useState("");
  const [errorMessage, setErrorMessage] = useState("");

  const hasCooldown = cooldown > 0;
  useEffect(() => {
    if (!hasCooldown) return;
    const timer = window.setInterval(() => {
      setCooldown((prev) => (prev <= 1 ? 0 : prev - 1));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [hasCooldown]);

  const handleResend = async () => {
    const trimmed = targetEmail.trim();
    if (!trimmed || loading || cooldown > 0) return;

    setLoading(true);
    setSentMessage("");
    setErrorMessage("");
    try {
      await axios.post("/api/resend-verification", { email: trimmed });
      setSentMessage(t("auth.verification.resendSuccess"));
      setCooldown(60);
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        const retrySecs = Number(err.response?.data?.retry_after_seconds);
        if (Number.isFinite(retrySecs) && retrySecs > 0) {
          setCooldown(Math.ceil(retrySecs));
        }
      }
      setErrorMessage(resolveApiError(err, t, t("auth.verification.resendFallbackError")));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-3">
      {allowEmailEdit && (
        <div>
          <label className="block text-sm font-medium text-graphite-200 mb-1.5">
            {t("auth.verification.emailInputLabel")}
          </label>
          <input
            type="email"
            value={targetEmail}
            onChange={(e) => setTargetEmail(e.target.value)}
            placeholder={t("auth.login.emailPlaceholder")}
            className="w-full px-3.5 py-2.5 bg-graphite-950 border border-graphite-800 rounded text-sm text-white placeholder-graphite-500 focus:outline-none focus:border-white transition-all"
          />
        </div>
      )}

      {sentMessage && (
        <div className="p-2.5 rounded bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 text-xs sm:text-sm flex items-center gap-2">
          <CheckCircle2 size={15} className="shrink-0" />
          <span>{sentMessage}</span>
        </div>
      )}

      {errorMessage && (
        <div className="p-2.5 rounded bg-red-500/10 border border-red-500/20 text-red-400 text-xs sm:text-sm">
          {errorMessage}
        </div>
      )}

      <button
        type="button"
        onClick={handleResend}
        disabled={loading || cooldown > 0 || !targetEmail.trim()}
        className="w-full px-4 py-2 rounded border border-graphite-700 bg-graphite-950 hover:border-white text-sm font-medium text-white transition-all flex items-center justify-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
      >
        {loading ? <Loader2 size={15} className="animate-spin" /> : <Mail size={15} />}
        <span>
          {loading
            ? t("auth.verification.resending")
            : cooldown > 0
              ? t("auth.verification.resendCooldown", { seconds: cooldown })
              : t("auth.verification.resendButton")}
        </span>
      </button>
    </div>
  );
}
