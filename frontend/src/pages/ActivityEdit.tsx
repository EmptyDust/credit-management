import { useState, useEffect } from "react";
import * as z from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useParams, useNavigate } from "react-router-dom";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import { ArrowLeft, Save, Loader2 } from "lucide-react";
import apiClient from "@/lib/api";
import toast from "react-hot-toast";
import { DialogFooter } from "@/components/ui/dialog";

const activitySchema = z.object({
  title: z
    .string()
    .min(1, "活动名称不能为空")
    .max(200, "活动名称不能超过200个字符"),
  description: z.string().optional(),
  category: z.string().min(1, "请选择活动类型"),
  start_date: z.string().optional(),
  end_date: z.string().optional(),
});

type ActivityForm = z.infer<typeof activitySchema>;

interface Activity {
  id: string;
  title: string;
  description: string;
  category: string;
  status: string;
  requirements?: string;
  start_date?: string;
  end_date?: string;
  created_at: string;
  updated_at: string;
}

export default function ActivityEditPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [activity, setActivity] = useState<Activity | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  const form = useForm<ActivityForm>({
    resolver: zodResolver(activitySchema),
    defaultValues: {
      title: "",
      description: "",
      category: "",
      start_date: "",
      end_date: "",
    },
  });

  useEffect(() => {
    if (!id) return;

    setLoading(true);
    apiClient
      .get(`/activities/${id}`)
      .then((response) => {
        const activityData = response.data.data || response.data;
        setActivity(activityData);

        // 格式化日期
        const startDate = activityData.start_date
          ? new Date(activityData.start_date).toISOString().split("T")[0]
          : "";
        const endDate = activityData.end_date
          ? new Date(activityData.end_date).toISOString().split("T")[0]
          : "";

        form.reset({
          title: activityData.title,
          description: activityData.description || "",
          category: activityData.category || "",
          start_date: startDate,
          end_date: endDate,
        });
      })
      .catch((error) => {
        console.error("Failed to fetch activity:", error);
        toast.error("获取活动信息失败");
        navigate("/affairs");
      })
      .finally(() => setLoading(false));
  }, [id, form, navigate]);

  const onSubmit = async (values: ActivityForm) => {
    if (!activity) return;

    setSaving(true);
    try {
      await apiClient.put(`/activities/${activity.id}`, values);
      toast.success("活动更新成功");
      navigate(`/affairs/${activity.id}`);
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || "更新活动失败";
      toast.error(errorMessage);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="flex items-center gap-2">
          <Loader2 className="h-8 w-8 animate-spin" />
          <span className="text-lg">加载中...</span>
        </div>
      </div>
    );
  }

  if (!activity) {
    return (
      <div className="flex flex-col items-center mt-16">
        <h2 className="text-xl font-semibold text-red-500 mb-2">
          未找到该活动
        </h2>
        <Button onClick={() => navigate("/affairs")}>返回活动列表</Button>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto p-4 md:p-8 space-y-8">
      {/* 返回按钮 */}
      <Button
        variant="ghost"
        onClick={() => navigate(`/affairs/${activity.id}`)}
        className="mb-2"
      >
        <ArrowLeft className="h-4 w-4 mr-2" /> 返回活动详情
      </Button>

      {/* 编辑表单 */}
      <Card className="rounded-xl shadow-lg">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Save className="h-5 w-5" />
            编辑活动
          </CardTitle>
        </CardHeader>
        <CardContent>
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
                        <SelectItem value="创新创业">创新创业</SelectItem>
                        <SelectItem value="学科竞赛">学科竞赛</SelectItem>
                        <SelectItem value="志愿服务">志愿服务</SelectItem>
                        <SelectItem value="学术研究">学术研究</SelectItem>
                        <SelectItem value="文体活动">文体活动</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>活动描述</FormLabel>
                    <FormControl>
                      <Textarea placeholder="请输入活动描述" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <FormField
                  control={form.control}
                  name="start_date"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>开始时间</FormLabel>
                      <FormControl>
                        <Input type="date" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="end_date"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>结束时间</FormLabel>
                      <FormControl>
                        <Input type="date" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
              <DialogFooter>
                <Button type="submit" className="w-full">
                  保存修改
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}
