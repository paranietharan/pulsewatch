import { apiClient } from "./client";
import {
  AlertChannel,
  CreateAlertChannelInput,
  MessageResponse,
} from "@/types/api";

export async function listAlertChannels(): Promise<AlertChannel[]> {
  return apiClient<AlertChannel[]>("/api/v1/alert-channels");
}

export async function createAlertChannel(
  data: CreateAlertChannelInput
): Promise<AlertChannel> {
  return apiClient<AlertChannel>("/api/v1/alert-channels", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function deleteAlertChannel(id: string): Promise<MessageResponse> {
  return apiClient<MessageResponse>(`/api/v1/alert-channels/${id}`, {
    method: "DELETE",
  });
}
