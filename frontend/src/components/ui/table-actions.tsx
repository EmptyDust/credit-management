import { Button } from "./button";
import { Edit, Trash, Eye, Download, Upload } from "lucide-react";

interface TableAction {
  icon: React.ComponentType<any>;
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

// 预定义的操作组合
export const createEditDeleteActions = (
  onEdit: () => void,
  onDelete: () => void,
  canEdit: boolean = true,
  canDelete: boolean = true
): TableAction[] => [
  ...(canEdit ? [{
    icon: Edit,
    label: "编辑",
    onClick: onEdit,
    variant: "ghost" as const,
  }] : []),
  ...(canDelete ? [{
    icon: Trash,
    label: "删除",
    onClick: onDelete,
    variant: "ghost" as const,
    className: "text-red-600 hover:text-red-700",
  }] : []),
];

export const createViewEditDeleteActions = (
  onView: () => void,
  onEdit: () => void,
  onDelete: () => void,
  canView: boolean = true,
  canEdit: boolean = true,
  canDelete: boolean = true
): TableAction[] => [
  ...(canView ? [{
    icon: Eye,
    label: "查看",
    onClick: onView,
    variant: "ghost" as const,
  }] : []),
  ...(canEdit ? [{
    icon: Edit,
    label: "编辑",
    onClick: onEdit,
    variant: "ghost" as const,
  }] : []),
  ...(canDelete ? [{
    icon: Trash,
    label: "删除",
    onClick: onDelete,
    variant: "ghost" as const,
    className: "text-red-600 hover:text-red-700",
  }] : []),
]; 