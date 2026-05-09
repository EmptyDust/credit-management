import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Download, RefreshCw, Upload } from "lucide-react";
import { useState } from "react";
import toast from "react-hot-toast";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { UserImportDialog } from "@/modules/imports/components/UserImportDialog";
import {
  createTeacher,
  deleteTeacher,
  exportTeachers,
  listTeachers,
  resetUserPassword,
  updateTeacher,
  updateUserStatus,
  type AdminUser,
  type UserMutationInput,
} from "@/modules/users/api/users";

interface TeacherFormState {
  username: string;
  password: string;
  real_name: string;
  teacher_id: string;
  title: string;
}

const initialTeacherForm: TeacherFormState = {
  username: "",
  password: "",
  real_name: "",
  teacher_id: "",
  title: "",
};

interface TeacherEditState {
  id: string;
  username: string;
  real_name: string;
  teacher_id: string;
  title: string;
}

function toTeacherPayload(form: TeacherFormState): UserMutationInput {
  const payload: UserMutationInput = {};

  if (form.username.trim()) payload.username = form.username.trim();
  if (form.password.trim()) payload.password = form.password.trim();
  if (form.real_name.trim()) payload.real_name = form.real_name.trim();
  if (form.teacher_id.trim()) payload.teacher_id = form.teacher_id.trim();
  if (form.title.trim()) payload.title = form.title.trim();

  return payload;
}

function toTeacherEditPayload(form: TeacherEditState): UserMutationInput {
  const payload: UserMutationInput = {};

  if (form.username.trim()) payload.username = form.username.trim();
  if (form.real_name.trim()) payload.real_name = form.real_name.trim();
  if (form.teacher_id.trim()) payload.teacher_id = form.teacher_id.trim();
  if (form.title.trim()) payload.title = form.title.trim();

  return payload;
}

export default function TeachersPage() {
  const queryClient = useQueryClient();
  const [keyword, setKeyword] = useState("");
  const [importOpen, setImportOpen] = useState(false);
  const [teacherForm, setTeacherForm] = useState<TeacherFormState>(initialTeacherForm);
  const [editingTeacher, setEditingTeacher] = useState<TeacherEditState | null>(null);
  const [newPassword, setNewPassword] = useState("");

  const query = useQuery({
    queryKey: ["teachers", keyword],
    queryFn: () => listTeachers({ keyword: keyword || undefined }),
  });

  const invalidateTeachers = () =>
    queryClient.invalidateQueries({ queryKey: ["teachers"] });

  const createMutation = useMutation({
    mutationFn: (input: UserMutationInput) => createTeacher(input),
    onSuccess: () => {
      setTeacherForm(initialTeacherForm);
      void invalidateTeachers();
      toast.success("教师已创建");
    },
    onError: () => {
      toast.error("创建教师失败");
    },
  });

  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      updateUserStatus(id, status),
    onSuccess: () => {
      void invalidateTeachers();
      toast.success("教师状态已更新");
    },
    onError: () => {
      toast.error("更新教师状态失败");
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: UserMutationInput }) =>
      updateTeacher(id, input),
    onSuccess: () => {
      setEditingTeacher(null);
      void invalidateTeachers();
      toast.success("教师已更新");
    },
    onError: () => {
      toast.error("更新教师失败");
    },
  });

  const resetPasswordMutation = useMutation({
    mutationFn: ({ id, password }: { id: string; password: string }) =>
      resetUserPassword(id, password),
    onSuccess: () => {
      setNewPassword("");
      toast.success("教师密码已重置");
    },
    onError: () => {
      toast.error("重置教师密码失败");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteTeacher(id),
    onSuccess: () => {
      void invalidateTeachers();
      toast.success("教师已删除");
    },
    onError: () => {
      toast.error("删除教师失败");
    },
  });

  const updateTeacherForm = (field: keyof TeacherFormState, value: string) => {
    setTeacherForm((current) => ({ ...current, [field]: value }));
  };

  const updateEditingTeacher = (field: keyof TeacherEditState, value: string) => {
    setEditingTeacher((current) => current && { ...current, [field]: value });
  };

  const handleCreateTeacher = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    createMutation.mutate(toTeacherPayload(teacherForm));
  };

  const startEditingTeacher = (teacher: AdminUser) => {
    setEditingTeacher({
      id: teacher.id,
      username: teacher.username,
      real_name: teacher.real_name,
      teacher_id: teacher.teacher_id ?? "",
      title: teacher.title ?? "",
    });
  };

  const handleUpdateTeacher = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!editingTeacher) return;
    updateMutation.mutate({
      id: editingTeacher.id,
      input: toTeacherEditPayload(editingTeacher),
    });
  };

  const handleResetPassword = (teacherId: string) => {
    const password = newPassword.trim();
    if (!password) {
      toast.error("请填写新密码");
      return;
    }
    resetPasswordMutation.mutate({ id: teacherId, password });
  };

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>创建教师</CardTitle>
          <p className="mt-2 text-sm text-muted-foreground">
            管理员可直接创建教师账号；导入仍用于批量场景。
          </p>
        </CardHeader>
        <CardContent>
          <form className="grid gap-4 md:grid-cols-6" onSubmit={handleCreateTeacher}>
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="teacher-username">教师用户名</Label>
              <Input
                id="teacher-username"
                value={teacherForm.username}
                onChange={(event) => updateTeacherForm("username", event.target.value)}
              />
            </div>
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="teacher-password">教师密码</Label>
              <Input
                id="teacher-password"
                type="password"
                value={teacherForm.password}
                onChange={(event) => updateTeacherForm("password", event.target.value)}
              />
            </div>
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="teacher-real-name">教师姓名</Label>
              <Input
                id="teacher-real-name"
                value={teacherForm.real_name}
                onChange={(event) => updateTeacherForm("real_name", event.target.value)}
              />
            </div>
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="teacher-number">教师工号</Label>
              <Input
                id="teacher-number"
                value={teacherForm.teacher_id}
                onChange={(event) => updateTeacherForm("teacher_id", event.target.value)}
              />
            </div>
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="teacher-title">教师职称</Label>
              <Input
                id="teacher-title"
                value={teacherForm.title}
                onChange={(event) => updateTeacherForm("title", event.target.value)}
              />
            </div>
            <div className="flex items-end md:col-span-2">
              <Button type="submit" disabled={createMutation.isPending}>
                创建教师
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {editingTeacher ? (
        <Card>
          <CardHeader>
            <CardTitle>编辑教师</CardTitle>
            <p className="mt-2 text-sm text-muted-foreground">
              更新教师基础资料，密码请使用单独的重置入口。
            </p>
          </CardHeader>
          <CardContent>
            <form className="grid gap-4 md:grid-cols-6" onSubmit={handleUpdateTeacher}>
              <div className="space-y-2 md:col-span-2">
                <Label htmlFor="edit-teacher-username">编辑教师用户名</Label>
                <Input
                  id="edit-teacher-username"
                  value={editingTeacher.username}
                  onChange={(event) => updateEditingTeacher("username", event.target.value)}
                />
              </div>
              <div className="space-y-2 md:col-span-2">
                <Label htmlFor="edit-teacher-real-name">编辑教师姓名</Label>
                <Input
                  id="edit-teacher-real-name"
                  value={editingTeacher.real_name}
                  onChange={(event) => updateEditingTeacher("real_name", event.target.value)}
                />
              </div>
              <div className="space-y-2 md:col-span-2">
                <Label htmlFor="edit-teacher-number">编辑教师工号</Label>
                <Input
                  id="edit-teacher-number"
                  value={editingTeacher.teacher_id}
                  onChange={(event) => updateEditingTeacher("teacher_id", event.target.value)}
                />
              </div>
              <div className="space-y-2 md:col-span-2">
                <Label htmlFor="edit-teacher-title">编辑教师职称</Label>
                <Input
                  id="edit-teacher-title"
                  value={editingTeacher.title}
                  onChange={(event) => updateEditingTeacher("title", event.target.value)}
                />
              </div>
              <div className="flex items-end gap-2 md:col-span-2">
                <Button type="submit" disabled={updateMutation.isPending}>
                  保存教师
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setEditingTeacher(null)}
                >
                  取消
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      ) : null}

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>教师列表</CardTitle>
            <p className="mt-2 text-sm text-muted-foreground">
              基于新 `/api/v1/teachers` 契约的教师管理基础页。
            </p>
          </div>
          <div className="flex items-center gap-3">
            <Input
              className="w-72"
              placeholder="搜索用户名或姓名"
              value={keyword}
              onChange={(event) => setKeyword(event.target.value)}
            />
            <Button variant="secondary" className="gap-2" onClick={() => query.refetch()}>
              <RefreshCw className="h-4 w-4" />
              刷新
            </Button>
            <Button variant="outline" className="gap-2" onClick={() => setImportOpen(true)}>
              <Upload className="h-4 w-4" />
              导入
            </Button>
            <Button variant="outline" className="gap-2" onClick={() => void exportTeachers()}>
              <Download className="h-4 w-4" />
              导出
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {query.isLoading ? (
            <div className="py-10 text-sm text-muted-foreground">正在加载教师数据...</div>
          ) : query.isError ? (
            <div className="py-10 text-sm text-destructive">教师列表加载失败。</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>姓名</TableHead>
                  <TableHead>用户名</TableHead>
                  <TableHead>工号</TableHead>
                  <TableHead>职称</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {query.data?.items.map((teacher) => (
                  <TableRow key={teacher.id}>
                    <TableCell>{teacher.real_name}</TableCell>
                    <TableCell>{teacher.username}</TableCell>
                    <TableCell>{teacher.teacher_id || "-"}</TableCell>
                    <TableCell>{teacher.title || "-"}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{teacher.status}</Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          aria-label={`编辑 ${teacher.username}`}
                          onClick={() => startEditingTeacher(teacher)}
                        >
                          编辑
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          aria-label={`${teacher.status === "active" ? "停用" : "启用"} ${teacher.username}`}
                          disabled={statusMutation.isPending}
                          onClick={() =>
                            statusMutation.mutate({
                              id: teacher.id,
                              status: teacher.status === "active" ? "disabled" : "active",
                            })
                          }
                        >
                          {teacher.status === "active" ? "停用" : "启用"}
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          aria-label={`重置 ${teacher.username} 密码`}
                          disabled={resetPasswordMutation.isPending}
                          onClick={() => handleResetPassword(teacher.id)}
                        >
                          重置密码
                        </Button>
                        <Button
                          variant="destructive"
                          size="sm"
                          aria-label={`删除 ${teacher.username}`}
                          disabled={deleteMutation.isPending}
                          onClick={() => deleteMutation.mutate(teacher.id)}
                        >
                          删除
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>重置密码</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="max-w-sm space-y-2">
            <Label htmlFor="teacher-new-password">新密码</Label>
            <Input
              id="teacher-new-password"
              type="password"
              value={newPassword}
              onChange={(event) => setNewPassword(event.target.value)}
            />
          </div>
        </CardContent>
      </Card>

      <UserImportDialog
        open={importOpen}
        onOpenChange={setImportOpen}
        userType="teacher"
        onCompleted={() => void query.refetch()}
      />
    </div>
  );
}
