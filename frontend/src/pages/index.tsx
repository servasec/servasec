import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { PageHeader } from "@/components/page-header";
import {
  ShieldAlert,
  Bug,
  Activity,
  Server,
  ArrowUpRight,
  Plus,
  Scan,
  FileText,
} from "lucide-react";
import Link from "next/link";
import axios from "@/lib/api";

const severityColors: Record<string, string> = {
  Critical: "bg-red-500/15 text-red-600 dark:text-red-400 border-red-500/30",
  High: "bg-orange-500/15 text-orange-600 dark:text-orange-400 border-orange-500/30",
  Medium: "bg-yellow-500/15 text-yellow-600 dark:text-yellow-400 border-yellow-500/30",
  Low: "bg-green-500/15 text-green-600 dark:text-green-400 border-green-500/30",
};

const statusColors: Record<string, string> = {
  Open: "text-red-500",
  "In Progress": "text-amber-500",
  Resolved: "text-green-500",
};

const mockFindings = [
  { id: 1, title: "SQL Injection in login endpoint", severity: "Critical", status: "Open", date: "2026-06-15" },
  { id: 2, title: "Stored XSS in user profile", severity: "High", status: "In Progress", date: "2026-06-14" },
  { id: 3, title: "Insecure JWT token handling", severity: "High", status: "Open", date: "2026-06-13" },
  { id: 4, title: "Missing CSP headers", severity: "Medium", status: "Resolved", date: "2026-06-12" },
  { id: 5, title: "Outdated SSL certificate", severity: "Low", status: "Open", date: "2026-06-11" },
];

const statCards = [
  {
    label: "Total Findings",
    value: "47",
    icon: Bug,
    color: "text-blue-500",
    bg: "bg-blue-500/10",
  },
  {
    label: "Critical",
    value: "3",
    icon: ShieldAlert,
    color: "text-red-500",
    bg: "bg-red-500/10",
  },
  {
    label: "Active Scans",
    value: "2",
    icon: Activity,
    color: "text-emerald-500",
    bg: "bg-emerald-500/10",
  },
  {
    label: "Projects",
    value: "12",
    icon: Server,
    color: "text-violet-500",
    bg: "bg-violet-500/10",
  },
];

export default function Dashboard() {
  const router = useRouter();
  const { loggedIn, user, authChecked } = useAuth();
  const [stats, setStats] = useState<any>(null);
  const [statsLoading, setStatsLoading] = useState(true);

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn) {
      axios
        .get("/api/dashboard/stats")
        .then((res) => setStats(res.data))
        .catch(() => {})
        .finally(() => setStatsLoading(false));
    }
  }, [authChecked, loggedIn]);

  if (!authChecked || !loggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-muted-foreground">Loading...</div>
      </div>
    );
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <PageHeader crumbs={[{ label: "Dashboard" }]} />
        <Button className="gap-2">
          <Plus className="h-4 w-4" />
          New Scan
        </Button>
      </div>

      <p className="text-sm text-muted-foreground">
        Welcome back, <span className="text-foreground font-medium">{user?.username}</span>
      </p>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {statCards.map((card) => {
          const Icon = card.icon;
          return (
            <Card key={card.label} className="border border-border/50">
              <CardContent className="p-5">
                <div className="flex items-start justify-between">
                  <div className="space-y-1.5">
                    <p className="text-sm text-muted-foreground">{card.label}</p>
                    <p className="text-2xl font-bold">{card.value}</p>
                  </div>
                  <div className={`rounded-lg p-2 ${card.bg}`}>
                    <Icon className={`h-5 w-5 ${card.color}`} />
                  </div>
                </div>
              </CardContent>
            </Card>
          );
        })}
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-base font-semibold">Recent Findings</h3>
            <Link
              href="/findings"
              className="text-sm text-primary hover:text-primary/80 flex items-center gap-1"
            >
              View all
              <ArrowUpRight className="h-3.5 w-3.5" />
            </Link>
          </div>

          <div className="rounded-lg border">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="text-left px-4 py-3 font-medium text-muted-foreground">Finding</th>
                    <th className="text-left px-4 py-3 font-medium text-muted-foreground hidden sm:table-cell">Severity</th>
                    <th className="text-left px-4 py-3 font-medium text-muted-foreground hidden md:table-cell">Status</th>
                    <th className="text-left px-4 py-3 font-medium text-muted-foreground hidden lg:table-cell">Date</th>
                  </tr>
                </thead>
                <tbody>
                  {mockFindings.map((f) => (
                    <tr key={f.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-3 font-medium">{f.title}</td>
                      <td className="px-4 py-3 hidden sm:table-cell">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${
                            severityColors[f.severity] || ""
                          }`}
                        >
                          {f.severity}
                        </span>
                      </td>
                      <td className={`px-4 py-3 hidden md:table-cell ${statusColors[f.status] || ""}`}>
                        <span className="inline-flex items-center gap-1.5 text-sm">
                          <span className="h-1.5 w-1.5 rounded-full bg-current" />
                          {f.status}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-muted-foreground hidden lg:table-cell">{f.date}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <div className="space-y-4">
          <h3 className="text-base font-semibold">Quick Actions</h3>
          <div className="grid gap-3">
            <Button variant="outline" className="justify-start gap-3 h-auto py-3 px-4">
              <div className="rounded-lg bg-primary/10 p-2">
                <Scan className="h-4 w-4 text-primary" />
              </div>
              <div className="text-left">
                <p className="text-sm font-medium">Start a Scan</p>
                <p className="text-xs text-muted-foreground">Run security scan on a project</p>
              </div>
            </Button>
            <Button variant="outline" className="justify-start gap-3 h-auto py-3 px-4">
              <div className="rounded-lg bg-emerald-500/10 p-2">
                <FileText className="h-4 w-4 text-emerald-500" />
              </div>
              <div className="text-left">
                <p className="text-sm font-medium">Generate Report</p>
                <p className="text-xs text-muted-foreground">Export findings summary</p>
              </div>
            </Button>
            <Button variant="outline" className="justify-start gap-3 h-auto py-3 px-4">
              <div className="rounded-lg bg-violet-500/10 p-2">
                <Server className="h-4 w-4 text-violet-500" />
              </div>
              <div className="text-left">
                <p className="text-sm font-medium">Add Project</p>
                <p className="text-xs text-muted-foreground">Register a new project</p>
              </div>
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
