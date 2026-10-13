"use client";

import React, { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { DashboardLayout } from "@/components/layout/DashboardLayout";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/Button";
import { Select } from "@/components/ui/Select";
import { EmptyState } from "@/components/ui/EmptyState";
import { Incident, Monitor } from "@/types/api";
import * as monitorsApi from "@/lib/api/monitors";
import {
  AlertTriangle,
  RefreshCw,
  ExternalLink,
  ShieldCheck,
  CheckCircle2,
  Clock,
} from "lucide-react";

export default function IncidentsPage() {
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [monitors, setMonitors] = useState<Monitor[]>([]);
  const [statusFilter, setStatusFilter] = useState<string>("ALL");
  const [loading, setLoading] = useState(true);

  const fetchIncidents = useCallback(async () => {
    try {
      setLoading(true);
      const monitorsRes = await monitorsApi.listMonitors({ limit: 100 });
      const monitorList = monitorsRes.data || [];
      setMonitors(monitorList);

      if (monitorList.length === 0) {
        setIncidents([]);
        return;
      }

      // Fetch incidents for all monitors concurrently
      const incidentPromises = monitorList.map(async (m) => {
        try {
          const res = await monitorsApi.getIncidents(m.id, { limit: 50 });
          return (res.data || []).map((inc) => ({
            ...inc,
            monitor_name: m.name,
          }));
        } catch {
          return [];
        }
      });

      const results = await Promise.all(incidentPromises);
      const combined = results.flat();

      // Sort descending by started_at
      combined.sort(
        (a, b) => new Date(b.started_at).getTime() - new Date(a.started_at).getTime()
      );

      setIncidents(combined);
    } catch {
      setIncidents([]);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchIncidents();
  }, [fetchIncidents]);

  const filteredIncidents = incidents.filter((inc) => {
    if (statusFilter === "ALL") return true;
    return inc.status === statusFilter;
  });

  const openCount = incidents.filter((i) => i.status === "OPEN").length;
  const resolvedCount = incidents.filter((i) => i.status === "RESOLVED").length;

  const formatDuration = (sec?: number | null) => {
    if (!sec && sec !== 0) return "Ongoing";
    if (sec < 60) return `${sec}s`;
    if (sec < 3600) return `${Math.floor(sec / 60)}m ${sec % 60}s`;
    return `${Math.floor(sec / 3600)}h ${Math.floor((sec % 3600) / 60)}m`;
  };

  return (
    <DashboardLayout
      title="Incidents"
      description="Track service outages, threshold breaches, and resolution timelines"
      action={
        <Button
          variant="secondary"
          size="sm"
          onClick={fetchIncidents}
          title="Refresh incidents"
        >
          <RefreshCw className="w-3.5 h-3.5" />
          <span>Refresh</span>
        </Button>
      }
    >
      <div className="flex flex-col gap-4">
        {/* Metric summary pills */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-3.5 shadow-2xs flex items-center justify-between">
            <div>
              <p className="text-[11px] font-semibold text-neutral-500">Total Incidents</p>
              <p className="text-xl font-bold font-mono text-neutral-900 dark:text-white mt-0.5">
                {incidents.length}
              </p>
            </div>
            <div className="w-8 h-8 rounded-lg bg-neutral-100 dark:bg-neutral-800 flex items-center justify-center text-neutral-600 dark:text-neutral-300">
              <Clock className="w-4 h-4" />
            </div>
          </div>

          <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-3.5 shadow-2xs flex items-center justify-between">
            <div>
              <p className="text-[11px] font-semibold text-neutral-500">Active Outages</p>
              <p className="text-xl font-bold font-mono text-rose-600 dark:text-rose-400 mt-0.5">
                {openCount}
              </p>
            </div>
            <div className="w-8 h-8 rounded-lg bg-rose-50 dark:bg-rose-950/40 flex items-center justify-center text-rose-600 dark:text-rose-400">
              <AlertTriangle className="w-4 h-4" />
            </div>
          </div>

          <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-3.5 shadow-2xs flex items-center justify-between">
            <div>
              <p className="text-[11px] font-semibold text-neutral-500">Resolved Incidents</p>
              <p className="text-xl font-bold font-mono text-emerald-600 dark:text-emerald-400 mt-0.5">
                {resolvedCount}
              </p>
            </div>
            <div className="w-8 h-8 rounded-lg bg-emerald-50 dark:bg-emerald-950/40 flex items-center justify-center text-emerald-600 dark:text-emerald-400">
              <CheckCircle2 className="w-4 h-4" />
            </div>
          </div>
        </div>

        {/* Filter bar */}
        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl p-3 shadow-2xs flex items-center justify-between">
          <span className="text-xs font-semibold text-neutral-700 dark:text-neutral-300">
            Incident History Log
          </span>

          <div className="flex items-center gap-2">
            <span className="text-xs text-neutral-500">Filter:</span>
            <div className="w-36">
              <Select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                options={[
                  { label: "All Incidents", value: "ALL" },
                  { label: "Open / Active", value: "OPEN" },
                  { label: "Resolved", value: "RESOLVED" },
                ]}
              />
            </div>
          </div>
        </div>

        {/* Incidents Table */}
        <div className="bg-white dark:bg-[#0a0e1a] border border-neutral-200 dark:border-neutral-800/80 rounded-xl shadow-2xs overflow-hidden">
          {filteredIncidents.length === 0 && !loading ? (
            <div className="py-12">
              <EmptyState
                icon={ShieldCheck}
                title="No incidents detected"
                description={
                  statusFilter !== "ALL"
                    ? "No incidents match your selected filter."
                    : monitors.length === 0
                    ? "No monitors configured yet. Add monitors to start outage tracking."
                    : "All configured monitors are currently healthy and operational."
                }
              />
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left border-collapse text-xs">
                <thead>
                  <tr className="border-b border-neutral-200 dark:border-neutral-800 text-[11px] font-semibold text-neutral-500 uppercase tracking-wider bg-neutral-50/50 dark:bg-neutral-900/30">
                    <th className="py-3 px-3.5">Status</th>
                    <th className="py-3 px-3.5">Monitor</th>
                    <th className="py-3 px-3.5">Cause Summary</th>
                    <th className="py-3 px-3.5">Started At</th>
                    <th className="py-3 px-3.5">Resolved At</th>
                    <th className="py-3 px-3.5">Duration</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-neutral-100 dark:divide-neutral-800/60">
                  {filteredIncidents.map((inc) => (
                    <tr
                      key={inc.id}
                      className="hover:bg-neutral-50/70 dark:hover:bg-neutral-900/40 transition-colors"
                    >
                      <td className="py-3 px-3.5 whitespace-nowrap">
                        <Badge status={inc.status} />
                      </td>
                      <td className="py-3 px-3.5 font-semibold text-neutral-900 dark:text-white whitespace-nowrap">
                        <Link
                          href={`/monitors/${inc.monitor_id}`}
                          className="hover:text-emerald-600 dark:hover:text-emerald-400 flex items-center gap-1 transition-colors"
                        >
                          <span>{inc.monitor_name || inc.monitor_id}</span>
                          <ExternalLink className="w-3 h-3 text-neutral-400 opacity-60" />
                        </Link>
                      </td>
                      <td className="py-3 px-3.5 text-neutral-700 dark:text-neutral-300">
                        {inc.cause_summary || "Automated threshold failure detected"}
                      </td>
                      <td className="py-3 px-3.5 font-mono text-[11px] text-neutral-500 whitespace-nowrap">
                        {new Date(inc.started_at).toLocaleString()}
                      </td>
                      <td className="py-3 px-3.5 font-mono text-[11px] text-neutral-500 whitespace-nowrap">
                        {inc.resolved_at
                          ? new Date(inc.resolved_at).toLocaleString()
                          : "—"}
                      </td>
                      <td className="py-3 px-3.5 font-mono text-[11px] whitespace-nowrap">
                        <span
                          className={
                            inc.status === "OPEN"
                              ? "text-rose-600 dark:text-rose-400 font-semibold"
                              : "text-neutral-600 dark:text-neutral-400"
                          }
                        >
                          {formatDuration(inc.duration_seconds)}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </DashboardLayout>
  );
}
