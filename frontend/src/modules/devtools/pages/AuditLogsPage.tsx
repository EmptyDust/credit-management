import { useQuery } from "@tanstack/react-query";
import { Download } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { exportAuditLogs, getAuditLogs } from "@/modules/devtools/api/devtools";

export default function AuditLogsPage() {
  const query = useQuery({
    queryKey: ["devtools", "audit-logs"],
    queryFn: getAuditLogs,
  });

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle>审计日志</CardTitle>
        <Button variant="outline" className="gap-2" onClick={() => void exportAuditLogs()}>
          <Download className="h-4 w-4" />
          导出
        </Button>
      </CardHeader>
      <CardContent>
        {query.isLoading ? (
          <div className="text-sm text-muted-foreground">正在加载审计日志...</div>
        ) : query.data?.items.length ? (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>时间</TableHead>
                <TableHead>动作</TableHead>
                <TableHead>结果</TableHead>
                <TableHead>目标</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {query.data.items.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>{item.created_at}</TableCell>
                  <TableCell>{item.action}</TableCell>
                  <TableCell>{item.result || "-"}</TableCell>
                  <TableCell>
                    {item.target_type || "-"} / {item.target_id || "-"}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        ) : (
          <div className="text-sm text-muted-foreground">当前没有审计记录。</div>
        )}
      </CardContent>
    </Card>
  );
}
