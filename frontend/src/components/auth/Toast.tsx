"use client";

import { useEffect } from "react";
import { CheckCircle2Icon, AlertCircleIcon, XIcon } from "lucide-react";

interface ToastProps {
  message: string;
  type: "error" | "success";
  onClose: () => void;
  duration?: number;
}

export default function Toast({
  message,
  type,
  onClose,
  duration = 3500,
}: ToastProps) {
  useEffect(() => {
    const timer = setTimeout(onClose, duration);
    return () => clearTimeout(timer);
  }, [onClose, duration]);

  const bgColor = type === "error" ? "bg-red-600" : "bg-green-600";
  const Icon = type === "error" ? AlertCircleIcon : CheckCircle2Icon;

  return (
    <div className="fixed bottom-6 left-1/2 -translate-x-1/2 z-50 animate-in fade-in slide-in-from-bottom-4">
      <div
        className={`flex items-center gap-3 px-4 py-3 rounded-lg text-white shadow-lg ${bgColor}`}
      >
        <Icon size={20} className="flex-shrink-0" />
        <p className="text-sm font-medium">{message}</p>
        <button
          onClick={onClose}
          className="ml-2 flex-shrink-0 hover:opacity-80 transition-opacity"
        >
          <XIcon size={18} />
        </button>
      </div>
    </div>
  );
}
