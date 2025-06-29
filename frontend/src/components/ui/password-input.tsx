import { useState, forwardRef } from "react";
import type { InputHTMLAttributes } from "react";
import { Input } from "@/components/ui/input";
import { Eye, EyeOff } from "lucide-react";

export interface PasswordInputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

export const PasswordInput = forwardRef<HTMLInputElement, PasswordInputProps>(
  ({ label, error, ...props }, ref) => {
    const [show, setShow] = useState(false);
    return (
      <div className="w-full">
        {label && <label className="block mb-1 text-sm font-medium">{label}</label>}
        <div className="relative">
          <Input
            ref={ref}
            type={show ? "text" : "password"}
            {...props}
            className={
              "pr-10 " +
              (props.className || "") +
              (error ? " border-red-500 focus:border-red-500" : "")
            }
          />
          <button
            type="button"
            tabIndex={-1}
            onClick={() => setShow((v) => !v)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            disabled={props.disabled}
          >
            {show ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </button>
        </div>
        {error && <div className="text-xs text-red-500 mt-1">{error}</div>}
      </div>
    );
  }
);
PasswordInput.displayName = "PasswordInput"; 