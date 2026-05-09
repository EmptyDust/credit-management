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
  listStudents,
  resetUserPassword,
  updateUserStatus,
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

function toStudentPayload(form: StudentFormState): UserMutationInput {
  const payload: UserMutationInput = {};

  if (form.username.trim()) payload.username = form.username.trim();
  if (form.password.trim()) payload.password = form.password.trim();
  if (form.real_name.trim()) payload.real_name = form.real_name.trim();
  if (form.student_id.trim()) payload.student_id = form.student_id.trim();
  if (form.grade.trim()) payload.grade = form.grade.trim();

  return payload;
}

export default function StudentsPage() {
  const queryClient = useQueryClient();
  const [keyword, setKeyword] = useState("");
  const [importOpen, setImportOpen] = useState(false);
  const [studentForm, setStudentForm] = useState<StudentFormState>(initialStudentForm);
  const [newPassword, setNewPassword] = useState("");

  const query = useQuery({
    queryKey: ["students", keyword],
    queryFn: () => listStudents({ keyword: keyword || undefined }),
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

  const handleCreateStudent = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    createMutation.mutate(toStudentPayload(studentForm));
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
