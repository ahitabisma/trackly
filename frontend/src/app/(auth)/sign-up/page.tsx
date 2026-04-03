import AuthLayout from "@/components/auth/AuthLayout";
import SignUpForm from "@/components/auth/SignUpForm";

export const metadata = {
  title: "Sign Up - Trackly",
  description: "Create a new Trackly account",
};

export default function SignUpPage() {
  return (
    <AuthLayout
      title="Create account"
      subtitle="Join us to start tracking your shareholdings"
    >
      <SignUpForm />
    </AuthLayout>
  );
}
