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
  createStudent,
  deleteStudent,
  exportStudents,
  getStudentStats,
  listStudents,
  resetUserPassword,
  updateStudent,
  updateUserStatus,
  type AdminUser,
  type UserMutationInput,
} from "@/modules/users/api/users";

interface StudentFormState {
  username: string;
  password: string;
  real_name: string;
  student_id: string;
  grade: string;
}

const initialStudentForm: StudentFormState = {
  username: "",
  password: "",
  real_name: "",
  student_id: "",
  grade: "",
};

const pageSize = 20;

interface StudentEditState {
  id: string;
  username: string;
  real_name: string;
  student_id: string;
  grade: string;
}

function toStudentPayload(form: StudentFormState): UserMutationInput {
  const payload: UserMutationInput = {};

  if (form.username.trim()) payload.username = form.username.trim();
  if (form.password.trim()) payload.password = form.password.trim();
  if (form.real_name.trim()) payload.real_name = form.real_name.trim();
  if (form.student_id.trim()) payload.student_id = form.student_id.trim();
  if (form.grade.trim()) payload.grade = form.grade.trim();

  return payload;
}

function toStudentEditPayload(form: StudentEditState): UserMutationInput {
  const payload: UserMutationInput = {};

  if (form.username.trim()) payload.username = form.username.trim();
  if (form.real_name.trim()) payload.real_name = form.real_name.trim();
  if (form.student_id.trim()) payload.student_id = form.student_id.trim();
  if (form.grade.trim()) payload.grade = form.grade.trim();

  return payload;
}

export default function StudentsPage() {
  const queryClient = useQueryClient();
  const [keyword, setKeyword] = useState("");
  const [page, setPage] = useState(1);
  const [importOpen, setImportOpen] = useState(false);
  const [studentForm, setStudentForm] = useState<StudentFormState>(initialStudentForm);
  const [editingStudent, setEditingStudent] = useState<StudentEditState | null>(null);
  const [newPassword, setNewPassword] = useState("");

  const query = useQuery({
    queryKey: ["students", "list", keyword, page],
    queryFn: () =>
      listStudents({
        keyword: keyword || undefined,
        page,
        page_size: pageSize,
      }),
  });

  const statsQuery = useQuery({
    queryKey: ["students", "stats", keyword],
    queryFn: () => getStudentStats({ keyword: keyword || undefined }),
  });

  const invalidateStudents = () =>
    queryClient.invalidateQueries({ queryKey: ["students"] });

  const createMutation = useMutation({
    mutationFn: (input: UserMutationInput) => createStudent(input),
    onSuccess: () => {
      setStudentForm(initialStudentForm);
      void invalidateStudents();
      toast.success("学生已创建");
    },
    onError: () => {
      toast.error("创建学生失败");
    },
  });

  const statusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      updateUserStatus(id, status),
    onSuccess: () => {
      void invalidateStudents();
      toast.success("学生状态已更新");
    },
    onError: () => {
      toast.error("更新学生状态失败");
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: UserMutationInput }) =>
      updateStudent(id, input),
    onSuccess: () => {
      setEditingStudent(null);
      void invalidateStudents();
      toast.success("学生已更新");
    },
    onError: () => {
      toast.error("更新学生失败");
    },
  });

  const resetPasswordMutation = useMutation({
    mutationFn: ({ id, password }: { id: string; password: string }) =>
      resetUserPassword(id, password),
    onSuccess: () => {
      setNewPassword("");
      toast.success("学生密码已重置");
    },
    onError: () => {
      toast.error("重置学生密码失败");
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteStudent(id),
    onSuccess: () => {
      void invalidateStudents();
      toast.success("学生已删除");
    },
    onError: () => {
      toast.error("删除学生失败");
    },
  });

  const updateStudentForm = (field: keyof StudentFormState, value: string) => {
    setStudentForm((current) => ({ ...current, [field]: value }));
  };

  const updateEditingStudent = (field: keyof StudentEditState, value: string) => {
    setEditingStudent((current) => current && { ...current, [field]: value });
  };

  const handleCreateStudent = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    createMutation.mutate(toStudentPayload(studentForm));
  };

  const startEditingStudent = (student: AdminUser) => {
    setEditingStudent({
      id: student.id,
      username: student.username,
      real_name: student.real_name,
      student_id: student.student_id ?? "",
      grade: student.grade ?? "",
    });
  };

  const handleUpdateStudent = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!editingStudent) return;
    updateMutation.mutate({
      id: editingStudent.id,
      input: toStudentEditPayload(editingStudent),
    });
  };

  const handleResetPassword = (studentId: string) => {
    const password = newPassword.trim();
    if (!password) {
      toast.error("请填写新密码");
      return;
    }
    resetPasswordMutation.mutate({ id: studentId, password });
  };

  return (
    <div className="space-y-6">
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              总学生数
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold">{statsQuery.data?.total ?? 0}</div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              活跃学生
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold">
              {statsQuery.data?.status_counts?.active ?? 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              停用学生
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-semibold">
              {statsQuery.data?.status_counts?.disabled ?? 0}
            </div>
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle className="text-sm font-medium text-muted-foreground">
              年级分布
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm text-muted-foreground">
              {(statsQuery.data?.grade_counts ?? [])
                .map((item) => `${item.label}: ${item.count}`)
                .join(" / ") || "-"}
            </div>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>创建学生</CardTitle>
          <p className="mt-2 text-sm text-muted-foreground">
            管理员与教师可直接创建学生账号；导入仍用于批量场景。
          </p>
        </CardHeader>
        <CardContent>
          <form className="grid gap-4 md:grid-cols-6" onSubmit={handleCreateStudent}>
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="student-username">学生用户名</Label>
              <Input
                id="student-username"
                value={studentForm.username}
                onChange={(event) => updateStudentForm("username", event.target.value)}
              />
            </div>
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="student-password">学生密码</Label>
              <Input
                id="student-password"
                type="password"
                value={studentForm.password}
                onChange={(event) => updateStudentForm("password", event.target.value)}
              />
            </div>
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="student-real-name">学生姓名</Label>
              <Input
                id="student-real-name"
                value={studentForm.real_name}
                onChange={(event) => updateStudentForm("real_name", event.target.value)}
              />
            </div>
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="student-number">学生学号</Label>
              <Input
                id="student-number"
                value={studentForm.student_id}
                onChange={(event) => updateStudentForm("student_id", event.target.value)}
              />
            </div>
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="student-grade">学生年级</Label>
              <Input
                id="student-grade"
                value={studentForm.grade}
                onChange={(event) => updateStudentForm("grade", event.target.value)}
              />
            </div>
            <div className="flex items-end md:col-span-2">
              <Button type="submit" disabled={createMutation.isPending}>
                创建学生
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {editingStudent ? (
        <Card>
          <CardHeader>
            <CardTitle>编辑学生</CardTitle>
            <p className="mt-2 text-sm text-muted-foreground">
              更新学生基础资料，密码请使用单独的重置入口。
            </p>
          </CardHeader>
          <CardContent>
            <form className="grid gap-4 md:grid-cols-6" onSubmit={handleUpdateStudent}>
              <div className="space-y-2 md:col-span-2">
                <Label htmlFor="edit-student-username">编辑学生用户名</Label>
                <Input
                  id="edit-student-username"
                  value={editingStudent.username}
                  onChange={(event) => updateEditingStudent("username", event.target.value)}
                />
              </div>
              <div className="space-y-2 md:col-span-2">
                <Label htmlFor="edit-student-real-name">编辑学生姓名</Label>
                <Input
                  id="edit-student-real-name"
                  value={editingStudent.real_name}
                  onChange={(event) => updateEditingStudent("real_name", event.target.value)}
                />
              </div>
              <div className="space-y-2 md:col-span-2">
                <Label htmlFor="edit-student-number">编辑学生学号</Label>
                <Input
                  id="edit-student-number"
                  value={editingStudent.student_id}
                  onChange={(event) => updateEditingStudent("student_id", event.target.value)}
                />
              </div>
              <div className="space-y-2 md:col-span-2">
                <Label htmlFor="edit-student-grade">编辑学生年级</Label>
                <Input
                  id="edit-student-grade"
                  value={editingStudent.grade}
                  onChange={(event) => updateEditingStudent("grade", event.target.value)}
                />
              </div>
              <div className="flex items-end gap-2 md:col-span-2">
                <Button type="submit" disabled={updateMutation.isPending}>
                  保存学生
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setEditingStudent(null)}
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
            <CardTitle>学生列表</CardTitle>
            <p className="mt-2 text-sm text-muted-foreground">
              基于新 `/api/v1/students` 契约的学生管理基础页。
            </p>
          </div>
          <div className="flex items-center gap-3">
            <Input
              className="w-72"
              placeholder="搜索用户名或姓名"
              value={keyword}
              onChange={(event) => {
                setKeyword(event.target.value);
                setPage(1);
              }}
            />
            <Button variant="secondary" className="gap-2" onClick={() => query.refetch()}>
              <RefreshCw className="h-4 w-4" />
              刷新
            </Button>
            <Button variant="outline" className="gap-2" onClick={() => setImportOpen(true)}>
              <Upload className="h-4 w-4" />
              导入
            </Button>
            <Button variant="outline" className="gap-2" onClick={() => void exportStudents()}>
              <Download className="h-4 w-4" />
              导出
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {query.isLoading ? (
            <div className="py-10 text-sm text-muted-foreground">正在加载学生数据...</div>
          ) : query.isError ? (
            <div className="py-10 text-sm text-destructive">学生列表加载失败。</div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>姓名</TableHead>
                  <TableHead>用户名</TableHead>
                  <TableHead>学号</TableHead>
                  <TableHead>年级</TableHead>
                  <TableHead>状态</TableHead>
                  <TableHead>操作</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {query.data?.items.map((student) => (
                  <TableRow key={student.id}>
                    <TableCell>{student.real_name}</TableCell>
                    <TableCell>{student.username}</TableCell>
                    <TableCell>{student.student_id || "-"}</TableCell>
                    <TableCell>{student.grade || "-"}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{student.status}</Badge>
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          aria-label={`编辑 ${student.username}`}
                          onClick={() => startEditingStudent(student)}
                        >
                          编辑
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          aria-label={`${student.status === "active" ? "停用" : "启用"} ${student.username}`}
                          disabled={statusMutation.isPending}
                          onClick={() =>
                            statusMutation.mutate({
                              id: student.id,
                              status: student.status === "active" ? "disabled" : "active",
                            })
                          }
                        >
                          {student.status === "active" ? "停用" : "启用"}
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          aria-label={`重置 ${student.username} 密码`}
                          disabled={resetPasswordMutation.isPending}
                          onClick={() => handleResetPassword(student.id)}
                        >
                          重置密码
                        </Button>
                        <Button
                          variant="destructive"
                          size="sm"
                          aria-label={`删除 ${student.username}`}
                          disabled={deleteMutation.isPending}
                          onClick={() => deleteMutation.mutate(student.id)}
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
          <div className="mt-4 flex items-center justify-between text-sm text-muted-foreground">
            <span>
              第 {query.data?.page ?? page} / {query.data?.total_pages ?? 0} 页，共{" "}
              {query.data?.total ?? 0} 条
            </span>
            <div className="flex items-center gap-2">
              <Button
                variant="outline"
                size="sm"
                disabled={(query.data?.page ?? page) <= 1}
                onClick={() => setPage((current) => Math.max(1, current - 1))}
              >
                上一页
              </Button>
              <Button
                variant="outline"
                size="sm"
                disabled={(query.data?.page ?? page) >= (query.data?.total_pages ?? 0)}
                onClick={() => setPage((current) => current + 1)}
              >
                下一页
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>重置密码</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="max-w-sm space-y-2">
            <Label htmlFor="student-new-password">新密码</Label>
            <Input
              id="student-new-password"
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
        userType="student"
        onCompleted={() => void query.refetch()}
      />
    </div>
  );
}
