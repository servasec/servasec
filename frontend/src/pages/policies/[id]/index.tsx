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
import { Pencil, Trash2, ChevronLeft, ShieldAlert, ScrollText } from "lucide-react";
import Link from "next/link";
import axios from "@/lib/api";
import { toast } from "sonner";
import type { Policy, PolicyLog, PolicyCondition, PolicyAction } from "@/lib/types";

export default function PolicyDetailPage() {
  const router = useRouter();
  const { id } = router.query;
  const { loggedIn, authChecked } = useAuth();
  const [policy, setPolicy] = useState<Policy | null>(null);
  const [logs, setLogs] = useState<PolicyLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleteTarget, setDeleteTarget] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const fetchData = () => {
    if (!id) return;
    Promise.all([
      axios.get(`/api/policies/${id}`),
      axios.get(`/api/policies/${id}/logs`),
    ])
      .then(([policyRes, logsRes]) => {
        setPolicy(policyRes.data);
        setLogs(logsRes.data || []);
      })
      .catch(() => { toast.error("Failed to load policy"); router.push("/policies"); })
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    if (!authChecked) return;
    if (!loggedIn) { router.push("/login"); return; }
    if (!id) return;
    fetchData();
  }, [authChecked, loggedIn, id, router]);

  const handleDelete = async () => {
    if (!policy) return;
    setDeleting(true);
    try {
      await axios.delete(`/api/policies/${policy.id}`);
      toast.success("Policy deleted");
      router.push("/policies");
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to delete policy");
      setDeleting(false);
    }
  };

  const parseJSON = (str: string) => {
    try { return JSON.parse(str); } catch { return null; }
  };

  if (!authChecked || !loggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-muted-foreground">Loading...</div>
      </div>
    );
  }

  if (loading || !policy) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-7 w-48 animate-pulse" />
        <Card><div className="p-4 space-y-3"><Skeleton className="h-5 w-64 animate-pulse" /><Skeleton className="h-3.5 w-full animate-pulse" /><Skeleton className="h-3.5 w-3/4 animate-pulse" /></div></Card>
      </div>
    );
  }

  const conditions: PolicyCondition[] | null = parseJSON(policy.conditions);
  const actions: PolicyAction[] | null = parseJSON(policy.actions);

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => router.push("/policies")} className="h-8 w-8 shrink-0">
          <ChevronLeft className="h-4 w-4" />
        </Button>
        <PageHeader crumbs={[
          { label: "Security" },
          { label: "Policies", href: "/policies" },
          { label: policy.name },
        ]} />
      </div>

      <div className="flex items-center gap-2">
        <Link href={`/policies/${policy.id}/edit`}>
          <Button variant="outline" size="sm" className="h-8 px-2.5 text-[11px] gap-1">
            <Pencil className="h-3 w-3" />
            Edit
          </Button>
        </Link>
        <Button variant="destructive-ghost" size="sm" className="h-8 px-2.5 text-[11px] gap-1" onClick={() => setDeleteTarget(true)}>
          <Trash2 className="h-3 w-3" />
          Delete
        </Button>
      </div>

      <Card>
        <div className="p-4 space-y-4">
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            <div>
              <p className="text-xs text-muted-foreground mb-0.5">Status</p>
              <span className={`inline-flex items-center gap-1.5 text-xs ${policy.isActive ? "text-emerald-500" : "text-muted-foreground"}`}>
                <span className="h-1.5 w-1.5 rounded-full bg-current" />
                {policy.isActive ? "Active" : "Inactive"}
              </span>
            </div>
            <div>
              <p className="text-xs text-muted-foreground mb-0.5">Scope</p>
              <p className="text-xs capitalize">
                {policy.scopeType}
                {policy.scopeValue && <span className="text-muted-foreground ml-1">(#{policy.scopeValue})</span>}
              </p>
            </div>
            <div>
              <p className="text-xs text-muted-foreground mb-0.5">Priority</p>
              <p className="text-xs">{policy.priority}</p>
            </div>
          </div>

          {policy.description && (
            <div>
              <p className="text-xs text-muted-foreground mb-0.5">Description</p>
              <p className="text-xs">{policy.description}</p>
            </div>
          )}

          <div>
            <p className="text-xs text-muted-foreground mb-0.5">Event Types</p>
            <div className="flex flex-wrap gap-1.5">
              {(policy.eventTypes || "").split(",").filter(Boolean).map((ev) => (
                <code key={ev} className="text-[11px] font-mono bg-muted rounded px-1.5 py-0.5">{ev.trim()}</code>
              ))}
              {!policy.eventTypes && <p className="text-xs text-muted-foreground">None</p>}
            </div>
          </div>

          <div>
            <p className="text-xs text-muted-foreground mb-1">Conditions</p>
            {Array.isArray(conditions) && conditions.length > 0 ? (
              <div className="space-y-1">
                {conditions.map((c, i) => (
                  <div key={i} className="flex items-center gap-2 text-xs">
                    <code className="text-[11px] font-mono bg-muted rounded px-1.5 py-0.5">{c.field}</code>
                    <span className="text-muted-foreground">{c.op}</span>
                    <code className="text-[11px] font-mono bg-muted rounded px-1.5 py-0.5">{c.value}</code>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-xs text-muted-foreground">No conditions (always matches)</p>
            )}
          </div>

          <div>
            <p className="text-xs text-muted-foreground mb-1">Actions</p>
            {Array.isArray(actions) && actions.length > 0 ? (
              <div className="space-y-1">
                {actions.map((a, i) => (
                  <div key={i} className="flex items-center gap-2 text-xs">
                    <code className="text-[11px] font-mono bg-muted rounded px-1.5 py-0.5">{a.type}</code>
                    {a.target && <span className="text-muted-foreground">→ <span className="text-foreground">{a.target}</span></span>}
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-xs text-muted-foreground">No actions</p>
            )}
          </div>

          <div className="grid grid-cols-2 gap-3 text-xs text-muted-foreground">
            <div>
              <p>Created: {new Date(policy.createdAt).toLocaleString()}</p>
            </div>
            <div>
              <p>Updated: {new Date(policy.updatedAt).toLocaleString()}</p>
            </div>
          </div>
        </div>
      </Card>

      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold">Execution Logs</h3>
      </div>

      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground">Finding</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden sm:table-cell">Event</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground">Action</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden sm:table-cell">Result</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden sm:table-cell">Detail</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden md:table-cell">Date</th>
              </tr>
            </thead>
            <tbody className="animate-pulse">
              {logs.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center text-muted-foreground">
                    <ScrollText className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>No execution logs yet</p>
                  </td>
                </tr>
              ) : (
                logs.map((log) => (
                  <tr key={log.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-2">
                      <Link href={`/findings?id=${log.findingId}`} className="hover:text-primary transition-colors">
                        #{log.findingId}
                      </Link>
                    </td>
                    <td className="px-4 py-2 hidden sm:table-cell">
                      <code className="text-[11px] font-mono bg-muted rounded px-1.5 py-0.5">{log.eventType}</code>
                    </td>
                    <td className="px-4 py-2 capitalize">{log.actionType}</td>
                    <td className="px-4 py-2 hidden sm:table-cell">
                      <span className={`inline-flex items-center gap-1.5 text-xs ${log.actionResult === "success" ? "text-emerald-500" : log.actionResult === "skipped" ? "text-amber-500" : "text-red-500"}`}>
                        <span className="h-1.5 w-1.5 rounded-full bg-current" />
                        {log.actionResult}
                      </span>
                    </td>
                    <td className="px-4 py-2 hidden sm:table-cell max-w-[200px] truncate text-muted-foreground">
                      {log.detail || "-"}
                    </td>
                    <td className="px-4 py-2 text-muted-foreground hidden md:table-cell text-xs">
                      {new Date(log.createdAt).toLocaleString()}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </Card>

      <Dialog open={deleteTarget} onOpenChange={setDeleteTarget}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete policy</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete <strong>{policy.name}</strong>? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setDeleteTarget(false)}>Cancel</Button>
            <Button type="button" variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
