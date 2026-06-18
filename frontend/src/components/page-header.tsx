"use client"

import Link from "next/link"

interface Crumb {
  label: string
  href?: string
}

export function PageHeader({ crumbs }: { crumbs: Crumb[] }) {
  return (
    <p className="text-xs text-muted-foreground leading-relaxed">
      {crumbs.map((crumb, i) => (
        <span key={i}>
          {i > 0 && <span className="mx-1.5 text-muted-foreground/30 select-none">/</span>}
          {crumb.href ? (
            <Link href={crumb.href} className="hover:text-foreground transition-colors">
              {crumb.label}
            </Link>
          ) : (
            <span className="text-foreground/80">{crumb.label}</span>
          )}
        </span>
      ))}
    </p>
  )
}
