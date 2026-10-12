export interface PingLatency {
  loc: string;
  ms: number;
  optimal: boolean;
}

export interface MonitorProfile {
  id: string;
  name: string;
  endpoint: string;
  method: "GET" | "POST" | "TCP" | "HEAD";
  status: "UP" | "DOWN" | "DEGRADED";
  httpCode: number;
  uptime90d: string;
  latency: number;
  dns: number;
  tcp: number;
  tls: number;
  ttfb: number;
  sslDays: number;
  region: string;
  pings: PingLatency[];
}

export interface SimulatorStage {
  title: string;
  status: string;
  badgeColor: string;
  description: string;
  alertSent: boolean;
}

export interface ArchitectureNode {
  title: string;
  service: string;
  badge: string;
  badgeColor: string;
  tech: string;
  description: string;
  connection: string;
  isHighlight?: boolean;
}

export interface FeatureItem {
  title: string;
  description: string;
  icon: "server" | "shield" | "bolt" | "bell" | "globe" | "code";
  accentColor: string;
}

export interface PricingPlan {
  id: string;
  name: string;
  price: number;
  period: string;
  tagline: string;
  features: string[];
  ctaText: string;
  isPopular?: boolean;
}

export interface HeroMetric {
  value: string;
  label: string;
  accentClass: string;
}
