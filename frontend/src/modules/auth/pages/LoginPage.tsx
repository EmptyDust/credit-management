import { useState } from "react";
import axios from "axios";
import { zodResolver } from "@hookform/resolvers/zod";
import { LogIn, User } from "lucide-react";
import { useForm } from "react-hook-form";
import { Link, useNavigate } from "react-router-dom";
import * as z from "zod";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { PasswordInput } from "@/components/ui/password-input";
import { useAuth } from "@/contexts/AuthContext";
import apiClient from "@/lib/api";
import type { ApiSuccessResponse, AuthPayload } from "@/shared/types/api";

const loginSchema = z.object({
  username: z.string().min(1, "用户名不能为空"),
  password: z.string().min(1, "密码不能为空"),
});

type LoginForm = z.infer<typeof loginSchema>;

interface AuthErrorBody {
  message?: string;
  error?: string;
  errors?: Array<{ message?: string; field?: string }>;
}

export default function LoginPage() {
  const [loading, setLoading] = useState(false);
  const [loginError, setLoginError] = useState("");
  const { login } = useAuth();
  const navigate = useNavigate();

  const form = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  const onSubmit = async (values: LoginForm) => {
    setLoading(true);
    setLoginError("");
    try {
      const response = await apiClient.post<ApiSuccessResponse<AuthPayload>>(
        "/auth/login",
        {
          login_id: values.username,
          password: values.password,
        }
      );

      const payload = response.data?.data;
      if (!payload?.access_token || !payload.user) {
        setLoginError("登录响应格式错误");
        return;
      }

      login(payload.access_token, payload.refresh_token || "", {
        ...payload.user,
        fullName: payload.user.real_name,
        avatar: payload.user.avatar_url,
      });
      navigate("/dashboard");
    } catch (error: unknown) {
      if (axios.isAxiosError<AuthErrorBody>(error) && error.response) {
        const { status, data } = error.response;
        switch (status) {
          case 401:
            setLoginError("用户名或密码错误");
            break;
          case 403:
            setLoginError("账户未激活或已被禁用");
            break;
          case 422:
            if (data.errors && Array.isArray(data.errors)) {
              const errorMessages = data.errors
                .map((item) => item.message || item.field)
                .join(", ");
              setLoginError(`数据验证失败: ${errorMessages}`);
            } else {
              setLoginError(data.message || data.error || "数据验证失败");
            }
            break;
          case 500:
            setLoginError("服务器内部错误，请稍后再试");
            break;
          default:
            setLoginError(data?.message || data?.error || "登录失败，请重试");
        }
      } else if (axios.isAxiosError(error) && error.request) {
        setLoginError("网络连接失败，请检查网络设置");
      } else {
        setLoginError("登录失败，请重试");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
      <Card className="w-full max-w-md border-0 bg-white/80 shadow-2xl backdrop-blur-sm sm:border dark:bg-gray-900/80">
        <CardHeader className="space-y-4 text-center">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
            <LogIn className="h-8 w-8 text-primary" />
          </div>
          <div>
            <CardTitle className="text-2xl font-bold">欢迎回来</CardTitle>
            <CardDescription className="mt-2">
              请输入您的账号信息登录系统
            </CardDescription>
          </div>
        </CardHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <CardContent className="space-y-4">
              <FormField
                control={form.control}
                name="username"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>用户名</FormLabel>
                    <FormControl>
                      <div className="relative">
                        <User className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                        <Input
                          aria-label="用户名"
                          {...field}
                          className="pl-10"
                          disabled={loading}
                          placeholder="请输入用户名"
                        />
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>密码</FormLabel>
                    <FormControl>
                      <PasswordInput
                        aria-label="密码"
                        {...field}
                        disabled={loading}
                        error={form.formState.errors.password?.message}
                        placeholder="请输入密码"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {loginError ? (
                <div className="rounded-md border border-red-200 bg-red-50 p-3 text-center text-sm text-red-500 dark:border-red-800 dark:bg-red-900/20">
                  {loginError}
                </div>
              ) : null}
              <div className="flex items-center justify-between text-sm">
                <Link className="text-primary transition-colors hover:underline" to="/register">
                  创建新账号
                </Link>
              </div>
            </CardContent>
            <CardFooter>
              <Button
                className="h-11 w-full transition-all duration-200 hover:scale-[1.02]"
                disabled={loading}
                type="submit"
              >
                {loading ? (
                  <div className="flex items-center gap-2">
                    <div className="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                    登录中...
                  </div>
                ) : (
                  <>
                    <LogIn className="mr-2 h-4 w-4" />
                    登录
                  </>
                )}
              </Button>
            </CardFooter>
          </form>
        </Form>
      </Card>
    </div>
  );
}
