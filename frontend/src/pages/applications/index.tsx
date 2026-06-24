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
import { Plus, Pencil, Trash2, Key, Copy, AppWindow } from "lucide-react";
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

interface Application {
  id: number;
  name: string;
  description: string;
  slug: string;
  groupId: number;
  repositoryUrl: string;
  apiToken?: string;
  createdAt: string;
}

interface Group {
  id: number;
  name: string;
}

interface FormData {
  name: string;
  description: string;
  slug: string;
  groupId: string;
  repositoryUrl: string;
  assetCriticality: string;
}

const emptyForm: FormData = { name: "", description: "", slug: "", groupId: "", repositoryUrl: "", assetCriticality: "medium" };

export default function ApplicationsListPage() {
  const router = useRouter();
  const { loggedIn, authChecked } = useAuth();
  const [apps, setApps] = useState<Application[]>([]);
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Application | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Application | null>(null);
  const [tokenApp, setTokenApp] = useState<Application | null>(null);
  const [newToken, setNewToken] = useState("");
  const [form, setForm] = useState<FormData>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [slugManuallyEdited, setSlugManuallyEdited] = useState(false);

  const slugify = (s: string) => s.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "").slice(0, 100);

  const fetchData = () => {
    Promise.all([
      axios.get("/api/applications"),
      axios.get("/api/groups"),
    ])
      .then(([appsRes, groupsRes]) => {
        setApps(appsRes.data);
        setGroups(groupsRes.data);
      })
      .catch(() => toast.error("Failed to load data"))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn) {
      fetchData();
    }
  }, [authChecked, loggedIn]);

  const groupMap = Object.fromEntries(groups.map((g) => [g.id, g.name]));

  const openCreate = () => {
    setEditing(null);
    setForm(emptyForm);
    setSlugManuallyEdited(false);
    setDialogOpen(true);
  };

  const openEdit = (a: Application) => {
    setEditing(a);
    setForm({
      name: a.name,
      description: a.description || "",
      slug: a.slug,
      groupId: String(a.groupId),
      repositoryUrl: a.repositoryUrl || "",
      assetCriticality: (a as any).assetCriticality || "medium",
    });
    setSlugManuallyEdited(true);
    setDialogOpen(true);
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    const payload = { ...form, groupId: Number(form.groupId) };
    try {
      if (editing) {
        await axios.put(`/api/applications/${editing.id}`, payload);
        toast.success("Application updated");
      } else {
        await axios.post("/api/applications", payload);
        toast.success("Application created");
      }
      setDialogOpen(false);
      fetchData();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to save application");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await axios.delete(`/api/applications/${deleteTarget.id}`);
      toast.success("Application deleted");
      setDeleteTarget(null);
      fetchData();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to delete application");
    }
  };

  const handleRegenerateToken = async (app: Application) => {
    try {
      const res = await axios.post(`/api/applications/${app.id}/regenerate-token`);
      setNewToken(res.data.apiToken);
      setTokenApp(app);
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to regenerate token");
    }
  };

  const copyToken = () => {
    navigator.clipboard.writeText(newToken);
    toast.success("Token copied to clipboard");
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
      <div className="flex items-center justify-between">
        <PageHeader crumbs={[{ label: "Security" }, { label: "Applications" }]} />
        <Button onClick={openCreate} className="gap-2">
          <Plus className="h-4 w-4" />
          New application
        </Button>
      </div>

      <Card>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">Name</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden md:table-cell">Slug</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Group</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden lg:table-cell">Repository</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden lg:table-cell">Created</th>
                <th className="w-36 px-4 py-3.5" />
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array.from({ length: 3 }).map((_, i) => (
                  <tr key={i} className="border-b last:border-0">
                    {Array.from({ length: 6 }).map((_, j) => (
                      <td key={j} className="px-4 py-3">
                        <Skeleton className="h-5 w-full max-w-[120px]" />
                      </td>
                    ))}
                  </tr>
                ))
              ) : apps.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center text-muted-foreground">
                    <AppWindow className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>No applications yet</p>
                    <p className="text-xs mt-1">Create an application to start scanning</p>
                  </td>
                </tr>
              ) : (
                apps.map((a) => (
                  <tr key={a.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors cursor-pointer" onClick={() => router.push(`/applications/${a.id}`)}>
                    <td className="px-4 py-3 font-medium">{a.name}</td>
                    <td className="px-4 py-3 hidden md:table-cell">
                      <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{a.slug}</code>
                    </td>
                    <td className="px-4 py-3 hidden sm:table-cell text-muted-foreground">
                      {groupMap[a.groupId] || `#${a.groupId}`}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground hidden lg:table-cell max-w-[160px] truncate">
                      {a.repositoryUrl || "-"}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground hidden lg:table-cell">
                      {a.createdAt ? new Date(a.createdAt).toLocaleDateString() : "-"}
                    </td>
                    <td className="px-4 py-3" onClick={(e) => e.stopPropagation()}>
                      <div className="flex items-center gap-1">
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleRegenerateToken(a)} title="Regenerate API token">
                          <Key className="h-3.5 w-3.5" />
                        </Button>
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => openEdit(a)}>
                          <Pencil className="h-3.5 w-3.5" />
                        </Button>
                        <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive hover:text-destructive" onClick={() => setDeleteTarget(a)}>
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

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{editing ? "Edit application" : "New application"}</DialogTitle>
            <DialogDescription>
              {editing ? "Update the application details below." : "Register a new application for security scanning."}
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSave}>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="name">Name</Label>
                <Input id="name" value={form.name} onChange={(e) => {
                  const newName = e.target.value;
                  setForm({ ...form, name: newName, slug: slugManuallyEdited ? form.slug : slugify(newName) });
                }} placeholder="My App" required maxLength={200} autoComplete="off" data-1p-ignore />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="slug">Slug</Label>
                <Input id="slug" value={form.slug} onChange={(e) => {
                  setSlugManuallyEdited(true);
                  setForm({ ...form, slug: e.target.value });
                }} placeholder="my-app" required maxLength={100} autoComplete="off" data-1p-ignore />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="groupId">Group</Label>
                <Select value={form.groupId} onValueChange={(v) => setForm({ ...form, groupId: v })}>
                  <SelectTrigger id="groupId">
                    <SelectValue placeholder="Select a group" />
                  </SelectTrigger>
                  <SelectContent>
                    {groups.map((g) => (
                      <SelectItem key={g.id} value={String(g.id)}>{g.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="description">Description</Label>
                <Input id="description" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} placeholder="Optional description" maxLength={1000} autoComplete="off" data-1p-ignore />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="repositoryUrl">Repository URL</Label>
                <Input id="repositoryUrl" value={form.repositoryUrl} onChange={(e) => setForm({ ...form, repositoryUrl: e.target.value })} placeholder="https://github.com/org/repo" maxLength={500} autoComplete="off" data-1p-ignore />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="assetCriticality">Asset criticality</Label>
                <Select value={form.assetCriticality} onValueChange={(v) => setForm({ ...form, assetCriticality: v })}>
                  <SelectTrigger id="assetCriticality">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="critical">Critical</SelectItem>
                    <SelectItem value="high">High</SelectItem>
                    <SelectItem value="medium">Medium</SelectItem>
                    <SelectItem value="low">Low</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={saving}>{saving ? "Saving..." : editing ? "Save changes" : "Create application"}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete application</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete <strong>{deleteTarget?.name}</strong>? This will also delete all associated scans, findings, and versions.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setDeleteTarget(null)}>Cancel</Button>
            <Button type="button" variant="destructive" onClick={handleDelete}>Delete</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!tokenApp} onOpenChange={(open) => { if (!open) { setTokenApp(null); setNewToken(""); } }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>API Token</DialogTitle>
            <DialogDescription>
              New API token for <strong>{tokenApp?.name}</strong>. Copy it now - you won&apos;t be able to see it again.
            </DialogDescription>
          </DialogHeader>
          <div className="flex items-center gap-2 py-4">
            <code className="flex-1 rounded border bg-muted px-3 py-2 text-sm font-mono break-all select-all">{newToken}</code>
            <Button variant="outline" size="icon" onClick={copyToken}>
              <Copy className="h-4 w-4" />
            </Button>
          </div>
          <DialogFooter>
            <Button onClick={() => { setTokenApp(null); setNewToken(""); }}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
