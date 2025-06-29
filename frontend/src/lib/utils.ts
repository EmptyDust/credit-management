import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"
import {
  File,
  Image,
  FileVideo,
  FileAudio,
  Archive,
  FileText,
} from "lucide-react";
import type { ComponentType } from "react";
// 导入统一的状态处理函数
export { 
  getStatusText, 
  getStatusStyle, 
  getStatusIcon, 
  getStatusBadge,
  getStatusColor,
  getStatusConfig,
  isActiveStatus,
  isPendingStatus,
  isRejectedStatus,
  isDraftStatus,
  type StatusType
} from "./status-utils";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

/**
 * 获取文件图标
 * @param filenameOrCategory 文件名或文件类别
 * @param isCategory 是否直接传入类别
 */
export const getFileIcon = (filenameOrCategory: string, isCategory = false): ComponentType<any> => {
  let category = filenameOrCategory;
  if (!isCategory) {
    // 通过扩展名推断类别
    const ext = filenameOrCategory.split(".").pop()?.toLowerCase();
    const fileType = "." + ext;
    const imageExts = [".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"];
    const videoExts = [".mp4", ".avi", ".mov", ".wmv", ".flv"];
    const audioExts = [".mp3", ".wav", ".ogg", ".aac"];
    const archiveExts = [".zip", ".rar", ".7z", ".tar", ".gz"];
    const documentExts = [".pdf", ".doc", ".docx", ".txt", ".rtf", ".odt"];
    const spreadsheetExts = [".xls", ".xlsx", ".csv"];
    const presentationExts = [".ppt", ".pptx"];
    if (imageExts.includes(fileType)) category = "image";
    else if (videoExts.includes(fileType)) category = "video";
    else if (audioExts.includes(fileType)) category = "audio";
    else if (archiveExts.includes(fileType)) category = "archive";
    else if (documentExts.includes(fileType)) category = "document";
    else if (spreadsheetExts.includes(fileType)) category = "spreadsheet";
    else if (presentationExts.includes(fileType)) category = "presentation";
    else category = "other";
  }
  if (!category) return File;
  if (category === "image") return Image;
  if (category === "video") return FileVideo;
  if (category === "audio") return FileAudio;
  if (category === "archive") return Archive;
  if (["document", "spreadsheet", "presentation"].includes(category)) return FileText;
  return File;
};

// 格式化文件大小
export const formatFileSize = (bytes: number) => {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

// 活动类别配置
export const activityCategories = [
  { value: "创新创业实践活动", label: "创新创业实践活动" },
  { value: "学科竞赛", label: "学科竞赛" },
  { value: "大学生创业项目", label: "大学生创业项目" },
  { value: "创业实践项目", label: "创业实践项目" },
  { value: "论文专利", label: "论文专利" },
];

// 活动详情配置
export const activityDetailConfigs = {
  "创新创业实践活动": {
    icon: "Lightbulb",
    color: "text-yellow-600",
    title: "创新创业详情",
    fields: [
      { key: "item", label: "实践事项", type: "text" },
      { key: "company", label: "实习公司", type: "text" },
      { key: "project_no", label: "课题编号", type: "text" },
      { key: "issuer", label: "发证机构", type: "text" },
      { key: "date", label: "实践日期", type: "date" },
      { key: "total_hours", label: "累计学时", type: "number" },
    ],
  },
  "学科竞赛": {
    icon: "Trophy",
    color: "text-yellow-600",
    title: "学科竞赛详情",
    fields: [
      { key: "competition", label: "竞赛名称", type: "text" },
      { key: "level", label: "竞赛级别", type: "text" },
      { key: "award_level", label: "获奖等级", type: "text" },
      { key: "rank", label: "排名", type: "text" },
    ],
  },
  "大学生创业项目": {
    icon: "Building2",
    color: "text-blue-600",
    title: "创业项目详情",
    fields: [
      { key: "project_name", label: "项目名称", type: "text" },
      { key: "project_level", label: "项目级别", type: "text" },
      { key: "project_rank", label: "项目排名", type: "text" },
    ],
  },
  "创业实践项目": {
    icon: "Building2",
    color: "text-green-600",
    title: "创业实践详情",
    fields: [
      { key: "company_name", label: "公司名称", type: "text" },
      { key: "legal_person", label: "法人代表", type: "text" },
      { key: "share_percent", label: "持股比例", type: "number" },
    ],
  },
  "论文专利": {
    icon: "FileText",
    color: "text-purple-600",
    title: "论文专利详情",
    fields: [
      { key: "name", label: "名称", type: "text" },
      { key: "category", label: "类别", type: "text" },
      { key: "rank", label: "排名", type: "text" },
    ],
  },
}; 