import React from "react";
import { Badge } from "@/components/ui/badge";
import { CheckCircle, XCircle, Clock, AlertCircle } from "lucide-react";
import type { ComponentType } from "react";

// 状态配置类型
export type StatusType = 
  | "draft" 
  | "pending_review" 
  | "pending"
  | "approved" 
  | "rejected" 
  | "unsubmitted"
  | "active"
  | "inactive"
  | "suspended";

// 统一的状态配置
const STATUS_CONFIG = {
  // 活动/申请状态
  draft: {
    text: "草稿",
    style: "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200",
    icon: Clock,
    color: "gray"
  },
  pending_review: {
    text: "待审核",
    style: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
    icon: AlertCircle,
    color: "yellow"
  },
  pending: {
    text: "待审核",
    style: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200",
    icon: AlertCircle,
    color: "yellow"
  },
  approved: {
    text: "已通过",
    style: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
    icon: CheckCircle,
    color: "green"
  },
  rejected: {
    text: "已拒绝",
    style: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
    icon: XCircle,
    color: "red"
  },
  unsubmitted: {
    text: "未提交",
    style: "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200",
    icon: Clock,
    color: "gray"
  },
  // 用户状态
  active: {
    text: "活跃",
    style: "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200",
    icon: CheckCircle,
    color: "green"
  },
  inactive: {
    text: "停用",
    style: "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200",
    icon: Clock,
    color: "gray"
  },
  suspended: {
    text: "暂停",
    style: "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200",
    icon: XCircle,
    color: "red"
  }
} as const;

// 获取状态文本
export const getStatusText = (status: string): string => {
  const config = STATUS_CONFIG[status as StatusType];
  return config?.text || status || "未知";
};

// 获取状态样式类名
export const getStatusStyle = (status: string): string => {
  const config = STATUS_CONFIG[status as StatusType];
  return config?.style || "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200";
};

// 获取状态图标组件
export const getStatusIcon = (status: string): ComponentType<any> => {
  const config = STATUS_CONFIG[status as StatusType];
  return config?.icon || Clock;
};

// 获取状态颜色
export const getStatusColor = (status: string): string => {
  const config = STATUS_CONFIG[status as StatusType];
  return config?.color || "gray";
};

// 获取状态徽章组件
export const getStatusBadge = (status: string): React.ReactElement => {
  const config = STATUS_CONFIG[status as StatusType];
  
  if (config) {
    return React.createElement(Badge, {
      variant: "default",
      className: config.style,
      children: config.text
    });
  }
  
  return React.createElement(Badge, {
    variant: "outline",
    className: "text-xs",
    children: status || "未知"
  });
};

// 获取状态配置对象
export const getStatusConfig = (status: string) => {
  return STATUS_CONFIG[status as StatusType] || {
    text: status || "未知",
    style: "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200",
    icon: Clock,
    color: "gray"
  };
};

// 检查状态是否为活跃状态
export const isActiveStatus = (status: string): boolean => {
  return status === "active" || status === "approved";
};

// 检查状态是否为待处理状态
export const isPendingStatus = (status: string): boolean => {
  return status === "pending" || status === "pending_review";
};

// 检查状态是否为拒绝状态
export const isRejectedStatus = (status: string): boolean => {
  return status === "rejected" || status === "suspended";
};

// 检查状态是否为草稿状态
export const isDraftStatus = (status: string): boolean => {
  return status === "draft" || status === "unsubmitted";
}; 