"use client"

import { useMemo } from "react"
import { PieChart, Pie, Cell, Tooltip, Legend, ResponsiveContainer } from "recharts"

const SEVERITY_ORDER = ["critical", "high", "medium", "low", "info"] as const

const SEVERITY_COLORS: Record<string, string> = {
  critical: "hsl(0 72% 51%)",
  high: "hsl(24 96% 53%)",
  medium: "hsl(47 96% 48%)",
  low: "hsl(222 89% 56%)",
  info: "hsl(215 14% 50%)",
}

const SEVERITY_LABELS: Record<string, string> = {
  critical: "Critical",
  high: "High",
  medium: "Medium",
  low: "Low",
  info: "Info",
}

interface SeverityChartProps {
  findings: { severity: string }[]
  loading?: boolean
}

interface CountEntry {
  severity: string
  count: number
  color: string
}

function CustomTooltip({ active, payload }: any) {
  if (!active || !payload?.length) return null
  const entry = payload[0].payload as CountEntry
  return (
    <div className="rounded-lg border bg-popover px-3 py-2 text-sm shadow-md">
      <p className="font-medium capitalize" style={{ color: entry.color }}>
        {SEVERITY_LABELS[entry.severity] || entry.severity}
      </p>
      <p className="text-muted-foreground">
        {entry.count} finding{entry.count !== 1 ? "s" : ""}
      </p>
    </div>
  )
}

export function SeverityChart({ findings, loading }: SeverityChartProps) {
  const data = useMemo<CountEntry[]>(() => {
    const counts: Record<string, number> = {}
    for (const f of findings) {
      const s = f.severity?.toLowerCase() || "unknown"
      counts[s] = (counts[s] || 0) + 1
    }
    return SEVERITY_ORDER
      .filter((s) => counts[s] && counts[s] > 0)
      .map((s) => ({ severity: s, count: counts[s], color: SEVERITY_COLORS[s] || "hsl(0 0% 60%)" }))
  }, [findings])

  const total = data.reduce((sum, d) => sum + d.count, 0)

  if (loading) {
    return (
      <div className="flex h-64 items-center justify-center text-muted-foreground">
        Loading...
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className="flex h-64 items-center justify-center text-muted-foreground">
        <p>No findings data for the default version</p>
      </div>
    )
  }

  return (
    <div className="w-full">
      <ResponsiveContainer width="100%" height={280}>
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            innerRadius={70}
            outerRadius={110}
            paddingAngle={3}
            dataKey="count"
            nameKey="severity"
            strokeWidth={0}
          >
            {data.map((entry) => (
              <Cell key={entry.severity} fill={entry.color} />
            ))}
          </Pie>
          <Tooltip content={<CustomTooltip />} />
          <Legend
            verticalAlign="bottom"
            height={36}
            formatter={(value: string) => (
              <span className="text-sm capitalize text-muted-foreground">
                {SEVERITY_LABELS[value] || value}
              </span>
            )}
          />
        </PieChart>
      </ResponsiveContainer>
      <p className="text-center text-sm text-muted-foreground -mt-2">
        {total} finding{total !== 1 ? "s" : ""} total
      </p>
    </div>
  )
}
