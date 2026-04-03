// Auth error type definitions
export interface AuthError {
  message: string;
  code?: string;
  status?: number;
}

// Error code mapping
export const AUTH_ERRORS: Record<string, string> = {
  user_already_exists:
    "This email is already registered. Please sign in instead.",
  weak_password:
    "Password must be at least 8 characters with uppercase, lowercase, numbers, and symbols.",
  invalid_email: "Please enter a valid email address.",
  invalid_credentials: "Invalid email or password.",
  unexpected_failure: "An unexpected error occurred. Please try again later.",
  database_error:
    "Database error. Our team has been notified. Please try again later.",
  auth_error: "Authentication failed. Please try again.",
  network_error: "Network error. Please check your connection and try again.",
  invalid_grant: "Invalid email or password.",
  over_email_send_rate_limit:
    "Too many requests. Please wait a moment and try again.",
  over_request_rate_limit:
    "Too many requests. Please wait a moment and try again.",
};

export function parseAuthError(error: any): AuthError {
  if (!error) return { message: "An unknown error occurred" };

  // Handle Supabase error format
  if (error.message) {
    const message = error.message.toLowerCase();

    // Check for specific error patterns
    if (message.includes("database error")) {
      return {
        message: AUTH_ERRORS.database_error,
        code: "database_error",
      };
    }

    if (message.includes("user already exists")) {
      return {
        message: AUTH_ERRORS.user_already_exists,
        code: "user_already_exists",
      };
    }

    if (message.includes("password")) {
      return {
        message: AUTH_ERRORS.weak_password,
        code: "weak_password",
      };
    }

    if (message.includes("email")) {
      return {
        message: AUTH_ERRORS.invalid_email,
        code: "invalid_email",
      };
    }

    // Default: return original message with safe handling
    return {
      message: AUTH_ERRORS[message] || error.message,
      code: error.code,
      status: error.status,
    };
  }

  return { message: "An unknown error occurred" };
}

export function getErrorCode(error: any): string {
  if (error?.code) return error.code;
  if (error?.message?.includes("database")) return "database_error";
  if (error?.message?.includes("already exists")) return "user_already_exists";
  return "unknown_error";
}
