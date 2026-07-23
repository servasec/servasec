"use client"

import { useRouter } from "next/router"
import { useEffect, useState, useCallback } from "react"
import { Command } from "cmdk"
import { useTheme } from "next-themes"
import {
  Bug, Scan, AppWindow, Users, Shield, ShieldAlert,
  FolderKanban, Settings, ScrollText, ScanLine, LogOut,
  Sun, Moon, Search,
} from "lucide-react"
import { useAuth } from "@/context/AuthContext"

interface CommandItem {
  label: string
  href?: string
  icon: React.ReactNode
  group: string
  action?: () => void
  shortcut?: string
}

export function CommandPalette() {
  const router = useRouter()
  const { user, logout } = useAuth()
  const { theme, setTheme } = useTheme()
  const [open, setOpen] = useState(false)
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  const toggleTheme = useCallback(() => {
    setTheme(theme === "dark" || theme === "catppuccin" || theme === "atom-one" || theme === "nord" ? "light" : "dark")
  }, [theme, setTheme])

  const runCommand = useCallback((action?: () => void, href?: string) => {
    setOpen(false)
    if (action) action()
    else if (href) router.push(href)
  }, [router])

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault()
        setOpen((prev) => !prev)
      }
    }
    document.addEventListener("keydown", down)
    return () => document.removeEventListener("keydown", down)
  }, [])

  if (!mounted) return null

  const isDark = theme === "dark" || theme === "catppuccin" || theme === "atom-one" || theme === "nord"

  const items: CommandItem[] = [
    { label: "Dashboard", href: "/", icon: <AppWindow className="h-4 w-4" />, group: "Navigation", shortcut: "G D" },
    { label: "Applications", href: "/applications", icon: <AppWindow className="h-4 w-4" />, group: "Navigation", shortcut: "G A" },
    { label: "Groups", href: "/groups", icon: <FolderKanban className="h-4 w-4" />, group: "Navigation", shortcut: "G G" },
    { label: "Scans", href: "/scans", icon: <Scan className="h-4 w-4" />, group: "Navigation", shortcut: "G S" },
    { label: "Findings", href: "/findings", icon: <Bug className="h-4 w-4" />, group: "Navigation", shortcut: "G F" },
    { label: "Policies", href: "/policies", icon: <ShieldAlert className="h-4 w-4" />, group: "Navigation", shortcut: "G P" },
    { label: "Teams", href: "/teams", icon: <Users className="h-4 w-4" />, group: "Navigation" },
    ...(user?.role === "admin" ? [
      { label: "Users", href: "/users", icon: <Users className="h-4 w-4" />, group: "Admin" },
      { label: "Permissions", href: "/admin/permissions", icon: <Shield className="h-4 w-4" />, group: "Admin" },
      { label: "Scanner Types", href: "/admin/scanner-types", icon: <ScanLine className="h-4 w-4" />, group: "Admin" },
      { label: "Audit Log", href: "/admin/audit-log", icon: <ScrollText className="h-4 w-4" />, group: "Admin" },
    ] : []),
    { label: "Profile", href: "/profile", icon: <Settings className="h-4 w-4" />, group: "Account" },
    { label: isDark ? "Light mode" : "Dark mode", icon: isDark ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />, group: "Appearance", action: toggleTheme },
    { label: "Sign out", icon: <LogOut className="h-4 w-4" />, group: "Account", action: () => logout(true) },
  ]

  return (
    <>
      <button
        onClick={() => setOpen(true)}
        className="flex items-center gap-2 rounded-md border bg-card px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
      >
        <Search className="h-3.5 w-3.5" />
        <span className="hidden sm:inline">Search...</span>
        <kbd className="pointer-events-none ml-2 hidden sm:inline-flex h-5 select-none items-center gap-0.5 rounded border bg-muted px-1 text-[10px] font-medium text-muted-foreground">
          <span className="text-xs">⌘</span>K
        </kbd>
      </button>

      {open && (
        <div className="fixed inset-0 z-50">
          <div className="fixed inset-0 bg-black/40 backdrop-blur-sm" onClick={() => setOpen(false)} />
          <div className="fixed left-1/2 top-[20%] -translate-x-1/2 w-full max-w-lg animate-fade-in">
            <Command className="rounded-xl border bg-popover shadow-2xl overflow-hidden">
              <div className="flex items-center border-b px-3">
                <Search className="h-4 w-4 shrink-0 text-muted-foreground" />
                <Command.Input
                  placeholder="Type a command or search..."
                  className="flex h-11 w-full rounded-md bg-transparent py-3 pl-2 text-sm outline-none placeholder:text-muted-foreground"
                  autoFocus
                />
              </div>
              <Command.List className="max-h-[300px] overflow-y-auto p-2">
                <Command.Empty className="py-6 text-center text-sm text-muted-foreground">
                  No results found.
                </Command.Empty>
                {["Navigation", "Admin", "Appearance", "Account"].map((group) => {
                  const groupItems = items.filter((i) => i.group === group)
                  if (groupItems.length === 0) return null
                  return (
                    <Command.Group key={group} heading={group} className="text-xs font-medium text-muted-foreground mb-1">
                      {groupItems.map((item) => (
                        <Command.Item
                          key={item.label}
                          value={item.label}
                          onSelect={() => runCommand(item.action, item.href)}
                          className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm cursor-pointer select-none data-[selected=true]:bg-accent data-[selected=true]:text-accent-foreground"
                        >
                          {item.icon}
                          <span className="flex-1">{item.label}</span>
                          {item.shortcut && (
                            <kbd className="text-[10px] text-muted-foreground">{item.shortcut}</kbd>
                          )}
                        </Command.Item>
                      ))}
                    </Command.Group>
                  )
                })}
              </Command.List>
            </Command>
          </div>
        </div>
      )}
    </>
  )
}
