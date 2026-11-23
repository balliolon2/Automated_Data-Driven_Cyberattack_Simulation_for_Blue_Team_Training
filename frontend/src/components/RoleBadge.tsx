import { cn } from "../lib/utils";
import { useI18n } from "../i18n";

interface RoleBadgeProps {
  role?: string;
  className?: string;
}

const BASE_BADGE_CLASS =
  "px-2 py-0.5 rounded text-xs font-mono font-bold border uppercase tracking-wide whitespace-nowrap shrink-0";

const ROLE_STYLES: Record<string, { tone: string; titleKey: string; badgeKey: string }> = {
  admin: {
    tone: "bg-amber-500/10 text-amber-400 border-amber-500/20",
    titleKey: "common.roles.admin",
    badgeKey: "common.roles.adminBadge",
  },
  specialist: {
    tone: "bg-indigo-500/10 text-indigo-400 border-indigo-500/20",
    titleKey: "common.roles.specialist",
    badgeKey: "common.roles.specialistBadge",
  },
  learner: {
    tone: "bg-graphite-900 text-graphite-300 border-graphite-800",
    titleKey: "common.roles.learner",
    badgeKey: "common.roles.learnerBadge",
  },
};

export default function RoleBadge({ role, className }: RoleBadgeProps) {
  const { t } = useI18n();
  if (!role) return null;

  const spec = ROLE_STYLES[role];
  if (spec) {
    return (
      <span
        title={t(spec.titleKey)}
        className={cn(BASE_BADGE_CLASS, spec.tone, className)}
      >
        {t(spec.badgeKey)}
      </span>
    );
  }

  return (
    <span
      className={cn(BASE_BADGE_CLASS, "bg-graphite-800 text-graphite-300 border-graphite-700", className)}
    >
      {role}
    </span>
  );
}
