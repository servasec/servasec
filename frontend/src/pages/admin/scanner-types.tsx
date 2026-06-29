import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { ShieldAlert } from "lucide-react";
import type { ScannerType } from "@/lib/types";
import axios from "@/lib/api";
import { toast } from "sonner";

export default function ScannerTypesPage() {
  const router = useRouter();
  const { loggedIn, user, authChecked } = useAuth();
  const [scannerTypes, setScannerTypes] = useState<ScannerType[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [togglingId, setTogglingId] = useState<number | null>(null);

  const fetchScannerTypes = async () => {
    setLoading(true);
    try {
      const res = await axios.get("/api/scanner-types");
      setScannerTypes(res.data);
      setLoadError(false);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!authChecked) return;
    if (!loggedIn) { router.push("/login"); return; }
    if (user?.role !== "admin") { router.push("/"); return; }
    fetchScannerTypes();
  }, [authChecked, loggedIn, user, router]);

  const handleToggle = async (st: ScannerType) => {
    setTogglingId(st.id);
    try {
      const res = await axios.put(`/api/admin/scanner-types/${st.id}`, { enabled: !st.enabled });
      setScannerTypes((prev) => prev.map((s) => (s.id === st.id ? res.data : s)));
      toast.success(`Scanner type ${res.data.enabled ? "enabled" : "disabled"}`);
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to update scanner type");
    } finally {
      setTogglingId(null);
    }
  };

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
        { label: "Scanner Types" },
      ]} />

      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">Name</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Description</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Parser</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">Status</th>
                <th className="w-24 px-4 py-3.5" />
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array.from({ length: 4 }).map((_, i) => (
                  <tr key={i} className="border-b">
                    <td className="px-4 py-3"><Skeleton className="h-4 w-24" /></td>
                    <td className="px-4 py-3 hidden sm:table-cell"><Skeleton className="h-4 w-40" /></td>
                    <td className="px-4 py-3 hidden sm:table-cell"><Skeleton className="h-4 w-16" /></td>
                    <td className="px-4 py-3"><Skeleton className="h-4 w-16" /></td>
                    <td className="px-4 py-3"><Skeleton className="h-4 w-12" /></td>
                  </tr>
                ))
              ) : loadError ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-muted-foreground">
                    <ShieldAlert className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>Failed to load scanner types</p>
                    <p className="text-xs mt-1">The server may be unavailable</p>
                  </td>
                </tr>
              ) : scannerTypes.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-muted-foreground">
                    <p>No scanner types found</p>
                  </td>
                </tr>
              ) : (
                scannerTypes.map((st) => (
                  <tr key={st.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3">
                      <span className="font-medium">{st.name}</span>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground hidden sm:table-cell">
                      {st.description}
                    </td>
                    <td className="px-4 py-3 hidden sm:table-cell">
                      <code className="text-xs bg-muted rounded px-1.5 py-0.5">{st.parser}</code>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center gap-1.5 text-sm ${st.enabled ? "text-green-500" : "text-muted-foreground"}`}>
                        <span className="h-1.5 w-1.5 rounded-full bg-current" />
                        {st.enabled ? "Active" : "Disabled"}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleToggle(st)}
                        disabled={togglingId === st.id}
                      >
                        {togglingId === st.id ? "..." : st.enabled ? "Disable" : "Enable"}
                      </Button>
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
