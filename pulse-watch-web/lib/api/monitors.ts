import { apiClient } from "./client";
import {
  Monitor,
  Heartbeat,
  Incident,
  DashboardSummary,
  PaginatedResponse,
  CreateMonitorInput,
  UpdateMonitorInput,
  MessageResponse,
} from "@/types/api";

export async function getDashboardSummary(): Promise<DashboardSummary> {
  return apiClient<DashboardSummary>("/api/v1/dashboard/summary");
}

export async function listMonitors(params: {
  page?: number;
  limit?: number;
  status?: string;
} = {}): Promise<PaginatedResponse<Monitor>> {
  const query = new URLSearchParams();
  if (params.page) query.set("page", params.page.toString());
  if (params.limit) query.set("limit", params.limit.toString());
  if (params.status) query.set("status", params.status);

  const qs = query.toString();
  return apiClient<PaginatedResponse<Monitor>>(
    `/api/v1/monitors${qs ? `?${qs}` : ""}`
  );
}

export async function getMonitor(id: string): Promise<Monitor> {
  return apiClient<Monitor>(`/api/v1/monitors/${id}`);
}

export async function createMonitor(data: CreateMonitorInput): Promise<Monitor> {
  return apiClient<Monitor>("/api/v1/monitors", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateMonitor(
  id: string,
  data: UpdateMonitorInput
): Promise<Monitor> {
  return apiClient<Monitor>(`/api/v1/monitors/${id}`, {
    method: "PATCH",
    body: JSON.stringify(data),
  });
}

export async function deleteMonitor(id: string): Promise<MessageResponse> {
  return apiClient<MessageResponse>(`/api/v1/monitors/${id}`, {
    method: "DELETE",
  });
}

export async function enableMonitor(id: string): Promise<MessageResponse> {
  return apiClient<MessageResponse>(`/api/v1/monitors/${id}/enable`, {
    method: "POST",
  });
}

export async function disableMonitor(id: string): Promise<MessageResponse> {
  return apiClient<MessageResponse>(`/api/v1/monitors/${id}/disable`, {
    method: "POST",
  });
}

export async function getResults(
  id: string,
  params: { page?: number; limit?: number } = {}
): Promise<PaginatedResponse<Heartbeat>> {
  const query = new URLSearchParams();
  if (params.page) query.set("page", params.page.toString());
  if (params.limit) query.set("limit", params.limit.toString());

  const qs = query.toString();
  return apiClient<PaginatedResponse<Heartbeat>>(
    `/api/v1/monitors/${id}/results${qs ? `?${qs}` : ""}`
  );
}

export async function getIncidents(
  id: string,
  params: { page?: number; limit?: number } = {}
): Promise<PaginatedResponse<Incident>> {
  const query = new URLSearchParams();
  if (params.page) query.set("page", params.page.toString());
  if (params.limit) query.set("limit", params.limit.toString());

  const qs = query.toString();
  return apiClient<PaginatedResponse<Incident>>(
    `/api/v1/monitors/${id}/incidents${qs ? `?${qs}` : ""}`
  );
}
