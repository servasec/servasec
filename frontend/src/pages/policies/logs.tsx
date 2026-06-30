import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { ScrollText } from "lucide-react";
import Link from "next/link";
import axios from "@/lib/api";
import { toast } from "sonner";
import type { PolicyLog } from "@/lib/types";

const EVENT_FILTERS = [
  { value: "", label: "All events" },
  { value: "finding.created", label: "Finding Created" },
  { value: "finding.status_changed", label: "Status Changed" },
  { value: "finding.reassigned", label: "Reassigned" },
];

export default function PolicyLogsPage() {
  const router = useRouter();
  const { loggedIn, authChecked } = useAuth();
  const [logs, setLogs] = useState<PolicyLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [eventFilter, setEventFilter] = useState("");

  const fetchLogs = (eventType: string) => {
    setLoading(true);
    const params = new URLSearchParams();
    if (eventType) params.set("eventType", eventType);

    axios.get(`/api/policies/logs?${params}`)
      .then((res) => setLogs(res.data || []))
      .catch(() => toast.error("Failed to load policy logs"))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    if (!authChecked) return;
    if (!loggedIn) { router.push("/login"); return; }
    fetchLogs(eventFilter);
  }, [authChecked, loggedIn, router]);

  const handleFilterChange = (val: string) => {
    setEventFilter(val);
    fetchLogs(val);
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
      <PageHeader crumbs={[
        { label: "Security" },
        { label: "Policies", href: "/policies" },
        { label: "Logs" },
      ]} />

      <div className="flex items-center gap-3">
        <Select value={eventFilter} onValueChange={handleFilterChange}>
          <SelectTrigger className="w-48 h-8 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {EVENT_FILTERS.map((f) => (
              <SelectItem key={f.value} value={f.value}>{f.label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <Card className="overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">Policy</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Finding</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">Event</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Action</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Result</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden md:table-cell">Date</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array.from({ length: 6 }).map((_, i) => (
                  <tr key={i} className="border-b">
                    <td className="px-4 py-3"><Skeleton className="h-4 w-20" /></td>
                    <td className="px-4 py-3 hidden sm:table-cell"><Skeleton className="h-4 w-16" /></td>
                    <td className="px-4 py-3"><Skeleton className="h-4 w-24" /></td>
                    <td className="px-4 py-3 hidden sm:table-cell"><Skeleton className="h-4 w-16" /></td>
                    <td className="px-4 py-3 hidden sm:table-cell"><Skeleton className="h-4 w-20" /></td>
                    <td className="px-4 py-3 hidden md:table-cell"><Skeleton className="h-4 w-24" /></td>
                  </tr>
                ))
              ) : logs.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center text-muted-foreground">
                    <ScrollText className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>No policy logs found</p>
                    <p className="text-xs mt-1">Logs appear when policies are evaluated against findings</p>
                  </td>
                </tr>
              ) : (
                logs.map((log) => (
                  <tr key={log.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3">
                      <Link href={`/policies/${log.policyId}`} className="hover:text-primary transition-colors">
                        Policy #{log.policyId}
                      </Link>
                    </td>
                    <td className="px-4 py-3 hidden sm:table-cell">
                      <Link href={`/findings?id=${log.findingId}`} className="text-muted-foreground hover:text-foreground transition-colors">
                        #{log.findingId}
                      </Link>
                    </td>
                    <td className="px-4 py-3">
                      <code className="text-xs bg-muted rounded px-1.5 py-0.5">{log.eventType}</code>
                    </td>
                    <td className="px-4 py-3 hidden sm:table-cell capitalize">{log.actionType}</td>
                    <td className="px-4 py-3 hidden sm:table-cell">
                      <span className={`inline-flex items-center gap-1.5 text-sm ${log.actionResult === "success" ? "text-emerald-500" : log.actionResult === "skipped" ? "text-amber-500" : "text-red-500"}`}>
                        <span className="h-1.5 w-1.5 rounded-full bg-current" />
                        {log.actionResult}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground hidden md:table-cell text-xs">
                      {new Date(log.createdAt).toLocaleString()}
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
