import { useMutation, useQuery } from "@tanstack/react-query";
import { ArrowLeft, FileText } from "lucide-react";
import { useState } from "react";
import toast from "react-hot-toast";
import { useNavigate } from "react-router-dom";

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
import {
  getActivityCategorySchema,
  listActivityCategories,
} from "@/modules/activities/api/activity-categories";
import { createActivity } from "@/modules/activities/api/activities";
import ActivitySchemaForm from "@/modules/activities/components/ActivitySchemaForm";
import type { ActivitySchemaNormalizedValues } from "@/modules/activities/schema/types";

type ActivityFixedFormState = {
  title: string;
  description: string;
  category: string;
  academic_year: string;
  semester: string;
  start_date: string;
  end_date: string;
};

const emptyFixedForm: ActivityFixedFormState = {
  title: "",
  description: "",
  category: "",
  academic_year: "",
  semester: "",
  start_date: "",
  end_date: "",
};

const semesterOptions = [
  { value: "autumn", label: "秋季学期" },
  { value: "spring", label: "春季学期" },
  { value: "summer", label: "夏季学期" },
];

const selectClassName =
  "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2";

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

export default function ActivityCreatePage() {
  const navigate = useNavigate();
  const [form, setForm] = useState<ActivityFixedFormState>(emptyFixedForm);

  const categoriesQuery = useQuery({
    queryKey: ["activity-categories"],
    queryFn: listActivityCategories,
  });

  const schemaQuery = useQuery({
    queryKey: ["activity-category-schema", form.category],
    queryFn: () => getActivityCategorySchema(form.category),
    enabled: form.category.length > 0,
  });

  const createMutation = useMutation({
    mutationFn: (details: ActivitySchemaNormalizedValues) =>
      createActivity(toMutationInput(form, details)),
    onSuccess: (activity) => {
      toast.success("活动草稿已保存");
      navigate(`/activities/${activity.id}/edit`);
    },
    onError: (error) => {
      const message =
        error instanceof Error ? error.message : "保存活动草稿失败，请稍后重试";
      toast.error(message);
    },
  });

  const updateForm = (key: keyof ActivityFixedFormState, value: string) => {
    setForm((current) => ({ ...current, [key]: value }));
  };

  const handleSubmit = async (details: ActivitySchemaNormalizedValues) => {
    if (
      !form.title.trim() ||
      !form.category ||
      !form.academic_year.trim() ||
      !form.semester
    ) {
      toast.error("请先填写标题、分类、学年和学期");
      return;
    }

    await createMutation.mutateAsync(details);
  };

  return (
    <div className="space-y-6">
      <PageHeader
        title="新建活动"
        description="先选择活动分类并填写基础字段，系统会按分类加载对应的动态字段。"
        actions={
          <Button
            className="gap-2"
            variant="outline"
            onClick={() => navigate("/activities")}
          >
            <ArrowLeft className="h-4 w-4" />
            返回列表
          </Button>
        }
      />

      <div className="grid gap-6 xl:grid-cols-[minmax(0,420px)_1fr]">
        <Card>
          <CardHeader>
            <CardTitle>基础信息</CardTitle>
            <CardDescription>
              这些字段属于活动主表，保存后用于列表、搜索、统计和申请派生。
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4">
            <div className="space-y-2">
              <Label htmlFor="activity-title">活动标题</Label>
              <Input
                id="activity-title"
                value={form.title}
                onChange={(event) => updateForm("title", event.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="activity-description">活动描述</Label>
              <Textarea
                id="activity-description"
                value={form.description}
                onChange={(event) => updateForm("description", event.target.value)}
                placeholder="说明活动目标、范围和面向对象"
              />
            </div>
            <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-1">
              <div className="space-y-2">
                <Label htmlFor="activity-category">活动分类</Label>
                <select
                  id="activity-category"
                  className={selectClassName}
                  value={form.category}
                  onChange={(event) => updateForm("category", event.target.value)}
                >
                  <option value="">请选择分类</option>
                  {(categoriesQuery.data ?? []).map((category) => (
                    <option key={category.key} value={category.key}>
                      {category.label}
                    </option>
                  ))}
                </select>
                {categoriesQuery.isError ? (
                  <p className="text-sm text-destructive">活动分类加载失败。</p>
                ) : null}
              </div>
              <div className="space-y-2">
                <Label htmlFor="activity-academic-year">学年</Label>
                <Input
                  id="activity-academic-year"
                  placeholder="2025-2026"
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
                  type="date"
                  value={form.start_date}
                  onChange={(event) => updateForm("start_date", event.target.value)}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="activity-end-date">结束日期</Label>
                <Input
                  id="activity-end-date"
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
              动态字段由后端分类 schema 定义，前端只负责展示和基础交互校验。
            </CardDescription>
          </CardHeader>
          <CardContent>
            {!form.category ? (
              <div className="rounded-lg border border-dashed p-8 text-sm text-muted-foreground">
                请先选择活动分类。
              </div>
            ) : schemaQuery.isLoading ? (
              <div className="rounded-lg border border-dashed p-8 text-sm text-muted-foreground">
                正在加载分类字段...
              </div>
            ) : schemaQuery.isError ? (
              <div className="rounded-lg border border-destructive/40 p-8 text-sm text-destructive">
                分类字段加载失败，请刷新后重试。
              </div>
            ) : schemaQuery.data ? (
              <ActivitySchemaForm
                key={form.category}
                disabled={createMutation.isPending}
                schema={schemaQuery.data.schema}
                submitLabel="保存草稿"
                onSubmit={handleSubmit}
              />
            ) : null}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
