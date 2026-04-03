import { createClient } from "@/lib/supabase/client";
import { parseAuthError, getErrorCode } from "@/utils/auth-errors";

export class AuthService {
  // ── CLIENT-SIDE METHODS ─────────────────────────────

  static async signUp(email: string, password: string, name?: string) {
    try {
      const supabase = createClient();
      const { data, error } = await supabase.auth.signUp({
        email,
        password,
      });

      if (error) throw error;

      // Update user profile with name if provided
      if (name && data.user) {
        const { error: updateError } = await supabase
          .from("users")
          .update({ name })
          .eq("id", data.user.id);

        if (updateError) {
          console.warn("Failed to update user name:", updateError);
          // Don't throw - name is optional
        }
      }

      return { data, error: null };
    } catch (error: any) {
      console.error("SignUp Error:", error);
      const parsedError = parseAuthError(error);
      return { data: null, error: parsedError };
    }
  }

  static async signIn(email: string, password: string) {
    try {
      const supabase = createClient();
      const { data, error } = await supabase.auth.signInWithPassword({
        email,
        password,
      });

      if (error) throw error;
      return { data, error: null };
    } catch (error: any) {
      console.error("SignIn Error:", error);
      const parsedError = parseAuthError(error);
      return { data: null, error: parsedError };
    }
  }

  static async signInWithGoogle() {
    try {
      const supabase = createClient();
      const { data, error } = await supabase.auth.signInWithOAuth({
        provider: "google",
        options: {
          redirectTo: `${window.location.origin}/auth/callback`,
        },
      });

      if (error) throw error;
      return { data, error: null };
    } catch (error: any) {
      console.error("OAuth Error:", error);
      const parsedError = parseAuthError(error);
      return { data: null, error: parsedError };
    }
  }

  static async signOut() {
    try {
      const supabase = createClient();
      const { error } = await supabase.auth.signOut();

      if (error) throw error;
      return { error: null };
    } catch (error: any) {
      console.error("SignOut Error:", error);
      const parsedError = parseAuthError(error);
      return { error: parsedError };
    }
  }

  static async getSession() {
    try {
      const supabase = createClient();
      const { data, error } = await supabase.auth.getSession();

      if (error) throw error;
      return { data, error: null };
    } catch (error: any) {
      console.error("GetSession Error:", error);
      const parsedError = parseAuthError(error);
      return { data: null, error: parsedError };
    }
  }

  static async getUser() {
    try {
      const supabase = createClient();
      const { data, error } = await supabase.auth.getUser();

      if (error) throw error;
      return { data, error: null };
    } catch (error: any) {
      console.error("GetUser Error:", error);
      const parsedError = parseAuthError(error);
      return { data: null, error: parsedError };
    }
  }

  static async refreshSession() {
    try {
      const supabase = createClient();
      const { data, error } = await supabase.auth.refreshSession();

      if (error) throw error;
      return { data, error: null };
    } catch (error: any) {
      console.error("RefreshSession Error:", error);
      const parsedError = parseAuthError(error);
      return { data: null, error: parsedError };
    }
  }
}
