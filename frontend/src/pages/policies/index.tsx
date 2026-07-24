import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { ShieldAlert, Plus, Trash2 } from "lucide-react";
import Link from "next/link";
import axios from "@/lib/api";
import { toast } from "sonner";
import type { Policy } from "@/lib/types";

export default function PoliciesPage() {
  const router = useRouter();
  const { loggedIn, user, authChecked } = useAuth();
  const [policies, setPolicies] = useState<Policy[]>([]);
  const [loading, setLoading] = useState(true);
  const [togglingId, setTogglingId] = useState<number | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Policy | null>(null);
  const [deleting, setDeleting] = useState(false);

  const fetchPolicies = () => {
    setLoading(true);
    axios.get("/api/policies")
      .then((res) => setPolicies(res.data || []))
      .catch(() => toast.error("Failed to load policies"))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    if (!authChecked) return;
    if (!loggedIn) { router.push("/login"); return; }
    fetchPolicies();
  }, [authChecked, loggedIn, router]);

  const handleToggle = async (p: Policy) => {
    setTogglingId(p.id);
    try {
      await axios.patch(`/api/policies/${p.id}/toggle`);
      setPolicies((prev) => prev.map((pl) => (pl.id === p.id ? { ...pl, isActive: !pl.isActive } : pl)));
      toast.success(`Policy ${p.isActive ? "disabled" : "enabled"}`);
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to toggle policy");
    } finally {
      setTogglingId(null);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);
    try {
      await axios.delete(`/api/policies/${deleteTarget.id}`);
      toast.success("Policy deleted");
      setDeleteTarget(null);
      fetchPolicies();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to delete policy");
    } finally {
      setDeleting(false);
    }
  };

  if (!authChecked || !loggedIn) {
    return (
      <div className="flex items-center justify-center h-64">
        <Skeleton className="h-8 w-48" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <PageHeader crumbs={[
          { label: "Security" },
          { label: "Policies" },
        ]} />
        <Link href="/policies/new">
          <Button size="sm" className="h-8 px-2.5 text-xs gap-1">
            <Plus className="h-3.5 w-3.5" />
            New Policy
          </Button>
        </Link>
      </div>

      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground">Name</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden sm:table-cell">Scope</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden sm:table-cell">Events</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden md:table-cell">Priority</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground">Status</th>
                <th className="w-28 px-4 py-2.5" />
              </tr>
            </thead>
            <tbody className="animate-pulse">
              {loading ? (
                Array.from({ length: 4 }).map((_, i) => (
                  <tr key={i} className="border-b">
                    <td className="px-4 py-2"><Skeleton className="h-3.5 w-32 animate-pulse" /></td>
                    <td className="px-4 py-2 hidden sm:table-cell"><Skeleton className="h-3.5 w-16 animate-pulse" /></td>
                    <td className="px-4 py-2 hidden sm:table-cell"><Skeleton className="h-3.5 w-24 animate-pulse" /></td>
                    <td className="px-4 py-2 hidden md:table-cell"><Skeleton className="h-3.5 w-8 animate-pulse" /></td>
                    <td className="px-4 py-2"><Skeleton className="h-3.5 w-16 animate-pulse" /></td>
                    <td className="px-4 py-2"><Skeleton className="h-3.5 w-20 animate-pulse" /></td>
                  </tr>
                ))
              ) : policies.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center text-muted-foreground">
                    <ShieldAlert className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>No policies configured</p>
                    <p className="text-xs mt-1">Create a policy to automate security responses</p>
                  </td>
                </tr>
              ) : (
                policies.map((p) => (
                  <tr key={p.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-2">
                      <Link href={`/policies/${p.id}`} className="font-medium hover:text-primary transition-colors">
                        {p.name}
                      </Link>
                    </td>
                    <td className="px-4 py-2 text-muted-foreground hidden sm:table-cell capitalize">
                      {p.scopeType}
                      {p.scopeValue && <span className="text-xs ml-1">#{p.scopeValue}</span>}
                    </td>
                    <td className="px-4 py-2 hidden sm:table-cell">
                      <code className="text-[11px] font-mono bg-muted rounded px-1.5 py-0.5">
                        {p.eventTypes || "none"}
                      </code>
                    </td>
                    <td className="px-4 py-2 text-muted-foreground hidden md:table-cell">
                      {p.priority}
                    </td>
                    <td className="px-4 py-2">
                      <span className={`inline-flex items-center gap-1.5 text-xs ${p.isActive ? "text-emerald-500" : "text-muted-foreground"}`}>
                        <span className="h-1.5 w-1.5 rounded-full bg-current" />
                        {p.isActive ? "Active" : "Inactive"}
                      </span>
                    </td>
                    <td className="px-4 py-2">
                      <div className="flex items-center gap-1">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleToggle(p)}
                          disabled={togglingId === p.id}
                          className="h-8 px-2.5 text-xs"
                        >
                          {togglingId === p.id ? "..." : p.isActive ? "Disable" : "Enable"}
                        </Button>
                        <Button
                          variant="destructive-ghost"
                          size="icon"
                          className="h-8 w-8"
                          onClick={() => setDeleteTarget(p)}
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </div>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </Card>

      <Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete policy</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete <strong>{deleteTarget?.name}</strong>? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setDeleteTarget(null)}>Cancel</Button>
            <Button type="button" variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
