import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import {
  Bug, Activity, User, Clock, ArrowUpRight, Plus, Server, ShieldAlert,
} from "lucide-react";
import Link from "next/link";
import axios from "@/lib/api";
import type { DashboardStats } from "@/lib/types";
import { severityBarColors, severityColors, statusColors } from "@/lib/constants";

const kpiConfig = [
  { label: "Total findings", key: "totalFindings" as const, icon: Bug, color: "text-blue-500", bg: "bg-blue-500/10" },
  { label: "Assigned to me & still open", key: "myOpenFindings" as const, icon: User, color: "text-amber-500", bg: "bg-amber-500/10" },
  { label: "Overdue", key: "overdueFindings" as const, icon: Clock, color: "text-red-500", bg: "bg-red-500/10" },
  { label: "Recent scans", key: "recentScans" as const, icon: Activity, color: "text-emerald-500", bg: "bg-emerald-500/10" },
];

export default function Dashboard() {
  const router = useRouter();
  const { loggedIn, user, authChecked } = useAuth();
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn) {
      axios
        .get("/api/dashboard/stats")
        .then((res) => { setStats(res.data); setLoadError(false); })
        .catch(() => { setLoadError(true); })
        .finally(() => setLoading(false));
    }
  }, [authChecked, loggedIn]);

  if (!authChecked || !loggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-muted-foreground">Loading...</div>
      </div>
    );
  }

  const totalSeverity = stats?.bySeverity.reduce((a, b) => a + b.count, 0) || 1;

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <PageHeader crumbs={[{ label: "Dashboard" }]} />
        <Link href="/applications">
          <Button size="sm" className="gap-2">
            <Plus className="h-4 w-4" />
            New scan
          </Button>
        </Link>
      </div>

      <p className="text-sm text-muted-foreground">
        Welcome back, <span className="text-foreground font-medium">{user?.username}</span>
      </p>

      {loadError ? (
        <Card>
          <CardContent className="p-8 text-center">
            <ShieldAlert className="h-8 w-8 mx-auto mb-3 text-destructive opacity-60" />
            <p className="text-sm font-medium text-foreground">Failed to load dashboard data</p>
            <p className="text-xs text-muted-foreground mt-1">The server may be unavailable. Try refreshing the page.</p>
          </CardContent>
        </Card>
      ) : loading ? (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <Card key={i}><CardContent className="p-5"><Skeleton className="h-5 w-24 mb-2" /><Skeleton className="h-8 w-16" /></CardContent></Card>
          ))}
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {kpiConfig.map((cfg) => {
            const Icon = cfg.icon;
            const value = stats ? stats[cfg.key] : 0;
            const isOverdue = cfg.key === "overdueFindings" && value > 0;
            const isMyOpen = cfg.key === "myOpenFindings" && value > 0;
            return (
              <Card key={cfg.key} className="border border-border/50">
                <CardContent className="p-5">
                  <div className="flex items-start justify-between">
                    <div className="space-y-1.5">
                      <p className="text-sm text-muted-foreground">{cfg.label}</p>
                      <p className={`text-2xl font-bold ${isOverdue ? "text-red-500" : isMyOpen ? "text-amber-500" : ""}`}>
                        {value}
                      </p>
                    </div>
                    <div className={`rounded-lg p-2 ${cfg.bg}`}>
                      <Icon className={`h-5 w-5 ${cfg.color}`} />
                    </div>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      {loading ? (
        <div className="grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2 space-y-4">
            <Skeleton className="h-6 w-48" />
            <Skeleton className="h-40 w-full" />
          </div>
          <div className="space-y-4">
            <Skeleton className="h-6 w-32" />
            <Skeleton className="h-32 w-full" />
          </div>
        </div>
      ) : stats ? (
        <div className="grid gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2 space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Severity distribution</CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                {stats.bySeverity.map((s) => (
                  <div key={s.severity} className="space-y-1">
                    <div className="flex items-center justify-between text-sm">
                      <span className="capitalize font-medium">{s.severity}</span>
                      <span className="text-muted-foreground">{s.count}</span>
                    </div>
                    <div className="h-2 rounded-full bg-muted overflow-hidden">
                      <div
                        className="h-full rounded-full transition-all"
                        style={{
                          width: `${(s.count / totalSeverity) * 100}%`,
                          backgroundColor: severityBarColors[s.severity.toLowerCase()] || "hsl(var(--primary))",
                        }}
                      />
                    </div>
                  </div>
                ))}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle className="text-base">Top findings</CardTitle>
                  <Link href="/findings" className="text-sm text-primary hover:text-primary/80 flex items-center gap-1">
                    View all <ArrowUpRight className="h-3.5 w-3.5" />
                  </Link>
                </div>
              </CardHeader>
              <CardContent className="p-0">
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b bg-muted/50">
                        <th className="text-left px-4 py-3 font-medium text-muted-foreground">Rule</th>
                        <th className="text-left px-4 py-3 font-medium text-muted-foreground">Title</th>
                        <th className="text-right px-4 py-3 font-medium text-muted-foreground">Count</th>
                      </tr>
                    </thead>
                    <tbody>
                      {stats.topFindings.length === 0 ? (
                        <tr><td colSpan={3} className="px-4 py-8 text-center text-muted-foreground">No findings yet</td></tr>
                      ) : (
                        stats.topFindings.map((f, i) => (
                          <tr key={i} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                            <td className="px-4 py-3"><code className="text-xs bg-muted px-1.5 py-0.5 rounded">{f.ruleId}</code></td>
                            <td className="px-4 py-3 max-w-[300px] truncate">{f.title}</td>
                            <td className="px-4 py-3 text-right font-medium">{f.count}</td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </CardContent>
            </Card>
          </div>

          <div className="space-y-6">
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Findings by status</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                {stats.byStatus.map((s) => (
                  <div key={s.status} className="flex items-center justify-between text-sm">
                    <span className="inline-flex items-center gap-1.5">
                      <span className={`h-1.5 w-1.5 rounded-full bg-current ${statusColors[s.status] || ""}`} />
                      <span className="capitalize">{s.status.replace("_", " ")}</span>
                    </span>
                    <span className="font-medium">{s.count}</span>
                  </div>
                ))}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle className="text-base">Findings by scanner</CardTitle>
              </CardHeader>
              <CardContent className="space-y-2">
                {stats.byScanner.length === 0 ? (
                  <p className="text-sm text-muted-foreground">No scanner data</p>
                ) : (
                  stats.byScanner.map((s, i) => (
                    <div key={i} className="flex items-center justify-between text-sm">
                      <span>{s.scannerType}</span>
                      <span className="font-medium">{s.count}</span>
                    </div>
                  ))
                )}
              </CardContent>
            </Card>

            {user?.features?.includes("risk_scoring") && stats.avgRiskScore != null && (
              <>
                <Card>
                  <CardHeader className="pb-3">
                    <CardTitle className="text-sm font-medium text-muted-foreground">Average risk score</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <div className={`text-3xl font-bold ${stats.avgRiskScore >= 0.6 ? "text-red-500" : stats.avgRiskScore >= 0.3 ? "text-orange-500" : "text-emerald-500"}`}>
                      {stats.avgRiskScore.toFixed(2)}
                    </div>
                  </CardContent>
                </Card>
                {stats.riskDistribution && stats.riskDistribution.length > 0 && (
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">Risk distribution</CardTitle>
                    </CardHeader>
                    <CardContent className="space-y-2">
                      {[...stats.riskDistribution].reverse().map((d) => {
                        const maxCount = Math.max(...stats.riskDistribution!.map((r) => r.count), 1);
                        return (
                          <div key={d.label} className="space-y-0.5">
                            <div className="flex items-center justify-between text-sm">
                              <span className="text-muted-foreground">{d.label}</span>
                              <span className="font-medium">{d.count}</span>
                            </div>
                            <div className="h-2 rounded-full bg-muted overflow-hidden">
                              <div
                                className="h-full rounded-full bg-primary transition-all"
                                style={{ width: `${(d.count / maxCount) * 100}%` }}
                              />
                            </div>
                          </div>
                        );
                      })}
                    </CardContent>
                  </Card>
                )}
                {stats.topRiskyFindings && stats.topRiskyFindings.length > 0 && (
                  <Card>
                    <CardHeader>
                      <CardTitle className="text-base">Top risky findings</CardTitle>
                    </CardHeader>
                    <CardContent className="p-0">
                      <div className="divide-y">
                        {stats.topRiskyFindings.map((f) => (
                          <Link key={f.id} href={`/findings/${f.id}`} className="flex items-center gap-3 px-4 py-2.5 hover:bg-muted/30 transition-colors">
                            <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium border ${severityColors[f.severity?.toLowerCase()] || ""}`}>
                              {f.severity}
                            </span>
                            <span className="flex-1 min-w-0 text-sm truncate">{f.title}</span>
                            <span className={`text-xs font-medium shrink-0 ${f.riskScore >= 0.6 ? "text-red-500" : f.riskScore >= 0.3 ? "text-orange-500" : "text-emerald-500"}`}>
                              {f.riskScore.toFixed(2)}
                            </span>
                          </Link>
                        ))}
                      </div>
                    </CardContent>
                  </Card>
                )}
              </>
            )}

          </div>
        </div>
      ) : null}
    </div>
  );
}
