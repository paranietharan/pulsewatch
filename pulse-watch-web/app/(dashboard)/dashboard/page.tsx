"use client";

import React, { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { SummaryCards } from "@/components/dashboard/SummaryCards";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { EmptyState } from "@/components/ui/EmptyState";
import { MonitorFormModal } from "@/components/monitors/MonitorFormModal";
import { DashboardSummary, Monitor, HttpMethod } from "@/types/api";
import * as monitorsApi from "@/lib/api/monitors";
import {
  Activity,
  Plus,
  RefreshCw,
  ExternalLink,
  ArrowRight,
  AlertTriangle,
} from "lucide-react";

export default function DashboardOverviewPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [recentMonitors, setRecentMonitors] = useState<Monitor[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [createModalOpen, setCreateModalOpen] = useState(false);

  const loadData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const [sumData, monitorsData] = await Promise.all([
        monitorsApi.getDashboardSummary(),
        monitorsApi.listMonitors({ limit: 5 }),
      ]);
      setSummary(sumData);
      setRecentMonitors(monitorsData?.data || []);
    } catch (err: unknown) {
      setError(
        err instanceof Error
          ? err.message
          : "Failed to load dashboard data. Check backend connection."
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

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
    await loadData();
  };

  return (
    <DashboardLayout
      title="Overview"
      description="System health, telemetry summary, and active incidents"
      action={
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            onClick={loadData}
            title="Refresh dashboard"
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
      <div className="flex flex-col gap-5">
        {error && (
          <div className="p-4 rounded-xl bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/50 flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
            <div className="flex items-center gap-2.5">
              <AlertTriangle className="w-4 h-4 text-rose-600 dark:text-rose-400 shrink-0" />
              <div>
                <p className="text-xs font-semibold text-rose-800 dark:text-rose-300">
                  Failed to load dashboard telemetry
                </p>
                <p className="text-[11px] text-rose-600 dark:text-rose-400 mt-0.5">
                  {error}
                </p>
              </div>
            </div>
            <Button variant="secondary" size="sm" onClick={loadData}>
              <RefreshCw className="w-3.5 h-3.5 mr-1" />
              Retry
            </Button>
          </div>
        )}

        {summary && summary.active_incidents > 0 && (
          <div className="p-3.5 rounded-xl bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/50 flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <AlertTriangle className="w-4 h-4 text-rose-600 dark:text-rose-400 shrink-0" />
              <p className="text-xs font-semibold text-rose-800 dark:text-rose-300">
                {summary.active_incidents} active incident
                {summary.active_incidents > 1 ? "s" : ""} detected across your
                monitors.
              </p>
            </div>
            <Link
              href="/incidents"
              className="text-xs font-semibold text-rose-700 dark:text-rose-400 hover:underline flex items-center gap-1"
            >
              <span>View Incidents</span>
              <ArrowRight className="w-3 h-3" />
            </Link>
          </div>
        )}

        <SummaryCards summary={summary} isLoading={loading} />

        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-5 shadow-2xs">
          <div className="flex items-center justify-between pb-3.5 border-b border-neutral-100 dark:border-neutral-800/80 mb-3">
            <div>
              <h2 className="text-xs font-bold uppercase tracking-wider text-neutral-500">
                Recent Monitors
              </h2>
            </div>
            {recentMonitors.length > 0 && (
              <Link
                href="/monitors"
                className="text-xs font-semibold text-neutral-700 dark:text-neutral-300 hover:text-neutral-900 dark:hover:text-white flex items-center gap-1"
              >
                <span>View all</span>
                <ArrowRight className="w-3 h-3" />
              </Link>
            )}
          </div>

          {error ? (
            <div className="py-8 text-center text-xs text-neutral-500">
              Could not load recent monitors due to an API error.
            </div>
          ) : recentMonitors.length === 0 && !loading ? (
            <EmptyState
              icon={Activity}
              title="No monitors found"
              description="You haven't set up any uptime monitors yet. Add your first HTTP or TCP endpoint to start probing."
              action={
                <Button
                  variant="primary"
                  size="sm"
                  onClick={() => setCreateModalOpen(true)}
                >
                  <Plus className="w-3.5 h-3.5 mr-1" />
                  Add First Monitor
                </Button>
              }
            />
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse text-xs">
                <thead>
                  <tr className="border-b border-neutral-100 dark:border-neutral-800 text-[11px] font-semibold text-neutral-500 uppercase tracking-wider">
                    <th className="py-2.5 px-3">Status</th>
                    <th className="py-2.5 px-3">Name</th>
                    <th className="py-2.5 px-3">Target URL</th>
                    <th className="py-2.5 px-3">Response</th>
                    <th className="py-2.5 px-3">Interval</th>
                    <th className="py-2.5 px-3">Last Checked</th>
                    <th className="py-2.5 px-3 text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-100 dark:divide-neutral-800/60">
                  {recentMonitors.map((m) => (
                    <tr
                      key={m.id}
                      className="hover:bg-neutral-50/70 dark:hover:bg-neutral-900/40 transition-colors"
                    >
                      <td className="py-2.5 px-3 whitespace-nowrap">
                        <Badge status={m.is_active ? m.status : "PAUSED"} />
                      </td>
                      <td className="py-2.5 px-3 font-semibold text-neutral-900 dark:text-white whitespace-nowrap">
                        <Link
                          href={`/monitors/${m.id}`}
                          className="hover:underline"
                        >
                          {m.name}
                        </Link>
                      </td>
                      <td className="py-2.5 px-3 font-mono text-[11px] text-neutral-500 max-w-xs truncate">
                        <span className="font-semibold text-neutral-700 dark:text-neutral-300 mr-1.5">
                          {m.method}
                        </span>
                        {m.url}
                      </td>
                      <td className="py-2.5 px-3 font-mono text-neutral-800 dark:text-neutral-200">
                        {m.avg_response_time_ms
                          ? `${m.avg_response_time_ms}ms`
                          : "—"}
                      </td>
                      <td className="py-2.5 px-3 text-neutral-500 font-mono text-[11px]">
                        {m.interval_seconds}s
                      </td>
                      <td className="py-2.5 px-3 text-neutral-500 font-mono text-[11px] whitespace-nowrap">
                        {m.last_checked_at
                          ? new Date(m.last_checked_at).toLocaleTimeString()
                          : "Pending"}
                      </td>
                      <td className="py-2.5 px-3 text-right">
                        <Link
                          href={`/monitors/${m.id}`}
                          className="text-neutral-400 hover:text-neutral-900 dark:hover:text-white p-1 inline-block"
                          title="View monitor telemetry"
                        >
                          <ExternalLink className="w-3.5 h-3.5" />
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>

      <MonitorFormModal
        isOpen={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        onSubmit={handleCreateMonitor}
      />
    </DashboardLayout>
  );
}
