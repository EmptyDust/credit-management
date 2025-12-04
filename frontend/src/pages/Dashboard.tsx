import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Progress } from "@/components/ui/progress";
import apiClient from "@/lib/api";
import {
  Users,
  UserCheck,
  School,
  Briefcase,
  FileText,
  GitPullRequest,
  Hourglass,
  Award,
  CheckCircle,
  XCircle,
  AlertCircle,
  RefreshCw,
  Activity,
  BarChart3,
  Target,
  BookOpen,
  Trophy,
  Zap,
  Lightbulb,
  Eye,
  Calendar,
} from "lucide-react";
import { Link } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import toast from "react-hot-toast";
import type { ActivityCategory } from "@/types/activity";
import { getActivityCategories } from "@/types/activity";
import { StatCard } from "@/components/ui/stat-card";
import { getStatusBadge } from "@/lib/status-utils";
import type { SelectOption } from "@/lib/options";
import { TopProgressBar } from "@/components/ui/top-progress-bar";

interface UserStats {
  total_users: number;
  student_users: number;
  teacher_users: number;
  admin_users: number;
  active_users: number;
  new_users_this_month: number;
}

interface ActivityStats {
  total_activities: number;
  active_activities: number;
  draft_activities: number;
  pending_activities: number;
  approved_activities: number;
  rejected_activities: number;
  recent_activities: Array<{
    id: string;
    title: string;
    description: string;
    category: ActivityCategory;
    status: string;
    participant_count: number;
    application_count: number;
    created_at: string;
    updated_at: string;
    start_date: string;
    end_date: string;
  }>;
  popular_activities: Array<{
    id: string;
    title: string;
    application_count: number;
    participant_count: number;
  }>;
  category_stats: {
    [key in ActivityCategory]: number;
  };
  total_participants: number;
  total_applications: number;
  total_credits_awarded: number;
  average_credits_per_activity: number;
}

interface RecentActivity {
  id: string;
  title: string;
  description: string;
  category: ActivityCategory;
  status: string;
  participant_count: number;
  application_count: number;
  created_at: string;
  updated_at: string;
  start_date: string;
  end_date: string;
  owner_name?: string;
}


const ActivityCard = ({ activity }: { activity: RecentActivity }) => {
  const getCategoryIcon = (category: ActivityCategory) => {
    // 动态图标映射 - 基于分类名称的关键词
    const categoryLower = category.toLowerCase();
    
    if (categoryLower.includes("创新") || categoryLower.includes("创业")) {
      return <Lightbulb className="h-4 w-4" />;
    } else if (categoryLower.includes("竞赛") || categoryLower.includes("比赛")) {
      return <Trophy className="h-4 w-4" />;
    } else if (categoryLower.includes("项目")) {
      return <Zap className="h-4 w-4" />;
    } else if (categoryLower.includes("实践")) {
      return <Briefcase className="h-4 w-4" />;
    } else if (categoryLower.includes("论文") || categoryLower.includes("专利")) {
      return <BookOpen className="h-4 w-4" />;
    } else {
      return <Activity className="h-4 w-4" />;
    }
  };

  const formatTime = (timestamp: string) => {
    try {
      const date = new Date(timestamp);
      if (isNaN(date.getTime())) {
        return "无效时间";
      }
      const now = new Date();
      const diffInHours = Math.floor(
        (now.getTime() - date.getTime()) / (1000 * 60 * 60)
      );

      if (diffInHours < 1) {
        return "刚刚";
      } else if (diffInHours < 24) {
        return `${diffInHours}小时前`;
      } else if (diffInHours < 24 * 7) {
        return `${Math.floor(diffInHours / 24)}天前`;
      } else {
        return date.toLocaleDateString();
      }
    } catch (error) {
      console.error("Time formatting error:", error);
      return "无效时间";
    }
  };

  const formatDate = (dateString: string) => {
    try {
      const date = new Date(dateString);
      if (isNaN(date.getTime())) {
        return "无效日期";
      }
      return date.toLocaleDateString("zh-CN", {
        month: "short",
        day: "numeric",
      });
    } catch (error) {
      console.error("Date formatting error:", error);
      return "无效日期";
    }
  };

  return (
    <div className="flex items-start space-x-3 p-3 rounded-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
      <div className="flex-shrink-0 p-2 rounded-lg bg-blue-100 dark:bg-blue-900/20">
        {getCategoryIcon(activity.category)}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between">
          <div className="flex-1 min-w-0">
            <h4 className="text-sm font-medium text-gray-900 dark:text-gray-100 truncate">
              {activity.title || "无标题"}
            </h4>
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1 line-clamp-2">
              {activity.description || "暂无描述"}
            </p>
          </div>
          <div className="flex-shrink-0 ml-2">
            {getStatusBadge(activity.status)}
          </div>
        </div>

        <div className="flex items-center justify-between mt-2">
          <div className="flex items-center space-x-4 text-xs text-gray-500 dark:text-gray-400">
            <div className="flex items-center space-x-1">
              <Calendar className="h-3 w-3" />
              <span>
                {formatDate(activity.start_date || "")} -{" "}
                {formatDate(activity.end_date || "")}
              </span>
            </div>
            <div className="flex items-center space-x-1">
              <Users className="h-3 w-3" />
              <span>{activity.participant_count || 0} 参与者</span>
            </div>
            <div className="flex items-center space-x-1">
              <FileText className="h-3 w-3" />
              <span>{activity.application_count || 0} 申请</span>
            </div>
          </div>
          <div className="text-xs text-gray-400">
            {formatTime(activity.updated_at || "")}
          </div>
        </div>
      </div>
      <Link to={`/activities/${activity.id}`}>
        <Button variant="ghost" size="sm" className="h-8 w-8 p-0">
          <Eye className="h-4 w-4" />
        </Button>
      </Link>
    </div>
  );
};

const CreditTypeCard = ({
  type,
  count,
  total,
  icon: Icon,
  color,
}: {
  type: string;
  count: number;
  total: number;
  icon: React.ElementType;
  color: string;
}) => {
  const percentage = total > 0 ? (count / total) * 100 : 0;

  return (
    <Card className="rounded-lg border-0 shadow-sm hover:shadow-md transition-shadow">
      <CardContent className="p-4">
        <div className="flex items-center justify-between mb-3">
          <div className={`p-2 rounded-lg ${color}`}>
            <Icon className="h-5 w-5 text-white" />
          </div>
          <span className="text-2xl font-bold">{count}</span>
        </div>
        <div className="space-y-2">
          <p className="text-sm font-medium">{type}</p>
          <Progress value={percentage} className="h-2" />
          <p className="text-xs text-muted-foreground">
            {percentage.toFixed(1)}% of total
          </p>
        </div>
      </CardContent>
    </Card>
  );
};

export default function Dashboard() {
  const { user, hasPermission } = useAuth();
  const [userStats, setUserStats] = useState<UserStats | null>(null);
  const [activityCategories, setActivityCategories] = useState<SelectOption[]>([]);
  const [activityStats, setActivityStats] = useState<ActivityStats>({
    total_activities: 0,
    active_activities: 0,
    draft_activities: 0,
    pending_activities: 0,
    approved_activities: 0,
    rejected_activities: 0,
    recent_activities: [],
    popular_activities: [],
    category_stats: {},
    total_participants: 0,
    total_applications: 0,
    total_credits_awarded: 0,
    average_credits_per_activity: 0,
  });
  const [creditTypeStats, setCreditTypeStats] = useState<Record<string, number>>({});
  const [recentActivities, setRecentActivities] = useState<RecentActivity[]>(
    []
  );
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const fetchDashboardData = async () => {
    try {
      setRefreshing(true);

      // Fetch user stats (admin only)
      if (hasPermission("view_user_stats") || user?.userType === "admin") {
        try {
          const userResponse = await apiClient.get("/users/stats");
          const userStatsData = userResponse.data.data || userResponse.data;
          if (userResponse.data.code === 0) {
            setUserStats(userStatsData);
          }
        } catch (error) {
          console.error("Failed to fetch user stats:", error);
          if (
            error instanceof Error &&
            error.message.includes("Network Error")
          ) {
            toast.error("网络连接失败，请检查网络连接");
            return;
          }
        }
      }

      // Fetch activity stats overview（与活动列表使用同一统计接口，保证“待审核”等数据一致）
      try {
        // 1) 先获取全局活动统计数据
        try {
          const statsResponse = await apiClient.get("/activities/stats");
          if (statsResponse.data.code === 0 && statsResponse.data.data) {
            const stats = statsResponse.data.data;
            setActivityStats((prev) => ({
              ...prev,
              total_activities: stats.total_activities ?? prev.total_activities,
              draft_activities: stats.draft_count ?? prev.draft_activities,
              pending_activities: stats.pending_count ?? prev.pending_activities,
              approved_activities: stats.approved_count ?? prev.approved_activities,
              rejected_activities: stats.rejected_count ?? prev.rejected_activities,
              total_participants:
                stats.total_participants ?? prev.total_participants,
              total_credits_awarded:
                stats.total_credits ?? prev.total_credits_awarded,
            }));
          }
        } catch (error) {
          console.error("Failed to fetch activity stats overview:", error);
        }

        // 2) 再拉取活动列表，用于最近活动、热门活动和按分类统计等
        const activitiesResponse = await apiClient.get("/activities", {
          // 不再按 status 过滤，避免导致待审核数量统计不准确
          params: { page: 1, page_size: 20 },
        });

        let activitiesData: any[] = [];
        if (activitiesResponse.data.code === 0 && activitiesResponse.data.data) {
          activitiesData = activitiesResponse.data.data.data || activitiesResponse.data.data.activities || [];
          setRecentActivities(activitiesData);
        }

        // Calculate category stats - 动态分类统计
        const categoryStats: Record<string, number> = {};

        activitiesData.forEach((activity: any) => {
          if (activity.category) {
            if (!categoryStats[activity.category]) {
              categoryStats[activity.category] = 0;
            }
            categoryStats[activity.category]++;
          }
        });

        // Calculate total participants and applications
        const totalApplications = activitiesData.reduce(
          (sum: number, activity: any) =>
            sum + (activity.applications?.length || 0),
          0
        );

        // Calculate total credits awarded
        const averageCreditsPerActivity =
          activitiesData.length > 0
            ? Math.round(
                (activitiesData.reduce(
                  (sum: number, activity: any) => {
                    const approvedApps =
                      activity.applications?.filter(
                        (app: any) => app.status === "approved"
                      ) || [];
                    return (
                      sum +
                      approvedApps.reduce(
                        (appSum: number, app: any) =>
                          appSum + (app.awarded_credits || 0),
                        0
                      )
                    );
                  },
                  0
                ) /
                  activitiesData.length) *
                  10
              ) / 10
            : 0;

        // Prepare recent activities for the card
        const recentActivitiesData = activitiesData
          .filter((activity: any) => activity && activity.id) // 过滤掉无效的活动
          .sort(
            (a: any, b: any) =>
              new Date(b.updated_at || 0).getTime() -
              new Date(a.updated_at || 0).getTime()
          )
          .slice(0, 6)
          .map((activity: any) => ({
            id: activity.id || "",
            title: activity.title || "无标题",
            description: activity.description || "暂无描述",
            category: activity.category || "未分类",
            status: activity.status || "draft",
            participant_count: activity.participants_count || 0,
            application_count: activity.applications_count || 0,
            created_at: activity.created_at || "",
            updated_at: activity.updated_at || "",
            start_date: activity.start_date || "",
            end_date: activity.end_date || "",
            owner_name: activity.owner?.name || activity.owner_name || "",
          }));

        // Prepare popular activities
        const popularActivities = activitiesData
          .filter((activity: any) => activity && activity.id) // 过滤掉无效的活动
          .sort(
            (a: any, b: any) =>
              (b.participants_count || 0) - (a.participants_count || 0)
          )
          .slice(0, 5)
          .map((activity: any) => ({
            id: activity.id || "",
            title: activity.title || "无标题",
            application_count: activity.applications_count || 0,
            participant_count: activity.participants_count || 0,
          }));

        // 这里不再覆盖从 /activities/stats 获取到的汇总统计，
        // 只更新需要列表数据支撑的字段，保证仪表盘与活动列表统计统一
        setActivityStats((prev) => ({
          ...prev,
          recent_activities: recentActivitiesData,
          popular_activities: popularActivities,
          category_stats: categoryStats,
          total_applications: totalApplications,
          average_credits_per_activity: averageCreditsPerActivity,
        }));

        // Set recent activities for the card
        setRecentActivities(recentActivitiesData);

        // Update credit type stats based on activities - 动态分类统计
        setCreditTypeStats(categoryStats);
      } catch (error) {
        console.error("Failed to fetch activities:", error);
        if (error instanceof Error && error.message.includes("Network Error")) {
          toast.error("网络连接失败，请检查网络连接");
          return;
        }
        toast.error("获取活动数据失败");
      }
    } catch (error) {
      console.error("Failed to fetch dashboard data:", error);
      toast.error("获取数据失败，请稍后重试");
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  };

  // 加载活动配置
  useEffect(() => {
    const loadActivityConfig = async () => {
      try {
        const categories = await getActivityCategories();
        setActivityCategories(categories);
      } catch (error) {
        console.error("Failed to load activity categories:", error);
        toast.error("加载活动配置失败");
      }
    };
    
    loadActivityConfig();
  }, []);

  useEffect(() => {
    fetchDashboardData();
  }, [hasPermission, user?.username]);

  const handleRefresh = () => {
    fetchDashboardData();
  };

  if (loading) {
    return (
      <div className="flex-1 space-y-4 p-4 md:p-8 bg-gradient-to-br from-gray-50 to-blue-50 dark:from-gray-900 dark:to-gray-800 min-h-screen">
        <TopProgressBar active={true} />
        <div className="flex justify-center items-center h-64">
          <div className="flex items-center gap-2">
            <Hourglass className="h-8 w-8 animate-spin" />
            <span>加载中...</span>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 space-y-8 p-4 md:p-8 bg-gradient-to-br from-gray-50 to-blue-50 dark:from-gray-900 dark:to-gray-800 min-h-screen">
      <TopProgressBar active={refreshing} />
      {/* Header */}
      <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-4">
        <div className="space-y-2">
          <h1 className="text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
            仪表板
          </h1>
          <p className="text-muted-foreground text-lg">
            欢迎回来，{user?.fullName || user?.username}！这里是系统概览。
          </p>
        </div>
        <Button
          onClick={handleRefresh}
          disabled={refreshing}
          variant="outline"
          className="rounded-lg shadow-lg hover:shadow-xl transition-all duration-200"
        >
          <RefreshCw
            className={`h-4 w-4 mr-2 ${refreshing ? "animate-spin" : ""}`}
          />
          刷新数据
        </Button>
      </div>

      {/* User Statistics - Admin Only */}
      {hasPermission("view_user_stats") && (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          <StatCard
            title="总用户数"
            value={userStats?.total_users ?? "N/A"}
            icon={Users}
            to="/students"
            description="所有注册用户"
            trend={{ value: 12, isPositive: true }}
            color="info"
            loading={refreshing}
          />
          <StatCard
            title="学生用户"
            value={userStats?.student_users ?? "N/A"}
            icon={School}
            to="/students"
            description="学生账户"
            color="success"
            loading={refreshing}
          />
          <StatCard
            title="教师用户"
            value={userStats?.teacher_users ?? "N/A"}
            icon={Briefcase}
            to="/teachers"
            description="教师账户"
            color="warning"
            loading={refreshing}
          />
          <StatCard
            title="管理员"
            value={userStats?.admin_users ?? "N/A"}
            icon={UserCheck}
            description="系统管理员"
            color="danger"
            loading={refreshing}
          />
        </div>
      )}

      {/* Activity Statistics - All data from activities */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-5">
        <StatCard
          title="总活动数"
          value={activityStats.total_activities}
          icon={Award}
          to="/activities"
          description="所有活动"
          loading={refreshing}
        />
        <StatCard
          title="待审核"
          value={activityStats.pending_activities}
          icon={GitPullRequest}
          to="/activities?status=pending_review"
          description="等待审核"
          color="warning"
          loading={refreshing}
        />
        <StatCard
          title="已通过"
          value={activityStats.approved_activities}
          icon={CheckCircle}
          to="/activities?status=approved"
          description="审核通过"
          color="success"
          loading={refreshing}
        />
        <StatCard
          title="已拒绝"
          value={activityStats.rejected_activities}
          icon={XCircle}
          to="/activities?status=rejected"
          description="审核拒绝"
          color="danger"
          loading={refreshing}
        />
        <StatCard
          title="申请数量"
          value={activityStats.total_applications}
          icon={Target}
          to="/applications"
          description="总申请数"
          color="info"
          loading={refreshing}
        />
      </div>

      {/* Activities and Recent Activities */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* Activities Statistics */}
        <Card className="rounded-xl shadow-lg border-0 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-xl">
              <Award className="h-6 w-6 text-purple-600 dark:text-purple-400" />
              活动统计
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid grid-cols-2 gap-4">
              <div className="text-center p-4 rounded-lg bg-gradient-to-br from-blue-50 to-blue-100 dark:from-blue-900/20 dark:to-blue-800/20">
                <div className="text-3xl font-bold text-blue-600">
                  {activityStats.total_activities}
                </div>
                <div className="text-sm text-muted-foreground">总活动数</div>
              </div>
              <div className="text-center p-4 rounded-lg bg-gradient-to-br from-green-50 to-green-100 dark:from-green-900/20 dark:to-green-800/20">
                <div className="text-3xl font-bold text-green-600">
                  {activityStats.total_participants}
                </div>
                <div className="text-sm text-muted-foreground">总参与者</div>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="text-center p-4 rounded-lg bg-gradient-to-br from-yellow-50 to-yellow-100 dark:from-yellow-900/20 dark:to-yellow-800/20">
                <div className="text-3xl font-bold text-yellow-600 dark:text-yellow-400">
                  {activityStats.total_applications}
                </div>
                <div className="text-sm text-muted-foreground">总申请数</div>
              </div>
              <div className="text-center p-4 rounded-lg bg-gradient-to-br from-purple-50 to-purple-100 dark:from-purple-900/20 dark:to-purple-800/20">
                <div className="text-3xl font-bold text-purple-600 dark:text-purple-400">
                  {activityStats.total_credits_awarded}
                </div>
                <div className="text-sm text-muted-foreground">总授予学分</div>
              </div>
            </div>

            {activityStats.popular_activities.length > 0 && (
              <div>
                <h4 className="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300">
                  热门活动
                </h4>
                <div className="space-y-3">
                  {activityStats.popular_activities
                    .slice(0, 3)
                    .map((activity) => (
                      <div
                        key={activity.id}
                        className="flex justify-between items-center p-3 rounded-lg bg-gray-50 dark:bg-gray-800/50"
                      >
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium truncate">
                            {activity.title}
                          </p>
                          <p className="text-xs text-muted-foreground">
                            {activity.participant_count} 参与者
                          </p>
                        </div>
                        <Link to={`/activities/${activity.id}`}>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 w-8 p-0"
                          >
                            <Eye className="h-4 w-4" />
                          </Button>
                        </Link>
                      </div>
                    ))}
                </div>
              </div>
            )}

            <div className="pt-4">
              <Link
                to="/activities"
                className="text-sm text-primary hover:underline font-medium"
              >
                查看所有活动 →
              </Link>
            </div>
          </CardContent>
        </Card>

        {/* Recent Activities */}
        <Card className="rounded-xl shadow-lg border-0 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-xl">
              <Activity className="h-6 w-6 text-blue-600" />
              最近活动
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 max-h-[600px] overflow-y-auto">
            {recentActivities.length > 0 ? (
              recentActivities.map((activity) => (
                <ActivityCard key={activity.id} activity={activity} />
              ))
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                <AlertCircle className="h-8 w-8 mx-auto mb-2" />
                <p>{refreshing ? "加载中..." : "暂无活动记录"}</p>
                {!refreshing && (
                  <p className="text-xs mt-1">请创建一些活动来查看最近活动</p>
                )}
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Credit Statistics - All data from activities */}
      <Card className="rounded-xl shadow-lg border-0 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-xl">
            <BarChart3 className="h-6 w-6 text-green-600" />
            学分统计
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-6 md:grid-cols-4 mb-8">
            <div className="text-center p-6 rounded-xl bg-gradient-to-br from-green-50 to-green-100 dark:from-green-900/20 dark:to-green-800/20">
              <div className="text-3xl font-bold text-green-600">
                {activityStats.total_credits_awarded}
              </div>
              <div className="text-sm text-muted-foreground">总授予学分</div>
            </div>
            <div className="text-center p-6 rounded-xl bg-gradient-to-br from-blue-50 to-blue-100 dark:from-blue-900/20 dark:to-blue-800/20">
              <div className="text-3xl font-bold text-blue-600">
                {activityStats.average_credits_per_activity}
              </div>
              <div className="text-sm text-muted-foreground">平均学分/活动</div>
            </div>
            <div className="text-center p-6 rounded-xl bg-gradient-to-br from-purple-50 to-purple-100 dark:from-purple-900/20 dark:to-purple-800/20">
              <div className="text-3xl font-bold text-purple-600 dark:text-purple-400">
                {activityStats.total_applications}
              </div>
              <div className="text-sm text-muted-foreground">总申请数</div>
            </div>
            <div className="text-center p-6 rounded-xl bg-gradient-to-br from-orange-50 to-orange-100 dark:from-orange-900/20 dark:to-orange-800/20">
              <div className="text-3xl font-bold text-orange-600">
                {activityStats.total_participants}
              </div>
              <div className="text-sm text-muted-foreground">总参与者</div>
            </div>
          </div>

          {/* Credit Types Distribution */}
          <div>
            <h4 className="text-lg font-medium mb-4 text-gray-700 dark:text-gray-300">
              活动类型分布
            </h4>
            <div className="grid gap-4 md:grid-cols-5">
              {activityCategories.map((category, index) => {
                const getCategoryIcon = (categoryName: string) => {
                  const categoryLower = categoryName.toLowerCase();
                  if (categoryLower.includes("创新") || categoryLower.includes("创业")) {
                    return Lightbulb;
                  } else if (categoryLower.includes("竞赛") || categoryLower.includes("比赛")) {
                    return Trophy;
                  } else if (categoryLower.includes("项目")) {
                    return Zap;
                  } else if (categoryLower.includes("实践")) {
                    return Briefcase;
                  } else if (categoryLower.includes("论文") || categoryLower.includes("专利")) {
                    return BookOpen;
                  } else {
                    return Activity;
                  }
                };

                const getCategoryColor = (index: number) => {
                  const colors = [
                    "bg-blue-500",
                    "bg-yellow-500", 
                    "bg-green-500",
                    "bg-purple-500",
                    "bg-red-500",
                    "bg-indigo-500",
                    "bg-pink-500",
                    "bg-orange-500"
                  ];
                  return colors[index % colors.length];
                };

                const IconComponent = getCategoryIcon(category.label);
                const totalCount = Object.values(creditTypeStats).reduce((a, b) => a + b, 0);

                return (
                  <CreditTypeCard
                    key={category.value}
                    type={category.label}
                    count={creditTypeStats[category.value] || 0}
                    total={totalCount}
                    icon={IconComponent}
                    color={getCategoryColor(index)}
                  />
                );
              })}
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
