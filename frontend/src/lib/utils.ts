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
// 旧的活动类别与详情配置已迁移到后端配置驱动，前端不再保留硬编码。