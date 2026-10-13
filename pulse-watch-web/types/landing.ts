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
