import { useState, useEffect, useRef } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  FileText,
  Download,
  Trash2,
  Eye,
  File,
  Image,
  FileVideo,
  FileAudio,
  Archive,
  Plus,
  Search,
  FolderOpen,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useAuth } from "@/contexts/AuthContext";
import apiClient from "@/lib/api";
import toast from "react-hot-toast";
import type { Activity } from "@/types/activity";

interface ActivityAttachmentsProps {
  activity: Activity;
  onRefresh?: () => void;
}

interface Attachment {
  id: string;
  activity_id: string;
  file_name: string;
  original_name: string;
  file_size: number;
  file_type: string;
  file_category: string;
  uploaded_by: string;
  uploaded_at: string;
  description?: string;
  download_count: number;
  download_url: string;
  uploader?: {
    name: string;
    username: string;
  };
}

// 获取文件类型图标
const getFileIcon = (fileCategory: string) => {
  if (!fileCategory) {
    return <File className="h-4 w-4" />;
  }
  if (fileCategory === "image") {
    return <Image className="h-4 w-4" />;
  } else if (fileCategory === "video") {
    return <FileVideo className="h-4 w-4" />;
  } else if (fileCategory === "audio") {
    return <FileAudio className="h-4 w-4" />;
  } else if (fileCategory === "archive") {
    return <Archive className="h-4 w-4" />;
  } else if (
    fileCategory === "document" ||
    fileCategory === "spreadsheet" ||
    fileCategory === "presentation"
  ) {
    return <FileText className="h-4 w-4" />;
  } else {
    return <File className="h-4 w-4" />;
  }
};

// 格式化文件大小
const formatFileSize = (bytes: number) => {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

// 根据文件名获取文件类别
const getFileCategory = (fileName: string): string => {
  const ext = fileName.toLowerCase().split(".").pop();
  if (!ext) return "other";

  // 添加点号以匹配后端逻辑
  const fileType = "." + ext;

  const imageExts = [".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"];
  const videoExts = [".mp4", ".avi", ".mov", ".wmv", ".flv"];
  const audioExts = [".mp3", ".wav", ".ogg", ".aac"];
  const archiveExts = [".zip", ".rar", ".7z", ".tar", ".gz"];
  const documentExts = [".pdf", ".doc", ".docx", ".txt", ".rtf", ".odt"];
  const spreadsheetExts = [".xls", ".xlsx", ".csv"];
  const presentationExts = [".ppt", ".pptx"];

  if (imageExts.includes(fileType)) return "image";
  if (videoExts.includes(fileType)) return "video";
  if (audioExts.includes(fileType)) return "audio";
  if (archiveExts.includes(fileType)) return "archive";
  if (documentExts.includes(fileType)) return "document";
  if (spreadsheetExts.includes(fileType)) return "spreadsheet";
  if (presentationExts.includes(fileType)) return "presentation";

  return "other";
};

export default function ActivityAttachments({
  activity,
}: ActivityAttachmentsProps) {
  const { user } = useAuth();
  const [attachments, setAttachments] = useState<Attachment[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [showUploadDialog, setShowUploadDialog] = useState(false);
  const [showPreviewDialog, setShowPreviewDialog] = useState(false);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploadDescription, setUploadDescription] = useState("");
  const [uploading, setUploading] = useState(false);
  const [selectedAttachment, setSelectedAttachment] =
    useState<Attachment | null>(null);
  const [isDragOver, setIsDragOver] = useState(false);
  const [dragFiles, setDragFiles] = useState<File[]>([]);

  const isOwner =
    user && (user.user_id === activity.owner_id || user.userType === "admin");
  const canUpload = isOwner;
  const canDelete = isOwner && user.user_id === activity.owner_id;

  // 获取附件列表
  const fetchAttachments = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get(
        `/activities/${activity.id}/attachments`
      );
      setAttachments(response.data.data?.attachments || []);
    } catch (error) {
      console.error("Failed to fetch attachments:", error);
      toast.error("获取附件列表失败");
    } finally {
      setLoading(false);
    }
  };

  // 上传附件
  const uploadAttachment = async () => {
    const fileToUpload = dragFiles.length > 0 ? dragFiles[0] : selectedFile;
    if (!fileToUpload) {
      toast.error("请选择要上传的文件");
      return;
    }

    setUploading(true);
    try {
      const formData = new FormData();
      formData.append("file", fileToUpload);
      formData.append("description", uploadDescription);

      await apiClient.post(`/activities/${activity.id}/attachments`, formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      toast.success("附件上传成功");
      setShowUploadDialog(false);
      setSelectedFile(null);
      setDragFiles([]);
      setUploadDescription("");
      fetchAttachments();
    } catch (error) {
      console.error("Failed to upload attachment:", error);
      toast.error("附件上传失败");
    } finally {
      setUploading(false);
    }
  };

  // 删除附件
  const deleteAttachment = async (attachmentId: string) => {
    try {
      await apiClient.delete(
        `/activities/${activity.id}/attachments/${attachmentId}`
      );
      toast.success("附件删除成功");
      fetchAttachments();
    } catch (error) {
      console.error("Failed to delete attachment:", error);
      toast.error("删除失败");
    }
  };

  // 下载附件
  const downloadAttachment = async (attachment: Attachment) => {
    try {
      const response = await apiClient.get(
        `/activities/${activity.id}/attachments/${attachment.id}/download`,
        {
          responseType: "blob",
        }
      );

      const blob = new Blob([response.data]);
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = attachment.original_name;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);

      toast.success("下载成功");
    } catch (error) {
      console.error("Failed to download attachment:", error);
      toast.error("下载失败");
    }
  };

  // 预览附件
  const previewAttachment = (attachment: Attachment) => {
    setSelectedAttachment(attachment);
    setShowPreviewDialog(true);
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

    const files = Array.from(e.dataTransfer.files);
    if (files.length > 0) {
      setDragFiles(files);
      setShowUploadDialog(true);
    }
  };

  // 批量上传附件
  const uploadMultipleAttachments = async () => {
    if (dragFiles.length === 0) {
      toast.error("请选择要上传的文件");
      return;
    }

    setUploading(true);
    try {
      const formData = new FormData();
      dragFiles.forEach((file) => {
        formData.append("files", file);
      });
      formData.append("description", uploadDescription);

      await apiClient.post(
        `/activities/${activity.id}/attachments/batch`,
        formData,
        {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        }
      );

      toast.success("附件上传成功");
      setShowUploadDialog(false);
      setDragFiles([]);
      setUploadDescription("");
      fetchAttachments();
    } catch (error) {
      console.error("Failed to upload attachments:", error);
      toast.error("附件上传失败");
    } finally {
      setUploading(false);
    }
  };

  // 获取带认证的预览URL
  const getPreviewUrl = (attachmentId: string) => {
    const token = localStorage.getItem("token");
    return `/api/activities/${activity.id}/attachments/${attachmentId}/preview?token=${token}`;
  };

  // 过滤附件
  const filteredAttachments = attachments.filter((attachment) => {
    if (!searchQuery) return true;
    return (
      attachment.original_name
        .toLowerCase()
        .includes(searchQuery.toLowerCase()) ||
      attachment.description?.toLowerCase().includes(searchQuery.toLowerCase())
    );
  });

  // 统计信息
  const stats = {
    total: attachments.length,
    totalSize: attachments.reduce((sum, a) => sum + a.file_size, 0),
    imageCount: attachments.filter((a) => a.file_category === "image").length,
    documentCount: attachments.filter(
      (a) =>
        a.file_category === "document" ||
        a.file_category === "spreadsheet" ||
        a.file_category === "presentation"
    ).length,
  };

  useEffect(() => {
    fetchAttachments();
  }, [activity.id]);

  return (
    <Card className="rounded-xl shadow-lg">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <FolderOpen className="h-5 w-5" />
            附件 ({attachments.length})
          </CardTitle>
          <div className="flex items-center gap-2">
            {canUpload && (
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowUploadDialog(true)}
              >
                <Plus className="h-4 w-4 mr-1" />
                上传
              </Button>
            )}
          </div>
        </div>

        {/* 统计信息 */}
        {attachments.length > 0 && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
            <div className="text-center p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
              <div className="text-2xl font-bold text-blue-600">
                {stats.total}
              </div>
              <div className="text-sm text-muted-foreground">总文件</div>
            </div>
            <div className="text-center p-3 bg-green-50 dark:bg-green-900/20 rounded-lg">
              <div className="text-2xl font-bold text-green-600">
                {formatFileSize(stats.totalSize)}
              </div>
              <div className="text-sm text-muted-foreground">总大小</div>
            </div>
            <div className="text-center p-3 bg-purple-50 dark:bg-purple-900/20 rounded-lg">
              <div className="text-2xl font-bold text-purple-600">
                {stats.imageCount}
              </div>
              <div className="text-sm text-muted-foreground">图片</div>
            </div>
            <div className="text-center p-3 bg-orange-50 dark:bg-orange-900/20 rounded-lg">
              <div className="text-2xl font-bold text-orange-600">
                {stats.documentCount}
              </div>
              <div className="text-sm text-muted-foreground">文档</div>
            </div>
          </div>
        )}
      </CardHeader>

      <CardContent>
        {/* 搜索 */}
        <div className="flex items-center gap-4 mb-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="搜索附件..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        </div>

        {/* 附件表格 */}
        {filteredAttachments.length > 0 ? (
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>文件信息</TableHead>
                  <TableHead>大小</TableHead>
                  <TableHead>上传者</TableHead>
                  <TableHead>上传时间</TableHead>
                  <TableHead className="w-20">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredAttachments.map((attachment) => (
                  <TableRow key={attachment.id}>
                    <TableCell>
                      <div
                        className="flex items-center gap-3 cursor-pointer hover:bg-muted/50 p-2 rounded-lg transition-colors"
                        onClick={() => previewAttachment(attachment)}
                      >
                        <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
                          {getFileIcon(attachment.file_category)}
                        </div>
                        <div>
                          <div className="font-medium">
                            {attachment.original_name}
                          </div>
                          {attachment.description && (
                            <div className="text-sm text-muted-foreground">
                              {attachment.description}
                            </div>
                          )}
                          <div className="text-xs text-muted-foreground">
                            {attachment.file_category || "未知类型"}
                          </div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div
                        className="text-sm cursor-pointer hover:bg-muted/50 p-2 rounded-lg transition-colors"
                        onClick={() => previewAttachment(attachment)}
                      >
                        {formatFileSize(attachment.file_size)}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div
                        className="text-sm cursor-pointer hover:bg-muted/50 p-2 rounded-lg transition-colors"
                        onClick={() => previewAttachment(attachment)}
                      >
                        {attachment.uploader?.name ||
                          attachment.uploader?.username ||
                          attachment.uploaded_by}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div
                        className="cursor-pointer hover:bg-muted/50 p-2 rounded-lg transition-colors"
                        onClick={() => previewAttachment(attachment)}
                      >
                        <div className="text-sm text-muted-foreground">
                          {new Date(
                            attachment.uploaded_at
                          ).toLocaleDateString()}
                        </div>
                        <div className="text-xs text-muted-foreground">
                          {new Date(
                            attachment.uploaded_at
                          ).toLocaleTimeString()}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            downloadAttachment(attachment);
                          }}
                        >
                          <Download className="h-4 w-4" />
                        </Button>
                        {canDelete && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={(e) => {
                              e.stopPropagation();
                              deleteAttachment(attachment.id);
                            }}
                            className="text-red-600 hover:text-red-700 hover:bg-red-50"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        ) : (
          <div className="text-center py-8 text-muted-foreground">
            {searchQuery ? "没有找到匹配的附件" : "暂无附件"}
          </div>
        )}
      </CardContent>

      {/* 上传对话框 */}
      <Dialog open={showUploadDialog} onOpenChange={setShowUploadDialog}>
        <DialogContent className="max-w-2xl rounded-2xl shadow-2xl border-0 bg-white dark:bg-zinc-900">
          <DialogHeader>
            <DialogTitle className="text-2xl font-bold">上传附件</DialogTitle>
            <DialogDescription className="text-base text-muted-foreground">
              选择要上传的文件，支持图片、文档、视频等格式，也可以拖拽文件到下方区域
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-6">
            {/* 拖拽上传区域 */}
            {(() => {
              const fileInputRef = useRef<HTMLInputElement>(null);
              return (
                <div
                  className={`border-2 border-dashed rounded-2xl p-10 text-center transition-all duration-200 ${
                    isDragOver
                      ? "border-blue-500 bg-blue-50 scale-105 shadow-lg"
                      : "border-zinc-200 hover:border-blue-400"
                  }`}
                  onDragEnter={handleDragEnter}
                  onDragLeave={handleDragLeave}
                  onDragOver={handleDragOver}
                  onDrop={handleDrop}
                  onClick={() => fileInputRef.current?.click()}
                  style={{ cursor: "pointer" }}
                >
                  <div className="flex flex-col items-center gap-4">
                    <FolderOpen className="h-16 w-16 text-blue-400" />
                    <div>
                      <p className="text-xl font-semibold text-blue-700">
                        {isDragOver ? "释放文件以上传" : "拖拽或点击上传文件"}
                      </p>
                      <p className="text-sm text-zinc-400 mt-1">
                        支持图片、文档、视频等格式，最大文件大小 50MB
                      </p>
                    </div>
                    <div className="flex items-center gap-2 my-2">
                      <div className="h-px bg-zinc-200 flex-1"></div>
                      <span className="text-xs text-zinc-400"></span>
                      <div className="h-px bg-zinc-200 flex-1"></div>
                    </div>
                    <input
                      ref={fileInputRef}
                      type="file"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        if (file) {
                          setDragFiles([file]);
                        }
                      }}
                      accept="image/*,video/*,audio/*,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.zip,.rar"
                      className="hidden"
                    />
                  </div>
                </div>
              );
            })()}

            {/* 文件列表 */}
            {dragFiles.length > 0 && (
              <div className="space-y-2">
                <label className="text-sm font-medium text-zinc-600">
                  待上传文件
                </label>
                <div className="space-y-2 max-h-40 overflow-y-auto">
                  {dragFiles.map((file, index) => (
                    <div
                      key={index}
                      className="flex items-center gap-3 p-3 bg-zinc-50 dark:bg-zinc-800 rounded-lg border border-zinc-100 dark:border-zinc-700"
                    >
                      <div className="w-8 h-8 bg-blue-100 dark:bg-blue-900 rounded-full flex items-center justify-center">
                        {getFileIcon(getFileCategory(file.name))}
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
                        onClick={() => {
                          setDragFiles(dragFiles.filter((_, i) => i !== index));
                        }}
                        className="text-red-600 hover:text-red-700"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div>
              <label className="text-sm font-medium text-zinc-600">
                文件描述（可选）
              </label>
              <Input
                value={uploadDescription}
                onChange={(e) => setUploadDescription(e.target.value)}
                placeholder="请输入文件描述"
                className="mt-2 rounded-lg border-zinc-200 focus:border-blue-400 focus:ring-2 focus:ring-blue-100"
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              className="rounded-lg border-zinc-200"
              onClick={() => {
                setShowUploadDialog(false);
                setDragFiles([]);
                setUploadDescription("");
              }}
            >
              取消
            </Button>
            <Button
              className="rounded-lg bg-blue-600 hover:bg-blue-700 text-white shadow"
              onClick={
                dragFiles.length > 1
                  ? uploadMultipleAttachments
                  : uploadAttachment
              }
              disabled={(dragFiles.length === 0 && !selectedFile) || uploading}
            >
              {uploading ? "上传中..." : "上传"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 预览对话框 */}
      <Dialog open={showPreviewDialog} onOpenChange={setShowPreviewDialog}>
        <DialogContent className="max-w-[1200px] min-h-[700px]">
          <DialogHeader>
            <DialogTitle>文件预览</DialogTitle>
            <DialogDescription>
              {selectedAttachment?.original_name}
            </DialogDescription>
          </DialogHeader>

          {selectedAttachment && (
            <div className="space-y-4">
              {/* 移除文件基本信息，仅保留预览区域 */}
              <div className="border rounded-lg p-4 min-h-[600px] flex items-center justify-center">
                {selectedAttachment.file_category === "image" ? (
                  <img
                    src={getPreviewUrl(selectedAttachment.id)}
                    alt={selectedAttachment.original_name}
                    className="max-w-full max-h-[650px] object-contain"
                  />
                ) : selectedAttachment.file_category === "video" ? (
                  <video
                    src={getPreviewUrl(selectedAttachment.id)}
                    controls
                    className="max-w-full max-h-[650px]"
                  >
                    您的浏览器不支持视频播放
                  </video>
                ) : selectedAttachment.file_category === "audio" ? (
                  <audio
                    src={getPreviewUrl(selectedAttachment.id)}
                    controls
                    className="w-full"
                  >
                    您的浏览器不支持音频播放
                  </audio>
                ) : selectedAttachment.file_category === "document" &&
                  selectedAttachment.file_type === ".pdf" ? (
                  <iframe
                    src={getPreviewUrl(selectedAttachment.id)}
                    className="w-full h-[650px] border-0"
                    title={selectedAttachment.original_name}
                  />
                ) : selectedAttachment.file_category === "document" &&
                  selectedAttachment.file_type === ".txt" ? (
                  <iframe
                    src={getPreviewUrl(selectedAttachment.id)}
                    className="w-full h-[650px] border-0 bg-white"
                    title={selectedAttachment.original_name}
                  />
                ) : (
                  <div className="text-center text-muted-foreground">
                    <FileText className="h-12 w-12 mx-auto mb-2" />
                    <p>此文件类型不支持预览</p>
                    <Button
                      variant="outline"
                      className="mt-2"
                      onClick={() => downloadAttachment(selectedAttachment)}
                    >
                      <Download className="h-4 w-4 mr-2" />
                      下载查看
                    </Button>
                  </div>
                )}
              </div>
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowPreviewDialog(false)}
            >
              关闭
            </Button>
            {selectedAttachment && (
              <Button onClick={() => downloadAttachment(selectedAttachment)}>
                <Download className="h-4 w-4 mr-2" />
                下载
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
