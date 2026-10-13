export interface SafeUser {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  role: "USER" | "ADMIN";
  is_active: boolean;
  is_verified: boolean;
  created_at: string;
  last_login_at?: string;
}

export interface AuthResponse {
  access_token: string;
  refresh_token?: string;
  expires_in: number;
  user: SafeUser;
}

export interface SignupResponse {
  message: string;
  user: SafeUser;
}

export interface MessageResponse {
  message: string;
}

export type MonitorStatus = "PENDING" | "UP" | "DOWN" | "DEGRADED";
export type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "HEAD" | "OPTIONS";

export interface Monitor {
  id: string;
  user_id: string;
  name: string;
  url: string;
  method: HttpMethod;
  interval_seconds: number;
  timeout_seconds: number;
  expected_status_code: number;
  headers: Record<string, string>;
  body?: string;
  status: MonitorStatus;
  is_active: boolean;
  ssl_check_enabled: boolean;
  failure_threshold: number;
  recovery_threshold: number;
  consecutive_failures: number;
  consecutive_successes: number;
  last_checked_at: string | null;
  next_check_at: string;
  uptime_24h: number;
  uptime_7d: number;
  uptime_30d: number;
  avg_response_time_ms: number;
  created_at: string;
  updated_at: string;
}

export interface Heartbeat {
  id: number;
  monitor_id: string;
  status: MonitorStatus;
  status_code: number;
  response_time_ms: number;
  dns_lookup_ms: number;
  tcp_connect_ms: number;
  tls_handshake_ms: number;
  error_message?: string;
  region: string;
  checked_at: string;
  job_id?: string;
}

export type IncidentStatus = "OPEN" | "ACKNOWLEDGED" | "RESOLVED";

export interface IncidentUpdate {
  id: number;
  incident_id: string;
  status: string;
  message: string;
  created_at: string;
}

export interface Incident {
  id: string;
  monitor_id: string;
  monitor_name?: string;
  status: IncidentStatus;
  cause_summary: string;
  started_at: string;
  resolved_at?: string | null;
  duration_seconds?: number | null;
  created_at: string;
  updated_at: string;
  updates?: IncidentUpdate[];
}

export type AlertChannelType = "EMAIL" | "WEBHOOK";

export interface AlertChannel {
  id: string;
  user_id: string;
  type: AlertChannelType;
  name: string;
  config: Record<string, unknown>;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface DashboardSummary {
  total_monitors: number;
  healthy_monitors: number;
  degraded_monitors: number;
  down_monitors: number;
  paused_monitors: number;
  active_incidents: number;
  avg_uptime_24h: number;
  avg_response_time_ms: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

export interface CreateMonitorInput {
  name: string;
  url: string;
  method?: HttpMethod;
  interval_seconds?: number;
  timeout_seconds?: number;
  expected_status_code?: number;
  headers?: Record<string, string>;
  body?: string;
  ssl_check_enabled?: boolean;
  failure_threshold?: number;
  recovery_threshold?: number;
  alert_channel_ids?: string[];
}

export interface UpdateMonitorInput {
  name?: string;
  url?: string;
  method?: HttpMethod;
  interval_seconds?: number;
  timeout_seconds?: number;
  expected_status_code?: number;
  headers?: Record<string, string>;
  body?: string;
  ssl_check_enabled?: boolean;
  failure_threshold?: number;
  recovery_threshold?: number;
}

export interface CreateAlertChannelInput {
  type: AlertChannelType;
  name: string;
  config: Record<string, unknown>;
}

export interface ApiError {
  error?: string;
  message?: string;
  status?: number;
}
