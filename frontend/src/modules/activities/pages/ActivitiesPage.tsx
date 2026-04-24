import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Eye, Plus, RefreshCw, Search } from "lucide-react";
import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import toast from "react-hot-toast";

import {
  createActivity,
  listActivities,
} from "@/modules/activities/api/activities";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  DialogDescription,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

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

export default function ActivitiesPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [keyword, setKeyword] = useState("");
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [draftForm, setDraftForm] = useState({
    title: "",
    description: "",
    category: "",
    academic_year: "",
    semester: "",
  });

  const query = useQuery({
    queryKey: ["activities", keyword],
    queryFn: () =>
      listActivities({
        keyword: keyword || undefined,
      }),
  });

  const categories = useMemo(() => {
    const values = new Set(query.data?.items.map((item) => item.category) || []);
    return Array.from(values);
  }, [query.data?.items]);

  const createMutation = useMutation({
    mutationFn: () =>
      createActivity({
        title: draftForm.title,
        description: draftForm.description || undefined,
        category: draftForm.category,
        academic_year: draftForm.academic_year,
        semester: draftForm.semester,
        start_date: undefined,
        end_date: undefined,
        details: {},
      }),
    onSuccess: async (activity) => {
      await queryClient.invalidateQueries({ queryKey: ["activities"] });
      toast.success("活动草稿已创建");
      setIsCreateOpen(false);
      setDraftForm({
        title: "",
        description: "",
        category: "",
        academic_year: "",
        semester: "",
      });
      navigate(`/activities/${activity.id}`);
    },
    onError: (error) => {
      const message =
        error instanceof Error ? error.message : "创建活动失败，请稍后重试";
      toast.error(message);
    },
  });

  return (
    <div className="space-y-6">
      <Card className="border-none bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-slate-50 shadow-xl">
        <CardHeader className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
          <div>
            <CardTitle className="text-2xl">活动列表</CardTitle>
            <p className="mt-2 text-sm text-slate-300">
              新版活动列表直接基于模块化单体 API 渲染，作为后续编辑、审核和导入流的入口。
            </p>
          </div>
          <div className="flex items-center gap-3">
            <div className="relative">
              <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
              <Input
                className="w-72 border-slate-700 bg-slate-950/70 pl-10 text-slate-50 placeholder:text-slate-500"
                placeholder="搜索活动标题"
                value={keyword}
                onChange={(event) => setKeyword(event.target.value)}
              />
            </div>
            <Button
              variant="secondary"
              className="gap-2"
              onClick={() => query.refetch()}
            >
              <RefreshCw className="h-4 w-4" />
              刷新
            </Button>
            <Button className="gap-2" onClick={() => setIsCreateOpen(true)}>
              <Plus className="h-4 w-4" />
              新建活动
            </Button>
          </div>
        </CardHeader>
        <CardContent className="flex flex-wrap gap-3">
          {categories.map((category) => (
            <Badge key={category} variant="secondary" className="bg-white/10 text-slate-100">
              {category}
            </Badge>
          ))}
          <Badge variant="outline" className="border-slate-700 text-slate-100">
            共 {query.data?.total ?? 0} 项
          </Badge>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>活动数据</CardTitle>
        </CardHeader>
        <CardContent>
          {query.isLoading ? (
            <div className="py-10 text-sm text-muted-foreground">正在加载活动数据...</div>
          ) : query.isError ? (
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
                {query.data?.items.map((activity) => (
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
                      <Button
                        variant="ghost"
                        size="sm"
                        className="gap-2"
                        onClick={() => navigate(`/activities/${activity.id}`)}
                      >
                        <Eye className="h-4 w-4" />
                        查看
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>创建活动草稿</DialogTitle>
            <DialogDescription>
              填写基础字段后立即创建一个新的 draft 活动，后续在详情页继续编辑参与者、附件和审核流程。
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-2">
            <div className="space-y-2">
              <Label htmlFor="create-title">标题</Label>
              <Input
                id="create-title"
                aria-label="标题"
                value={draftForm.title}
                onChange={(event) =>
                  setDraftForm((current) => ({ ...current, title: event.target.value }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="create-description">描述</Label>
              <Input
                id="create-description"
                aria-label="描述"
                value={draftForm.description}
                onChange={(event) =>
                  setDraftForm((current) => ({
                    ...current,
                    description: event.target.value,
                  }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="create-category">分类</Label>
              <Input
                id="create-category"
                aria-label="分类"
                value={draftForm.category}
                onChange={(event) =>
                  setDraftForm((current) => ({
                    ...current,
                    category: event.target.value,
                  }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="create-academic-year">学年</Label>
              <Input
                id="create-academic-year"
                aria-label="学年"
                value={draftForm.academic_year}
                onChange={(event) =>
                  setDraftForm((current) => ({
                    ...current,
                    academic_year: event.target.value,
                  }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="create-semester">学期</Label>
              <Input
                id="create-semester"
                aria-label="学期"
                value={draftForm.semester}
                onChange={(event) =>
                  setDraftForm((current) => ({
                    ...current,
                    semester: event.target.value,
                  }))
                }
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsCreateOpen(false)}>
              取消
            </Button>
            <Button disabled={createMutation.isPending} onClick={() => createMutation.mutate()}>
              创建草稿
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
