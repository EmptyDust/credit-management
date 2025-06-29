import React, { useState, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import { Input } from "@/components/ui/input";
import {
  Upload,
  X,
  CheckCircle,
  AlertCircle,
  Loader2,
} from "lucide-react";
import { getFileIcon, formatFileSize } from "@/lib/utils";

interface UploadProgress {
  fileName: string;
  progress: number;
  status: "pending" | "uploading" | "success" | "error";
  error?: string;
}

interface BatchUploadProps {
  onUpload: (files: File[], description: string) => Promise<any>;
  maxFiles?: number;
  maxFileSize?: number; // in bytes
  acceptedTypes?: string[];
  className?: string;
  disabled?: boolean;
}

export function BatchUpload({
  onUpload,
  maxFiles = 10,
  maxFileSize = 50 * 1024 * 1024, // 50MB
  acceptedTypes = [
    "image/*",
    "video/*",
    "audio/*",
    ".pdf",
    ".doc",
    ".docx",
    ".xls",
    ".xlsx",
    ".ppt",
    ".pptx",
    ".txt",
    ".zip",
    ".rar",
  ],
  className = "",
  disabled = false,
}: BatchUploadProps) {
  const [files, setFiles] = useState<File[]>([]);
  const [description, setDescription] = useState("");
  const [uploading, setUploading] = useState(false);
  const [isDragOver, setIsDragOver] = useState(false);
  const [uploadProgress, setUploadProgress] = useState<UploadProgress[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // 验证文件
  const validateFile = (file: File): string | null => {
    if (file.size > maxFileSize) {
      return `文件大小不能超过${formatFileSize(
        maxFileSize
      )}，当前文件大小：${formatFileSize(file.size)}`;
    }

    const isValidType = acceptedTypes.some(
      (type) =>
        file.type.startsWith(type) || file.name.toLowerCase().endsWith(type)
    );

    if (!isValidType) {
      return `不支持的文件类型：${file.type || "未知类型"}`;
    }

    return null;
  };

  // 处理文件选择
  const handleFileSelect = (selectedFiles: FileList | null) => {
    if (!selectedFiles) return;

    const fileArray = Array.from(selectedFiles);
    const validFiles: File[] = [];
    const errors: string[] = [];

    // 检查文件数量限制
    if (files.length + fileArray.length > maxFiles) {
      errors.push(`文件数量不能超过${maxFiles}个`);
    }

    fileArray.forEach((file) => {
      const error = validateFile(file);
      if (error) {
        errors.push(`${file.name}: ${error}`);
      } else {
        validFiles.push(file);
      }
    });

    if (errors.length > 0) {
      console.error("文件验证失败：", errors);
    }

    if (validFiles.length > 0) {
      setFiles((prev) => [...prev, ...validFiles]);
    }
  };

  // 拖拽事件处理
  const handleDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragOver(false);

    const droppedFiles = Array.from(e.dataTransfer.files);
    if (droppedFiles.length > 0) {
      handleFileSelect(e.dataTransfer.files);
    }
  };

  // 移除文件
  const removeFile = (index: number) => {
    setFiles((prev) => prev.filter((_, i) => i !== index));
  };

  // 上传文件
  const handleUpload = async () => {
    if (files.length === 0) {
      console.error("请选择要上传的文件");
      return;
    }

    setUploading(true);

    // 初始化上传进度
    const initialProgress: UploadProgress[] = files.map((file) => ({
      fileName: file.name,
      progress: 0,
      status: "pending",
    }));
    setUploadProgress(initialProgress);

    try {
      await onUpload(files, description);

      // 上传成功，清空状态
      setFiles([]);
      setDescription("");
      setUploadProgress([]);
    } catch (error) {
      console.error("上传失败:", error);

      // 更新所有文件状态为错误
      setUploadProgress((prev) =>
        prev.map((item) => ({
          ...item,
          status: "error",
          error: "上传失败",
        }))
      );
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className={`space-y-4 ${className}`}>
      {/* 拖拽区域 */}
      <div
        className={`border-2 border-dashed rounded-xl p-8 text-center transition-all duration-200 ${
          isDragOver
            ? "border-blue-400 bg-blue-50 dark:bg-blue-900/20"
            : "border-zinc-200 hover:border-blue-400"
        } ${disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer"}`}
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDragOver={handleDragOver}
        onDrop={handleDrop}
        onClick={() => !disabled && fileInputRef.current?.click()}
      >
        <div className="flex flex-col items-center gap-4">
          <Upload className="h-16 w-16 text-blue-400" />
          <div>
            <p className="text-xl font-semibold text-blue-700">
              {isDragOver ? "释放文件以上传" : "拖拽或点击上传文件"}
            </p>
            <p className="text-sm text-zinc-400 mt-1">
              支持图片、文档、视频等格式，最大文件大小{" "}
              {formatFileSize(maxFileSize)}
            </p>
            <p className="text-xs text-zinc-400 mt-1">
              支持多文件选择，最多{maxFiles}个文件
            </p>
          </div>
          <div className="flex items-center gap-2 my-2">
            <div className="h-px bg-zinc-200 flex-1"></div>
            <span className="text-xs text-zinc-400">或</span>
            <div className="h-px bg-zinc-200 flex-1"></div>
          </div>
          <Button variant="outline" size="sm" disabled={disabled}>
            选择文件
          </Button>
          <input
            ref={fileInputRef}
            type="file"
            multiple
            onChange={(e) => handleFileSelect(e.target.files)}
            accept={acceptedTypes.join(",")}
            className="hidden"
            disabled={disabled}
          />
        </div>
      </div>

      {/* 文件列表 */}
      {files.length > 0 && (
        <div className="space-y-2">
          <label className="text-sm font-medium text-zinc-600">
            待上传文件 ({files.length})
          </label>
          <div className="space-y-2 max-h-40 overflow-y-auto">
            {files.map((file, index) => (
              <div
                key={index}
                className="flex items-center gap-3 p-3 bg-zinc-50 dark:bg-zinc-800 rounded-lg border border-zinc-100 dark:border-zinc-700"
              >
                <div className="w-8 h-8 bg-blue-100 dark:bg-blue-900 rounded-full flex items-center justify-center">
                  {React.createElement(getFileIcon(file.name))}
                </div>
                <div className="flex-1">
                  <div className="font-medium text-sm">{file.name}</div>
                  <div className="text-xs text-zinc-400">
                    {formatFileSize(file.size)}
                  </div>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => removeFile(index)}
                  className="text-red-600 hover:text-red-700"
                  disabled={uploading}
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 上传进度 */}
      {uploadProgress.length > 0 && (
        <div className="space-y-2">
          <label className="text-sm font-medium text-zinc-600">上传进度</label>
          <div className="space-y-2">
            {uploadProgress.map((item, index) => (
              <div key={index} className="space-y-1">
                <div className="flex items-center justify-between text-sm">
                  <span className="truncate flex-1">{item.fileName}</span>
                  <div className="flex items-center gap-2">
                    {item.status === "pending" && (
                      <Loader2 className="h-4 w-4 animate-spin" />
                    )}
                    {item.status === "success" && (
                      <CheckCircle className="h-4 w-4 text-green-600" />
                    )}
                    {item.status === "error" && (
                      <AlertCircle className="h-4 w-4 text-red-600" />
                    )}
                  </div>
                </div>
                <Progress value={item.progress} className="h-2" />
                {item.error && (
                  <p className="text-xs text-red-600">{item.error}</p>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {/* 文件描述 */}
      <div>
        <label className="text-sm font-medium text-zinc-600">
          文件描述（可选）
        </label>
        <Input
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="请输入文件描述"
          className="mt-2 rounded-lg border-zinc-200 focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
          disabled={uploading}
        />
      </div>

      {/* 上传按钮 */}
      <div className="flex justify-end gap-2">
        <Button
          variant="outline"
          onClick={() => {
            setFiles([]);
            setDescription("");
            setUploadProgress([]);
          }}
          disabled={uploading}
        >
          清空
        </Button>
        <Button
          onClick={handleUpload}
          disabled={files.length === 0 || uploading || disabled}
        >
          {uploading ? (
            <>
              <Loader2 className="h-4 w-4 mr-2 animate-spin" />
              上传中...
            </>
          ) : (
            <>
              <Upload className="h-4 w-4 mr-2" />
              {files.length > 1 ? "批量上传" : "上传"}
            </>
          )}
        </Button>
      </div>
    </div>
  );
}
