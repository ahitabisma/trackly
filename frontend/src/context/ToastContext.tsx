"use client";

import {
  createContext,
  useContext,
  useState,
  useCallback,
  ReactNode,
} from "react";
import type { ToastProps } from "@/components/ui/toast/toast";

interface ToastContextType {
  toasts: ToastProps[];
  toast: (options: Omit<ToastProps, "id" | "onClose">) => void;
  dismiss: (id?: string) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastProps[]>([]);
  let toastCount = 0;

  const toast = useCallback(
    ({
      message,
      description,
      variant = "default",
      duration = 3000,
    }: Omit<ToastProps, "id" | "onClose">) => {
      const id = `toast-${toastCount++}-${Date.now()}`;

      const dismiss = (dismissId: string) => {
        setToasts((prev) => prev.filter((t) => t.id !== dismissId));
      };

      const newToast: ToastProps = {
        id,
        message,
        description,
        variant,
        duration,
        onClose: dismiss,
      };

      setToasts((prev) => [...prev, newToast]);
    },
    [],
  );

  const dismiss = useCallback((id?: string) => {
    if (id) {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    } else {
      setToasts([]);
    }
  }, []);

  return (
    <ToastContext.Provider value={{ toasts, toast, dismiss }}>
      {children}
    </ToastContext.Provider>
  );
}

export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error("useToast must be used within a ToastProvider");
  }
  return context;
}
