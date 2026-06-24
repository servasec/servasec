import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  ChevronLeft, Plus, Pencil, Trash2, Upload, Bug, Star, StarOff, Globe, GitCompare, Shield, Scan as ScanIcon, Calendar,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import axios from "@/lib/api";
import { toast } from "sonner";
import type { Webhook, ScanItem } from "@/lib/types";
import { statusScanColors } from "@/lib/constants";
import { UserSearch } from "@/components/user-search";
import { SeverityChart } from "@/components/findings/severity-chart";

interface ApplicationVersion {
  id: number;
  applicationId: number;
  name: string;
  branch: string;
  tag: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

interface Application {
  id: number;
  name: string;
  description: string;
  slug: string;
  groupId: number;
  repositoryUrl: string;
  versions: ApplicationVersion[];
  defaultVersion: ApplicationVersion | null;
  createdAt: string;
  updatedAt: string;
}

interface Group {
  id: number;
  name: string;
}

interface ScannerType {
  id: number;
  name: string;
  description: string;
  parser: string;
}

interface AppPermission {
  id: number;
  subject: string;
  resource: string;
  action: string;
}

export default function ApplicationDetailPage() {
  const router = useRouter();
  const { id } = router.query;
  const { loggedIn, authChecked } = useAuth();
  const [app, setApp] = useState<Application | null>(null);
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);

  const [versionDialogOpen, setVersionDialogOpen] = useState(false);
  const [editingVersion, setEditingVersion] = useState<ApplicationVersion | null>(null);
  const [versionForm, setVersionForm] = useState({ name: "", branch: "", tag: "", isDefault: false });
  const [savingVersion, setSavingVersion] = useState(false);
  const [deleteVersionTarget, setDeleteVersionTarget] = useState<ApplicationVersion | null>(null);

  const [ingestDialogOpen, setIngestDialogOpen] = useState(false);
  const [scannerTypes, setScannerTypes] = useState<ScannerType[]>([]);
  const [ingestFile, setIngestFile] = useState<File | null>(null);
  const [ingestScannerType, setIngestScannerType] = useState("");
  const [ingestVersionName, setIngestVersionName] = useState("");
  const [ingestVersionMode, setIngestVersionMode] = useState<"select" | "new">("select");
  const [ingestBranch, setIngestBranch] = useState("");
  const [ingesting, setIngesting] = useState(false);

  const [webhooks, setWebhooks] = useState<Webhook[]>([]);
  const [webhookDialogOpen, setWebhookDialogOpen] = useState(false);
  const [webhookForm, setWebhookForm] = useState({ url: "", events: "", secret: "", isActive: true });
  const [savingWebhook, setSavingWebhook] = useState(false);

  const [appPerms, setAppPerms] = useState<AppPermission[]>([]);
  const [canManagePerms, setCanManagePerms] = useState(false);
  const [permDialogOpen, setPermDialogOpen] = useState(false);
  const [permForm, setPermForm] = useState({ subjectType: "user", subjectValue: "", action: "read" });
  const [savingPerm, setSavingPerm] = useState(false);
  const [permTeams, setPermTeams] = useState<{ id: number; name: string }[]>([]);

  const [activeTab, setActiveTab] = useState<"overview" | "versions" | "settings">("overview");
  const [versionFindings, setVersionFindings] = useState<{ severity: string }[]>([]);
  const [loadingFindings, setLoadingFindings] = useState(false);
  const [latestScan, setLatestScan] = useState<ScanItem | null>(null);
  const [loadingScan, setLoadingScan] = useState(false);

  const groupMap = Object.fromEntries(groups.map((g) => [g.id, g.name]));

  const fetchApp = () => {
    if (!id) return;
    Promise.all([
      axios.get(`/api/applications/${id}`),
      axios.get("/api/groups"),
      axios.get(`/api/applications/${id}/webhooks`),
    ])
      .then(([appRes, groupsRes, whRes]) => {
        setApp(appRes.data);
        setGroups(groupsRes.data);
        setWebhooks(whRes.data);
      })
      .catch(() => toast.error("Failed to load application"))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn && id) {
      fetchApp();
    }
  }, [authChecked, loggedIn, id]);

  useEffect(() => {
    if (!id || !app?.defaultVersion?.id) return;
    setLoadingFindings(true);
    axios.get(`/api/applications/${id}/versions/${app.defaultVersion.id}/findings`)
      .then((res) => setVersionFindings(res.data || []))
      .catch(() => setVersionFindings([]))
      .finally(() => setLoadingFindings(false));
  }, [id, app?.defaultVersion?.id]);

  useEffect(() => {
    if (!id) return;
    setLoadingScan(true);
    axios.get(`/api/scans?applicationId=${id}`)
      .then((res) => {
        const scans = res.data || [];
        setLatestScan(scans.length > 0 ? scans[0] : null);
      })
      .catch(() => setLatestScan(null))
      .finally(() => setLoadingScan(false));
  }, [id]);

  const openCreateVersion = () => {
    setEditingVersion(null);
    setVersionForm({ name: "", branch: "", tag: "", isDefault: false });
    setVersionDialogOpen(true);
  };

  const openEditVersion = (v: ApplicationVersion) => {
    setEditingVersion(v);
    setVersionForm({ name: v.name, branch: v.branch || "", tag: v.tag || "", isDefault: v.isDefault });
    setVersionDialogOpen(true);
  };

  const handleSaveVersion = async (e: React.FormEvent) => {
    e.preventDefault();
    setSavingVersion(true);
    try {
      if (editingVersion) {
        await axios.patch(`/api/applications/${id}/versions/${editingVersion.id}`, versionForm);
        toast.success("Version updated");
      } else {
        await axios.post(`/api/applications/${id}/versions`, versionForm);
        toast.success("Version created");
      }
      setVersionDialogOpen(false);
      fetchApp();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to save version");
    } finally {
      setSavingVersion(false);
    }
  };

  const handleDeleteVersion = async () => {
    if (!deleteVersionTarget) return;
    try {
      await axios.delete(`/api/applications/${id}/versions/${deleteVersionTarget.id}`);
      toast.success("Version deleted");
      setDeleteVersionTarget(null);
      fetchApp();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to delete version");
    }
  };

  const handleSetDefault = async (v: ApplicationVersion) => {
    try {
      await axios.patch(`/api/applications/${id}/versions/${v.id}`, { isDefault: true });
      toast.success(`"${v.name}" is now the default version`);
      fetchApp();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to set default version");
    }
  };

  const openIngest = () => {
    setIngestFile(null);
    setIngestScannerType("");
    setIngestVersionName(app?.defaultVersion?.name || "");
    setIngestVersionMode("select");
    setIngestBranch("");
    axios.get("/api/scanner-types")
      .then((res) => setScannerTypes(res.data))
      .catch(() => toast.error("Failed to load scanner types"));
    setIngestDialogOpen(true);
  };

  const handleIngest = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!ingestFile) {
      toast.error("Please select a file to upload");
      return;
    }
    setIngesting(true);
    const formData = new FormData();
    formData.append("file", ingestFile);
    if (ingestScannerType) {
      const st = scannerTypes.find((s) => String(s.id) === ingestScannerType);
      if (st) formData.append("scannerType", st.name);
    }
    formData.append("version", ingestVersionName);
    if (ingestBranch) formData.append("branch", ingestBranch);

    try {
      const res = await axios.post(`/api/applications/${id}/ingest`, formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      toast.success(`Scan completed - ${res.data.findingsCount} findings`);
      setIngestDialogOpen(false);
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to upload scan");
    } finally {
      setIngesting(false);
    }
  };

  const openCreateWebhook = () => {
    setWebhookForm({ url: "", events: "", secret: "", isActive: true });
    setWebhookDialogOpen(true);
  };

  const handleSaveWebhook = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!id) return;
    setSavingWebhook(true);
    try {
      await axios.post(`/api/applications/${id}/webhooks`, webhookForm);
      toast.success("Webhook created");
      setWebhookDialogOpen(false);
      fetchApp();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to create webhook");
    } finally {
      setSavingWebhook(false);
    }
  };

  const handleDeleteWebhook = async (wh: Webhook) => {
    if (!id) return;
    try {
      await axios.delete(`/api/applications/${id}/webhooks/${wh.id}`);
      toast.success("Webhook deleted");
      fetchApp();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to delete webhook");
    }
  };

  const fetchAppPermissions = async () => {
    if (!id) return;
    try {
      const res = await axios.get(`/api/applications/${id}/permissions`);
      setAppPerms(res.data || []);
      setCanManagePerms(true);
    } catch {
      setCanManagePerms(false);
    }
  };

  useEffect(() => {
    if (authChecked && loggedIn && id) {
      fetchAppPermissions();
    }
  }, [authChecked, loggedIn, id]);

  useEffect(() => {
    if (!permDialogOpen) {
      setPermForm({ subjectType: "user", subjectValue: "", action: "read" });
      setPermTeams([]);
    }
  }, [permDialogOpen]);

  const openPermGrant = () => {
    if (permForm.subjectType === "team") {
      axios.get("/api/teams")
        .then((res) => setPermTeams(res.data || []))
        .catch(() => {});
    }
    setPermDialogOpen(true);
  };

  const handleGrantPerm = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!permForm.subjectValue) return;
    setSavingPerm(true);
    try {
      await axios.post(`/api/applications/${id}/permissions`, {
        subject: `${permForm.subjectType}:${permForm.subjectValue}`,
        action: permForm.action,
      });
      toast.success("Permission granted");
      setPermDialogOpen(false);
      fetchAppPermissions();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to grant permission");
    } finally {
      setSavingPerm(false);
    }
  };

  const handleRevokePerm = async (p: AppPermission) => {
    try {
      await axios.delete(`/api/applications/${id}/permissions`, {
        data: { subject: p.subject, action: p.action },
      });
      toast.success("Permission revoked");
      fetchAppPermissions();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to revoke permission");
    }
  };

  if (!authChecked || !loggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-muted-foreground">Loading...</div>
      </div>
    );
  }

  if (loading || !app) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-48" />
        <Card><div className="p-6 space-y-4"><Skeleton className="h-6 w-64" /><Skeleton className="h-4 w-full" /><Skeleton className="h-4 w-3/4" /></div></Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Button variant="ghost" size="icon" onClick={() => router.push("/applications")} className="h-8 w-8 shrink-0">
          <ChevronLeft className="h-4 w-4" />
        </Button>
        <PageHeader crumbs={[{ label: "Security" }, { label: "Applications", href: "/applications" }, { label: app.name }]} />
      </div>

      <div className="border-b border-border">
        <div className="flex gap-0 -mb-px">
          {(["overview", "versions", "settings"] as const).map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`px-4 py-2.5 text-sm font-medium border-b-2 transition-colors
                ${activeTab === tab
                  ? "border-primary text-foreground"
                  : "border-transparent text-muted-foreground hover:text-foreground"
                }`}
            >
              {tab === "settings" ? "Settings" : tab.charAt(0).toUpperCase() + tab.slice(1)}
            </button>
          ))}
        </div>
      </div>

      {activeTab === "overview" && (
        <div key="overview" className="animate-in fade-in duration-200">
          <Card>
            <div className="p-6">
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  <h2 className="text-xl font-semibold truncate">{app.name}</h2>
                  {app.description && <p className="text-muted-foreground mt-1">{app.description}</p>}
                </div>
                <div className="flex items-center gap-2 shrink-0">
                  <Button variant="outline" size="sm" onClick={() => router.push(`/findings?applicationId=${app.id}`)}>
                    <Bug className="h-4 w-4 mr-1.5" />
                    View findings
                  </Button>
                  <Button size="sm" onClick={openIngest}>
                    <Upload className="h-4 w-4 mr-1.5" />
                    Upload scan
                  </Button>
                </div>
              </div>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-6">
                <div>
                  <p className="text-xs text-muted-foreground mb-0.5">Slug</p>
                  <code className="text-sm bg-muted px-1.5 py-0.5 rounded">{app.slug}</code>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground mb-0.5">Group</p>
                  <p className="text-sm">{groupMap[app.groupId] || `#${app.groupId}`}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground mb-0.5">Repository</p>
                  <p className="text-sm truncate">{app.repositoryUrl || "-"}</p>
                </div>
                <div>
                  <p className="text-xs text-muted-foreground mb-0.5">Default Version</p>
                  <p className="text-sm">{app.defaultVersion?.name || "-"}</p>
                </div>
              </div>
            </div>
          </Card>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold mb-4">Latest scan</h3>
                {loadingScan ? (
                  <div className="space-y-3">
                    <Skeleton className="h-5 w-32" />
                    <Skeleton className="h-5 w-48" />
                    <Skeleton className="h-5 w-24" />
                  </div>
                ) : latestScan ? (
                  <div className="space-y-3">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Scanner</span>
                      <span className="font-medium">{latestScan.scannerType?.name || "-"}</span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Status</span>
                      <span className={`inline-flex items-center gap-1.5 text-sm ${statusScanColors[latestScan.status] || ""}`}>
                        <span className="h-1.5 w-1.5 rounded-full bg-current" />
                        {latestScan.status}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Version</span>
                      <span className="font-medium">{latestScan.applicationVersion?.name || `#${latestScan.applicationVersionId}`}</span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Findings</span>
                      <span className="font-medium">{latestScan.findingsCount}</span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">Date</span>
                      <span className="text-muted-foreground">{new Date(latestScan.createdAt).toLocaleDateString()}</span>
                    </div>
                  </div>
                ) : (
                  <div className="text-center py-6 text-muted-foreground">
                    <ScanIcon className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p className="text-sm">No scans yet</p>
                    <p className="text-xs mt-1">Upload a scan to get started</p>
                  </div>
                )}
              </div>
            </Card>

            <Card>
              <div className="p-6">
                <h3 className="text-lg font-semibold mb-4">Findings severity</h3>
                <SeverityChart findings={versionFindings} loading={loadingFindings} />
              </div>
            </Card>
          </div>
        </div>
      )}

      {activeTab === "versions" && (
        <div key="versions" className="animate-in fade-in duration-200">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold">Versions</h3>
            <div className="flex items-center gap-2">
              <Button variant="outline" size="sm" onClick={() => router.push(`/applications/${id}/compare`)} className="gap-1.5">
                <GitCompare className="h-4 w-4" />
                Compare
              </Button>
              <Button size="sm" onClick={openCreateVersion} className="gap-1.5">
                <Plus className="h-4 w-4" />
                New version
              </Button>
            </div>
          </div>

          <Card>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">Name</th>
                    <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Branch</th>
                    <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Tag</th>
                    <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden md:table-cell">Created</th>
                    <th className="w-32 px-4 py-3.5" />
                  </tr>
                </thead>
                <tbody>
                  {!app.versions || app.versions.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="px-4 py-12 text-center text-muted-foreground">
                        <p>No versions yet</p>
                        <p className="text-xs mt-1">Create a version to track scanner results</p>
                      </td>
                    </tr>
                  ) : (
                    app.versions.map((v) => (
                      <tr key={v.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-2">
                            <span className="font-medium">{v.name}</span>
                            {v.isDefault && (
                              <span className="inline-flex items-center gap-1 rounded-full bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-300 px-2 py-0.5 text-[10px] font-medium">
                                <Star className="h-3 w-3" />
                                default
                              </span>
                            )}
                          </div>
                        </td>
                        <td className="px-4 py-3 text-muted-foreground hidden sm:table-cell">
                          {v.branch || "-"}
                        </td>
                        <td className="px-4 py-3 text-muted-foreground hidden sm:table-cell">
                          {v.tag || "-"}
                        </td>
                        <td className="px-4 py-3 text-muted-foreground hidden md:table-cell">
                          {new Date(v.createdAt).toLocaleDateString()}
                        </td>
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-1">
                            {!v.isDefault && (
                              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleSetDefault(v)} title="Set as default">
                                <Star className="h-3.5 w-3.5" />
                              </Button>
                            )}
                            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => router.push(`/findings?applicationId=${v.applicationId}&applicationVersionId=${v.id}`)} title="View findings">
                              <Bug className="h-3.5 w-3.5" />
                            </Button>
                            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => openEditVersion(v)}>
                              <Pencil className="h-3.5 w-3.5" />
                            </Button>
                            <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive" onClick={() => setDeleteVersionTarget(v)}>
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
        </div>
      )}

      {activeTab === "settings" && (
        <div key="settings" className="animate-in fade-in duration-200">
          <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold">Webhooks</h3>
            <Button size="sm" onClick={openCreateWebhook} className="gap-1.5">
              <Plus className="h-4 w-4" />
              Add webhook
            </Button>
          </div>

          <Card>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">URL</th>
                    <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Events</th>
                    <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Status</th>
                    <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden md:table-cell">Created</th>
                    <th className="w-12 px-4 py-3.5" />
                  </tr>
                </thead>
                <tbody>
                  {webhooks.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="px-4 py-12 text-center text-muted-foreground">
                        <Globe className="h-8 w-8 mx-auto mb-2 opacity-40" />
                        <p>No webhooks configured</p>
                        <p className="text-xs mt-1">Add a webhook to receive security alerts</p>
                      </td>
                    </tr>
                  ) : (
                    webhooks.map((wh) => (
                      <tr key={wh.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                        <td className="px-4 py-3 font-medium max-w-[200px] truncate">{wh.url}</td>
                        <td className="px-4 py-3 text-muted-foreground hidden sm:table-cell">
                          <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{wh.events}</code>
                        </td>
                        <td className="px-4 py-3 hidden sm:table-cell">
                          <span className={`inline-flex items-center gap-1.5 text-sm ${wh.isActive ? "text-emerald-500" : "text-muted-foreground"}`}>
                            <span className="h-1.5 w-1.5 rounded-full bg-current" />
                            {wh.isActive ? "Active" : "Inactive"}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-muted-foreground hidden md:table-cell">
                          {wh.createdAt ? new Date(wh.createdAt).toLocaleDateString() : "-"}
                        </td>
                        <td className="px-4 py-3">
                          <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive" onClick={() => handleDeleteWebhook(wh)} title="Delete webhook">
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </td>
                      </tr>
                    ))
                  )}
                </tbody>
              </table>
            </div>
          </Card>

          {canManagePerms && (
            <>
              <div className="flex items-center justify-between">
                  <h3 className="text-lg font-semibold">Permissions</h3>
                <Button size="sm" onClick={openPermGrant} className="gap-1.5">
                  <Plus className="h-4 w-4" />
                  Add permission
                </Button>
              </div>

              <Card>
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b bg-muted/50">
                        <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">Subject</th>
                        <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">Action</th>
                        <th className="w-12 px-4 py-3.5" />
                      </tr>
                    </thead>
                    <tbody>
                      {appPerms.length === 0 ? (
                        <tr>
                          <td colSpan={3} className="px-4 py-12 text-center text-muted-foreground">
                            <Shield className="h-8 w-8 mx-auto mb-2 opacity-40" />
                            <p>No permissions configured</p>
                            <p className="text-xs mt-1">Grant access to users or teams</p>
                          </td>
                        </tr>
                      ) : (
                        appPerms.map((p, i) => (
                          <tr key={p.id || i} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                            <td className="px-4 py-3">
                              <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{p.subject}</code>
                            </td>
                            <td className="px-4 py-3 capitalize">{p.action}</td>
                            <td className="px-4 py-3">
                              <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive" onClick={() => handleRevokePerm(p)} title="Revoke">
                                <Trash2 className="h-3.5 w-3.5" />
                              </Button>
                            </td>
                          </tr>
                        ))
                      )}
                    </tbody>
                  </table>
                </div>
              </Card>
            </>
          )}
        </div>
      )}

      <Dialog open={permDialogOpen} onOpenChange={setPermDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add permission</DialogTitle>
            <DialogDescription>Grant a user or team access to this application.</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleGrantPerm}>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label>Subject type</Label>
                <div className="flex items-center gap-3">
                  <label className="flex items-center gap-1.5 text-sm">
                    <input
                      type="radio"
                      name="permSubjectType"
                      checked={permForm.subjectType === "user"}
                      onChange={() => { setPermForm({ ...permForm, subjectType: "user", subjectValue: "" }); }}
                      className="text-purple-600 focus:ring-purple-500"
                    />
                    User
                  </label>
                  <label className="flex items-center gap-1.5 text-sm">
                    <input
                      type="radio"
                      name="permSubjectType"
                      checked={permForm.subjectType === "team"}
                      onChange={() => { setPermForm({ ...permForm, subjectType: "team", subjectValue: "" }); }}
                      className="text-purple-600 focus:ring-purple-500"
                    />
                    Team
                  </label>
                </div>
              </div>

              {permForm.subjectType === "user" ? (
                <div className="grid gap-2">
                  <Label>User</Label>
                  <UserSearch
                    value={permForm.subjectValue}
                    onSelect={(userId) => setPermForm({ ...permForm, subjectValue: userId })}
                    onClear={() => setPermForm({ ...permForm, subjectValue: "" })}
                    placeholder="Search users (min 2 characters)..."
                  />
                </div>
              ) : (
                <div className="grid gap-2">
                  <Label>Team</Label>
                  <Select value={permForm.subjectValue} onValueChange={(v) => setPermForm({ ...permForm, subjectValue: v })}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select a team..." />
                    </SelectTrigger>
                    <SelectContent>
                      {permTeams.map((t) => (
                        <SelectItem key={t.id} value={String(t.id)}>{t.name}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              )}

              <div className="grid gap-2">
                <Label>Action</Label>
                <Select value={permForm.action} onValueChange={(v: any) => setPermForm({ ...permForm, action: v })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="read">Read</SelectItem>
                    <SelectItem value="write">Write</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setPermDialogOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={savingPerm || !permForm.subjectValue}>
                {savingPerm ? "Granting..." : "Grant access"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={versionDialogOpen} onOpenChange={setVersionDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editingVersion ? "Edit version" : "New version"}</DialogTitle>
            <DialogDescription>
              {editingVersion ? "Update the version details below." : "Create a new version for this application."}
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSaveVersion}>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="versionName">Name *</Label>
                <Input id="versionName" value={versionForm.name} onChange={(e) => setVersionForm({ ...versionForm, name: e.target.value })} placeholder="v1.0.0" required maxLength={100} autoComplete="off" data-1p-ignore />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="versionBranch">Branch</Label>
                <Input id="versionBranch" value={versionForm.branch} onChange={(e) => setVersionForm({ ...versionForm, branch: e.target.value })} placeholder="main" maxLength={200} autoComplete="off" data-1p-ignore />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="versionTag">Tag</Label>
                <Input id="versionTag" value={versionForm.tag} onChange={(e) => setVersionForm({ ...versionForm, tag: e.target.value })} placeholder="v1.0.0" maxLength={100} autoComplete="off" data-1p-ignore />
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="versionIsDefault"
                  checked={versionForm.isDefault}
                  onChange={(e) => setVersionForm({ ...versionForm, isDefault: e.target.checked })}
                  className="h-4 w-4 rounded border-gray-300 text-purple-600 focus:ring-purple-500"
                />
                <Label htmlFor="versionIsDefault" className="text-sm font-normal">Set as default version</Label>
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setVersionDialogOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={savingVersion}>{savingVersion ? "Saving..." : editingVersion ? "Save changes" : "Create version"}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={!!deleteVersionTarget} onOpenChange={(open) => !open && setDeleteVersionTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete version</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete version <strong>{deleteVersionTarget?.name}</strong>? All scans and findings associated with this version will also be deleted.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setDeleteVersionTarget(null)}>Cancel</Button>
            <Button type="button" variant="destructive" onClick={handleDeleteVersion}>Delete</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={ingestDialogOpen} onOpenChange={setIngestDialogOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>Upload scan results</DialogTitle>
            <DialogDescription>
              Upload a scanner output file for <strong>{app.name}</strong>. The file will be parsed and findings will be recorded.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleIngest}>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label>Scanner Output File *</Label>
                <div className="flex items-center gap-3">
                  <Button type="button" variant="outline" size="sm" onClick={() => document.getElementById('ingestFile')?.click()} className="shrink-0">
                    Choose File
                  </Button>
                  <span className="text-sm text-muted-foreground truncate">
                    {ingestFile ? ingestFile.name : 'No file selected'}
                  </span>
                  <input
                    id="ingestFile"
                    type="file"
                    accept=".json,.sarif,.txt"
                    onChange={(e) => setIngestFile(e.target.files?.[0] || null)}
                    className="hidden"
                  />
                </div>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="ingestScannerType">Scanner Type</Label>
                <Select value={ingestScannerType} onValueChange={setIngestScannerType}>
                  <SelectTrigger id="ingestScannerType">
                    <SelectValue placeholder="Auto-detect" />
                  </SelectTrigger>
                  <SelectContent>
                    {scannerTypes.map((st) => (
                      <SelectItem key={st.id} value={String(st.id)}>{st.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <p className="text-xs text-muted-foreground">Optional - will be auto-detected if not specified</p>
              </div>
              <div className="grid gap-2">
                <div className="flex items-center justify-between">
                  <Label>Version</Label>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <button type="button" className={`${ingestVersionMode === "select" ? "text-foreground font-medium" : ""}`} onClick={() => setIngestVersionMode("select")}>Existing</button>
                    <span>/</span>
                    <button type="button" className={`${ingestVersionMode === "new" ? "text-foreground font-medium" : ""}`} onClick={() => setIngestVersionMode("new")}>New</button>
                  </div>
                </div>
                {ingestVersionMode === "select" ? (
                  <Select value={ingestVersionName} onValueChange={setIngestVersionName}>
                    <SelectTrigger>
                      <SelectValue placeholder="Select a version" />
                    </SelectTrigger>
                    <SelectContent>
                        {(app.versions || []).map((v) => (
                        <SelectItem key={v.id} value={v.name}>{v.name}{v.isDefault ? " (default)" : ""}</SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                ) : (
                  <Input value={ingestVersionName} onChange={(e) => setIngestVersionName(e.target.value)} placeholder="v1.0.0" maxLength={100} autoComplete="off" data-1p-ignore />
                )}
                <p className="text-xs text-muted-foreground">Select an existing version or type a new name (will be created automatically)</p>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="ingestBranch">Branch</Label>
                <Input id="ingestBranch" value={ingestBranch} onChange={(e) => setIngestBranch(e.target.value)} placeholder="main" maxLength={200} autoComplete="off" data-1p-ignore />
                <p className="text-xs text-muted-foreground">Optional - will be stored with the version</p>
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setIngestDialogOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={ingesting || !ingestFile}>
                {ingesting ? "Uploading..." : "Upload & scan"}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={webhookDialogOpen} onOpenChange={setWebhookDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add webhook</DialogTitle>
            <DialogDescription>
              Configure a webhook to receive security alerts for this application.
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSaveWebhook}>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="whUrl">URL *</Label>
                <Input id="whUrl" value={webhookForm.url} onChange={(e) => setWebhookForm({ ...webhookForm, url: e.target.value })} placeholder="https://hooks.example.com/..." required maxLength={500} autoComplete="off" data-1p-ignore />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="whEvents">Events</Label>
                <Input id="whEvents" value={webhookForm.events} onChange={(e) => setWebhookForm({ ...webhookForm, events: e.target.value })} placeholder="finding.critical,finding.high" maxLength={200} autoComplete="off" data-1p-ignore />
                <p className="text-xs text-muted-foreground">Comma-separated event types (leave empty for all)</p>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="whSecret">Secret</Label>
                <Input id="whSecret" type="password" value={webhookForm.secret} onChange={(e) => setWebhookForm({ ...webhookForm, secret: e.target.value })} placeholder="Optional HMAC secret" maxLength={200} autoComplete="off" data-1p-ignore />
                <p className="text-xs text-muted-foreground">Used for HMAC-SHA256 signature header</p>
              </div>
              <div className="flex items-center gap-2">
                <input
                  type="checkbox"
                  id="whActive"
                  checked={webhookForm.isActive}
                  onChange={(e) => setWebhookForm({ ...webhookForm, isActive: e.target.checked })}
                  className="h-4 w-4 rounded border-gray-300 text-purple-600 focus:ring-purple-500"
                />
                <Label htmlFor="whActive" className="text-sm font-normal">Active</Label>
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setWebhookDialogOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={savingWebhook}>{savingWebhook ? "Saving..." : "Create webhook"}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
