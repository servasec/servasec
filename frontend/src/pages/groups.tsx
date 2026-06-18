import { useEffect } from "react";
import { useRouter } from "next/router";
import { useAuth } from "@/context/AuthContext";
import { PageHeader } from "@/components/page-header";

export default function GroupsPage() {
  const router = useRouter();
  const { loggedIn, authChecked } = useAuth();

  useEffect(() => {
    if (authChecked && !loggedIn) {
      router.replace("/login");
    }
  }, [authChecked, loggedIn, router]);

  if (!authChecked || !loggedIn) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-pulse text-muted-foreground">Loading...</div>
      </div>
    );
  }

  return (
    <div>
      <PageHeader crumbs={[{ label: "Groups" }]} />
    </div>
  );
}
