import AuthLayout from '@/components/auth/AuthLayout';
import SignInForm from '@/components/auth/SignInForm';

export const metadata = {
  title: 'Sign In - Trackly',
  description: 'Sign in to your Trackly account',
};

export default function SignInPage() {
  return (
    <AuthLayout
      title="Welcome back"
      subtitle="Sign in to your account to continue"
    >
      <SignInForm />
    </AuthLayout>
  );
}
