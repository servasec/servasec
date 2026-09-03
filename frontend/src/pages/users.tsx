import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import {
  Shield,
  ShieldCheck,
  Ban,
  ShieldAlert,
  MoreHorizontal,
  Pencil,
  Trash2,
  Loader2,
} from "lucide-react";
import { PageHeader } from "@/components/page-header";
import axios from "@/lib/api";
import { toast } from "sonner";
import type { User } from "@/lib/types";

type UserAction = "edit" | "delete" | null;

export default function UsersPage() {
  const router = useRouter();
  const { loggedIn, user, authChecked } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [busyId, setBusyId] = useState<number | null>(null);

  const [editUser, setEditUser] = useState<User | null>(null);
  const [deleteUser, setDeleteUser] = useState<User | null>(null);
  const [editRole, setEditRole] = useState<"admin" | "member">("member");
  const [editBanned, setEditBanned] = useState(false);
  const [savingEdit, setSavingEdit] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const fetchUsers = () => {
    setLoading(true);
    axios
      .get<User[]>("/api/users")
      .then((res) => { setUsers(res.data); setLoadError(false); })
      .catch(() => { toast.error("Failed to load users"); setLoadError(true); })
      .finally(() => setLoading(false));
  };

  const fetchUsersTimer = () => setTimeout(() => { fetchUsers(); }, 300);

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn) {
      fetchUsers();
    }
  }, [authChecked, loggedIn]);

  const isSelf = (u: User) => u.id === user?.id;

  const openEdit = (u: User) => {
    setEditUser(u);
    setEditRole(u.role === "admin" ? "admin" : "member");
    setEditBanned(u.banned);
  };

  const closeEdit = () => {
    setEditUser(null);
    setSavingEdit(false);
  };

  const handleSaveEdit = async () => {
    if (!editUser || savingEdit) return;
    const self = isSelf(editUser);
    if (self && editRole !== "admin") {
      toast.error("You cannot remove your own admin role");
      return;
    }
    if (self && editBanned) {
      toast.error("You cannot ban yourself");
      return;
    }
    setSavingEdit(true);
    setBusyId(editUser.id);
    const prev = editUser;
    setUsers((prevUsers) =>
      prevUsers.map((u) =>
        u.id === prev.id ? { ...u, role: editRole, banned: editBanned } : u
      )
    );
    try {
      await axios.put(`/api/users/${editUser.id}`, {
        role: editRole,
        banned: editBanned,
      });
      toast.success(`Updated ${editUser.username}`);
      setEditUser(null);
      fetchUsersTimer();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to update user");
      setUsers((prevUsers) =>
        prevUsers.map((u) =>
          u.id === prev.id ? { ...u, role: prev.role, banned: prev.banned } : u
        )
      );
    } finally {
      setSavingEdit(false);
      setBusyId(null);
    }
  };

  const handleDelete = async () => {
    if (!deleteUser || deleting) return;
    setDeleting(true);
    setBusyId(deleteUser.id);
    try {
      await axios.delete(`/api/users/${deleteUser.id}`);
      toast.success(`Deleted ${deleteUser.username}`);
      setDeleteUser(null);
      fetchUsersTimer();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to delete user");
    } finally {
      setDeleting(false);
      setBusyId(null);
    }
  };

  if (!authChecked || !loggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-muted-foreground">Loading...</div>
      </div>
    );
  }

  if (user?.role !== "admin") {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <p className="text-muted-foreground">You do not have permission to view this page.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader crumbs={[{ label: "Administration" }, { label: "Users" }]} />

      <Card>
        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground">User</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden md:table-cell">Email</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden sm:table-cell">Role</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden lg:table-cell">Status</th>
                <th className="text-left px-4 py-2.5 font-medium text-muted-foreground hidden lg:table-cell">Joined</th>
                <th className="w-12 px-4 py-2.5" />
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array.from({ length: 3 }).map((_, i) => (
                  <tr key={i} className="border-b last:border-0">
                    {Array.from({ length: 6 }).map((_, j) => (
                      <td key={j} className="px-4 py-2">
                        <Skeleton className="h-5 w-full max-w-[120px]" />
                      </td>
                    ))}
                  </tr>
                ))
              ) : loadError ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center text-muted-foreground">
                    <ShieldAlert className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>Failed to load users</p>
                    <p className="text-xs mt-1">The server may be unavailable</p>
                  </td>
                </tr>
              ) : users.length === 0 ? (
                <tr>
                  <td colSpan={6} className="px-4 py-12 text-center text-muted-foreground">
                    No users found
                  </td>
                </tr>
              ) : (
                users.map((u) => {
                  const self = isSelf(u);
                  const busy = busyId === u.id;
                  const canDelete = !self;
                  return (
                    <tr key={u.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                      <td className="px-4 py-2">
                        <div className="flex items-center gap-3">
                          <Avatar className="h-6 w-6">
                            <AvatarFallback className="bg-primary/10 text-primary text-xs">
                              {u.username.charAt(0).toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <span className="font-medium">
                            {u.username}
                            {self && <span className="ml-2 text-[10px] font-medium text-muted-foreground">(you)</span>}
                            {u.oauthProvider && (
                              <span className="ml-2 inline-flex items-center gap-1 rounded-md border px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
                                {u.oauthProvider === "github" ? "GitHub" : u.oauthProvider === "gitlab" ? "GitLab" : u.oauthProvider}
                              </span>
                            )}
                          </span>
                        </div>
                      </td>
                      <td className="px-4 py-2 text-muted-foreground hidden md:table-cell">
                        {u.email}
                      </td>
                      <td className="px-4 py-2 hidden sm:table-cell">
                        <span className="inline-flex items-center gap-1.5 text-xs">
                          {u.role === "admin" ? (
                            <ShieldCheck className="h-3.5 w-3.5 text-primary" />
                          ) : (
                            <Shield className="h-3.5 w-3.5 text-muted-foreground" />
                          )}
                          <span className="capitalize">{u.role}</span>
                        </span>
                      </td>
                      <td className="px-4 py-2 hidden lg:table-cell">
                        {u.banned ? (
                          <span className="inline-flex items-center gap-1.5 text-destructive text-xs">
                            <Ban className="h-3.5 w-3.5" />
                            Banned
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1.5 text-emerald-500 text-xs">
                            <span className="h-1.5 w-1.5 rounded-full bg-current" />
                            Active
                          </span>
                        )}
                      </td>
                      <td className="px-4 py-2 text-muted-foreground hidden lg:table-cell">
                        {u.createdAt ? new Date(u.createdAt).toLocaleDateString() : "-"}
                      </td>
                      <td className="px-4 py-2">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon" className="h-7 w-7" disabled={busy} title="Actions">
                              {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <MoreHorizontal className="h-4 w-4" />}
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuItem onClick={() => openEdit(u)}>
                              <Pencil className="h-3.5 w-3.5 mr-2" />
                              Edit
                            </DropdownMenuItem>
                            {canDelete && (
                              <>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                  className="text-destructive focus:text-destructive"
                                  onClick={() => setDeleteUser(u)}
                                >
                                  <Trash2 className="h-3.5 w-3.5 mr-2" />
                                  Delete
                                </DropdownMenuItem>
                              </>
                            )}
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </td>
                    </tr>
                  );
                })
              )}
            </tbody>
          </table>
        </div>
      </Card>

      <Dialog open={!!editUser} onOpenChange={(open) => !open && closeEdit()}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit user</DialogTitle>
            <DialogDescription>Update this user&apos;s role or account status.</DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label>User</Label>
              <p className="text-sm text-foreground">
                {editUser?.username}
                {editUser && editUser.oauthProvider && (
                  <span className="ml-2 rounded-md border px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
                    {editUser.oauthProvider === "github" ? "GitHub" : editUser.oauthProvider === "gitlab" ? "GitLab" : editUser.oauthProvider}
                  </span>
                )}
              </p>
            </div>
            <div className="grid gap-2">
              <Label htmlFor="editRole">Role</Label>
              <Select value={editRole} onValueChange={(v: "admin" | "member") => setEditRole(v)}>
                <SelectTrigger id="editRole">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="admin">Admin</SelectItem>
                  <SelectItem value="member">Member</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid gap-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="editBanned">Banned</Label>
                <Switch id="editBanned" checked={editBanned} onCheckedChange={setEditBanned} />
              </div>
              <p className="text-xs text-muted-foreground">
                Banned users cannot sign in.
              </p>
            </div>
          </div>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={closeEdit}>Cancel</Button>
            <Button onClick={handleSaveEdit} disabled={savingEdit}>
              {savingEdit ? "Saving..." : "Save changes"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!deleteUser} onOpenChange={(open) => !open && setDeleteUser(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete user</DialogTitle>
            <DialogDescription>
              This will permanently delete <span className="font-medium text-foreground">{deleteUser?.username}</span> and revoke their access. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setDeleteUser(null)}>Cancel</Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleting}>
              {deleting ? "Deleting..." : "Delete user"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}