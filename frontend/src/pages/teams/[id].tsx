import { useEffect, useState, useMemo } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Shield, ShieldCheck, UserMinus, Plus, UsersRound, ArrowLeft } from "lucide-react";
import Link from "next/link";
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

interface Team {
  id: number;
  name: string;
  description: string;
  createdAt: string;
  members?: Member[];
}

interface Member {
  userId: number;
  role: string;
  userName: string;
}

interface SearchUser {
  id: number;
  username: string;
}

export default function TeamDetailPage() {
  const router = useRouter();
  const { id } = router.query;
  const { loggedIn, user, authChecked } = useAuth();
  const [team, setTeam] = useState<Team | null>(null);
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [removeTarget, setRemoveTarget] = useState<Member | null>(null);
  const [addForm, setAddForm] = useState({ userId: "", role: "member" });
  const [saving, setSaving] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<SearchUser[]>([]);
  const [searching, setSearching] = useState(false);

  const currentMemberRole = useMemo(() => {
    if (!user || !team) return null;
    const m = (team.members || []).find((m) => m.userId === Number(user.id));
    return m?.role || null;
  }, [user, team]);

  const fetchTeam = () => {
    if (!id) return;
    axios.get(`/api/teams/${id}`)
      .then((teamRes) => {
        const t = teamRes.data;
        setTeam(t);
        setMembers(t.members || []);
      })
      .catch(() => {
        toast.error("Failed to load team");
        router.replace("/teams");
      })
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn && id) {
      fetchTeam();
    }
  }, [authChecked, loggedIn, id]);

  useEffect(() => {
    if (!addDialogOpen) {
      setSearchQuery("");
      setSearchResults([]);
      setAddForm({ userId: "", role: "member" });
    }
  }, [addDialogOpen]);

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

  const handleAddMember = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!addForm.userId) return;
    setSaving(true);
    try {
      await axios.post(`/api/teams/${id}/members`, {
        userId: Number(addForm.userId),
        role: addForm.role,
      });
      toast.success("Member added");
      setAddDialogOpen(false);
      fetchTeam();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to add member");
    } finally {
      setSaving(false);
    }
  };

  const handleRemoveMember = async () => {
    if (!removeTarget) return;
    try {
      await axios.delete(`/api/teams/${id}/members/${removeTarget.userId}`);
      toast.success("Member removed");
      setRemoveTarget(null);
      fetchTeam();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to remove member");
    }
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
        <PageHeader crumbs={[{ label: "Administration", href: "/teams" }, { label: "..." }]} />
        <Card><CardContent className="p-8"><Skeleton className="h-6 w-48" /></CardContent></Card>
        <Card><CardHeader><Skeleton className="h-5 w-32" /></CardHeader><CardContent><Skeleton className="h-20 w-full" /></CardContent></Card>
      </div>
    );
  }

  if (!team) return null;

  return (
    <div className="space-y-6">
      <PageHeader crumbs={[{ label: "Administration", href: "/teams" }, { label: team.name }]} />

      <div className="flex items-center gap-4">
        <Link href="/teams">
          <Button variant="ghost" size="icon" className="h-8 w-8">
            <ArrowLeft className="h-4 w-4" />
          </Button>
        </Link>
        <div>
          <h1 className="text-xl font-bold">{team.name}</h1>
          {team.description && (
            <p className="text-sm text-muted-foreground mt-0.5">{team.description}</p>
          )}
        </div>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle className="text-base flex items-center gap-2">
              <UsersRound className="h-4 w-4" />
              Members
            </CardTitle>
            {currentMemberRole === "admin" && (
              <Button size="sm" className="gap-2" onClick={() => setAddDialogOpen(true)}>
                <Plus className="h-3.5 w-3.5" />
                Add member
              </Button>
            )}
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="text-left px-4 py-3 font-medium text-muted-foreground">User</th>
                  <th className="text-left px-4 py-3 font-medium text-muted-foreground">Role</th>
                  <th className="w-12 px-4 py-3" />
                </tr>
              </thead>
              <tbody>
                  {(members || []).length === 0 ? (
                    <tr>
                      <td colSpan={3} className="px-4 py-8 text-center text-muted-foreground">
                        No members yet
                      </td>
                    </tr>
                  ) : (
                    members.map((m) => (
                      <tr key={m.userId} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                        <td className="px-4 py-3">
                          <div className="flex items-center gap-3">
                            <Avatar className="h-7 w-7">
                              <AvatarFallback className="bg-primary/10 text-primary text-[11px]">
                                {(m.userName || "?").charAt(0).toUpperCase()}
                              </AvatarFallback>
                            </Avatar>
                            <span className="font-medium">{m.userName || `User #${m.userId}`}</span>
                          </div>
                        </td>
                      <td className="px-4 py-3">
                        <span className="inline-flex items-center gap-1.5 text-sm">
                          {m.role === "admin" ? (
                            <ShieldCheck className="h-3.5 w-3.5 text-primary" />
                          ) : (
                            <Shield className="h-3.5 w-3.5 text-muted-foreground" />
                          )}
                          <span className="capitalize">{m.role}</span>
                        </span>
                      </td>
                      <td className="px-4 py-3">
                        {currentMemberRole === "admin" && (
                          <Button variant="ghost" size="icon" className="h-7 w-7 text-muted-foreground hover:text-destructive" onClick={() => setRemoveTarget(m)}>
                            <UserMinus className="h-3.5 w-3.5" />
                          </Button>
                        )}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>

      <Dialog open={addDialogOpen} onOpenChange={setAddDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Add member</DialogTitle>
            <DialogDescription>Search and add a user to this team.</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleAddMember}>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label>User</Label>
                <Input
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Search users (min 2 characters)..."
                  autoComplete="off"
                  data-1p-ignore
                />
                {searching && (
                  <p className="text-xs text-muted-foreground">Searching...</p>
                )}
                {searchResults.length > 0 && !addForm.userId && (
                  <div className="border rounded-md divide-y max-h-40 overflow-y-auto">
                    {searchResults.filter((u) => !(members || []).some((m) => m.userId === u.id)).map((u) => (
                      <button
                        key={u.id}
                        type="button"
                        className="w-full text-left px-3 py-2 text-sm hover:bg-accent transition-colors"
                        onClick={() => {
                          setAddForm({ ...addForm, userId: String(u.id) });
                          setSearchResults([]);
                          setSearchQuery(u.username);
                        }}
                      >
                        {u.username} <span className="text-muted-foreground">(id: {u.id})</span>
                      </button>
                    ))}
                  </div>
                )}
                {searchQuery.length >= 2 && !searching && searchResults.length === 0 && !addForm.userId && (
                  <p className="text-xs text-muted-foreground">No users found</p>
                )}
                {addForm.userId && (
                  <p className="text-sm text-muted-foreground">
                    Selected: <span className="font-medium text-foreground">{searchQuery}</span>
                    {" "}<button type="button" className="text-xs text-primary hover:underline" onClick={() => { setAddForm({ ...addForm, userId: "" }); setSearchQuery(""); }}>Change</button>
                  </p>
                )}
              </div>
              <div className="grid gap-2">
                <Label>Role</Label>
                <Select value={addForm.role} onValueChange={(v) => setAddForm({ ...addForm, role: v })}>
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="member">Member</SelectItem>
                    <SelectItem value="admin">Admin</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setAddDialogOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={saving}>{saving ? "Adding..." : "Add member"}</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={!!removeTarget} onOpenChange={(open) => !open && setRemoveTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Remove member</DialogTitle>
            <DialogDescription>
              Remove <strong>{removeTarget ? removeTarget.userName || `User #${removeTarget.userId}` : ""}</strong> from this team?
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRemoveTarget(null)}>Cancel</Button>
            <Button variant="destructive" onClick={handleRemoveMember}>Remove</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
