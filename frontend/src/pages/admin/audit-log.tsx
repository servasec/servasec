import { useEffect, useState, useCallback, Fragment } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { RefreshCw, ArrowLeft, ArrowRight, ChevronDown, ChevronRight, ShieldAlert } from "lucide-react";
import axios from "@/lib/api";

interface AuditEntry {
  id: number;
  userId: number;
  username: string;
  action: string;
  resource: string;
  status: number;
  details: string;
  ip: string;
  createdAt: string;
}

interface PaginatedResponse {
  data: AuditEntry[];
  total: number;
  page: number;
  perPage: number;
  totalPages: number;
}

function formatDateTime(iso: string) {
  const d = new Date(iso);
  return d.toLocaleString("fr-FR", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

function statusColor(status: number) {
  if (status >= 200 && status < 300) return "text-green-500";
  if (status >= 400 && status < 500) return "text-orange-500";
  if (status >= 500) return "text-red-500";
  return "text-muted-foreground";
}

const actionColors: Record<string, string> = {
  POST: "text-blue-500",
  PUT: "text-amber-500",
  PATCH: "text-amber-500",
  DELETE: "text-red-500",
};

export default function AuditLogPage() {
  const router = useRouter();
  const { loggedIn, user, authChecked } = useAuth();
  const [data, setData] = useState<PaginatedResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [actionFilter, setActionFilter] = useState("");
  const [resourceFilter, setResourceFilter] = useState("");
  const [statusFilter, setStatusFilter] = useState("");
  const [page, setPage] = useState(1);

  const fetchLogs = useCallback(async () => {
    setLoading(true);
    try {
      const params: Record<string, string> = { page: String(page), perPage: "50" };
      if (actionFilter) params.action = actionFilter;
      if (resourceFilter) params.resource = resourceFilter;
      if (statusFilter) params.status = statusFilter;
      const res = await axios.get("/api/admin/audit-logs", { params });
      setData(res.data);
      setLoadError(false);
    } catch {
      setData(null);
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  }, [page, actionFilter, resourceFilter, statusFilter]);

  useEffect(() => {
    if (!authChecked) return;
    if (!loggedIn) { router.push("/login"); return; }
    if (user?.role !== "admin") { router.push("/"); return; }
    fetchLogs();
  }, [authChecked, loggedIn, user, router, fetchLogs]);

  if (!authChecked || !loggedIn || user?.role !== "admin") {
    return (
      <div className="flex items-center justify-center h-64">
        <Skeleton className="h-8 w-48" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader crumbs={[
        { label: "Administration", href: "#" },
        { label: "Audit Log" },
      ]} />

      <div className="flex flex-wrap gap-3">
        <Select value={actionFilter} onValueChange={(v) => { setActionFilter(v); setPage(1); }}>
          <SelectTrigger className="w-32">
            <SelectValue placeholder="Action" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All</SelectItem>
            <SelectItem value="POST">POST</SelectItem>
            <SelectItem value="PUT">PUT</SelectItem>
            <SelectItem value="PATCH">PATCH</SelectItem>
            <SelectItem value="DELETE">DELETE</SelectItem>
          </SelectContent>
        </Select>

        <Input
          placeholder="Filter by resource path..."
          className="w-64"
          value={resourceFilter}
          onChange={(e) => { setResourceFilter(e.target.value); setPage(1); }}
        />

        <Select value={statusFilter} onValueChange={(v) => { setStatusFilter(v); setPage(1); }}>
          <SelectTrigger className="w-36">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="">All</SelectItem>
            <SelectItem value="200">2xx Success</SelectItem>
            <SelectItem value="400">4xx Error</SelectItem>
            <SelectItem value="500">5xx Error</SelectItem>
          </SelectContent>
        </Select>

        <Button variant="outline" size="icon" onClick={fetchLogs}>
          <RefreshCw className="h-4 w-4" />
        </Button>
      </div>

      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="w-8 px-4 py-3" />
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">Date</th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">User</th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">Action</th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">Resource</th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">Status</th>
                <th className="text-left px-4 py-3 font-medium text-muted-foreground">IP</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array.from({ length: 8 }).map((_, i) => (
                  <tr key={i} className="border-b">
                    <td className="px-4 py-3"><Skeleton className="h-4 w-4" /></td>
                    {Array.from({ length: 6 }).map((_, j) => (
                      <td key={j} className="px-4 py-3">
                        <Skeleton className="h-4 w-24" />
                      </td>
                    ))}
                  </tr>
                ))
              ) : loadError ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center text-muted-foreground">
                    <ShieldAlert className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>Failed to load audit log</p>
                    <p className="text-xs mt-1">The server may be unavailable</p>
                  </td>
                </tr>
              ) : data?.data.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center text-muted-foreground">
                    <p>No audit log entries found</p>
                    <p className="text-xs mt-1">No matching events recorded yet</p>
                  </td>
                </tr>
              ) : (
                data?.data.map((entry) => (
                  <Fragment key={entry.id}>
                    <tr className="border-b hover:bg-muted/30 transition-colors cursor-pointer" onClick={() => setExpandedId(expandedId === entry.id ? null : entry.id)}>
                      <td className="px-4 py-3">
                        {expandedId === entry.id ? (
                          <ChevronDown className="h-4 w-4 text-muted-foreground" />
                        ) : (
                          <ChevronRight className="h-4 w-4 text-muted-foreground" />
                        )}
                      </td>
                      <td className="px-4 py-3 text-muted-foreground text-xs whitespace-nowrap">
                        {formatDateTime(entry.createdAt)}
                      </td>
                      <td className="px-4 py-3">
                        <span className="font-medium">{entry.username}</span>
                        <span className="text-muted-foreground text-xs ml-1">#{entry.userId}</span>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`font-mono text-xs font-semibold ${actionColors[entry.action] || ""}`}>
                          {entry.action}
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        <code className="text-xs bg-muted rounded px-1.5 py-0.5 break-all">
                          {entry.resource}
                        </code>
                      </td>
                      <td className="px-4 py-3">
                        <span className={`font-mono text-xs font-semibold ${statusColor(entry.status)}`}>
                          {entry.status}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-xs text-muted-foreground font-mono">
                        {entry.ip}
                      </td>
                    </tr>
                    {expandedId === entry.id && entry.details && (
                      <tr key={`details-${entry.id}`} className="border-b bg-muted/30">
                        <td colSpan={7} className="px-4 py-3">
                          <pre className="text-xs font-mono whitespace-pre-wrap break-all bg-background rounded p-3 max-h-48 overflow-y-auto">
                            {(() => {
                              try {
                                return JSON.stringify(JSON.parse(entry.details), null, 2);
                              } catch {
                                return entry.details;
                              }
                            })()}
                          </pre>
                        </td>
                      </tr>
                    )}
                  </Fragment>
                ))
              )}
            </tbody>
          </table>
        </div>
      </Card>

      {data && data.totalPages > 1 && (
        <div className="flex items-center justify-between">
          <p className="text-sm text-muted-foreground">
            Page {data.page} of {data.totalPages} ({data.total} entries)
          </p>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              disabled={page <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              <ArrowLeft className="h-4 w-4 mr-1" /> Previous
            </Button>
            <Button
              variant="outline"
              size="sm"
              disabled={page >= data.totalPages}
              onClick={() => setPage((p) => p + 1)}
            >
              Next <ArrowRight className="h-4 w-4 ml-1" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
