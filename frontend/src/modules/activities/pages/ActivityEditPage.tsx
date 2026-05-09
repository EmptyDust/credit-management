import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, FileText, Send } from "lucide-react";
import { useState } from "react";
import toast from "react-hot-toast";
import { useNavigate, useParams } from "react-router-dom";

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
import { PageHeader } from "@/components/ui/page-header";
import { Textarea } from "@/components/ui/textarea";
import { getActivityCategorySchema } from "@/modules/activities/api/activity-categories";
import {
  getActivityDetail,
  submitActivity,
  updateActivity,
  type ActivityDetail,
} from "@/modules/activities/api/activities";
import ActivitySchemaForm from "@/modules/activities/components/ActivitySchemaForm";
import type {
  ActivitySchemaNormalizedValues,
  ActivitySchemaValues,
} from "@/modules/activities/schema/types";

type ActivityFixedFormState = {
  title: string;
  description: string;
  category: string;
  academic_year: string;
  semester: string;
  start_date: string;
  end_date: string;
};

const semesterOptions = [
  { value: "autumn", label: "秋季学期" },
  { value: "spring", label: "春季学期" },
  { value: "summer", label: "夏季学期" },
];

const selectClassName =
  "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50";

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

function createFixedFormState(activity: ActivityDetail): ActivityFixedFormState {
  return {
    title: activity.title || "",
    description: activity.description || "",
    category: activity.category || "",
    academic_year: activity.academic_year || "",
    semester: activity.semester || "",
    start_date: activity.start_date || "",
    end_date: activity.end_date || "",
  };
}

function toSchemaValues(details?: Record<string, unknown>): ActivitySchemaValues {
  return (details || {}) as ActivitySchemaValues;
}

function toMutationInput(
  form: ActivityFixedFormState,
  details: ActivitySchemaNormalizedValues
) {
  return {
    title: form.title.trim(),
    description: form.description.trim() || undefined,
    category: form.category,
    academic_year: form.academic_year.trim(),
    semester: form.semester,
    start_date: form.start_date || undefined,
    end_date: form.end_date || undefined,
    details,
  };
}

function ActivityEditForm({ activity }: { activity: ActivityDetail }) {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [form, setForm] = useState<ActivityFixedFormState>(() =>
    createFixedFormState(activity)
  );
  const isDraft = activity.status === "draft";

  const schemaQuery = useQuery({
    queryKey: ["activity-category-schema", form.category],
    queryFn: () => getActivityCategorySchema(form.category),
    enabled: form.category.length > 0,
  });

  const updateMutation = useMutation({
    mutationFn: (details: ActivitySchemaNormalizedValues) =>
      updateActivity(activity.id, toMutationInput(form, details)),
    onSuccess: (updatedActivity) => {
      queryClient.setQueryData(["activity", activity.id], updatedActivity);
      toast.success("活动已保存");
    },
    onError: (error) => {
      const message =
        error instanceof Error ? error.message : "保存活动失败，请稍后重试";
      toast.error(message);
    },
  });

  const submitMutation = useMutation({
    mutationFn: () => submitActivity(activity.id),
    onSuccess: (updatedActivity) => {
      queryClient.setQueryData(["activity", activity.id], updatedActivity);
      toast.success("活动已提交审核");
      navigate(`/activities/${activity.id}`);
    },
    onError: (error) => {
      const message =
        error instanceof Error ? error.message : "提交审核失败，请稍后重试";
      toast.error(message);
    },
  });

  const updateForm = (key: keyof ActivityFixedFormState, value: string) => {
    setForm((current) => ({ ...current, [key]: value }));
  };

  const handleSave = async (details: ActivitySchemaNormalizedValues) => {
    if (
      !form.title.trim() ||
      !form.category ||
      !form.academic_year.trim() ||
      !form.semester
    ) {
      toast.error("请先填写标题、分类、学年和学期");
      return;
    }

    await updateMutation.mutateAsync(details);
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="编辑活动"
        description="草稿阶段可编辑基础字段和分类动态字段，提交审核后由后端权限规则控制后续操作。"
        actions={
          <>
            <Button
              className="gap-2"
              variant="outline"
              onClick={() => navigate(`/activities/${activity.id}`)}
            >
              <ArrowLeft className="h-4 w-4" />
              返回详情
            </Button>
            {isDraft ? (
              <Button
                className="gap-2"
                disabled={submitMutation.isPending}
                onClick={() => submitMutation.mutate()}
              >
                <Send className="h-4 w-4" />
                提交审核
              </Button>
            ) : null}
          </>
        }
      />

      <Card className="border-slate-200 bg-gradient-to-br from-slate-950 via-slate-900 to-slate-800 text-slate-50">
        <CardContent className="flex flex-col gap-3 p-6 md:flex-row md:items-center md:justify-between">
          <div>
            <div className="text-sm text-slate-300">当前状态</div>
            <div className="mt-1 text-lg font-semibold">{activity.title}</div>
          </div>
          <Badge variant="secondary" className="bg-white/10 text-slate-50">
            {statusLabel(activity.status)}
          </Badge>
        </CardContent>
      </Card>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,420px)_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>基础信息</CardTitle>
            <CardDescription>
              只有草稿活动可直接保存；非草稿状态会进入只读展示。
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4">
            <div className="space-y-2">
              <Label htmlFor="activity-title">活动标题</Label>
              <Input
                id="activity-title"
                disabled={!isDraft}
                value={form.title}
                onChange={(event) => updateForm("title", event.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="activity-description">活动描述</Label>
              <Textarea
                id="activity-description"
                disabled={!isDraft}
                value={form.description}
                onChange={(event) => updateForm("description", event.target.value)}
              />
            </div>
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-1">
              <div className="space-y-2">
                <Label htmlFor="activity-category">活动分类</Label>
                <Input
                  id="activity-category"
                  disabled
                  value={form.category}
                  onChange={(event) => updateForm("category", event.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="activity-academic-year">学年</Label>
                <Input
                  id="activity-academic-year"
                  disabled={!isDraft}
                  value={form.academic_year}
                  onChange={(event) =>
                    updateForm("academic_year", event.target.value)
                  }
                />
              </div>
            </div>
            <div className="grid gap-4 md:grid-cols-3 xl:grid-cols-1">
              <div className="space-y-2">
                <Label htmlFor="activity-semester">学期</Label>
                <select
                  id="activity-semester"
                  className={selectClassName}
                  disabled={!isDraft}
                  value={form.semester}
                  onChange={(event) => updateForm("semester", event.target.value)}
                >
                  <option value="">请选择学期</option>
                  {semesterOptions.map((option) => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="activity-start-date">开始日期</Label>
                <Input
                  id="activity-start-date"
                  disabled={!isDraft}
                  type="date"
                  value={form.start_date}
                  onChange={(event) => updateForm("start_date", event.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="activity-end-date">结束日期</Label>
                <Input
                  id="activity-end-date"
                  disabled={!isDraft}
                  type="date"
                  value={form.end_date}
                  onChange={(event) => updateForm("end_date", event.target.value)}
                />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              分类字段
            </CardTitle>
            <CardDescription>
              这里展示所属分类的 schema 字段，保存时会统一提交到活动详情。
            </CardDescription>
          </CardHeader>
          <CardContent>
            {schemaQuery.isLoading ? (
              <div className="rounded-lg border border-dashed p-8 text-sm text-muted-foreground">
                正在加载分类字段...
              </div>
            ) : schemaQuery.isError ? (
              <div className="rounded-lg border border-destructive/40 p-8 text-sm text-destructive">
                分类字段加载失败，请刷新后重试。
              </div>
            ) : schemaQuery.data ? (
              <ActivitySchemaForm
                disabled={!isDraft || updateMutation.isPending}
                initialValues={toSchemaValues(activity.details)}
                schema={schemaQuery.data.schema}
                submitLabel="保存修改"
                onSubmit={handleSave}
              />
            ) : null}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

export default function ActivityEditPage() {
  const { id } = useParams<{ id: string }>();

  const activityQuery = useQuery({
    queryKey: ["activity", id],
    queryFn: () => getActivityDetail(id || ""),
    enabled: Boolean(id),
  });

  if (!id) {
    return (
      <div className="rounded-lg border border-destructive/40 p-8 text-sm text-destructive">
        缺少活动 ID。
      </div>
    );
  }

  if (activityQuery.isLoading) {
    return (
      <div className="rounded-lg border border-dashed p-8 text-sm text-muted-foreground">
        正在加载活动...
      </div>
    );
  }

  if (activityQuery.isError || !activityQuery.data) {
    return (
      <div className="rounded-lg border border-destructive/40 p-8 text-sm text-destructive">
        活动加载失败，请返回列表重试。
      </div>
    );
  }

  return (
    <ActivityEditForm key={activityQuery.data.id} activity={activityQuery.data} />
  );
}
