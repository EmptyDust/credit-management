import { useState, useEffect } from "react";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import apiClient from "@/lib/api";
import { getOptions } from "@/lib/options";
import { useNavigate, Link } from "react-router-dom";
import { 
  UserPlus, 
  User, 
  Mail, 
  Phone, 
  FileSignature, 
  GraduationCap,
} from "lucide-react";
import * as z from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import toast from "react-hot-toast";
import { PasswordInput } from "@/components/ui/password-input";

// 学生注册表单验证规则
const studentRegisterSchema = z.object({
  username: z.string()
    .min(3, "用户名至少3个字符")
    .max(20, "用户名最多20个字符")
    .regex(/^[a-zA-Z0-9_]+$/, "用户名只能包含字母、数字和下划线"),
  password: z.string()
    .min(8, "密码至少8个字符")
    .regex(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/, "密码必须包含大小写字母和数字"),
  confirm_password: z.string(),
  email: z.string().email("请输入有效的邮箱地址"),
  phone: z.string()
    .length(11, "手机号必须是11位数字")
    .regex(/^1[3-9]\d{9}$/, "请输入有效的手机号"),
  real_name: z.string().min(2, "真实姓名至少2个字符").max(50, "真实姓名最多50个字符"),
  id: z.string()
    .length(8, "学号必须是8位数字")
    .regex(/^\d{8}$/, "学号必须是8位数字"),
  college: z.string().min(1, "请选择学院").max(100, "学院名称最多100个字符"),
  major: z.string().min(1, "请选择专业").max(100, "专业名称最多100个字符"),
  class: z.string().min(1, "请选择班级").max(50, "班级名称最多50个字符"),
  grade: z.string().length(4, "年级必须是4位数字").regex(/^\d{4}$/, "年级必须是4位数字"),
}).refine((data) => data.password === data.confirm_password, {
  message: "两次密码输入不一致",
  path: ["confirm_password"],
});

type StudentRegisterForm = z.infer<typeof studentRegisterSchema>;

// 学院和专业数据（学院从后端获取）

// 专业、班级、年级将从后端获取

export default function Register() {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const [collegeOptions, setCollegeOptions] = useState<{ value: string; label: string }[]>([]);
  const [majorOptions, setMajorOptions] = useState<Record<string, { value: string; label: string }[]>>({});
  const [classOptions, setClassOptions] = useState<Record<string, { value: string; label: string }[]>>({});
  const [gradeOptions, setGradeOptions] = useState<{ value: string; label: string }[]>([]);

  const form = useForm<StudentRegisterForm>({
    resolver: zodResolver(studentRegisterSchema),
    defaultValues: {
      username: "",
      password: "",
      confirm_password: "",
      email: "",
      phone: "",
      real_name: "",
      id: "",
      college: "",
      major: "",
      class: "",
      grade: "",
    },
  });

  const selectedCollege = form.watch("college");
  const availableMajors = selectedCollege ? (majorOptions[selectedCollege] || []) : [];
  const selectedMajor = form.watch("major");
  const availableClasses = selectedMajor ? (classOptions[selectedMajor] || []) : [];

  useEffect(() => {
    (async () => {
      try {
        const opts = await getOptions();
        setCollegeOptions(opts.colleges);
        setMajorOptions(opts.majors || {});
        setClassOptions(opts.classes || {});
        setGradeOptions(opts.grades || []);
      } catch (e) {
        console.error("Failed to load options", e);
      }
    })();
  }, []);

  const onSubmit = async (values: StudentRegisterForm) => {
    setLoading(true);
    try {
      // 构造符合后端API的请求数据
      const registerData = {
        ...values,
        user_type: "student", // 固定为学生类型
      };

      const response = await apiClient.post("/users/register", registerData);
      
      if (response.data.code === 0) {
        toast.success("注册成功！正在跳转到登录页面...");
        setTimeout(() => {
          navigate("/login");
        }, 2000);
      } else {
        toast.error(response.data.message || "注册失败");
      }
    } catch (err: unknown) {
      console.error("Register error:", err);
      // 错误处理由API拦截器完成
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-green-50 to-emerald-100 dark:from-gray-900 dark:to-gray-800 p-4">
      <Card className="w-full max-w-2xl border-0 shadow-2xl sm:border bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm">
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <CardHeader className="text-center space-y-4">
              <div className="mx-auto w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center">
                <GraduationCap className="h-8 w-8 text-primary" />
              </div>
              <div>
                <CardTitle className="text-2xl font-bold">学生注册</CardTitle>
                <CardDescription className="mt-2">
                  请填写以下信息完成学生账号注册
                </CardDescription>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              {/* 基础信息 */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
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
                  name="real_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>真实姓名</FormLabel>
                      <FormControl>
                        <div className="relative">
                          <FileSignature className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                          <Input
                            {...field}
                            placeholder="请输入真实姓名"
                            className="pl-10"
                            disabled={loading}
                          />
                        </div>
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>密码</FormLabel>
                    <FormControl>
                      <PasswordInput {...field} placeholder="请输入密码" error={form.formState.errors.password?.message} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="confirm_password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>确认密码</FormLabel>
                    <FormControl>
                      <PasswordInput {...field} placeholder="请再次输入密码" error={form.formState.errors.confirm_password?.message} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <FormField
                  control={form.control}
                  name="email"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>邮箱</FormLabel>
                      <FormControl>
                        <div className="relative">
                          <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                          <Input
                            {...field}
                            type="email"
                            placeholder="请输入邮箱地址"
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
                  name="phone"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>手机号码</FormLabel>
                      <FormControl>
                        <div className="relative">
                          <Phone className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                          <Input
                            {...field}
                            placeholder="请输入手机号码"
                            className="pl-10"
                            disabled={loading}
                          />
                        </div>
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              {/* 学号 */}
              <FormField
                control={form.control}
                name="id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>学号</FormLabel>
                    <FormControl>
                      <div className="relative">
                        <GraduationCap className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                        <Input
                          {...field}
                          placeholder="请输入8位学号"
                          className="pl-10"
                          disabled={loading}
                        />
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              {/* 学院和专业 */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <FormField
                  control={form.control}
                  name="college"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>学院</FormLabel>
                      <Select onValueChange={field.onChange} defaultValue={field.value}>
                        <FormControl>
                          <SelectTrigger disabled={loading}>
                            <SelectValue placeholder="请选择学院" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {collegeOptions.map((college) => (
                            <SelectItem key={college.value} value={college.value}>
                              {college.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="major"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>专业</FormLabel>
                      <Select 
                        onValueChange={field.onChange} 
                        defaultValue={field.value}
                        disabled={!selectedCollege || loading}
                      >
                        <FormControl>
                          <SelectTrigger>
                            <SelectValue placeholder={selectedCollege ? "请选择专业" : "请先选择学院"} />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {availableMajors.map((major) => (
                            <SelectItem key={major.value} value={major.value}>
                              {major.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>

              {/* 班级和年级 */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <FormField
                  control={form.control}
                  name="class"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>班级</FormLabel>
                      <FormControl>
                        <Select 
                          onValueChange={field.onChange} 
                          defaultValue={field.value} 
                          disabled={loading || !selectedMajor}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder={selectedMajor ? "请选择班级" : "请先选择专业"} />
                          </SelectTrigger>
                          <SelectContent>
                            {availableClasses.map((c) => (
                              <SelectItem key={c.value} value={c.value}>{c.label}</SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="grade"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>年级</FormLabel>
                      <FormControl>
                        <Select onValueChange={field.onChange} defaultValue={field.value} disabled={loading}>
                          <SelectTrigger>
                            <SelectValue placeholder="请选择年级" />
                          </SelectTrigger>
                          <SelectContent>
                            {gradeOptions.map((g) => (
                              <SelectItem key={g.value} value={g.value}>{g.label}</SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-4">
              <Button
                type="submit"
                className="w-full h-11 transition-all duration-200 hover:scale-[1.02]"
                disabled={loading}
              >
                {loading ? (
                  <div className="flex items-center gap-2">
                    <div className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                    注册中...
                  </div>
                ) : (
                  <>
                    <UserPlus className="mr-2 h-4 w-4" />
                    注册学生账号
                  </>
                )}
              </Button>
              <Link
                to="/login"
                className="text-sm text-primary hover:underline transition-colors"
              >
                已有账号？立即登录
              </Link>
            </CardFooter>
          </form>
        </Form>
      </Card>
    </div>
  );
} 