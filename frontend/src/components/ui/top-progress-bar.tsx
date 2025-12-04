import React from "react";
import { cn } from "@/lib/utils";

interface TopProgressBarProps {
  active: boolean;
  className?: string;
}

export const TopProgressBar: React.FC<TopProgressBarProps> = ({
  active,
  className,
}) => {
  if (!active) return null;

  return (
    <div
      className={cn(
        "mb-4 h-1 w-full overflow-hidden rounded-full bg-gray-200",
        className
      )}
    >
      <div className="h-full w-1/3 animate-progressSlide bg-gradient-to-r from-blue-400 via-blue-500 to-blue-400" />
    </div>
  );
};


