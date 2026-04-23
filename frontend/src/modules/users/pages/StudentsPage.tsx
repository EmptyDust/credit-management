import { useQuery } from "@tanstack/react-query";
import { Download, RefreshCw } from "lucide-react";
import { useState } from "react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { exportStudents, listStudents } from "@/modules/users/api/users";

export default function StudentsPage() {
  const [keyword, setKeyword] = useState("");

  const query = useQuery({
    queryKey: ["students", keyword],
    queryFn: () => listStudents({ keyword: keyword || undefined }),
  });

  return (
    <div className="space-y-6">
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
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
