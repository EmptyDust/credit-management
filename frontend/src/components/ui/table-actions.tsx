import { Button } from "./button";

interface TableAction {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  onClick: () => void;
  variant?: "default" | "destructive" | "outline" | "secondary" | "ghost" | "link";
  size?: "default" | "sm" | "lg" | "icon";
  className?: string;
  disabled?: boolean;
}

interface TableActionsProps {
  actions: TableAction[];
  className?: string;
}

export function TableActions({ actions, className = "" }: TableActionsProps) {
  return (
    <div className={`flex items-center gap-2 ${className}`}>
      {actions.map((action, index) => (
        <Button
          key={index}
          variant={action.variant || "ghost"}
          size={action.size || "sm"}
          onClick={action.onClick}
          className={action.className}
          disabled={action.disabled}
          title={action.label}
        >
          <action.icon className="h-4 w-4" />
        </Button>
      ))}
    </div>
  );
}
