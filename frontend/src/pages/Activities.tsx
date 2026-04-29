import { useState, useEffect, useCallback } from "react";
import { useAuth } from "@/contexts/AuthContext";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { useNavigate, useSearchParams } from "react-router-dom";
import { getStatusText, getStatusStyle, getStatusIcon } from "@/lib/status-utils";
import React from "react";
import { StatCard } from "@/components/ui/stat-card";
import { FilterCard } from "@/components/ui/filter-card";
import type { Activity } from "@/types/activity";
import { getActivityOptions } from "@/lib/options";
import { ActivityEditDialog } from "@/components/activity-common";
import { TopProgressBar } from "@/components/ui/top-progress-bar";

export default function ActivitiesPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { user, hasPermission } = useAuth();
  const [activities, setActivities] = useState<Activity[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [searchTerm, setSearchTerm] = useState("");
  const [categoryFilter, setCategoryFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState("all");
  const [activityCategories, setActivityCategories] = useState<{ value: string; label: string }[]>([]);
  const [activityStatuses, setActivityStatuses] = useState<{ value: string; label: string }[]>([]);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingActivity, setEditingActivity] = useState<Activity | null>(null);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [deletingActivity, setDeletingActivity] = useState<Activity | null>(
    null
  );
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const [importFile, setImportFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);
  const [importErrors, setImportErrors] = useState<string[]>([]);

  // 分页状态
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [totalItems, setTotalItems] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [activityStats, setActivityStats] = useState({
    totalActivities: 0,
    approvedCount: 0,
    pendingCount: 0,
    totalParticipants: 0,
  });
  const [globalApplicationCount, setGlobalApplicationCount] = useState(0);

  // 根据URL参数设置初始过滤器
  useEffect(() => {
    const statusFromUrl = searchParams.get("status");
    if (statusFromUrl) {
      setStatusFilter(statusFromUrl);
    }
  }, [searchParams]);

  const fetchActivities = useCallback(
    async (
      page = currentPage,
      size = pageSize,
      options?: { isInitial?: boolean }
    ) => {
      // 显式由调用方控制是否为首次加载，避免依赖 loading 造成循环
      const isInitial = options?.isInitial ?? false;
      try {
        if (isInitial) {
          setLoading(true);
        } else {
          setRefreshing(true);
        }

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

      // 调试：打印响应数据
      console.log("API Response:", response.data);

      // 处理响应数据
      let activitiesData = [];
      let paginationData: any = {};

      if (response.data.code === 0 && response.data.data) {
        if (response.data.data.data && Array.isArray(response.data.data.data)) {
          // 分页数据结构
          activitiesData = response.data.data.data;
          paginationData = {
            total: response.data.data.total || 0,
            page: response.data.data.page || 1,
            page_size: response.data.data.page_size || 10,
            total_pages: response.data.data.total_pages || 0,
          };
        } else {
          // 非分页数据结构
          activitiesData = response.data.data.activities || response.data.data || [];
          paginationData = {
            total: activitiesData.length,
            page: 1,
            page_size: activitiesData.length,
            total_pages: 1,
          };
        }
      } else {
        activitiesData = [];
        paginationData = {
          total: 0,
          page: 1,
          page_size: 10,
          total_pages: 0,
        };
      }

      // 调试：打印处理后的数据
      console.log("Activities Data:", activitiesData);
      console.log("Is Array:", Array.isArray(activitiesData));
      
        setActivities(activitiesData);
        setTotalItems(paginationData.total);
        setTotalPages(paginationData.total_pages);
      } catch (error) {
        console.error("Failed to fetch activities:", error);
        toast.error("获取活动列表失败");
        setActivities([]);
        setTotalItems(0);
        setTotalPages(0);
      } finally {
        if (isInitial) {
          setLoading(false);
        } else {
          setRefreshing(false);
        }
      }
    },
    [currentPage, pageSize, searchTerm, categoryFilter, statusFilter]
  );

  const fetchActivityStats = useCallback(async () => {
    try {
      const response = await apiClient.get("/activities/stats");
      if (response.data.code === 0) {
        const data = response.data.data || {};
        setActivityStats({
          totalActivities: data.total_activities || 0,
          approvedCount: data.approved_count || 0,
          pendingCount: data.pending_count || 0,
          totalParticipants: data.total_participants || 0,
        });
      }
    } catch (error) {
      console.error("Failed to fetch activity stats:", error);
    }
  }, []);

  const fetchGlobalApplicationCount = useCallback(async () => {
    try {
      const endpoint =
        user?.userType === "student" ? "/applications" : "/applications/all";
      const response = await apiClient.get(endpoint, {
        params: { page: 1, page_size: 1 },
      });
      if (response.data.code === 0 && response.data.data) {
        const total =
          response.data.data.total ??
          response.data.data?.total_count ??
          (response.data.data.data?.length ?? 0);
        setGlobalApplicationCount(Number(total) || 0);
      }
    } catch (error) {
      console.error("Failed to fetch application stats:", error);
    }
  }, [user?.userType]);

  useEffect(() => {
    fetchActivityStats();
    fetchGlobalApplicationCount();
  }, [fetchActivityStats, fetchGlobalApplicationCount]);


  useEffect(() => {
    (async () => {
      try {
        const opts = await getActivityOptions();
        setActivityCategories(opts.categories || []);
        setActivityStatuses(opts.statuses || []);
      } catch (e) {
        console.error("Failed to load activity options", e);
      }
    })();
  }, []);

  const handlePageChange = useCallback(
    (page: number) => {
      setCurrentPage(page);
      fetchActivities(page, pageSize, { isInitial: false });
    },
    [fetchActivities, pageSize]
  );

  const handlePageSizeChange = useCallback(
    (size: number) => {
      setPageSize(size);
      setCurrentPage(1);
      fetchActivities(1, size, { isInitial: false });
    },
    [fetchActivities]
  );

  const handleSearchAndFilter = useCallback(
    (options?: { isInitial?: boolean }) => {
      setCurrentPage(1);
      fetchActivities(1, pageSize, options);
    },
    [fetchActivities, pageSize]
  );

  useEffect(() => {
    // 首次进入列表页使用 loading，其余筛选/分页使用 refreshing 细进度条
    handleSearchAndFilter({ isInitial: true });
  }, [handleSearchAndFilter]);

  // 根据 URL 参数自动打开编辑弹窗，例如 /activities?edit=123
  useEffect(() => {
    const editId = searchParams.get("edit");
    if (!editId) return;
    if (!activities || activities.length === 0) return;

    const target = activities.find((a) => String(a.id) === editId);
    if (target) {
      handleDialogOpen(target);
    }
  }, [searchParams, activities]);

  const handleDialogOpen = (activity: Activity | null) => {
    setEditingActivity(activity);
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
      fetchActivities(currentPage, pageSize, { isInitial: false });
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
    if (!importFile) {
      toast.error("请选择要导入的文件");
      return;
    }

    setImporting(true);
    setImportErrors([]);
    try {
      const formData = new FormData();
      formData.append("file", importFile);

      const response = await apiClient.post("/activities/import", formData, {
        headers: {
          // 让浏览器自动设置 multipart boundary
        },
      });

      if (response.data.code === 0) {
        toast.success("活动导入成功");
        setIsImportDialogOpen(false);
        setImportFile(null);
        setImportErrors([]);
        fetchActivities();
      } else {
        const errors =
          response.data.data?.errors || response.data.errors || [];
        if (Array.isArray(errors) && errors.length > 0) {
          setImportErrors(
            errors.map((e: any) =>
              typeof e === "string" ? e : JSON.stringify(e)
            )
          );
        } else {
          toast.error(response.data.message || "活动导入失败");
        }
      }
    } catch (error: any) {
      console.error("Failed to import activities:", error);
      const apiErrors =
        error?.response?.data?.data?.errors ||
        error?.response?.data?.errors ||
        [];
      if (Array.isArray(apiErrors) && apiErrors.length > 0) {
        setImportErrors(
          apiErrors.map((e: any) =>
            typeof e === "string" ? e : JSON.stringify(e)
          )
        );
      } else {
        toast.error(
          error?.response?.data?.message ||
            error?.response?.data?.error ||
            "活动导入失败"
        );
      }
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

  // 获取所有活动类别
  const safeActivities = Array.isArray(activities) ? activities : [];

  return (
    <div className="space-y-8 p-4 md:p-8">
      <TopProgressBar active={refreshing} />
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
          value={activityStats.totalActivities}
          icon={Award}
          color="info"
          subtitle={`已通过: ${activityStats.approvedCount}`}
        />
        <StatCard
          title="参与学生"
          value={activityStats.totalParticipants}
          icon={Users}
          color="success"
          subtitle="总参与人次"
        />
        <StatCard
          title="申请数量"
          value={globalApplicationCount}
          icon={FileText}
          color="purple"
          subtitle="总申请数量"
        />
        <StatCard
          title="待审核"
          value={activityStats.pendingCount}
          icon={Clock}
          color="warning"
          subtitle="待审核活动"
        />
      </div>

      {/* Filters */}
      <FilterCard icon={Filter} contentClassName="space-y-0">
        <div className="flex flex-col md:flex-row gap-4 items-center">
          <div className="relative w-full max-w-xs">
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
              className="pl-10 rounded-lg shadow-sm"
            />
          </div>
          <Select value={categoryFilter} onValueChange={setCategoryFilter}>
            <SelectTrigger className="w-48 rounded-lg">
              <SelectValue placeholder="选择类别" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部类别</SelectItem>
              {activityCategories.map((c) => (
                <SelectItem key={c.value} value={c.value}>{c.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="w-32 rounded-lg">
              <SelectValue placeholder="状态" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部状态</SelectItem>
              {activityStatuses.map((s) => (
                <SelectItem key={s.value} value={s.value}>{s.label}</SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            variant="outline"
            onClick={() => handleSearchAndFilter({ isInitial: false })}
            disabled={loading}
            className="rounded-lg shadow"
          >
            <Search className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            onClick={() => {
              setSearchTerm("");
              setCategoryFilter("all");
              setStatusFilter("all");
              setCurrentPage(1);
              fetchActivities(1, pageSize);
            }}
            disabled={loading}
            className="rounded-lg shadow"
          >
            <RefreshCw
              className={`h-4 w-4 ${loading ? "animate-spin" : ""}`}
            />
          </Button>
        </div>
      </FilterCard>

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
                    <TableCell colSpan={7} className="py-8 text-center">
                      <div className="flex flex-col items-center gap-2 text-muted-foreground">
                        <RefreshCw className="h-6 w-6 animate-spin" />
                        <span>加载中...</span>
                      </div>
                    </TableCell>
                  </TableRow>
                ) : !safeActivities || safeActivities.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="py-12 text-center">
                      <div className="flex flex-col items-center text-muted-foreground">
                        <AlertCircle className="w-10 h-10 mb-2" />
                        <p>暂无活动记录</p>
                      </div>
                    </TableCell>
                  </TableRow>
                ) : (
                  safeActivities.map((activity) => (
                    <TableRow
                      key={activity.id}
                      className="hover:bg-muted/40 transition-colors"
                    >
                      <TableCell>
                        <div>
                          <div className="font-medium max-w-md truncate" title={activity.title}>
                            {activity.title}
                          </div>
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
                          {React.createElement(getStatusIcon(activity.status))}
                          {getStatusText(activity.status)}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        <span className="font-bold text-blue-600 dark:text-blue-400">
                          {activity.participants_count ??
                            activity.participants?.length ??
                            0}
                        </span>
                      </TableCell>
                      <TableCell>
                        <span className="font-bold text-green-600 dark:text-green-400">
                          {activity.applications_count ??
                            activity.applications?.length ??
                            0}
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
                          user?.uuid === activity.owner_id) && (
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

        </CardContent>
      </Card>

      {/* 分页组件 */}
      {!loading && totalItems > 0 && (
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

      {/* Create/Edit Dialog — 统一使用 ActivityEditDialog 组件 */}
      <ActivityEditDialog
        open={isDialogOpen}
        onOpenChange={(open) => {
          setIsDialogOpen(open);
          if (!open) {
            setEditingActivity(null);
          }
        }}
        activity={editingActivity}
        onSuccess={fetchActivities}
      />

      {/* Delete Confirmation Dialog */}
      <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
                                      <AlertCircle className="h-5 w-5 text-red-500 dark:text-red-400" />
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
        <DialogContent className="sm:max-w-[520px]">
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
                  apiHelpers.downloadFile(
                    "/activities/csv-template",
                    "activity_template.csv"
                  );
                }}
              >
                <Download className="mr-2 h-4 w-4" />
                下载CSV模板
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  apiHelpers.downloadFile(
                    "/activities/excel-template",
                    "activity_template.xlsx"
                  );
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
              <div className="p-3 bg-muted rounded-lg space-y-1">
                <p className="text-sm font-medium">已选择文件：</p>
                <p className="text-sm text-muted-foreground">
                  {importFile.name}
                </p>
                <p className="text-xs text-muted-foreground">
                  大小：{(importFile.size / 1024 / 1024).toFixed(2)} MB
                </p>
              </div>
            )}
            {importErrors.length > 0 && (
              <div className="p-3 border rounded-lg bg-muted/60 max-h-60 overflow-y-auto">
                <p className="text-sm font-medium mb-2">
                  导入失败，共 {importErrors.length} 条错误：
                </p>
                <ul className="space-y-1 text-xs text-red-600">
                  {importErrors.map((msg, idx) => (
                    <li key={idx} className="whitespace-pre-wrap">
                      {msg}
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setIsImportDialogOpen(false);
                setImportFile(null);
                  setImportErrors([]);
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
