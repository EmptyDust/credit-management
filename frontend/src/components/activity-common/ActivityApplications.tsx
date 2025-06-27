import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  FileText,
  CheckCircle,
  XCircle,
  Clock,
  AlertCircle,
  User,
  Eye,
  Edit,
  Trash2,
  Download,
  Filter,
  Search,
  ExternalLink,
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useAuth } from "@/contexts/AuthContext";
import apiClient from "@/lib/api";
import toast from "react-hot-toast";
import type { Activity, Application } from "@/types/activity";

interface ActivityApplicationsProps {
  activity: Activity;
  onRefresh?: () => void;
}

// 获取状态显示文本
const getStatusText = (status: string) => {
  switch (status) {
    case "approved":
      return "已通过";
    case "pending":
      return "待审核";
    case "rejected":
      return "已拒绝";
    case "draft":
      return "草稿";
    default:
      return status;
  }
};

// 获取状态样式
const getStatusStyle = (status: string) => {
  switch (status) {
    case "approved":
      return "bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-400";
    case "pending":
      return "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20 dark:text-yellow-400";
    case "rejected":
      return "bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-400";
    case "draft":
      return "bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-400";
    default:
      return "bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-400";
  }
};

// 获取状态图标
const getStatusIcon = (status: string) => {
  switch (status) {
    case "approved":
      return <CheckCircle className="w-3 h-3 mr-1 inline" />;
    case "pending":
      return <Clock className="w-3 h-3 mr-1 inline" />;
    case "rejected":
      return <XCircle className="w-3 h-3 mr-1 inline" />;
    case "draft":
      return <FileText className="w-3 h-3 mr-1 inline" />;
    default:
      return <AlertCircle className="w-3 h-3 mr-1 inline" />;
  }
};

export default function ActivityApplications({
  activity,
  onRefresh,
}: ActivityApplicationsProps) {
  const { user } = useAuth();
  const [applications, setApplications] = useState<Application[]>([]);
  const [, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [showDetailDialog, setShowDetailDialog] = useState(false);
  const [selectedApplication, setSelectedApplication] =
    useState<Application | null>(null);
  const [showReviewDialog, setShowReviewDialog] = useState(false);
  const [reviewCredits, setReviewCredits] = useState(0);
  const [reviewComments, setReviewComments] = useState("");
  const navigate = useNavigate();

  const isOwner =
    user && (user.user_id === activity.owner_id || user.userType === "admin");
  const isReviewer =
    user && (user.userType === "teacher" || user.userType === "admin");

  // 获取申请列表
  const fetchApplications = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get(
        `/applications?activity_id=${activity.id}`
      );
      setApplications(response.data.data?.applications || []);
    } catch (error) {
      console.error("Failed to fetch applications:", error);
      toast.error("获取申请列表失败");
    } finally {
      setLoading(false);
    }
  };

  // 审核申请
  const reviewApplication = async (
    applicationId: string,
    status: string,
    credits?: number,
    comments?: string
  ) => {
    try {
      await apiClient.put(`/applications/${applicationId}/review`, {
        status,
        awarded_credits: credits,
        review_comments: comments,
      });
      toast.success("审核完成");
      setShowReviewDialog(false);
      setSelectedApplication(null);
      fetchApplications();
      onRefresh?.();
    } catch (error) {
      console.error("Failed to review application:", error);
      toast.error("审核失败");
    }
  };

  // 删除申请
  const deleteApplication = async (applicationId: string) => {
    try {
      await apiClient.delete(`/applications/${applicationId}`);
      toast.success("申请删除成功");
      fetchApplications();
      onRefresh?.();
    } catch (error) {
      console.error("Failed to delete application:", error);
      toast.error("删除失败");
    }
  };

  // 导出申请列表
  const exportApplications = async () => {
    try {
      const response = await apiClient.get(
        `/applications/export?activity_id=${activity.id}`,
        {
          params: { format: "json" },
        }
      );

      const data = response.data.data;
      const blob = new Blob([JSON.stringify(data, null, 2)], {
        type: "application/json",
      });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `申请列表_${activity.title}_${
        new Date().toISOString().split("T")[0]
      }.json`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);

      toast.success("导出成功");
    } catch (error) {
      console.error("Failed to export applications:", error);
      toast.error("导出失败");
    }
  };

  // 过滤申请
  const filteredApplications = applications.filter((application) => {
    const matchesSearch =
      !searchQuery ||
      application.user_info?.name
        ?.toLowerCase()
        .includes(searchQuery.toLowerCase()) ||
      application.user_info?.student_id
        ?.toLowerCase()
        .includes(searchQuery.toLowerCase()) ||
      application.user_info?.username
        ?.toLowerCase()
        .includes(searchQuery.toLowerCase()) ||
      application.user_info?.college
        ?.toLowerCase()
        .includes(searchQuery.toLowerCase()) ||
      application.user_info?.major
        ?.toLowerCase()
        .includes(searchQuery.toLowerCase()) ||
      application.user_info?.class
        ?.toLowerCase()
        .includes(searchQuery.toLowerCase());

    const matchesStatus =
      statusFilter === "all" || application.status === statusFilter;

    return matchesSearch && matchesStatus;
  });

  // 统计信息
  const stats = {
    total: applications.length,
    pending: applications.filter((a) => a.status === "pending").length,
    approved: applications.filter((a) => a.status === "approved").length,
    rejected: applications.filter((a) => a.status === "rejected").length,
    totalCredits: applications.reduce((sum, a) => sum + a.awarded_credits, 0),
  };

  useEffect(() => {
    fetchApplications();
  }, [activity.id]);

  if (applications.length === 0 && !isOwner) {
    return (
      <Card className="rounded-xl shadow-lg">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            申请列表
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            暂无申请记录
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="rounded-xl shadow-lg">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            申请列表 ({applications.length})
          </CardTitle>
          <div className="flex items-center gap-2">
            {isOwner && (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={exportApplications}
                >
                  <Download className="h-4 w-4 mr-1" />
                  导出
                </Button>
              </>
            )}
          </div>
        </div>

        {/* 统计信息 */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mt-4">
          <div className="text-center p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
            <div className="text-2xl font-bold text-blue-600">
              {stats.total}
            </div>
            <div className="text-sm text-muted-foreground">总申请</div>
          </div>
          <div className="text-center p-3 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
            <div className="text-2xl font-bold text-yellow-600">
              {stats.pending}
            </div>
            <div className="text-sm text-muted-foreground">待审核</div>
          </div>
          <div className="text-center p-3 bg-green-50 dark:bg-green-900/20 rounded-lg">
            <div className="text-2xl font-bold text-green-600">
              {stats.approved}
            </div>
            <div className="text-sm text-muted-foreground">已通过</div>
          </div>
          <div className="text-center p-3 bg-purple-50 dark:bg-purple-900/20 rounded-lg">
            <div className="text-2xl font-bold text-purple-600">
              {stats.totalCredits}
            </div>
            <div className="text-sm text-muted-foreground">总学分</div>
          </div>
        </div>
      </CardHeader>

      <CardContent>
        {/* 搜索和筛选 */}
        <div className="flex items-center gap-4 mb-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="搜索申请人姓名、学号、学院、专业、班级..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="w-40">
              <SelectValue placeholder="状态筛选" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="pending">待审核</SelectItem>
              <SelectItem value="approved">已通过</SelectItem>
              <SelectItem value="rejected">已拒绝</SelectItem>
              <SelectItem value="draft">草稿</SelectItem>
            </SelectContent>
          </Select>
        </div>

        {/* 申请表格 */}
        <div className="border rounded-lg">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>申请人信息</TableHead>
                <TableHead>活动信息</TableHead>
                <TableHead>申请学分</TableHead>
                <TableHead>授予学分</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>申请时间</TableHead>
                <TableHead className="w-20">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredApplications.map((application) => (
                <TableRow key={application.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
                        <User className="h-4 w-4 text-primary" />
                      </div>
                      <div>
                        <div className="font-medium">
                          {application.user_info?.name ||
                            `用户 ${application.user_id}`}
                        </div>
                        <div className="text-sm text-muted-foreground">
                          {application.user_info?.student_id ||
                            application.user_info?.username ||
                            application.user_id}
                        </div>
                        {application.user_info?.college && (
                          <div className="text-xs text-muted-foreground">
                            {application.user_info.college} -{" "}
                            {application.user_info.major}
                          </div>
                        )}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="space-y-1">
                      <Badge variant="outline" className="text-xs">
                        {activity.category || "未知类别"}
                      </Badge>
                      <div className="text-sm font-medium">
                        {activity.title}
                      </div>
                      {activity.description && (
                        <div className="text-xs text-muted-foreground line-clamp-2">
                          {activity.description}
                        </div>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="font-medium text-blue-600">
                      {application.applied_credits} 学分
                    </div>
                  </TableCell>
                  <TableCell>
                    <div className="font-bold text-primary">
                      {application.awarded_credits} 学分
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge
                      className={`${getStatusStyle(
                        application.status
                      )} rounded-lg px-2 py-1`}
                    >
                      {getStatusIcon(application.status)}
                      {getStatusText(application.status)}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="text-sm text-muted-foreground">
                      {new Date(application.submitted_at).toLocaleDateString()}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {new Date(application.submitted_at).toLocaleTimeString()}
                    </div>
                  </TableCell>
                  <TableCell>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="sm">
                          <Filter className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuLabel>操作</DropdownMenuLabel>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          onClick={() => {
                            setSelectedApplication(application);
                            setShowDetailDialog(true);
                          }}
                        >
                          <Eye className="h-4 w-4 mr-2" />
                          查看详情
                        </DropdownMenuItem>
                        {isReviewer && application.status === "pending" && (
                          <>
                            <DropdownMenuItem
                              onClick={() => {
                                setSelectedApplication(application);
                                setReviewCredits(application.applied_credits);
                                setReviewComments("");
                                setShowReviewDialog(true);
                              }}
                            >
                              <Edit className="h-4 w-4 mr-2" />
                              审核
                            </DropdownMenuItem>
                          </>
                        )}
                        {isOwner && (
                          <DropdownMenuItem
                            onClick={() => deleteApplication(application.id)}
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

        {filteredApplications.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            {searchQuery || statusFilter !== "all"
              ? "没有找到匹配的申请"
              : "暂无申请记录"}
          </div>
        )}
      </CardContent>

      {/* 申请详情对话框 */}
      <Dialog open={showDetailDialog} onOpenChange={setShowDetailDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>申请详情</DialogTitle>
            <DialogDescription>查看申请的详细信息</DialogDescription>
          </DialogHeader>

          {selectedApplication && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-sm font-medium text-gray-500">
                    申请人
                  </label>
                  <p className="text-sm">
                    {selectedApplication.user_info?.name ||
                      `用户 ${selectedApplication.user_id}`}
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium text-gray-500">
                    学号
                  </label>
                  <p className="text-sm">
                    {selectedApplication.user_info?.student_id || "-"}
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium text-gray-500">
                    申请学分
                  </label>
                  <p className="text-sm font-bold text-blue-600">
                    {selectedApplication.applied_credits} 学分
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium text-gray-500">
                    授予学分
                  </label>
                  <p className="text-sm font-bold text-primary">
                    {selectedApplication.awarded_credits} 学分
                  </p>
                </div>
                <div>
                  <label className="text-sm font-medium text-gray-500">
                    状态
                  </label>
                  <Badge
                    className={`${getStatusStyle(
                      selectedApplication.status
                    )} rounded-lg px-2 py-1`}
                  >
                    {getStatusIcon(selectedApplication.status)}
                    {getStatusText(selectedApplication.status)}
                  </Badge>
                </div>
                <div>
                  <label className="text-sm font-medium text-gray-500">
                    申请时间
                  </label>
                  <p className="text-sm">
                    {new Date(
                      selectedApplication.submitted_at
                    ).toLocaleString()}
                  </p>
                </div>
              </div>

              {selectedApplication.user_info?.college && (
                <div>
                  <label className="text-sm font-medium text-gray-500">
                    学院专业
                  </label>
                  <p className="text-sm">
                    {selectedApplication.user_info.college} -{" "}
                    {selectedApplication.user_info.major}
                  </p>
                </div>
              )}

              {/* 活动信息 */}
              <div className="border-t pt-4">
                <h4 className="text-sm font-medium text-gray-500 mb-2">
                  关联活动
                </h4>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-sm font-medium text-gray-500">
                      活动名称
                    </label>
                    <p className="text-sm font-medium">{activity.title}</p>
                  </div>
                  <div>
                    <label className="text-sm font-medium text-gray-500">
                      活动类别
                    </label>
                    <Badge variant="outline" className="text-xs">
                      {activity.category}
                    </Badge>
                  </div>
                  <div className="col-span-2">
                    <label className="text-sm font-medium text-gray-500">
                      活动描述
                    </label>
                    <p className="text-sm text-muted-foreground">
                      {activity.description || "暂无描述"}
                    </p>
                  </div>
                  {activity.start_date && activity.end_date && (
                    <div className="col-span-2">
                      <label className="text-sm font-medium text-gray-500">
                        活动时间
                      </label>
                      <p className="text-sm text-muted-foreground">
                        {new Date(activity.start_date).toLocaleDateString()} -{" "}
                        {new Date(activity.end_date).toLocaleDateString()}
                      </p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => navigate(`/activities/${activity.id}`)}
              className="flex items-center gap-2"
            >
              <ExternalLink className="h-4 w-4" />
              查看关联活动
            </Button>
            <Button onClick={() => setShowDetailDialog(false)}>关闭</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 审核对话框 */}
      <Dialog open={showReviewDialog} onOpenChange={setShowReviewDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>审核申请</DialogTitle>
            <DialogDescription>审核申请并设置授予学分</DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">申请人</label>
              <p className="text-sm">
                {selectedApplication?.user_info?.name ||
                  `用户 ${selectedApplication?.user_id}`}
              </p>
            </div>
            <div>
              <label className="text-sm font-medium">申请学分</label>
              <p className="text-sm">
                {selectedApplication?.applied_credits} 学分
              </p>
            </div>
            <div>
              <label className="text-sm font-medium">授予学分</label>
              <Input
                type="number"
                step="0.1"
                min="0"
                max="10"
                value={reviewCredits}
                onChange={(e) =>
                  setReviewCredits(parseFloat(e.target.value) || 0)
                }
                placeholder="请输入授予学分"
              />
            </div>
            <div>
              <label className="text-sm font-medium">审核意见</label>
              <Input
                value={reviewComments}
                onChange={(e) => setReviewComments(e.target.value)}
                placeholder="请输入审核意见（可选）"
              />
            </div>
          </div>

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowReviewDialog(false)}
            >
              取消
            </Button>
            <Button
              variant="destructive"
              onClick={() =>
                reviewApplication(
                  selectedApplication!.id,
                  "rejected",
                  reviewCredits,
                  reviewComments
                )
              }
            >
              拒绝
            </Button>
            <Button
              onClick={() =>
                reviewApplication(
                  selectedApplication!.id,
                  "approved",
                  reviewCredits,
                  reviewComments
                )
              }
            >
              通过
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
