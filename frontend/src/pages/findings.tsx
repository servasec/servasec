import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Bug, ChevronDown, User as UserIcon, Clock, ChevronLeft, ChevronRight, ShieldAlert } from "lucide-react";
import Link from "next/link";
import axios from "@/lib/api";
import { toast } from "sonner";
import type { Finding, Application, ScannerType, ApplicationVersion } from "@/lib/types";
import { severityBadgeColors, statusColors, statusLabels, nextStatuses, severityOptions, riskScoreColor } from "@/lib/constants";

const filterKeys = ["applicationId", "applicationVersionId", "scanId", "severity", "status", "scannerTypeId", "assignedTo"] as const;

export default function FindingsPage() {
  const router = useRouter();
  const { loggedIn, user, authChecked } = useAuth();
  const hasRisk = user?.features?.includes("risk_scoring");
  const [findings, setFindings] = useState<Finding[]>([]);
  const [apps, setApps] = useState<Application[]>([]);
  const [scannerTypes, setScannerTypes] = useState<ScannerType[]>([]);
  const [versions, setVersions] = useState<ApplicationVersion[]>([]);
  const [scans, setScans] = useState<{ id: number; createdAt: string; scannerType?: { name: string } | null }[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const perPage = 50;
  const [sortBy, setSortBy] = useState("");
  const [order, setOrder] = useState("desc");

  const filters: Record<string, string> = {};
  for (const key of filterKeys) {
    const val = router.query[key];
    filters[key] = typeof val === "string" ? val : "";
  }

  const fetchFindings = (p: number, sort?: string, ord?: string) => {
    const params = new URLSearchParams();
    for (const key of filterKeys) {
      if (filters[key]) params.set(key, filters[key]);
    }
    params.set("page", String(p));
    params.set("perPage", String(perPage));
    if (sort && hasRisk) {
      params.set("sortBy", sort);
      params.set("order", ord || "desc");
    }

    Promise.all([
      axios.get(`/api/findings?${params}`),
      axios.get("/api/applications"),
      axios.get("/api/scanner-types", { params: { enabled: "true" } }),
    ])
      .then(([findingsRes, appsRes, stRes]) => {
        setFindings(findingsRes.data.data || []);
        setTotal(findingsRes.data.total || 0);
        setPage(findingsRes.data.page || 1);
        setApps(appsRes.data);
        setScannerTypes(stRes.data);
      })
      .catch(() => { toast.error("Failed to load findings"); setLoadError(true); })
      .finally(() => setLoading(false));
  };

  const goToPage = (p: number) => {
    if (p < 1 || (total > 0 && p > Math.ceil(total / perPage))) return;
    setLoading(true);
    fetchFindings(p);
  };

  const fetchVersions = (appId: string) => {
    if (!appId) { setVersions([]); return; }
    axios.get(`/api/applications/${appId}/versions`)
      .then((res) => setVersions(res.data))
      .catch(() => {});
  };

  const fetchScans = (appId: string) => {
    if (!appId) { setScans([]); return; }
    axios.get(`/api/scans?applicationId=${appId}`)
      .then((res) => setScans(res.data || []))
      .catch(() => {});
  };

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn && router.isReady) {
      fetchFindings(1, sortBy, order);
    }
  }, [authChecked, loggedIn, router.isReady, filters.applicationId, filters.applicationVersionId, filters.scanId, filters.severity, filters.status, filters.scannerTypeId, filters.assignedTo, sortBy, order]);

  useEffect(() => {
    if (authChecked && loggedIn && router.isReady) {
      fetchVersions(filters.applicationId);
      fetchScans(filters.applicationId);
    }
  }, [authChecked, loggedIn, router.isReady, filters.applicationId]);

  const appMap = Object.fromEntries(apps.map((a) => [a.id, a.name]));

  const toggleSort = () => {
    if (sortBy === "risk_score") {
      setOrder(order === "desc" ? "asc" : "desc");
    } else {
      setSortBy("risk_score");
      setOrder("desc");
    }
  };

  const handleStatusChange = async (findingId: number, newStatus: string) => {
    try {
      await axios.patch(`/api/findings/${findingId}/status`, { status: newStatus });
      toast.success(`Status updated to ${statusLabels[newStatus] || newStatus}`);
      fetchFindings(page, sortBy, order);
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to update status");
    }
  };

  const applyFilter = (key: string, value: string) => {
    setLoading(true);
    const newQuery: Record<string, string> = {};
    for (const k of filterKeys) {
      if (k === key) {
        if (value) newQuery[k] = value;
      } else if ((k === "applicationVersionId" || k === "scanId") && key === "applicationId") {
        continue;
      } else if (filters[k]) {
        newQuery[k] = filters[k];
      }
    }
    router.replace({ pathname: router.pathname, query: newQuery }, undefined, { shallow: true });
  };

  const isOverdue = (f: Finding) => {
    if (!f.dueDate) return false;
    if (f.status === "fixed" || f.status === "false_positive") return false;
    return new Date(f.dueDate) < new Date();
  };

  if (!authChecked || !loggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-muted-foreground">Loading...</div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full">
      <PageHeader crumbs={[{ label: "Findings" }]} />

      {loadError ? (
        <Card>
          <div className="p-8 text-center">
            <ShieldAlert className="h-8 w-8 mx-auto mb-3 text-destructive opacity-60" />
            <p className="text-sm font-medium text-foreground">Failed to load findings</p>
            <p className="text-xs text-muted-foreground mt-1">The server may be unavailable. Try refreshing the page.</p>
          </div>
        </Card>
      ) : (
      <Card className="flex flex-col flex-1 min-h-0 mt-4">
        <div className="shrink-0 p-3 flex flex-wrap gap-2 border-b items-center">
          <div className="w-44">
            <Select value={filters.applicationId} onValueChange={(v) => applyFilter("applicationId", v === "All" ? "" : v)}>
              <SelectTrigger className="h-8 text-xs">
                <SelectValue placeholder="All applications" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="All">All applications</SelectItem>
                {apps.map((a) => (
                  <SelectItem key={a.id} value={String(a.id)}>{a.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="w-44">
            <Select value={filters.applicationVersionId} onValueChange={(v) => applyFilter("applicationVersionId", v === "All" ? "" : v)}>
              <SelectTrigger className="h-8 text-xs">
                <SelectValue placeholder="All versions" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="All">All versions</SelectItem>
                {versions.map((v) => (
                  <SelectItem key={v.id} value={String(v.id)}>{v.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="w-36">
            <Select value={filters.scanId} onValueChange={(v) => applyFilter("scanId", v === "All" ? "" : v)}>
              <SelectTrigger className="h-8 text-xs">
                <SelectValue placeholder="All scans" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="All">All scans</SelectItem>
                {scans.map((s) => (
                  <SelectItem key={s.id} value={String(s.id)}>
                    Scan #{s.id}{s.scannerType ? ` (${s.scannerType.name})` : ""} - {new Date(s.createdAt).toLocaleDateString()}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="w-32">
            <Select value={filters.severity} onValueChange={(v) => applyFilter("severity", v === "All" ? "" : v)}>
              <SelectTrigger className="h-8 text-xs">
                <SelectValue placeholder="All severities" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="All">All severities</SelectItem>
                {severityOptions.map((s) => (
                  <SelectItem key={s} value={s}>{s}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="w-36">
            <Select value={filters.status} onValueChange={(v) => applyFilter("status", v === "All" ? "" : v)}>
              <SelectTrigger className="h-8 text-xs">
                <SelectValue placeholder="All statuses" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="All">All statuses</SelectItem>
                {Object.entries(statusLabels).map(([k, v]) => (
                  <SelectItem key={k} value={k}>{v}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="w-36">
            <Select value={filters.scannerTypeId} onValueChange={(v) => applyFilter("scannerTypeId", v === "All" ? "" : v)}>
              <SelectTrigger className="h-8 text-xs">
                <SelectValue placeholder="All scanners" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="All">All scanners</SelectItem>
                {scannerTypes.map((st) => (
                  <SelectItem key={st.id} value={String(st.id)}>{st.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <Button
            variant={filters.assignedTo === "me" ? "default" : "outline"}
            size="sm"
            className="gap-1.5 h-8 text-xs"
            onClick={() => applyFilter("assignedTo", filters.assignedTo === "me" ? "" : "me")}
          >
            <UserIcon className="h-3.5 w-3.5" />
            My findings
          </Button>
        </div>
        <div className="flex-1 min-h-0 overflow-y-auto">
          <table className="w-full text-xs">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground">Finding</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden sm:table-cell">Severity</th>
                {hasRisk && (
                  <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden md:table-cell cursor-pointer hover:text-foreground select-none" onClick={toggleSort}>
                    Risk {sortBy === "risk_score" ? (order === "desc" ? "↓" : "↑") : ""}
                  </th>
                )}
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden md:table-cell">Status</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden lg:table-cell">Assigned to</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden lg:table-cell">Due date</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden lg:table-cell">Version</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden lg:table-cell">Scanner</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array.from({ length: 4 }).map((_, i) => (
                  <tr key={i} className="border-b last:border-0">
                    {Array.from({ length: hasRisk ? 8 : 7 }).map((_, j) => (
                      <td key={j} className="px-4 py-2">
                        <Skeleton className="h-4 w-full max-w-[140px]" />
                      </td>
                    ))}
                  </tr>
                ))
              ) : findings.length === 0 ? (
                <tr>
                  <td colSpan={hasRisk ? 8 : 7} className="px-4 py-12 text-center text-muted-foreground">
                    <Bug className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>No findings found</p>
                    <p className="text-xs mt-1">Run a scan or adjust your filters</p>
                  </td>
                </tr>
              ) : (
                findings.map((f) => (
                  <tr key={f.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors cursor-pointer" onClick={() => router.push(`/findings/${f.id}`)}>
                    <td className="px-4 py-2">
                      <span className="font-medium hover:text-primary transition-colors">
                        {f.title}
                      </span>
                      {f.applicationVersion && (
                        <p className="text-xs text-muted-foreground mt-0.5">
                          {f.applicationVersion.name}
                          <span className="mx-1">·</span>
                          {appMap[f.applicationVersion.applicationId] || `App #${f.applicationVersion.applicationId}`}
                        </p>
                      )}
                    </td>
                    <td className="px-4 py-2 hidden sm:table-cell">
                      <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium border ${severityBadgeColors[f.severity] || ""}`}>
                        {f.severity}
                      </span>
                    </td>
                    {hasRisk && (
                      <td className="px-4 py-2 hidden md:table-cell">
                        <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium border ${riskScoreColor(f.riskScore)}`}>
                          {f.riskScore != null ? f.riskScore.toFixed(2) : "N/A"}
                        </span>
                      </td>
                    )}
                    <td className="px-4 py-2 hidden md:table-cell" onClick={(e) => e.stopPropagation()}>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <button className={`inline-flex items-center gap-1.5 text-xs hover:underline ${statusColors[f.status] || ""}`}>
                            <span className="h-1.5 w-1.5 rounded-full bg-current" />
                            {statusLabels[f.status] || f.status}
                            <ChevronDown className="h-3 w-3" />
                          </button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="start">
                          {nextStatuses[f.status]?.map((ns) => (
                            <DropdownMenuItem key={ns} onClick={() => handleStatusChange(f.id, ns)}>
                              <span className={`h-1.5 w-1.5 rounded-full bg-current mr-2 ${statusColors[ns] || ""}`} />
                              {statusLabels[ns] || ns}
                            </DropdownMenuItem>
                          ))}
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </td>
                    <td className="px-4 py-2 text-muted-foreground hidden lg:table-cell">
                      {f.assignedToUser ? (
                        <span className="inline-flex items-center gap-1">
                          <UserIcon className="h-3 w-3" />
                          {f.assignedToUser.username}
                        </span>
                      ) : "-"}
                    </td>
                    <td className="px-4 py-2 hidden lg:table-cell">
                      {f.dueDate ? (
                        <span className={`inline-flex items-center gap-1 ${isOverdue(f) ? "text-red-500 font-medium" : "text-muted-foreground"}`}>
                          <Clock className="h-3 w-3" />
                          {new Date(f.dueDate).toLocaleDateString()}
                        </span>
                      ) : "-"}
                    </td>
                    <td className="px-4 py-2 text-muted-foreground hidden lg:table-cell">
                       <span className="text-[11px] font-mono">{f.applicationVersion?.name || `#${f.applicationVersionId}`}</span>
                     </td>
                     <td className="px-4 py-2 text-muted-foreground hidden lg:table-cell">
                       <span className="text-[11px] font-mono">{f.scannerType?.name || `#${f.scannerTypeId}`}</span>
                     </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        {!loading && total > perPage && (
          <div className="shrink-0 flex items-center justify-between px-4 py-2 border-t text-xs">
            <span className="text-muted-foreground">
              {total} finding{(total) !== 1 ? "s" : ""}
            </span>
            <div className="flex items-center gap-1">
              <Button variant="ghost" size="sm" disabled={page <= 1} onClick={() => goToPage(page - 1)}>
                <ChevronLeft className="h-4 w-4" />
              </Button>
              {(() => {
                const totalPages = Math.ceil(total / perPage);
                const pages: (number | "...")[] = [];
                if (totalPages <= 7) {
                  for (let i = 1; i <= totalPages; i++) pages.push(i);
                } else {
                  pages.push(1);
                  if (page > 3) pages.push("...");
                  for (let i = Math.max(2, page - 1); i <= Math.min(totalPages - 1, page + 1); i++) {
                    pages.push(i);
                  }
                  if (page < totalPages - 2) pages.push("...");
                  pages.push(totalPages);
                }
                return pages.map((p, i) =>
                  p === "..." ? (
                    <span key={`ellipsis-${i}`} className="px-1 text-muted-foreground">...</span>
                  ) : (
                    <Button key={p} variant={p === page ? "default" : "ghost"} size="sm" className="min-w-[32px]" onClick={() => goToPage(p)}>
                      {p}
                    </Button>
                  )
                );
              })()}
              <Button variant="ghost" size="sm" disabled={page >= Math.ceil(total / perPage)} onClick={() => goToPage(page + 1)}>
                <ChevronRight className="h-4 w-4" />
              </Button>
            </div>
          </div>
        )}
      </Card>
      )}
    </div>
  );
}
