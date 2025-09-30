import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useAuth } from "@/contexts/AuthContext";
import apiClient from "@/lib/api";
import {
  User,
  Mail,
  Phone,
  FileSignature,
  Edit3,
  Save,
  X,
  Shield,
  Calendar,
  MapPin,
  GraduationCap,
  Building,
  Award,
  Clock,
  Lock,
} from "lucide-react";
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
import toast from "react-hot-toast";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { getStatusBadge } from "@/lib/status-utils";
import { PasswordInput } from "@/components/ui/password-input";
import { getOptions } from "@/lib/options";

const colleges: { value: string; label: string }[] = [];
let majors: Record<string, { value: string; label: string }[]> = {};

const profileSchema = z.object({
  username: z.string().optional(),
  email: z.string().email("请输入有效的邮箱地址").optional(),
  phone: z.string().optional(),
  real_name: z.string().optional(),
  student_id: z.string().optional(),
  college: z.string().optional(),
  major: z.string().optional(),
  class: z.string().optional(),
  grade: z.string().optional(),
  department: z.string().optional(),
  title: z.string().optional(),
});

const passwordSchema = z
  .object({
    current_password: z.string().min(1, "请输入当前密码"),
    new_password: z.string().min(6, "新密码至少6位").max(50, "新密码最多50位"),
    confirm_password: z.string().min(1, "请确认新密码"),
  })
  .refine((data) => data.new_password === data.confirm_password, {
    message: "两次输入的密码不一致",
    path: ["confirm_password"],
  });

type ProfileForm = z.infer<typeof profileSchema>;
type PasswordForm = z.infer<typeof passwordSchema>;

type UserProfile = {
  email: string;
  phone: string;
  real_name: string;
  username: string;
  user_type: string;
  status: string;
  created_at: string;
  updated_at: string;
  // Student specific fields
  student_id?: string;
  college?: string;
  major?: string;
  class?: string;
  grade?: string;
  // Teacher specific fields
  department?: string;
  title?: string;
};

export default function ProfilePage() {
  const { user, updateUser } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [passwordDialogOpen, setPasswordDialogOpen] = useState(false);
  const [changingPassword, setChangingPassword] = useState(false);
  const [collegeOptions, setCollegeOptions] = useState<{ value: string; label: string }[]>([]);
  const [majorOptions, setMajorOptions] = useState<Record<string, { value: string; label: string }[]>>({});

  const form = useForm<ProfileForm>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      email: "",
      phone: "",
      real_name: "",
      college: "",
      major: "",
      class: "",
      grade: "",
      department: "",
      title: "",
    },
  });

  const passwordForm = useForm<PasswordForm>({
    resolver: zodResolver(passwordSchema),
    defaultValues: {
      current_password: "",
      new_password: "",
      confirm_password: "",
    },
  });

  useEffect(() => {
    const fetchProfile = async () => {
      if (!user) return;
      try {
        setLoading(true);
        const response = await apiClient.get(`/users/profile`);
        const profileData = response.data.data || response.data;
        setProfile(profileData);
        form.reset({
          username: profileData.username || "",
          email: profileData.email || "",
          phone: profileData.phone || "",
          real_name: profileData.real_name || "",
          college: profileData.college || "",
          major: profileData.major || "",
          class: profileData.class || "",
          grade: profileData.grade || "",
          department: profileData.department || "",
          title: profileData.title || "",
          student_id: profileData.student_id || "",
        });
      } catch (err) {
        setError("获取个人资料失败");
        console.error(err);
        toast.error("获取个人资料失败");
      } finally {
        setLoading(false);
      }
    };
    fetchProfile();
    (async () => {
      try {
        const opts = await getOptions();
        setCollegeOptions(opts.colleges);
        setMajorOptions(opts.majors || {});
      } catch (e) {
        console.error("Failed to load options", e);
      }
    })();
  }, [user, form]);

  const handleSave = async (values: ProfileForm) => {
    if (!profile) return;
    setError("");
    setSaving(true);
    try {
      // Only send fields that are allowed by the backend UserUpdateRequest
      const updateData: any = {};
      
      // Only include fields that have values and are valid for updates
      if (values.email && values.email !== profile.email) {
        updateData.email = values.email;
      }
      if (values.phone && values.phone !== profile.phone) {
        updateData.phone = values.phone;
      }
      if (values.real_name && values.real_name !== profile.real_name) {
        updateData.real_name = values.real_name;
      }
      if (values.department && values.department !== profile.department) {
        updateData.department_id = values.department;
      }
      
      // Handle grade field - only send if it's a valid 4-digit string
      if (values.grade && values.grade.length === 4 && /^\d{4}$/.test(values.grade)) {
        updateData.grade = values.grade;
      }
      
      // Handle title field
      if (values.title && values.title !== profile.title) {
        updateData.title = values.title;
      }
      
      // Only make the request if there are fields to update
      if (Object.keys(updateData).length > 0) {
        await apiClient.put("/users/profile", updateData);
      }
      
      // Update local state with the new values
      const updatedProfile = { ...profile, ...values };
      setProfile(updatedProfile);
      updateUser({
        ...user,
        ...values,
      });
      setIsEditing(false);
      toast.success("个人资料更新成功！");
    } catch (err) {
      setError("更新个人资料失败");
      console.error(err);
      toast.error("更新个人资料失败");
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    if (profile) {
      form.reset({
        username: profile.username || "",
        email: profile.email || "",
        phone: profile.phone || "",
        real_name: profile.real_name || "",
        college: profile.college || "",
        major: profile.major || "",
        class: profile.class || "",
        grade: profile.grade || "",
        department: profile.department || "",
        title: profile.title || "",
        student_id: profile.student_id || "",
      });
    }
    setIsEditing(false);
    setError("");
  };

  const handlePasswordChange = async (values: PasswordForm) => {
    setChangingPassword(true);
    try {
      await apiClient.post("/users/change_password", {
        old_password: values.current_password,
        new_password: values.new_password,
      });

      setPasswordDialogOpen(false);
      passwordForm.reset();
      toast.success("密码修改成功！");
    } catch (err: any) {
      let errorMessage = err.response?.data?.message || "密码修改失败";
      if (errorMessage.includes("管理员权限")) {
        errorMessage = "您没有权限修改密码，请重新登录或联系管理员。";
      }
      toast.error(errorMessage);
    } finally {
      setChangingPassword(false);
    }
  };


  const getUserTypeLabel = (userType: string) => {
    const labels = {
      student: "学生",
      teacher: "教师",
      admin: "管理员",
    };
    return labels[userType as keyof typeof labels] || userType;
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="flex items-center gap-2">
          <Clock className="h-8 w-8 animate-spin" />
          <span>加载中...</span>
        </div>
      </div>
    );
  }

  if (error && !profile) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-center">
          <div className="text-red-500 mb-2">{error}</div>
          <Button onClick={() => window.location.reload()}>重试</Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-8 p-4 md:p-8">
      <div>
        <h1 className="text-3xl font-bold">个人资料</h1>
        <p className="text-muted-foreground">查看和管理您的个人信息</p>
      </div>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(handleSave)} className="space-y-8">
          <div className="grid gap-6 md:grid-cols-2">
            {/* 基本信息 */}
            <Card>
              <CardHeader className="flex flex-row items-center justify-between">
                <div>
                  <CardTitle>基本信息</CardTitle>
                  <CardDescription>您的账户基本信息</CardDescription>
                </div>
                {!isEditing ? (
                  <Button
                    onClick={() => setIsEditing(true)}
                    variant="outline"
                    type="button"
                  >
                    <Edit3 className="mr-2 h-4 w-4" /> 编辑
                  </Button>
                ) : (
                  <div className="flex gap-2">
                    <Button
                      onClick={handleCancel}
                      variant="outline"
                      size="sm"
                      type="button"
                    >
                      <X className="h-4 w-4" />
                    </Button>
                    <Button type="submit" size="sm" disabled={saving}>
                      <Save className="mr-2 h-4 w-4" />
                      {saving ? "保存中..." : "保存"}
                    </Button>
                  </div>
                )}
              </CardHeader>
              <CardContent className="space-y-4">
                <FormField
                  control={form.control}
                  name="username"
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex items-center gap-4">
                        <User className="h-5 w-5 text-muted-foreground" />
                        <div className="flex-1">
                          <FormLabel>用户名</FormLabel>
                          <FormControl>
                            <Input {...field} disabled className="mt-1" />
                          </FormControl>
                          <FormMessage />
                        </div>
                      </div>
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="real_name"
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex items-center gap-4">
                        <FileSignature className="h-5 w-5 text-muted-foreground" />
                        <div className="flex-1">
                          <FormLabel>真实姓名</FormLabel>
                          <FormControl>
                            <Input
                              {...field}
                              disabled={!isEditing}
                              className="mt-1"
                            />
                          </FormControl>
                          <FormMessage />
                        </div>
                      </div>
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="email"
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex items-center gap-4">
                        <Mail className="h-5 w-5 text-muted-foreground" />
                        <div className="flex-1">
                          <FormLabel>邮箱地址</FormLabel>
                          <FormControl>
                            <Input
                              {...field}
                              type="email"
                              disabled={!isEditing}
                              className="mt-1"
                            />
                          </FormControl>
                          <FormMessage />
                        </div>
                      </div>
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name="phone"
                  render={({ field }) => (
                    <FormItem>
                      <div className="flex items-center gap-4">
                        <Phone className="h-5 w-5 text-muted-foreground" />
                        <div className="flex-1">
                          <FormLabel>手机号码</FormLabel>
                          <FormControl>
                            <Input
                              {...field}
                              disabled={!isEditing}
                              className="mt-1"
                            />
                          </FormControl>
                          <FormMessage />
                        </div>
                      </div>
                    </FormItem>
                  )}
                />
              </CardContent>
            </Card>
            {/* 账户信息 */}
            <Card>
              <CardHeader>
                <CardTitle>账户信息</CardTitle>
                <CardDescription>您的账户状态和类型</CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex items-center gap-4">
                  <Shield className="h-5 w-5 text-muted-foreground" />
                  <div className="flex-1">
                    <label className="text-sm font-medium">用户类型</label>
                    <div className="mt-1">
                      <Badge variant="outline">
                        {getUserTypeLabel(profile?.user_type || "")}
                      </Badge>
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <Shield className="h-5 w-5 text-muted-foreground" />
                  <div className="flex-1">
                    <label className="text-sm font-medium">账户状态</label>
                    <div className="mt-1">
                      {getStatusBadge(profile?.status || "")}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <Calendar className="h-5 w-5 text-muted-foreground" />
                  <div className="flex-1">
                    <label className="text-sm font-medium">注册时间</label>
                    <p className="text-sm text-muted-foreground mt-1">
                      {profile?.created_at
                        ? new Date(profile.created_at).toLocaleDateString()
                        : "未知"}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <Clock className="h-5 w-5 text-muted-foreground" />
                  <div className="flex-1">
                    <label className="text-sm font-medium">最后更新</label>
                    <p className="text-sm text-muted-foreground mt-1">
                      {profile?.updated_at
                        ? new Date(profile.updated_at).toLocaleDateString()
                        : "未知"}
                    </p>
                  </div>
                </div>
                <div className="pt-4 border-t">
                  <Dialog
                    open={passwordDialogOpen}
                    onOpenChange={setPasswordDialogOpen}
                  >
                    <DialogTrigger asChild>
                      <Button variant="outline" className="w-full">
                        <Lock className="mr-2 h-4 w-4" />
                        修改密码
                      </Button>
                    </DialogTrigger>
                    <DialogContent>
                      <DialogHeader>
                        <DialogTitle>修改密码</DialogTitle>
                        <DialogDescription>
                          请输入当前密码和新密码来修改您的账户密码
                        </DialogDescription>
                      </DialogHeader>
                      <Form {...passwordForm}>
                        <form
                          onSubmit={passwordForm.handleSubmit(
                            handlePasswordChange
                          )}
                          className="space-y-4"
                        >
                          <FormField
                            control={passwordForm.control}
                            name="current_password"
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>当前密码</FormLabel>
                                <FormControl>
                                  <PasswordInput {...field} placeholder="请输入当前密码" error={passwordForm.formState.errors.current_password?.message} />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <FormField
                            control={passwordForm.control}
                            name="new_password"
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>新密码</FormLabel>
                                <FormControl>
                                  <PasswordInput {...field} placeholder="请输入新密码" error={passwordForm.formState.errors.new_password?.message} />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <FormField
                            control={passwordForm.control}
                            name="confirm_password"
                            render={({ field }) => (
                              <FormItem>
                                <FormLabel>确认新密码</FormLabel>
                                <FormControl>
                                  <PasswordInput {...field} placeholder="请再次输入新密码" error={passwordForm.formState.errors.confirm_password?.message} />
                                </FormControl>
                                <FormMessage />
                              </FormItem>
                            )}
                          />
                          <div className="flex gap-2 pt-4">
                            <Button
                              type="button"
                              variant="outline"
                              onClick={() => setPasswordDialogOpen(false)}
                              className="flex-1"
                            >
                              取消
                            </Button>
                            <Button
                              type="submit"
                              disabled={changingPassword}
                              className="flex-1"
                            >
                              {changingPassword ? "修改中..." : "确认修改"}
                            </Button>
                          </div>
                        </form>
                      </Form>
                    </DialogContent>
                  </Dialog>
                </div>
              </CardContent>
            </Card>
          </div>
          {/* 学生信息表单 */}
          {profile?.user_type === "student" && (
            <Card>
              <CardHeader>
                <CardTitle>学生信息</CardTitle>
                <CardDescription>您的学生档案信息</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 md:grid-cols-2">
                  <FormField
                    control={form.control}
                    name="student_id"
                    render={({ field }) => (
                      <FormItem>
                        <div className="flex items-center gap-4">
                          <GraduationCap className="h-5 w-5 text-muted-foreground" />
                          <div className="flex-1">
                            <FormLabel>学号</FormLabel>
                            <FormControl>
                              <Input
                                {...field}
                                disabled={!isEditing}
                                className="mt-1"
                              />
                            </FormControl>
                            <FormMessage />
                          </div>
                        </div>
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="college"
                    render={({ field }) => (
                      <FormItem>
                        <div className="flex items-center gap-4">
                          <MapPin className="h-5 w-5 text-muted-foreground" />
                          <div className="flex-1">
                            <FormLabel>学院</FormLabel>
                            <FormControl>
                              <Select
                                disabled={!isEditing}
                                value={field.value || ""}
                                onValueChange={field.onChange}
                              >
                                <SelectTrigger>
                                  <SelectValue placeholder="请选择学院" />
                                </SelectTrigger>
                                <SelectContent>
                                  {collegeOptions.map((c) => (
                                    <SelectItem key={c.value} value={c.value}>
                                      {c.label}
                                    </SelectItem>
                                  ))}
                                </SelectContent>
                              </Select>
                            </FormControl>
                            <FormMessage />
                          </div>
                        </div>
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="major"
                    render={({ field }) => {
                      const selectedCollege = form.watch("college");
                      const availableMajors = selectedCollege
                        ? majorOptions[selectedCollege || ""] || []
                        : [];
                      return (
                        <FormItem>
                          <div className="flex items-center gap-4">
                            <GraduationCap className="h-5 w-5 text-muted-foreground" />
                            <div className="flex-1">
                              <FormLabel>专业</FormLabel>
                              <FormControl>
                                <Select
                                  disabled={!isEditing}
                                  value={field.value || ""}
                                  onValueChange={field.onChange}
                                >
                                  <SelectTrigger>
                                    <SelectValue placeholder="请选择专业" />
                                  </SelectTrigger>
                                  <SelectContent>
                                    {availableMajors.map(
                                      (m: { value: string; label: string }) => (
                                        <SelectItem
                                          key={m.value}
                                          value={m.value}
                                        >
                                          {m.label}
                                        </SelectItem>
                                      )
                                    )}
                                  </SelectContent>
                                </Select>
                              </FormControl>
                              <FormMessage />
                            </div>
                          </div>
                        </FormItem>
                      );
                    }}
                  />
                  <FormField
                    control={form.control}
                    name="class"
                    render={({ field }) => (
                      <FormItem>
                        <div className="flex items-center gap-4">
                          <Building className="h-5 w-5 text-muted-foreground" />
                          <div className="flex-1">
                            <FormLabel>班级</FormLabel>
                            <FormControl>
                              <Input
                                {...field}
                                disabled={!isEditing}
                                className="mt-1"
                              />
                            </FormControl>
                            <FormMessage />
                          </div>
                        </div>
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="grade"
                    render={({ field }) => (
                      <FormItem>
                        <div className="flex items-center gap-4">
                          <Award className="h-5 w-5 text-muted-foreground" />
                          <div className="flex-1">
                            <FormLabel>年级</FormLabel>
                            <FormControl>
                              <Input
                                {...field}
                                disabled={!isEditing}
                                className="mt-1"
                              />
                            </FormControl>
                            <FormMessage />
                          </div>
                        </div>
                      </FormItem>
                    )}
                  />
                </div>
              </CardContent>
            </Card>
          )}
          {/* 教师信息表单 */}
          {profile?.user_type === "teacher" && (
            <Card>
              <CardHeader>
                <CardTitle>教师信息</CardTitle>
                <CardDescription>您的教师档案信息</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4 md:grid-cols-2">
                  <FormField
                    control={form.control}
                    name="department"
                    render={({ field }) => (
                      <FormItem>
                        <div className="flex items-center gap-4">
                          <Building className="h-5 w-5 text-muted-foreground" />
                          <div className="flex-1">
                            <FormLabel>院系</FormLabel>
                            <FormControl>
                              <Input
                                {...field}
                                disabled={!isEditing}
                                className="mt-1"
                              />
                            </FormControl>
                            <FormMessage />
                          </div>
                        </div>
                      </FormItem>
                    )}
                  />
                  <FormField
                    control={form.control}
                    name="title"
                    render={({ field }) => (
                      <FormItem>
                        <div className="flex items-center gap-4">
                          <Award className="h-5 w-5 text-muted-foreground" />
                          <div className="flex-1">
                            <FormLabel>职称</FormLabel>
                            <FormControl>
                              <Input
                                {...field}
                                disabled={!isEditing}
                                className="mt-1"
                              />
                            </FormControl>
                            <FormMessage />
                          </div>
                        </div>
                      </FormItem>
                    )}
                  />
                </div>
              </CardContent>
            </Card>
          )}
        </form>
      </Form>
    </div>
  );
}
