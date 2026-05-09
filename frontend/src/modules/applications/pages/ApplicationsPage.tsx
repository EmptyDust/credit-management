import { useQuery } from "@tanstack/react-query";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Download, Eye, RefreshCw } from "lucide-react";
import { useState } from "react";

import {
  exportApplications,
  getApplicationDetail,
  getApplicationStats,
  listApplications,
  type ApplicationFilters,
} from "@/modules/applications/api/applications";

function statusLabel(status: string) {
  switch (status) {
    case "approved":
      return "已通过";
    case "pending_review":
      return "审核中";
    case "draft":
      return "草稿";
    case "rejected":
      return "已拒绝";
    default:
      return status;
  }
}

function buildFilters(input: {
  keyword: string;
  academicYear: string;
  semester: string;
  status: string;
}): ApplicationFilters {
  const filters: ApplicationFilters = {};
  const keyword = input.keyword.trim();
  const academicYear = input.academicYear.trim();
  const semester = input.semester.trim();
  const status = input.status.trim();

  if (keyword) {
    filters.keyword = keyword;
  }
  if (academicYear) {
    filters.academic_year = academicYear;
  }
  if (semester) {
    filters.semester = semester;
  }
  if (status) {
    filters.status = status;
  }

  return filters;
}

export default function ApplicationsPage() {
  const [selectedApplicationID, setSelectedApplicationID] = useState<string | null>(
    null
  );
  const [keyword, setKeyword] = useState("");
  const [academicYear, setAcademicYear] = useState("");
  const [semester, setSemester] = useState("");
  const [status, setStatus] = useState("");
  const filters = buildFilters({ keyword, academicYear, semester, status });

  const listQuery = useQuery({
    queryKey: ["applications", filters],
    queryFn: () => listApplications(filters),
  });

  const statsQuery = useQuery({
    queryKey: ["applications", "stats", filters],
    queryFn: () => getApplicationStats(filters),
  });

  const detailQuery = useQuery({
    queryKey: ["applications", selectedApplicationID],
    queryFn: () => getApplicationDetail(selectedApplicationID as string),
    enabled: selectedApplicationID !== null,
  });

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>申请总数</CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-semibold">
            {statsQuery.data?.total ?? 0}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>通过数量</CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-semibold">
            {statsQuery.data?.approved_count ?? 0}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>累计学分</CardTitle>
          </CardHeader>
          <CardContent className="text-3xl font-semibold">
            {statsQuery.data?.total_credits ?? 0}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="gap-4">
          <div className="flex flex-row items-center justify-between gap-4">
            <CardTitle>申请列表</CardTitle>
            <div className="flex flex-wrap justify-end gap-2">
              <Button
                variant="outline"
                className="gap-2"
                onClick={() => {
                  void exportApplications(filters);
                }}
              >
                <Download className="h-4 w-4" />
                导出申请
              </Button>
              <Button
                variant="secondary"
                className="gap-2"
                onClick={() => {
                  void listQuery.refetch();
                  void statsQuery.refetch();
                }}
              >
                <RefreshCw className="h-4 w-4" />
                刷新
              </Button>
            </div>
          </div>
          <div className="grid gap-3 md:grid-cols-4">
            <Input
              aria-label="申请搜索关键字"
              placeholder="搜索活动、用户或申请"
              value={keyword}
              onChange={(event) => setKeyword(event.target.value)}
            />
            <Input
              aria-label="申请学年"
              placeholder="学年，如 2025-2026"
              value={academicYear}
              onChange={(event) => setAcademicYear(event.target.value)}
            />
            <Input
              aria-label="申请学期"
              placeholder="学期，如 autumn"
              value={semester}
              onChange={(event) => setSemester(event.target.value)}
            />
            <Input
              aria-label="申请状态"
              placeholder="状态，如 approved"
              value={status}
              onChange={(event) => setStatus(event.target.value)}
            />
          </div>
        </CardHeader>
        <CardContent>
          {listQuery.isLoading ? (
            <div className="py-10 text-sm text-muted-foreground">正在加载申请数据...</div>
          ) : listQuery.isError ? (
            <div className="py-10 text-sm text-destructive">申请列表加载失败。</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>活动</TableHead>
                  <TableHead>学年 / 学期</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>学分</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {listQuery.data?.items.map((application) => (
                  <TableRow key={application.id}>
                    <TableCell>
                      <div className="font-medium">
                        {application.activity?.title || "未命名活动"}
                      </div>
                      <div className="mt-1 text-xs text-muted-foreground">
                        {application.activity?.category || "未分类"}
                      </div>
                    </TableCell>
                    <TableCell>
                      {application.academic_year || "-"} / {application.semester || "-"}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{statusLabel(application.status)}</Badge>
                    </TableCell>
                    <TableCell>
                      {application.awarded_credits || application.applied_credits}
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="gap-2"
                        onClick={() => setSelectedApplicationID(application.id)}
                      >
                        <Eye className="h-4 w-4" />
                        详情
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog
        open={selectedApplicationID !== null}
        onOpenChange={(open) => {
          if (!open) {
            setSelectedApplicationID(null);
          }
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>申请详情</DialogTitle>
          </DialogHeader>
          {detailQuery.isLoading ? (
            <div className="text-sm text-muted-foreground">正在加载详情...</div>
          ) : detailQuery.data ? (
            <div className="space-y-3 text-sm">
              <div>
                <span className="font-medium">活动：</span>
                {detailQuery.data.activity?.title || "未命名活动"}
              </div>
              <div>
                <span className="font-medium">状态：</span>
                {statusLabel(detailQuery.data.status)}
              </div>
              <div>
                <span className="font-medium">学分：</span>
                {detailQuery.data.awarded_credits || detailQuery.data.applied_credits}
              </div>
              <div>
                <span className="font-medium">学年 / 学期：</span>
                {detailQuery.data.academic_year || "-"} /{" "}
                {detailQuery.data.semester || "-"}
              </div>
            </div>
          ) : (
            <div className="text-sm text-destructive">申请详情加载失败。</div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}
