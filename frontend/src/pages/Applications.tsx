import { useState, useEffect } from "react";
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
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Pagination } from "@/components/ui/pagination";
import apiClient from "@/lib/api";
import {
  Search,
  Eye,
  Download,
  Filter,
  RefreshCw,
  AlertCircle,
  User,
} from "lucide-react";
import toast from "react-hot-toast";
import { getFileIcon, formatFileSize } from "@/lib/utils";
import React from "react";
import { getStatusBadge } from "@/lib/status-utils";
import { getActivityOptions } from "@/lib/options";

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
  attachments?: string; // JSON string of attachments
  user_id?: string; // 用户ID，用于查询用户信息
  user_info?: {
    username: string;
    name: string;
    student_id?: string;
    college?: string;
    major?: string;
    class?: string;
    grade?: string;
    user_type?: string;
  };
  activity?: {
    id: string;
    title: string;
    description: string;
    category: string;
    start_date: string;
    end_date: string;
    status: string;
    owner_id: string;
  };
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

export default function ApplicationsPage() {
  const { user } = useAuth();
  const [applications, setApplications] = useState<Application[]>([]);
  const [loading, setLoading] = useState(true);
  const [isDetailDialogOpen, setDetailDialogOpen] = useState(false);
  const [selectedApp, setSelectedApp] = useState<Application | null>(null);
  const [searchTerm, setSearchTerm] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [applicationStatuses, setApplicationStatuses] = useState<{ value: string; label: string }[]>([]);
  const [selectedAppAttachments, setSelectedAppAttachments] = useState<
    ApplicationAttachment[]
  >([]);

  // 分页状态
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [totalItems, setTotalItems] = useState(0);
  const [totalPages, setTotalPages] = useState(0);

  const parseAttachments = (
    attachmentsJson: string
  ): ApplicationAttachment[] => {
    try {
      return JSON.parse(attachmentsJson);
    } catch (error) {
      console.error("Failed to parse attachments:", error);
      return [];
    }
  };

  const fetchApplications = async (page = currentPage, size = pageSize) => {
    try {
      setLoading(true);

      // 构建查询参数
      const params: any = {
        page,
        page_size: size,
      };

      if (searchTerm) {
        params.query = searchTerm;
      }
      if (statusFilter !== "all") {
        params.status = statusFilter;
      }

      // 根据用户类型选择不同的API端点
      const endpoint =
        user?.userType === "student"
          ? "/applications" // 学生只能看到自己的申请
          : "/applications/all"; // 教师和管理员可以看到所有申请

      const response = await apiClient.get(endpoint, { params });

      // 调试：打印响应数据
      console.log("Applications API Response:", response.data);

      // 处理响应数据
      let applicationsData = [];
      let paginationData: any = {};

      if (response.data.code === 0 && response.data.data) {
        if (response.data.data.data && Array.isArray(response.data.data.data)) {
          // 分页数据结构
          applicationsData = response.data.data.data;
          paginationData = {
            total: response.data.data.total || 0,
            page: response.data.data.page || 1,
            page_size: response.data.data.page_size || 10,
            total_pages: response.data.data.total_pages || 0,
          };
        } else {
          // 非分页数据结构
          applicationsData = response.data.data.applications || response.data.data || [];
          paginationData = {
            total: applicationsData.length,
            page: 1,
            page_size: applicationsData.length,
            total_pages: 1,
          };
        }
      } else {
        applicationsData = [];
        paginationData = {
          total: 0,
          page: 1,
          page_size: 10,
          total_pages: 0,
        };
      }

      // 处理申请数据（确保为数组）
      const applicationsArray = Array.isArray(applicationsData)
        ? applicationsData
        : applicationsData && typeof applicationsData === "object"
        ? Object.values(applicationsData)
        : [];

      const processedApplications = applicationsArray
        .filter((app: any) => app && app.id) // 过滤掉 null/undefined 或没有 id 的项目
        .map((app: any) => ({
          id: app.id,
          affair_id: app.activity_id,
          affair_name: app.activity?.title || app.affair_name,
          student_number: app.user_info?.student_id || app.student_number || "",
          student_name: app.user_info?.name || app.student_name || "",
          submission_time: app.submitted_at || app.submission_time,
          status: app.status,
          reviewer_id: app.reviewer_id,
          reviewer_name: app.reviewer_name,
          review_comment: app.review_comment,
          review_time: app.reviewed_at || app.review_time,
          applied_credits: app.applied_credits,
          approved_credits: app.awarded_credits || app.approved_credits,
          attachments: app.attachments,
          user_id: app.user_info?.id,
          user_info: app.user_info,
          activity: app.activity,
        }));

      // 调试：打印处理后的数据
      console.log("Processed Applications:", processedApplications);
      console.log("Is Array:", Array.isArray(processedApplications));
      
      setApplications(processedApplications);

      // 更新分页信息
      const safeTotal = Array.isArray(applicationsArray)
        ? applicationsArray.length
        : Number(paginationData.total) || 0;
      setTotalItems(safeTotal);
      setTotalPages(Number(paginationData.total_pages) || 1);
      setCurrentPage(Number(paginationData.page) || 1);
    } catch (err) {
      console.error("Failed to fetch applications:", err);
      toast.error("获取申请列表失败");
      setApplications([]);
      setTotalItems(0);
      setTotalPages(0);
    } finally {
      setLoading(false);
    }
  };

  // 分页处理函数
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    fetchApplications(page, pageSize);
  };

  const handlePageSizeChange = (size: number) => {
    setPageSize(size);
    setCurrentPage(1);
    fetchApplications(1, size);
  };

  // 搜索和筛选处理
  const handleSearchAndFilter = () => {
    setCurrentPage(1);
    fetchApplications(1, pageSize);
  };

  useEffect(() => {
    fetchApplications();
  }, [user]);

  useEffect(() => {
    (async () => {
      try {
        const opts = await getActivityOptions();
        // 申请状态通常与活动审核状态一致，若不同可在后端单独提供 application_statuses
        setApplicationStatuses(opts.statuses || []);
      } catch (e) {
        console.error("Failed to load application statuses", e);
      }
    })();
  }, []);

  // 本地搜索过滤功能
  const safeApplications = Array.isArray(applications) ? applications : [];
  const filteredApplications = safeApplications.filter((app) => {
    if (!searchTerm) return true;

    const searchLower = searchTerm.toLowerCase();

    // 搜索申请人姓名
    if (app.user_info?.name?.toLowerCase().includes(searchLower)) return true;
    if (app.student_name?.toLowerCase().includes(searchLower)) return true;

    // 搜索学号
    if (app.user_info?.student_id?.toLowerCase().includes(searchLower))
      return true;
    if (app.student_number?.toLowerCase().includes(searchLower)) return true;

    // 搜索用户名
    if (app.user_info?.username?.toLowerCase().includes(searchLower))
      return true;

    // 搜索学院
    if (app.user_info?.college?.toLowerCase().includes(searchLower))
      return true;

    // 搜索专业
    if (app.user_info?.major?.toLowerCase().includes(searchLower)) return true;

    // 搜索班级
    if (app.user_info?.class?.toLowerCase().includes(searchLower)) return true;

    // 搜索活动名称
    if (app.affair_name?.toLowerCase().includes(searchLower)) return true;
    if (app.activity?.title?.toLowerCase().includes(searchLower)) return true;

    // 搜索活动类别
    if (app.activity?.category?.toLowerCase().includes(searchLower))
      return true;

    // 搜索活动描述
    if (app.activity?.description?.toLowerCase().includes(searchLower))
      return true;

    return false;
  });

  // 状态过滤
  const statusFilteredApplications = filteredApplications.filter((app) => {
    if (statusFilter === "all") return true;
    return app.status === statusFilter;
  });

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

  return (
    <div className="space-y-8 p-4 md:p-8">
      <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">申请管理</h1>
          <p className="text-muted-foreground">查看和管理所有学分申请记录</p>
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
                placeholder="搜索申请人姓名、学号、学院、专业、班级或事务名称..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") {
                    handleSearchAndFilter();
                  }
                }}
                className="pl-10 rounded-lg shadow-sm"
              />
            </div>
            <Select
              value={statusFilter}
              onValueChange={(value) => {
                setStatusFilter(value);
                handleSearchAndFilter();
              }}
            >
              <SelectTrigger className="w-48 rounded-lg">
                <SelectValue placeholder="选择状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                {applicationStatuses.map((s) => (
                  <SelectItem key={s.value} value={s.value}>{s.label}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              onClick={handleSearchAndFilter}
              disabled={loading}
              className="rounded-lg shadow"
            >
              <Search className="h-4 w-4" />
            </Button>
            <Button
              variant="outline"
              onClick={() => {
                setSearchTerm("");
                setStatusFilter("all");
                setCurrentPage(1);
                fetchApplications(1, pageSize);
              }}
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
          <CardTitle>申请记录</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-xl overflow-x-auto">
            <Table>
              <TableHeader className="bg-muted/60">
                <TableRow>
                  <TableHead>申请人信息</TableHead>
                  <TableHead>活动信息</TableHead>
                  <TableHead>申请学分</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>提交时间</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <tr>
                    <td colSpan={6} className="py-8 text-center">
                      <div className="flex flex-col items-center gap-2">
                        <RefreshCw className="h-6 w-6 animate-spin text-muted-foreground" />
                        <span className="text-muted-foreground">加载中...</span>
                      </div>
                    </td>
                  </tr>
                ) : !statusFilteredApplications || statusFilteredApplications.length === 0 ? (
                  <tr>
                    <td colSpan={6} className="py-12">
                      <div className="flex flex-col items-center text-muted-foreground">
                        <AlertCircle className="w-12 h-12 mb-2" />
                        <p>暂无申请记录</p>
                      </div>
                    </td>
                  </tr>
                ) : (
                  (statusFilteredApplications || []).map((app) => (
                    <TableRow
                      key={app.id}
                      className="hover:bg-muted/40 transition-colors"
                    >
                      <TableCell>
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
                            <User className="h-4 w-4 text-primary" />
                          </div>
                          <div>
                            <div className="font-medium">
                              {app.user_info?.name ||
                                app.student_name ||
                                (app.user_id ? `用户 ${app.user_id}` : `申请 ${app.id}`)}
                            </div>
                            <div className="text-sm text-muted-foreground">
                              {app.user_info?.student_id ||
                                app.student_number ||
                                app.user_info?.username ||
                                app.id}
                            </div>
                            {app.user_info?.college && (
                              <div className="text-xs text-muted-foreground">
                                {app.user_info.college} - {app.user_info.major}
                                {app.user_info.class &&
                                  ` - ${app.user_info.class}`}
                                {app.user_info.grade &&
                                  ` (${app.user_info.grade}级)`}
                              </div>
                            )}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="space-y-1">
                          <Badge
                            variant="secondary"
                            className="rounded px-2 py-1"
                          >
                            {app.activity?.category || "未知类别"}
                          </Badge>
                          <div className="text-sm font-medium">
                            {app.affair_name || `活动#${app.affair_id}`}
                          </div>
                          {app.activity?.description && (
                            <div className="text-xs text-muted-foreground line-clamp-2">
                              {app.activity.description}
                            </div>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <span className="font-bold text-blue-600 dark:text-blue-400">
                          {app.applied_credits}
                        </span>
                      </TableCell>
                      <TableCell>
                        {getStatusBadge(app.status)}
                      </TableCell>
                      <TableCell>
                        {app.submission_time?.split("T")[0] || "-"}
                      </TableCell>
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

      {/* Pagination */}
      {totalItems > 0 && (
        <Card className="rounded-xl shadow-lg">
          <CardContent className="pt-6">
            <Pagination
              currentPage={currentPage}
              totalPages={totalPages}
              totalItems={totalItems}
              pageSize={pageSize}
              onPageChange={handlePageChange}
              onPageSizeChange={handlePageSizeChange}
            />
          </CardContent>
        </Card>
      )}

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
                  <label className="text-sm font-medium">申请人</label>
                  <p className="text-sm text-muted-foreground">
                    {selectedApp.user_info?.name ||
                      selectedApp.student_name ||
                      selectedApp.student_number}
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium">学号</label>
                  <p className="text-sm text-muted-foreground">
                    {selectedApp.user_info?.student_id ||
                      selectedApp.student_number}
                  </p>
                </div>
                {selectedApp.user_info?.college && (
                  <div>
                    <label className="text-sm font-medium">学院</label>
                    <p className="text-sm text-muted-foreground">
                      {selectedApp.user_info.college}
                    </p>
                  </div>
                )}
                {selectedApp.user_info?.major && (
                  <div>
                    <label className="text-sm font-medium">专业</label>
                    <p className="text-sm text-muted-foreground">
                      {selectedApp.user_info.major}
                    </p>
                  </div>
                )}
                {selectedApp.user_info?.class && (
                  <div>
                    <label className="text-sm font-medium">班级</label>
                    <p className="text-sm text-muted-foreground">
                      {selectedApp.user_info.class}
                    </p>
                  </div>
                )}
                <div>
                  <label className="text-sm font-medium">活动信息</label>
                  <div className="mt-1 space-y-1">
                    <p className="text-sm text-muted-foreground">
                      {selectedApp.affair_name ||
                        `活动#${selectedApp.affair_id}`}
                    </p>
                    {selectedApp.activity?.category && (
                      <Badge variant="outline" className="text-xs">
                        {selectedApp.activity.category}
                      </Badge>
                    )}
                    {selectedApp.activity?.description && (
                      <p className="text-xs text-muted-foreground mt-1">
                        {selectedApp.activity.description}
                      </p>
                    )}
                    {selectedApp.activity?.start_date &&
                      selectedApp.activity?.end_date && (
                        <p className="text-xs text-muted-foreground">
                          {selectedApp.activity.start_date.split("T")[0]} -{" "}
                          {selectedApp.activity.end_date.split("T")[0]}
                        </p>
                      )}
                  </div>
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
              {selectedApp.review_comment && (
                <div>
                  <label className="text-sm font-medium">审核意见</label>
                  <p className="text-sm text-muted-foreground mt-1">
                    {selectedApp.review_comment}
                  </p>
                </div>
              )}
              {selectedAppAttachments && selectedAppAttachments.length > 0 && (
                <div>
                  <label className="text-sm font-medium">附件</label>
                  <div className="mt-2 space-y-2">
                    {(selectedAppAttachments || []).map(
                      (attachment: ApplicationAttachment) => (
                        <div
                          key={attachment.id}
                          className="flex items-center justify-between p-2 border rounded"
                        >
                          <div className="flex items-center gap-2">
                            {React.createElement(getFileIcon(attachment.original_name))}
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
    </div>
  );
}
