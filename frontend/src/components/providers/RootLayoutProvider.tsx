"use client";

import { ToastProvider } from "@/context/ToastContext";
import { ToastContainer } from "@/components/ui/toast/toast-container";
import { useToast } from "@/context/ToastContext";
import { ReactNode } from "react";

function Inner({ children }: { children: ReactNode }) {
  const { toasts, dismiss } = useToast();

  return (
    <>
      {children}
      <ToastContainer toasts={toasts} onClose={dismiss} />
    </>
  );
}

export function RootLayoutProvider({ children }: { children: ReactNode }) {
  return (
    <ToastProvider>
      <Inner>{children}</Inner>
    </ToastProvider>
  );
}
