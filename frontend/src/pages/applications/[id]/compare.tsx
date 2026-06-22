import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ChevronLeft, GitCompare, Bug, CheckCircle2, PlusCircle, MinusCircle } from "lucide-react";
import axios from "@/lib/api";
import { toast } from "sonner";
import type { ApplicationVersion, Finding, CompareResult } from "@/lib/types";

const severityColors: Record<string, string> = {
  Critical: "bg-red-500/15 text-red-600 dark:text-red-400 border-red-500/30",
  High: "bg-orange-500/15 text-orange-600 dark:text-orange-400 border-orange-500/30",
  Medium: "bg-yellow-500/15 text-yellow-600 dark:text-yellow-400 border-yellow-500/30",
  Low: "bg-green-500/15 text-green-600 dark:text-green-400 border-green-500/30",
  Info: "bg-blue-500/15 text-blue-600 dark:text-blue-400 border-blue-500/30",
};

export default function CompareVersionsPage() {
  const router = useRouter();
  const { id } = router.query;
  const { loggedIn, authChecked } = useAuth();
  const [versions, setVersions] = useState<ApplicationVersion[]>([]);
  const [loadingVersions, setLoadingVersions] = useState(true);
  const [fromId, setFromId] = useState("");
  const [toId, setToId] = useState("");
  const [result, setResult] = useState<CompareResult | null>(null);
  const [comparing, setComparing] = useState(false);

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn && id) {
      axios.get(`/api/applications/${id}/versions`)
        .then((res) => setVersions(res.data))
        .catch(() => toast.error("Failed to load versions"))
        .finally(() => setLoadingVersions(false));
    }
  }, [authChecked, loggedIn, id]);

  const handleCompare = async () => {
    if (!fromId || !toId || !id) return;
    setComparing(true);
    setResult(null);
    try {
      const res = await axios.get(`/api/applications/${id}/versions/compare?from=${fromId}&to=${toId}`);
      setResult(res.data);
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to compare versions");
    } finally {
      setComparing(false);
    }
  };

  if (!authChecked || !loggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-muted-foreground">Loading...</div>
      </div>
    );
  }

  const renderFindingsTable = (findings: Finding[], variant: "fixed" | "new" | "stillPresent") => {
    const config = {
      fixed: { icon: CheckCircle2, color: "text-emerald-500", bg: "bg-emerald-500/5", label: "Fixed" },
      new: { icon: PlusCircle, color: "text-red-500", bg: "bg-red-500/5", label: "New" },
      stillPresent: { icon: MinusCircle, color: "text-muted-foreground", bg: "bg-muted/20", label: "Still present" },
    };
    const cfg = config[variant];
    const Icon = cfg.icon;

    return (
      <Card className={cfg.bg}>
        <CardHeader>
          <CardTitle className="text-base flex items-center gap-2">
            <Icon className={`h-4 w-4 ${cfg.color}`} />
            {cfg.label}
            <span className="text-sm font-normal text-muted-foreground ml-auto">{findings.length}</span>
          </CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {findings.length === 0 ? (
            <p className="px-4 pb-4 text-sm text-muted-foreground">No findings</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-muted/30">
                    <th className="text-left px-4 py-2.5 font-medium text-muted-foreground text-xs">Rule</th>
                    <th className="text-left px-4 py-2.5 font-medium text-muted-foreground text-xs">Title</th>
                    <th className="text-left px-4 py-2.5 font-medium text-muted-foreground text-xs hidden sm:table-cell">Severity</th>
                    <th className="text-left px-4 py-2.5 font-medium text-muted-foreground text-xs hidden lg:table-cell">File</th>
                  </tr>
                </thead>
                <tbody>
                  {findings.map((f) => (
                    <tr key={f.ruleId + f.filePath} className="border-b last:border-0 hover:bg-muted/20 transition-colors">
                      <td className="px-4 py-2.5">
                        <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{f.ruleId}</code>
                      </td>
                      <td className="px-4 py-2.5 max-w-[300px] truncate">{f.title}</td>
                      <td className="px-4 py-2.5 hidden sm:table-cell">
                        <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium border ${severityColors[f.severity] || ""}`}>
                          {f.severity}
                        </span>
                      </td>
                      <td className="px-4 py-2.5 text-muted-foreground text-xs hidden lg:table-cell max-w-[200px] truncate">
                        {f.filePath || "-"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>
    );
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => router.push(`/applications/${id}`)} className="h-8 w-8 shrink-0">
          <ChevronLeft className="h-4 w-4" />
        </Button>
        <PageHeader crumbs={[{ label: "Applications", href: "/applications" }, { label: "Version comparison" }]} />
      </div>

      <Card>
        <CardContent className="p-4">
          <div className="flex flex-wrap items-end gap-3">
            <div className="w-44">
              <Select value={fromId} onValueChange={setFromId}>
                <SelectTrigger>
                  <SelectValue placeholder="From version" />
                </SelectTrigger>
                <SelectContent>
                  {versions.map((v) => (
                    <SelectItem key={v.id} value={String(v.id)}>{v.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="w-44">
              <Select value={toId} onValueChange={setToId}>
                <SelectTrigger>
                  <SelectValue placeholder="To version" />
                </SelectTrigger>
                <SelectContent>
                  {versions.map((v) => (
                    <SelectItem key={v.id} value={String(v.id)}>{v.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <Button onClick={handleCompare} disabled={!fromId || !toId || fromId === toId || comparing} className="gap-2">
              <GitCompare className="h-4 w-4" />
              {comparing ? "Comparing..." : "Compare"}
            </Button>
          </div>
        </CardContent>
      </Card>

      {loadingVersions ? (
        <div className="space-y-4">
          <Skeleton className="h-48 w-full" />
        </div>
      ) : comparing ? (
        <div className="space-y-6">
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <Skeleton className="h-5 w-24 inline-block" />
            <span>→</span>
            <Skeleton className="h-5 w-24 inline-block" />
          </div>
          {[
            { label: "Fixed", color: "bg-emerald-500/5" },
            { label: "New", color: "bg-red-500/5" },
            { label: "Still present", color: "bg-muted/20" },
          ].map((cfg) => (
            <Card key={cfg.label} className={cfg.color}>
              <CardHeader>
                <CardTitle className="text-base">{cfg.label}</CardTitle>
              </CardHeader>
              <CardContent className="p-0">
                <div className="p-4 space-y-3">
                  {Array.from({ length: 3 }).map((_, i) => (
                    <div key={i} className="flex items-center gap-3">
                      <Skeleton className="h-4 w-16" />
                      <Skeleton className="h-4 flex-1" />
                      <Skeleton className="h-5 w-14" />
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : result ? (
        <div className="space-y-6">
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <span className="font-medium text-foreground">{result.from.name}</span>
            <span>→</span>
            <span className="font-medium text-foreground">{result.to.name}</span>
          </div>
          {renderFindingsTable(result.fixed || [], "fixed")}
          {renderFindingsTable(result.new || [], "new")}
          {renderFindingsTable(result.stillPresent || [], "stillPresent")}
        </div>
      ) : null}
    </div>
  );
}
