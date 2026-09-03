"use client"

import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { PageHeader } from "@/components/page-header";
import { ShieldAlert } from "lucide-react";
import axios from "@/lib/api";
import { toast } from "sonner";

const MASKED = "•••";

const KEYS: Record<string, string[]> = {
  github: ["sso_github_client_id", "sso_github_client_secret"],
  gitlab: ["sso_gitlab_client_id", "sso_gitlab_client_secret", "sso_gitlab_base_url"],
  oidc: ["sso_oidc_client_id", "sso_oidc_client_secret", "sso_oidc_issuer_url", "sso_oidc_scopes"],
};

const FIELD_META: Record<string, { label: string; secret: boolean; placeholder: string }> = {
  sso_github_client_id: { label: "Client ID", secret: false, placeholder: "GitHub OAuth app Client ID" },
  sso_github_client_secret: { label: "Client Secret", secret: true, placeholder: "GitHub OAuth app Client Secret" },
  sso_gitlab_client_id: { label: "Client ID", secret: false, placeholder: "GitLab application Client ID" },
  sso_gitlab_client_secret: { label: "Client Secret", secret: true, placeholder: "GitLab application Client Secret" },
  sso_gitlab_base_url: { label: "Base URL", secret: false, placeholder: "https://gitlab.com" },
  sso_oidc_client_id: { label: "Client ID", secret: false, placeholder: "OIDC Client ID" },
  sso_oidc_client_secret: { label: "Client Secret", secret: true, placeholder: "OIDC Client Secret" },
  sso_oidc_issuer_url: { label: "Issuer URL", secret: false, placeholder: "https://idp.example.com/realms/your-realm" },
  sso_oidc_scopes: { label: "Scopes", secret: false, placeholder: "openid profile email" },
};

interface SsoSettingsResponse {
  keys: string[];
  vals: Record<string, string>;
}

export default function SsoSettingsPage() {
  const router = useRouter();
  const { loggedIn, user, authChecked } = useAuth();
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState(false);
  const [saving, setSaving] = useState(false);
  const [vals, setVals] = useState<Record<string, string>>({});
  const [orig, setOrig] = useState<Record<string, string>>({});

  const fetchSSOSettings = async () => {
    setLoading(true);
    try {
      const res = await axios.get<SsoSettingsResponse>("/api/admin/settings/sso");
      setVals(res.data.vals);
      setOrig({ ...res.data.vals });
      setLoadError(false);
    } catch {
      setLoadError(true);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (!authChecked) return;
    if (!loggedIn) { router.push("/login"); return; }
    if (user?.role !== "admin") { router.push("/"); return; }
    fetchSSOSettings();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [authChecked, loggedIn, user, router]);

  const handleChange = (key: string, value: string) => {
    setVals((prev) => ({ ...prev, [key]: value }));
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await axios.put("/api/admin/settings/sso", { vals });
      toast.success("SSO settings saved");
      setOrig({ ...vals });
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to save SSO settings");
    } finally {
      setSaving(false);
    }
  };

  if (!authChecked || !loggedIn || user?.role !== "admin") {
    return (
      <div className="flex items-center justify-center h-64">
        <Skeleton className="h-8 w-48" />
      </div>
    );
  }

  if (loadError) {
    return (
      <div className="space-y-6">
        <PageHeader crumbs={[{ label: "Administration", href: "#" }, { label: "SSO / Single Sign-On" }]} />
        <Card>
          <CardContent className="py-12 text-center text-muted-foreground">
            <ShieldAlert className="h-8 w-8 mx-auto mb-2 opacity-40" />
            <p>Failed to load SSO settings</p>
            <p className="text-xs mt-1">The server may be unavailable</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader crumbs={[{ label: "Administration", href: "#" }, { label: "SSO / Single Sign-On" }]} />

      <Tabs defaultValue="github">
        <TabsList>
          <TabsTrigger value="github">GitHub</TabsTrigger>
          <TabsTrigger value="gitlab">GitLab</TabsTrigger>
          <TabsTrigger value="oidc">OIDC</TabsTrigger>
        </TabsList>

        {(["github", "gitlab", "oidc"] as const).map((provider) => (
          <TabsContent key={provider} value={provider}>
            <Card>
              <CardHeader>
                <CardDescription>
                  Configure single sign-on settings. Secret fields are stored encrypted and shown as {MASKED} when set.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                {loading ? (
                  Array.from({ length: KEYS[provider].length }).map((_, i) => (
                    <div key={i} className="space-y-2">
                      <Skeleton className="h-4 w-24" />
                      <Skeleton className="h-10 w-full" />
                    </div>
                  ))
                ) : (
                  KEYS[provider].map((key) => {
                    const meta = FIELD_META[key];
                    const isSecretPlaceholder = meta.secret && orig[key] && orig[key] !== "";
                    return (
                      <div key={key} className="space-y-2">
                        <Label htmlFor={key}>{meta.label}</Label>
                        <Input
                          id={key}
                          type={meta.secret ? "password" : "text"}
                          placeholder={isSecretPlaceholder ? MASKED : meta.placeholder}
                          value={vals[key] ?? ""}
                          onChange={(e) => handleChange(key, e.target.value)}
                          autoComplete="off"
                        />
                      </div>
                    );
                  })
                )}
                <div className="flex justify-end">
                  <Button onClick={handleSave} disabled={saving || loading}>
                    {saving ? "Saving..." : "Save changes"}
                  </Button>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        ))}
      </Tabs>
    </div>
  );
}