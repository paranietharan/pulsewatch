"use client";

import React, { useState, useEffect, useCallback, Suspense } from "react";
import Link from "next/link";
import { useRouter, useParams } from "next/navigation";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { LoadingSpinner } from "@/components/ui/LoadingSpinner";
import { LatencyChart } from "@/components/monitors/LatencyChart";
import { HeartbeatResultsTable } from "@/components/monitors/HeartbeatResultsTable";
import { IncidentHistoryTable } from "@/components/monitors/IncidentHistoryTable";
import { MonitorFormModal } from "@/components/monitors/MonitorFormModal";
import { DeleteConfirmModal } from "@/components/monitors/DeleteConfirmModal";
import { Monitor, Heartbeat, Incident, HttpMethod } from "@/types/api";
import { FetchError } from "@/lib/api/client";
import * as monitorsApi from "@/lib/api/monitors";
import {
  ArrowLeft,
  RefreshCw,
  Edit2,
  Trash2,
  Pause,
  Play,
  ExternalLink,
  ShieldCheck,
  Activity,
  AlertTriangle,
  ChevronLeft,
  ChevronRight,
  X,
} from "lucide-react";

function MonitorDetailsContent() {
  const router = useRouter();
  const params = useParams();
  const id = params?.id as string;

  const [monitor, setMonitor] = useState<Monitor | null>(null);
  const [results, setResults] = useState<Heartbeat[]>([]);
  const [resultsTotal, setResultsTotal] = useState(0);
  const [resultsPage, setResultsPage] = useState(1);
  const [resultsTotalPages, setResultsTotalPages] = useState(1);

  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [incidentsTotal, setIncidentsTotal] = useState(0);

  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);
  const [fetchError, setFetchError] = useState<string | null>(null);
  const [resultsError, setResultsError] = useState<string | null>(null);
  const [incidentsError, setIncidentsError] = useState<string | null>(null);
  const [actionError, setActionError] = useState<string | null>(null);

  const [activeTab, setActiveTab] = useState<"telemetry" | "incidents">("telemetry");

  // Modals state
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [isToggling, setIsToggling] = useState(false);

  const loadMonitorData = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    setFetchError(null);
    setNotFound(false);

    try {
      const mon = await monitorsApi.getMonitor(id);
      setMonitor(mon);

      // Fetch telemetry and incidents in parallel
      const [resResult, incResult] = await Promise.allSettled([
        monitorsApi.getResults(id, { page: resultsPage, limit: 20 }),
        monitorsApi.getIncidents(id, { limit: 20 }),
      ]);

      if (resResult.status === "fulfilled") {
        setResults(resResult.value.data || []);
        setResultsTotal(resResult.value.total || 0);
        setResultsTotalPages(resResult.value.total_pages || 1);
        setResultsError(null);
      } else {
        setResultsError(
          resResult.reason instanceof Error
            ? resResult.reason.message
            : "Failed to load probe results."
        );
      }

      if (incResult.status === "fulfilled") {
        setIncidents(incResult.value.data || []);
        setIncidentsTotal(incResult.value.total || 0);
        setIncidentsError(null);
      } else {
        setIncidentsError(
          incResult.reason instanceof Error
            ? incResult.reason.message
            : "Failed to load incidents."
        );
      }
    } catch (err: unknown) {
      if (err instanceof FetchError && err.status === 404) {
        setNotFound(true);
      } else {
        setFetchError(
          err instanceof Error
            ? err.message
            : "Failed to load monitor telemetry."
        );
      }
    } finally {
      setLoading(false);
    }
  }, [id, resultsPage]);

  useEffect(() => {
    loadMonitorData();
  }, [loadMonitorData]);

  const handleToggleActive = async () => {
    if (!monitor) return;
    setIsToggling(true);
    setActionError(null);
    try {
      if (monitor.is_active) {
        await monitorsApi.disableMonitor(monitor.id);
      } else {
        await monitorsApi.enableMonitor(monitor.id);
      }
      await loadMonitorData();
    } catch (err: unknown) {
      setActionError(
        err instanceof Error
          ? err.message
          : `Failed to ${monitor.is_active ? "pause" : "resume"} monitor.`
      );
    } finally {
      setIsToggling(false);
    }
  };

  const handleEdit = async (data: {
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
    if (!monitor) return;
    await monitorsApi.updateMonitor(monitor.id, data);
    setEditModalOpen(false);
    await loadMonitorData();
  };

  const handleDelete = async () => {
    if (!monitor) return;
    await monitorsApi.deleteMonitor(monitor.id);
    setDeleteModalOpen(false);
    router.replace("/monitors");
  };

  if (loading && !monitor && !fetchError && !notFound) {
    return (
      <DashboardLayout title="Monitor Details">
        <div className="py-24 flex items-center justify-center">
          <LoadingSpinner text="Loading monitor telemetry..." />
        </div>
      </DashboardLayout>
    );
  }

  if (notFound) {
    return (
      <DashboardLayout title="Monitor Not Found">
        <div className="py-16 text-center">
          <p className="text-sm text-neutral-500 mb-4">
            The requested monitor does not exist or has been deleted.
          </p>
          <Link href="/monitors">
            <Button variant="secondary" size="sm">
              <ArrowLeft className="w-3.5 h-3.5 mr-1" />
              Back to Monitors
            </Button>
          </Link>
        </div>
      </DashboardLayout>
    );
  }

  if (fetchError || !monitor) {
    return (
      <DashboardLayout title="Error Loading Monitor">
        <div className="py-16 max-w-md mx-auto text-center flex flex-col items-center">
          <div className="w-10 h-10 rounded-full bg-rose-50 dark:bg-rose-950/40 text-rose-600 dark:text-rose-400 flex items-center justify-center mb-3">
            <AlertTriangle className="w-5 h-5" />
          </div>
          <h2 className="text-sm font-bold text-neutral-900 dark:text-white">
            Could not retrieve monitor data
          </h2>
          <p className="text-xs text-neutral-500 mt-1 mb-5 leading-relaxed">
            {fetchError || "An unexpected error occurred while communicating with the backend."}
          </p>
          <div className="flex items-center gap-2">
            <Link href="/monitors">
              <Button variant="secondary" size="sm">
                <ArrowLeft className="w-3.5 h-3.5 mr-1" />
                Back to Monitors
              </Button>
            </Link>
            <Button variant="primary" size="sm" onClick={loadMonitorData}>
              <RefreshCw className="w-3.5 h-3.5 mr-1" />
              Retry
            </Button>
          </div>
        </div>
      </DashboardLayout>
    );
  }

  const isPaused = !monitor.is_active;

  return (
    <DashboardLayout
      title={monitor.name}
      description={monitor.url}
      action={
        <div className="flex items-center gap-2">
          <Button
            variant="secondary"
            size="sm"
            onClick={loadMonitorData}
            title="Refresh monitor telemetry"
          >
            <RefreshCw className="w-3.5 h-3.5" />
          </Button>

          <Button
            variant="secondary"
            size="sm"
            onClick={handleToggleActive}
            isLoading={isToggling}
          >
            {isPaused ? (
              <>
                <Play className="w-3.5 h-3.5 text-emerald-500" />
                <span>Resume</span>
              </>
            ) : (
              <>
                <Pause className="w-3.5 h-3.5 text-amber-500" />
                <span>Pause</span>
              </>
            )}
          </Button>

          <Button
            variant="secondary"
            size="sm"
            onClick={() => setEditModalOpen(true)}
          >
            <Edit2 className="w-3.5 h-3.5" />
            <span>Edit</span>
          </Button>

          <Button
            variant="danger"
            size="sm"
            onClick={() => setDeleteModalOpen(true)}
          >
            <Trash2 className="w-3.5 h-3.5" />
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-5">
        {/* Action Error Alert */}
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

        {/* Breadcrumb back link */}
        <div className="flex items-center justify-between">
          <Link
            href="/monitors"
            className="text-xs text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-200 flex items-center gap-1 transition-colors"
          >
            <ArrowLeft className="w-3.5 h-3.5" />
            <span>Back to Monitors</span>
          </Link>

          <div className="flex items-center gap-2">
            <span className="text-xs text-neutral-500">Status:</span>
            <Badge status={isPaused ? "PAUSED" : monitor.status} />
          </div>
        </div>

        {/* Top metrics summary cards */}
        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
          <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-3.5 shadow-2xs">
            <p className="text-[11px] font-semibold text-neutral-500">Uptime (24h)</p>
            <p className="text-lg font-bold font-mono text-neutral-900 dark:text-white mt-1">
              {typeof monitor.uptime_24h === "number" ? `${monitor.uptime_24h.toFixed(2)}%` : "—"}
            </p>
          </div>

          <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-3.5 shadow-2xs">
            <p className="text-[11px] font-semibold text-neutral-500">Uptime (7d)</p>
            <p className="text-lg font-bold font-mono text-neutral-900 dark:text-white mt-1">
              {typeof monitor.uptime_7d === "number" ? `${monitor.uptime_7d.toFixed(2)}%` : "—"}
            </p>
          </div>

          <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-3.5 shadow-2xs">
            <p className="text-[11px] font-semibold text-neutral-500">Uptime (30d)</p>
            <p className="text-lg font-bold font-mono text-neutral-900 dark:text-white mt-1">
              {typeof monitor.uptime_30d === "number" ? `${monitor.uptime_30d.toFixed(2)}%` : "—"}
            </p>
          </div>

          <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-3.5 shadow-2xs">
            <p className="text-[11px] font-semibold text-neutral-500">Avg Latency</p>
            <p className="text-lg font-bold font-mono text-neutral-900 dark:text-white mt-1">
              {typeof monitor.avg_response_time_ms === "number"
                ? `${monitor.avg_response_time_ms}ms`
                : "—"}
            </p>
          </div>

          <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-3.5 shadow-2xs">
            <p className="text-[11px] font-semibold text-neutral-500">Consecutive OK</p>
            <p className="text-lg font-bold font-mono text-emerald-600 dark:text-emerald-400 mt-1">
              {monitor.consecutive_successes}
            </p>
          </div>

          <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-3.5 shadow-2xs">
            <p className="text-[11px] font-semibold text-neutral-500">Consecutive Fails</p>
            <p className="text-lg font-bold font-mono text-rose-600 dark:text-rose-400 mt-1">
              {monitor.consecutive_failures}
            </p>
          </div>
        </div>

        {/* Configuration Details Panel */}
        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-4 shadow-2xs">
          <div className="flex items-center justify-between pb-3 border-b border-neutral-100 dark:border-neutral-800/80 mb-3.5">
            <h2 className="text-xs font-bold uppercase tracking-wider text-neutral-500 flex items-center gap-1.5">
              <ShieldCheck className="w-3.5 h-3.5 text-emerald-500" />
              Endpoint Configuration
            </h2>
            <a
              href={monitor.url}
              target="_blank"
              rel="noreferrer"
              className="text-xs text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-200 flex items-center gap-1 font-mono"
            >
              <span>Visit Target</span>
              <ExternalLink className="w-3 h-3" />
            </a>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-xs">
            <div>
              <span className="text-neutral-500 block text-[11px]">HTTP Method & URL</span>
              <span className="font-mono font-semibold text-neutral-900 dark:text-white break-all">
                {monitor.method} {monitor.url}
              </span>
            </div>

            <div>
              <span className="text-neutral-500 block text-[11px]">Interval / Timeout</span>
              <span className="font-mono font-medium text-neutral-800 dark:text-neutral-200">
                {monitor.interval_seconds}s check / {monitor.timeout_seconds}s timeout
              </span>
            </div>

            <div>
              <span className="text-neutral-500 block text-[11px]">Expected Status Code</span>
              <span className="font-mono font-medium text-neutral-800 dark:text-neutral-200">
                HTTP {monitor.expected_status_code}
              </span>
            </div>

            <div>
              <span className="text-neutral-500 block text-[11px]">SSL Validation</span>
              <span className="font-medium text-neutral-800 dark:text-neutral-200">
                {monitor.ssl_check_enabled ? "Enabled (Strict)" : "Disabled"}
              </span>
            </div>

            <div>
              <span className="text-neutral-500 block text-[11px]">Failure / Recovery Threshold</span>
              <span className="font-mono text-neutral-800 dark:text-neutral-200">
                {monitor.failure_threshold} fails / {monitor.recovery_threshold} oks
              </span>
            </div>

            <div>
              <span className="text-neutral-500 block text-[11px]">Last Checked</span>
              <span className="font-mono text-neutral-800 dark:text-neutral-200">
                {monitor.last_checked_at
                  ? new Date(monitor.last_checked_at).toLocaleString()
                  : "Never"}
              </span>
            </div>

            <div>
              <span className="text-neutral-500 block text-[11px]">Next Scheduled Check</span>
              <span className="font-mono text-neutral-800 dark:text-neutral-200">
                {monitor.next_check_at
                  ? new Date(monitor.next_check_at).toLocaleString()
                  : "—"}
              </span>
            </div>

            <div>
              <span className="text-neutral-500 block text-[11px]">Created At</span>
              <span className="font-mono text-neutral-800 dark:text-neutral-200">
                {new Date(monitor.created_at).toLocaleDateString()}
              </span>
            </div>
          </div>
        </div>

        {/* Telemetry Visualizer Card */}
        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-4 shadow-2xs">
          <div className="pb-3 border-b border-neutral-100 dark:border-neutral-800/80 mb-3 flex items-center justify-between">
            <h2 className="text-xs font-bold uppercase tracking-wider text-neutral-500 flex items-center gap-1.5">
              <Activity className="w-3.5 h-3.5 text-emerald-500" />
              Response Time Latency Telemetry
            </h2>
            <span className="text-xs text-neutral-400 font-mono">
              Last {results.length} heartbeats
            </span>
          </div>

          <LatencyChart results={results} />
        </div>

        {/* Results & Incidents Tabs Card */}
        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl shadow-2xs overflow-hidden">
          <div className="flex items-center gap-4 px-4 pt-3 border-b border-neutral-200 dark:border-neutral-800">
            <button
              onClick={() => setActiveTab("telemetry")}
              className={`pb-3 text-xs font-semibold flex items-center gap-1.5 border-b-2 transition-colors cursor-pointer ${
                activeTab === "telemetry"
                  ? "border-emerald-500 text-neutral-900 dark:text-white"
                  : "border-transparent text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-200"
              }`}
            >
              <Activity className="w-3.5 h-3.5" />
              <span>Heartbeat Logs ({resultsTotal})</span>
            </button>

            <button
              onClick={() => setActiveTab("incidents")}
              className={`pb-3 text-xs font-semibold flex items-center gap-1.5 border-b-2 transition-colors cursor-pointer ${
                activeTab === "incidents"
                  ? "border-rose-500 text-neutral-900 dark:text-white"
                  : "border-transparent text-neutral-500 hover:text-neutral-800 dark:hover:text-neutral-200"
              }`}
            >
              <AlertTriangle className="w-3.5 h-3.5" />
              <span>Incidents ({incidentsTotal})</span>
            </button>
          </div>

          <div className="p-4">
            {activeTab === "telemetry" ? (
              <div>
                {resultsError ? (
                  <div className="p-4 rounded-xl bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/50 flex items-center justify-between">
                    <span className="text-xs text-rose-700 dark:text-rose-400">
                      {resultsError}
                    </span>
                    <Button variant="secondary" size="sm" onClick={loadMonitorData}>
                      <RefreshCw className="w-3.5 h-3.5 mr-1" />
                      Retry
                    </Button>
                  </div>
                ) : (
                  <>
                    <HeartbeatResultsTable results={results} />
                    {resultsTotalPages > 1 && (
                      <div className="pt-4 mt-3 border-t border-neutral-100 dark:border-neutral-800 flex items-center justify-between text-xs text-neutral-500">
                        <span>
                          Page {resultsPage} of {resultsTotalPages} ({resultsTotal} total probes)
                        </span>
                        <div className="flex items-center gap-1.5">
                          <Button
                            variant="secondary"
                            size="sm"
                            disabled={resultsPage <= 1}
                            onClick={() => setResultsPage((p) => Math.max(1, p - 1))}
                          >
                            <ChevronLeft className="w-3.5 h-3.5 mr-0.5" />
                            Previous
                          </Button>
                          <Button
                            variant="secondary"
                            size="sm"
                            disabled={resultsPage >= resultsTotalPages}
                            onClick={() => setResultsPage((p) => Math.min(resultsTotalPages, p + 1))}
                          >
                            Next
                            <ChevronRight className="w-3.5 h-3.5 ml-0.5" />
                          </Button>
                        </div>
                      </div>
                    )}
                  </>
                )}
              </div>
            ) : (
              <div>
                {incidentsError ? (
                  <div className="p-4 rounded-xl bg-rose-50 dark:bg-rose-950/30 border border-rose-200 dark:border-rose-900/50 flex items-center justify-between">
                    <span className="text-xs text-rose-700 dark:text-rose-400">
                      {incidentsError}
                    </span>
                    <Button variant="secondary" size="sm" onClick={loadMonitorData}>
                      <RefreshCw className="w-3.5 h-3.5 mr-1" />
                      Retry
                    </Button>
                  </div>
                ) : (
                  <IncidentHistoryTable incidents={incidents} />
                )}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Edit Modal */}
      <MonitorFormModal
        isOpen={editModalOpen}
        onClose={() => setEditModalOpen(false)}
        initialData={monitor}
        title="Edit Monitor"
        onSubmit={handleEdit}
      />

      {/* Delete Modal */}
      <DeleteConfirmModal
        isOpen={deleteModalOpen}
        onClose={() => setDeleteModalOpen(false)}
        onConfirm={handleDelete}
        title="Delete Monitor"
        message="Are you sure you want to permanently delete this monitor and all associated heartbeat records?"
        itemName={monitor.name}
      />
    </DashboardLayout>
  );
}

export default function MonitorDetailsPage() {
  return (
    <Suspense
      fallback={
        <DashboardLayout title="Loading...">
          <div className="py-24 flex items-center justify-center">
            <LoadingSpinner text="Loading..." />
          </div>
        </DashboardLayout>
      }
    >
      <MonitorDetailsContent />
    </Suspense>
  );
}
