export interface Monitor {
  id: number
  url: string
  name: string
  interval_sec: number
  created_at: string
}

export interface Check {
  id: number
  monitor_id: number
  status_code: number
  response_ms: number
  is_up: boolean
  error: string
  checked_at: string
}

export interface MonitorStatus {
  monitor: Monitor
  latest_check: Check | null
}
