import type { Metadata, Viewport } from "next";
import { Inter, Plus_Jakarta_Sans, JetBrains_Mono } from "next/font/google";
import { ThemeProvider } from "@/components/theme/ThemeProvider";
import "./globals.css";

const inter = Inter({
  subsets: ["latin"],
  variable: "--font-inter",
  display: "swap",
});

const jakarta = Plus_Jakarta_Sans({
  subsets: ["latin"],
  variable: "--font-heading",
  display: "swap",
});

const mono = JetBrains_Mono({
  subsets: ["latin"],
  variable: "--font-mono",
  display: "swap",
});

export const metadata: Metadata = {
  title: "PulseWatch — Distributed Uptime & API Performance Monitoring",
  description:
    "Continuous HTTP, TLS, & TCP probing with distributed edge workers, microsecond latency telemetry, and instant threshold-driven alerts before customers notice.",
  keywords: [
    "uptime monitoring",
    "synthetic monitoring",
    "API monitoring",
    "ping monitor",
    "status page",
    "incident management",
    "SSL certificate monitor",
    "PulseWatch",
  ],
  authors: [{ name: "PulseWatch Core Team" }],
  icons: {
    icon: [
      { url: "/icon.svg", type: "image/svg+xml" },
      { url: "/favicon.svg", type: "image/svg+xml" },
    ],
    apple: "/apple-icon.svg",
  },
  openGraph: {
    title: "PulseWatch — Distributed Uptime & API Performance Monitoring",
    description:
      "Enterprise-grade distributed uptime probing with zero false positives and sub-second alert dispatch.",
    siteName: "PulseWatch",
    type: "website",
  },
};

export const viewport: Viewport = {
  themeColor: [
    { media: "(prefers-color-scheme: light)", color: "#ffffff" },
    { media: "(prefers-color-scheme: dark)", color: "#07090e" },
  ],
  width: "device-width",
  initialScale: 1,
};

const themeInitScript = `
  (function() {
    try {
      var stored = localStorage.getItem('pulsewatch-theme');
      var isDark = stored ? stored === 'dark' : true; // Default dark
      if (stored === 'system') {
        isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      }
      if (isDark) {
        document.documentElement.classList.add('dark');
        document.documentElement.style.colorScheme = 'dark';
      } else {
        document.documentElement.classList.remove('dark');
        document.documentElement.style.colorScheme = 'light';
      }
    } catch (e) {}
  })();
`;

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={`${inter.variable} ${jakarta.variable} ${mono.variable} dark`}
    >
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeInitScript }} />
      </head>
      <body className="min-h-full flex flex-col bg-white dark:bg-[#07090e] text-neutral-900 dark:text-neutral-100 transition-colors duration-150 antialiased selection:bg-emerald-500/20 selection:text-emerald-700 dark:selection:text-emerald-300">
        <ThemeProvider>{children}</ThemeProvider>
      </body>
    </html>
  );
}
