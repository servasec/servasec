import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { Separator } from "@/components/ui/separator";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  ChevronDown, Bug, ArrowLeft, ShieldAlert, AlertTriangle, Info,
  FileCode, BookOpen, Hash, ExternalLink, User as UserIcon, Calendar, MessageSquare, CheckCircle2,
} from "lucide-react";
import Link from "next/link";
import axios from "@/lib/api";
import { toast } from "sonner";
import type { Finding, Comment } from "@/lib/types";

const severityConfig: Record<string, { icon: any; color: string; bg: string }> = {
  Critical: { icon: ShieldAlert, color: "text-red-600 dark:text-red-400", bg: "bg-red-500/10" },
  High: { icon: AlertTriangle, color: "text-orange-600 dark:text-orange-400", bg: "bg-orange-500/10" },
  Medium: { icon: AlertTriangle, color: "text-yellow-600 dark:text-yellow-400", bg: "bg-yellow-500/10" },
  Low: { icon: Info, color: "text-green-600 dark:text-green-400", bg: "bg-green-500/10" },
  Info: { icon: Info, color: "text-blue-600 dark:text-blue-400", bg: "bg-blue-500/10" },
};

const statusLabels: Record<string, string> = {
  open: "Open",
  confirmed: "Confirmed",
  false_positive: "False Positive",
  fixed: "Fixed",
};

const statusColors: Record<string, string> = {
  open: "text-red-500",
  confirmed: "text-amber-500",
  false_positive: "text-emerald-500",
  fixed: "text-blue-500",
};

const nextStatuses: Record<string, string[]> = {
  open: ["confirmed", "false_positive"],
  confirmed: ["fixed", "false_positive"],
  false_positive: ["open"],
  fixed: ["open"],
};

export default function FindingDetailPage() {
  const router = useRouter();
  const { id } = router.query;
  const { loggedIn, user, authChecked } = useAuth();
  const [finding, setFinding] = useState<Finding | null>(null);
  const [loading, setLoading] = useState(true);

  const [assignUserId, setAssignUserId] = useState("");
  const [assignDueDate, setAssignDueDate] = useState("");
  const [assigning, setAssigning] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<{ id: number; username: string }[]>([]);
  const [searching, setSearching] = useState(false);

  const [comments, setComments] = useState<Comment[]>([]);
  const [commentBody, setCommentBody] = useState("");
  const [postingComment, setPostingComment] = useState(false);

  const fetchFinding = () => {
    if (!id) return;
    axios.get(`/api/findings/${id}`)
      .then((res) => {
        setFinding(res.data);
        setAssignUserId(String(res.data.assignedTo || ""));
        setAssignDueDate(res.data.dueDate ? res.data.dueDate.split("T")[0] : "");
      })
      .catch(() => {
        toast.error("Finding not found");
        router.replace("/findings");
      })
      .finally(() => setLoading(false));
  };

  const fetchComments = () => {
    if (!id) return;
    axios.get(`/api/findings/${id}/comments`)
      .then((res) => setComments(res.data))
      .catch(() => {});
  };

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn && id) {
      fetchFinding();
      fetchComments();
    }
  }, [authChecked, loggedIn, id]);

  useEffect(() => {
    if (searchQuery.length < 2) {
      setSearchResults([]);
      return;
    }
    const timer = setTimeout(() => {
      setSearching(true);
      axios.get(`/api/users/search?q=${encodeURIComponent(searchQuery)}`)
        .then((res) => setSearchResults(res.data || []))
        .catch(() => setSearchResults([]))
        .finally(() => setSearching(false));
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const handleStatusChange = async (newStatus: string) => {
    if (!finding) return;
    try {
      await axios.patch(`/api/findings/${finding.id}/status`, { status: newStatus });
      toast.success(`Status updated to ${statusLabels[newStatus] || newStatus}`);
      setFinding({ ...finding, status: newStatus });
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to update status");
    }
  };

  const handleAssign = async () => {
    if (!finding) return;
    setAssigning(true);
    try {
      const payload: any = {};
      payload.userId = assignUserId ? Number(assignUserId) : null;
      if (assignDueDate) payload.dueDate = assignDueDate;
      await axios.patch(`/api/findings/${finding.id}/assign`, payload);
      toast.success(assignUserId ? "Finding assigned" : "Finding unassigned");
      fetchFinding();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to assign");
    } finally {
      setAssigning(false);
    }
  };

  const handleReview = async () => {
    if (!finding) return;
    try {
      await axios.patch(`/api/findings/${finding.id}/review`, {});
      toast.success("Finding reviewed");
      fetchFinding();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to review");
    }
  };

  const handlePostComment = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!finding || !commentBody.trim()) return;
    setPostingComment(true);
    try {
      await axios.post(`/api/findings/${finding.id}/comments`, { body: commentBody });
      setCommentBody("");
      toast.success("Comment added");
      fetchComments();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to post comment");
    } finally {
      setPostingComment(false);
    }
  };

  const cweNumber = finding?.cweId?.match(/CWE-(\d+)/)?.[1];

  if (!authChecked || !loggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-muted-foreground">Loading...</div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="space-y-6">
        <PageHeader crumbs={[{ label: "Findings", href: "/findings" }, { label: "..." }]} />
        <Card><CardContent className="p-8 space-y-4"><Skeleton className="h-8 w-64" /><Skeleton className="h-4 w-96" /></CardContent></Card>
      </div>
    );
  }

  if (!finding) return null;

  const sevConf = severityConfig[finding.severity] || severityConfig.Info;
  const SevIcon = sevConf.icon;
  const isOverdue = finding.dueDate && finding.status !== "fixed" && finding.status !== "false_positive" && new Date(finding.dueDate) < new Date();

  return (
    <div className="space-y-6">
      <PageHeader crumbs={[{ label: "Findings", href: "/findings" }, { label: finding.title }]} />

      <div className="flex items-start gap-4">
        <Link href="/findings">
          <Button variant="ghost" size="icon" className="h-8 w-8 mt-1 shrink-0">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-3 flex-wrap">
            <h1 className="text-xl font-bold break-words">{finding.title}</h1>
            <span className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs font-medium ${sevConf.bg} ${sevConf.color}`}>
              <SevIcon className="h-3.5 w-3.5" />
              {finding.severity}
            </span>
          </div>
          <p className="text-sm text-muted-foreground mt-1">
            {finding.scannerType?.name || "Unknown scanner"}
            {finding.cweId && <span className="mx-1.5">·</span>}
            {finding.cweId && <span>{finding.cweId}</span>}
            <span className="mx-1.5">·</span>
            Created {new Date(finding.createdAt).toLocaleDateString()}
          </p>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        <div className="lg:col-span-2 space-y-6">
          {finding.description && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <Bug className="h-4 w-4" />
                  Description
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm leading-relaxed whitespace-pre-wrap">{finding.description}</p>
              </CardContent>
            </Card>
          )}

          {finding.remediation && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <BookOpen className="h-4 w-4" />
                  Remediation
                </CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-sm leading-relaxed whitespace-pre-wrap">{finding.remediation}</p>
              </CardContent>
            </Card>
          )}

          {(finding.filePath || finding.lineStart || finding.lineEnd) && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base flex items-center gap-2">
                  <FileCode className="h-4 w-4" />
                  Location
                </CardTitle>
              </CardHeader>
              <CardContent>
                {finding.filePath && (
                  <div className="flex items-center gap-2 text-sm">
                    <code className="rounded bg-muted px-2 py-1 text-xs flex-1 break-all">{finding.filePath}</code>
                    {(finding.lineStart || finding.lineEnd) && (
                      <span className="text-muted-foreground whitespace-nowrap text-xs">
                        Lines {finding.lineStart || "?"}{finding.lineEnd && finding.lineEnd !== finding.lineStart ? `-${finding.lineEnd}` : ""}
                      </span>
                    )}
                  </div>
                )}
              </CardContent>
            </Card>
          )}

          <Card>
            <CardHeader>
              <CardTitle className="text-base flex items-center gap-2">
                <MessageSquare className="h-4 w-4" />
                Comments
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <form onSubmit={handlePostComment} className="space-y-2">
                <Textarea
                  placeholder="Add a comment..."
                  value={commentBody}
                  onChange={(e) => setCommentBody(e.target.value)}
                  rows={3}
                />
                <div className="flex justify-end">
                  <Button type="submit" size="sm" disabled={postingComment || !commentBody.trim()}>
                    {postingComment ? "Posting..." : "Add comment"}
                  </Button>
                </div>
              </form>
              <Separator />
              {comments.length === 0 ? (
                <p className="text-sm text-muted-foreground">No comments yet.</p>
              ) : (
                <div className="space-y-4 max-h-80 overflow-y-auto">
                  {comments.map((c) => (
                    <div key={c.id} className="flex gap-3">
                      <div className="h-7 w-7 rounded-full bg-primary/10 flex items-center justify-center shrink-0 mt-0.5">
                        <span className="text-xs font-medium text-primary">{c.user?.username?.charAt(0).toUpperCase() || "?"}</span>
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-medium">{c.user?.username || "Unknown"}</span>
                          <span className="text-xs text-muted-foreground">{new Date(c.createdAt).toLocaleString()}</span>
                        </div>
                        <p className="text-sm text-muted-foreground mt-0.5 whitespace-pre-wrap">{c.body}</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>

        <div className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Status</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" className={`w-full justify-between gap-2 ${statusColors[finding.status] || ""}`}>
                    <span className="inline-flex items-center gap-1.5">
                      <span className="h-1.5 w-1.5 rounded-full bg-current" />
                      {statusLabels[finding.status] || finding.status}
                    </span>
                    <ChevronDown className="h-4 w-4 opacity-50" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="w-48">
                  {nextStatuses[finding.status]?.map((ns) => (
                    <DropdownMenuItem key={ns} onClick={() => handleStatusChange(ns)}>
                      <span className={`h-1.5 w-1.5 rounded-full bg-current mr-2 ${statusColors[ns] || ""}`} />
                      {statusLabels[ns] || ns}
                    </DropdownMenuItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>

              {finding.status === "fixed" && finding.fixedAt && (
                <p className="text-xs text-muted-foreground text-center">
                  Fixed {new Date(finding.fixedAt).toLocaleDateString()}
                </p>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Assignee</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              {finding.assignedToUser && (
                <div className="flex items-center gap-2 text-sm text-muted-foreground pb-2 border-b">
                  <div className="h-6 w-6 rounded-full bg-primary/10 flex items-center justify-center">
                    <span className="text-[10px] font-medium text-primary">{finding.assignedToUser.username.charAt(0).toUpperCase()}</span>
                  </div>
                  <span className="font-medium text-foreground">{finding.assignedToUser.username}</span>
                  {isOverdue && <span className="text-[10px] font-medium text-red-500 ml-auto">OVERDUE</span>}
                </div>
              )}
              <div className="space-y-2">
                <Label>Assign to</Label>
                <Input
                  value={searchQuery}
                  onChange={(e) => { setSearchQuery(e.target.value); setAssignUserId(""); }}
                  placeholder="Search users (min 2 chars)..."
                  autoComplete="off"
                  data-1p-ignore
                />
                {searching && (
                  <p className="text-xs text-muted-foreground">Searching...</p>
                )}
                {searchResults.length > 0 && !assignUserId && (
                  <div className="border rounded-md divide-y max-h-40 overflow-y-auto">
                    {searchResults.map((u) => (
                      <button
                        key={u.id}
                        type="button"
                        className="w-full text-left px-3 py-2 text-sm hover:bg-accent transition-colors"
                        onClick={() => { setAssignUserId(String(u.id)); setSearchResults([]); setSearchQuery(u.username); }}
                      >
                        {u.username} <span className="text-muted-foreground">(id: {u.id})</span>
                      </button>
                    ))}
                  </div>
                )}
                {searchQuery.length >= 2 && !searching && searchResults.length === 0 && !assignUserId && (
                  <p className="text-xs text-muted-foreground">No users found</p>
                )}
                {assignUserId && (
                  <p className="text-sm text-muted-foreground">
                    Selected: <span className="font-medium text-foreground">{searchQuery}</span>
                    {" "}<button type="button" className="text-xs text-primary hover:underline" onClick={() => { setAssignUserId(""); setSearchQuery(""); }}>Change</button>
                  </p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="assignDueDate">Due date</Label>
                <Input id="assignDueDate" type="date" value={assignDueDate} onChange={(e) => setAssignDueDate(e.target.value)} />
              </div>
              <Button onClick={handleAssign} disabled={assigning} className="w-full" size="sm">
                {assigning ? "Assigning..." : "Assign"}
              </Button>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Review</CardTitle>
            </CardHeader>
            <CardContent>
              {finding.reviewedByUser ? (
                <div className="flex items-center gap-2 text-sm text-muted-foreground">
                  <CheckCircle2 className="h-4 w-4 text-emerald-500" />
                  <span>Reviewed by <span className="font-medium text-foreground">{finding.reviewedByUser.username}</span></span>
                </div>
              ) : (
                <Button onClick={handleReview} variant="outline" className="w-full gap-2">
                  <CheckCircle2 className="h-4 w-4" />
                  Mark as reviewed
                </Button>
              )}
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Details</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Rule ID</span>
                <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{finding.ruleId || "-"}</code>
              </div>
              <Separator />
              <div className="flex justify-between">
                <span className="text-muted-foreground">Scanner</span>
                <span>{finding.scannerType?.name || "-"}</span>
              </div>
              <Separator />
              <div className="flex justify-between">
                <span className="text-muted-foreground">Version</span>
                <span>{finding.applicationVersion?.name || `#${finding.applicationVersionId}`}</span>
              </div>
              <Separator />
              <div className="flex justify-between">
                <span className="text-muted-foreground">Scan ID</span>
                <span>#{finding.scanId}</span>
              </div>
              <Separator />
              {finding.cweId && cweNumber && (
                <>
                  <div className="flex justify-between items-center">
                    <span className="text-muted-foreground">CWE</span>
                    <a
                      href={`https://cwe.mitre.org/data/definitions/${cweNumber}.html`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="inline-flex items-center gap-1 text-primary hover:underline"
                    >
                      {cweNumber}
                      <ExternalLink className="h-3 w-3" />
                    </a>
                  </div>
                  <Separator />
                </>
              )}
              <div className="flex justify-between">
                <span className="text-muted-foreground">Created</span>
                <span>{new Date(finding.createdAt).toLocaleDateString()}</span>
              </div>
              <Separator />
              <div className="flex justify-between">
                <span className="text-muted-foreground">Updated</span>
                <span>{new Date(finding.updatedAt).toLocaleDateString()}</span>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
