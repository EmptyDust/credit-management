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
import { useNavigate, Link } from "react-router-dom";
import { 
  UserPlus, 
  User, 
  KeyRound, 
  Mail, 
  Phone, 
  FileSignature, 
  Eye, 
  EyeOff,
  GraduationCap,
  Building,
  Users
} from "lucide-react";
import * as z from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import toast from "react-hot-toast";

// 学生注册表单验证规则
const studentRegisterSchema = z.object({
  username: z.string()
    .min(3, "用户名至少3个字符")
    .max(20, "用户名最多20个字符")
    .regex(/^[a-zA-Z0-9_]+$/, "用户名只能包含字母、数字和下划线"),
  password: z.string()
    .min(8, "密码至少8个字符")
    .regex(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/, "密码必须包含大小写字母和数字"),
  email: z.string().email("请输入有效的邮箱地址"),
  phone: z.string()
    .min(11, "手机号必须是11位数字")
    .max(11, "手机号必须是11位数字")
    .regex(/^1[3-9]\d{9}$/, "请输入有效的手机号"),
  real_name: z.string().min(2, "真实姓名至少2个字符").max(50, "真实姓名最多50个字符"),
  student_id: z.string()
    .min(8, "学号必须是8位数字")
    .max(8, "学号必须是8位数字")
    .regex(/^\d{8}$/, "学号必须是8位数字"),
  college: z.string().min(1, "请选择学院"),
  major: z.string().min(1, "请选择专业"),
  class: z.string().min(1, "请选择班级"),
  grade: z.string().length(4, "年级必须是4位数字").regex(/^\d{4}$/, "年级必须是4位数字"),
});

type StudentRegisterForm = z.infer<typeof studentRegisterSchema>;

// 学院和专业数据
const colleges = [
  { value: "计算机学院", label: "计算机学院" },
  { value: "信息工程学院", label: "信息工程学院" },
  { value: "数学学院", label: "数学学院" },
  { value: "物理学院", label: "物理学院" },
  { value: "化学学院", label: "化学学院" },
  { value: "生命科学学院", label: "生命科学学院" },
  { value: "经济管理学院", label: "经济管理学院" },
  { value: "外国语学院", label: "外国语学院" },
  { value: "文学院", label: "文学院" },
  { value: "法学院", label: "法学院" },
];

const majors = {
  "计算机学院": [
    { value: "软件工程", label: "软件工程" },
    { value: "计算机科学与技术", label: "计算机科学与技术" },
    { value: "人工智能", label: "人工智能" },
    { value: "数据科学与大数据技术", label: "数据科学与大数据技术" },
  ],
  "信息工程学院": [
    { value: "通信工程", label: "通信工程" },
    { value: "电子信息工程", label: "电子信息工程" },
    { value: "自动化", label: "自动化" },
    { value: "物联网工程", label: "物联网工程" },
  ],
  "数学学院": [
    { value: "数学与应用数学", label: "数学与应用数学" },
    { value: "信息与计算科学", label: "信息与计算科学" },
    { value: "统计学", label: "统计学" },
  ],
  "物理学院": [
    { value: "物理学", label: "物理学" },
    { value: "应用物理学", label: "应用物理学" },
  ],
  "化学学院": [
    { value: "化学", label: "化学" },
    { value: "应用化学", label: "应用化学" },
  ],
  "生命科学学院": [
    { value: "生物科学", label: "生物科学" },
    { value: "生物技术", label: "生物技术" },
  ],
  "经济管理学院": [
    { value: "工商管理", label: "工商管理" },
    { value: "会计学", label: "会计学" },
    { value: "金融学", label: "金融学" },
  ],
  "外国语学院": [
    { value: "英语", label: "英语" },
    { value: "日语", label: "日语" },
    { value: "德语", label: "德语" },
  ],
  "文学院": [
    { value: "汉语言文学", label: "汉语言文学" },
    { value: "新闻学", label: "新闻学" },
  ],
  "法学院": [
    { value: "法学", label: "法学" },
    { value: "知识产权", label: "知识产权" },
  ],
};

export default function Register() {
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const form = useForm<StudentRegisterForm>({
    resolver: zodResolver(studentRegisterSchema),
    defaultValues: {
      username: "",
      password: "",
      email: "",
      phone: "",
      real_name: "",
      student_id: "",
      college: "",
      major: "",
      class: "",
      grade: "",
    },
  });

  const selectedCollege = form.watch("college");
  const availableMajors = selectedCollege ? majors[selectedCollege as keyof typeof majors] || [] : [];

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
                      <div className="relative">
                        <KeyRound className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                        <Input
                          {...field}
                          type={showPassword ? "text" : "password"}
                          placeholder="请输入密码"
                          className="pl-10 pr-10"
                          disabled={loading}
                        />
                        <button
                          type="button"
                          onClick={() => setShowPassword(!showPassword)}
                          className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                          disabled={loading}
                        >
                          {showPassword ? (
                            <EyeOff className="h-4 w-4" />
                          ) : (
                            <Eye className="h-4 w-4" />
                          )}
                        </button>
                      </div>
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
                name="student_id"
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
                          {colleges.map((college) => (
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
                        <div className="relative">
                          <Users className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                          <Input
                            {...field}
                            placeholder="请输入班级"
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
                  name="grade"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>年级</FormLabel>
                      <FormControl>
                        <div className="relative">
                          <Building className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                          <Input
                            {...field}
                            placeholder="请输入年级（如：2023）"
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