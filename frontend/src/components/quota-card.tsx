import { useEffect, useState } from "react";
import axios from "@/lib/api";
import type { QuotaStatus } from "@/lib/types";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import { Skeleton } from "@/components/ui/skeleton";
import { Gauge } from "lucide-react";

const LABELS: Record<QuotaStatus["resource"], string> = {
  groups: "Groups",
  applications: "Applications",
  versions: "Versions per app",
  users: "Users",
};

export default function QuotaCard() {
  const [statuses, setStatuses] = useState<QuotaStatus[] | null>(null);
  const [error, setError] = useState(false);

  useEffect(() => {
    axios
      .get("/api/me/quotas")
      .then((res) => setStatuses(res.data.resources || []))
      .catch(() => setError(true));
  }, []);

  if (error) return null;

  const limited = (statuses || []).filter((s) => s.limit != null);

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center gap-3">
          <Gauge className="h-5 w-5 text-muted-foreground" />
          <div>
            <CardTitle>Plan limits</CardTitle>
            <p className="text-sm text-muted-foreground mt-0.5">
              Usage against your current license
            </p>
          </div>
        </div>
      </CardHeader>
      <CardContent>
        {!statuses ? (
          <div className="space-y-4">
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-full" />
          </div>
        ) : limited.length === 0 ? (
          <p className="text-sm text-muted-foreground">
            Unlimited - no license quotas apply to this instance.
          </p>
        ) : (
          <div className="space-y-5">
            {limited.map((s) => {
              const pct = Math.min(100, Math.round((s.usage / s.limit!) * 100));
              const nearLimit = pct >= 80;
              return (
                <div key={s.resource} className="space-y-1.5">
                  <div className="flex items-center justify-between text-sm">
                    <span className="font-medium">{LABELS[s.resource]}</span>
                    <span
                      className={
                        nearLimit
                          ? "text-amber-600 dark:text-amber-400 font-medium"
                          : "text-muted-foreground"
                      }
                    >
                      {s.usage} / {s.limit}
                    </span>
                  </div>
                  <Progress value={pct} className={nearLimit ? "bg-amber-500/20" : ""} />
                </div>
              );
            })}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
