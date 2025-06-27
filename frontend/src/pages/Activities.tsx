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
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Pagination } from "@/components/ui/pagination";
import apiClient, { apiHelpers } from "@/lib/api";
import {
  PlusCircle,
  Edit,
  Trash,
  Search,
  Filter,
  RefreshCw,
  Award,
  Users,
  AlertCircle,
  CheckCircle,
  XCircle,
  Eye,
  Upload,
  Download,
  FileText,
  Clock,
} from "lucide-react";
import toast from "react-hot-toast";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import { useNavigate, useSearchParams } from "react-router-dom";

// Types
interface Activity {
  id: string;
  title: string;
  description?: string;
  status: string;
  category?: string;
  created_at: string;
  updated_at: string;
  participants?: any[];
  applications?: any[];
  owner_id: string;
  owner_info?: {
    name: string;
    role: string;
  };
}

interface PaginatedResponse {
  data: Activity[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

type CreateActivityForm = z.infer<typeof activitySchema>;

const activitySchema = z.object({
  title: z
    .string()
    .min(1, "活动名称不能为空")
    .max(200, "活动名称不能超过200个字符"),
  category: z.string().min(1, "请选择活动类型"),
});

// 统计卡片样式与仪表盘一致
const StatCard = ({
  title,
  value,
  icon: Icon,
  color = "default",
  subtitle,
}: {
  title: string;
  value: string | number;
  icon: React.ElementType;
  color?: "default" | "success" | "warning" | "danger" | "info" | "purple";
  subtitle?: string;
}) => {
  const colorClasses = {
    default: "text-muted-foreground",
    success: "text-green-600",
    warning: "text-yellow-600",
    danger: "text-red-600",
    info: "text-blue-600",
    purple: "text-purple-600",
  };
  const bgClasses = {
    default: "bg-muted/20",
    success: "bg-green-100 dark:bg-green-900/20",
    warning: "bg-yellow-100 dark:bg-yellow-900/20",
    danger: "bg-red-100 dark:bg-red-900/20",
    info: "bg-blue-100 dark:bg-blue-900/20",
    purple: "bg-purple-100 dark:bg-purple-900/20",
  };
  return (
    <Card className="rounded-xl shadow-lg hover:shadow-2xl transition-all duration-300 bg-gradient-to-br from-white to-gray-50 dark:from-gray-900 dark:to-gray-800 border-0">
      <CardHeader className="flex flex-row items-center justify-between pb-3">
        <div className={`p-3 rounded-xl ${bgClasses[color]}`}>
          <Icon className={`h-6 w-6 ${colorClasses[color]}`} />
        </div>
        <CardTitle className="text-lg font-semibold text-gray-900 dark:text-gray-100">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-bold mb-1 text-gray-900 dark:text-gray-100">
          {value}
        </div>
        {subtitle && (
          <div className="text-sm text-muted-foreground mb-2">{subtitle}</div>
        )}
      </CardContent>
    </Card>
  );
};

export default function ActivitiesPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { user, hasPermission } = useAuth();
  const [activities, setActivities] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState("");
  const [categoryFilter, setCategoryFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingActivity, setEditingActivity] = useState<Activity | null>(null);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [deletingActivity, setDeletingActivity] = useState<Activity | null>(
    null
  );
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const [importFile, setImportFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);

  // 分页状态
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [totalItems, setTotalItems] = useState(0);
  const [totalPages, setTotalPages] = useState(0);

  const form = useForm<CreateActivityForm>({
    resolver: zodResolver(activitySchema),
    defaultValues: {
      title: "",
      category: "",
    },
  });

  // 根据URL参数设置初始过滤器
  useEffect(() => {
    const statusFromUrl = searchParams.get("status");
    if (statusFromUrl) {
      setStatusFilter(statusFromUrl);
    }
  }, [searchParams]);

  const fetchActivities = async (page = currentPage, size = pageSize) => {
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
      if (categoryFilter !== "all") {
        params.category = categoryFilter;
      }
      if (statusFilter !== "all") {
        params.status = statusFilter;
      }

      const response = await apiClient.get("/activities", { params });
      console.log("API Response:", response.data); // 调试日志

      // 处理不同的响应数据结构
      let activitiesData: Activity[] = [];
      let paginationData: any = {};

      if (response.data.data && response.data.data.data) {
        // 嵌套数据结构
        activitiesData = response.data.data.data;
        paginationData = {
          total: response.data.data.total || 0,
          page: response.data.data.page || 1,
          page_size: response.data.data.page_size || 10,
          total_pages: response.data.data.total_pages || 0,
        };
      } else if (response.data.data && Array.isArray(response.data.data)) {
        // 直接数组结构
        activitiesData = response.data.data;
        paginationData = {
          total: activitiesData.length,
          page: 1,
          page_size: activitiesData.length,
          total_pages: 1,
        };
      } else if (Array.isArray(response.data)) {
        // 直接数组
        activitiesData = response.data;
        paginationData = {
          total: activitiesData.length,
          page: 1,
          page_size: activitiesData.length,
          total_pages: 1,
        };
      } else if (
        response.data.activities &&
        Array.isArray(response.data.activities)
      ) {
        // 旧格式
        activitiesData = response.data.activities;
        paginationData = {
          total: response.data.total || activitiesData.length,
          page: response.data.page || 1,
          page_size: response.data.page_size || activitiesData.length,
          total_pages: response.data.total_pages || 1,
        };
      } else {
        // 默认处理
        activitiesData = [];
        paginationData = {
          total: 0,
          page: 1,
          page_size: 10,
          total_pages: 0,
        };
      }

      setActivities(activitiesData);
      setTotalItems(paginationData.total);
      setTotalPages(paginationData.total_pages);
    } catch (error) {
      console.error("Failed to fetch activities:", error);
      toast.error("获取活动列表失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchActivities();
  }, []);

  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    fetchActivities(page, pageSize);
  };

  const handlePageSizeChange = (size: number) => {
    setPageSize(size);
    setCurrentPage(1);
    fetchActivities(1, size);
  };

  const handleSearchAndFilter = () => {
    setCurrentPage(1);
    fetchActivities(1, pageSize);
  };

  const handleDialogOpen = (activity: Activity | null) => {
    setEditingActivity(activity);
    if (activity) {
      form.reset({
        title: activity.title,
        category: activity.category || "",
      });
    } else {
      form.reset({
        title: "",
        category: "",
      });
    }
    setIsDialogOpen(true);
  };

  const handleDeleteClick = (activity: Activity) => {
    setDeletingActivity(activity);
    setIsDeleteDialogOpen(true);
  };

  const handleDeleteConfirm = async () => {
    if (!deletingActivity) return;

    try {
      await apiClient.delete(`/activities/${deletingActivity.id}`);
      toast.success("活动删除成功");
      fetchActivities();
    } catch (error) {
      console.error("Failed to delete activity:", error);
      toast.error("删除活动失败");
    } finally {
      setIsDeleteDialogOpen(false);
      setDeletingActivity(null);
    }
  };

  const handleImportFile = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      // 验证文件类型
      const allowedTypes = [
        "application/vnd.ms-excel",
        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "text/csv",
      ];
      if (
        !allowedTypes.includes(file.type) &&
        !file.name.toLowerCase().endsWith(".csv")
      ) {
        toast.error("请选择Excel或CSV文件");
        return;
      }
      setImportFile(file);
    }
  };

  const handleImport = async () => {
    if (!importFile) return;

    try {
      setImporting(true);

      // Create FormData manually to ensure proper file upload
      const formData = new FormData();
      formData.append("file", importFile);

      // Debug: Log the FormData contents
      console.log("Import file:", importFile);
      console.log("FormData entries:");
      for (let [key, value] of formData.entries()) {
        console.log(key, value);
      }

      // Use axios directly for file upload to ensure proper handling
      const response = await apiClient.post("/activities/import", formData);

      toast.success("批量导入成功");
      setIsImportDialogOpen(false);
      setImportFile(null);
      fetchActivities();
    } catch (error: any) {
      console.error("Failed to import activities:", error);
      console.error("Error details:", error.response?.data);
      toast.error("批量导入失败");
    } finally {
      setImporting(false);
    }
  };

  const handleExport = async () => {
    try {
      const response = await apiClient.get("/activities/export", {
        responseType: "blob",
      });

      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement("a");
      link.href = url;
      link.setAttribute(
        "download",
        `activities_${new Date().toISOString().split("T")[0]}.xlsx`
      );
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);

      toast.success("导出成功");
    } catch (err) {
      toast.error("导出失败");
    }
  };

  const onSubmit = async (values: CreateActivityForm) => {
    try {
      if (editingActivity) {
        await apiClient.put(`/activities/${editingActivity.id}`, values);
        toast.success("活动更新成功");
      } else {
        await apiClient.post("/activities", values);
        toast.success("活动创建成功");
      }
      setIsDialogOpen(false);
      fetchActivities();
    } catch (error) {
      console.error("Failed to save activity:", error);
      toast.error(editingActivity ? "更新活动失败" : "创建活动失败");
    }
  };

  const getStatusText = (status: string) => {
    switch (status) {
      case "draft":
        return "草稿";
      case "pending_review":
        return "待审核";
      case "approved":
        return "已通过";
      case "rejected":
        return "已拒绝";
      default:
        return status;
    }
  };

  const getStatusStyle = (status: string) => {
    switch (status) {
      case "draft":
        return "bg-gray-100 text-gray-800";
      case "pending_review":
        return "bg-yellow-100 text-yellow-800";
      case "approved":
        return "bg-green-100 text-green-800";
      case "rejected":
        return "bg-red-100 text-red-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case "draft":
        return <FileText className="w-3 h-3 mr-1 inline" />;
      case "pending_review":
        return <Clock className="w-3 h-3 mr-1 inline" />;
      case "approved":
        return <CheckCircle className="w-3 h-3 mr-1 inline" />;
      case "rejected":
        return <XCircle className="w-3 h-3 mr-1 inline" />;
      default:
        return <AlertCircle className="w-3 h-3 mr-1 inline" />;
    }
  };

  // 获取所有活动类别
  const categories = Array.from(
    new Set(activities.map((a) => a.category).filter(Boolean))
  );

  return (
    <div className="space-y-8 p-4 md:p-8 bg-gray-50 min-h-screen">
      <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">活动列表</h1>
          <p className="text-muted-foreground">管理学分相关活动</p>
        </div>
        <div className="flex items-center gap-2">
          {(hasPermission("manage_activities") ||
            user?.userType === "student") && (
            <Button
              onClick={() => handleDialogOpen(null)}
              className="rounded-lg shadow transition-all duration-200 hover:scale-105"
            >
              <PlusCircle className="mr-2 h-4 w-4" />
              新建活动
            </Button>
          )}
          {(hasPermission("manage_activities") ||
            user?.userType === "admin") && (
            <>
              <Button
                onClick={() => setIsImportDialogOpen(true)}
                variant="outline"
                className="rounded-lg shadow transition-all duration-200 hover:scale-105"
              >
                <Upload className="mr-2 h-4 w-4" />
                批量导入
              </Button>
              <Button
                onClick={handleExport}
                variant="outline"
                className="rounded-lg shadow transition-all duration-200 hover:scale-105"
              >
                <Download className="mr-2 h-4 w-4" />
                导出数据
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard
          title="总活动数"
          value={totalItems}
          icon={Award}
          color="info"
          subtitle={`已通过: ${
            activities.filter((a) => a.status === "approved").length
          }`}
        />
        <StatCard
          title="参与学生"
          value={activities.reduce(
            (sum, activity) => sum + (activity.participants?.length || 0),
            0
          )}
          icon={Users}
          color="success"
          subtitle="总参与人次"
        />
        <StatCard
          title="申请数量"
          value={activities.reduce(
            (sum, activity) => sum + (activity.applications?.length || 0),
            0
          )}
          icon={FileText}
          color="purple"
          subtitle="总申请数量"
        />
        <StatCard
          title="待审核"
          value={activities.filter((a) => a.status === "pending_review").length}
          icon={Clock}
          color="warning"
          subtitle="待审核活动"
        />
      </div>

      {/* Filters */}
      <Card className="bg-white/80 backdrop-blur border-0 shadow-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Filter className="h-5 w-5" />
            筛选和搜索
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="搜索活动名称..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      handleSearchAndFilter();
                    }
                  }}
                  className="pl-10"
                />
              </div>
            </div>
            <Select value={categoryFilter} onValueChange={setCategoryFilter}>
              <SelectTrigger className="w-48">
                <SelectValue placeholder="选择类别" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部类别</SelectItem>
                {categories
                  .filter((category): category is string => Boolean(category))
                  .map((category) => (
                    <SelectItem key={category} value={category}>
                      {category}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-32">
                <SelectValue placeholder="状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="draft">草稿</SelectItem>
                <SelectItem value="pending_review">待审核</SelectItem>
                <SelectItem value="approved">已通过</SelectItem>
                <SelectItem value="rejected">已拒绝</SelectItem>
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              onClick={handleSearchAndFilter}
              disabled={loading}
            >
              <RefreshCw
                className={`h-4 w-4 ${loading ? "animate-spin" : ""}`}
              />
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Activities Table */}
      <Card className="bg-gray-100/80 dark:bg-gray-900/40 border-0 shadow-sm">
        <CardHeader>
          <CardTitle>活动列表</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-md bg-white dark:bg-gray-900/60">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>名称</TableHead>
                  <TableHead>类别</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>参与人数</TableHead>
                  <TableHead>申请数量</TableHead>
                  <TableHead>创建时间</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-8">
                      <div className="flex items-center justify-center gap-2">
                        <RefreshCw className="h-4 w-4 animate-spin" />
                        加载中...
                      </div>
                    </TableCell>
                  </TableRow>
                ) : activities.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-8">
                      <div className="flex flex-col items-center gap-2 text-muted-foreground">
                        <AlertCircle className="w-8 h-8" />
                        <p>暂无活动记录</p>
                      </div>
                    </TableCell>
                  </TableRow>
                ) : (
                  activities.map((activity) => (
                    <TableRow key={activity.id}>
                      <TableCell>
                        <div>
                          <div className="font-medium">{activity.title}</div>
                          <div className="text-sm text-muted-foreground max-w-xs truncate">
                            {activity.description || "暂无描述"}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <Badge variant="secondary" className="rounded">
                          {activity.category || "-"}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge
                          className={`rounded ${getStatusStyle(
                            activity.status
                          )}`}
                        >
                          {getStatusIcon(activity.status)}
                          {getStatusText(activity.status)}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <span className="font-bold text-blue-600">
                          {activity.participants?.length || 0}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="font-bold text-green-600">
                          {activity.applications?.length || 0}
                        </span>
                      </TableCell>
                      <TableCell>
                        {activity.created_at?.split("T")[0] || "-"}
                      </TableCell>
                      <TableCell className="text-right space-x-2">
                        <Button
                          variant="outline"
                          size="icon"
                          onClick={() => navigate(`/activities/${activity.id}`)}
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        {(hasPermission("manage_activities") ||
                          user?.user_id === activity.owner_id) && (
                          <>
                            <Button
                              variant="outline"
                              size="icon"
                              onClick={() => handleDialogOpen(activity)}
                            >
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="destructive"
                              size="icon"
                              onClick={() => handleDeleteClick(activity)}
                            >
                              <Trash className="h-4 w-4" />
                            </Button>
                          </>
                        )}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          {/* 分页组件 */}
          {!loading && totalItems > 0 && (
            <Pagination
              currentPage={currentPage}
              totalPages={totalPages}
              totalItems={totalItems}
              pageSize={pageSize}
              onPageChange={handlePageChange}
              onPageSizeChange={handlePageSizeChange}
            />
          )}
        </CardContent>
      </Card>

      {/* Create/Edit Dialog */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>
              {editingActivity ? "编辑活动" : "添加新活动"}
            </DialogTitle>
            <DialogDescription>
              {editingActivity ? "修改活动信息" : "创建新的活动"}
            </DialogDescription>
          </DialogHeader>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <FormField
                control={form.control}
                name="title"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>活动名称</FormLabel>
                    <FormControl>
                      <Input placeholder="请输入活动名称" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="category"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>活动类型</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="选择活动类型" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="创新创业实践活动">
                          创新创业实践活动
                        </SelectItem>
                        <SelectItem value="学科竞赛">学科竞赛</SelectItem>
                        <SelectItem value="大学生创业项目">
                          大学生创业项目
                        </SelectItem>
                        <SelectItem value="创业实践项目">
                          创业实践项目
                        </SelectItem>
                        <SelectItem value="论文专利">论文专利</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button type="submit" className="w-full">
                  {editingActivity ? "更新活动" : "创建活动"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertCircle className="h-5 w-5 text-red-500" />
              确认删除
            </DialogTitle>
            <DialogDescription>
              您确定要删除活动 "{deletingActivity?.title}" 吗？此操作不可撤销。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsDeleteDialogOpen(false)}
            >
              取消
            </Button>
            <Button variant="destructive" onClick={handleDeleteConfirm}>
              确认删除
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Import Dialog */}
      <Dialog open={isImportDialogOpen} onOpenChange={setIsImportDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Upload className="h-5 w-5" />
              批量导入活动
            </DialogTitle>
            <DialogDescription>
              请选择Excel或CSV文件进行批量导入。文件应包含活动的基本信息。
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  const link = document.createElement("a");
                  link.href = "/api/activities/csv-template";
                  link.download = "activity_template.csv";
                  document.body.appendChild(link);
                  link.click();
                  link.remove();
                }}
              >
                <Download className="mr-2 h-4 w-4" />
                下载CSV模板
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  const link = document.createElement("a");
                  link.href = "/api/activities/excel-template";
                  link.download = "activity_template.xlsx";
                  document.body.appendChild(link);
                  link.click();
                  link.remove();
                }}
              >
                <Download className="mr-2 h-4 w-4" />
                下载Excel模板
              </Button>
            </div>
            <div>
              <label className="text-sm font-medium">选择文件</label>
              <Input
                type="file"
                accept=".xlsx,.xls,.csv"
                onChange={handleImportFile}
                className="mt-1"
              />
              <p className="text-xs text-muted-foreground mt-1">
                支持Excel (.xlsx, .xls) 和CSV格式，文件大小不超过10MB
              </p>
            </div>
            {importFile && (
              <div className="p-3 bg-muted rounded-lg">
                <p className="text-sm font-medium">已选择文件：</p>
                <p className="text-sm text-muted-foreground">
                  {importFile.name}
                </p>
                <p className="text-xs text-muted-foreground">
                  大小：{(importFile.size / 1024 / 1024).toFixed(2)} MB
                </p>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setIsImportDialogOpen(false);
                setImportFile(null);
              }}
            >
              取消
            </Button>
            <Button
              onClick={handleImport}
              disabled={!importFile || importing}
              className="bg-blue-600 hover:bg-blue-700"
            >
              {importing ? (
                <div className="flex items-center gap-2">
                  <RefreshCw className="h-4 w-4 animate-spin" />
                  导入中...
                </div>
              ) : (
                <>
                  <Upload className="mr-2 h-4 w-4" />
                  开始导入
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
