"use client";

import { createContext, useContext } from "react";

interface DashboardShellContextType {
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
}

export const DashboardShellContext = createContext<DashboardShellContextType>({
  sidebarOpen: false,
  setSidebarOpen: () => {},
});

export function useDashboardShell() {
  return useContext(DashboardShellContext);
}
