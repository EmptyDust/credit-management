import { useQuery } from "@tanstack/react-query";
import {
  Activity,
  FileText,
  Layers3,
  Sparkles,
  Trophy,
  Users,
} from "lucide-react";
import type { ReactNode } from "react";
import { Link } from "react-router-dom";

import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  type ActivitySummary,
  type KeyLabelCount,
  getDashboardOverview,
} from "@/modules/dashboard/api/dashboard";

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

const statusLabels: Record<string, string> = {
  draft: "草稿",
  pending_review: "审核中",
  approved: "已通过",
  rejected: "已拒绝",
};

export default function DashboardPage() {
  const query = useQuery({
    queryKey: ["dashboard", "overview"],
    queryFn: getDashboardOverview,
  });

  const overview = query.data;
  const statusCounts = overview?.activity_stats?.status_counts ?? {};
  const categoryDistribution = overview?.category_distribution ?? [];

  return (
    <div className="space-y-6">
      <Card className="border-none bg-gradient-to-r from-sky-950 via-slate-900 to-slate-950 text-slate-50 shadow-xl">
        <CardHeader>
          <CardTitle className="text-2xl">仪表板</CardTitle>
          <p className="text-sm text-slate-300">
            按当前登录用户可见范围聚合用户、活动、申请与参与热度。
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

      {query.isError ? (
        <Card>
          <CardContent className="py-8 text-sm text-destructive">
            仪表板数据加载失败
          </CardContent>
        </Card>
      ) : null}

      <div className="grid gap-4 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Users className="h-4 w-4" />
              学生 / 教师 / 管理员
            </CardTitle>
          </CardHeader>
          <CardContent className="grid grid-cols-3 gap-3">
            <MetricBlock label="学生" value={overview?.user_stats?.student_count} />
            <MetricBlock label="教师" value={overview?.user_stats?.teacher_count} />
            <MetricBlock label="管理员" value={overview?.user_stats?.admin_count} />
            <MetricBlock label="启用" value={overview?.user_stats?.active_count} />
            <MetricBlock label="禁用" value={overview?.user_stats?.disabled_count} />
            <MetricBlock label="总计" value={overview?.user_stats?.total} />
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Activity className="h-4 w-4" />
              活动状态
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {Object.entries(statusLabels).map(([status, label]) => (
              <Link
                key={status}
                to={`/activities?status=${status}`}
                className="flex items-center justify-between rounded-lg border px-3 py-2 text-sm transition hover:bg-muted"
              >
                <span>{label}</span>
                <Badge variant="secondary">{statusCounts[status] ?? 0}</Badge>
              </Link>
            ))}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <FileText className="h-4 w-4" />
              申请统计
            </CardTitle>
          </CardHeader>
          <CardContent className="grid grid-cols-3 gap-3">
            <MetricBlock label="申请" value={overview?.application_stats?.total} />
            <MetricBlock
              label="已通过"
              value={overview?.application_stats?.approved_count}
            />
            <div className="rounded-xl border bg-muted/30 p-3">
              <div className="text-xs text-muted-foreground">总学分</div>
              <div className="mt-1 text-2xl font-semibold">
                {overview?.application_stats?.total_credits ?? 0}
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 lg:grid-cols-[1fr_1.4fr]">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-base">
              <Layers3 className="h-4 w-4" />
              分类分布
            </CardTitle>
          </CardHeader>
          <CardContent>
            <DistributionList items={categoryDistribution} />
          </CardContent>
        </Card>

        <div className="grid gap-4 xl:grid-cols-2">
          <ActivityListCard
            title="最近活动"
            icon={<Sparkles className="h-4 w-4" />}
            activities={overview?.recent_activities ?? []}
            emptyText="暂无最近活动"
          />
          <ActivityListCard
            title="热门活动"
            icon={<Trophy className="h-4 w-4" />}
            activities={overview?.popular_activities ?? []}
            emptyText="暂无热门活动"
            showParticipants
          />
        </div>
      </div>
    </div>
  );
}

function MetricBlock({ label, value }: { label: string; value?: number }) {
  return (
    <div className="rounded-xl border bg-muted/30 p-3">
      <div className="text-xs text-muted-foreground">{label}</div>
      <div className="mt-1 text-2xl font-semibold">{value ?? 0}</div>
    </div>
  );
}

function DistributionList({ items }: { items: KeyLabelCount[] }) {
  if (items.length === 0) {
    return <div className="text-sm text-muted-foreground">暂无分类数据</div>;
  }

  const max = Math.max(...items.map((item) => item.count), 1);

  return (
    <div className="space-y-4">
      {items.map((item) => (
        <div key={item.key} className="space-y-2">
          <div className="flex items-center justify-between text-sm">
            <span className="font-medium">{item.label || item.key}</span>
            <span className="text-muted-foreground">{item.count}</span>
          </div>
          <div className="h-2 overflow-hidden rounded-full bg-muted">
            <div
              className="h-full rounded-full bg-sky-600"
              style={{ width: `${Math.max((item.count / max) * 100, 6)}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  );
}

function ActivityListCard({
  title,
  icon,
  activities,
  emptyText,
  showParticipants = false,
}: {
  title: string;
  icon: ReactNode;
  activities: ActivitySummary[];
  emptyText: string;
  showParticipants?: boolean;
}) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-base">
          {icon}
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {activities.length === 0 ? (
          <div className="text-sm text-muted-foreground">{emptyText}</div>
        ) : (
          activities.map((activity) => (
            <Link
              key={activity.id}
              to={`/activities/${activity.id}`}
              className="block rounded-xl border p-3 transition hover:bg-muted"
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <div className="font-medium">{activity.title}</div>
                  <div className="mt-1 text-xs text-muted-foreground">
                    {activity.academic_year} · {activity.semester} ·{" "}
                    {activity.category}
                  </div>
                </div>
                <Badge variant="outline">
                  {statusLabels[activity.status] ?? activity.status}
                </Badge>
              </div>
              {showParticipants ? (
                <div className="mt-2 text-xs text-muted-foreground">
                  参与者 {activity.participant_count}
                </div>
              ) : null}
            </Link>
          ))
        )}
      </CardContent>
    </Card>
  );
}
