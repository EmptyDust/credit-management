import { useMutation, useQuery } from "@tanstack/react-query";
import { Save } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import toast from "react-hot-toast";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { getProfile, updateProfile } from "@/modules/users/api/users";

type ProfileFormValues = {
  real_name: string;
  email: string;
  phone: string;
};

export default function ProfilePage() {
  const query = useQuery({
    queryKey: ["me", "profile"],
    queryFn: getProfile,
  });

  const form = useForm<ProfileFormValues>({
    defaultValues: {
      real_name: "",
      email: "",
      phone: "",
    },
  });

  useEffect(() => {
    if (!query.data) {
      return;
    }

    form.reset({
      real_name: query.data.real_name || "",
      email: query.data.email || "",
      phone: query.data.phone || "",
    });
  }, [form, query.data]);

  const mutation = useMutation({
    mutationFn: updateProfile,
    onSuccess: () => {
      toast.success("个人资料已更新");
      void query.refetch();
    },
  });

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>个人资料</CardTitle>
        </CardHeader>
        <CardContent>
          {query.isLoading ? (
            <div className="text-sm text-muted-foreground">正在加载个人资料...</div>
          ) : query.isError ? (
            <div className="text-sm text-destructive">个人资料加载失败。</div>
          ) : (
            <Form {...form}>
              <form
                className="space-y-4"
                onSubmit={form.handleSubmit((values) => mutation.mutate(values))}
              >
                <div className="grid gap-4 md:grid-cols-2">
                  <FormField
                    control={form.control}
                    name="real_name"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>姓名</FormLabel>
                        <FormControl>
                          <Input {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormItem>
                    <FormLabel>用户名</FormLabel>
                    <FormControl>
                      <Input value={query.data?.username || ""} disabled />
                    </FormControl>
                  </FormItem>
                  <FormField
                    control={form.control}
                    name="email"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>邮箱</FormLabel>
                        <FormControl>
                          <Input {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="phone"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>手机号</FormLabel>
                        <FormControl>
                          <Input {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>
                <Button type="submit" className="gap-2" disabled={mutation.isPending}>
                  <Save className="h-4 w-4" />
                  保存资料
                </Button>
              </form>
            </Form>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
