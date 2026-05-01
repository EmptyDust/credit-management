import { useEffect, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  FileSignature,
  GraduationCap,
  Mail,
  Phone,
  User,
  UserPlus,
} from "lucide-react";
import { useForm, useWatch } from "react-hook-form";
import toast from "react-hot-toast";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import apiClient from "@/lib/api";
import { getOptions } from "@/lib/options";
import type { ApiSuccessResponse } from "@/shared/types/api";

const studentRegisterSchema = z
  .object({
    username: z
      .string()
      .min(3, "用户名至少3个字符")
      .max(20, "用户名最多20个字符")
      .regex(/^[a-zA-Z0-9_]+$/, "用户名只能包含字母、数字和下划线"),
    password: z
      .string()
      .min(8, "密码至少8个字符")
      .regex(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/, "密码必须包含大小写字母和数字"),
    confirm_password: z.string(),
    email: z.string().email("请输入有效的邮箱地址"),
    phone: z
      .string()
      .length(11, "手机号必须是11位数字")
      .regex(/^1[3-9]\d{9}$/, "请输入有效的手机号"),
    real_name: z.string().min(2, "真实姓名至少2个字符").max(50, "真实姓名最多50个字符"),
    student_id: z
      .string()
      .length(8, "学号必须是8位数字")
      .regex(/^\d{8}$/, "学号必须是8位数字"),
    college: z.string().min(1, "请选择学部").max(100, "学部名称最多100个字符"),
    major: z.string().min(1, "请选择专业").max(100, "专业名称最多100个字符"),
    class: z.string().min(1, "请选择班级").max(50, "班级名称最多50个字符"),
    grade: z
      .string()
      .length(4, "年级必须是4位数字")
      .regex(/^\d{4}$/, "年级必须是4位数字"),
  })
  .refine((data) => data.password === data.confirm_password, {
    message: "两次密码输入不一致",
    path: ["confirm_password"],
  });

type StudentRegisterForm = z.infer<typeof studentRegisterSchema>;

type SelectOption = { value: string; label: string };

export default function RegisterPage() {
  const [loading, setLoading] = useState(false);
  const [collegeOptions, setCollegeOptions] = useState<SelectOption[]>([]);
  const [majorOptions, setMajorOptions] = useState<Record<string, SelectOption[]>>({});
  const [classOptions, setClassOptions] = useState<Record<string, SelectOption[]>>({});
  const [gradeOptions, setGradeOptions] = useState<SelectOption[]>([]);
  const navigate = useNavigate();

  const form = useForm<StudentRegisterForm>({
    resolver: zodResolver(studentRegisterSchema),
    defaultValues: {
      username: "",
      password: "",
      confirm_password: "",
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

  const selectedCollege = useWatch({
    control: form.control,
    name: "college",
  }) || "";
  const selectedMajor = useWatch({
    control: form.control,
    name: "major",
  }) || "";
  const availableMajors = selectedCollege ? majorOptions[selectedCollege] || [] : [];
  const availableClasses = selectedMajor ? classOptions[selectedMajor] || [] : [];

  useEffect(() => {
    void (async () => {
      try {
        const options = await getOptions();
        setCollegeOptions(options.colleges);
        setMajorOptions(options.majors || {});
        setClassOptions(options.classes || {});
        setGradeOptions(options.grades || []);
      } catch (error) {
        console.error("Failed to load options", error);
      }
    })();
  }, []);

  const onSubmit = async (values: StudentRegisterForm) => {
    setLoading(true);
    try {
      const response = await apiClient.post<
        ApiSuccessResponse<{ user: { username: string } }>
      >("/auth/register/student", {
        student_id: values.student_id,
        username: values.username,
        password: values.password,
        email: values.email,
        phone: values.phone,
        real_name: values.real_name,
        department_id: values.class,
        grade: values.grade,
      });

      if (response.data.code === "OK") {
        toast.success("注册成功！正在跳转到登录页面...");
        setTimeout(() => {
          navigate("/login");
        }, 2000);
      } else {
        toast.error(response.data.message || "注册失败");
      }
    } catch (error) {
      console.error("Register error:", error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-green-50 to-emerald-100 p-4 dark:from-gray-900 dark:to-gray-800">
      <Card className="w-full max-w-2xl border-0 bg-white/80 shadow-2xl backdrop-blur-sm sm:border dark:bg-gray-900/80">
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)}>
            <CardHeader className="space-y-4 text-center">
              <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-primary/10">
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
              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
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
                  name="real_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>真实姓名</FormLabel>
                      <FormControl>
                        <div className="relative">
                          <FileSignature className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                          <Input
                            aria-label="真实姓名"
                            {...field}
                            className="pl-10"
                            disabled={loading}
                            placeholder="请输入真实姓名"
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
                      <PasswordInput
                        aria-label="密码"
                        {...field}
                        error={form.formState.errors.password?.message}
                        placeholder="请输入密码"
                      />
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
                      <PasswordInput
                        aria-label="确认密码"
                        {...field}
                        error={form.formState.errors.confirm_password?.message}
                        placeholder="请再次输入密码"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <FormField
                  control={form.control}
                  name="email"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>邮箱</FormLabel>
                      <FormControl>
                        <div className="relative">
                          <Mail className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                          <Input
                            aria-label="邮箱"
                            {...field}
                            className="pl-10"
                            disabled={loading}
                            placeholder="请输入邮箱地址"
                            type="email"
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
                          <Phone className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                          <Input
                            aria-label="手机号码"
                            {...field}
                            className="pl-10"
                            disabled={loading}
                            placeholder="请输入手机号码"
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
                name="student_id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>学号</FormLabel>
                    <FormControl>
                      <div className="relative">
                        <GraduationCap className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                        <Input
                          aria-label="学号"
                          {...field}
                          className="pl-10"
                          disabled={loading}
                          placeholder="请输入8位学号"
                        />
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <FormField
                  control={form.control}
                  name="college"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>学部</FormLabel>
                      <Select onValueChange={field.onChange} defaultValue={field.value}>
                        <FormControl>
                          <SelectTrigger aria-label="学部" disabled={loading}>
                            <SelectValue placeholder="请选择学部" />
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
                        defaultValue={field.value}
                        disabled={!selectedCollege || loading}
                        onValueChange={field.onChange}
                      >
                        <FormControl>
                          <SelectTrigger aria-label="专业">
                            <SelectValue
                              placeholder={selectedCollege ? "请选择专业" : "请先选择学部"}
                            />
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

              <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                <FormField
                  control={form.control}
                  name="class"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>班级</FormLabel>
                      <Select
                        defaultValue={field.value}
                        disabled={loading || !selectedMajor}
                        onValueChange={field.onChange}
                      >
                        <FormControl>
                          <SelectTrigger aria-label="班级">
                            <SelectValue
                              placeholder={selectedMajor ? "请选择班级" : "请先选择专业"}
                            />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {availableClasses.map((item) => (
                            <SelectItem key={item.value} value={item.value}>
                              {item.label}
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
                  name="grade"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>年级</FormLabel>
                      <Select
                        defaultValue={field.value}
                        disabled={loading}
                        onValueChange={field.onChange}
                      >
                        <FormControl>
                          <SelectTrigger aria-label="年级">
                            <SelectValue placeholder="请选择年级" />
                          </SelectTrigger>
                        </FormControl>
                        <SelectContent>
                          {gradeOptions.map((grade) => (
                            <SelectItem key={grade.value} value={grade.value}>
                              {grade.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              </div>
            </CardContent>
            <CardFooter className="flex flex-col gap-4">
              <Button
                className="h-11 w-full transition-all duration-200 hover:scale-[1.02]"
                disabled={loading}
                type="submit"
              >
                {loading ? (
                  <div className="flex items-center gap-2">
                    <div className="h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
                    注册中...
                  </div>
                ) : (
                  <>
                    <UserPlus className="mr-2 h-4 w-4" />
                    注册学生账号
                  </>
                )}
              </Button>
              <Link className="text-sm text-primary transition-colors hover:underline" to="/login">
                已有账号？立即登录
              </Link>
            </CardFooter>
          </form>
        </Form>
      </Card>
    </div>
  );
}
