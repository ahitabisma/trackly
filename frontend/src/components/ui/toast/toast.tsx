"use client";

import { useEffect } from "react";
import {
  XIcon,
  CheckCircleIcon,
  AlertCircleIcon,
  InfoIcon,
} from "lucide-react";

export interface ToastProps {
  id: string;
  message: string;
  description?: string;
  variant?: "default" | "success" | "error" | "warning" | "info";
  duration?: number;
  onClose: (id: string) => void;
}

export function Toast({
  id,
  message,
  description,
  variant = "default",
  duration = 3000,
  onClose,
}: ToastProps) {
  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => onClose(id), duration);
      return () => clearTimeout(timer);
    }
  }, [id, duration, onClose]);

  const variantStyles = {
    default: {
      bg: "bg-zinc-900",
      border: "border-zinc-700",
      text: "text-zinc-50",
      icon: null,
      accentColor: "text-zinc-400",
    },
    success: {
      bg: "bg-emerald-50",
      border: "border-emerald-200",
      text: "text-emerald-950",
      icon: CheckCircleIcon,
      accentColor: "text-emerald-600",
    },
    error: {
      bg: "bg-red-50",
      border: "border-red-200",
      text: "text-red-950",
      icon: AlertCircleIcon,
      accentColor: "text-red-600",
    },
    warning: {
      bg: "bg-amber-50",
      border: "border-amber-200",
      text: "text-amber-950",
      icon: AlertCircleIcon,
      accentColor: "text-amber-600",
    },
    info: {
      bg: "bg-blue-50",
      border: "border-blue-200",
      text: "text-blue-950",
      icon: InfoIcon,
      accentColor: "text-blue-600",
    },
  };

  const style = variantStyles[variant];
  const IconComponent = style.icon;

  return (
    <div
      className={`
        animate-in slide-in-from-top-2 fade-in-0 duration-300
        ${style.bg} ${style.border} ${style.text}
        border rounded-lg p-4 shadow-lg
        flex items-start gap-3 max-w-md
        backdrop-blur-sm
      `}
      role="alert"
    >
      {IconComponent && (
        <IconComponent
          className={`w-5 h-5 flex-shrink-0 mt-0.5 ${style.accentColor}`}
        />
      )}

      <div className="flex-1">
        <p className="font-medium text-sm">{message}</p>
        {description && (
          <p className={`text-xs mt-1 opacity-90`}>{description}</p>
        )}
      </div>

      <button
        onClick={() => onClose(id)}
        className={`flex-shrink-0 ${style.accentColor} hover:opacity-70 transition-opacity ml-2`}
        aria-label="Close"
      >
        <XIcon className="w-4 h-4" />
      </button>
    </div>
  );
}
