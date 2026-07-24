import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card } from "@/components/ui/card";
import { Switch } from "@/components/ui/switch";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { ChevronLeft, ExternalLink, Trash2, Zap } from "lucide-react";
import Link from "next/link";
import axios from "@/lib/api";
import { toast } from "sonner";
import type { Application, IssueTracker, IssueTrackerIssue } from "@/lib/types";
import { severityBadgeColors } from "@/lib/constants";

const severityOptions = ["critical", "high", "medium", "low", "info"];

export default function IssueTrackerPage() {
  const router = useRouter();
  const { id } = router.query;
  const { loggedIn, authChecked } = useAuth();

  const [app, setApp] = useState<Application | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);

  const [config, setConfig] = useState<IssueTracker | null>(null);
  const [form, setForm] = useState({
    provider: "github" as "github" | "gitlab",
    authType: "pat" as "pat" | "github_app",
    token: "",
    githubAppId: "",
    githubInstallationId: "",
    githubAppKey: "",
    severityThreshold: "medium",
    isActive: true,
  });
  const [hasExistingToken, setHasExistingToken] = useState(false);
  const [hasExistingGitHubAppKey, setHasExistingGitHubAppKey] = useState(false);

  const [issues, setIssues] = useState<IssueTrackerIssue[]>([]);
  const [loadingIssues, setLoadingIssues] = useState(true);

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (!authChecked || !loggedIn || !id) return;
    setLoading(true);
    Promise.all([
      axios.get(`/api/applications/${id}`),
      axios.get(`/api/applications/${id}/issue-tracker`).catch(() => null),
    ])
      .then(([appRes, itRes]) => {
        setApp(appRes.data);
        if (itRes?.data) {
          const cfg = itRes.data as IssueTracker;
          setConfig(cfg);
          setForm({
            provider: cfg.provider,
            authType: cfg.authType || "pat",
            token: "",
            githubAppId: cfg.githubAppId ? String(cfg.githubAppId) : "",
            githubInstallationId: cfg.githubInstallationId ? String(cfg.githubInstallationId) : "",
            githubAppKey: "",
            severityThreshold: cfg.severityThreshold,
            isActive: cfg.isActive,
          });
          setHasExistingToken(cfg.hasToken);
          setHasExistingGitHubAppKey(cfg.hasGitHubAppKey || false);
        }
      })
      .catch(() => toast.error("Failed to load configuration"))
      .finally(() => setLoading(false));
  }, [authChecked, loggedIn, id]);

  useEffect(() => {
    if (!authChecked || !loggedIn || !id) return;
    setLoadingIssues(true);
    axios.get(`/api/applications/${id}/issue-tracker/issues`)
      .then((res) => setIssues(res.data || []))
      .catch(() => {})
      .finally(() => setLoadingIssues(false));
  }, [authChecked, loggedIn, id]);

  const handleProviderChange = (provider: "github" | "gitlab") => {
    setForm((f) => ({
      ...f,
      provider,
      authType: provider === "gitlab" ? "pat" : f.authType,
      token: "",
      githubAppId: "",
      githubInstallationId: "",
      githubAppKey: "",
    }));
    setHasExistingToken(false);
    setHasExistingGitHubAppKey(false);
  };

  const handleAuthTypeChange = (authType: "pat" | "github_app") => {
    setForm((f) => ({
      ...f,
      authType,
      token: "",
      githubAppId: "",
      githubInstallationId: "",
      githubAppKey: "",
    }));
    setHasExistingToken(false);
    setHasExistingGitHubAppKey(false);
  };

  const handleTest = async () => {
    setTesting(true);
    try {
      const payload: Record<string, unknown> = { provider: form.provider };
      if (form.authType === "pat") {
        payload.token = form.token || undefined;
      }
      await axios.post(`/api/applications/${id}/issue-tracker/test`, payload);
      toast.success("Connection successful");
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Connection failed");
    } finally {
      setTesting(false);
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      const payload: Record<string, unknown> = {
        provider: form.provider,
        authType: form.authType,
        severityThreshold: form.severityThreshold,
        isActive: form.isActive,
      };

      if (form.authType === "pat") {
        if (form.token) payload.token = form.token;
      } else {
        payload.githubAppId = Number(form.githubAppId);
        payload.githubInstallationId = Number(form.githubInstallationId);
        if (form.githubAppKey) payload.githubAppKey = form.githubAppKey;
      }

      const res = await axios.post(`/api/applications/${id}/issue-tracker`, payload);
      setConfig(res.data);
      setHasExistingToken(form.authType === "pat" ? true : false);
      setHasExistingGitHubAppKey(form.authType === "github_app" ? true : false);
      setForm((f) => ({ ...f, token: "", githubAppKey: "" }));
      toast.success("Configuration saved");
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to save configuration");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    setDeleting(true);
    try {
      await axios.delete(`/api/applications/${id}/issue-tracker`);
      setConfig(null);
      setForm({
        provider: "github",
        authType: "pat",
        token: "",
        githubAppId: "",
        githubInstallationId: "",
        githubAppKey: "",
        severityThreshold: "medium",
        isActive: true,
      });
      setHasExistingToken(false);
      setHasExistingGitHubAppKey(false);
      setIssues([]);
      setDeleteDialogOpen(false);
      toast.success("Configuration deleted");
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to delete configuration");
    } finally {
      setDeleting(false);
    }
  };

  const isFormValid = () => {
    if (!app?.repositoryUrl) return false;
    if (form.authType === "pat") {
      return true; // token can be empty if existing
    }
    return form.githubAppId && form.githubInstallationId;
  };

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
        <Skeleton className="h-8 w-64" />
        <Card><div className="p-6 space-y-4"><Skeleton className="h-6 w-48" /><Skeleton className="h-4 w-full" /><Skeleton className="h-4 w-3/4" /></div></Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => router.push(`/applications/${id}`)} className="h-8 w-8 shrink-0">
          <ChevronLeft className="h-4 w-4" />
        </Button>
        <PageHeader crumbs={[{ label: "Security" }, { label: "Applications", href: "/applications" }, { label: app?.name || "App", href: `/applications/${id}` }, { label: "Issue Tracker" }]} />
      </div>

      <Card>
        <form onSubmit={handleSave}>
          <div className="p-5 space-y-5">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="text-lg font-semibold">Configuration</h3>
                <p className="text-sm text-muted-foreground mt-0.5">Connect a repository to automatically create issues from findings.</p>
              </div>
              {config && (
                <Button type="button" variant="destructive-ghost" size="sm" onClick={() => setDeleteDialogOpen(true)} className="gap-1.5">
                  <Trash2 className="h-3.5 w-3.5" />
                  Delete
                </Button>
              )}
            </div>

            <div className="grid gap-4 max-w-xl">
              <div className="grid gap-2">
                <Label>Provider</Label>
                <Select value={form.provider} onValueChange={(v) => handleProviderChange(v as "github" | "gitlab")}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="github">GitHub</SelectItem>
                    <SelectItem value="gitlab">GitLab</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {form.provider === "github" && (
                <div className="grid gap-2">
                  <Label>Authentication Type</Label>
                  <Select value={form.authType} onValueChange={(v) => handleAuthTypeChange(v as "pat" | "github_app")}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="pat">Personal Access Token</SelectItem>
                      <SelectItem value="github_app">GitHub App</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              )}

              <div className="grid gap-2">
                <Label>Repository</Label>
                <Input
                  value={app?.repositoryUrl || ""}
                  disabled
                  placeholder="Set in application settings"
                />
                {!app?.repositoryUrl && (
                  <p className="text-[11px] text-muted-foreground">No repository URL set. Configure it in the application settings first.</p>
                )}
              </div>

              {form.authType === "pat" ? (
                <div className="grid gap-2">
                  <Label>Personal Access Token</Label>
                  <Input
                    type="password"
                    placeholder={hasExistingToken ? "••••••••••••••" : "ghp_... or glpat-..."}
                    value={form.token}
                    onChange={(e) => setForm((f) => ({ ...f, token: e.target.value }))}
                  />
                  {hasExistingToken && (
                    <p className="text-[11px] text-muted-foreground">Token already set. Leave blank to keep existing.</p>
                  )}
                </div>
              ) : (
                <>
                  <div className="grid gap-2">
                    <Label>GitHub App ID</Label>
                    <Input
                      type="number"
                      placeholder="12345"
                      value={form.githubAppId}
                      onChange={(e) => setForm((f) => ({ ...f, githubAppId: e.target.value }))}
                    />
                  </div>

                  <div className="grid gap-2">
                    <Label>Installation ID</Label>
                    <Input
                      type="number"
                      placeholder="67890"
                      value={form.githubInstallationId}
                      onChange={(e) => setForm((f) => ({ ...f, githubInstallationId: e.target.value }))}
                    />
                  </div>

                  <div className="grid gap-2">
                    <Label>Private Key (PEM)</Label>
                    <textarea
                      className="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 font-mono text-[11px]"
                      placeholder={hasExistingGitHubAppKey ? "••••••••••••••••" : "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"}
                      value={form.githubAppKey}
                      onChange={(e) => setForm((f) => ({ ...f, githubAppKey: e.target.value }))}
                    />
                    {hasExistingGitHubAppKey && (
                      <p className="text-[11px] text-muted-foreground">Private key already set. Leave blank to keep existing.</p>
                    )}
                  </div>
                </>
              )}

              <div className="grid gap-2">
                <Label>Severity threshold</Label>
                <Select value={form.severityThreshold} onValueChange={(v) => setForm((f) => ({ ...f, severityThreshold: v }))}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {severityOptions.map((s) => (
                      <SelectItem key={s} value={s}>{s.charAt(0).toUpperCase() + s.slice(1)}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-[11px] text-muted-foreground">Findings at or above this severity will create issues.</p>
              </div>

              <div className="flex items-center justify-between rounded-lg border p-3">
                <div className="space-y-0.5">
                  <Label>Active</Label>
                  <p className="text-[11px] text-muted-foreground">Enable automatic issue creation for new findings.</p>
                </div>
                <Switch
                  checked={form.isActive}
                  onCheckedChange={(checked) => setForm((f) => ({ ...f, isActive: checked }))}
                />
              </div>
            </div>
          </div>

          <div className="px-5 pb-5 flex items-center gap-2">
            <Button type="button" variant="outline" size="sm" onClick={handleTest} disabled={testing || !app?.repositoryUrl} className="gap-1.5">
              <Zap className="h-3.5 w-3.5" />
              {testing ? "Testing..." : "Test connection"}
            </Button>
            <Button type="submit" size="sm" disabled={saving || !isFormValid()}>
              {saving ? "Saving..." : "Save configuration"}
            </Button>
          </div>
        </form>
      </Card>

      <Card>
        <div className="p-5">
          <h3 className="text-lg font-semibold mb-4">Created issues</h3>
          {loadingIssues ? (
            <div className="space-y-3">
              {Array.from({ length: 3 }).map((_, i) => (
                <Skeleton key={i} className="h-10 w-full" />
              ))}
            </div>
          ) : issues.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <ExternalLink className="h-8 w-8 mx-auto mb-2 opacity-40" />
              <p className="text-sm">No issues created yet</p>
              <p className="text-xs mt-1">Issues will appear here once findings are pushed to your repository.</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-xs">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="text-left px-4 py-2.5 font-medium text-muted-foreground">Finding</th>
                    <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden sm:table-cell">Severity</th>
                    <th className="text-left px-4 py-2.5 font-medium text-muted-foreground">Issue</th>
                    <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden sm:table-cell">Status</th>
                    <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden md:table-cell">Created</th>
                  </tr>
                </thead>
                <tbody>
                  {issues.map((issue) => (
                    <tr key={issue.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-2 font-medium max-w-[200px] truncate">
                        <Link href={`/findings/${issue.findingId}`} className="hover:text-primary transition-colors">
                          {issue.finding?.title || `Finding #${issue.findingId}`}
                        </Link>
                      </td>
                      <td className="px-4 py-2 hidden sm:table-cell">
                        <span className={`inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium border ${severityBadgeColors[issue.finding?.severity] || ""}`}>
                          {issue.finding?.severity || "-"}
                        </span>
                      </td>
                      <td className="px-4 py-2">
                        <a
                          href={issue.externalIssueUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="inline-flex items-center gap-1 text-primary hover:underline"
                          onClick={(e) => e.stopPropagation()}
                        >
                          {issue.externalIssueId}
                          <ExternalLink className="h-3 w-3" />
                        </a>
                      </td>
                      <td className="px-4 py-2 text-muted-foreground hidden sm:table-cell">{issue.status}</td>
                      <td className="px-4 py-2 text-muted-foreground hidden md:table-cell">
                        {new Date(issue.createdAt).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </Card>

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Issue Tracker</DialogTitle>
            <DialogDescription>
              This will disconnect the issue tracker from this application. Existing issues in your repository will not be affected.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting ? "Deleting..." : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
