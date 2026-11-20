import { cn } from "../lib/utils";
import { useI18n } from "../i18n";

interface RoleBadgeProps {
  role?: string;
  className?: string;
}

export default function RoleBadge({ role, className }: RoleBadgeProps) {
  const { t } = useI18n();
  if (!role) return null;

  if (role === "admin") {
    return (
      <span
        title={t("common.roles.admin")}
        className={cn(
          "px-1.5 py-0.5 rounded text-[9px] font-mono font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20 uppercase tracking-wide",
          className
        )}
      >
        {t("common.roles.adminBadge")}
      </span>
    );
  }

  if (role === "specialist") {
    return (
      <span
        title={t("common.roles.specialist")}
        className={cn(
          "px-1.5 py-0.5 rounded text-[9px] font-mono font-bold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 uppercase tracking-wide",
          className
        )}
      >
        {t("common.roles.specialistBadge")}
      </span>
    );
  }

  if (role === "learner") {
    return (
      <span
        title={t("common.roles.learner")}
        className={cn(
          "px-1.5 py-0.5 rounded text-[9px] font-mono font-bold bg-graphite-900 text-graphite-400 border border-graphite-800 uppercase tracking-wide",
          className
        )}
      >
        {t("common.roles.learnerBadge")}
      </span>
    );
  }

  return (
    <span
      className={cn(
        "px-1.5 py-0.5 rounded text-[9px] font-mono font-bold bg-graphite-800 text-graphite-300 border border-graphite-700 uppercase tracking-wide",
        className
      )}
    >
      {role}
    </span>
  );
}
