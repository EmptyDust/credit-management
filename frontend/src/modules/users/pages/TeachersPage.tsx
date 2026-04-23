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
import { exportTeachers, listTeachers } from "@/modules/users/api/users";

export default function TeachersPage() {
  const [keyword, setKeyword] = useState("");

  const query = useQuery({
    queryKey: ["teachers", keyword],
    queryFn: () => listTeachers({ keyword: keyword || undefined }),
  });

  return (
    <div className="space-y-6">
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
