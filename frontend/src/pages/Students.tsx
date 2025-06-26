import { useState, useEffect } from "react";
import * as z from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useAuth } from "@/contexts/AuthContext";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import apiClient from "@/lib/api";
import {
  PlusCircle,
  Search,
  Edit,
  Trash,
  Filter,
  RefreshCw,
  Users,
  School,
  GraduationCap,
  AlertCircle,
  UserCheck,
  Eye,
  Upload,
  Download,
} from "lucide-react";
import toast from "react-hot-toast";

// Updated Student type based on student.go
export type Student = {
  user_id?: string;
  username: string;
  real_name: string;
  student_id?: string;
  college?: string;
  major?: string;
  class?: string;
  grade?: string;
  email?: string;
  phone?: string;
  status: "active" | "inactive";
  avatar?: string;
  register_time?: string;
  created_at?: string;
  updated_at?: string;
};

// Form schema for validation
const formSchema = z.object({
  username: z
    .string()
    .min(3, "用户名至少3个字符")
    .max(20, "用户名最多20个字符")
    .regex(/^[a-zA-Z0-9_]+$/, "用户名只能包含字母、数字和下划线"),
  password: z
    .string()
    .min(8, "密码至少8个字符")
    .regex(/[A-Z]/, "密码必须包含至少一个大写字母")
    .regex(/[a-z]/, "密码必须包含至少一个小写字母")
    .regex(/[0-9]/, "密码必须包含至少一个数字")
    .optional(),
  student_id: z
    .string()
    .length(8, "学号必须是8位数字")
    .regex(/^\d{8}$/, "学号必须是8位数字")
    .optional(),
  real_name: z.string().min(2, "姓名至少2个字符").max(50, "姓名最多50个字符"),
  college: z.string().min(1, "学院不能为空").max(100, "学院名称最多100个字符"),
  major: z.string().min(1, "专业不能为空").max(100, "专业名称最多100个字符"),
  class: z.string().min(1, "班级不能为空").max(50, "班级名称最多50个字符"),
  phone: z
    .string()
    .regex(/^1[3-9]\d{9}$/, "请输入有效的11位手机号")
    .optional()
    .or(z.literal("")),
  email: z
    .string()
    .email({ message: "请输入有效的邮箱地址" })
    .optional()
    .or(z.literal("")),
  grade: z
    .string()
    .length(4, "年级必须是4位数字")
    .regex(/^\d{4}$/, "年级必须是4位数字"),
  status: z.enum(["active", "inactive"]),
  user_type: z.literal("student"),
});

// 统计卡片样式与仪表盘一致
const StatCard = ({
  title,
  value,
  icon: Icon,
  color = "default",
  subtitle,
}: {
  title: string;
  value: string | number;
  icon: React.ElementType;
  color?: "default" | "success" | "warning" | "danger" | "info" | "purple";
  subtitle?: string;
}) => {
  const colorClasses = {
    default: "text-muted-foreground",
    success: "text-green-600",
    warning: "text-yellow-600",
    danger: "text-red-600",
    info: "text-blue-600",
    purple: "text-purple-600",
  };
  const bgClasses = {
    default: "bg-muted/20",
    success: "bg-green-100 dark:bg-green-900/20",
    warning: "bg-yellow-100 dark:bg-yellow-900/20",
    danger: "bg-red-100 dark:bg-red-900/20",
    info: "bg-blue-100 dark:bg-blue-900/20",
    purple: "bg-purple-100 dark:bg-purple-900/20",
  };
  return (
    <Card className="rounded-xl shadow-lg hover:shadow-2xl transition-all duration-300 bg-gradient-to-br from-white to-gray-50 dark:from-gray-900 dark:to-gray-800 border-0">
      <CardHeader className="flex flex-row items-center justify-between pb-3">
        <div className={`p-3 rounded-xl ${bgClasses[color]}`}>
          <Icon className={`h-6 w-6 ${colorClasses[color]}`} />
        </div>
        <CardTitle className="text-lg font-semibold text-gray-900 dark:text-gray-100">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-bold mb-1 text-gray-900 dark:text-gray-100">
          {value}
        </div>
        {subtitle && (
          <div className="text-sm text-muted-foreground mb-2">{subtitle}</div>
        )}
      </CardContent>
    </Card>
  );
};

export default function StudentsPage() {
  const { user, hasPermission } = useAuth();
  const [students, setStudents] = useState<Student[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [collegeFilter, setCollegeFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingStudent, setEditingStudent] = useState<Student | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [studentToDelete, setStudentToDelete] = useState<Student | null>(null);
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const [importFile, setImportFile] = useState<File | null>(null);
  const [importing, setImporting] = useState(false);

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: "",
      password: "",
      student_id: "",
      real_name: "",
      college: "",
      major: "",
      class: "",
      phone: "",
      email: "",
      grade: "",
      status: "active",
      user_type: "student",
    },
  });

  const fetchStudents = async () => {
    try {
      setLoading(true);
      setError("");
      let endpoint = "/students";
      const params = new URLSearchParams();

      if (searchQuery) {
        params.append("q", searchQuery);
      }
      if (collegeFilter !== "all") {
        params.append("college", collegeFilter);
      }
      if (statusFilter !== "all") {
        params.append("status", statusFilter);
      }

      if (params.toString()) {
        endpoint += `?${params.toString()}`;
      }

      console.log("Fetching students from:", endpoint);
      const response = await apiClient.get(endpoint);
      console.log("API response:", response.data);

      // 处理不同的响应格式
      let studentsData = [];
      if (response.data?.data?.users) {
        studentsData = response.data.data.users;
      } else if (response.data?.students) {
        studentsData = response.data.students;
      } else if (Array.isArray(response.data)) {
        studentsData = response.data;
      } else {
        studentsData = [];
      }

      setStudents(studentsData);
      console.log("Students loaded:", studentsData.length);
    } catch (err: any) {
      const errorMessage =
        err.response?.data?.error || err.message || "获取学生列表失败";
      setError(errorMessage);
      console.error("Error fetching students:", err);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchStudents();
  }, [searchQuery, collegeFilter, statusFilter]);

  const handleDialogOpen = (student: Student | null) => {
    setEditingStudent(student);
    if (student) {
      form.reset(student);
    } else {
      form.reset({
        username: "",
        password: "",
        student_id: "",
        real_name: "",
        college: "",
        major: "",
        class: "",
        phone: "",
        email: "",
        grade: "",
        status: "active",
        user_type: "student",
      });
    }
    setIsDialogOpen(true);
  };

  const handleDelete = async (studentId: string) => {
    try {
      // Find the student by student_id to get the user_id
      const student = students.find((s) => s.student_id === studentId);
      if (!student || !student.user_id) {
        toast.error("无法找到学生信息");
        return;
      }
      await apiClient.delete(`/students/${student.user_id}`);
      fetchStudents();
      toast.success("学生删除成功");
    } catch (err) {
      toast.error("删除学生失败");
      console.error(err);
    }
  };

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    setIsSubmitting(true);
    try {
      if (editingStudent) {
        if (!editingStudent.user_id) {
          toast.error("无法找到学生ID");
          return;
        }
        await apiClient.put(`/students/${editingStudent.user_id}`, values);
        toast.success("学生信息更新成功");
      } else {
        // For creating a new student, ensure required fields are included
        const createData = {
          ...values,
          password: values.password || "Password123", // Default password that meets requirements
          user_type: "student",
        };
        await apiClient.post("/students", createData);
        toast.success("学生创建成功");
      }
      setIsDialogOpen(false);
      fetchStudents();
    } catch (err: any) {
      // 只保留兜底错误提示，手机号等冲突交给全局拦截器
      if (!err.response || err.response.status !== 409) {
        toast.error(`学生${editingStudent ? "更新" : "创建"}失败`);
      }
      console.error(err);
    } finally {
      setIsSubmitting(false);
    }
  };

  const getStatusBadge = (status: string) => {
    const statusConfig = {
      active: { label: "活跃", color: "bg-green-100 text-green-800" },
      inactive: { label: "停用", color: "bg-gray-100 text-gray-800" },
    };

    const config =
      statusConfig[status as keyof typeof statusConfig] ||
      statusConfig.inactive;
    return <Badge className={config.color}>{config.label}</Badge>;
  };

  const filteredStudents = students.filter((student) => {
    const matchesSearch =
      (student.real_name || "")
        .toLowerCase()
        .includes(searchQuery.toLowerCase()) ||
      (student.student_id || "")
        .toLowerCase()
        .includes(searchQuery.toLowerCase()) ||
      (student.major || "").toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCollege =
      collegeFilter === "all" || student.college === collegeFilter;
    const matchesStatus =
      statusFilter === "all" || student.status === statusFilter;
    return matchesSearch && matchesCollege && matchesStatus;
  });

  const colleges = Array.from(
    new Set(students.map((s) => s.college).filter(Boolean))
  );
  const canManageStudents = hasPermission("manage_students");
  const canViewStudents = hasPermission("view_students");

  const handleDeleteConfirm = async () => {
    if (!studentToDelete) return;

    try {
      await apiClient.delete(`/users/${studentToDelete.id}`);
      toast.success("学生删除成功");
      fetchStudents();
    } catch (err) {
      toast.error("删除学生失败");
    } finally {
      setDeleteDialogOpen(false);
      setStudentToDelete(null);
    }
  };

  const handleImportFile = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      const allowedTypes = [
        "application/vnd.ms-excel",
        "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "text/csv",
      ];
      if (!allowedTypes.includes(file.type)) {
        toast.error("请选择Excel或CSV文件");
        return;
      }
      setImportFile(file);
    }
  };

  const handleImport = async () => {
    if (!importFile) return;

    setImporting(true);
    try {
      const formData = new FormData();
      formData.append("file", importFile);

      const response = await apiClient.post(
        "/users/import/students",
        formData,
        {
          headers: {
            "Content-Type": "multipart/form-data",
          },
        }
      );

      if (response.data.code === 0) {
        toast.success("批量导入成功");
        setIsImportDialogOpen(false);
        setImportFile(null);
        fetchStudents();
      } else {
        toast.error(response.data.message || "导入失败");
      }
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || "导入失败";
      toast.error(errorMessage);
    } finally {
      setImporting(false);
    }
  };

  const handleExport = async () => {
    try {
      const response = await apiClient.get("/users/export/students", {
        responseType: "blob",
      });

      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement("a");
      link.href = url;
      link.setAttribute(
        "download",
        `students_${new Date().toISOString().split("T")[0]}.xlsx`
      );
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);

      toast.success("导出成功");
    } catch (err) {
      toast.error("导出失败");
    }
  };

  return (
    <div className="space-y-8 p-4 md:p-8 bg-gray-50 min-h-screen">
      <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">学生列表</h1>
          <p className="text-muted-foreground">管理学生用户信息</p>
        </div>
        <div className="flex items-center gap-2">
          {canManageStudents && (
            <Button
              onClick={() => setIsDialogOpen(true)}
              className="rounded-lg shadow transition-all duration-200 hover:scale-105"
            >
              <PlusCircle className="mr-2 h-4 w-4" />
              添加学生
            </Button>
          )}
          {canManageStudents && (
            <>
              <Button
                onClick={() => setIsImportDialogOpen(true)}
                variant="outline"
                className="rounded-lg shadow transition-all duration-200 hover:scale-105"
              >
                <Upload className="mr-2 h-4 w-4" />
                批量导入
              </Button>
              <Button
                onClick={handleExport}
                variant="outline"
                className="rounded-lg shadow transition-all duration-200 hover:scale-105"
              >
                <Download className="mr-2 h-4 w-4" />
                导出数据
              </Button>
            </>
          )}
        </div>
      </div>

      {/* Statistics Cards */}
      <div className="grid gap-4 md:grid-cols-4">
        <StatCard
          title="总学生数"
          value={students.length}
          icon={Users}
          color="info"
          subtitle={`活跃学生: ${
            students.filter((s) => s.status === "active").length
          }`}
        />
        <StatCard
          title="学院数量"
          value={colleges.length}
          icon={School}
          color="purple"
          subtitle="不同学院"
        />
        <StatCard
          title="专业数量"
          value={
            Array.from(new Set(students.map((s) => s.major).filter(Boolean)))
              .length
          }
          icon={GraduationCap}
          color="success"
          subtitle="不同专业"
        />
        <StatCard
          title="年级分布"
          value={
            Array.from(new Set(students.map((s) => s.grade).filter(Boolean)))
              .length
          }
          icon={UserCheck}
          color="warning"
          subtitle="不同年级"
        />
      </div>

      {/* Filters */}
      <Card className="bg-white/80 backdrop-blur border-0 shadow-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Filter className="h-5 w-5" />
            筛选和搜索
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="搜索姓名、学号或专业..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>
            <Select value={collegeFilter} onValueChange={setCollegeFilter}>
              <SelectTrigger className="w-48">
                <SelectValue placeholder="选择学院" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部学院</SelectItem>
                {colleges.map((college) => (
                  <SelectItem key={college || "unknown"} value={college || ""}>
                    {college || "未知学院"}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Select value={statusFilter} onValueChange={setStatusFilter}>
              <SelectTrigger className="w-32">
                <SelectValue placeholder="状态" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部状态</SelectItem>
                <SelectItem value="active">活跃</SelectItem>
                <SelectItem value="inactive">停用</SelectItem>
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              onClick={fetchStudents}
              disabled={loading}
            >
              <RefreshCw
                className={`h-4 w-4 ${loading ? "animate-spin" : ""}`}
              />
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Students Table */}
      <Card className="bg-gray-100/80 dark:bg-gray-900/40 border-0 shadow-sm">
        <CardHeader>
          <CardTitle>学生列表</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-md bg-white dark:bg-gray-900/60">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>学号</TableHead>
                  <TableHead>姓名</TableHead>
                  <TableHead>学院</TableHead>
                  <TableHead>专业</TableHead>
                  <TableHead>班级</TableHead>
                  <TableHead>年级</TableHead>
                  <TableHead>状态</TableHead>
                  {canManageStudents && (
                    <TableHead className="text-right">操作</TableHead>
                  )}
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell
                      colSpan={canManageStudents ? 8 : 7}
                      className="text-center py-8"
                    >
                      <div className="flex items-center justify-center gap-2">
                        <RefreshCw className="h-4 w-4 animate-spin" />
                        加载中...
                      </div>
                    </TableCell>
                  </TableRow>
                ) : error ? (
                  <TableRow>
                    <TableCell
                      colSpan={canManageStudents ? 8 : 7}
                      className="text-center py-8 text-red-500"
                    >
                      {error}
                    </TableCell>
                  </TableRow>
                ) : filteredStudents.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={canManageStudents ? 8 : 7}
                      className="text-center py-8"
                    >
                      <div className="flex flex-col items-center gap-2 text-muted-foreground">
                        <AlertCircle className="h-8 w-8" />
                        <p>暂无学生记录</p>
                      </div>
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredStudents.map((student, index) => (
                    <TableRow key={student.student_id || `student-${index}`}>
                      <TableCell className="font-medium">
                        {student.student_id || "-"}
                      </TableCell>
                      <TableCell>
                        <div>
                          <div className="font-medium">{student.real_name}</div>
                          <div className="text-sm text-muted-foreground">
                            {student.username}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>{student.college || "-"}</TableCell>
                      <TableCell>{student.major || "-"}</TableCell>
                      <TableCell>{student.class || "-"}</TableCell>
                      <TableCell>{student.grade || "-"}</TableCell>
                      <TableCell>{getStatusBadge(student.status)}</TableCell>
                      {canManageStudents && (
                        <TableCell className="text-right space-x-2">
                          <Button
                            variant="outline"
                            size="icon"
                            onClick={() => handleDialogOpen(student)}
                          >
                            <Edit className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="destructive"
                            size="icon"
                            onClick={() => {
                              setStudentToDelete(student);
                              setDeleteDialogOpen(true);
                            }}
                          >
                            <Trash className="h-4 w-4" />
                          </Button>
                        </TableCell>
                      )}
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* Create/Edit Dialog */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>
              {editingStudent ? "编辑学生" : "添加新学生"}
            </DialogTitle>
            <DialogDescription>
              {editingStudent ? "修改学生信息" : "填写学生详细信息"}
            </DialogDescription>
          </DialogHeader>
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(onSubmit)}
              className="grid grid-cols-2 gap-4 py-4"
            >
              <FormField
                control={form.control}
                name="username"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>用户名</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        disabled={!!editingStudent}
                        placeholder="请输入用户名"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {!editingStudent && (
                <FormField
                  control={form.control}
                  name="password"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>密码</FormLabel>
                      <FormControl>
                        <Input
                          {...field}
                          type="password"
                          placeholder="至少8位，包含大小写字母和数字"
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              )}
              <FormField
                control={form.control}
                name="student_id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>学号</FormLabel>
                    <FormControl>
                      <Input {...field} disabled={!!editingStudent} />
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
                    <FormLabel>姓名</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="email"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>邮箱</FormLabel>
                    <FormControl>
                      <Input {...field} type="email" />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="college"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>学院</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="选择学院" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {colleges.map((college) => (
                          <SelectItem key={college} value={college ?? ""}>
                            {college ?? "未知学院"}
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
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="class"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>班级</FormLabel>
                    <FormControl>
                      <Input {...field} />
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
                  <FormItem className="col-span-2">
                    <FormLabel>联系方式</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        placeholder="请输入11位手机号，如：13812345678"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="status"
                render={({ field }) => (
                  <FormItem className="col-span-2">
                    <FormLabel>状态</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="active">活跃</SelectItem>
                        <SelectItem value="inactive">停用</SelectItem>
                      </SelectContent>
                    </Select>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <DialogFooter className="col-span-2">
                <Button type="submit" disabled={isSubmitting}>
                  {isSubmitting ? "保存中..." : "保存"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>确认删除</DialogTitle>
            <DialogDescription>
              您确定要删除学生 "{studentToDelete?.real_name}"
              吗？此操作不可撤销。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
            >
              取消
            </Button>
            <Button variant="destructive" onClick={handleDeleteConfirm}>
              删除
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Import Dialog */}
      <Dialog open={isImportDialogOpen} onOpenChange={setIsImportDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Upload className="h-5 w-5" />
              批量导入学生
            </DialogTitle>
            <DialogDescription>
              请选择Excel或CSV文件进行批量导入。文件应包含学生的基本信息。
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <label className="text-sm font-medium">选择文件</label>
              <Input
                type="file"
                accept=".xlsx,.xls,.csv"
                onChange={handleImportFile}
                className="mt-1"
              />
              <p className="text-xs text-muted-foreground mt-1">
                支持Excel (.xlsx, .xls) 和CSV格式，文件大小不超过10MB
              </p>
            </div>
            {importFile && (
              <div className="p-3 bg-muted rounded-lg">
                <p className="text-sm font-medium">已选择文件：</p>
                <p className="text-sm text-muted-foreground">
                  {importFile.name}
                </p>
                <p className="text-xs text-muted-foreground">
                  大小：{(importFile.size / 1024 / 1024).toFixed(2)} MB
                </p>
              </div>
            )}
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => {
                setIsImportDialogOpen(false);
                setImportFile(null);
              }}
            >
              取消
            </Button>
            <Button
              onClick={handleImport}
              disabled={!importFile || importing}
              className="bg-blue-600 hover:bg-blue-700"
            >
              {importing ? (
                <div className="flex items-center gap-2">
                  <RefreshCw className="h-4 w-4 animate-spin" />
                  导入中...
                </div>
              ) : (
                <>
                  <Upload className="mr-2 h-4 w-4" />
                  开始导入
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
