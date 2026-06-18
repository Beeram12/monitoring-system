"use client"

import { useCallback, useEffect, useState } from "react"
import { AddMonitorDialog } from "@/components/add-monitor-dialog"
import { MonitorTable } from "@/components/monitor-table"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { listStatuses } from "@/lib/api"
import type { MonitorStatus } from "@/lib/types"

const POLL_INTERVAL_MS = 5000

export default function DashboardPage() {
  const [statuses, setStatuses] = useState<MonitorStatus[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = useCallback(async () => {
    try {
      const data = await listStatuses()
      setStatuses(data)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load monitors")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    refresh()
    const interval = setInterval(refresh, POLL_INTERVAL_MS)
    return () => clearInterval(interval)
  }, [refresh])

  const upCount = statuses.filter((s) => s.latest_check?.is_up).length
  const downCount = statuses.filter((s) => s.latest_check && !s.latest_check.is_up).length

  return (
    <main className="mx-auto max-w-5xl space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Uptime Monitor</h1>
          <p className="text-sm text-muted-foreground">
            Tracking {statuses.length} URL{statuses.length === 1 ? "" : "s"} · refreshes every{" "}
            {POLL_INTERVAL_MS / 1000}s
          </p>
        </div>
        <AddMonitorDialog onCreated={refresh} />
      </div>

      <div className="grid grid-cols-3 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Total</CardTitle>
          </CardHeader>
          <CardContent className="text-2xl font-semibold">{statuses.length}</CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Up</CardTitle>
          </CardHeader>
          <CardContent className="text-2xl font-semibold text-emerald-600">{upCount}</CardContent>
        </Card>
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Down</CardTitle>
          </CardHeader>
          <CardContent className="text-2xl font-semibold text-destructive">{downCount}</CardContent>
        </Card>
      </div>

      <Card>
        <CardContent className="pt-6">
          {error && <p className="mb-4 text-sm text-destructive">{error}</p>}
          {loading ? (
            <div className="space-y-2">
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-8 w-full" />
              <Skeleton className="h-8 w-full" />
            </div>
          ) : (
            <MonitorTable statuses={statuses} onChanged={refresh} />
          )}
        </CardContent>
      </Card>
    </main>
  )
}
