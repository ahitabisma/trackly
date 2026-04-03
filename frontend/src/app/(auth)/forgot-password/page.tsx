"use client";

import { useState } from "react";
import Link from "next/link";
import AuthLayout from "@/components/auth/AuthLayout";
import Toast from "@/components/auth/Toast";

export default function ForgotPasswordPage() {
  const [email, setEmail] = useState("");
  const [loading, setLoading] = useState(false);
  const [sent, setSent] = useState(false);
  const [error, setError] = useState("");
  const [toast, setToast] = useState<{
    message: string;
    type: "error" | "success";
  } | null>(null);

  const validateEmail = () => {
    if (!email) {
      setError("Email is required");
      return false;
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      setError("Invalid email format");
      return false;
    }
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!validateEmail()) return;

    setLoading(true);
    try {
      // This would be implemented with your backend
      // For now, simulating the request
      await new Promise((resolve) => setTimeout(resolve, 2000));

      setToast({
        message: "Password reset link sent! Check your email.",
        type: "success",
      });
      setSent(true);
    } catch (err) {
      setToast({
        message: err instanceof Error ? err.message : "An error occurred",
        type: "error",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <AuthLayout
      title={sent ? "Check your email" : "Forgot password?"}
      subtitle={
        sent
          ? `We've sent a password reset link to ${email}`
          : "Enter your email to receive a password reset link"
      }
    >
      {!sent ? (
        <>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <label className="block text-sm font-medium text-gray-700">
                Email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => {
                  setEmail(e.target.value);
                  setError("");
                }}
                placeholder="your@email.com"
                className={`w-full px-4 py-2.5 rounded-lg border-2 transition-all ${
                  error
                    ? "border-red-500 bg-red-50"
                    : "border-gray-300 focus:border-blue-600"
                } focus:outline-none focus:ring-0`}
                disabled={loading}
              />
              {error && <p className="text-sm text-red-600">{error}</p>}
            </div>

            <button
              type="submit"
              disabled={loading}
              className="w-full mt-6 px-4 py-2.5 bg-gray-900 text-white rounded-lg font-medium hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
            >
              {loading && (
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              )}
              {loading ? "Sending..." : "Send Reset Link"}
            </button>
          </form>

          <div className="mt-6 text-center">
            <Link
              href="/auth/sign-in"
              className="text-sm text-blue-600 hover:text-blue-700 font-medium"
            >
              ← Back to Sign In
            </Link>
          </div>
        </>
      ) : (
        <>
          <div className="text-center py-8">
            <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg
                className="w-8 h-8 text-green-600"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
            </div>
            <p className="text-gray-600 mb-6">
              Didn't receive the email? Check your spam folder or try a
              different email.
            </p>
            <button
              onClick={() => {
                setSent(false);
                setEmail("");
              }}
              className="text-blue-600 hover:text-blue-700 font-medium"
            >
              Try another email
            </button>
          </div>

          <div className="mt-6 text-center">
            <Link
              href="/auth/sign-in"
              className="text-sm text-gray-600 hover:text-gray-700 font-medium"
            >
              ← Back to Sign In
            </Link>
          </div>
        </>
      )}

      {toast && (
        <Toast
          message={toast.message}
          type={toast.type}
          onClose={() => setToast(null)}
        />
      )}
    </AuthLayout>
  );
}
