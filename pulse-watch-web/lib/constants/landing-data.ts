import { FeatureItem, PricingPlan, HeroMetric } from "@/types/landing";

export const HERO_METRICS: HeroMetric[] = [
  { value: "99.99%", label: "Target SLA", accentClass: "text-white" },
  { value: "10s", label: "Check Cadence", accentClass: "text-cyan-400" },
  { value: "< 50ms", label: "RabbitMQ Dispatch", accentClass: "text-emerald-400" },
  { value: "0 Noise", label: "3-Tier Verification", accentClass: "text-sky-400" },
];

export const FEATURES: FeatureItem[] = [
  {
    title: "Multi-Protocol Probing",
    description: "Probe HTTP/S endpoints, raw TCP sockets, TLS cert validity, and WebSocket handshakes with microsecond timestamps.",
    icon: "server",
    accentColor: "cyan",
  },
  {
    title: "SSRF Security Shield",
    description: "Worker dialers automatically validate destination IP addresses, rejecting loopback, link-local, and RFC1918 subnets.",
    icon: "shield",
    accentColor: "emerald",
  },
  {
    title: "RabbitMQ Event Backbone",
    description: "Check dispatching and result processing are decoupled via AMQP queues with backpressure management.",
    icon: "bolt",
    accentColor: "sky",
  },
  {
    title: "HMAC-Signed Webhooks",
    description: "Send tamper-proof incident alerts to Slack, Discord, or internal microservices signed with SHA-256 tokens.",
    icon: "bell",
    accentColor: "amber",
  },
  {
    title: "Multi-Region Probing",
    description: "Execute synthetic checks across US, Europe, and Asia-Pacific edge zones to detect localized internet peering outages.",
    icon: "globe",
    accentColor: "indigo",
  },
  {
    title: "100% Open & Self-Hostable",
    description: "Single `docker compose up` spins up Postgres, RabbitMQ, Mailpit, API, and workers. Full control over data privacy.",
    icon: "code",
    accentColor: "violet",
  },
];

export const PRICING_PLANS: PricingPlan[] = [
  {
    id: "community",
    name: "Community",
    price: 0,
    period: "/ forever",
    tagline: "Self-hosted Docker setup",
    features: [
      "Up to 15 Monitors",
      "60-Second Check Cadence",
      "Mailpit & SMTP Alerts",
      "Full Docker Compose Access",
    ],
    ctaText: "Deploy Self-Hosted",
  },
  {
    id: "pro",
    name: "Cloud Pro",
    price: 29,
    period: "/ month",
    tagline: "For growing engineering teams",
    isPopular: true,
    features: [
      "Up to 100 Monitors",
      "10-Second High Frequency",
      "Multi-Region Consensus",
      "HMAC Webhooks + Slack",
      "1-Year Telemetry Retention",
    ],
    ctaText: "Start 14-Day Free Trial",
  },
  {
    id: "enterprise",
    name: "Enterprise",
    price: 99,
    period: "/ month",
    tagline: "Mission-critical infrastructure",
    features: [
      "Unlimited Monitors",
      "5-Second Probing Frequency",
      "Dedicated RabbitMQ Cluster",
      "99.99% Guaranteed SLA",
    ],
    ctaText: "Contact Enterprise",
  },
];
