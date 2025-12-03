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
  Filter,
} from "lucide-react";
import { getOptions } from "@/lib/options";
import { getStatusBadge } from "@/lib/status-utils";
import { DeleteConfirmDialog } from "@/components/ui/delete-confirm-dialog";
import { useUserManagement } from "@/hooks/useUserManagement";
import { useListPage } from "@/hooks/useListPage";
import { SearchFilterBar } from "@/components/ui/search-filter-bar";
import { TableActions, createEditDeleteActions } from "@/components/ui/table-actions";
import { DataTable } from "@/components/ui/data-table";
import { Pagination } from "@/components/ui/pagination";
import { PageHeader, createPageActions } from "@/components/ui/page-header";
import { StatsGrid } from "@/components/ui/stats-grid";
import { FilterCard } from "@/components/ui/filter-card";

// Updated Student type based on student.go
export type Student = {
  uuid: string;
  student_id?: string;
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
  student_id: z
    .string()
    .length(8, "学号必须是8位数字")
    .regex(/^\d{8}$/, "学号必须是8位数字"),
  real_name: z.string().min(2, "姓名至少2个字符").max(50, "姓名最多50个字符"),
  college: z.string().min(1, "学部不能为空").max(100, "学部名称最多100个字符"),
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
  student_id: "",
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

type StudentStats = {
  total: number;
  active: number;
  collegeCount: number;
  majorCount: number;
  gradeCount: number;
};

export default function StudentsPage() {
  const { hasPermission } = useAuth();
  const [students, setStudents] = useState<Student[]>([]);
  const [studentStats, setStudentStats] = useState<StudentStats>({
    total: 0,
    active: 0,
    collegeCount: 0,
    majorCount: 0,
    gradeCount: 0,
  });

  // 选项
  const [collegeOptions, setCollegeOptions] = useState<{ value: string; label: string }[]>([]);
  const [majorOptions, setMajorOptions] = useState<Record<string, { value: string; label: string }[]>>({});
  const [classOptions, setClassOptions] = useState<Record<string, { value: string; label: string }[]>>({});
  const [gradeOptions, setGradeOptions] = useState<{ value: string; label: string }[]>([]);
  
  // 筛选状态 - 使用useListPage hook提供的状态

  // 获取统计数据
  const fetchStats = async () => {
    try {
      const response = await apiClient.get("/users/stats/students");
      if (response.data.code === 0) {
        const data = response.data.data || {};
        setStudentStats({
          total: data.total_students || 0,
          active: data.active_students || 0,
          collegeCount: Object.keys(data.students_by_college || {}).length,
          majorCount: Object.keys(data.students_by_major || {}).length,
          gradeCount: Object.keys(data.students_by_grade || {}).length,
        });
      }
    } catch (error) {
      console.error("Failed to fetch stats:", error);
    }
  };

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
    onSuccess: fetchStats,
  });

  useEffect(() => {
    listPage.fetchList();
    fetchStats();
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

  // 实时监听搜索与筛选变化
  useEffect(() => {
    listPage.handleSearchAndFilter();
  }, [
    listPage.searchQuery,
    listPage.filterValue,
    listPage.statusFilter,
    listPage.gradeFilter,
  ]);

  // 处理分页变化
  const handlePageChange = (page: number) => {
    listPage.handlePageChange(page);
  };

  const handlePageSizeChange = (size: number) => {
    listPage.handlePageSizeChange(size);
  };

  return (
    <div className="space-y-8 p-4 md:p-8">
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
            value: studentStats.total,
            icon: Users,
            color: "info",
            subtitle: `活跃学生: ${studentStats.active}`,
          },
          {
            title: "学部数量",
            value: studentStats.collegeCount,
            icon: School,
            color: "purple",
            subtitle: "不同学部",
          },
          {
            title: "专业数量",
            value: studentStats.majorCount,
            icon: GraduationCap,
            color: "success",
            subtitle: "不同专业",
          },
          {
            title: "年级分布",
            value: studentStats.gradeCount,
            icon: UserCheck,
            color: "warning",
            subtitle: "不同年级",
          },
        ]}
      />

      {/* 搜索与筛选 */}
      <FilterCard icon={Filter}>
        <SearchFilterBar
          searchQuery={listPage.searchQuery}
          onSearchChange={listPage.setSearchQuery}
          onSearch={listPage.handleSearchAndFilter}
          onRefresh={() => listPage.fetchList()}
          filterOptions={collegeOptions}
          filterValue={listPage.filterValue}
          onFilterChange={listPage.setFilterValue}
          filterPlaceholder="选择学部"
          searchPlaceholder="搜索学生姓名、学号..."
          className="flex-col md:flex-row items-stretch md:items-center"
        />
        <div className="flex flex-wrap gap-4">
          <Select value={listPage.statusFilter} onValueChange={listPage.setStatusFilter}>
            <SelectTrigger className="w-32 rounded-lg">
              <SelectValue placeholder="状态" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部状态</SelectItem>
              <SelectItem value="active">活跃</SelectItem>
              <SelectItem value="inactive">停用</SelectItem>
              <SelectItem value="suspended">暂停</SelectItem>
            </SelectContent>
          </Select>

          <Select value={listPage.gradeFilter} onValueChange={listPage.setGradeFilter}>
            <SelectTrigger className="w-32 rounded-lg">
              <SelectValue placeholder="年级" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">全部年级</SelectItem>
              {gradeOptions.map((g) => (
                <SelectItem key={g.value} value={g.value}>
                  {g.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </FilterCard>

      {/* Students Table */}
      <Card className="rounded-xl shadow-lg">
        <CardHeader>
          <CardTitle>学生列表</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="border rounded-xl overflow-x-auto bg-white dark:bg-gray-900/60">
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
                  key: "student_id",
                  header: "学号",
                  render: (student: Student) => student.student_id || "-",
                },
                {
                  key: "college",
                  header: "学部",
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
              className="min-w-full"
            />
          </div>
        </CardContent>
      </Card>

      {/* 分页 */}
      {!listPage.loading && listPage.totalItems > 0 && (
        <Card className="rounded-xl shadow-lg">
          <CardContent className="pt-6">
            <Pagination
              currentPage={listPage.currentPage}
              totalPages={listPage.totalPages}
              totalItems={listPage.totalItems}
              pageSize={listPage.pageSize}
              onPageChange={handlePageChange}
              onPageSizeChange={handlePageSizeChange}
            />
          </CardContent>
        </Card>
      )}

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
                name="student_id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>学号</FormLabel>
                    <FormControl>
                      <Input {...field} />
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
                    <FormLabel>学部</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="选择学部" />
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
                render={({ field }) => {
                  const selectedCollege = userManagement.form.watch("college");
                  const availableMajors = selectedCollege
                    ? majorOptions[selectedCollege] || []
                    : [];
                  return (
                    <FormItem>
                      <FormLabel>专业</FormLabel>
                      <FormControl>
                        <Select
                          value={field.value || ""}
                          onValueChange={field.onChange}
                          disabled={!selectedCollege}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder={selectedCollege ? "请选择专业" : "请先选择学部"} />
                          </SelectTrigger>
                          <SelectContent>
                            {availableMajors.map((m) => (
                              <SelectItem key={m.value} value={m.value}>
                                {m.label}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  );
                }}
              />
              <FormField
                control={userManagement.form.control}
                name="class"
                render={({ field }) => {
                  const selectedMajor = userManagement.form.watch("major");
                  const availableClasses = selectedMajor
                    ? classOptions[selectedMajor] || []
                    : [];
                  return (
                    <FormItem>
                      <FormLabel>班级</FormLabel>
                      <FormControl>
                        <Select
                          value={field.value || ""}
                          onValueChange={field.onChange}
                          disabled={!selectedMajor}
                        >
                          <SelectTrigger>
                            <SelectValue placeholder={selectedMajor ? "请选择班级" : "请先选择专业"} />
                          </SelectTrigger>
                          <SelectContent>
                            {availableClasses.map((c) => (
                              <SelectItem key={c.value} value={c.value}>
                                {c.label}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  );
                }}
              />
              <FormField
                control={userManagement.form.control}
                name="grade"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>年级</FormLabel>
                    <FormControl>
                      <Select
                        value={field.value || ""}
                        onValueChange={field.onChange}
                      >
                        <SelectTrigger>
                          <SelectValue placeholder="请选择年级" />
                        </SelectTrigger>
                        <SelectContent>
                          {gradeOptions.map((g) => (
                            <SelectItem key={g.value} value={g.value}>
                              {g.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
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
