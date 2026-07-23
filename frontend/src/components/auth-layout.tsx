"use client"

import { useRouter } from "next/router"
import { useTheme } from "next-themes"
import { useState, useEffect } from "react"
import { Sun, Moon } from "lucide-react"
import { AuthBackground } from "./auth-background"

export function AuthLayout({ children }: { children: React.ReactNode }) {
  const { theme, setTheme } = useTheme()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  return (
    <div className="flex min-h-[100dvh] bg-background">
      <div className="relative flex w-full lg:w-[30%] items-center justify-center p-8 bg-card">
        {mounted && (
          <button
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
            className="absolute bottom-4 left-4 flex h-9 w-9 items-center justify-center rounded-lg border bg-card text-muted-foreground hover:text-foreground transition-colors shadow-sm"
          >
            {theme === "dark" || theme === "catppuccin" || theme === "github" || theme === "nord" ? (
              <Sun className="h-4 w-4" />
            ) : (
              <Moon className="h-4 w-4" />
            )}
          </button>
        )}

        <div className="w-full max-w-sm">{children}</div>
      </div>

      <div className="hidden lg:flex w-[70%]">
        <AuthBackground />
      </div>
    </div>
  )
}
