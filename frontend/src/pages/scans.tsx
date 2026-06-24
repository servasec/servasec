import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Scan, ArrowRight, ShieldAlert } from "lucide-react";
import Link from "next/link";
import axios from "@/lib/api";
import { toast } from "sonner";
import { statusScanColors } from "@/lib/constants";

interface ApplicationVersion {
  id: number;
  applicationId: number;
  name: string;
}

interface ScannerType {
  id: number;
  name: string;
  description: string;
}

interface ScanItem {
  id: number;
  applicationVersionId: number;
  applicationVersion: ApplicationVersion | null;
  scannerTypeId: number;
  scannerType: ScannerType | null;
  status: string;
  startedAt: string | null;
  completedAt: string | null;
  createdAt: string;
}

interface Application {
  id: number;
  name: string;
}

export default function ScansPage() {
  const router = useRouter();
  const { loggedIn, authChecked } = useAuth();
  const [scans, setScans] = useState<ScanItem[]>([]);
  const [apps, setApps] = useState<Application[]>([]);
  const [versions, setVersions] = useState<ApplicationVersion[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);

  const appFilter = (router.query.applicationId as string) || "";
  const versionFilter = (router.query.applicationVersionId as string) || "";

  const fetchScans = () => {
    const params = new URLSearchParams();
    if (appFilter) params.set("applicationId", appFilter);
    if (versionFilter) params.set("applicationVersionId", versionFilter);

    Promise.all([
      axios.get(`/api/scans?${params}`),
      axios.get("/api/applications"),
    ])
      .then(([scansRes, appsRes]) => {
        setScans(scansRes.data);
        setApps(appsRes.data);
        setLoadError(false);
      })
      .catch(() => { toast.error("Failed to load scans"); setLoadError(true); })
      .finally(() => setLoading(false));
  };

  const fetchVersions = (appId: string) => {
    if (!appId) { setVersions([]); return; }
    axios.get(`/api/applications/${appId}/versions`)
      .then((res) => setVersions(res.data))
      .catch(() => {});
  };

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn && router.isReady) {
      fetchScans();
    }
  }, [authChecked, loggedIn, router.isReady, appFilter, versionFilter]);

  useEffect(() => {
    if (authChecked && loggedIn && router.isReady) {
      fetchVersions(appFilter);
    }
  }, [authChecked, loggedIn, router.isReady, appFilter]);

  const appMap = Object.fromEntries(apps.map((a) => [a.id, a.name]));

  const formatDate = (d: string | null) => {
    if (!d) return "-";
    return new Date(d).toLocaleDateString();
  };

  const handleAppFilter = (v: string) => {
    setLoading(true);
    const value = v === "All" ? "" : v;
    router.replace(
      { pathname: router.pathname, query: value ? { applicationId: value } : {} },
      undefined,
      { shallow: true },
    );
  };

  const handleVersionFilter = (v: string) => {
    setLoading(true);
    const value = v === "All" ? "" : v;
    const newQuery: Record<string, string> = {};
    if (appFilter) newQuery.applicationId = appFilter;
    if (value) newQuery.applicationVersionId = value;
    router.replace(
      { pathname: router.pathname, query: newQuery },
      undefined,
      { shallow: true },
    );
  };

  if (!authChecked || !loggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-muted-foreground">Loading...</div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader crumbs={[{ label: "Scans" }]} />

      <Card>
        <div className="p-4 flex flex-wrap gap-3 border-b">
          <div className="w-44">
            <Select value={appFilter} onValueChange={handleAppFilter}>
              <SelectTrigger>
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
            <Select value={versionFilter} onValueChange={handleVersionFilter}>
              <SelectTrigger>
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
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">Scanner</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden md:table-cell">Application</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Version</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Status</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden lg:table-cell">Started</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden lg:table-cell">Completed</th>
                <th className="w-12 px-4 py-3.5" />
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array.from({ length: 3 }).map((_, i) => (
                  <tr key={i} className="border-b last:border-0">
                    {Array.from({ length: 7 }).map((_, j) => (
                      <td key={j} className="px-4 py-3">
                        <Skeleton className="h-5 w-full max-w-[100px]" />
                      </td>
                    ))}
                  </tr>
                ))
              ) : loadError ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center text-muted-foreground">
                    <ShieldAlert className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>Failed to load scans</p>
                    <p className="text-xs mt-1">The server may be unavailable</p>
                  </td>
                </tr>
              ) : scans.length === 0 ? (
                <tr>
                  <td colSpan={7} className="px-4 py-12 text-center text-muted-foreground">
                    <Scan className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>No scans yet</p>
                    <p className="text-xs mt-1">Run a scan via the API to see results here</p>
                  </td>
                </tr>
              ) : (
                scans.map((s) => (
                  <tr key={s.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3 font-medium">{s.scannerType?.name || "-"}</td>
                    <td className="px-4 py-3 text-muted-foreground hidden md:table-cell">
                      {s.applicationVersion
                        ? (appMap[s.applicationVersion.applicationId] || `App #${s.applicationVersion.applicationId}`)
                        : "-"}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground hidden sm:table-cell">
                      {s.applicationVersion?.name || `#${s.applicationVersionId}`}
                    </td>
                    <td className="px-4 py-3 hidden sm:table-cell">
                      <span className={`inline-flex items-center gap-1.5 text-sm ${statusScanColors[s.status] || ""}`}>
                        <span className="h-1.5 w-1.5 rounded-full bg-current" />
                        {s.status}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground hidden lg:table-cell">{formatDate(s.startedAt)}</td>
                    <td className="px-4 py-3 text-muted-foreground hidden lg:table-cell">{formatDate(s.completedAt)}</td>
                    <td className="px-4 py-3">
                      <Link href={`/findings?scanId=${s.id}`}>
                        <Button variant="ghost" size="icon" className="h-8 w-8" title="View findings">
                          <ArrowRight className="h-3.5 w-3.5" />
                        </Button>
                      </Link>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
