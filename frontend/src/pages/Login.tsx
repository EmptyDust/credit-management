import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useAuth } from "@/contexts/AuthContext";
import apiClient from "@/lib/api";
import { useNavigate, Link } from "react-router-dom";
import { LogIn, User } from "lucide-react";
import * as z from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { PasswordInput } from "@/components/ui/password-input";

const loginSchema = z.object({
  username: z.string().min(1, "用户名不能为空"),
  password: z.string().min(1, "密码不能为空"),
});

type LoginForm = z.infer<typeof loginSchema>;

export default function Login() {
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
      const response = await apiClient.post("/auth/login", values);

      // 检查响应格式
      if (response.data && response.data.code === 0 && response.data.data) {
        const { token, refresh_token, user } = response.data.data;

        if (token && user) {
          // 转换用户数据格式以匹配前端接口（以数据库模型为准）
          const normalizedUser = {
            uuid: user.uuid,
            user_id: user.user_id,
            username: user.username,
            userType: user.user_type,
            email: user.email,
            fullName: user.real_name,
            department: user.department,
            college: user.college,
            major: user.major,
            class: user.class,
            status: user.status,
            createdAt: user.created_at,
            updatedAt: user.updated_at,
          } as const;

          login(token, refresh_token || "", normalizedUser);
          navigate("/dashboard");
        } else {
          setLoginError("登录响应格式错误");
        }
      } else {
        setLoginError(response.data?.message || "登录失败");
      }
    } catch (err: any) {
      console.error("Login error:", err);

      // 处理不同类型的错误
      if (err.response) {
        const { status, data } = err.response;

        switch (status) {
          case 401:
            setLoginError("用户名或密码错误");
            break;
          case 403:
            setLoginError("账户未激活或已被禁用");
            break;
          case 422:
            // 验证错误
            if (data.errors && Array.isArray(data.errors)) {
              const errorMessages = data.errors
                .map((err: any) => err.message || err.field)
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
      } else if (err.request) {
        // 网络错误
        setLoginError("网络连接失败，请检查网络设置");
      } else {
        // 其他错误
        setLoginError("登录失败，请重试");
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
      <Card className="w-full max-w-md border-0 shadow-2xl sm:border bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm">
        <CardHeader className="text-center space-y-4">
          <div className="mx-auto w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center">
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
                        <User className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                        <Input
                          {...field}
                          placeholder="请输入用户名"
                          className="pl-10"
                          disabled={loading}
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
                      <PasswordInput {...field} placeholder="请输入密码" disabled={loading} error={form.formState.errors.password?.message} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {loginError && (
                <div className="text-red-500 text-sm text-center bg-red-50 dark:bg-red-900/20 p-3 rounded-md border border-red-200 dark:border-red-800">
                  {loginError}
                </div>
              )}
              <div className="flex items-center justify-between text-sm">
                <Link
                  to="/register"
                  className="text-primary hover:underline transition-colors"
                >
                  创建新账号
                </Link>
              </div>
            </CardContent>
            <CardFooter>
              <Button
                type="submit"
                className="w-full h-11 transition-all duration-200 hover:scale-[1.02]"
                disabled={loading}
              >
                {loading ? (
                  <div className="flex items-center gap-2">
                    <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
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
