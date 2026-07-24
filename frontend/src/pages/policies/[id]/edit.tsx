import { useEffect, useState } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Card } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { PolicyForm } from "@/components/policies/policy-form";
import axios from "@/lib/api";
import { toast } from "sonner";
import type { Policy } from "@/lib/types";

export default function EditPolicyPage() {
  const router = useRouter();
  const { id } = router.query;
  const { loggedIn, authChecked } = useAuth();
  const [policy, setPolicy] = useState<Policy | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (!authChecked) return;
    if (!loggedIn) { router.push("/login"); return; }
    if (!id) return;

    axios.get(`/api/policies/${id}`)
      .then((res) => setPolicy(res.data))
      .catch(() => { toast.error("Failed to load policy"); router.push("/policies"); })
      .finally(() => setLoading(false));
  }, [authChecked, loggedIn, id, router]);

  const handleSubmit = async (data: Partial<Policy>) => {
    setSaving(true);
    try {
      await axios.put(`/api/policies/${id}`, data);
      toast.success("Policy updated");
      router.push(`/policies/${id}`);
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to update policy");
      setSaving(false);
      throw error;
    }
  };

  if (!authChecked || !loggedIn) {
    return (
      <div className="flex items-center justify-center h-64">
        <Skeleton className="h-8 w-48" />
      </div>
    );
  }

  if (loading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-5 w-48 animate-pulse" />
        <Card><div className="p-4 space-y-3"><Skeleton className="h-5 w-64 animate-pulse" /><Skeleton className="h-3.5 w-full animate-pulse" /><Skeleton className="h-3.5 w-3/4 animate-pulse" /></div></Card>
      </div>
    );
  }

  if (!policy) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        <p>Policy not found</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <PageHeader crumbs={[
        { label: "Security" },
        { label: "Policies", href: "/policies" },
        { label: policy.name, href: `/policies/${policy.id}` },
        { label: "Edit" },
      ]} />

      <div className="rounded-lg border p-4">
        <h2 className="text-sm font-semibold mb-4">Edit policy</h2>
        <PolicyForm initialData={policy} onSubmit={handleSubmit} saving={saving} />
      </div>
    </div>
  );
}
