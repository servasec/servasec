import { useState, useEffect } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import axios from "@/lib/api";
import { toast } from "sonner";
import { CheckCircle } from "lucide-react";
import { AuthLayout } from "@/components/auth-layout";

const RegisterPage = () => {
  const router = useRouter();
  const { loggedIn, authChecked } = useAuth();

  const [form, setForm] = useState({ username: "", email: "", password: "" });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (authChecked && loggedIn) {
      router.replace("/");
    }
  }, [authChecked, loggedIn, router]);

  const onChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm({ ...form, [e.target.name]: e.target.value });
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    try {
      await axios.post("/api/register", form);
      toast.success("Account created successfully!", {
        icon: <CheckCircle className="w-4 h-4" />,
      });
      router.push("/login");
    } catch (error: any) {
      const errMsg = error?.response?.data?.error || "An error has occurred";
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
          <form onSubmit={handleRegister}>
            <div className="flex flex-col gap-6">
              <div className="flex flex-col items-center gap-3 text-center">
                <div>
                  <h1 className="text-xl font-bold tracking-tight">servasec</h1>
                  <p className="text-sm text-muted-foreground mt-0.5">
                    Create your account
                  </p>
                </div>
              </div>

              <div className="grid gap-5">
                <div className="grid gap-2">
                  <Label htmlFor="username" className="text-sm font-medium">
                    Username
                  </Label>
                  <Input
                    id="username"
                    name="username"
                    value={form.username}
                    onChange={onChange}
                    placeholder="username"
                    autoComplete="username"
                    required
                    maxLength={32}
                  />
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="email" className="text-sm font-medium">
                    Email
                  </Label>
                  <Input
                    id="email"
                    name="email"
                    type="email"
                    value={form.email}
                    onChange={onChange}
                    placeholder="you@example.com"
                    autoComplete="email"
                    required
                    maxLength={254}
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
                    placeholder="Min 8 characters"
                    autoComplete="new-password"
                    required
                    minLength={8}
                    maxLength={72}
                  />
                </div>
              </div>

              <Button type="submit" className="w-full" disabled={loading}>
                {loading ? "Creating account..." : "Register"}
              </Button>

              <p className="text-center text-sm text-muted-foreground">
                Already have an account?{" "}
                <Link
                  href="/login"
                  className="font-medium text-primary hover:text-primary/80 underline-offset-4 hover:underline"
                >
                  Sign in
                </Link>
              </p>
            </div>
          </form>
        </div>
      </div>
    </AuthLayout>
  );
};

export default RegisterPage;
