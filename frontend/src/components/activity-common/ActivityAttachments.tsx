import { useState, useEffect } from "react";
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
  onRefresh,
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

  const isOwner =
    user && (user.user_id === activity.owner_id || user.userType === "admin");
  const canUpload =
    user &&
    (user.userType === "teacher" ||
      user.userType === "admin" ||
      user.user_id === activity.owner_id);

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
    if (!selectedFile) {
      toast.error("请选择要上传的文件");
      return;
    }

    setUploading(true);
    try {
      const formData = new FormData();
      formData.append("file", selectedFile);
      formData.append("description", uploadDescription);

      await apiClient.post(`/activities/${activity.id}/attachments`, formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      toast.success("附件上传成功");
      setShowUploadDialog(false);
      setSelectedFile(null);
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
                      <div className="flex items-center gap-3">
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
                      <div className="text-sm">
                        {formatFileSize(attachment.file_size)}
                      </div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm">{attachment.uploaded_by}</div>
                    </TableCell>
                    <TableCell>
                      <div className="text-sm text-muted-foreground">
                        {new Date(attachment.uploaded_at).toLocaleDateString()}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {new Date(attachment.uploaded_at).toLocaleTimeString()}
                      </div>
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <FileText className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuLabel>操作</DropdownMenuLabel>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            onClick={() => previewAttachment(attachment)}
                          >
                            <Eye className="h-4 w-4 mr-2" />
                            预览
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => downloadAttachment(attachment)}
                          >
                            <Download className="h-4 w-4 mr-2" />
                            下载
                          </DropdownMenuItem>
                          {isOwner && (
                            <DropdownMenuItem
                              onClick={() => deleteAttachment(attachment.id)}
                              className="text-red-600"
                            >
                              <Trash2 className="h-4 w-4 mr-2" />
                              删除
                            </DropdownMenuItem>
                          )}
                        </DropdownMenuContent>
                      </DropdownMenu>
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
        <DialogContent>
          <DialogHeader>
            <DialogTitle>上传附件</DialogTitle>
            <DialogDescription>
              选择要上传的文件，支持图片、文档、视频等格式
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">选择文件</label>
              <Input
                type="file"
                onChange={(e) => setSelectedFile(e.target.files?.[0] || null)}
                accept="image/*,video/*,audio/*,.pdf,.doc,.docx,.xls,.xlsx,.ppt,.pptx,.txt,.zip,.rar"
              />
            </div>
            <div>
              <label className="text-sm font-medium">文件描述（可选）</label>
              <Input
                value={uploadDescription}
                onChange={(e) => setUploadDescription(e.target.value)}
                placeholder="请输入文件描述"
              />
            </div>
            {selectedFile && (
              <div className="p-3 bg-muted/50 rounded-lg">
                <div className="flex items-center gap-2">
                  {getFileIcon(getFileCategory(selectedFile.name))}
                  <div>
                    <div className="font-medium">{selectedFile.name}</div>
                    <div className="text-sm text-muted-foreground">
                      {formatFileSize(selectedFile.size)}
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowUploadDialog(false)}
            >
              取消
            </Button>
            <Button
              onClick={uploadAttachment}
              disabled={!selectedFile || uploading}
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
