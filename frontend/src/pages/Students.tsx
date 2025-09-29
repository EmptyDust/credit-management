import { useState, useEffect } from "react";
import * as z from "zod";
import { useAuth } from "@/contexts/AuthContext";
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
import { ImportDialog } from "@/components/ui/import-dialog";
import apiClient from "@/lib/api";
import {
  PlusCircle,
  Users,
  School,
  GraduationCap,
  UserCheck,
  Upload,
  Download,
} from "lucide-react";
import { getOptions } from "@/lib/options";
import { getStatusBadge } from "@/lib/common-utils.tsx";
import { DeleteConfirmDialog } from "@/components/ui/delete-confirm-dialog";
import { useUserManagement } from "@/hooks/useUserManagement";
import { useListPage } from "@/hooks/useListPage";
import { SearchFilterBar } from "@/components/ui/search-filter-bar";
import { TableActions, createEditDeleteActions } from "@/components/ui/table-actions";
import { DataTable } from "@/components/ui/data-table";
import { PageHeader, createPageActions } from "@/components/ui/page-header";
import { StatsGrid } from "@/components/ui/stats-grid";

// Updated Student type based on student.go
export type Student = {
  id?: string;
  username: string;
  real_name: string;
  college?: string | null;
  major?: string | null;
  class?: string | null;
  grade?: string | null;
  email?: string;
  phone?: string | null;
  status: "active" | "inactive" | "suspended";
  avatar?: string;
  last_login_at?: string | null;
  register_time?: string;
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
  id: z
    .string()
    .length(8, "学号必须是8位数字")
    .regex(/^\d{8}$/, "学号必须是8位数字")
    .optional()
    .or(z.literal("")),
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
  status: z.enum(["active", "inactive", "suspended"]),
  user_type: z.literal("student"),
});

const defaultValues = {
  username: "",
  password: "",
  id: "",
  real_name: "",
  college: "",
  major: "",
  class: "",
  phone: "",
  email: "",
  grade: "",
  status: "active" as const,
  user_type: "student" as const,
};

export default function StudentsPage() {
  const { hasPermission } = useAuth();
  const [students, setStudents] = useState<Student[]>([]);

  // 选项
  const [collegeOptions, setCollegeOptions] = useState<{ value: string; label: string }[]>([]);

  // 使用新的通用列表页面hook
  const listPage = useListPage({
    endpoint: "/search/users",
    setData: setStudents,
    errorMessage: "获取学生列表失败",
    userType: "student" // 添加用户类型参数
  });

  // 使用用户管理hook
  const userManagement = useUserManagement({
    userType: "student",
    formSchema,
    defaultValues,
    fetchFunction: listPage.fetchList,
  });

  // 获取统计数据
  const fetchStats = async () => {
    try {
      await apiClient.get("/users/stats", {
        params: { user_type: "student" },
      });
      // 处理统计数据...
    } catch (error) {
      console.error("Failed to fetch stats:", error);
    }
  };

  useEffect(() => {
    listPage.fetchList();
    fetchStats();
    (async () => {
      try {
        const opts = await getOptions();
        setCollegeOptions(opts.colleges);
      } catch (e) {
        console.error("Failed to load options", e);
      }
    })();
  }, []);

  // 处理搜索和过滤
  const handleSearchAndFilter = () => {
    listPage.handleSearchAndFilter();
  };

  // 处理分页变化
  const handlePageChange = (page: number) => {
    listPage.handlePageChange(page);
  };

  const handlePageSizeChange = (size: number) => {
    listPage.handlePageSizeChange(size);
  };

  return (
    <div className="space-y-8 p-4 md:p-8 bg-background min-h-screen">
      <PageHeader
        title="学生列表"
        description="管理学生用户信息"
        actions={createPageActions(
          hasPermission("manage_students") ? {
            label: "添加学生",
            onClick: () => userManagement.handleDialogOpen(null),
            icon: PlusCircle,
          } : undefined,
          hasPermission("manage_students") ? [
            {
              label: "批量导入",
              onClick: () => userManagement.setIsImportDialogOpen(true),
              icon: Upload,
            },
            {
              label: "导出数据",
              onClick: userManagement.handleExport,
              icon: Download,
            },
          ] : undefined
        )}
      />

      {/* Statistics Cards */}
      <StatsGrid
        stats={[
          {
            title: "总学生数",
            value: students.length,
            icon: Users,
            color: "info",
            subtitle: `活跃学生: ${students.filter((s) => s.status === "active").length}`,
          },
          {
            title: "学院数量",
            value: collegeOptions.length,
            icon: School,
            color: "purple",
            subtitle: "不同学院",
          },
          {
            title: "专业数量",
            value: Array.from(new Set(students.map((s) => s.major).filter(Boolean))).length,
            icon: GraduationCap,
            color: "success",
            subtitle: "不同专业",
          },
          {
            title: "年级分布",
            value: Array.from(new Set(students.map((s) => s.grade).filter(Boolean))).length,
            icon: UserCheck,
            color: "warning",
            subtitle: "不同年级",
          },
        ]}
      />

      {/* 搜索和过滤栏 */}
      <div className="mb-6">
        <SearchFilterBar
          searchQuery={listPage.searchQuery}
          onSearchChange={listPage.setSearchQuery}
          onSearch={handleSearchAndFilter}
          onRefresh={() => listPage.fetchList()}
          filterOptions={collegeOptions}
          filterValue={listPage.filterValue}
          onFilterChange={listPage.setFilterValue}
          filterPlaceholder="选择学院"
          searchPlaceholder="搜索学生姓名、学号..."
        />
      </div>

      {/* 状态过滤 */}
      <div className="mb-4">
        <Select value={userManagement.statusFilter} onValueChange={userManagement.setStatusFilter}>
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
      </div>

      {/* Students Table */}
      <Card className="bg-gray-100/80 dark:bg-gray-900/40 border-0 shadow-sm">
        <CardHeader>
          <CardTitle>学生列表</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-md bg-white dark:bg-gray-900/60">
            <DataTable
              data={students}
              columns={[
                {
                  key: "name",
                  header: "姓名",
                  render: (student: Student) => (
                    <span className="font-medium">{student.real_name}</span>
                  ),
                },
                {
                  key: "id",
                  header: "学号",
                  render: (student: Student) => student.id || "-",
                },
                {
                  key: "college",
                  header: "学院",
                  render: (student: Student) => student.college || "-",
                },
                {
                  key: "major",
                  header: "专业",
                  render: (student: Student) => student.major || "-",
                },
                {
                  key: "class",
                  header: "班级",
                  render: (student: Student) => student.class || "-",
                },
                {
                  key: "status",
                  header: "状态",
                  render: (student: Student) => getStatusBadge(student.status),
                },
                {
                  key: "actions",
                  header: "操作",
                  render: (student: Student) => (
                    <TableActions
                      actions={createEditDeleteActions(
                        () => userManagement.handleDialogOpen(student),
                        () => {
                          userManagement.setItemToDelete(student);
                          userManagement.setDeleteDialogOpen(true);
                        },
                        hasPermission("update_student"),
                        hasPermission("delete_student")
                      )}
                    />
                  ),
                },
              ]}
              loading={listPage.loading}
              emptyState={{
                icon: Users,
                title: "暂无学生数据",
                description: "还没有添加任何学生信息",
                action: hasPermission("create_student") ? (
                  <Button
                    onClick={() => userManagement.handleDialogOpen(null)}
                    className="bg-blue-600 hover:bg-blue-700"
                  >
                    <PlusCircle className="h-4 w-4 mr-2" />
                    添加学生
                  </Button>
                ) : undefined,
              }}
              pagination={
                listPage.totalItems > 0
                  ? {
                      currentPage: listPage.currentPage,
                      totalPages: listPage.totalPages,
                      totalItems: listPage.totalItems,
                      pageSize: listPage.pageSize,
                      onPageChange: handlePageChange,
                      onPageSizeChange: handlePageSizeChange,
                    }
                  : undefined
              }
            />
          </div>
        </CardContent>
      </Card>

      {/* Create/Edit Dialog */}
      <Dialog open={userManagement.isDialogOpen} onOpenChange={userManagement.setIsDialogOpen}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>
              {userManagement.editingItem ? "编辑学生" : "添加新学生"}
            </DialogTitle>
            <DialogDescription>
              {userManagement.editingItem ? "修改学生信息" : "填写学生详细信息"}
            </DialogDescription>
          </DialogHeader>
          <Form {...userManagement.form}>
            <form
              onSubmit={userManagement.form.handleSubmit(userManagement.onSubmit)}
              className="grid grid-cols-2 gap-4 py-4"
            >
              <FormField
                control={userManagement.form.control}
                name="username"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>用户名</FormLabel>
                    <FormControl>
                      <Input
                        {...field}
                        disabled={!!userManagement.editingItem}
                        placeholder="请输入用户名"
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {!userManagement.editingItem && (
                <FormField
                  control={userManagement.form.control}
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
                control={userManagement.form.control}
                name="id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>学号</FormLabel>
                    <FormControl>
                      <Input {...field} disabled={!!userManagement.editingItem} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={userManagement.form.control}
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
                control={userManagement.form.control}
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
                control={userManagement.form.control}
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
                control={userManagement.form.control}
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
                control={userManagement.form.control}
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
                control={userManagement.form.control}
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
                control={userManagement.form.control}
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
                control={userManagement.form.control}
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
                <Button type="submit" disabled={userManagement.isSubmitting}>
                  {userManagement.isSubmitting ? "保存中..." : "保存"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <DeleteConfirmDialog
        open={userManagement.deleteDialogOpen}
        onOpenChange={userManagement.setDeleteDialogOpen}
        title="确认删除学生"
        itemName={userManagement.itemToDelete?.real_name}
        onConfirm={userManagement.handleDeleteConfirm}
      />

      {/* Import Dialog */}
      <ImportDialog
        open={userManagement.isImportDialogOpen}
        onOpenChange={userManagement.setIsImportDialogOpen}
        title="批量导入学生"
        description="请选择Excel或CSV文件进行批量导入。文件应包含学生的基本信息。"
        userType="student"
        onImport={userManagement.handleImport}
        importing={userManagement.importing}
      />
    </div>
  );
}
