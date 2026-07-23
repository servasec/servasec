import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { Plus, Pencil, Trash2, FolderKanban } from "lucide-react";
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

interface Group {
  id: number;
  name: string;
  description: string;
  path: string;
  createdAt: string;
}

interface FormData {
  name: string;
  description: string;
  path: string;
}

const emptyForm = { name: "", description: "", path: "" };

export default function GroupsPage() {
  const router = useRouter();
  const { loggedIn, authChecked } = useAuth();
  const [groups, setGroups] = useState<Group[]>([]);
  const [loading, setLoading] = useState(true);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editing, setEditing] = useState<Group | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Group | null>(null);
  const [form, setForm] = useState<FormData>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [pathManuallyEdited, setPathManuallyEdited] = useState(false);

  const slugify = (s: string) => s.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "").slice(0, 100);

  const fetchGroups = () => {
    axios
      .get("/api/groups")
      .then((res) => setGroups(res.data))
      .catch(() => toast.error("Failed to load groups"))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn) {
      fetchGroups();
    }
  }, [authChecked, loggedIn]);

  const openCreate = () => {
    setEditing(null);
    setForm(emptyForm);
    setPathManuallyEdited(false);
    setDialogOpen(true);
  };

  const openEdit = (g: Group) => {
    setEditing(g);
    setForm({ name: g.name, description: g.description || "", path: g.path });
    setPathManuallyEdited(true);
    setDialogOpen(true);
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      if (editing) {
        await axios.put(`/api/groups/${editing.id}`, form);
        toast.success("Group updated");
      } else {
        await axios.post("/api/groups", form);
        toast.success("Group created");
      }
      setDialogOpen(false);
      fetchGroups();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to save group");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await axios.delete(`/api/groups/${deleteTarget.id}`);
      toast.success("Group deleted");
      setDeleteTarget(null);
      fetchGroups();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to delete group");
    }
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
        <PageHeader crumbs={[{ label: "Security" }, { label: "Groups" }]} />
          <Button onClick={openCreate} className="h-8 text-xs gap-1.5">
            <Plus className="h-3.5 w-3.5" />
            New group
          </Button>
      </div>

      <Card>
        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground">Name</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden md:table-cell">Description</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden sm:table-cell">Path</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden lg:table-cell">Created</th>
                <th className="w-20 px-4 py-2.5" />
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array.from({ length: 3 }).map((_, i) => (
                  <tr key={i} className="border-b last:border-0">
                    {Array.from({ length: 5 }).map((_, j) => (
                      <td key={j} className="px-4 py-2">
                        <Skeleton className="h-4 w-full max-w-[120px] animate-pulse" />
                      </td>
                    ))}
                  </tr>
                ))
              ) : groups.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-muted-foreground">
                    <FolderKanban className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>No groups yet</p>
                    <p className="text-xs mt-1">Create a group to organize your applications</p>
                  </td>
                </tr>
              ) : (
                groups.map((g) => (
                  <tr key={g.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-2 font-medium">{g.name}</td>
                    <td className="px-4 py-2 text-muted-foreground hidden md:table-cell max-w-[200px] truncate">
                      {g.description || "-"}
                    </td>
                    <td className="px-4 py-2 hidden sm:table-cell">
                      <code className="text-[11px] font-mono bg-muted px-1.5 py-0.5 rounded">{g.path}</code>
                    </td>
                    <td className="px-4 py-2 text-muted-foreground hidden lg:table-cell">
                      {g.createdAt ? new Date(g.createdAt).toLocaleDateString() : "-"}
                    </td>
                    <td className="px-4 py-2">
                      <div className="flex items-center gap-1">
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => openEdit(g)}>
                          <Pencil className="h-3.5 w-3.5" />
                        </Button>
                        <Button variant="destructive-ghost" size="icon" className="h-8 w-8" onClick={() => setDeleteTarget(g)}>
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
            <DialogTitle>{editing ? "Edit group" : "New group"}</DialogTitle>
            <DialogDescription>
              {editing ? "Update the group details below." : "Create a new group to organize your applications."}
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSave}>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="name">Name</Label>
                <Input id="name" type="text" value={form.name} onChange={(e) => {
                  const newName = e.target.value;
                  setForm({ ...form, name: newName, path: pathManuallyEdited ? form.path : slugify(newName) });
                }} placeholder="My Group" required maxLength={100} autoComplete="off" data-1p-ignore />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="description">Description</Label>
                <Input id="description" type="text" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} placeholder="Optional description" maxLength={500} autoComplete="off" data-1p-ignore />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="path">Path</Label>
                <Input id="path" type="text" value={form.path} onChange={(e) => {
                  setPathManuallyEdited(true);
                  setForm({ ...form, path: e.target.value });
                }} placeholder="my-group" required maxLength={100} autoComplete="off" data-1p-ignore />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={saving}>{saving ? "Saving..." : editing ? "Save changes" : "Create group"}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete group</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete <strong>{deleteTarget?.name}</strong>? This action cannot be undone. Applications in this group will also be deleted.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setDeleteTarget(null)}>Cancel</Button>
            <Button type="button" variant="destructive" onClick={handleDelete}>Delete</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
