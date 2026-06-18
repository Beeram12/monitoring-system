"use client"

import { useState } from "react"
import { toast } from "sonner"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { deleteMonitor } from "@/lib/api"
import type { MonitorStatus } from "@/lib/types"

function StatusBadge({ status }: { status: MonitorStatus }) {
  if (!status.latest_check) {
    return <Badge variant="secondary">Pending</Badge>
  }
  return status.latest_check.is_up ? (
    <Badge className="bg-emerald-600 text-white hover:bg-emerald-600">Up</Badge>
  ) : (
    <Badge variant="destructive">Down</Badge>
  )
}

export function MonitorTable({
  statuses,
  onChanged,
}: {
  statuses: MonitorStatus[]
  onChanged: () => void
}) {
  const [deletingId, setDeletingId] = useState<number | null>(null)

  async function handleDelete(id: number) {
    setDeletingId(id)
    try {
      await deleteMonitor(id)
      toast.success("Monitor removed")
      onChanged()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "Failed to delete monitor")
    } finally {
      setDeletingId(null)
    }
  }

  if (statuses.length === 0) {
    return (
      <div className="rounded-lg border border-dashed p-10 text-center text-muted-foreground">
        No monitors yet. Add a URL to start tracking uptime.
      </div>
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>URL</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>HTTP code</TableHead>
          <TableHead>Response time</TableHead>
          <TableHead>Last checked</TableHead>
          <TableHead className="text-right">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {statuses.map((s) => (
          <TableRow key={s.monitor.id}>
            <TableCell className="font-medium">{s.monitor.name || s.monitor.url}</TableCell>
            <TableCell className="max-w-xs truncate text-muted-foreground">
              {s.monitor.url}
            </TableCell>
            <TableCell>
              <StatusBadge status={s} />
            </TableCell>
            <TableCell>{s.latest_check?.status_code || "—"}</TableCell>
            <TableCell>
              {s.latest_check ? `${s.latest_check.response_ms} ms` : "—"}
            </TableCell>
            <TableCell className="text-muted-foreground">
              {s.latest_check
                ? new Date(s.latest_check.checked_at).toLocaleTimeString()
                : "—"}
            </TableCell>
            <TableCell className="text-right">
              <Button
                variant="ghost"
                size="sm"
                disabled={deletingId === s.monitor.id}
                onClick={() => handleDelete(s.monitor.id)}
              >
                Remove
              </Button>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
