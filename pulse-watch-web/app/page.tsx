import React from "react";
import { Navbar } from "@/components/landing/Navbar";
import { HeroSection } from "@/components/landing/HeroSection";
import { LiveConsole } from "@/components/landing/live-console/LiveConsole";
import { SimulatorSection } from "@/components/landing/SimulatorSection";
import { ArchitectureSection } from "@/components/landing/ArchitectureSection";
import { RoiCalculatorSection } from "@/components/landing/RoiCalculatorSection";
import { FeaturesSection } from "@/components/landing/FeaturesSection";
import { PricingSection } from "@/components/landing/PricingSection";
import { CtaSection } from "@/components/landing/CtaSection";
import { Footer } from "@/components/landing/Footer";

export default function HomePage() {
  return (
    <div className="min-h-screen bg-neutral-50 dark:bg-[#07090e] text-neutral-900 dark:text-neutral-100 flex flex-col relative selection:bg-emerald-500/20 selection:text-emerald-800 dark:selection:text-emerald-200">
      {/* Background Ambience */}
      <div className="absolute inset-0 bg-grid-mesh opacity-60 dark:opacity-40 pointer-events-none" />
      <div className="absolute inset-0 bg-radial-ambient pointer-events-none" />

      {/* 1. Header Navigation */}
      <Navbar />

      {/* 2. Hero Presentation */}
      <HeroSection />

      {/* 3. Live Observability Console */}
      <LiveConsole />

      {/* 4. Smart Threshold Noise Filter */}
      <SimulatorSection />

      {/* 5. Decoupled Architecture Pipeline */}
      <ArchitectureSection />

      {/* 6. Downtime ROI Calculator */}
      <RoiCalculatorSection />

      {/* 7. Reliability Feature Bento Grid */}
      <FeaturesSection />

      {/* 8. Pricing Matrix */}
      <PricingSection />

      {/* 9. Final CTA */}
      <CtaSection />

      {/* 10. Developer Footer */}
      <Footer />
    </div>
  );
}
