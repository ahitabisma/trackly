"use client";

import { ReactNode } from "react";
import Link from "next/link";

interface AuthLayoutProps {
  children: ReactNode;
  title: string;
  subtitle: string;
}

export default function AuthLayout({
  children,
  title,
  subtitle,
}: AuthLayoutProps) {
  return (
    <div className="min-h-screen grid grid-cols-1 lg:grid-cols-2">
      {/* Left Panel - Dark with Branding */}
      <div className="hidden lg:flex flex-col justify-between bg-gradient-to-br from-gray-900 to-gray-800 p-12 text-white relative overflow-hidden">
        {/* Background Pattern */}
        <div className="absolute inset-0 opacity-5">
          <svg width="100%" height="100%" xmlns="http://www.w3.org/2000/svg">
            <defs>
              <pattern
                id="grid"
                width="40"
                height="40"
                patternUnits="userSpaceOnUse"
              >
                <path
                  d="M 40 0 L 0 0 0 40"
                  fill="none"
                  stroke="white"
                  strokeWidth="0.5"
                />
              </pattern>
            </defs>
            <rect width="100%" height="100%" fill="url(#grid)" />
          </svg>
        </div>

        {/* Logo */}
        <div className="relative z-10">
          <h1 className="text-4xl font-bold tracking-tight mb-2">Trackly</h1>
          <p className="text-sm text-gray-400 tracking-widest uppercase">
            v1.0
          </p>
        </div>

        {/* Hero Section */}
        <div className="relative z-10 max-w-md">
          <h2 className="text-4xl font-bold leading-tight mb-4">
            Track Your <span className="text-blue-400">Shareholding</span>{" "}
            Effortlessly
          </h2>
          <p className="text-gray-300 text-sm leading-relaxed mb-8">
            Manage, analyze, and monitor all your investments in one place.
            Beautiful dashboards, powerful insights.
          </p>

          {/* Features */}
          <div className="space-y-3">
            {[
              "Real-time portfolio tracking",
              "Advanced reporting tools",
              "Secure and encrypted data",
              "Multi-asset support",
            ].map((feature) => (
              <div
                key={feature}
                className="flex items-center gap-3 text-sm text-gray-300"
              >
                <div className="w-2 h-2 rounded-full bg-blue-400 flex-shrink-0" />
                {feature}
              </div>
            ))}
          </div>
        </div>

        {/* Footer */}
        <div className="relative z-10 text-xs text-gray-500 tracking-wider uppercase">
          © 2026 Trackly Inc. All rights reserved.
        </div>
      </div>

      {/* Right Panel - Light with Form */}
      <div className="flex flex-col items-center justify-center px-6 py-12 bg-white relative">
        {/* Mobile Logo */}
        <div className="lg:hidden mb-8 text-center">
          <h1 className="text-3xl font-bold text-gray-900 mb-1">Trackly</h1>
          <p className="text-xs text-gray-500 tracking-widest uppercase">
            Shareholding Platform
          </p>
        </div>

        {/* Background Pattern (optional) */}
        <div className="absolute inset-0 opacity-3 pointer-events-none">
          <svg width="100%" height="100%" xmlns="http://www.w3.org/2000/svg">
            <defs>
              <pattern
                id="dot-pattern"
                width="30"
                height="30"
                patternUnits="userSpaceOnUse"
              >
                <circle cx="15" cy="15" r="1" fill="currentColor" />
              </pattern>
            </defs>
            <rect width="100%" height="100%" fill="url(#dot-pattern)" />
          </svg>
        </div>

        {/* Form Container */}
        <div className="w-full max-w-md relative z-10">
          {/* Header */}
          <div className="text-center mb-8">
            <h2 className="text-3xl font-bold text-gray-900 mb-2">{title}</h2>
            <p className="text-gray-600 text-sm">{subtitle}</p>
          </div>

          {/* Form */}
          {children}
        </div>

        {/* Footer Links */}
        <div className="mt-12 text-center text-xs text-gray-500 space-y-1">
          <p>
            <Link
              href="/terms"
              className="text-blue-600 hover:text-blue-700 font-medium"
            >
              Terms of Service
            </Link>
            {" • "}
            <Link
              href="/privacy"
              className="text-blue-600 hover:text-blue-700 font-medium"
            >
              Privacy Policy
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
