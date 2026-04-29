import { useQuery } from "@tanstack/react-query";
import { Activity, FileText, Users } from "lucide-react";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { getDashboardOverview } from "@/modules/dashboard/api/dashboard";

const cards = [
  {
    key: "user_count",
    label: "用户总数",
    icon: Users,
  },
  {
    key: "activity_count",
    label: "活动总数",
    icon: Activity,
  },
  {
    key: "application_count",
    label: "申请总数",
    icon: FileText,
  },
] as const;

export default function DashboardPage() {
  const query = useQuery({
    queryKey: ["dashboard", "overview"],
    queryFn: getDashboardOverview,
  });

  return (
    <div className="space-y-6">
      <Card className="border-none bg-gradient-to-r from-sky-950 via-slate-900 to-slate-950 text-slate-50 shadow-xl">
        <CardHeader>
          <CardTitle className="text-2xl">仪表板</CardTitle>
          <p className="text-sm text-slate-300">
            这里展示重构后系统的关键汇总指标，作为后续统计模块的基础入口。
          </p>
        </CardHeader>
      </Card>

      <div className="grid gap-4 md:grid-cols-3">
        {cards.map((item) => (
          <Card key={item.key}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <CardTitle className="text-sm font-medium">{item.label}</CardTitle>
              <item.icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {query.isLoading ? (
                <div className="text-sm text-muted-foreground">加载中...</div>
              ) : query.isError ? (
                <div className="text-sm text-destructive">加载失败</div>
              ) : (
                <div className="text-3xl font-semibold">
                  {query.data?.[item.key] ?? 0}
                </div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
