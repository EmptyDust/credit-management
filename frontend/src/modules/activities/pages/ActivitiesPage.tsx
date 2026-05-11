import { useMutation, useQuery } from "@tanstack/react-query";
import {
  Copy,
  Download,
  Edit,
  Eye,
  Plus,
  RefreshCw,
  Search,
  Trash2,
  Upload,
} from "lucide-react";
import { useMemo, useState } from "react";
import toast from "react-hot-toast";
import { useNavigate } from "react-router-dom";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Pagination } from "@/components/ui/pagination";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  exportActivities,
  getActivityStats,
  type ActivityStats,
  type ActivityStatsFilters,
} from "@/modules/activities/api/activity-stats";
import {
  copyActivity,
  deleteActivity,
  listActivities,
  type ActivityListFilters,
  type ActivityListItem,
} from "@/modules/activities/api/activities";
import ActivityImportDialog from "@/modules/activities/components/ActivityImportDialog";

const statusOptions = [
  { value: "draft", label: "草稿" },
  { value: "pending_review", label: "审核中" },
  { value: "approved", label: "已通过" },
  { value: "rejected", label: "已拒绝" },
];

const selectClassName =
  "h-10 rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2";

function statusLabel(status: string) {
  switch (status) {
    case "draft":
      return "草稿";
    case "pending_review":
      return "审核中";
    case "approved":
      return "已通过";
    case "rejected":
      return "已拒绝";
    default:
      return status;
  }
}

function statusCount(stats: ActivityStats | undefined, status: string) {
  return stats?.status_counts?.[status] ?? 0;
}

function totalCount(stats: ActivityStats | undefined) {
  return Object.values(stats?.status_counts ?? {}).reduce(
    (total, count) => total + count,
    0
  );
}

export default function ActivitiesPage() {
  const navigate = useNavigate();
  const [keyword, setKeyword] = useState("");
  const [status, setStatus] = useState("");
  const [category, setCategory] = useState("");
  const [academicYear, setAcademicYear] = useState("");
  const [semester, setSemester] = useState("");
  const [importOpen, setImportOpen] = useState(false);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);

  const filters: ActivityStatsFilters = useMemo(
    () => ({
      keyword: keyword.trim() || undefined,
      status: status || undefined,
      category: category || undefined,
      academic_year: academicYear || undefined,
      semester: semester || undefined,
    }),
    [academicYear, category, keyword, semester, status]
  );
  const listFilters: ActivityListFilters = {
    ...filters,
    page,
    page_size: pageSize,
  };

  const listQuery = useQuery({
    queryKey: ["activities", listFilters],
    queryFn: () => listActivities(listFilters),
  });

  const statsQuery = useQuery({
    queryKey: ["activities", "stats", filters],
    queryFn: () => getActivityStats(filters),
  });

  const exportMutation = useMutation({
    mutationFn: () => exportActivities({ ...filters, format: "xlsx" }),
    onSuccess: () => {
      toast.success("活动导出已开始");
    },
    onError: (error) => {
      const message =
        error instanceof Error ? error.message : "活动导出失败，请稍后重试";
      toast.error(message);
    },
  });

  const refreshWorkspace = () => {
    void listQuery.refetch();
    void statsQuery.refetch();
  };

  const activityActionMutation = useMutation({
    mutationFn: async (action: { type: "copy" | "delete"; activityID: string }) => {
      if (action.type === "copy") {
        return copyActivity(action.activityID);
      }
      await deleteActivity(action.activityID);
      return null;
    },
    onSuccess: (_, action) => {
      toast.success(action.type === "copy" ? "活动已复制为草稿" : "活动已删除");
      refreshWorkspace();
    },
    onError: (error) => {
      const message =
        error instanceof Error ? error.message : "活动操作失败，请稍后重试";
      toast.error(message);
    },
  });

  const resetPage = () => setPage(1);

  const handleDelete = (activity: ActivityListItem) => {
    if (!window.confirm(`确定删除活动「${activity.title}」吗？`)) {
      return;
    }
    activityActionMutation.mutate({ type: "delete", activityID: activity.id });
  };

  const categories = statsQuery.data?.category_counts ?? [];
  const academicYears = statsQuery.data?.academic_year_counts ?? [];
  const semesters = statsQuery.data?.semester_counts ?? [];

  return (
    <div className="space-y-6">
      <Card className="border-none bg-gradient-to-br from-slate-950 via-slate-900 to-slate-800 text-slate-50 shadow-xl">
        <CardHeader className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
          <div>
            <CardTitle className="text-3xl">活动工作台</CardTitle>
            <CardDescription className="mt-3 max-w-2xl text-slate-300">
              从这里进入 schema 驱动的新建、草稿编辑、统计、导入和导出流程。
            </CardDescription>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <Button
              className="gap-2 bg-white text-slate-950 hover:bg-slate-100"
              onClick={() => navigate("/activities/new")}
            >
              <Plus className="h-4 w-4" />
              新建活动
            </Button>
            <Button
              className="gap-2 border-slate-600 bg-slate-950/40 text-slate-50 hover:bg-slate-800"
              variant="outline"
              onClick={() => setImportOpen(true)}
            >
              <Upload className="h-4 w-4" />
              导入
            </Button>
            <Button
              className="gap-2 border-slate-600 bg-slate-950/40 text-slate-50 hover:bg-slate-800"
              variant="outline"
              disabled={exportMutation.isPending}
              onClick={() => exportMutation.mutate()}
            >
              <Download className="h-4 w-4" />
              导出
            </Button>
            <Button
              className="gap-2"
              variant="secondary"
              onClick={refreshWorkspace}
            >
              <RefreshCw className="h-4 w-4" />
              刷新
            </Button>
          </div>
        </CardHeader>
      </Card>

      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              活动总数
            </CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-semibold">
            {totalCount(statsQuery.data)}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              我的活动
            </CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-semibold">
            {statsQuery.data?.mine_count ?? 0}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              已参与
            </CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-semibold">
            {statsQuery.data?.participated_count ?? 0}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              审核中
            </CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-semibold">
            {statusCount(statsQuery.data, "pending_review")}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              已通过
            </CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-semibold">
            {statusCount(statsQuery.data, "approved")}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>筛选</CardTitle>
          <CardDescription>筛选条件同时作用于列表、统计和导出。</CardDescription>
        </CardHeader>
        <CardContent className="grid gap-4 lg:grid-cols-[minmax(260px,1fr)_repeat(4,180px)]">
          <div className="space-y-2">
            <Label htmlFor="activity-keyword">关键字</Label>
            <div className="relative">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                id="activity-keyword"
                className="pl-10"
                placeholder="搜索活动标题"
                value={keyword}
                onChange={(event) => {
                  setKeyword(event.target.value);
                  resetPage();
                }}
              />
            </div>
          </div>
          <div className="space-y-2">
            <Label htmlFor="activity-status">状态</Label>
            <select
              id="activity-status"
              className={selectClassName}
              value={status}
              onChange={(event) => {
                setStatus(event.target.value);
                resetPage();
              }}
            >
              <option value="">全部状态</option>
              {statusOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="activity-category">分类</Label>
            <select
              id="activity-category"
              className={selectClassName}
              value={category}
              onChange={(event) => {
                setCategory(event.target.value);
                resetPage();
              }}
            >
              <option value="">全部分类</option>
              {categories.map((item) => (
                <option key={item.key} value={item.key}>
                  {item.label || item.key}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="activity-academic-year">学年</Label>
            <select
              id="activity-academic-year"
              className={selectClassName}
              value={academicYear}
              onChange={(event) => {
                setAcademicYear(event.target.value);
                resetPage();
              }}
            >
              <option value="">全部学年</option>
              {academicYears.map((item) => (
                <option key={item.value} value={item.value}>
                  {item.value}
                </option>
              ))}
            </select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="activity-semester">学期</Label>
            <select
              id="activity-semester"
              className={selectClassName}
              value={semester}
              onChange={(event) => {
                setSemester(event.target.value);
                resetPage();
              }}
            >
              <option value="">全部学期</option>
              {semesters.map((item) => (
                <option key={item.value} value={item.value}>
                  {item.value}
                </option>
              ))}
            </select>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <CardTitle>活动数据</CardTitle>
            <CardDescription>
              共 {listQuery.data?.total ?? 0} 项，点击查看进入详情页。
            </CardDescription>
          </div>
          <div className="flex flex-wrap gap-2">
            {categories.slice(0, 4).map((item) => (
              <Badge key={item.key} variant="secondary">
                {item.label || item.key}: {item.count}
              </Badge>
            ))}
          </div>
        </CardHeader>
        <CardContent>
          {listQuery.isLoading ? (
            <div className="py-10 text-sm text-muted-foreground">
              正在加载活动数据...
            </div>
          ) : listQuery.isError ? (
            <div className="py-10 text-sm text-destructive">活动列表加载失败。</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>标题</TableHead>
                  <TableHead>分类</TableHead>
                  <TableHead>学年 / 学期</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {listQuery.data?.items.map((activity) => (
                  <TableRow key={activity.id}>
                    <TableCell>
                      <div className="font-medium">{activity.title}</div>
                      {activity.description ? (
                        <div className="mt-1 text-xs text-muted-foreground">
                          {activity.description}
                        </div>
                      ) : null}
                    </TableCell>
                    <TableCell>{activity.category}</TableCell>
                    <TableCell>
                      {activity.academic_year} / {activity.semester}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{statusLabel(activity.status)}</Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap items-center gap-2">
                        <Button
                          className="gap-2"
                          size="sm"
                          variant="ghost"
                          onClick={() => navigate(`/activities/${activity.id}`)}
                        >
                          <Eye className="h-4 w-4" />
                          查看
                        </Button>
                        <Button
                          className="gap-2"
                          size="sm"
                          variant="outline"
                          aria-label={`复制 ${activity.title}`}
                          onClick={() =>
                            activityActionMutation.mutate({
                              type: "copy",
                              activityID: activity.id,
                            })
                          }
                        >
                          <Copy className="h-4 w-4" />
                          复制
                        </Button>
                        {activity.status === "draft" ? (
                          <Button
                            className="gap-2"
                            size="sm"
                            variant="outline"
                            aria-label={`编辑 ${activity.title}`}
                            onClick={() => navigate(`/activities/${activity.id}/edit`)}
                          >
                            <Edit className="h-4 w-4" />
                            编辑
                          </Button>
                        ) : null}
                        <Button
                          className="gap-2"
                          size="sm"
                          variant="destructive"
                          aria-label={`删除 ${activity.title}`}
                          onClick={() => handleDelete(activity)}
                        >
                          <Trash2 className="h-4 w-4" />
                          删除
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
          {listQuery.data ? (
            <Pagination
              currentPage={listQuery.data.page}
              totalPages={listQuery.data.total_pages}
              totalItems={listQuery.data.total}
              pageSize={listQuery.data.page_size}
              onPageChange={setPage}
              onPageSizeChange={(nextPageSize) => {
                setPageSize(nextPageSize);
                setPage(1);
              }}
            />
          ) : null}
        </CardContent>
      </Card>

      <ActivityImportDialog
        open={importOpen}
        onOpenChange={setImportOpen}
        onCompleted={refreshWorkspace}
      />
    </div>
  );
}
