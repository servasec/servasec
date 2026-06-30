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
      <div className="space-y-6">
        <Skeleton className="h-6 w-48" />
        <Card><div className="p-6 space-y-4"><Skeleton className="h-6 w-64" /><Skeleton className="h-4 w-full" /><Skeleton className="h-4 w-3/4" /></div></Card>
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

      <div className="rounded-lg border p-6">
        <h2 className="text-lg font-semibold mb-6">Edit policy</h2>
        <PolicyForm initialData={policy} onSubmit={handleSubmit} saving={saving} />
      </div>
    </div>
  );
}
