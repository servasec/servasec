"use client"

import { useRouter } from "next/router"
import { useAuth } from "@/context/AuthContext"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { BackToTop } from "@/components/back-to-top"
import { useTheme } from "next-themes"
import {
  AppWindow,
  FolderKanban,
  Users,
  UsersRound,
  Shield,
  Sun,
  Moon,
  ChevronLeft,
  ChevronDown,
  User,
  LogOut,
  Scan,
  Bug,
  ScrollText,
  ScanLine,
  ShieldAlert,
} from "lucide-react"
import { useState, useEffect, useRef } from "react"
import Link from "next/link"
import type { User as UserProfile } from "@/lib/types"

interface NavItem {
  label: string
  href: string
  icon: React.ReactNode
  adminOnly?: boolean
  feature?: string
}

interface NavGroup {
  label: string
  items: NavItem[]
}

const navGroups: NavGroup[] = [
    {
      label: "Security",
      items: [
        { label: "Applications", href: "/applications", icon: <AppWindow className="h-4 w-4" /> },
        { label: "Groups", href: "/groups", icon: <FolderKanban className="h-4 w-4" /> },
        { label: "Scans", href: "/scans", icon: <Scan className="h-4 w-4" /> },
        { label: "Findings", href: "/findings", icon: <Bug className="h-4 w-4" /> },
        { label: "Policies", href: "/policies", icon: <ShieldAlert className="h-4 w-4" /> },
      ],
    },
    {
      label: "Administration",
      items: [
        { label: "Users", href: "/users", icon: <Users className="h-4 w-4" />, adminOnly: true },
        { label: "Teams", href: "/teams", icon: <UsersRound className="h-4 w-4" /> },
        { label: "Permissions", href: "/admin/permissions", icon: <Shield className="h-4 w-4" />, adminOnly: true },
        { label: "Scanner Types", href: "/admin/scanner-types", icon: <ScanLine className="h-4 w-4" />, adminOnly: true },
        { label: "Audit Log", href: "/admin/audit-log", icon: <ScrollText className="h-4 w-4" />, adminOnly: true, feature: "audit_log" },
      ],
    },
]

function UserMenu({
  user,
  onLogout,
  collapsed,
}: {
  user: UserProfile | null
  onLogout: () => void
  collapsed: boolean
}) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener("mousedown", handleClick)
    return () => document.removeEventListener("mousedown", handleClick)
  }, [])

  if (collapsed) {
    return (
      <Link
        href="/profile"
        className="flex items-center justify-center w-full rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground"
        title="Profile"
      >
        <Avatar className="h-7 w-7 shrink-0">
          <AvatarFallback className="bg-accent text-accent-foreground text-[11px]">
            {user?.username?.charAt(0).toUpperCase() || "U"}
          </AvatarFallback>
        </Avatar>
      </Link>
    )
  }

  return (
    <div ref={ref} className="relative">
      <button
        onClick={() => setOpen(!open)}
        className={`flex w-full items-center gap-3 rounded-md px-3 py-2 text-sm text-muted-foreground transition-colors hover:bg-accent hover:text-accent-foreground ${
          collapsed ? "justify-center" : ""
        }`}
      >
        <Avatar className="h-7 w-7 shrink-0">
          <AvatarFallback className="bg-accent text-accent-foreground text-[11px]">
            {user?.username?.charAt(0).toUpperCase() || "U"}
          </AvatarFallback>
        </Avatar>
        {!collapsed && (
          <>
            <div className="flex-1 min-w-0 text-left">
              <p className="text-sm font-medium truncate text-foreground">
                {user?.username}
              </p>
              <p className="text-[11px] text-muted-foreground/60 capitalize truncate">
                {user?.role}
              </p>
            </div>
            <ChevronDown
              className={`h-3.5 w-3.5 shrink-0 text-muted-foreground/50 transition-transform ${
                open ? "rotate-180" : ""
              }`}
            />
          </>
        )}
      </button>

      {open && (
        <div className="absolute bottom-full left-0 right-0 mb-1 rounded-md border bg-popover text-popover-foreground py-1 shadow-lg">
          <div className="px-3 py-2 text-xs text-muted-foreground border-b">
            Signed in as
          </div>
          <Link
            href="/profile"
            onClick={() => setOpen(false)}
            className="flex w-full items-center gap-2 px-3 py-2 text-sm text-foreground/80 hover:bg-accent transition-colors"
          >
            <User className="h-3.5 w-3.5" />
            Profile
          </Link>
          <button
            onClick={onLogout}
            className="flex w-full items-center gap-2 px-3 py-2 text-sm text-foreground/80 hover:bg-accent transition-colors"
          >
            <LogOut className="h-3.5 w-3.5" />
            Sign out
          </button>
        </div>
      )}
    </div>
  )
}

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter()
  const { loggedIn, user, logout, authChecked } = useAuth()
  const { theme, setTheme } = useTheme()
  const [collapsed, setCollapsed] = useState(false)
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  useEffect(() => {
    const saved = localStorage.getItem("servasec-sidebar-collapsed")
    if (saved === "true") {
      setCollapsed(true)
    }
  }, [])

  useEffect(() => {
    if (typeof window !== "undefined") {
      localStorage.setItem("servasec-sidebar-collapsed", String(collapsed))
    }
  }, [collapsed])

  const isPublicPage =
    router.pathname === "/login" || router.pathname === "/register"

  if (!authChecked || !loggedIn || isPublicPage) {
    return <>{children}</>
  }

  const handleLogout = async () => {
    await logout(true)
  }

  const isActive = (href: string) => {
    if (href === "/") return router.pathname === "/"
    return router.pathname.startsWith(href)
  }

  return (
    <div className="h-screen overflow-hidden bg-background flex">
      <aside
        className={`flex flex-col bg-sidebar shrink-0 transition-all duration-200 ${
          collapsed ? "w-16" : "w-64"
        }`}
      >
        <Link href="/" className="h-14 flex items-center gap-3 px-4 shrink-0 hover:opacity-80 transition-opacity">
          <img
            src="/assets/servasec-mark.svg"
            alt="servasec"
            className="h-6 w-6 shrink-0"
          />
          {!collapsed && (
            <span className="text-sm font-bold tracking-tight text-sidebar-foreground">
              servasec
            </span>
          )}
        </Link>

        <nav className="flex-1 overflow-y-auto py-3 space-y-1">
          {navGroups.map((group) => {
            const visibleItems = group.items.filter((item) => {
              if (item.adminOnly && user?.role !== "admin") return false
              if (item.feature && !user?.features?.includes(item.feature)) return false
              return true
            })
            if (visibleItems.length === 0) return null

            return (
              <div key={group.label}>
                {!collapsed && (
                  <div className="px-4 py-1.5">
                    <span className="text-xs font-semibold text-sidebar-foreground/60">
                      {group.label}
                    </span>
                  </div>
                )}
                {visibleItems.map((item) => {
                  const active = isActive(item.href)
                  return (
                    <Link
                      key={item.href}
                      href={item.href}
                      className={`relative flex items-center h-9 text-sm transition-colors ${
                        collapsed ? "justify-center" : "gap-3 px-4"
                      } ${
                        active
                          ? "bg-sidebar-accent text-sidebar-accent-foreground font-medium"
                          : "text-sidebar-foreground/85 hover:bg-sidebar-accent hover:text-sidebar-foreground"
                      }`}
                      title={collapsed ? item.label : undefined}
                    >
                      {active && (
                        <span className="absolute left-0 top-1/2 -translate-y-1/2 w-[3px] h-5 rounded-r-full bg-primary" />
                      )}
                      <span className="shrink-0">{item.icon}</span>
                      {!collapsed && <span>{item.label}</span>}
                    </Link>
                  )
                })}
              </div>
            )
          })}
        </nav>

        <div className="pt-1 pb-2 px-2 space-y-0.5 shrink-0">
          {mounted && (
            <button
              onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
              className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm text-sidebar-foreground/85 hover:bg-sidebar-accent hover:text-sidebar-foreground w-full transition-colors ${
                collapsed ? "justify-center" : ""
              }`}
            >
              {theme === "dark" || theme === "catppuccin" || theme === "atom-one" || theme === "nord" ? (
                <Sun className="h-4 w-4 shrink-0" />
              ) : (
                <Moon className="h-4 w-4 shrink-0" />
              )}
              {!collapsed && (
                <span>{theme === "light" ? "Dark mode" : "Light mode"}</span>
              )}
            </button>
          )}

          <UserMenu user={user} onLogout={handleLogout} collapsed={collapsed} />

          <button
            onClick={() => setCollapsed(!collapsed)}
            className={`flex items-center gap-3 px-3 py-2 rounded-md text-sm text-sidebar-foreground/70 hover:bg-sidebar-accent hover:text-sidebar-foreground w-full transition-colors ${
              collapsed ? "justify-center" : ""
            }`}
          >
            <ChevronLeft
              className={`h-4 w-4 shrink-0 transition-transform ${
                collapsed ? "rotate-180" : ""
              }`}
            />
            {!collapsed && <span>Collapse sidebar</span>}
          </button>
        </div>
      </aside>

      <main className="flex-1 flex flex-col min-w-0 min-h-0 pt-6 lg:pt-8 pr-6 lg:pr-8 pb-6 lg:pb-8">
        <div id="main-scroll" className="bg-card rounded-xl border shadow-sm flex-1 min-h-0 overflow-y-auto p-6 lg:p-8">
          {children}
        </div>
        <BackToTop />
      </main>
    </div>
  )
}
