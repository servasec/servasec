import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { ChevronUp } from "lucide-react";

export function BackToTop() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const el = document.getElementById("main-scroll");
    if (!el) return;
    const onScroll = () => setVisible(el.scrollTop > 400);
    el.addEventListener("scroll", onScroll, { passive: true });
    return () => el.removeEventListener("scroll", onScroll);
  }, []);

  if (!visible) return null;

  return (
    <Button
      variant="secondary"
      size="icon"
      className="fixed bottom-6 right-6 h-10 w-10 rounded-full shadow-lg z-50"
      onClick={() => {
        const el = document.getElementById("main-scroll");
        if (el) el.scrollTo({ top: 0, behavior: "smooth" });
      }}
    >
      <ChevronUp className="h-5 w-5" />
    </Button>
  );
}
