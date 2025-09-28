import { useState, useEffect } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  FileText,
  Download,
  Trash2,
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
import { useAuth } from "@/contexts/AuthContext";
import apiClient from "@/lib/api";
import toast from "react-hot-toast";
import type { Activity } from "@/types/activity";
import { getFileIcon, formatFileSize } from "@/lib/utils";

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
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);

  const isOwner =
    user && (user.id === activity.owner_id || user.userType === "admin");

  // 添加活动状态检查：只有草稿状态的活动才能编辑附件
  const canEditAttachments = isOwner && activity.status === "draft";

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
    if (!confirm("确定要删除这个附件吗？")) {
      return;
    }

    try {
      await apiClient.delete(
        `/activities/${activity.id}/attachments/${attachmentId}`
      );
      toast.success("附件删除成功");
      fetchAttachments();
    } catch (error) {
      console.error("Failed to delete attachment:", error);
      toast.error("附件删除失败");
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

      // 更新下载次数
      fetchAttachments();
    } catch (error) {
      console.error("Failed to download attachment:", error);
      toast.error("附件下载失败");
    }
  };

  // 获取预览URL

  // 拖拽处理
  const handleDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);

    const files = Array.from(e.dataTransfer.files);
    if (files.length > 0) {
      setDragFiles(files);
      setSelectedFile(files[0]);
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
      for (const file of dragFiles) {
        const formData = new FormData();
        formData.append("file", file);
        formData.append("description", uploadDescription);

        await apiClient.post(`/activities/${activity.id}/attachments`, formData, {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        });
      }

      toast.success(`成功上传 ${dragFiles.length} 个附件`);
      setShowUploadDialog(false);
      setDragFiles([]);
      setSelectedFile(null);
      setUploadDescription("");
      fetchAttachments();
    } catch (error) {
      console.error("Failed to upload multiple attachments:", error);
      toast.error("批量上传失败");
    } finally {
      setUploading(false);
    }
  };

  useEffect(() => {
    fetchAttachments();
  }, [activity.id]);

  // 过滤附件
  const filteredAttachments = attachments.filter((attachment) =>
    attachment.original_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    attachment.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            附件管理
          </CardTitle>
          {canEditAttachments && (
            <Button
              onClick={() => setShowUploadDialog(true)}
              className="bg-blue-600 hover:bg-blue-700"
            >
              <Plus className="h-4 w-4 mr-2" />
              上传附件
            </Button>
          )}
        </div>
      </CardHeader>
      <CardContent>
        {/* 搜索栏 */}
        <div className="mb-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              placeholder="搜索附件..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
        </div>

        {/* 附件列表 */}
        {loading ? (
          <div className="text-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            <p className="mt-2 text-gray-500">加载中...</p>
          </div>
        ) : filteredAttachments.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <FolderOpen className="h-12 w-12 mx-auto mb-4 text-gray-300" />
            <p>暂无附件</p>
          </div>
        ) : (
          <div className="space-y-2">
            {filteredAttachments.map((attachment) => {
              const FileIcon = getFileIcon(attachment.file_category, true);
              return (
                <div
                  key={attachment.id}
                  className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
                >
                  <div className="flex items-center gap-3 flex-1">
                    <FileIcon className="h-5 w-5 text-blue-500" />
                    <div className="flex-1 min-w-0">
                      <p className="font-medium truncate">
                        {attachment.original_name}
                      </p>
                      <p className="text-sm text-gray-500">
                        {formatFileSize(attachment.file_size)} • 上传于{" "}
                        {new Date(attachment.uploaded_at).toLocaleDateString()}
                      </p>
                      {attachment.description && (
                        <p className="text-sm text-gray-600 mt-1">
                          {attachment.description}
                        </p>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge variant="secondary">
                      {attachment.download_count} 次下载
                    </Badge>
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => downloadAttachment(attachment)}
                    >
                      <Download className="h-4 w-4" />
                    </Button>
                    {canEditAttachments && (
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => deleteAttachment(attachment.id)}
                        className="text-red-600 hover:text-red-700"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </CardContent>

      {/* 上传对话框 */}
      <Dialog open={showUploadDialog} onOpenChange={setShowUploadDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>上传附件</DialogTitle>
            <DialogDescription>
              选择要上传的文件，支持拖拽上传
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div
              className={`border-2 border-dashed rounded-lg p-6 text-center transition-colors ${
                isDragOver
                  ? "border-blue-500 bg-blue-50"
                  : "border-gray-300 hover:border-gray-400"
              }`}
              onDragEnter={handleDragEnter}
              onDragLeave={handleDragLeave}
              onDragOver={handleDragOver}
              onDrop={handleDrop}
            >
              {dragFiles.length > 0 ? (
                <div>
                  <p className="font-medium">已选择 {dragFiles.length} 个文件</p>
                  <p className="text-sm text-gray-500 mt-1">
                    {dragFiles.map((file) => file.name).join(", ")}
                  </p>
                </div>
              ) : (
                <div>
                  <FolderOpen className="h-8 w-8 mx-auto mb-2 text-gray-400" />
                  <p>拖拽文件到此处或点击选择</p>
                  <input
                    type="file"
                    onChange={(e) =>
                      setSelectedFile(e.target.files?.[0] || null)
                    }
                    className="hidden"
                    id="file-upload"
                  />
                  <label
                    htmlFor="file-upload"
                    className="cursor-pointer text-blue-600 hover:text-blue-700"
                  >
                    选择文件
                  </label>
                </div>
              )}
            </div>
            <div>
              <label className="block text-sm font-medium mb-1">
                文件描述（可选）
              </label>
              <Input
                value={uploadDescription}
                onChange={(e) => setUploadDescription(e.target.value)}
                placeholder="请输入文件描述"
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowUploadDialog(false)}
            >
              取消
            </Button>
            <Button
              onClick={
                dragFiles.length > 1
                  ? uploadMultipleAttachments
                  : uploadAttachment
              }
              disabled={uploading || (!selectedFile && dragFiles.length === 0)}
            >
              {uploading ? "上传中..." : "上传"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 预览对话框 */}
      <Dialog open={showPreviewDialog} onOpenChange={setShowPreviewDialog}>
        <DialogContent className="max-w-4xl">
          <DialogHeader>
            <DialogTitle>
              预览: {selectedAttachment?.original_name}
            </DialogTitle>
          </DialogHeader>
          <div className="min-h-[400px] flex items-center justify-center">
            {previewLoading ? (
              <div className="text-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
                <p className="mt-2 text-gray-500">加载预览中...</p>
              </div>
            ) : previewUrl ? (
              <iframe
                src={previewUrl}
                className="w-full h-[600px] border rounded"
                title="文件预览"
              />
            ) : (
              <div className="text-center text-gray-500">
                <p>无法预览此文件</p>
              </div>
            )}
          </div>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
