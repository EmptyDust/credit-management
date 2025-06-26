import { useState, useEffect } from "react";
import * as z from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useAuth } from "@/contexts/AuthContext";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Progress } from "@/components/ui/progress";
import apiClient from "@/lib/api";
import {
  PlusCircle,
  Search,
  Eye,
  FileText,
  Upload,
  Download,
  Filter,
  RefreshCw,
  CheckCircle,
  XCircle,
  Clock,
  AlertCircle,
  Trash2,
  File,
  Image,
  FileVideo,
  FileAudio,
} from "lucide-react";
import toast from "react-hot-toast";

// Types
interface Application {
  id: string;
  affair_id: string;
  affair_name?: string;
  student_number: string;
  student_name?: string;
  submission_time: string;
  status: "unsubmitted" | "pending" | "approved" | "rejected";
  reviewer_id?: string;
  reviewer_name?: string;
  review_comment?: string;
  review_time?: string;
  applied_credits: number;
  approved_credits: number;
  details: string;
  attachments?: string; // JSON string of attachments
}

interface ApplicationAttachment {
  id: string;
  file_name: string;
  original_name: string;
  file_size: number;
  file_type: string;
  file_path: string;
  upload_time: string;
  description: string;
  download_url: string;
}

interface Affair {
  id: string;
  name: string;
  description?: string;
  max_credits?: number;
}

interface CreateApplicationForm {
  affair_id: string;
  details: string;
  applied_credits: number;
}

interface ReviewApplicationForm {
  status: "approved" | "rejected";
  review_comment: string;
  approved_credits: number;
}

const createApplicationSchema = z.object({
  affair_id: z.string().min(1, "请选择事务类型"),
  details: z
    .string()
    .min(10, "详情至少10个字符")
    .max(1000, "详情不能超过1000个字符"),
  applied_credits: z
    .number()
    .min(0.5, "申请学分至少0.5")
    .max(10, "申请学分最多10"),
});

const reviewApplicationSchema = z.object({
  status: z.enum(["approved", "rejected"]),
  review_comment: z
    .string()
    .min(1, "请填写审核意见")
    .max(500, "审核意见不能超过500个字符"),
  approved_credits: z
    .number()
    .min(0, "批准学分不能为负数")
    .max(10, "批准学分最多10"),
});

const getFileIcon = (filename: string) => {
  const ext = filename.split(".").pop()?.toLowerCase();
  switch (ext) {
    case "pdf":
      return <FileText className="h-4 w-4 text-red-500" />;
    case "doc":
    case "docx":
      return <FileText className="h-4 w-4 text-blue-500" />;
    case "jpg":
    case "jpeg":
    case "png":
    case "gif":
      return <Image className="h-4 w-4 text-green-500" />;
    case "mp4":
    case "avi":
    case "mov":
      return <FileVideo className="h-4 w-4 text-purple-500" />;
    case "mp3":
    case "wav":
      return <FileAudio className="h-4 w-4 text-orange-500" />;
    default:
      return <File className="h-4 w-4 text-gray-500" />;
  }
};

const formatFileSize = (bytes: number) => {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
};

export default function ApplicationsPage() {
  const { user } = useAuth();
  const [applications, setApplications] = useState<Application[]>([]);
  const [affairs, setAffairs] = useState<Affair[]>([]);
  const [loading, setLoading] = useState(true);
  const [isCreateDialogOpen, setCreateDialogOpen] = useState(false);
  const [isDetailDialogOpen, setDetailDialogOpen] = useState(false);
  const [isReviewDialogOpen, setReviewDialogOpen] = useState(false);
  const [selectedApp, setSelectedApp] = useState<Application | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [uploadingFiles, setUploadingFiles] = useState<File[]>([]);
  const [uploadProgress, setUploadProgress] = useState<number>(0);
  const [isUploading, setIsUploading] = useState(false);
  const [selectedAppAttachments, setSelectedAppAttachments] = useState<
    ApplicationAttachment[]
  >([]);

  const createForm = useForm<CreateApplicationForm>({
    resolver: zodResolver(createApplicationSchema),
    defaultValues: { affair_id: "", details: "", applied_credits: 1 },
  });

  const reviewForm = useForm<ReviewApplicationForm>({
    resolver: zodResolver(reviewApplicationSchema),
    defaultValues: {
      status: "approved",
      review_comment: "",
      approved_credits: 0,
    },
  });

  // 解析附件JSON字符串
  const parseAttachments = (
    attachmentsJson: string
  ): ApplicationAttachment[] => {
    try {
      return attachmentsJson ? JSON.parse(attachmentsJson) : [];
    } catch {
      return [];
    }
  };

  const fetchApplications = async () => {
    try {
      setLoading(true);
      // 根据用户类型选择不同的API端点
      const endpoint =
        user?.userType === "student"
          ? "/applications" // 学生只能看到自己的申请
          : "/applications/all"; // 教师和管理员可以看到所有申请

      const response = await apiClient.get(endpoint);
      console.log("Applications response:", response.data); // 调试日志

      // 正确处理后端返回的数据结构
      let applicationsData = [];
      if (response.data.code === 0 && response.data.data) {
        if (response.data.data.data && Array.isArray(response.data.data.data)) {
          // 分页数据结构
          applicationsData = response.data.data.data;
        } else if (Array.isArray(response.data.data)) {
          // 直接数组结构
          applicationsData = response.data.data;
        }
      }

      setApplications(
        applicationsData.map((app: any) => ({
          id: app.id,
          affair_id: app.activity_id,
          affair_name: app.activity?.title || app.affair_name,
          student_number: app.user_info?.student_id || app.student_number,
          student_name: app.user_info?.name || app.student_name,
          submission_time: app.submitted_at || app.submission_time,
          status: app.status,
          reviewer_id: app.reviewer_id,
          reviewer_name: app.reviewer_name,
          review_comment: app.review_comment,
          review_time: app.reviewed_at || app.review_time,
          applied_credits: app.applied_credits,
          approved_credits: app.awarded_credits || app.approved_credits,
          details: app.details || "",
          attachments: app.attachments,
        }))
      );
    } catch (err) {
      console.error("Failed to fetch applications:", err);
      toast.error("获取申请列表失败");
    } finally {
      setLoading(false);
    }
  };

  const fetchAffairs = async () => {
    try {
      const response = await apiClient.get("/activities");
      console.log("Activities response:", response.data); // 调试日志

      // 正确处理后端返回的数据结构
      let activities = [];
      if (response.data.code === 0 && response.data.data) {
        if (response.data.data.data && Array.isArray(response.data.data.data)) {
          // 分页数据结构
          activities = response.data.data.data;
        } else if (Array.isArray(response.data.data)) {
          // 直接数组结构
          activities = response.data.data;
        }
      }

      setAffairs(
        activities.map((activity: any) => ({
          id: activity.id,
          name: activity.title,
          description: activity.description,
          max_credits: activity.max_credits || 1,
        }))
      );
    } catch (err) {
      console.error("Failed to fetch activities:", err);
      toast.error("获取活动列表失败");
    }
  };

  useEffect(() => {
    fetchApplications();
    fetchAffairs();
  }, [user]);

  const handleCreateApplication = async (values: CreateApplicationForm) => {
    if (uploadingFiles.length === 0) {
      toast.error("请至少上传一个文件");
      return;
    }

    try {
      setIsUploading(true);
      setUploadProgress(0);

      const formData = new FormData();
      formData.append("affair_id", values.affair_id);
      formData.append("student_id", user?.id || "");
      formData.append("user_id", user?.id || "");
      formData.append("details", values.details);
      formData.append("applied_credits", values.applied_credits.toString());

      // Add files
      uploadingFiles.forEach((file) => {
        formData.append(`files`, file);
      });

      await apiClient.post("/applications", formData, {
        headers: { "Content-Type": "multipart/form-data" },
        onUploadProgress: (progressEvent) => {
          if (progressEvent.total) {
            const progress = Math.round(
              (progressEvent.loaded * 100) / progressEvent.total
            );
            setUploadProgress(progress);
          }
        },
      });

      setCreateDialogOpen(false);
      createForm.reset();
      setUploadingFiles([]);
      setUploadProgress(0);
      fetchApplications();
      toast.success("申请提交成功！");
    } catch (err) {
      console.error("Failed to create application:", err);
      toast.error("申请提交失败");
    } finally {
      setIsUploading(false);
    }
  };

  const handleReviewApplication = async (values: ReviewApplicationForm) => {
    if (!selectedApp) return;

    try {
      await apiClient.put(`/applications/${selectedApp.id}/status`, {
        status: values.status,
        review_comment: values.review_comment,
        approved_credits: values.approved_credits,
        reviewer_id: user?.id,
      });

      setReviewDialogOpen(false);
      reviewForm.reset();
      setSelectedApp(null);
      fetchApplications();
      toast.success("审核完成！");
    } catch (err) {
      console.error("Failed to review application:", err);
      toast.error("审核失败");
    }
  };

  const handleFileUpload = (event: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files || []);
    const maxFileSize = 10 * 1024 * 1024; // 10MB
    const allowedTypes = ["pdf", "doc", "docx", "jpg", "jpeg", "png", "gif"];

    const validFiles = files.filter((file) => {
      const ext = file.name.split(".").pop()?.toLowerCase();
      if (!ext || !allowedTypes.includes(ext)) {
        toast.error(`不支持的文件类型: ${file.name}`);
        return false;
      }
      if (file.size > maxFileSize) {
        toast.error(`文件过大: ${file.name} (最大10MB)`);
        return false;
      }
      return true;
    });

    setUploadingFiles((prev) => [...prev, ...validFiles]);
  };

  const removeFile = (index: number) => {
    setUploadingFiles((prev) => prev.filter((_, i) => i !== index));
  };

  const handleFileDownload = async (
    _attachmentId: string,
    fileName: string
  ) => {
    try {
      const response = await apiClient.get(
        `/applications/${selectedApp?.id}/attachments/${fileName}/download`,
        {
          responseType: "blob",
        }
      );

      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement("a");
      link.href = url;
      link.setAttribute("download", fileName);
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);

      toast.success("文件下载成功");
    } catch (err) {
      console.error("Failed to download file:", err);
      toast.error("文件下载失败");
    }
  };

  const handleViewApplication = (application: Application) => {
    setSelectedApp(application);
    // 解析附件
    const attachments = parseAttachments(application.attachments || "[]");
    setSelectedAppAttachments(attachments);
    setDetailDialogOpen(true);
  };

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      unsubmitted: { label: "Unsubmitted", color: "bg-gray-100 text-gray-800" },
      pending: { label: "Pending", color: "bg-yellow-100 text-yellow-800" },
      approved: { label: "Approved", color: "bg-green-100 text-green-800" },
      rejected: { label: "Rejected", color: "bg-red-100 text-red-800" },
    };

    const config =
      statusConfig[status as keyof typeof statusConfig] || statusConfig.pending;
    return <Badge className={config.color}>{config.label}</Badge>;
  };

  const filteredApplications = applications.filter((app) => {
    const matchesSearch =
      app.student_name?.includes(searchTerm) ||
      app.student_number.includes(searchTerm) ||
      app.affair_name?.includes(searchTerm);
    const matchesStatus = statusFilter === "all" || app.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  return (
    <div className="space-y-8 p-4 md:p-8">
      <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">申请列表</h1>
          <p className="text-muted-foreground">管理学分申请和审核流程</p>
        </div>
      </div>

      {/* Filters */}
      <Card className="rounded-xl shadow-lg">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Filter className="h-5 w-5" />
            筛选和搜索
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col md:flex-row gap-4 items-center">
            <div className="relative w-full max-w-xs">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="搜索学生姓名、学号或事务名称..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10 rounded-lg shadow-sm"
              />
            </div>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-48 rounded-lg">
                <SelectValue placeholder="选择状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="unsubmitted">Unsubmitted</SelectItem>
                <SelectItem value="pending">Pending</SelectItem>
                <SelectItem value="approved">Approved</SelectItem>
                <SelectItem value="rejected">Rejected</SelectItem>
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              onClick={fetchApplications}
              disabled={loading}
              className="rounded-lg shadow"
            >
              <RefreshCw
                className={`h-4 w-4 ${loading ? "animate-spin" : ""}`}
              />
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Applications Table */}
      <Card className="rounded-xl shadow-lg">
        <CardHeader>
          <CardTitle>申请列表</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-xl overflow-x-auto">
            <Table>
              <TableHeader className="bg-muted/60">
                <TableRow>
                  <TableHead className="font-bold text-primary">
                    申请ID
                  </TableHead>
                  <TableHead>学生</TableHead>
                  <TableHead>事务类型</TableHead>
                  <TableHead>申请学分</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>提交时间</TableHead>
                  <TableHead>审核人</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <tr>
                    <td colSpan={8} className="py-8 text-center">
                      <div className="flex flex-col items-center gap-2">
                        <RefreshCw className="h-6 w-6 animate-spin text-muted-foreground" />
                        <span className="text-muted-foreground">加载中...</span>
                      </div>
                    </td>
                  </tr>
                ) : filteredApplications.length === 0 ? (
                  <tr>
                    <td colSpan={8} className="py-12">
                      <div className="flex flex-col items-center text-muted-foreground">
                        <AlertCircle className="w-12 h-12 mb-2" />
                        <p>暂无申请记录</p>
                      </div>
                    </td>
                  </tr>
                ) : (
                  filteredApplications.map((app) => (
                    <TableRow
                      key={app.id}
                      className="hover:bg-muted/40 transition-colors"
                    >
                      <TableCell className="font-semibold text-primary">
                        #{app.id}
                      </TableCell>
                      <TableCell>
                        <div className="font-medium">{app.student_number}</div>
                        <div className="text-xs text-muted-foreground">
                          {app.student_name}
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant="secondary"
                          className="rounded px-2 py-1"
                        >
                          {app.affair_name || `事务#${app.affair_id}`}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <span className="font-bold text-blue-600">
                          {app.applied_credits}
                        </span>
                      </TableCell>
                      <TableCell>
                        <Badge
                          className={
                            app.status === "unsubmitted"
                              ? "bg-gray-100 text-gray-800 rounded-lg px-2 py-1"
                              : app.status === "pending"
                              ? "bg-yellow-100 text-yellow-800 rounded-lg px-2 py-1"
                              : app.status === "approved"
                              ? "bg-green-100 text-green-800 rounded-lg px-2 py-1"
                              : "bg-red-100 text-red-800 rounded-lg px-2 py-1"
                          }
                        >
                          {app.status === "unsubmitted" ? (
                            <Clock className="w-3 h-3 mr-1 inline" />
                          ) : null}
                          {app.status === "pending" ? (
                            <Clock className="w-3 h-3 mr-1 inline" />
                          ) : null}
                          {app.status === "approved" ? (
                            <CheckCircle className="w-3 h-3 mr-1 inline" />
                          ) : null}
                          {app.status === "rejected" ? (
                            <XCircle className="w-3 h-3 mr-1 inline" />
                          ) : null}
                          {app.status === "unsubmitted"
                            ? "Unsubmitted"
                            : app.status === "pending"
                            ? "Pending"
                            : app.status === "approved"
                            ? "Approved"
                            : "Rejected"}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {app.submission_time?.split("T")[0] || "-"}
                      </TableCell>
                      <TableCell>{app.reviewer_name || "-"}</TableCell>
                      <TableCell>
                        <Button
                          size="icon"
                          variant="ghost"
                          className="rounded-full hover:bg-primary/10"
                          title="查看详情"
                          onClick={() => {
                            handleViewApplication(app);
                          }}
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* Create Application Dialog */}
      <Dialog open={isCreateDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>提交新申请</DialogTitle>
            <DialogDescription>选择事务类型并填写详细信息</DialogDescription>
          </DialogHeader>
          <Form {...createForm}>
            <form
              onSubmit={createForm.handleSubmit(handleCreateApplication)}
              className="space-y-6"
            >
              <FormField
                control={createForm.control}
                name="affair_id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>事务类型</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="选择事务类型" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {affairs.map((affair) => (
                          <SelectItem key={affair.id} value={String(affair.id)}>
                            {affair.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={createForm.control}
                name="applied_credits"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>申请学分</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        step="0.5"
                        min="0.5"
                        max="10"
                        {...field}
                        onChange={(e) =>
                          field.onChange(parseFloat(e.target.value) || 0)
                        }
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={createForm.control}
                name="details"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>申请详情</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="请详细描述您参与的事务内容、您的贡献和获得的成果..."
                        className="min-h-[120px]"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <div>
                <FormLabel>上传文件</FormLabel>
                <div className="mt-2">
                  <Input
                    type="file"
                    multiple
                    onChange={handleFileUpload}
                    accept=".pdf,.doc,.docx,.jpg,.jpeg,.png,.gif"
                    disabled={isUploading}
                  />
                  <p className="text-sm text-muted-foreground mt-1">
                    支持PDF、Word文档和图片格式，单个文件不超过10MB
                  </p>
                </div>
                {uploadingFiles.length > 0 && (
                  <div className="mt-2 space-y-2">
                    {uploadingFiles.map((file, index) => (
                      <div
                        key={index}
                        className="flex items-center justify-between p-2 border rounded"
                      >
                        <div className="flex items-center gap-2">
                          {getFileIcon(file.name)}
                          <span className="text-sm">{file.name}</span>
                          <span className="text-xs text-muted-foreground">
                            ({formatFileSize(file.size)})
                          </span>
                        </div>
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          onClick={() => removeFile(index)}
                          disabled={isUploading}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                )}
                {isUploading && (
                  <div className="mt-2">
                    <Progress value={uploadProgress} className="h-2" />
                    <p className="text-sm text-muted-foreground mt-1">
                      上传进度: {uploadProgress}%
                    </p>
                  </div>
                )}
              </div>
              <DialogFooter>
                <Button type="submit" className="w-full" disabled={isUploading}>
                  {isUploading ? (
                    <div className="flex items-center gap-2">
                      <RefreshCw className="h-4 w-4 animate-spin" />
                      上传中...
                    </div>
                  ) : (
                    <>
                      <Upload className="mr-2 h-4 w-4" />
                      提交申请
                    </>
                  )}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>

      {/* Application Detail Dialog */}
      <Dialog open={isDetailDialogOpen} onOpenChange={setDetailDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>申请详情</DialogTitle>
          </DialogHeader>
          {selectedApp && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium">申请ID</label>
                  <p className="text-sm text-muted-foreground">
                    #{selectedApp.id}
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium">状态</label>
                  <div className="mt-1">
                    {getStatusBadge(selectedApp.status)}
                  </div>
                </div>
                <div>
                  <label className="text-sm font-medium">学生</label>
                  <p className="text-sm text-muted-foreground">
                    {selectedApp.student_name || selectedApp.student_number}
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium">事务类型</label>
                  <p className="text-sm text-muted-foreground">
                    {selectedApp.affair_name || `事务#${selectedApp.affair_id}`}
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium">申请学分</label>
                  <p className="text-sm text-muted-foreground">
                    {selectedApp.applied_credits}
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium">批准学分</label>
                  <p className="text-sm text-muted-foreground">
                    {selectedApp.approved_credits}
                  </p>
                </div>
              </div>
              <div>
                <label className="text-sm font-medium">申请详情</label>
                <p className="text-sm text-muted-foreground mt-1 whitespace-pre-wrap">
                  {selectedApp.details}
                </p>
              </div>
              {selectedApp.review_comment && (
                <div>
                  <label className="text-sm font-medium">审核意见</label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {selectedApp.review_comment}
                  </p>
                </div>
              )}
              {selectedAppAttachments.length > 0 && (
                <div>
                  <label className="text-sm font-medium">附件</label>
                  <div className="mt-2 space-y-2">
                    {selectedAppAttachments.map(
                      (attachment: ApplicationAttachment) => (
                        <div
                          key={attachment.id}
                          className="flex items-center justify-between p-2 border rounded"
                        >
                          <div className="flex items-center gap-2">
                            {getFileIcon(attachment.original_name)}
                            <span className="text-sm">
                              {attachment.original_name}
                            </span>
                            <span className="text-xs text-muted-foreground">
                              ({formatFileSize(attachment.file_size)})
                            </span>
                          </div>
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() =>
                              handleFileDownload(
                                attachment.id,
                                attachment.file_name
                              )
                            }
                          >
                            <Download className="h-4 w-4" />
                          </Button>
                        </div>
                      )
                    )}
                  </div>
                </div>
              )}
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Review Application Dialog */}
      <Dialog open={isReviewDialogOpen} onOpenChange={setReviewDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>审核申请</DialogTitle>
            <DialogDescription>请填写审核意见和批准的学分</DialogDescription>
          </DialogHeader>
          <Form {...reviewForm}>
            <form
              onSubmit={reviewForm.handleSubmit(handleReviewApplication)}
              className="space-y-4"
            >
              <FormField
                control={reviewForm.control}
                name="status"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>审核结果</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="approved">通过</SelectItem>
                        <SelectItem value="rejected">拒绝</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={reviewForm.control}
                name="approved_credits"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>批准学分</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        step="0.5"
                        min="0"
                        max="10"
                        {...field}
                        onChange={(e) =>
                          field.onChange(parseFloat(e.target.value) || 0)
                        }
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={reviewForm.control}
                name="review_comment"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>审核意见</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="请填写详细的审核意见..."
                        className="min-h-[100px]"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button type="submit" className="w-full">
                  <CheckCircle className="mr-2 h-4 w-4" />
                  提交审核
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
