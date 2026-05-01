import { Edit, Eye, Trash } from "lucide-react";

type ActionIcon = React.ComponentType<{ className?: string }>;
type ButtonVariant =
  | "default"
  | "destructive"
  | "outline"
  | "secondary"
  | "ghost"
  | "link";
type ButtonSize = "default" | "sm" | "lg" | "icon";

export interface TableAction {
  icon: ActionIcon;
  label: string;
  onClick: () => void;
  variant?: ButtonVariant;
  size?: ButtonSize;
  className?: string;
  disabled?: boolean;
}

export const createEditDeleteActions = (
  onEdit: () => void,
  onDelete: () => void,
  canEdit = true,
  canDelete = true,
): TableAction[] => [
  ...(canEdit
    ? [
        {
          icon: Edit,
          label: "编辑",
          onClick: onEdit,
          variant: "ghost" as const,
        },
      ]
    : []),
  ...(canDelete
    ? [
        {
          icon: Trash,
          label: "删除",
          onClick: onDelete,
          variant: "ghost" as const,
          className: "text-red-600 hover:text-red-700",
        },
      ]
    : []),
];

export const createViewEditDeleteActions = (
  onView: () => void,
  onEdit: () => void,
  onDelete: () => void,
  canView = true,
  canEdit = true,
  canDelete = true,
): TableAction[] => [
  ...(canView
    ? [
        {
          icon: Eye,
          label: "查看",
          onClick: onView,
          variant: "ghost" as const,
        },
      ]
    : []),
  ...(canEdit
    ? [
        {
          icon: Edit,
          label: "编辑",
          onClick: onEdit,
          variant: "ghost" as const,
        },
      ]
    : []),
  ...(canDelete
    ? [
        {
          icon: Trash,
          label: "删除",
          onClick: onDelete,
          variant: "ghost" as const,
          className: "text-red-600 hover:text-red-700",
        },
      ]
    : []),
];
