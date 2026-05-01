import type { ReactNode } from "react";
import { Button } from "./button";

type ButtonVariant =
  | "default"
  | "destructive"
  | "outline"
  | "secondary"
  | "ghost"
  | "link";

type HeaderActionIcon = React.ComponentType<{ className?: string }>;

interface HeaderActionConfig {
  label: string;
  onClick: () => void;
  icon?: HeaderActionIcon;
  disabled?: boolean;
  variant?: ButtonVariant;
}

export const createPageActions = (
  primaryAction?: HeaderActionConfig,
  secondaryActions?: HeaderActionConfig[],
): ReactNode => {
  return (
    <>
      {primaryAction && (
        <Button
          onClick={primaryAction.onClick}
          disabled={primaryAction.disabled}
          className="rounded-lg shadow transition-all duration-200 hover:scale-105"
        >
          {primaryAction.icon && (
            <primaryAction.icon className="mr-2 h-4 w-4" />
          )}
          {primaryAction.label}
        </Button>
      )}
      {secondaryActions?.map((action, index) => (
        <Button
          key={index}
          onClick={action.onClick}
          variant={action.variant || "outline"}
          disabled={action.disabled}
          className="rounded-lg shadow transition-all duration-200 hover:scale-105"
        >
          {action.icon && <action.icon className="mr-2 h-4 w-4" />}
          {action.label}
        </Button>
      ))}
    </>
  );
};
