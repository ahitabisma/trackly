"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { AuthService } from "@/services/auth_service";
import { useToast } from "@/context/ToastContext";
import { signUpSchema, type SignUpInput } from "@/lib/validations/auth";
import { EyeIcon, EyeOffIcon } from "lucide-react";
import { ZodError } from "zod";

type PasswordStrength = "weak" | "fair" | "good" | "strong";

export default function SignUpForm() {
  const router = useRouter();
  const { toast } = useToast();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [agreeTerms, setAgreeTerms] = useState(false);
  const [loading, setLoading] = useState(false);
  const [googleLoading, setGoogleLoading] = useState(false);
  const [passwordStrength, setPasswordStrength] =
    useState<PasswordStrength | null>(null);
  const [errors, setErrors] = useState<{
    name?: string;
    email?: string;
    password?: string;
    confirmPassword?: string;
    agreeTerms?: string;
  }>({});

  const calculatePasswordStrength = (pwd: string): PasswordStrength => {
    let score = 0;

    if (pwd.length >= 8) score++;
    if (pwd.length >= 12) score++;
    if (/[A-Z]/.test(pwd)) score++;
    if (/[a-z]/.test(pwd)) score++;
    if (/[0-9]/.test(pwd)) score++;
    if (/[^A-Za-z0-9]/.test(pwd)) score++;

    if (score <= 2) return "weak";
    if (score <= 3) return "fair";
    if (score <= 4) return "good";
    return "strong";
  };

  const handlePasswordChange = (value: string) => {
    setPassword(value);
    if (value) {
      setPasswordStrength(calculatePasswordStrength(value));
    } else {
      setPasswordStrength(null);
    }
    if (errors.password) setErrors({ ...errors, password: undefined });
  };

  const validateForm = (): boolean => {
    try {
      const data: SignUpInput = {
        name,
        email,
        password,
        confirmPassword,
        agreeTerms,
      };

      signUpSchema.parse(data);
      setErrors({});
      return true;
    } catch (error) {
      if (error instanceof ZodError) {
        const newErrors: typeof errors = {};
        error.errors.forEach((err) => {
          const path = err.path[0] as string;
          newErrors[path as keyof typeof errors] = err.message;
        });
        setErrors(newErrors);
      }
      return false;
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) return;

    setLoading(true);
    try {
      const { error } = await AuthService.signUp(email, password, name);

      if (error) {
        // Error is now a parsed AuthError with user-friendly message
        const errorMessage =
          error.message || "Failed to create account. Please try again.";
        toast({
          message: "Sign up failed",
          description: errorMessage,
          variant: "error",
        });
        return;
      }

      toast({
        message: "Account created!",
        description:
          "Check your email to verify your address. You can sign in once verified.",
        variant: "success",
      });
      setTimeout(() => router.push("/auth/sign-in"), 2500);
    } catch (err) {
      toast({
        message: "An error occurred",
        description: err instanceof Error ? err.message : "Please try again",
        variant: "error",
      });
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleSignUp = async () => {
    setGoogleLoading(true);
    try {
      await AuthService.signInWithGoogle();
    } catch (err) {
      toast({
        message: "Google sign up failed",
        variant: "error",
      });
      setGoogleLoading(false);
    }
  };

  const getStrengthColor = (strength: PasswordStrength | null) => {
    if (!strength) return "bg-gray-200";
    if (strength === "weak") return "bg-red-500";
    if (strength === "fair") return "bg-yellow-500";
    if (strength === "good") return "bg-blue-500";
    return "bg-green-500";
  };

  return (
    <>
      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Name Field */}
        <div className="space-y-2">
          <label className="block text-sm font-medium text-gray-700">
            Full Name
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => {
              setName(e.target.value);
              if (errors.name) setErrors({ ...errors, name: undefined });
            }}
            placeholder="John Doe"
            className={`w-full px-4 py-2.5 rounded-lg border-2 transition-all ${
              errors.name
                ? "border-red-500 bg-red-50"
                : "border-gray-300 focus:border-blue-600"
            } focus:outline-none focus:ring-0`}
            disabled={loading}
          />
          {errors.name && <p className="text-sm text-red-600">{errors.name}</p>}
        </div>

        {/* Email Field */}
        <div className="space-y-2">
          <label className="block text-sm font-medium text-gray-700">
            Email
          </label>
          <input
            type="email"
            value={email}
            onChange={(e) => {
              setEmail(e.target.value);
              if (errors.email) setErrors({ ...errors, email: undefined });
            }}
            placeholder="your@email.com"
            className={`w-full px-4 py-2.5 rounded-lg border-2 transition-all ${
              errors.email
                ? "border-red-500 bg-red-50"
                : "border-gray-300 focus:border-blue-600"
            } focus:outline-none focus:ring-0`}
            disabled={loading}
          />
          {errors.email && (
            <p className="text-sm text-red-600">{errors.email}</p>
          )}
        </div>

        {/* Password Field */}
        <div className="space-y-2">
          <label className="block text-sm font-medium text-gray-700">
            Password
          </label>
          <div className="relative">
            <input
              type={showPassword ? "text" : "password"}
              value={password}
              onChange={(e) => handlePasswordChange(e.target.value)}
              placeholder="••••••••"
              className={`w-full px-4 py-2.5 rounded-lg border-2 transition-all pr-11 ${
                errors.password
                  ? "border-red-500 bg-red-50"
                  : "border-gray-300 focus:border-blue-600"
              } focus:outline-none focus:ring-0`}
              disabled={loading}
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700"
              disabled={loading}
            >
              {showPassword ? <EyeOffIcon size={18} /> : <EyeIcon size={18} />}
            </button>
          </div>

          {/* Password Strength */}
          {password && (
            <div className="space-y-2">
              <div className="flex gap-1">
                {[1, 2, 3, 4].map((i) => {
                  const strengths: PasswordStrength[] = [
                    "weak",
                    "fair",
                    "good",
                    "strong",
                  ];
                  const isActive =
                    (passwordStrength === "weak" && i === 1) ||
                    (passwordStrength === "fair" && i <= 2) ||
                    (passwordStrength === "good" && i <= 3) ||
                    (passwordStrength === "strong" && i <= 4);

                  return (
                    <div
                      key={i}
                      className={`flex-1 h-1 rounded-full transition-colors ${
                        isActive
                          ? getStrengthColor(passwordStrength)
                          : "bg-gray-200"
                      }`}
                    />
                  );
                })}
              </div>
              <p className="text-xs text-gray-600 capitalize">
                Strength:{" "}
                <span
                  className={`font-medium ${
                    passwordStrength === "weak"
                      ? "text-red-600"
                      : passwordStrength === "fair"
                        ? "text-yellow-600"
                        : passwordStrength === "good"
                          ? "text-blue-600"
                          : "text-green-600"
                  }`}
                >
                  {passwordStrength}
                </span>
              </p>
            </div>
          )}

          {errors.password && (
            <p className="text-sm text-red-600">{errors.password}</p>
          )}
        </div>

        {/* Confirm Password Field */}
        <div className="space-y-2">
          <label className="block text-sm font-medium text-gray-700">
            Confirm Password
          </label>
          <div className="relative">
            <input
              type={showConfirmPassword ? "text" : "password"}
              value={confirmPassword}
              onChange={(e) => {
                setConfirmPassword(e.target.value);
                if (errors.confirmPassword)
                  setErrors({ ...errors, confirmPassword: undefined });
              }}
              placeholder="••••••••"
              className={`w-full px-4 py-2.5 rounded-lg border-2 transition-all pr-11 ${
                errors.confirmPassword
                  ? "border-red-500 bg-red-50"
                  : "border-gray-300 focus:border-blue-600"
              } focus:outline-none focus:ring-0`}
              disabled={loading}
            />
            <button
              type="button"
              onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-500 hover:text-gray-700"
              disabled={loading}
            >
              {showConfirmPassword ? (
                <EyeOffIcon size={18} />
              ) : (
                <EyeIcon size={18} />
              )}
            </button>
          </div>
          {errors.confirmPassword && (
            <p className="text-sm text-red-600">{errors.confirmPassword}</p>
          )}
        </div>

        {/* Terms & Conditions */}
        <div className="flex items-start gap-3 py-2">
          <input
            type="checkbox"
            id="terms"
            checked={agreeTerms}
            onChange={(e) => {
              setAgreeTerms(e.target.checked);
              if (errors.agreeTerms)
                setErrors({ ...errors, agreeTerms: undefined });
            }}
            className="w-5 h-5 mt-0.5 rounded border-2 border-gray-300 cursor-pointer"
            disabled={loading}
          />
          <label
            htmlFor="terms"
            className="text-sm text-gray-600 cursor-pointer leading-relaxed"
          >
            I agree to the{" "}
            <Link
              href="/terms"
              className="text-blue-600 hover:text-blue-700 font-medium"
            >
              Terms & Conditions
            </Link>{" "}
            and{" "}
            <Link
              href="/privacy"
              className="text-blue-600 hover:text-blue-700 font-medium"
            >
              Privacy Policy
            </Link>
          </label>
        </div>
        {errors.agreeTerms && (
          <p className="text-sm text-red-600">{errors.agreeTerms}</p>
        )}

        {/* Sign Up Button */}
        <button
          type="submit"
          disabled={loading}
          className="w-full mt-6 px-4 py-2.5 bg-gray-900 text-white rounded-lg font-medium hover:bg-gray-800 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
        >
          {loading && (
            <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
          )}
          {loading ? "Creating account..." : "Create Account"}
        </button>
      </form>

      {/* Divider */}
      <div className="relative my-6">
        <div className="absolute inset-0 flex items-center">
          <div className="w-full border-t border-gray-300" />
        </div>
        <div className="relative flex justify-center text-sm">
          <span className="px-2 bg-white text-gray-500">or</span>
        </div>
      </div>

      {/* Google Sign Up */}
      <button
        type="button"
        onClick={handleGoogleSignUp}
        disabled={googleLoading || loading}
        className="w-full px-4 py-2.5 border-2 border-gray-300 rounded-lg font-medium text-gray-900 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2"
      >
        <svg
          width="18"
          height="18"
          viewBox="0 0 24 24"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <path
            d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
            fill="#4285F4"
          />
          <path
            d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
            fill="#34A853"
          />
          <path
            d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
            fill="#FBBC05"
          />
          <path
            d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
            fill="#EA4335"
          />
        </svg>
        {googleLoading ? "Signing up..." : "Continue with Google"}
      </button>

      {/* Sign In Link */}
      <p className="text-center text-sm text-gray-600 mt-6">
        Already have an account?{" "}
        <Link
          href="/auth/sign-in"
          className="font-medium text-blue-600 hover:text-blue-700"
        >
          Sign in
        </Link>
      </p>
    </>
  );
}
