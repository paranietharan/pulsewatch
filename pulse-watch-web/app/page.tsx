import React from "react";
import { Navbar } from "@/components/landing/Navbar";
import { HeroSection } from "@/components/landing/HeroSection";
import { FeaturesSection } from "@/components/landing/FeaturesSection";
import { PricingSection } from "@/components/landing/PricingSection";
import { CtaSection } from "@/components/landing/CtaSection";
import { Footer } from "@/components/landing/Footer";

export default function HomePage() {
  return (
    <div className="min-h-screen bg-neutral-50 dark:bg-[#07090e] text-neutral-900 dark:text-neutral-100 flex flex-col relative selection:bg-emerald-500/20 selection:text-emerald-800 dark:selection:text-emerald-200">
      <div className="absolute inset-0 bg-grid-mesh opacity-60 dark:opacity-40 pointer-events-none" />
      <div className="absolute inset-0 bg-radial-ambient pointer-events-none" />
      <Navbar />
      <HeroSection />
      <FeaturesSection />
      <PricingSection />
      <CtaSection />
      <Footer />
    </div>
  );
}
