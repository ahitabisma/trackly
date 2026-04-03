"use client";

import { useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { AuthService } from "@/services/auth_service";
import { useToast } from "@/context/ToastContext";

export default function AuthCallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { toast } = useToast();

  useEffect(() => {
    const handleCallback = async () => {
      try {
        // Check for error in URL
        const error = searchParams.get("error");
        const errorCode = searchParams.get("error_code");
        const errorDescription = searchParams.get("error_description");

        if (error) {
          console.error("Auth Error:", {
            error,
            errorCode,
            errorDescription,
          });

          // Decode error description
          const decodedDesc = errorDescription
            ? decodeURIComponent(errorDescription)
            : "Unknown error";

          toast({
            message: "Authentication failed",
            description: decodedDesc,
            variant: "error",
            duration: 5000,
          });

          // Redirect to error page with details
          router.replace(
            `/auth/error?error=${error}&code=${errorCode}&desc=${encodeURIComponent(decodedDesc)}`,
          );
          return;
        }

        // Get session to verify auth was successful
        const { data, error: sessionError } = await AuthService.getSession();

        if (sessionError || !data?.session) {
          throw new Error("Failed to establish session");
        }

        toast({
          message: "Signed in successfully!",
          description: "Welcome to Trackly",
          variant: "success",
        });

        // Redirect to dashboard
        setTimeout(() => {
          router.replace("/dashboard");
        }, 1500);
      } catch (err) {
        console.error("Callback Error:", err);
        toast({
          message: "Authentication error",
          description: err instanceof Error ? err.message : "Please try again",
          variant: "error",
        });

        setTimeout(() => {
          router.replace("/auth/sign-in");
        }, 2000);
      }
    };

    handleCallback();
  }, [router, searchParams, toast]);

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <div className="w-12 h-12 border-4 border-gray-200 border-t-blue-600 rounded-full animate-spin mx-auto mb-4" />
        <p className="text-gray-600 font-medium">
          Completing authentication...
        </p>
      </div>
    </div>
  );
}
