import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { Shield, CheckCircle2, AlertTriangle, Loader2, ArrowRight } from "lucide-react";
import axios from "axios";
import { useI18n } from "../i18n/index.tsx";
import { resolveApiError } from "../i18n/format.ts";
import ResendVerificationPanel from "../components/ResendVerificationPanel.tsx";

interface VerificationAttemptOutcome {
  token: string;
  status: "success" | "error";
  verifiedEmail: string;
  errorObj: unknown;
}

// Share an in-flight/settled request per token so React 19 StrictMode double-mounts
// never fire duplicate verify POSTs or discard the first response cleanly.
const tokenVerificationCache = new Map<string, Promise<VerificationAttemptOutcome>>();

function runVerifyTokenOnce(token: string): Promise<VerificationAttemptOutcome> {
  const cached = tokenVerificationCache.get(token);
  if (cached) {
    return cached;
  }
  const promise = axios
    .post("/api/verify-email", { token })
    .then(
      (res): VerificationAttemptOutcome => ({
        token,
        status: "success",
        verifiedEmail: res.data?.email ?? "",
        errorObj: null,
      }),
    )
    .catch(
      (err: unknown): VerificationAttemptOutcome => ({
        token,
        status: "error",
        verifiedEmail: "",
        errorObj: err,
      }),
    );
  tokenVerificationCache.set(token, promise);
  return promise;
}

export default function VerifyEmailPage() {
  const [searchParams] = useSearchParams();
  const { t } = useI18n();
  const token = (searchParams.get("token") ?? "").trim();

  const [outcome, setOutcome] = useState<VerificationAttemptOutcome | null>(null);

  useEffect(() => {
    if (!token) return;
    let cancelled = false;

    runVerifyTokenOnce(token).then((result) => {
      if (!cancelled) {
        setOutcome(result);
      }
    });

    return () => {
      cancelled = true;
    };
  }, [token]);

  const status: "loading" | "success" | "error" = !token
    ? "error"
    : outcome?.token === token
      ? outcome.status
      : "loading";

  const errorMessage = !token
    ? t("auth.verification.missingToken")
    : outcome?.errorObj
      ? resolveApiError(outcome.errorObj, t, t("errors.AUTH_VERIFICATION_TOKEN_INVALID"))
      : t("errors.AUTH_VERIFICATION_TOKEN_INVALID");

  return (
    <div className="min-h-[80vh] flex items-center justify-center py-12">
      <div className="w-full max-w-md bg-graphite-900 border border-graphite-800 rounded-lg p-8 shadow-xl">
        <div className="flex flex-col items-center text-center mb-6">
          <div className="w-12 h-12 bg-graphite-950 text-white rounded border border-graphite-800 flex items-center justify-center mb-4">
            <Shield size={20} />
          </div>

          {status === "loading" && (
            <>
              <Loader2 size={28} className="animate-spin text-white mb-3" />
              <h2 className="text-xl font-bold text-white tracking-tight">
                {t("auth.verification.pageTitleVerifying")}
              </h2>
              <p className="text-graphite-300 text-sm mt-2 leading-relaxed">
                {t("auth.verification.pageSubtitleVerifying")}
              </p>
            </>
          )}

          {status === "success" && (
            <>
              <div className="w-12 h-12 rounded-full bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400 mb-3">
                <CheckCircle2 size={24} />
              </div>
              <h2 className="text-xl font-bold text-white tracking-tight">
                {t("auth.verification.verifiedTitle")}
              </h2>
              <p className="text-graphite-200 text-sm mt-2 leading-relaxed">
                {t("auth.verification.verifiedBody", {
                  email: outcome?.verifiedEmail || "SOC Trainer",
                })}
              </p>
              <Link
                to="/login"
                className="w-full mt-6 bg-white text-black font-semibold py-2.5 px-4 rounded text-sm hover:bg-white/90 transition-all flex items-center justify-center gap-2"
              >
                <span>{t("auth.verification.goToLogin")}</span>
                <ArrowRight size={16} />
              </Link>
            </>
          )}

          {status === "error" && (
            <>
              <div className="w-12 h-12 rounded-full bg-amber-500/10 border border-amber-500/30 flex items-center justify-center text-amber-400 mb-3">
                <AlertTriangle size={24} />
              </div>
              <h2 className="text-xl font-bold text-white tracking-tight">
                {t("auth.verification.failedTitle")}
              </h2>
              <p className="text-red-400 text-sm mt-2 leading-relaxed">{errorMessage}</p>
            </>
          )}
        </div>

        {status === "error" && (
          <div className="pt-4 border-t border-graphite-800 space-y-4">
            <p className="text-xs sm:text-sm text-graphite-300 leading-relaxed">
              {t("auth.verification.recoveryPrompt")}
            </p>
            <ResendVerificationPanel allowEmailEdit />
            <div className="text-center pt-1">
              <Link to="/login" className="text-sm text-graphite-300 hover:text-white underline">
                {t("auth.verification.goToLogin")}
              </Link>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
