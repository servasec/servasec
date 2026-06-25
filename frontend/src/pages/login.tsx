import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import { Github, Gitlab, LogIn } from "lucide-react";

import Link from "next/link";
import axios from "@/lib/api";
import { toast } from "sonner";
import { AuthLayout } from "@/components/auth-layout";

interface SSOProvider {
  name: string;
  url: string;
}

const providerLabel: Record<string, string> = {
  github: "GitHub",
  gitlab: "GitLab",
  oidc: "OIDC",
};

const LoginPage = () => {
  const router = useRouter();
  const { login, loggedIn, authChecked } = useAuth();

  const [form, setForm] = useState({ username: "", password: "" });
  const [loading, setLoading] = useState(false);
  const [ssoProviders, setSSOProviders] = useState<SSOProvider[]>([]);

  useEffect(() => {
    if (authChecked && loggedIn) {
      router.replace("/");
    }
  }, [authChecked, loggedIn, router]);

  useEffect(() => {
    axios.get("/api/auth/providers").then((res) => {
      setSSOProviders(res.data.providers || []);
    }).catch(() => {
      // SSO not configured
    });
  }, []);

  const onChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const res = await axios.post("/api/login", form);
      login(res.data.user);
      router.push("/");
    } catch (error: any) {
      const errMsg = error?.response?.data?.error || "Error during login";
      toast.error(errMsg);
    } finally {
      setLoading(false);
    }
  };

  if (authChecked && loggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-muted-foreground">Redirecting...</div>
      </div>
    );
  }

  return (
    <AuthLayout>
      <div className="rounded-xl border bg-card text-card-foreground shadow-sm">
        <div className="p-8">
          <div className="flex flex-col gap-6">
            <div className="flex flex-col items-center gap-3 text-center">
              <div>
                <h1 className="text-xl font-bold tracking-tight">servasec</h1>
                <p className="text-sm text-muted-foreground mt-0.5">
                  Sign in to your account
                </p>
              </div>
            </div>

            {ssoProviders.length > 0 && (
              <>
                <div className="grid gap-2">
                  {ssoProviders.map((p) => {
                    const Icon = p.name === "github" ? Github : p.name === "gitlab" ? Gitlab : LogIn;
                    return (
                      <Button
                        key={p.name}
                        variant="outline"
                        className="w-full gap-2"
                        onClick={() => { window.location.href = p.url; }}
                      >
                        <Icon className="h-4 w-4" />
                        Sign in with {providerLabel[p.name] || p.name}
                      </Button>
                    );
                  })}
                </div>

                <div className="relative">
                  <div className="absolute inset-0 flex items-center">
                    <span className="w-full border-t" />
                  </div>
                  <div className="relative flex justify-center text-xs uppercase">
                    <span className="bg-card px-2 text-muted-foreground">
                      Or continue with
                    </span>
                  </div>
                </div>
              </>
            )}

            <form onSubmit={handleLogin}>
              <div className="grid gap-5">
                <div className="grid gap-2">
                  <Label htmlFor="username" className="text-sm font-medium">
                    Username or email
                  </Label>
                  <Input
                    id="username"
                    name="username"
                    value={form.username}
                    onChange={onChange}
                    placeholder="you@example.com"
                    autoComplete="username"
                    required
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="password" className="text-sm font-medium">
                    Password
                  </Label>
                  <Input
                    id="password"
                    name="password"
                    type="password"
                    value={form.password}
                    onChange={onChange}
                    placeholder="Enter your password"
                    autoComplete="current-password"
                    required
                  />
                </div>
              </div>

              <Button type="submit" className="w-full mt-5" disabled={loading}>
                {loading ? "Signing in..." : "Sign in"}
              </Button>
            </form>

            <p className="text-center text-sm text-muted-foreground">
              Don&apos;t have an account?{" "}
              <Link
                href="/register"
                className="font-medium text-primary hover:text-primary/80 underline-offset-4 hover:underline"
              >
                Register
              </Link>
            </p>
          </div>
        </div>
      </div>
    </AuthLayout>
  );
};

export default LoginPage;
