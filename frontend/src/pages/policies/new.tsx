import { useEffect } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { Skeleton } from "@/components/ui/skeleton";
import { PageHeader } from "@/components/page-header";
import { PolicyForm } from "@/components/policies/policy-form";
import axios from "@/lib/api";
import { toast } from "sonner";
import type { Policy } from "@/lib/types";

export default function NewPolicyPage() {
  const router = useRouter();
  const { loggedIn, authChecked } = useAuth();

  useEffect(() => {
    if (!authChecked) return;
    if (!loggedIn) { router.push("/login"); }
  }, [authChecked, loggedIn, router]);

  const handleSubmit = async (data: Partial<Policy>) => {
    try {
      await axios.post("/api/policies", data);
      toast.success("Policy created");
      router.push("/policies");
    } catch (error: any) {
      toast.error(error?.response?.data?.error || "Failed to create policy");
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

  return (
    <div className="space-y-6">
      <PageHeader crumbs={[
        { label: "Security" },
        { label: "Policies", href: "/policies" },
        { label: "New policy" },
      ]} />

      <div className="rounded-lg border p-6">
        <h2 className="text-lg font-semibold mb-6">Create policy</h2>
        <PolicyForm onSubmit={handleSubmit} saving={false} />
      </div>
    </div>
  );
}
