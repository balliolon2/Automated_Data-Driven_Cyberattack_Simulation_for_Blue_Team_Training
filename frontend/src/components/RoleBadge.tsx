import { cn } from "../lib/utils";

interface RoleBadgeProps {
  role?: string;
  className?: string;
}

export default function RoleBadge({ role, className }: RoleBadgeProps) {
  if (!role) return null;

  if (role === "admin") {
    return (
      <span
        className={cn(
          "px-1.5 py-0.5 rounded text-[9px] font-mono font-bold bg-amber-500/10 text-amber-400 border border-amber-500/20 uppercase tracking-wide",
          className
        )}
      >
        ADMIN
      </span>
    );
  }

  if (role === "specialist") {
    return (
      <span
        className={cn(
          "px-1.5 py-0.5 rounded text-[9px] font-mono font-bold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 uppercase tracking-wide",
          className
        )}
      >
        SPECIALIST
      </span>
    );
  }

  if (role === "learner") {
    return (
      <span
        className={cn(
          "px-1.5 py-0.5 rounded text-[9px] font-mono font-bold bg-graphite-900 text-graphite-400 border border-graphite-800 uppercase tracking-wide",
          className
        )}
      >
        LEARNER
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
