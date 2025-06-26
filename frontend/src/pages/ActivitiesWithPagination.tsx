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
import apiClient from "@/lib/api";
import {
  PlusCircle,
  Edit,
  Trash,
  Search,
  Filter,
  RefreshCw,
  Award,
  Users,
  Calendar,
  AlertCircle,
  CheckCircle,
  XCircle,
  Eye,
  Upload,
  Download,
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
  participants_count?: number;
  applications_count?: number;
  owner_id: string;
  owner_info?: {
    name: string;
    role: string;
  };
}

type CreateActivityForm = z.infer<typeof activitySchema>;

const activitySchema = z.object({
  title: z
    .string()
    .min(1, "活动名称不能为空")
    .max(200, "活动名称不能超过200个字符"),
  category: z.string().min(1, "请选择活动类型"),
});

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
      console.log("API Response:", response.data);

      // 处理响应数据
      let activitiesData: Activity[] = [];
      let paginationData: any = {};

      if (response.data.data && response.data.data.data) {
        activitiesData = response.data.data.data;
        paginationData = {
          total: response.data.data.total || 0,
          page: response.data.data.page || 1,
          page_size: response.data.data.page_size || 10,
          total_pages: response.data.data.total_pages || 0,
        };
      } else if (response.data.data && Array.isArray(response.data.data)) {
        activitiesData = response.data.data;
        paginationData = {
          total: activitiesData.length,
          page: 1,
          page_size: activitiesData.length,
          total_pages: 1,
        };
      } else {
        console.warn("Unexpected response structure:", response.data);
        activitiesData = [];
        paginationData = {
          total: 0,
          page: 1,
          page_size: 10,
          total_pages: 0,
        };
      }

      setActivities(
        activitiesData.map((activity: any) => ({
          id: activity.id,
          title: activity.title,
          description: activity.description,
          status: activity.status,
          category: activity.category,
          created_at: activity.created_at,
          updated_at: activity.updated_at,
          participants_count: activity.participants?.length || 0,
          applications_count: activity.applications?.length || 0,
          owner_id: activity.owner_id,
          owner_info: activity.owner_info,
        }))
      );

      // 更新分页信息
      setTotalItems(paginationData.total);
      setTotalPages(paginationData.total_pages);
      setCurrentPage(paginationData.page);
      setPageSize(paginationData.page_size);
    } catch (err) {
      console.error("Failed to fetch activities:", err);
      toast.error("获取活动列表失败");
      // Fallback to mock data
      const mockActivities = [
        {
          id: "1",
          title: "创新创业项目",
          description: "参与各类创新创业项目，包括创业计划书撰写、项目路演等",
          status: "approved",
          category: "创新创业",
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          participants_count: 25,
          applications_count: 18,
          owner_id: "admin",
          owner_info: { name: "张三", role: "teacher" },
        },
      ];
      setActivities(mockActivities);
      setTotalItems(mockActivities.length);
      setTotalPages(1);
      setCurrentPage(1);
      setPageSize(10);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchActivities();
  }, []);

  // 处理分页变化
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    fetchActivities(page, pageSize);
  };

  // 处理每页数量变化
  const handlePageSizeChange = (size: number) => {
    setPageSize(size);
    setCurrentPage(1);
    fetchActivities(1, size);
  };

  // 处理搜索和筛选
  const handleSearchAndFilter = () => {
    setCurrentPage(1);
    fetchActivities(1, pageSize);
  };

  useEffect(() => {
    handleSearchAndFilter();
  }, [searchTerm, categoryFilter, statusFilter]);

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
      fetchActivities(currentPage, pageSize);
    } catch (err) {
      toast.error("删除活动失败");
    } finally {
      setIsDeleteDialogOpen(false);
      setDeletingActivity(null);
    }
  };

  const handleImportFile = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const allowedTypes = [
        "application/vnd.ms-excel",
        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "text/csv",
      ];
      if (!allowedTypes.includes(file.type)) {
        toast.error("请选择Excel或CSV文件");
        return;
      }
      setImportFile(file);
    }
  };

  const handleImport = async () => {
    if (!importFile) return;

    setImporting(true);
    try {
      const formData = new FormData();
      formData.append("file", importFile);

      const response = await apiClient.post("/activities/import", formData, {
        headers: {
          "Content-Type": "multipart/form-data",
        },
      });

      if (response.data.code === 0) {
        toast.success("批量导入成功");
        setIsImportDialogOpen(false);
        setImportFile(null);
        fetchActivities(currentPage, pageSize);
      } else {
        toast.error(response.data.message || "导入失败");
      }
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || "导入失败";
      toast.error(errorMessage);
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
        setIsDialogOpen(false);
        fetchActivities(currentPage, pageSize);
      } else {
        const response = await apiClient.post("/activities", values);
        const createdActivity = response.data.data;
        toast.success("活动创建成功");
        setIsDialogOpen(false);
        navigate(`/activities/${createdActivity.id}`);
      }
    } catch (err: any) {
      const errorMessage =
        err.response?.data?.message ||
        `活动${editingActivity ? "更新" : "创建"}失败`;
      toast.error(errorMessage);
    }
  };

  const categories = Array.from(
    new Set(activities.map((a) => a.category).filter(Boolean))
  );

  // 获取状态显示文本
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

  // 获取状态样式
  const getStatusStyle = (status: string) => {
    switch (status) {
      case "approved":
        return "bg-green-100 text-green-800";
      case "pending_review":
        return "bg-yellow-100 text-yellow-800";
      case "rejected":
        return "bg-red-100 text-red-800";
      case "draft":
        return "bg-gray-100 text-gray-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  // 获取状态图标
  const getStatusIcon = (status: string) => {
    switch (status) {
      case "approved":
        return <CheckCircle className="w-3 h-3 mr-1 inline" />;
      case "pending_review":
        return <AlertCircle className="w-3 h-3 mr-1 inline" />;
      case "rejected":
        return <XCircle className="w-3 h-3 mr-1 inline" />;
      default:
        return <AlertCircle className="w-3 h-3 mr-1 inline" />;
    }
  };

  return (
    <div className="space-y-8 p-4 md:p-8">
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
              className="rounded-lg shadow transition-all duration-200 hover:scale-105 bg-primary text-white font-bold"
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
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">总活动数</CardTitle>
            <Award className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{totalItems}</div>
            <p className="text-xs text-muted-foreground">
              已通过: {activities.filter((a) => a.status === "approved").length}
            </p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">参与学生</CardTitle>
            <Users className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {activities.reduce(
                (sum, activity) => sum + (activity.participants_count || 0),
                0
              )}
            </div>
            <p className="text-xs text-muted-foreground">总参与人次</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">申请数量</CardTitle>
            <Calendar className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {activities.reduce(
                (sum, activity) => sum + (activity.applications_count || 0),
                0
              )}
            </div>
            <p className="text-xs text-muted-foreground">总申请数量</p>
          </CardContent>
        </Card>
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">待审核</CardTitle>
            <AlertCircle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {activities.filter((a) => a.status === "pending_review").length}
            </div>
            <p className="text-xs text-muted-foreground">待审核活动</p>
          </CardContent>
        </Card>
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
                placeholder="搜索活动名称..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10 rounded-lg shadow-sm"
              />
            </div>
            <Select value={categoryFilter} onValueChange={setCategoryFilter}>
              <SelectTrigger className="w-40">
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
              onClick={() => fetchActivities(currentPage, pageSize)}
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

      {/* Activities Table */}
      <Card className="rounded-xl shadow-lg">
        <CardHeader>
          <CardTitle>活动列表</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-xl overflow-x-auto">
            <Table>
              <TableHeader className="bg-muted/60">
                <TableRow>
                  <TableHead>名称</TableHead>
                  <TableHead>描述</TableHead>
                  <TableHead>类别</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>创建人</TableHead>
                  <TableHead>参与人数</TableHead>
                  <TableHead>申请数</TableHead>
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
                ) : activities.length === 0 ? (
                  <tr>
                    <td colSpan={8} className="py-12">
                      <div className="flex flex-col items-center text-muted-foreground">
                        <AlertCircle className="w-12 h-12 mb-2" />
                        <p>暂无活动记录</p>
                      </div>
                    </td>
                  </tr>
                ) : (
                  activities.map((activity) => (
                    <TableRow
                      key={activity.id}
                      className="hover:bg-muted/40 transition-colors"
                    >
                      <TableCell className="font-medium">
                        {activity.title}
                      </TableCell>
                      <TableCell className="truncate max-w-xs">
                        {activity.description}
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant="secondary"
                          className="rounded px-2 py-1"
                        >
                          {activity.category || "-"}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <Badge
                          className={`${getStatusStyle(
                            activity.status
                          )} rounded-lg px-2 py-1`}
                        >
                          {getStatusIcon(activity.status)}
                          {getStatusText(activity.status)}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <div className="flex flex-col">
                          <span className="font-medium text-sm">
                            {activity.owner_info?.name || "未知用户"}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            {activity.owner_info?.role === "student"
                              ? "学生"
                              : activity.owner_info?.role === "teacher"
                              ? "教师"
                              : activity.owner_info?.role === "admin"
                              ? "管理员"
                              : "用户"}
                          </span>
                        </div>
                      </TableCell>
                      <TableCell>{activity.participants_count || 0}</TableCell>
                      <TableCell>{activity.applications_count || 0}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1">
                          <Button
                            size="icon"
                            variant="ghost"
                            className="rounded-full hover:bg-primary/10"
                            title="查看详情"
                            onClick={() =>
                              navigate(`/activities/${activity.id}`)
                            }
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                          {(user?.userType === "admin" ||
                            (user?.user_id === activity.owner_id &&
                              activity.status === "draft")) && (
                            <Button
                              size="icon"
                              variant="ghost"
                              className="rounded-full hover:bg-primary/10"
                              title="编辑"
                              onClick={() =>
                                navigate(`/activities/${activity.id}?edit=1`)
                              }
                            >
                              <Edit className="h-4 w-4" />
                            </Button>
                          )}
                          {(user?.userType === "admin" ||
                            user?.user_id === activity.owner_id) && (
                            <Button
                              size="icon"
                              variant="ghost"
                              className="rounded-full hover:bg-red-100"
                              title="删除"
                              onClick={() => handleDeleteClick(activity)}
                            >
                              <Trash className="h-4 w-4" />
                            </Button>
                          )}
                        </div>
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
