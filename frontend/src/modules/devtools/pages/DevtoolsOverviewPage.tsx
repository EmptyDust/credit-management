import { useQuery } from "@tanstack/react-query";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getDevtoolsHealth } from "@/modules/devtools/api/devtools";

export default function DevtoolsOverviewPage() {
  const query = useQuery({
    queryKey: ["devtools", "health"],
    queryFn: getDevtoolsHealth,
  });

  const items = [
    ["数据库", query.data?.database.status || "unknown"],
    ["Redis", query.data?.redis.status || "unknown"],
    ["文件存储", query.data?.storage.status || "unknown"],
    ["版本", query.data?.version || "dev"],
  ];

  return (
    <Card>
      <CardHeader>
        <CardTitle>系统健康概览</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-4 md:grid-cols-2">
        {items.map(([label, value]) => (
          <div key={label} className="rounded-lg border p-4">
            <div className="text-sm text-muted-foreground">{label}</div>
            <div className="mt-2 text-xl font-semibold">
              {query.isLoading ? "加载中..." : value}
            </div>
          </div>
        ))}
      </CardContent>
    </Card>
  );
}
