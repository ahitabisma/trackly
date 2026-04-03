"use client";

import { Toast } from "./toast";
import type { ToastProps } from "./toast";

interface ToastContainerProps {
  toasts: ToastProps[];
  onClose: (id: string) => void;
}

export function ToastContainer({ toasts, onClose }: ToastContainerProps) {
  return (
    <div className="fixed top-4 right-4 z-50 flex flex-col gap-3 pointer-events-none">
      {toasts.map((toast) => (
        <div
          key={toast.id}
          className="pointer-events-auto animate-out slide-out-to-top-2 fade-out-0 duration-300 data-[state=closed]:animate-out"
        >
          <Toast {...toast} onClose={onClose} />
        </div>
      ))}
    </div>
  );
}
