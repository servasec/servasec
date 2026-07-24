export const severityColors: Record<string, string> = {
  critical: "bg-red-500/15 text-red-600 dark:text-red-400",
  high: "bg-orange-500/15 text-orange-600 dark:text-orange-400",
  medium: "bg-amber-500/15 text-amber-600 dark:text-amber-400",
  low: "bg-slate-500/15 text-slate-600 dark:text-slate-400",
  info: "bg-blue-500/15 text-blue-600 dark:text-blue-400",
};

export const severityBadgeColors: Record<string, string> = {
  Critical: "bg-red-500/15 text-red-600 dark:text-red-400 border-red-500/30",
  High: "bg-orange-500/15 text-orange-600 dark:text-orange-400 border-orange-500/30",
  Medium: "bg-amber-500/15 text-amber-600 dark:text-amber-400 border-amber-500/30",
  Low: "bg-slate-500/15 text-slate-600 dark:text-slate-400 border-slate-500/30",
  Info: "bg-blue-500/15 text-blue-600 dark:text-blue-400 border-blue-500/30",
};

export const severityBarColors: Record<string, string> = {
  critical: "#ef4444",
  high: "#f97316",
  medium: "#f59e0b",
  low: "#64748b",
  info: "#3b82f6",
};

export const statusColors: Record<string, string> = {
  open: "text-red-500",
  confirmed: "text-amber-500",
  false_positive: "text-emerald-500",
  fixed: "text-blue-500",
};

export const statusLabels: Record<string, string> = {
  open: "Open",
  confirmed: "Confirmed",
  false_positive: "False Positive",
  fixed: "Fixed",
};

export const severityOptions = ["Critical", "High", "Medium", "Low", "Info"];

export const statusScanColors: Record<string, string> = {
  pending: "text-amber-500",
  running: "text-blue-500",
  completed: "text-emerald-500",
  failed: "text-red-500",
};

export const riskScoreColor = (score: number | null | undefined) => {
  if (score == null) return "bg-muted text-muted-foreground border-border";
  if (score >= 0.6) return "bg-red-500/15 text-red-600 dark:text-red-400 border-red-500/30";
  if (score >= 0.3) return "bg-orange-500/15 text-orange-600 dark:text-orange-400 border-orange-500/30";
  return "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border-emerald-500/30";
};

export const nextStatuses: Record<string, string[]> = {
  open: ["confirmed", "false_positive"],
  confirmed: ["fixed", "false_positive"],
  false_positive: ["open"],
  fixed: ["open"],
};
