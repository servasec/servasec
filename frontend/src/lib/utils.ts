import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

const QUOTA_RESOURCE_LABELS: Record<string, string> = {
  groups: "groups",
  applications: "applications",
  versions: "versions",
  users: "users",
}

export function quotaErrorMessage(err: any, fallback: string): string {
  const error = err?.response?.data?.error;
  if (error === "quota_exceeded") {
    const resource = QUOTA_RESOURCE_LABELS[err?.response?.data?.resource] ?? "resource";
    const limit = err?.response?.data?.limit;
    if (limit != null) {
      return `Plan limit reached: this instance allows up to ${limit} ${resource}. Upgrade your plan to continue.`;
    }
    return `Plan limit reached for ${resource}. Upgrade your plan to continue.`;
  }
  return err?.response?.data?.error || fallback;
}
