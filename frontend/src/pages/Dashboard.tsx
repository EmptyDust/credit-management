import { useEffect, useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
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
  TrendingUp,
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
  MapPin,
} from "lucide-react";
import { Link } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import toast from "react-hot-toast";
import type { ActivityWithDetails, ActivityCategory } from "@/types/activity";

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

interface CreditTypeStats {
  innovation_practice: number;
  discipline_competition: number;
  entrepreneurship_project: number;
  entrepreneurship_practice: number;
  paper_patent: number;
}

const StatCard = ({
  title,
  value,
  icon: Icon,
  to,
  description,
  trend,
  color = "default",
  subtitle,
}: {
  title: string;
  value: string | number;
  icon: React.ElementType;
  to?: string;
  description?: string;
  trend?: { value: number; isPositive: boolean };
  color?: "default" | "success" | "warning" | "danger" | "info" | "purple";
  loading?: boolean;
  subtitle?: string;
}) => {
  const colorClasses = {
    default: "text-muted-foreground",
    success: "text-green-600",
    warning: "text-yellow-600",
    danger: "text-red-600",
    info: "text-blue-600",
    purple: "text-purple-600",
  };

  const bgClasses = {
    default: "bg-muted/20",
    success: "bg-green-100 dark:bg-green-900/20",
    warning: "bg-yellow-100 dark:bg-yellow-900/20",
    danger: "bg-red-100 dark:bg-red-900/20",
    info: "bg-blue-100 dark:bg-blue-900/20",
    purple: "bg-purple-100 dark:bg-purple-900/20",
  };

  const content = (
    <Card className="rounded-xl shadow-lg hover:shadow-2xl transition-all duration-300 bg-gradient-to-br from-white to-gray-50 dark:from-gray-900 dark:to-gray-800 border-0">
      <CardHeader className="flex flex-row items-center justify-between pb-3">
        <div className={`p-3 rounded-xl ${bgClasses[color]}`}>
          <Icon className={`h-6 w-6 ${colorClasses[color]}`} />
        </div>
        <CardTitle className="text-lg font-semibold text-gray-900 dark:text-gray-100">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-bold mb-1 text-gray-900 dark:text-gray-100">
          {value}
        </div>
        {subtitle && (
          <div className="text-sm text-muted-foreground mb-2">{subtitle}</div>
        )}
        {description && (
          <p className="text-xs text-muted-foreground">{description}</p>
        )}
        {trend && (
          <div className="flex items-center mt-3">
            <TrendingUp
              className={`h-4 w-4 mr-1 ${
                trend.isPositive ? "text-green-600" : "text-red-600"
              }`}
            />
            <span
              className={`text-xs font-medium ${
                trend.isPositive ? "text-green-600" : "text-red-600"
              }`}
            >
              {trend.isPositive ? "+" : ""}
              {trend.value}%
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  );

  return to ? (
    <Link to={to} className="block">
      {content}
    </Link>
  ) : (
    content
  );
};

const ActivityCard = ({ activity }: { activity: RecentActivity }) => {
  const getCategoryIcon = (category: ActivityCategory) => {
    switch (category) {
      case "创新创业实践活动":
        return <Lightbulb className="h-4 w-4" />;
      case "学科竞赛":
        return <Trophy className="h-4 w-4" />;
      case "大学生创业项目":
        return <Zap className="h-4 w-4" />;
      case "创业实践项目":
        return <Briefcase className="h-4 w-4" />;
      case "论文专利":
        return <BookOpen className="h-4 w-4" />;
      default:
        return <Activity className="h-4 w-4" />;
    }
  };

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "approved":
        return (
          <Badge
            variant="default"
            className="bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200"
          >
            已通过
          </Badge>
        );
      case "pending_review":
        return (
          <Badge
            variant="default"
            className="bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200"
          >
            待审核
          </Badge>
        );
      case "rejected":
        return (
          <Badge
            variant="default"
            className="bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200"
          >
            已拒绝
          </Badge>
        );
      case "draft":
        return (
          <Badge
            variant="default"
            className="bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200"
          >
            草稿
          </Badge>
        );
      default:
        return <Badge variant="outline">{status}</Badge>;
    }
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
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
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString("zh-CN", {
      month: "short",
      day: "numeric",
    });
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
              {activity.title}
            </h4>
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-1 line-clamp-2">
              {activity.description}
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
                {formatDate(activity.start_date)} -{" "}
                {formatDate(activity.end_date)}
              </span>
            </div>
            <div className="flex items-center space-x-1">
              <Users className="h-3 w-3" />
              <span>{activity.participant_count} 参与者</span>
            </div>
            <div className="flex items-center space-x-1">
              <FileText className="h-3 w-3" />
              <span>{activity.application_count} 申请</span>
            </div>
          </div>
          <div className="text-xs text-gray-400">
            {formatTime(activity.updated_at)}
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
  const [activityStats, setActivityStats] = useState<ActivityStats>({
    total_activities: 0,
    active_activities: 0,
    draft_activities: 0,
    pending_activities: 0,
    approved_activities: 0,
    rejected_activities: 0,
    recent_activities: [],
    popular_activities: [],
    category_stats: {
      创新创业实践活动: 0,
      学科竞赛: 0,
      大学生创业项目: 0,
      创业实践项目: 0,
      论文专利: 0,
    },
    total_participants: 0,
    total_applications: 0,
    total_credits_awarded: 0,
    average_credits_per_activity: 0,
  });
  const [creditTypeStats, setCreditTypeStats] = useState<CreditTypeStats>({
    innovation_practice: 0,
    discipline_competition: 0,
    entrepreneurship_project: 0,
    entrepreneurship_practice: 0,
    paper_patent: 0,
  });
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

      // Fetch activities to calculate all stats
      try {
        const activitiesResponse = await apiClient.get("/activities");
        console.log("Dashboard activities response:", activitiesResponse.data);

        let activities = [];
        if (
          activitiesResponse.data.data &&
          Array.isArray(activitiesResponse.data.data)
        ) {
          activities = activitiesResponse.data.data;
        } else if (
          activitiesResponse.data.data &&
          Array.isArray(activitiesResponse.data.data.data)
        ) {
          activities = activitiesResponse.data.data.data;
        } else if (Array.isArray(activitiesResponse.data)) {
          activities = activitiesResponse.data;
        } else if (
          activitiesResponse.data.activities &&
          Array.isArray(activitiesResponse.data.activities)
        ) {
          activities = activitiesResponse.data.activities;
        } else {
          console.warn(
            "Unexpected activities response structure:",
            activitiesResponse.data
          );
          activities = [];
        }

        // Calculate comprehensive activity stats
        const totalActivities = activities.length;
        const activeActivities = activities.filter(
          (activity: any) => activity.status === "approved"
        ).length;
        const draftActivities = activities.filter(
          (activity: any) => activity.status === "draft"
        ).length;
        const pendingActivities = activities.filter(
          (activity: any) => activity.status === "pending_review"
        ).length;
        const approvedActivities = activities.filter(
          (activity: any) => activity.status === "approved"
        ).length;
        const rejectedActivities = activities.filter(
          (activity: any) => activity.status === "rejected"
        ).length;

        // Calculate category stats
        const categoryStats = {
          创新创业实践活动: 0,
          学科竞赛: 0,
          大学生创业项目: 0,
          创业实践项目: 0,
          论文专利: 0,
        };

        activities.forEach((activity: any) => {
          if (
            activity.category &&
            categoryStats.hasOwnProperty(activity.category)
          ) {
            categoryStats[activity.category as ActivityCategory]++;
          }
        });

        // Calculate total participants and applications
        const totalParticipants = activities.reduce(
          (sum: number, activity: any) =>
            sum + (activity.participants?.length || 0),
          0
        );
        const totalApplications = activities.reduce(
          (sum: number, activity: any) =>
            sum + (activity.applications?.length || 0),
          0
        );

        // Calculate total credits awarded
        const totalCreditsAwarded = activities.reduce(
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
        );

        const averageCreditsPerActivity =
          totalActivities > 0
            ? Math.round((totalCreditsAwarded / totalActivities) * 10) / 10
            : 0;

        // Prepare recent activities for the card
        const recentActivitiesData = activities
          .sort(
            (a: any, b: any) =>
              new Date(b.updated_at).getTime() -
              new Date(a.updated_at).getTime()
          )
          .slice(0, 10)
          .map((activity: any) => ({
            id: activity.id,
            title: activity.title,
            description: activity.description,
            category: activity.category,
            status: activity.status,
            participant_count: activity.participants?.length || 0,
            application_count: activity.applications?.length || 0,
            created_at: activity.created_at,
            updated_at: activity.updated_at,
            start_date: activity.start_date,
            end_date: activity.end_date,
            owner_name: activity.owner?.name || activity.owner_name,
          }));

        // Prepare popular activities
        const popularActivities = activities
          .sort(
            (a: any, b: any) =>
              (b.participants?.length || 0) - (a.participants?.length || 0)
          )
          .slice(0, 5)
          .map((activity: any) => ({
            id: activity.id,
            title: activity.title,
            application_count: activity.applications?.length || 0,
            participant_count: activity.participants?.length || 0,
          }));

        setActivityStats({
          total_activities: totalActivities,
          active_activities: activeActivities,
          draft_activities: draftActivities,
          pending_activities: pendingActivities,
          approved_activities: approvedActivities,
          rejected_activities: rejectedActivities,
          recent_activities: recentActivitiesData,
          popular_activities: popularActivities,
          category_stats: categoryStats,
          total_participants: totalParticipants,
          total_applications: totalApplications,
          total_credits_awarded: totalCreditsAwarded,
          average_credits_per_activity: averageCreditsPerActivity,
        });

        // Set recent activities for the card
        setRecentActivities(recentActivitiesData);

        // Update credit type stats based on activities
        setCreditTypeStats({
          innovation_practice: categoryStats["创新创业实践活动"],
          discipline_competition: categoryStats["学科竞赛"],
          entrepreneurship_project: categoryStats["大学生创业项目"],
          entrepreneurship_practice: categoryStats["创业实践项目"],
          paper_patent: categoryStats["论文专利"],
        });
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

  useEffect(() => {
    fetchDashboardData();
  }, [hasPermission, user?.username]);

  const handleRefresh = () => {
    fetchDashboardData();
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="flex items-center gap-2">
          <Hourglass className="h-8 w-8 animate-spin" />
          <span>加载中...</span>
        </div>
      </div>
    );
  }

  return (
    <div className="flex-1 space-y-8 p-4 md:p-8 bg-gradient-to-br from-gray-50 to-blue-50 dark:from-gray-900 dark:to-gray-800 min-h-screen">
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
              <Award className="h-6 w-6 text-purple-600" />
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
                <div className="text-3xl font-bold text-yellow-600">
                  {activityStats.total_applications}
                </div>
                <div className="text-sm text-muted-foreground">总申请数</div>
              </div>
              <div className="text-center p-4 rounded-lg bg-gradient-to-br from-purple-50 to-purple-100 dark:from-purple-900/20 dark:to-purple-800/20">
                <div className="text-3xl font-bold text-purple-600">
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
          <CardContent className="space-y-2 h-96 overflow-y-auto">
            {recentActivities.length > 0 ? (
              recentActivities.map((activity) => (
                <ActivityCard key={activity.id} activity={activity} />
              ))
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                <AlertCircle className="h-8 w-8 mx-auto mb-2" />
                <p>暂无活动记录</p>
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
              <div className="text-3xl font-bold text-purple-600">
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
              <CreditTypeCard
                type="创新创业实践"
                count={creditTypeStats.innovation_practice}
                total={Object.values(creditTypeStats).reduce(
                  (a, b) => a + b,
                  0
                )}
                icon={Lightbulb}
                color="bg-blue-500"
              />
              <CreditTypeCard
                type="学科竞赛"
                count={creditTypeStats.discipline_competition}
                total={Object.values(creditTypeStats).reduce(
                  (a, b) => a + b,
                  0
                )}
                icon={Trophy}
                color="bg-yellow-500"
              />
              <CreditTypeCard
                type="创业项目"
                count={creditTypeStats.entrepreneurship_project}
                total={Object.values(creditTypeStats).reduce(
                  (a, b) => a + b,
                  0
                )}
                icon={Zap}
                color="bg-green-500"
              />
              <CreditTypeCard
                type="创业实践"
                count={creditTypeStats.entrepreneurship_practice}
                total={Object.values(creditTypeStats).reduce(
                  (a, b) => a + b,
                  0
                )}
                icon={Briefcase}
                color="bg-purple-500"
              />
              <CreditTypeCard
                type="论文专利"
                count={creditTypeStats.paper_patent}
                total={Object.values(creditTypeStats).reduce(
                  (a, b) => a + b,
                  0
                )}
                icon={BookOpen}
                color="bg-red-500"
              />
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
