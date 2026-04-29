import React, { useEffect, useState } from "react";
import * as z from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
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
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectTrigger,
  SelectValue,
  SelectContent,
  SelectItem,
} from "@/components/ui/select";
import toast from "react-hot-toast";
import apiClient from "@/lib/api";
import type { Activity } from "@/types/activity";
import { getActivityOptions } from "@/lib/options";

const activitySchema = z.object({
  title: z
    .string()
    .min(1, "活动名称不能为空")
    .max(200, "活动名称不能超过200个字符"),
  category: z.string().min(1, "请选择活动类型"),
  description: z
    .string()
    .max(1000, "活动简介不能超过1000个字符")
    .optional()
    .or(z.literal("")),
  start_date: z
    .string()
    .optional()
    .or(z.literal("")),
  end_date: z
    .string()
    .optional()
    .or(z.literal("")),
});

type ActivityFormValues = z.infer<typeof activitySchema>;

interface ActivityEditDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  activity?: Activity | null;
  onSuccess?: () => void;
}

export const ActivityEditDialog: React.FC<ActivityEditDialogProps> = ({
  open,
  onOpenChange,
  activity,
  onSuccess,
}) => {
  const [loading, setLoading] = useState(false);
  const [activityCategories, setActivityCategories] = useState<
    { value: string; label: string }[]
  >([]);
  const [categoryFields, setCategoryFields] = useState<
    Record<
      string,
      Array<{
        name: string;
        label: string;
        type: string;
        required?: boolean;
        options?: { value: string; label: string }[];
        min?: number;
        max?: number;
        maxLength?: number;
      }>
    >
  >({});
  const [details, setDetails] = useState<Record<string, any>>({});

  const form = useForm<ActivityFormValues>({
    resolver: zodResolver(activitySchema),
    defaultValues: {
      title: "",
      category: "",
      description: "",
      start_date: "",
      end_date: "",
    },
  });

  // 加载活动配置（类别、动态字段）
  useEffect(() => {
    (async () => {
      try {
        const opts = await getActivityOptions();
        setActivityCategories(opts.categories || []);
        setCategoryFields(opts.category_fields || {});
      } catch (e) {
        console.error("Failed to load activity options in ActivityEditDialog", e);
      }
    })();
  }, []);

  // 根据当前 activity 初始化表单和 details
  useEffect(() => {
    if (activity) {
      form.reset({
        title: activity.title,
        category: activity.category || "",
        description: activity.description || "",
        start_date: activity.start_date
          ? activity.start_date.split("T")[0]
          : "",
        end_date: activity.end_date ? activity.end_date.split("T")[0] : "",
      });
      // @ts-ignore
      setDetails((activity as any).details || {});
    } else {
      form.reset({
        title: "",
        category: "",
        description: "",
        start_date: "",
        end_date: "",
      });
      setDetails({});
    }
  }, [activity, form]);

  const onSubmit = async (values: ActivityFormValues) => {
    try {
      setLoading(true);

      const payload: any = {
        ...values,
        description: values.description || "",
        start_date: values.start_date || "",
        end_date: values.end_date || "",
        details,
      };

      if (activity) {
        await apiClient.put(`/activities/${activity.id}`, payload);
        toast.success("活动更新成功");
      } else {
        await apiClient.post("/activities", payload);
        toast.success("活动创建成功");
      }

      onOpenChange(false);
      if (onSuccess) {
        onSuccess();
      }
    } catch (error) {
      console.error("Failed to save activity:", error);
      toast.error(activity ? "更新活动失败" : "创建活动失败");
    } finally {
      setLoading(false);
    }
  };

  const selectedCategory = form.watch("category");
  const dynamicFields = categoryFields[selectedCategory] || [];

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>{activity ? "编辑活动" : "添加新活动"}</DialogTitle>
          <DialogDescription>
            {activity ? "修改活动信息" : "创建新的活动"}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(onSubmit)}
            className="space-y-6"
            autoComplete="off"
          >
            {/* 活动名称 + 活动类型 并排 */}
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
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
                        {activityCategories.map((c) => (
                          <SelectItem key={c.value} value={c.value}>
                            {c.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>活动简介</FormLabel>
                  <FormControl>
                    <Textarea
                      placeholder="请输入活动简介（可选）"
                      rows={3}
                      {...field}
                    />
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
                    <FormLabel>开始日期</FormLabel>
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
                    <FormLabel>结束日期</FormLabel>
                    <FormControl>
                      <Input type="date" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            {/* 动态详情字段 */}
            {dynamicFields.length > 0 && (
              <div className="space-y-4">
                {dynamicFields.map((f) => (
                  <div key={f.name} className="grid gap-2">
                    <FormLabel>
                      {f.label}
                      {f.required ? " *" : ""}
                    </FormLabel>
                    {f.type === "select" ? (
                      <Select
                        onValueChange={(v: string) =>
                          setDetails((d) => ({ ...d, [f.name]: v }))
                        }
                        defaultValue={details?.[f.name] ?? ""}
                      >
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue
                              placeholder={`请选择${f.label}`}
                            />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {(f.options || []).map((opt) => (
                            <SelectItem key={opt.value} value={opt.value}>
                              {opt.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    ) : (
                      <Input
                        type={
                          f.type === "number"
                            ? "number"
                            : f.type === "date"
                            ? "date"
                            : "text"
                        }
                        defaultValue={details?.[f.name] ?? ""}
                        onChange={(e) =>
                          setDetails((d) => ({
                            ...d,
                            [f.name]:
                              f.type === "number"
                                ? Number(e.target.value)
                                : e.target.value,
                          }))
                        }
                      />
                    )}
                  </div>
                ))}
              </div>
            )}

            <DialogFooter>
              <Button
                type="submit"
                className="w-full"
                disabled={loading}
              >
                {loading
                  ? activity
                    ? "更新中..."
                    : "创建中..."
                  : activity
                  ? "更新活动"
                  : "创建活动"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
};


