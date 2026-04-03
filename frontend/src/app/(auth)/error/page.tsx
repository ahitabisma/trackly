"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { AlertCircleIcon } from "lucide-react";

export default function AuthErrorPage() {
  const searchParams = useSearchParams();

  const error = searchParams.get("error");
  const errorCode = searchParams.get("code");
  const errorDesc = searchParams.get("desc");

  const errorMessages: Record<string, string> = {
    server_error: "Server error occurred. Please try again later.",
    invalid_grant: "Invalid email or password.",
    user_already_exists: "This email is already registered.",
    weak_password: "Password is too weak. Use uppercase, numbers, and symbols.",
    invalid_email: "Invalid email address.",
    auth_error: "Authentication failed. Please try again.",
  };

  const getIcon = (errorCode: string | null) => {
    return (
      <div className="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <AlertCircleIcon className="w-8 h-8 text-red-600" />
      </div>
    );
  };

  const getErrorMessage = (code: string | null) => {
    if (!code) return "An authentication error occurred";
    return (
      errorMessages[code] || errorDesc || errorMessages[error || "auth_error"]
    );
  };

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="max-w-md w-full text-center">
        {getIcon(errorCode)}

        <h1 className="text-2xl font-bold text-gray-900 mb-2">
          {error === "server_error" ? "Server Error" : "Authentication Error"}
        </h1>

        <p className="text-gray-600 mb-6 leading-relaxed">
          {getErrorMessage(errorCode)}
        </p>

        {/* Error Details (Development) */}
        {process.env.NODE_ENV === "development" && (
          <div className="bg-gray-100 rounded-lg p-4 mb-6 text-left text-xs text-gray-700 font-mono overflow-auto max-h-48">
            <p className="font-bold mb-2">Error Details:</p>
            {error && <p>Error: {error}</p>}
            {errorCode && <p>Code: {errorCode}</p>}
            {errorDesc && <p>Message: {decodeURIComponent(errorDesc)}</p>}
          </div>
        )}

        {/* Action Buttons */}
        <div className="space-y-3">
          <Link
            href="/auth/sign-in"
            className="block w-full px-4 py-2.5 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-colors"
          >
            Back to Sign In
          </Link>

          {error === "server_error" && (
            <Link
              href="/auth/sign-up"
              className="block w-full px-4 py-2.5 border-2 border-gray-300 text-gray-900 rounded-lg font-medium hover:bg-gray-50 transition-colors"
            >
              Try Sign Up Again
            </Link>
          )}

          <Link
            href="/"
            className="block w-full px-4 py-2.5 text-gray-600 hover:text-gray-900 font-medium"
          >
            Go Home
          </Link>
        </div>

        {/* Help Text */}
        <div className="mt-8 pt-6 border-t border-gray-200">
          <p className="text-sm text-gray-500 mb-3">
            Still having issues? Contact support or check our documentation.
          </p>
          <div className="flex justify-center gap-4 text-sm">
            <a
              href="mailto:support@trackly.com"
              className="text-blue-600 hover:text-blue-700"
            >
              Email Support
            </a>
            <span className="text-gray-300">|</span>
            <a
              href="https://trackly.com/docs"
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-600 hover:text-blue-700"
            >
              Documentation
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}
