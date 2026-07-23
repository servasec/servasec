import { useEffect, useState, useMemo } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Shield, ShieldOff, Plus, Trash2 } from "lucide-react";
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

interface Permission {
  id: number;
  subject: string;
  resource: string;
  action: string;
  createdAt: string;
}

interface UserItem {
  id: number;
  username: string;
  email: string;
}

interface TeamItem {
  id: number;
  name: string;
}

export default function PermissionsPage() {
  const router = useRouter();
  const { loggedIn, user, authChecked } = useAuth();
  const [subjectType, setSubjectType] = useState<"user" | "team">("user");
  const [subjectValue, setSubjectValue] = useState("");
  const [permissions, setPermissions] = useState<Permission[]>([]);
  const [loadingPerms, setLoadingPerms] = useState(false);
  const [users, setUsers] = useState<UserItem[]>([]);
  const [teams, setTeams] = useState<TeamItem[]>([]);
  const [apps, setApps] = useState<{ id: number; name: string }[]>([]);
  const [groups, setGroups] = useState<{ id: number; name: string }[]>([]);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [grantForm, setGrantForm] = useState<{ subjectType: "user" | "team"; subjectValue: string; resourceType: "applications" | "groups"; resourceId: string; action: string }>({ subjectType: "user", subjectValue: "", resourceType: "applications", resourceId: "", action: "read" });
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn && user?.role === "admin") {
      Promise.all([
        axios.get("/api/users"),
        axios.get("/api/teams"),
        axios.get("/api/applications"),
        axios.get("/api/groups"),
      ]).then(([usersRes, teamsRes, appsRes, groupsRes]) => {
        setUsers(usersRes.data || []);
        setTeams(teamsRes.data || []);
        setApps(appsRes.data || []);
        setGroups(groupsRes.data || []);
      }).catch(() => {});
    }
  }, [authChecked, loggedIn, user]);

  const subject = subjectValue ? `${subjectType}:${subjectValue}` : "";

  const subjectLabel = useMemo(() => {
    if (!subjectValue) return "";
    if (subjectType === "user") {
      const u = users.find((u) => String(u.id) === subjectValue);
      return u ? `${u.username} (${u.email})` : `user:${subjectValue}`;
    }
    const t = teams.find((t) => String(t.id) === subjectValue);
    return t ? t.name : `team:${subjectValue}`;
  }, [subjectType, subjectValue, users, teams]);

  const fetchPermissions = () => {
    if (!subject) { setPermissions([]); return; }
    setLoadingPerms(true);
    axios.get(`/api/admin/permissions?subject=${subject}`)
      .then((res) => setPermissions(res.data || []))
      .catch(() => toast.error("Failed to load permissions"))
      .finally(() => setLoadingPerms(false));
  };

  useEffect(() => {
    if (subject) fetchPermissions();
  }, [subject]);

  const openGrant = () => {
    setGrantForm({ subjectType, subjectValue, resourceType: "applications", resourceId: "", action: "read" });
    setDialogOpen(true);
  };

  const handleGrant = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!grantForm.subjectValue || !grantForm.resourceId) {
      toast.error("Please fill all required fields");
      return;
    }
    setSaving(true);
    try {
      await axios.post("/api/admin/permissions", {
        subject: `${grantForm.subjectType}:${grantForm.subjectValue}`,
        resource: `/${grantForm.resourceType}/${grantForm.resourceId}`,
        action: grantForm.action,
      });
      toast.success("Permission granted");
      setDialogOpen(false);
      setSubjectType(grantForm.subjectType);
      setSubjectValue(grantForm.subjectValue);
      fetchPermissions();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to grant permission");
    } finally {
      setSaving(false);
    }
  };

  const handleRevoke = async (perm: Permission) => {
    try {
      await axios.delete("/api/admin/permissions", {
        data: { subject: perm.subject, resource: perm.resource, action: perm.action },
      });
      toast.success("Permission revoked");
      fetchPermissions();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to revoke permission");
    }
  };

  const availableOptions = subjectType === "user" ? users : teams;

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
      <PageHeader crumbs={[{ label: "Administration" }, { label: "Permissions" }]} />

      <div className="flex items-center gap-2 flex-wrap">
        <Select value={subjectType} onValueChange={(v: "user" | "team") => { setSubjectType(v); setSubjectValue(""); setPermissions([]); }}>
          <SelectTrigger className="w-24">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="user">User</SelectItem>
            <SelectItem value="team">Team</SelectItem>
          </SelectContent>
        </Select>
        {subjectType === "user" ? (
          <Select value={subjectValue} onValueChange={setSubjectValue}>
            <SelectTrigger className="w-64">
              <SelectValue placeholder="Select a user..." />
            </SelectTrigger>
            <SelectContent>
              {users.map((u) => (
                <SelectItem key={u.id} value={String(u.id)}>
                  {u.username} ({u.email})
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        ) : (
          <Select value={subjectValue} onValueChange={setSubjectValue}>
            <SelectTrigger className="w-64">
              <SelectValue placeholder="Select a team..." />
            </SelectTrigger>
            <SelectContent>
              {teams.map((t) => (
                <SelectItem key={t.id} value={String(t.id)}>{t.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
        {subject && (
          <Button onClick={openGrant} size="sm" className="gap-1.5">
            <Plus className="h-4 w-4" />
            Grant access
          </Button>
        )}
      </div>

      {subject && (
        <div>
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground mb-3">
              Permissions for <span className="font-medium text-foreground">{subjectLabel}</span> (<code className="text-xs bg-muted px-1 py-0.5 rounded">{subject}</code>)
            </p>
          </div>
          <Card>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b bg-muted/50">
                    <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">Subject</th>
                    <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">Resource</th>
                    <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Action</th>
                    <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden md:table-cell">Granted</th>
                    <th className="w-12 px-4 py-3.5" />
                  </tr>
                </thead>
                <tbody>
                  {loadingPerms ? (
                    Array.from({ length: 2 }).map((_, i) => (
                      <tr key={i} className="border-b last:border-0">
                        {Array.from({ length: 5 }).map((_, j) => (
                          <td key={j} className="px-4 py-3"><Skeleton className="h-5 w-full max-w-[100px]" /></td>
                        ))}
                      </tr>
                    ))
                  ) : permissions.length === 0 ? (
                    <tr>
                      <td colSpan={5} className="px-4 py-12 text-center text-muted-foreground">
                        <Shield className="h-8 w-8 mx-auto mb-2 opacity-40" />
                        <p>No permissions for {subjectLabel}</p>
                        <p className="text-xs mt-1">Grant access to applications or groups</p>
                      </td>
                    </tr>
                  ) : (
                    permissions.map((p, i) => (
                      <tr key={p.id || i} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                        <td className="px-4 py-3">
                          <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{p.subject}</code>
                        </td>
                        <td className="px-4 py-3">
                          <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{p.resource}</code>
                        </td>
                        <td className="px-4 py-3 hidden sm:table-cell capitalize">{p.action}</td>
                        <td className="px-4 py-3 text-muted-foreground hidden md:table-cell">
                          {p.createdAt ? new Date(p.createdAt).toLocaleDateString() : "-"}
                        </td>
                        <td className="px-4 py-3">
                          <Button variant="destructive-ghost" size="icon" className="h-8 w-8" onClick={() => handleRevoke(p)} title="Revoke">
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
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Grant permission</DialogTitle>
            <DialogDescription>Grant access to a specific resource.</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleGrant}>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label>Subject</Label>
                <p className="text-sm">
                  <span className="font-medium capitalize">{grantForm.subjectType}</span>{": "}
                  {grantForm.subjectType === "user"
                    ? users.find((u) => String(u.id) === grantForm.subjectValue)?.username || `#${grantForm.subjectValue}`
                    : teams.find((t) => String(t.id) === grantForm.subjectValue)?.name || `#${grantForm.subjectValue}`
                  }
                  {" "}<code className="text-xs bg-muted px-1 py-0.5 rounded">{grantForm.subjectType}:{grantForm.subjectValue}</code>
                </p>
              </div>
              <div className="grid gap-2">
                <Label>Resource type</Label>
                <div className="flex items-center gap-3">
                  <label className="flex items-center gap-1.5 text-sm">
                    <input
                      type="radio"
                      name="resourceType"
                      checked={grantForm.resourceType === "applications"}
                      onChange={() => setGrantForm({ ...grantForm, resourceType: "applications" })}
                      className="text-purple-600 focus:ring-purple-500"
                    />
                    Application
                  </label>
                  <label className="flex items-center gap-1.5 text-sm">
                    <input
                      type="radio"
                      name="resourceType"
                      checked={grantForm.resourceType === "groups"}
                      onChange={() => setGrantForm({ ...grantForm, resourceType: "groups" })}
                      className="text-purple-600 focus:ring-purple-500"
                    />
                    Group
                  </label>
                </div>
              </div>
              <div className="grid gap-2">
                <Label>Resource</Label>
                <Select value={grantForm.resourceId} onValueChange={(v) => setGrantForm({ ...grantForm, resourceId: v })}>
                  <SelectTrigger>
                    <SelectValue placeholder={grantForm.resourceType === "applications" ? "Select an application..." : "Select a group..."} />
                  </SelectTrigger>
                  <SelectContent>
                    {(grantForm.resourceType === "applications" ? apps : groups).map((r) => (
                      <SelectItem key={r.id} value={String(r.id)}>{r.name}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="grantAction">Action</Label>
                <Select value={grantForm.action} onValueChange={(v) => setGrantForm({ ...grantForm, action: v })}>
                  <SelectTrigger id="grantAction">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="read">Read</SelectItem>
                    <SelectItem value="write">Write</SelectItem>
                    <SelectItem value="*">* (All)</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={saving}>{saving ? "Granting..." : "Grant access"}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
