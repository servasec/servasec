import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Shield, ShieldCheck, Ban, ShieldAlert } from "lucide-react";
import { PageHeader } from "@/components/page-header";
import axios from "@/lib/api";
import { toast } from "sonner";

interface User {
  id: number;
  username: string;
  email: string;
  role: string;
  banned: boolean;
  oauthProvider?: string;
  createdAt: string;
}

export default function UsersPage() {
  const router = useRouter();
  const { loggedIn, user, authChecked } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    if (authChecked && loggedIn) {
      axios
        .get("/api/users")
        .then((res) => { setUsers(res.data); setLoadError(false); })
        .catch(() => { toast.error("Failed to load users"); setLoadError(true); })
        .finally(() => setLoading(false));
    }
  }, [authChecked, loggedIn]);

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
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b bg-muted/50">
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground">User</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden md:table-cell">Email</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden sm:table-cell">Role</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden lg:table-cell">Status</th>
                <th className="text-left px-4 py-3.5 font-medium text-muted-foreground hidden lg:table-cell">Joined</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                Array.from({ length: 3 }).map((_, i) => (
                  <tr key={i} className="border-b last:border-0">
                    {Array.from({ length: 5 }).map((_, j) => (
                      <td key={j} className="px-4 py-3">
                        <Skeleton className="h-5 w-full max-w-[120px]" />
                      </td>
                    ))}
                  </tr>
                ))
              ) : loadError ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-muted-foreground">
                    <ShieldAlert className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    <p>Failed to load users</p>
                    <p className="text-xs mt-1">The server may be unavailable</p>
                  </td>
                </tr>
              ) : users.length === 0 ? (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center text-muted-foreground">
                    No users found
                  </td>
                </tr>
              ) : (
                users.map((u) => (
                  <tr key={u.id} className="border-b last:border-0 hover:bg-muted/30 transition-colors">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-3">
                        <Avatar className="h-8 w-8">
                          <AvatarFallback className="bg-primary/10 text-primary text-xs">
                            {u.username.charAt(0).toUpperCase()}
                          </AvatarFallback>
                        </Avatar>
                        <span className="font-medium">
                          {u.username}
                          {u.oauthProvider && (
                            <span className="ml-2 inline-flex items-center gap-1 rounded-md border px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
                              {u.oauthProvider === "github" ? "GitHub" : u.oauthProvider === "gitlab" ? "GitLab" : u.oauthProvider}
                            </span>
                          )}
                        </span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground hidden md:table-cell">
                      {u.email}
                    </td>
                    <td className="px-4 py-3 hidden sm:table-cell">
                      <span className="inline-flex items-center gap-1.5 text-sm">
                        {u.role === "admin" ? (
                          <ShieldCheck className="h-3.5 w-3.5 text-primary" />
                        ) : (
                          <Shield className="h-3.5 w-3.5 text-muted-foreground" />
                        )}
                        <span className="capitalize">{u.role}</span>
                      </span>
                    </td>
                    <td className="px-4 py-3 hidden lg:table-cell">
                      {u.banned ? (
                        <span className="inline-flex items-center gap-1.5 text-destructive text-sm">
                          <Ban className="h-3.5 w-3.5" />
                          Banned
                        </span>
                      ) : (
                        <span className="inline-flex items-center gap-1.5 text-emerald-500 text-sm">
                          <span className="h-1.5 w-1.5 rounded-full bg-current" />
                          Active
                        </span>
                      )}
                    </td>
                    <td className="px-4 py-3 text-muted-foreground hidden lg:table-cell">
                      {u.createdAt ? new Date(u.createdAt).toLocaleDateString() : "-"}
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
