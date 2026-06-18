import type { Check, Monitor, MonitorStatus } from "./types"

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080"

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: { "Content-Type": "application/json", ...init?.headers },
    cache: "no-store",
  })
  if (!res.ok) {
    const body = await res.text()
    throw new Error(`${res.status} ${res.statusText}: ${body}`)
  }
  if (res.status === 204) return undefined as T
  return res.json() as Promise<T>
}

export function listStatuses(): Promise<MonitorStatus[]> {
  return request<MonitorStatus[]>("/api/statuses")
}

export function listChecks(monitorId: number): Promise<Check[]> {
  return request<Check[]>(`/api/monitors/${monitorId}/checks`)
}

export function createMonitor(input: { url: string; name: string; interval_sec: number }): Promise<Monitor> {
  return request<Monitor>("/api/monitors", {
    method: "POST",
    body: JSON.stringify(input),
  })
}

export function deleteMonitor(id: number): Promise<void> {
  return request<void>(`/api/monitors/${id}`, { method: "DELETE" })
}
