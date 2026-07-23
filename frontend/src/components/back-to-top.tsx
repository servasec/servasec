import { useEffect, useState } from "react"
import { Button } from "@/components/ui/button"
import { ChevronUp } from "lucide-react"

export function BackToTop() {
  const [visible, setVisible] = useState(false)

  useEffect(() => {
    const el = document.getElementById("main-scroll")
    if (!el) return
    const onScroll = () => setVisible(el.scrollTop > 400)
    el.addEventListener("scroll", onScroll, { passive: true })
    return () => el.removeEventListener("scroll", onScroll)
  }, [])

  return (
    <div
      className={`fixed bottom-6 right-6 z-50 transition-all duration-200 ${
        visible ? "opacity-100 translate-y-0" : "opacity-0 translate-y-2 pointer-events-none"
      }`}
    >
      <Button
        variant="secondary"
        size="icon"
        className="h-10 w-10 rounded-full shadow-lg"
        onClick={() => {
          const el = document.getElementById("main-scroll")
          if (el) el.scrollTo({ top: 0, behavior: "smooth" })
        }}
      >
        <ChevronUp className="h-5 w-5" />
      </Button>
    </div>
  )
}
