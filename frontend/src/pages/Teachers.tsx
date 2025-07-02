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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Pagination } from "@/components/ui/pagination";
import apiClient from "@/lib/api";
import {
  PlusCircle,
  Search,
  Edit,
  Trash,
  Filter,
  RefreshCw,
  Users,
  Building,
  AlertCircle,
  Upload,
  Download,
} from "lucide-react";
import toast from "react-hot-toast";
import { StatCard } from "@/components/ui/stat-card";
import { getStatusBadge } from "@/lib/status-utils";
import { DeleteConfirmDialog } from "@/components/ui/delete-confirm-dialog";
import { ImportDialog } from "@/components/ui/import-dialog";
import { apiHelpers } from "@/lib/api";

// Teacher type based on teacher.go
export type Teacher = {
  user_id?: string;
  username: string;
  real_name: string;
  email?: string;
  phone?: string | null;
  department?: string | null;
  title?: string | null;
  status: "active" | "inactive" | "suspended";
  avatar?: string;
  last_login_at?: string | null;
  register_time?: string;
};

// Form schema for validation
const formSchema = z.object({
  username: z.string().min(1, "用户名不能为空").max(20, "用户名最多20个字符"),
  password: z
    .string()
    .min(8, "密码至少8个字符")
    .regex(/[A-Z]/, "密码必须包含至少一个大写字母")
    .regex(/[a-z]/, "密码必须包含至少一个小写字母")
    .regex(/[0-9]/, "密码必须包含至少一个数字")
    .optional(),
  real_name: z.string().min(1, "姓名不能为空").max(50, "姓名最多50个字符"),
  email: z.string().email({ message: "请输入有效的邮箱地址" }),
  phone: z
    .string()
    .regex(/^1[3-9]\d{9}$/, "请输入有效的11位手机号")
    .optional()
    .or(z.literal("")),
  department: z.string().min(1, "院系不能为空"),
  title: z.string().optional().or(z.literal("")),
  status: z.enum(["active", "inactive", "suspended"]),
  user_type: z.literal("teacher"),
});

export default function TeachersPage() {
  const { hasPermission } = useAuth();
  const [teachers, setTeachers] = useState<Teacher[]>([]);
  const [loading, setLoading] = useState(true);
  const [error] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [departmentFilter, setDepartmentFilter] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingTeacher, setEditingTeacher] = useState<Teacher | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [teacherToDelete, setTeacherToDelete] = useState<Teacher | null>(null);
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const [importing, setImporting] = useState(false);

  // 分页状态
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [totalItems, setTotalItems] = useState(0);
  const [totalPages, setTotalPages] = useState(0);

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      username: "",
      password: "",
      real_name: "",
      email: "",
      phone: "",
      department: "",
      status: "active",
      user_type: "teacher",
    },
  });

  const fetchTeachers = async (page = currentPage, size = pageSize) => {
    try {
      setLoading(true);

      // 构建查询参数
      const params: any = {
        page,
        page_size: size,
        user_type: "teacher", 
      };

      if (searchQuery) {
        params.query = searchQuery;
      }
      if (departmentFilter !== "all") {
        params.department = departmentFilter;
      }
      if (statusFilter !== "all") {
        params.status = statusFilter;
      }

      const response = await apiClient.get("/search/users", { params });

      // 使用统一的响应处理函数
      const { data: teachersData, pagination: paginationData } = apiHelpers.processPaginatedResponse(response);

      setTeachers(teachersData);
      setTotalItems(paginationData.total);
      setTotalPages(paginationData.total_pages);
    } catch (error) {
      console.error("Failed to fetch teachers:", error);
      toast.error("获取教师列表失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTeachers();
  }, []);

  // 处理分页变化
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    fetchTeachers(page, pageSize);
  };

  // 处理每页数量变化
  const handlePageSizeChange = (size: number) => {
    setPageSize(size);
    setCurrentPage(1);
    fetchTeachers(1, size);
  };

  // 处理搜索和筛选
  const handleSearchAndFilter = () => {
    setCurrentPage(1);
    fetchTeachers(1, pageSize);
  };

  useEffect(() => {
    handleSearchAndFilter();
  }, [searchQuery, departmentFilter, statusFilter]);

  const handleDialogOpen = (teacher: Teacher | null) => {
    setEditingTeacher(teacher);
    if (teacher) {
      // 转换null值为空字符串以匹配表单schema
      const formData = {
        ...teacher,
        department: teacher.department || "",
        title: teacher.title || "",
        phone: teacher.phone || "",
      };
      form.reset(formData);
    } else {
      form.reset({
        username: "",
        password: "",
        real_name: "",
        email: "",
        phone: "",
        department: "",
        title: "",
        status: "active",
        user_type: "teacher",
      });
    }
    setIsDialogOpen(true);
  };

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    setIsSubmitting(true);
    try {
      if (editingTeacher) {
        if (!editingTeacher.user_id) {
          toast.error("无法找到教师ID");
          return;
        }
        await apiClient.put(`/users/${editingTeacher.user_id}`, values);
        toast.success("教师信息更新成功");
      } else {
        // 使用正确的API端点创建教师
        const createData = {
          ...values,
          password: values.password || "Password123", // Default password that meets requirements
          user_type: "teacher",
        };
        await apiClient.post("/users/teachers", createData);
        toast.success("教师创建成功");
      }
      setIsDialogOpen(false);
      fetchTeachers();
    } catch (err: any) {
      if (err.response?.status === 409) {
        toast.error("用户名或邮箱已存在");
      } else {
        toast.error(`教师${editingTeacher ? "更新" : "创建"}失败`);
      }
      console.error(err);
    } finally {
      setIsSubmitting(false);
    }
  };

  const departments = Array.from(
    new Set(teachers.map((t) => t.department).filter(Boolean))
  );
  const canManageTeachers = hasPermission("manage_teachers");

  const handleDeleteConfirm = async () => {
    if (!teacherToDelete) return;

    try {
      await apiClient.delete(`/users/${teacherToDelete.user_id}`);
      toast.success("教师删除成功");
      fetchTeachers();
    } catch (err) {
      toast.error("删除教师失败");
    } finally {
      setDeleteDialogOpen(false);
      setTeacherToDelete(null);
    }
  };

  const handleImport = async (file: File) => {
    setImporting(true);
    try {
      const formData = new FormData();
      formData.append("file", file);
      formData.append("user_type", "teacher");

      const response = await apiClient.post("/users/import", formData, {
        headers: {
          // Remove Content-Type header to let browser set it with boundary
        },
      });

      if (response.data.code === 0) {
        toast.success("批量导入成功");
        setIsImportDialogOpen(false);
        fetchTeachers();
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
      const response = await apiClient.get("/users/export", {
        params: { user_type: "teacher" },
        responseType: "blob",
      });

      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement("a");
      link.href = url;
      link.setAttribute(
        "download",
        `teachers_${new Date().toISOString().split("T")[0]}.xlsx`
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
    <div className="space-y-8 p-4 md:p-8 bg-background min-h-screen">
      <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">教师列表</h1>
          <p className="text-muted-foreground">管理教师用户信息</p>
        </div>
        <div className="flex items-center gap-2">
          {canManageTeachers && (
            <Button
              onClick={() => handleDialogOpen(null)}
              className="rounded-lg shadow transition-all duration-200 hover:scale-105"
            >
              <PlusCircle className="mr-2 h-4 w-4" />
              添加教师
            </Button>
          )}
          {canManageTeachers && (
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
          title="总教师数"
          value={teachers.length}
          icon={Users}
          color="info"
          subtitle={`活跃教师: ${
            teachers.filter((t) => t.status === "active").length
          }`}
        />
        <StatCard
          title="院系数量"
          value={departments.length}
          icon={Building}
          color="purple"
          subtitle="不同院系"
        />
        <StatCard
          title="活跃教师数"
          value={teachers.filter((t) => t.status === "active").length}
          icon={Users}
          color="success"
          subtitle="当前活跃"
        />
      </div>

      {/* Filters */}
      <Card className="bg-white/80 dark:bg-gray-900/80 backdrop-blur border-0 shadow-sm">
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
                  placeholder="搜索姓名、院系或专业..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>
            <Select
              value={departmentFilter}
              onValueChange={setDepartmentFilter}
            >
              <SelectTrigger className="w-48">
                <SelectValue placeholder="选择院系" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="all">全部院系</SelectItem>
                {departments.map((department) => (
                  <SelectItem
                    key={department ?? "unknown"}
                    value={department ?? ""}
                  >
                    {department ?? "未知院系"}
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
                <SelectItem value="suspended">暂停</SelectItem>
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              onClick={handleSearchAndFilter}
              disabled={loading}
            >
              <RefreshCw
                className={`h-4 w-4 ${loading ? "animate-spin" : ""}`}
              />
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Teachers Table */}
      <Card className="bg-gray-100/80 dark:bg-gray-900/40 border-0 shadow-sm">
        <CardHeader>
          <CardTitle>教师列表</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-md bg-white dark:bg-gray-900/60">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>用户名</TableHead>
                  <TableHead>姓名</TableHead>
                  <TableHead>院系</TableHead>
                  <TableHead>职称</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-8">
                      <div className="flex items-center justify-center gap-2">
                        <RefreshCw className="h-4 w-4 animate-spin" />
                        加载中...
                      </div>
                    </TableCell>
                  </TableRow>
                ) : error ? (
                  <TableRow>
                    <TableCell
                      colSpan={7}
                      className="text-center py-8 text-red-500 dark:text-red-400"
                    >
                      {error}
                    </TableCell>
                  </TableRow>
                ) : teachers.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={canManageTeachers ? 6 : 5}
                      className="text-center py-8"
                    >
                      <div className="flex flex-col items-center gap-2 text-muted-foreground">
                        <AlertCircle className="w-8 h-8" />
                        <p>暂无教师记录</p>
                      </div>
                    </TableCell>
                  </TableRow>
                ) : (
                  teachers.map((teacher) => (
                    <TableRow key={teacher.username}>
                      <TableCell className="font-medium">
                        {teacher.username}
                      </TableCell>
                      <TableCell>
                        <div>
                          <div className="font-medium">{teacher.real_name}</div>
                          <div className="text-sm text-muted-foreground">
                            {teacher.email}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>{teacher.department || "-"}</TableCell>
                      <TableCell>{teacher.title || "-"}</TableCell>
                      <TableCell>{getStatusBadge(teacher.status)}</TableCell>
                      <TableCell className="text-right space-x-2">
                        {canManageTeachers && (
                          <>
                            <Button
                              variant="outline"
                              size="icon"
                              onClick={() => handleDialogOpen(teacher)}
                            >
                              <Edit className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="destructive"
                              size="icon"
                              onClick={() => {
                                setTeacherToDelete(teacher);
                                setDeleteDialogOpen(true);
                              }}
                            >
                              <Trash className="h-4 w-4" />
                            </Button>
                          </>
                        )}
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          {/* 分页组件 */}
          {!loading && totalItems > 0 && (
            <Pagination
              currentPage={currentPage}
              totalPages={totalPages}
              totalItems={totalItems}
              pageSize={pageSize}
              onPageChange={handlePageChange}
              onPageSizeChange={handlePageSizeChange}
            />
          )}
        </CardContent>
      </Card>

      {/* Create/Edit Dialog */}
      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>
              {editingTeacher ? "编辑教师" : "添加新教师"}
            </DialogTitle>
            <DialogDescription>
              {editingTeacher ? "修改教师信息" : "填写教师详细信息"}
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
                        disabled={!!editingTeacher}
                        placeholder="请输入用户名"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {!editingTeacher && (
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
                name="department"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>院系</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="选择院系" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {departments.map((dep) => (
                          <SelectItem key={dep} value={dep ?? ""}>
                            {dep ?? "未知院系"}
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
                name="title"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>职称</FormLabel>
                    <FormControl>
                      <Input {...field} placeholder="请输入职称" />
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
                        <SelectItem value="suspended">暂停</SelectItem>
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
      <DeleteConfirmDialog
        open={deleteDialogOpen}
        onOpenChange={setDeleteDialogOpen}
        title="确认删除教师"
        itemName={teacherToDelete?.real_name}
        onConfirm={handleDeleteConfirm}
      />

      {/* Import Dialog */}
      <ImportDialog
        open={isImportDialogOpen}
        onOpenChange={setIsImportDialogOpen}
        title="批量导入教师"
        description="请选择Excel或CSV文件进行批量导入。文件应包含教师的基本信息。"
        userType="teacher"
        onImport={handleImport}
        importing={importing}
      />
    </div>
  );
}
