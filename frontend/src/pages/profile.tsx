import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { User, Mail, Lock, Save, Key, Plus, Trash2, Copy, Check, Palette } from "lucide-react";
import { PageHeader } from "@/components/page-header";
import axios from "@/lib/api";
import { toast } from "sonner";
import { useTheme } from "next-themes";
import { THEMES } from "@/lib/types";
import type { ApiKey, CreateApiKeyResponse } from "@/lib/types";

export default function ProfilePage() {
  const router = useRouter();
  const { loggedIn, user: authUser, authChecked, checkAuth } = useAuth();
  const { theme, setTheme } = useTheme();
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [changingPassword, setChangingPassword] = useState(false);

  const [profile, setProfile] = useState({ username: "", email: "" });
  const [passwordForm, setPasswordForm] = useState({
    currentPassword: "",
    newPassword: "",
    confirmPassword: "",
  });

  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [keysLoading, setKeysLoading] = useState(true);
  const [genDialogOpen, setGenDialogOpen] = useState(false);
  const [genName, setGenName] = useState("");
  const [generating, setGenerating] = useState(false);
  const [createdKey, setCreatedKey] = useState<CreateApiKeyResponse | null>(null);
  const [copied, setCopied] = useState(false);
  const [revokeTarget, setRevokeTarget] = useState<ApiKey | null>(null);
  const [revoking, setRevoking] = useState(false);

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn) {
      axios
        .get("/api/me")
        .then((res) => {
          setProfile({
            username: res.data.username || "",
            email: res.data.email || "",
          });
        })
        .catch(() => {})
        .finally(() => setLoading(false));
    }
  }, [authChecked, loggedIn]);

  const handleProfileUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      await axios.patch("/api/me", profile);
      toast.success("Profile updated");
      checkAuth();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to update profile");
    } finally {
      setSaving(false);
    }
  };

  const handlePasswordChange = async (e: React.FormEvent) => {
    e.preventDefault();
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      toast.error("Passwords do not match");
      return;
    }
    setChangingPassword(true);
    try {
      await axios.put("/api/me/password", {
        currentPassword: passwordForm.currentPassword,
        newPassword: passwordForm.newPassword,
      });
      toast.success("Password changed");
      setPasswordForm({
        currentPassword: "",
        newPassword: "",
        confirmPassword: "",
      });
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to change password");
    } finally {
      setChangingPassword(false);
    }
  };

  const handleThemeChange = (newTheme: string) => {
    setTheme(newTheme);
  };

  const fetchKeys = () => {
    setKeysLoading(true);
    axios.get("/api/api-keys")
      .then((res) => setApiKeys(res.data || []))
      .catch(() => toast.error("Failed to load API keys"))
      .finally(() => setKeysLoading(false));
  };

  useEffect(() => {
    if (authChecked && loggedIn) fetchKeys();
  }, [authChecked, loggedIn]);

  const handleGenerate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!genName.trim()) return;
    setGenerating(true);
    try {
      const res = await axios.post("/api/api-keys", { name: genName.trim() });
      setCreatedKey(res.data);
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to generate key");
    } finally {
      setGenerating(false);
    }
  };

  const handleCopy = async () => {
    if (!createdKey) return;
    try {
      await navigator.clipboard.writeText(createdKey.key);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      toast.error("Failed to copy");
    }
  };

  const handleCloseGenDialog = () => {
    if (createdKey) {
      setGenDialogOpen(false);
      setCreatedKey(null);
      setGenName("");
      setCopied(false);
      fetchKeys();
    } else {
      setGenDialogOpen(false);
      setGenName("");
    }
  };

  const handleRevoke = async () => {
    if (!revokeTarget) return;
    setRevoking(true);
    try {
      await axios.delete(`/api/api-keys/${revokeTarget.id}`);
      toast.success("API key revoked");
      setRevokeTarget(null);
      fetchKeys();
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to revoke key");
    } finally {
      setRevoking(false);
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
      <PageHeader crumbs={[{ label: "Profile" }]} />

      <Card>
        <CardHeader>
          <div className="flex items-center gap-4">
            <Avatar className="h-14 w-14">
              <AvatarFallback className="bg-primary/10 text-primary text-lg">
                {profile.username.charAt(0).toUpperCase() || "U"}
              </AvatarFallback>
            </Avatar>
            <div>
              <CardTitle>Profile</CardTitle>
              <p className="text-sm text-muted-foreground mt-0.5">
                Manage your personal information
              </p>
              {authUser?.oauthProvider && (
                <span className="inline-flex items-center gap-1 text-xs text-muted-foreground mt-1">
                  Connected via {authUser.oauthProvider === "github" ? "GitHub" : authUser.oauthProvider === "gitlab" ? "GitLab" : authUser.oauthProvider}
                </span>
              )}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="space-y-4">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : (
            <form onSubmit={handleProfileUpdate} className="space-y-5">
              <div className="grid gap-2">
                <Label htmlFor="username" className="text-sm font-medium">
                  Username
                </Label>
                <div className="relative">
                  <User className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="username"
                    name="username"
                    value={profile.username}
                    onChange={(e) =>
                      setProfile({ ...profile, username: e.target.value })
                    }
                    className="pl-9"
                    required
                  />
                </div>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="email" className="text-sm font-medium">
                  Email
                </Label>
                <div className="relative">
                  <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="email"
                    name="email"
                    type="email"
                    value={profile.email}
                    onChange={(e) =>
                      setProfile({ ...profile, email: e.target.value })
                    }
                    className="pl-9"
                    required
                    disabled={!!authUser?.oauthProvider}
                  />
                </div>
                {authUser?.oauthProvider && (
                  <p className="text-xs text-muted-foreground">
                    Email managed by {authUser.oauthProvider === "github" ? "GitHub" : authUser.oauthProvider === "gitlab" ? "GitLab" : "your SSO provider"}
                  </p>
                )}
              </div>

              <Button type="submit" disabled={saving} className="gap-2">
                <Save className="h-4 w-4" />
                {saving ? "Saving..." : "Save changes"}
              </Button>
            </form>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-3">
            <Palette className="h-5 w-5 text-muted-foreground" />
            <div>
              <CardTitle>Theme</CardTitle>
              <p className="text-sm text-muted-foreground mt-0.5">
                Choose your preferred color scheme
              </p>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
            {THEMES.map((t) => {
                  const active = theme === t.name;
              return (
                <button
                  key={t.name}
                  onClick={() => handleThemeChange(t.name)}
                  className={`relative flex flex-col items-center gap-2.5 rounded-lg border-2 p-4 transition-all hover:border-primary/50 ${
                    active
                      ? "border-primary bg-primary/5"
                      : "border-border"
                  }`}
                >
                  <div className="flex gap-1.5">
                    <span
                      className="h-4 w-4 rounded-full border"
                      style={{
                        backgroundColor: t.dark ? "#1e1e2e" : "#f5f5f5",
                        borderColor: t.dark ? "#313244" : "#e5e5e5",
                      }}
                    />
                    <span
                      className="h-4 w-4 rounded-full border"
                      style={{
                        backgroundColor: t.dark ? "#cdd6f4" : "#1e1e2e",
                        borderColor: t.dark ? "#45475a" : "#313244",
                      }}
                    />
                    <span
                      className="h-4 w-4 rounded-full border"
                      style={{
                        backgroundColor:
                          t.name === "catppuccin" ? "#cba6f7" :
                          t.name === "atom-one" ? "#a6e22e" :
                          t.name === "nord" ? "#88c0d0" :
                          t.name === "dark" ? "#a298ab" :
                          "#7c3aed",
                        borderColor: "transparent",
                      }}
                    />
                  </div>
                  <span className="text-xs font-medium text-foreground">
                    {t.label}
                  </span>
                  {active && (
                    <span className="absolute top-2 right-2 h-2 w-2 rounded-full bg-primary" />
                  )}
                </button>
              );
            })}
          </div>
        </CardContent>
      </Card>

      {!authUser?.oauthProvider && (
        <Card>
          <CardHeader>
            <CardTitle>Change password</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handlePasswordChange} className="space-y-5">
              <div className="grid gap-2">
                <Label htmlFor="currentPassword" className="text-sm font-medium">
                  Current password
                </Label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="currentPassword"
                    name="currentPassword"
                    type="password"
                    value={passwordForm.currentPassword}
                    onChange={(e) =>
                      setPasswordForm({
                        ...passwordForm,
                        currentPassword: e.target.value,
                      })
                    }
                    className="pl-9"
                    required
                  />
                </div>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="newPassword" className="text-sm font-medium">
                  New password
                </Label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="newPassword"
                    name="newPassword"
                    type="password"
                    value={passwordForm.newPassword}
                    onChange={(e) =>
                      setPasswordForm({
                        ...passwordForm,
                        newPassword: e.target.value,
                      })
                    }
                    className="pl-9"
                    placeholder="Min 8 characters"
                    required
                    minLength={8}
                  />
                </div>
              </div>

              <div className="grid gap-2">
                <Label htmlFor="confirmPassword" className="text-sm font-medium">
                  Confirm new password
                </Label>
                <div className="relative">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                  <Input
                    id="confirmPassword"
                    name="confirmPassword"
                    type="password"
                    value={passwordForm.confirmPassword}
                    onChange={(e) =>
                      setPasswordForm({
                        ...passwordForm,
                        confirmPassword: e.target.value,
                      })
                    }
                    className="pl-9"
                    placeholder="Repeat your new password"
                    required
                  />
                </div>
              </div>

              <Button
                type="submit"
                disabled={changingPassword}
                variant="outline"
                className="gap-2"
              >
                <Lock className="h-4 w-4" />
                {changingPassword ? "Updating..." : "Change password"}
              </Button>
            </form>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <CardTitle>API Keys</CardTitle>
            <Button
              variant="outline"
              size="sm"
              className="h-8 px-2.5 text-xs gap-1"
              onClick={() => { setGenName(""); setCreatedKey(null); setCopied(false); setGenDialogOpen(true); }}
            >
              <Plus className="h-3.5 w-3.5" />
              Generate new key
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {keysLoading ? (
            <div className="space-y-3">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-full" />
            </div>
          ) : apiKeys.length === 0 ? (
            <div className="text-center py-8 text-muted-foreground">
              <Key className="h-8 w-8 mx-auto mb-2 opacity-40" />
              <p>No API keys yet</p>
              <p className="text-xs mt-1">Generate a key to use in your CI/CD pipelines</p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left pb-3 font-medium text-muted-foreground">Prefix</th>
                    <th className="text-left pb-3 font-medium text-muted-foreground">Name</th>
                    <th className="text-left pb-3 font-medium text-muted-foreground hidden sm:table-cell">Created</th>
                    <th className="text-left pb-3 font-medium text-muted-foreground hidden sm:table-cell">Last Used</th>
                    <th className="w-16 pb-3" />
                  </tr>
                </thead>
                <tbody>
                  {apiKeys.map((key) => (
                    <tr key={key.id} className="border-b last:border-0">
                      <td className="py-3 pr-4">
                        <code className="text-xs bg-muted rounded px-1.5 py-0.5 font-mono">{key.keyPrefix}...</code>
                      </td>
                      <td className="py-3 pr-4 font-medium">{key.name}</td>
                      <td className="py-3 pr-4 text-muted-foreground hidden sm:table-cell">
                        {new Date(key.createdAt).toLocaleDateString()}
                      </td>
                      <td className="py-3 pr-4 text-muted-foreground hidden sm:table-cell">
                        {key.lastUsedAt ? new Date(key.lastUsedAt).toLocaleDateString() : "never"}
                      </td>
                      <td className="py-3">
                        <Button
                          variant="destructive-ghost"
                          size="icon"
                          className="h-8 w-8"
                          onClick={() => setRevokeTarget(key)}
                          title="Revoke key"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      <Dialog open={genDialogOpen} onOpenChange={handleCloseGenDialog}>
        <DialogContent>
          {createdKey ? (
            <>
              <DialogHeader>
                <DialogTitle>API Key Generated</DialogTitle>
                <DialogDescription>
                  Copy this key now. You won&apos;t be able to see it again.
                </DialogDescription>
              </DialogHeader>
              <div className="space-y-4 py-4">
                <div className="p-3 rounded-md bg-amber-500/10 border border-amber-500/30 text-amber-600 dark:text-amber-400 text-xs">
                  ⚠ Copy this key now. It will never be shown again.
                </div>
                <div className="grid gap-2">
                  <Label>Name</Label>
                  <Input value={createdKey.name} readOnly />
                </div>
                <div className="grid gap-2">
                  <Label>Key</Label>
                  <div className="flex gap-2">
                    <Input
                      value={createdKey.key}
                      readOnly
                      className="font-mono text-xs"
                    />
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      className="h-10 w-10 shrink-0"
                      onClick={handleCopy}
                    >
                      {copied ? <Check className="h-4 w-4 text-emerald-500" /> : <Copy className="h-4 w-4" />}
                    </Button>
                  </div>
                </div>
              </div>
              <DialogFooter>
                <Button type="button" onClick={handleCloseGenDialog}>Done</Button>
              </DialogFooter>
            </>
          ) : (
            <>
              <DialogHeader>
                <DialogTitle>Generate new API key</DialogTitle>
                <DialogDescription>
                  Create a key for use in your CI/CD pipelines or automation scripts.
                </DialogDescription>
              </DialogHeader>
              <form onSubmit={handleGenerate}>
                <div className="py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="keyName">Key name *</Label>
                    <Input
                      id="keyName"
                      value={genName}
                      onChange={(e) => setGenName(e.target.value)}
                      placeholder="e.g. ci-cd, deploy"
                      required
                      maxLength={100}
                      autoComplete="off"
                      data-1p-ignore
                    />
                  </div>
                </div>
                <DialogFooter>
                  <Button type="button" variant="outline" onClick={() => setGenDialogOpen(false)}>Cancel</Button>
                  <Button type="submit" disabled={generating || !genName.trim()}>
                    {generating ? "Generating..." : "Generate"}
                  </Button>
                </DialogFooter>
              </form>
            </>
          )}
        </DialogContent>
      </Dialog>

      <Dialog open={!!revokeTarget} onOpenChange={(open) => !open && setRevokeTarget(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Revoke API Key</DialogTitle>
            <DialogDescription>
              Are you sure? This action is irreversible. Any services using this key will immediately lose access.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => setRevokeTarget(null)}>Cancel</Button>
            <Button type="button" variant="destructive" onClick={handleRevoke} disabled={revoking}>
              {revoking ? "Revoking..." : "Revoke"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
