import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";

import Link from "next/link";
import axios from "@/lib/api";
import { toast } from "sonner";
import { AuthLayout } from "@/components/auth-layout";

const LoginPage = () => {
  const router = useRouter();
  const { login, loggedIn, authChecked } = useAuth();

  const [form, setForm] = useState({ username: "", password: "" });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (authChecked && loggedIn) {
      router.replace("/");
    }
  }, [authChecked, loggedIn, router]);

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
    return null;
  }

  return (
    <AuthLayout>
      <div className="rounded-xl border bg-card text-card-foreground shadow-sm">
        <div className="p-8">
          <form onSubmit={handleLogin}>
            <div className="flex flex-col gap-6">
              <div className="flex flex-col items-center gap-3 text-center">
                <div>
                  <h1 className="text-xl font-bold tracking-tight">servasec</h1>
                  <p className="text-sm text-muted-foreground mt-0.5">
                    Sign in to your account
                  </p>
                </div>
              </div>

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

              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? "Signing in..." : "Sign in"}
              </Button>

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
          </form>
        </div>
      </div>
    </AuthLayout>
  );
};

export default LoginPage;
