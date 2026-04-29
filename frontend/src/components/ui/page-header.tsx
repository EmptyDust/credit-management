import type { ReactNode } from "react";
import { Button } from "./button";

interface PageHeaderProps {
  title: string;
  description?: string;
  actions?: ReactNode;
  className?: string;
}

export function PageHeader({ 
  title, 
  description, 
  actions, 
  className = "" 
}: PageHeaderProps) {
  return (
    <div className={`flex flex-col md:flex-row md:justify-between md:items-center gap-4 ${className}`}>
      <div>
        <h1 className="text-3xl font-bold tracking-tight">{title}</h1>
        {description && (
          <p className="text-muted-foreground">{description}</p>
        )}
      </div>
      {actions && (
        <div className="flex items-center gap-2">
          {actions}
        </div>
      )}
    </div>
  );
}

// 预定义的操作按钮组合
export const createPageActions = (
  primaryAction?: {
    label: string;
    onClick: () => void;
    icon?: React.ComponentType<any>;
    disabled?: boolean;
  },
  secondaryActions?: Array<{
    label: string;
    onClick: () => void;
    icon?: React.ComponentType<any>;
    variant?: "default" | "destructive" | "outline" | "secondary" | "ghost" | "link";
    disabled?: boolean;
  }>
): ReactNode => {
  return (
    <>
      {primaryAction && (
        <Button
          onClick={primaryAction.onClick}
          disabled={primaryAction.disabled}
          className="rounded-lg shadow transition-all duration-200 hover:scale-105"
        >
          {primaryAction.icon && <primaryAction.icon className="mr-2 h-4 w-4" />}
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