"use client";

import React, { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Select } from "@/components/ui/Select";
import { EmptyState } from "@/components/ui/EmptyState";
import { MonitorFormModal } from "@/components/monitors/MonitorFormModal";
import { DeleteConfirmModal } from "@/components/monitors/DeleteConfirmModal";
import { Monitor, HttpMethod, PaginatedResponse } from "@/types/api";
import * as monitorsApi from "@/lib/api/monitors";
import {
  Activity,
  Plus,
  Search,
  RefreshCw,
  Edit2,
  Trash2,
  Pause,
  Play,
  ExternalLink,
  ChevronLeft,
  ChevronRight,
  AlertTriangle,
  X,
} from "lucide-react";

export default function MonitorsPage() {
  const [monitors, setMonitors] = useState<Monitor[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const limit = 10;

  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("ALL");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  // Modals state
  const [createModalOpen, setCreateModalOpen] = useState(false);
  const [editingMonitor, setEditingMonitor] = useState<Monitor | null>(null);
  const [deletingMonitor, setDeletingMonitor] = useState<Monitor | null>(null);
  const [actionLoadingId, setActionLoadingId] = useState<string | null>(null);

  const fetchMonitors = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res: PaginatedResponse<Monitor> = await monitorsApi.listMonitors({
        page,
        limit,
        status: statusFilter === "ALL" ? undefined : statusFilter,
      });

      // Clamp out-of-range page if records were deleted
      if (res.total_pages > 0 && page > res.total_pages) {
        setPage(res.total_pages);
        return;
      }

      setMonitors(res.data || []);
      setTotal(res.total || 0);
      setTotalPages(res.total_pages || 1);
    } catch (err: unknown) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to load monitors. Please check your network connection."
      );
    } finally {
      setLoading(false);
    }
  }, [page, statusFilter]);

  useEffect(() => {
    fetchMonitors();
  }, [fetchMonitors]);

  const handleCreateMonitor = async (data: {
    name: string;
    url: string;
    method: HttpMethod;
    interval_seconds: number;
    timeout_seconds: number;
    expected_status_code: number;
    ssl_check_enabled: boolean;
    failure_threshold: number;
    recovery_threshold: number;
    alert_channel_ids: string[];
  }) => {
    await monitorsApi.createMonitor(data);
    await fetchMonitors();
  };

  const handleEditMonitor = async (data: {
    name: string;
    url: string;
    method: HttpMethod;
    interval_seconds: number;
    timeout_seconds: number;
    expected_status_code: number;
    ssl_check_enabled: boolean;
    failure_threshold: number;
    recovery_threshold: number;
  }) => {
    if (!editingMonitor) return;
    await monitorsApi.updateMonitor(editingMonitor.id, data);
    setEditingMonitor(null);
    await fetchMonitors();
  };

  const handleToggleActive = async (monitor: Monitor) => {
    setActionLoadingId(monitor.id);
    setActionError(null);
    try {
      if (monitor.is_active) {
        await monitorsApi.disableMonitor(monitor.id);
      } else {
        await monitorsApi.enableMonitor(monitor.id);
      }
      await fetchMonitors();
    } catch (err: unknown) {
      setActionError(
        err instanceof Error
          ? err.message
          : `Failed to ${monitor.is_active ? "pause" : "resume"} monitor.`
      );
    } finally {
      setActionLoadingId(null);
    }
  };

  const handleDelete = async () => {
    if (!deletingMonitor) return;
    setActionError(null);
    try {
      await monitorsApi.deleteMonitor(deletingMonitor.id);
      setDeletingMonitor(null);
      // If we deleted the only item on the current page, step back one page
      if (monitors.length === 1 && page > 1) {
        setPage(page - 1);
      } else {
        await fetchMonitors();
      }
    } catch (err: unknown) {
      setActionError(
        err instanceof Error ? err.message : "Failed to delete monitor."
      );
    }
  };

  // Client-side filtering applies to the current page
  const isFiltering = search.trim().length > 0;
  const filteredMonitors = monitors.filter((m) => {
    if (!isFiltering) return true;
    const q = search.toLowerCase();
    return m.name.toLowerCase().includes(q) || m.url.toLowerCase().includes(q);
  });

  return (
    <DashboardLayout
      title="Monitors"
      description={`Manage and configure ${total} uptime probes`}
      action={
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            onClick={fetchMonitors}
            title="Refresh monitors"
          >
            <RefreshCw className="w-3.5 h-3.5" />
          </Button>
          <Button
            variant="primary"
            size="sm"
            onClick={() => setCreateModalOpen(true)}
          >
            <Plus className="w-3.5 h-3.5" />
            <span>Create Monitor</span>
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-4">
        {/* Error Banners */}
        {error && (
          <div className="p-4 rounded-xl bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/50 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
            <div className="flex items-center gap-2.5">
              <AlertTriangle className="w-4 h-4 text-rose-600 dark:text-rose-400 shrink-0" />
              <div>
                <p className="text-xs font-semibold text-rose-800 dark:text-rose-300">
                  Failed to load monitors
                </p>
                <p className="text-[11px] text-rose-600 dark:text-rose-400 mt-0.5">
                  {error}
                </p>
              </div>
            </div>
            <Button variant="secondary" size="sm" onClick={fetchMonitors}>
              <RefreshCw className="w-3.5 h-3.5 mr-1" />
              Retry
            </Button>
          </div>
        )}

        {actionError && (
          <div className="p-3 rounded-lg bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/50 flex items-center justify-between text-xs text-rose-700 dark:text-rose-400">
            <div className="flex items-center gap-2">
              <AlertTriangle className="w-3.5 h-3.5 shrink-0" />
              <span>{actionError}</span>
            </div>
            <button
              onClick={() => setActionError(null)}
              className="text-rose-500 hover:text-rose-800 dark:hover:text-rose-200"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          </div>
        )}

        {/* Filters bar */}
        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-3 shadow-2xs flex flex-col sm:flex-row items-center justify-between gap-3">
          <div className="relative w-full sm:w-80">
            <Search className="w-4 h-4 text-neutral-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              placeholder="Filter this page by name or URL..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-9 pr-8 py-1.5 text-xs rounded-lg border border-neutral-200 dark:border-neutral-800 bg-white dark:bg-[#07090e] text-neutral-900 dark:text-neutral-100 placeholder-neutral-400 focus:outline-none focus:ring-1 focus:ring-emerald-500"
            />
            {search && (
              <button
                onClick={() => setSearch("")}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-neutral-400 hover:text-neutral-600 dark:hover:text-neutral-200"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            )}
          </div>

          <div className="flex items-center gap-2 w-full sm:w-auto justify-end">
            <span className="text-xs text-neutral-500">Status:</span>
            <div className="w-36">
              <Select
                value={statusFilter}
                onChange={(e) => {
                  setStatusFilter(e.target.value);
                  setPage(1);
                }}
                options={[
                  { label: "All Statuses", value: "ALL" },
                  { label: "Up", value: "UP" },
                  { label: "Down", value: "DOWN" },
                  { label: "Degraded", value: "DEGRADED" },
                  { label: "Pending", value: "PENDING" },
                ]}
              />
            </div>
          </div>
        </div>

        {/* Filter Indicator */}
        {isFiltering && (
          <div className="text-[11px] text-neutral-500 flex items-center justify-between px-1">
            <span>
              Showing {filteredMonitors.length} matching monitors on page {page}
            </span>
            <button
              onClick={() => setSearch("")}
              className="text-emerald-600 dark:text-emerald-400 hover:underline"
            >
              Clear page filter
            </button>
          </div>
        )}

        {/* Monitors List */}
        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl shadow-2xs overflow-hidden">
          {error ? (
            <div className="py-12 text-center text-xs text-neutral-500">
              Could not display monitors due to a server error.
            </div>
          ) : filteredMonitors.length === 0 && !loading ? (
            <div className="py-12">
              <EmptyState
                icon={Activity}
                title={
                  isFiltering
                    ? "No matches on this page"
                    : statusFilter !== "ALL"
                    ? "No monitors match status"
                    : "No monitors found"
                }
                description={
                  isFiltering
                    ? "No monitors on the current page match your search term. Clear the search or check other pages."
                    : statusFilter !== "ALL"
                    ? `No monitors found with status "${statusFilter}".`
                    : "You haven't created any uptime monitors yet."
                }
                action={
                  <Button
                    variant="primary"
                    size="sm"
                    onClick={() => {
                      if (isFiltering) {
                        setSearch("");
                      } else if (statusFilter !== "ALL") {
                        setStatusFilter("ALL");
                      } else {
                        setCreateModalOpen(true);
                      }
                    }}
                  >
                    {isFiltering
                      ? "Clear Search"
                      : statusFilter !== "ALL"
                      ? "Show All Statuses"
                      : "Create First Monitor"}
                  </Button>
                }
              />
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse text-xs">
                <thead>
                  <tr className="border-b border-neutral-200 dark:border-neutral-800 text-[11px] font-semibold text-neutral-500 uppercase tracking-wider bg-neutral-50/50 dark:bg-neutral-900/30">
                    <th className="py-3 px-3.5">Status</th>
                    <th className="py-3 px-3.5">Monitor Name</th>
                    <th className="py-3 px-3.5">Target URL</th>
                    <th className="py-3 px-3.5">Response Time</th>
                    <th className="py-3 px-3.5">Interval</th>
                    <th className="py-3 px-3.5">Last Checked</th>
                    <th className="py-3 px-3.5 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-100 dark:divide-neutral-800/60">
                  {filteredMonitors.map((m) => {
                    const isPaused = !m.is_active;
                    const isActionBusy = actionLoadingId === m.id;

                    return (
                      <tr
                        key={m.id}
                        className="hover:bg-neutral-50/70 dark:hover:bg-neutral-900/40 transition-colors"
                      >
                        <td className="py-3 px-3.5 whitespace-nowrap">
                          <Badge status={isPaused ? "PAUSED" : m.status} />
                        </td>
                        <td className="py-3 px-3.5 font-semibold text-neutral-900 dark:text-white whitespace-nowrap">
                          <Link
                            href={`/monitors/${m.id}`}
                            className="hover:text-emerald-600 dark:hover:text-emerald-400 flex items-center gap-1.5 transition-colors"
                          >
                            <span>{m.name}</span>
                            <ExternalLink className="w-3 h-3 text-neutral-400 opacity-60" />
                          </Link>
                        </td>
                        <td className="py-3 px-3.5 font-mono text-[11px] text-neutral-600 dark:text-neutral-400 max-w-xs truncate">
                          <span className="font-semibold text-neutral-800 dark:text-neutral-200 mr-1.5">
                            {m.method}
                          </span>
                          {m.url}
                        </td>
                        <td className="py-3 px-3.5 font-mono text-neutral-800 dark:text-neutral-200 whitespace-nowrap">
                          {m.avg_response_time_ms
                            ? `${m.avg_response_time_ms}ms`
                            : "—"}
                        </td>
                        <td className="py-3 px-3.5 text-neutral-500 font-mono text-[11px] whitespace-nowrap">
                          {m.interval_seconds}s
                        </td>
                        <td className="py-3 px-3.5 text-neutral-500 font-mono text-[11px] whitespace-nowrap">
                          {m.last_checked_at
                            ? new Date(m.last_checked_at).toLocaleTimeString()
                            : "Pending"}
                        </td>
                        <td className="py-3 px-3.5 text-right whitespace-nowrap">
                          <div className="flex items-center justify-end gap-1">
                            <button
                              onClick={() => handleToggleActive(m)}
                              disabled={isActionBusy}
                              title={
                                isPaused ? "Resume probing" : "Pause probing"
                              }
                              className={`p-1.5 rounded-md text-xs transition-colors cursor-pointer ${
                                isPaused
                                  ? "text-emerald-600 hover:bg-emerald-50 dark:hover:bg-emerald-950/40"
                                  : "text-neutral-500 hover:bg-neutral-100 dark:hover:bg-neutral-800 hover:text-neutral-800 dark:hover:text-neutral-200"
                              }`}
                            >
                              {isPaused ? (
                                <Play className="w-3.5 h-3.5" />
                              ) : (
                                <Pause className="w-3.5 h-3.5" />
                              )}
                            </button>

                            <button
                              onClick={() => setEditingMonitor(m)}
                              title="Edit monitor configuration"
                              className="p-1.5 rounded-md text-xs text-neutral-500 hover:bg-neutral-100 dark:hover:bg-neutral-800 hover:text-neutral-800 dark:hover:text-neutral-200 transition-colors cursor-pointer"
                            >
                              <Edit2 className="w-3.5 h-3.5" />
                            </button>

                            <button
                              onClick={() => setDeletingMonitor(m)}
                              title="Delete monitor"
                              className="p-1.5 rounded-md text-xs text-neutral-400 hover:bg-rose-50 dark:hover:bg-rose-950/40 hover:text-rose-600 dark:hover:text-rose-400 transition-colors cursor-pointer"
                            >
                              <Trash2 className="w-3.5 h-3.5" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}

          {/* Pagination bar (only shown when not searching within page) */}
          {totalPages > 1 && !isFiltering && !error && (
            <div className="p-3 border-t border-neutral-100 dark:border-neutral-800 flex items-center justify-between text-xs text-neutral-500">
              <span>
                Page {page} of {totalPages} ({total} total monitors)
              </span>
              <div className="flex items-center gap-1.5">
                <Button
                  variant="secondary"
                  size="sm"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => Math.max(1, p - 1))}
                >
                  <ChevronLeft className="w-3.5 h-3.5 mr-0.5" />
                  Previous
                </Button>
                <Button
                  variant="secondary"
                  size="sm"
                  disabled={page >= totalPages}
                  onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                >
                  Next
                  <ChevronRight className="w-3.5 h-3.5 ml-0.5" />
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Modals */}
      <MonitorFormModal
        isOpen={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        onSubmit={handleCreateMonitor}
      />

      <MonitorFormModal
        isOpen={Boolean(editingMonitor)}
        onClose={() => setEditingMonitor(null)}
        initialData={editingMonitor}
        title="Edit Monitor"
        onSubmit={handleEditMonitor}
      />

      <DeleteConfirmModal
        isOpen={Boolean(deletingMonitor)}
        onClose={() => setDeletingMonitor(null)}
        onConfirm={handleDelete}
        title="Delete Monitor"
        message="Are you sure you want to delete this monitor? All historical heartbeat telemetry and incidents will be permanently deleted."
        itemName={deletingMonitor?.name}
      />
    </DashboardLayout>
  );
}
