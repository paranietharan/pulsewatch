"use client";

import React from "react";
import { Header } from "./Header";
import { useDashboardShell } from "./DashboardShellContext";

interface DashboardLayoutProps {
  children: React.ReactNode;
  title: string;
  description?: string;
  action?: React.ReactNode;
}

export function DashboardLayout({
  children,
  title,
  description,
  action,
}: DashboardLayoutProps) {
  const { setSidebarOpen } = useDashboardShell();

  return (
    <>
      <Header
        title={title}
        description={description}
        action={action}
        onOpenSidebar={() => setSidebarOpen(true)}
      />
      <main className="flex-1 p-4 sm:p-6 max-w-7xl w-full mx-auto">
        {children}
      </main>
    </>
  );
}
