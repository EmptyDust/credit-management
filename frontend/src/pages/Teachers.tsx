import { useState, useEffect } from "react";
import * as z from "zod";
import { useAuth } from "@/contexts/AuthContext";
import { useListPage } from "@/hooks/useListPage";
import { useUserManagement } from "@/hooks/useUserManagement";
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
import {
  PlusCircle,
  RefreshCw,
  Users,
  Building,
  AlertCircle,
  Upload,
  Download,
} from "lucide-react";
import { getStatusBadge } from "@/lib/status-utils";
import { DeleteConfirmDialog } from "@/components/ui/delete-confirm-dialog";
import { ImportDialog } from "@/components/ui/import-dialog";
import { getOptions } from "@/lib/options";
import { PageHeader, createPageActions } from "@/components/ui/page-header";
import { StatsGrid } from "@/components/ui/stats-grid";
import { TableActions, createEditDeleteActions } from "@/components/ui/table-actions";
import { SearchFilterBar } from "@/components/ui/search-filter-bar";

// Teacher type based on teacher.go
export type Teacher = {
  id?: string;
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
  department: z.string().min(1, "学院不能为空"),
  title: z.string().optional().or(z.literal("")),
  status: z.enum(["active", "inactive", "suspended"]),
  user_type: z.literal("teacher"),
});

export default function TeachersPage() {
  const { hasPermission } = useAuth();
  const [teachers, setTeachers] = useState<Teacher[]>([]);
  const [userStatuses, setUserStatuses] = useState<{ value: string; label: string }[]>([]);
  const [teacherTitles, setTeacherTitles] = useState<{ value: string; label: string }[]>([]);
  const [collegeOptions, setCollegeOptions] = useState<{ value: string; label: string }[]>([]);

  // 使用新的通用列表页面hook
  const listPage = useListPage({
    endpoint: "/search/users",
    setData: setTeachers,
    errorMessage: "获取教师列表失败",
    userType: "teacher"
  });

  // 使用用户管理hook
  const userManagement = useUserManagement({
    userType: "teacher",
    formSchema,
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
    fetchFunction: listPage.fetchList,
  });

  useEffect(() => {
    // 初始化数据
    listPage.fetchList();
    
    // 加载选项数据
    (async () => {
      try {
        const opts = await getOptions();
        setCollegeOptions(opts.colleges || []);
        setUserStatuses(opts.user_statuses || []);
        setTeacherTitles(opts.teacher_titles || []);
      } catch (e) {
        console.error("Failed to load options", e);
      }
    })();
  }, []);

  // 处理搜索和筛选变化
  useEffect(() => {
    listPage.handleSearchAndFilter();
  }, [listPage.searchQuery, listPage.filterValue, listPage.statusFilter]);

  // 使用hook提供的handleDialogOpen
  const handleDialogOpen = userManagement.handleDialogOpen;

  // 使用hook提供的onSubmit
  const onSubmit = userManagement.onSubmit;

  const canManageTeachers = hasPermission("manage_teachers");

  // 使用hook提供的函数
  const handleDeleteConfirm = userManagement.handleDeleteConfirm;
  const handleImport = userManagement.handleImport;
  const handleExport = userManagement.handleExport;


  return (
    <div className="space-y-8 p-4 md:p-8 bg-background min-h-screen">
      <PageHeader
        title="教师列表"
        description="管理教师用户信息"
        actions={createPageActions(
          canManageTeachers ? {
            label: "添加教师",
            onClick: () => handleDialogOpen(null),
            icon: PlusCircle,
          } : undefined,
          canManageTeachers ? [
            {
              label: "批量导入",
              onClick: () => userManagement.setIsImportDialogOpen(true),
              icon: Upload,
            },
            {
              label: "导出数据",
              onClick: handleExport,
              icon: Download,
            },
          ] : undefined
        )}
      />

      {/* Statistics Cards */}
      <StatsGrid
        stats={[
          {
            title: "总教师数",
            value: teachers.length,
            icon: Users,
            color: "info",
            subtitle: `活跃教师: ${
              teachers.filter((t) => t.status === "active").length
            }`,
          },
          {
            title: "学院数量",
            value: collegeOptions.length,
            icon: Building,
            color: "purple",
            subtitle: "不同学院",
          },
          {
            title: "活跃教师数",
            value: teachers.filter((t) => t.status === "active").length,
            icon: Users,
            color: "success",
            subtitle: "当前活跃",
          },
        ]}
      />

      {/* 搜索和过滤栏 */}
      <div className="mb-6">
        <SearchFilterBar
          searchQuery={listPage.searchQuery}
          onSearchChange={listPage.setSearchQuery}
          onSearch={listPage.handleSearchAndFilter}
          onRefresh={() => listPage.fetchList()}
          filterOptions={collegeOptions}
          filterValue={listPage.filterValue}
          onFilterChange={listPage.setFilterValue}
          filterPlaceholder="选择学院"
          searchPlaceholder="搜索教师姓名、学院..."
        />
      </div>

      {/* 状态过滤 */}
      <div className="mb-4">
        <Select value={listPage.statusFilter} onValueChange={listPage.setStatusFilter}>
          <SelectTrigger className="w-32">
            <SelectValue placeholder="状态" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部状态</SelectItem>
            {userStatuses.map((s) => (
              <SelectItem key={s.value} value={s.value}>{s.label}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

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
                   <TableHead>学院</TableHead>
                  <TableHead>职称</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead className="text-right">操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {listPage.loading ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-8">
                      <div className="flex items-center justify-center gap-2">
                        <RefreshCw className="h-4 w-4 animate-spin" />
                        加载中...
                      </div>
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
                       <TableCell className="text-right">
                         {canManageTeachers && (
                           <TableActions
                             actions={createEditDeleteActions(
                               () => handleDialogOpen(teacher),
                               () => {
                                 userManagement.setItemToDelete(teacher);
                                 userManagement.setDeleteDialogOpen(true);
                               }
                             )}
                           />
                         )}
                       </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </div>

          {/* 分页组件 */}
          {!listPage.loading && listPage.totalItems > 0 && (
            <Pagination
              currentPage={listPage.currentPage}
              totalPages={listPage.totalPages}
              totalItems={listPage.totalItems}
              pageSize={listPage.pageSize}
              onPageChange={listPage.handlePageChange}
              onPageSizeChange={listPage.handlePageSizeChange}
            />
          )}
        </CardContent>
      </Card>

      {/* Create/Edit Dialog */}
      <Dialog open={userManagement.isDialogOpen} onOpenChange={userManagement.setIsDialogOpen}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>
              {userManagement.editingItem ? "编辑教师" : "添加新教师"}
            </DialogTitle>
            <DialogDescription>
              {userManagement.editingItem ? "修改教师信息" : "填写教师详细信息"}
            </DialogDescription>
          </DialogHeader>
          <Form {...userManagement.form}>
            <form
              onSubmit={userManagement.form.handleSubmit(onSubmit)}
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
                 name="department"
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
                name="title"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>职称</FormLabel>
                    <FormControl>
                      <Select onValueChange={field.onChange} defaultValue={field.value}>
                        <SelectTrigger>
                          <SelectValue placeholder="请选择职称" />
                        </SelectTrigger>
                        <SelectContent>
                          {teacherTitles.map((t) => (
                            <SelectItem key={t.value} value={t.value}>{t.label}</SelectItem>
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
                        {userStatuses.map((s) => (
                          <SelectItem key={s.value} value={s.value}>{s.label}</SelectItem>
                        ))}
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
        title="确认删除教师"
        itemName={userManagement.itemToDelete?.real_name}
        onConfirm={handleDeleteConfirm}
      />

      {/* Import Dialog */}
      <ImportDialog
        open={userManagement.isImportDialogOpen}
        onOpenChange={userManagement.setIsImportDialogOpen}
        title="批量导入教师"
        description="请选择Excel或CSV文件进行批量导入。文件应包含教师的基本信息。"
        userType="teacher"
        onImport={handleImport}
        importing={userManagement.importing}
      />
    </div>
  );
}
