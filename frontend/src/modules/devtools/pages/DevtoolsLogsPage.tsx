import { useQuery } from "@tanstack/react-query";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getDevtoolsLogs } from "@/modules/devtools/api/devtools";

export default function DevtoolsLogsPage() {
  const query = useQuery({
    queryKey: ["devtools", "logs"],
    queryFn: getDevtoolsLogs,
  });

  return (
    <Card>
      <CardHeader>
        <CardTitle>系统日志</CardTitle>
      </CardHeader>
      <CardContent>
        {query.isLoading ? (
          <div className="text-sm text-muted-foreground">正在加载日志...</div>
        ) : query.data?.items.length ? (
          <div className="space-y-3">
            {query.data.items.map((item, index) => (
              <div key={item.id || index} className="rounded-lg border p-3">
                <div className="text-xs text-muted-foreground">{item.created_at || "-"}</div>
                <div className="mt-1 font-medium">{item.level || "INFO"}</div>
                <div className="mt-1 text-sm">{item.message || "空日志消息"}</div>
              </div>
            ))}
          </div>
        ) : (
          <div className="text-sm text-muted-foreground">当前没有日志输出。</div>
        )}
      </CardContent>
    </Card>
  );
}
