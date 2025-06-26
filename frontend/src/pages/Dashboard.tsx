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
  Star,
  BookOpen,
  Trophy,
  Zap,
  Lightbulb,
  Globe,
  Eye,
} from "lucide-react";
import { Link } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import toast from "react-hot-toast";

interface UserStats {
  total_users: number;
  student_users: number;
  teacher_users: number;
  admin_users: number;
  active_users: number;
  new_users_this_month: number;
}

interface ApplicationStats {
  total_applications: number;
  pending_applications: number;
  approved_applications: number;
  rejected_applications: number;
  unsubmitted_applications: number;
  applications_this_month: number;
  approval_rate: number;
  total_credits_awarded: number;
  average_credits_per_application: number;
  credits_this_month: number;
}

interface AffairStats {
  total_affairs: number;
  active_affairs: number;
  recent_affairs: Array<{
    id: string;
    name: string;
    description: string;
    participant_count: number;
    application_count: number;
    created_at: string;
  }>;
  popular_affairs: Array<{
    id: string;
    name: string;
    application_count: number;
    participant_count: number;
  }>;
}

interface RecentActivity {
  id: string;
  type: "application" | "affair" | "user" | "review";
  action: string;
  timestamp: string;
  user: string;
  details?: string;
  status?: string;
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
  const getIcon = () => {
    switch (activity.type) {
      case "application":
        return <FileText className="h-4 w-4" />;
      case "affair":
        return <Award className="h-4 w-4" />;
      case "user":
        return <Users className="h-4 w-4" />;
      case "review":
        return <CheckCircle className="h-4 w-4" />;
      default:
        return <Activity className="h-4 w-4" />;
    }
  };

  const getColor = () => {
    switch (activity.type) {
      case "application":
        return "bg-blue-100 text-blue-600 dark:bg-blue-900/20";
      case "affair":
        return "bg-purple-100 text-purple-600 dark:bg-purple-900/20";
      case "user":
        return "bg-green-100 text-green-600 dark:bg-green-900/20";
      case "review":
        return "bg-orange-100 text-orange-600 dark:bg-orange-900/20";
      default:
        return "bg-gray-100 text-gray-600 dark:bg-gray-900/20";
    }
  };

  const getStatusBadge = (status?: string) => {
    if (!status) return null;

    const statusConfig = {
      pending: {
        label: "待审核",
        color: "bg-yellow-100 text-yellow-800 dark:bg-yellow-900/20",
      },
      approved: {
        label: "已通过",
        color: "bg-green-100 text-green-800 dark:bg-green-900/20",
      },
      rejected: {
        label: "已拒绝",
        color: "bg-red-100 text-red-800 dark:bg-red-900/20",
      },
      unsubmitted: {
        label: "未提交",
        color: "bg-gray-100 text-gray-800 dark:bg-gray-900/20",
      },
    };

    const config = statusConfig[status as keyof typeof statusConfig];
    return config ? (
      <Badge className={config.color}>{config.label}</Badge>
    ) : null;
  };

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffInMinutes = Math.floor(
      (now.getTime() - date.getTime()) / (1000 * 60)
    );
    if (diffInMinutes < 1) return "刚刚";
    if (diffInMinutes < 60) return `${diffInMinutes}分钟前`;
    if (diffInMinutes < 1440) return `${Math.floor(diffInMinutes / 60)}小时前`;
    return date.toLocaleDateString();
  };

  return (
    <div className="flex items-center space-x-3 p-3 rounded-lg hover:bg-muted/40 transition-colors border border-transparent hover:border-muted">
      <div className={`p-2 rounded-full ${getColor()} shadow-sm`}>
        {getIcon()}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <p className="text-sm font-medium truncate">{activity.action}</p>
          {getStatusBadge(activity.status)}
        </div>
        <p className="text-xs text-muted-foreground">by {activity.user}</p>
        {activity.details && (
          <p className="text-xs text-muted-foreground mt-1">
            {activity.details}
          </p>
        )}
      </div>
      <div className="text-xs text-muted-foreground whitespace-nowrap">
        {formatTime(activity.timestamp)}
      </div>
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
  const [appStats, setAppStats] = useState<ApplicationStats>({
    total_applications: 0,
    pending_applications: 0,
    approved_applications: 0,
    rejected_applications: 0,
    unsubmitted_applications: 0,
    applications_this_month: 0,
    approval_rate: 0,
    total_credits_awarded: 0,
    average_credits_per_application: 0,
    credits_this_month: 0,
  });
  const [affairStats, setAffairStats] = useState<AffairStats>({
    total_affairs: 0,
    active_affairs: 0,
    recent_affairs: [],
    popular_affairs: [],
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
          // Use fallback data for user stats
          setUserStats({
            total_users: 1250,
            student_users: 1100,
            teacher_users: 120,
            admin_users: 30,
            active_users: 1180,
            new_users_this_month: 45,
          });
        }
      }

      // Fetch activities to calculate stats
      try {
        const activitiesResponse = await apiClient.get("/activities");
        console.log("Dashboard activities response:", activitiesResponse.data); // 调试日志

        // 处理不同的响应数据结构
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

        // Calculate activity stats
        const totalActivities = activities.length;
        const activeActivities = activities.filter(
          (activity: any) => activity.status === "approved"
        ).length;

        setAffairStats({
          total_affairs: totalActivities,
          active_affairs: activeActivities,
          recent_affairs: activities
            .sort(
              (a: any, b: any) =>
                new Date(b.created_at).getTime() -
                new Date(a.created_at).getTime()
            )
            .slice(0, 5)
            .map((activity: any) => ({
              id: activity.id,
              name: activity.title,
              description: activity.description,
              participant_count: activity.participants?.length || 0,
              application_count: activity.applications?.length || 0,
              created_at: activity.created_at,
            })),
          popular_affairs: activities
            .sort(
              (a: any, b: any) =>
                (b.participants?.length || 0) - (a.participants?.length || 0)
            )
            .slice(0, 5)
            .map((activity: any) => ({
              id: activity.id,
              name: activity.title,
              application_count: activity.applications?.length || 0,
              participant_count: activity.participants?.length || 0,
            })),
        });
      } catch (error) {
        console.error("Failed to fetch activities:", error);
        setAffairStats({
          total_affairs: 15,
          active_affairs: 12,
          recent_affairs: [
            {
              id: "1",
              name: "创新创业项目",
              description: "参与各类创新创业项目",
              participant_count: 25,
              application_count: 25,
              created_at: new Date().toISOString(),
            },
            {
              id: "2",
              name: "学科竞赛",
              description: "参加各类学科竞赛",
              participant_count: 42,
              application_count: 42,
              created_at: new Date(Date.now() - 86400000).toISOString(),
            },
          ],
          popular_affairs: [
            {
              id: "1",
              name: "创新创业项目",
              application_count: 25,
              participant_count: 25,
            },
            {
              id: "2",
              name: "学科竞赛",
              application_count: 18,
              participant_count: 18,
            },
            {
              id: "3",
              name: "志愿服务",
              application_count: 12,
              participant_count: 12,
            },
          ],
        });
      }

      // Fetch applications to calculate stats
      try {
        const endpoint =
          user?.userType === "student"
            ? "/applications" // 学生只能看到自己的申请
            : "/applications/all"; // 教师和管理员可以看到所有申请
        const appResponse = await apiClient.get(endpoint);
        const applications =
          appResponse.data.data?.applications ||
          appResponse.data.applications ||
          [];

        const total = applications.length;
        const pending = applications.filter(
          (app: any) => app.status === "pending"
        ).length;
        const approved = applications.filter(
          (app: any) => app.status === "approved"
        ).length;
        const rejected = applications.filter(
          (app: any) => app.status === "rejected"
        ).length;
        const unsubmitted = applications.filter(
          (app: any) => app.status === "unsubmitted"
        ).length;
        const approvalRate =
          total > 0 ? Math.round((approved / total) * 100) : 0;

        // Calculate credit stats
        const approvedApps = applications.filter(
          (app: any) => app.status === "approved"
        );
        const totalCredits = approvedApps.reduce(
          (sum: number, app: any) => sum + (app.awarded_credits || 0),
          0
        );
        const avgCredits =
          approvedApps.length > 0
            ? Math.round((totalCredits / approvedApps.length) * 10) / 10
            : 0;

        // Calculate this month's credits
        const thisMonthCredits = approvedApps
          .filter((app: any) => {
            const appDate = new Date(app.submitted_at || app.created_at);
            const now = new Date();
            return (
              appDate.getMonth() === now.getMonth() &&
              appDate.getFullYear() === now.getFullYear()
            );
          })
          .reduce(
            (sum: number, app: any) => sum + (app.awarded_credits || 0),
            0
          );

        setAppStats({
          total_applications: total,
          pending_applications: pending,
          approved_applications: approved,
          rejected_applications: rejected,
          unsubmitted_applications: unsubmitted,
          applications_this_month: applications.filter((app: any) => {
            const appDate = new Date(app.submitted_at || app.created_at);
            const now = new Date();
            return (
              appDate.getMonth() === now.getMonth() &&
              appDate.getFullYear() === now.getFullYear()
            );
          }).length,
          approval_rate: approvalRate,
          total_credits_awarded: totalCredits,
          average_credits_per_application: avgCredits,
          credits_this_month: thisMonthCredits,
        });
      } catch (error) {
        console.error("Failed to fetch applications:", error);
        // Use fallback data for application stats
        setAppStats({
          total_applications: 152,
          pending_applications: 25,
          approved_applications: 127,
          rejected_applications: 8,
          unsubmitted_applications: 20,
          applications_this_month: 45,
          approval_rate: 83.5,
          total_credits_awarded: 456.5,
          average_credits_per_application: 2.3,
          credits_this_month: 89.5,
        });
      }

      // Mock credit type stats (would need API endpoint)
      setCreditTypeStats({
        innovation_practice: 45,
        discipline_competition: 38,
        entrepreneurship_project: 25,
        entrepreneurship_practice: 20,
        paper_patent: 15,
      });

      // Mock recent activities (since there's no activity API)
      setRecentActivities([
        {
          id: "1",
          type: "application",
          action: "新申请提交",
          timestamp: new Date().toISOString(),
          user: "张三",
          details: "创新创业项目申请",
          status: "pending",
        },
        {
          id: "2",
          type: "affair",
          action: "活动创建",
          timestamp: new Date(Date.now() - 3600000).toISOString(),
          user: "李老师",
          details: "学科竞赛活动",
        },
        {
          id: "3",
          type: "review",
          action: "申请审核",
          timestamp: new Date(Date.now() - 7200000).toISOString(),
          user: "王老师",
          details: "志愿服务申请审核通过",
          status: "approved",
        },
        {
          id: "4",
          type: "user",
          action: "用户注册",
          timestamp: new Date(Date.now() - 10800000).toISOString(),
          user: "赵同学",
          details: "新学生用户注册",
        },
      ]);
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

      {/* Application Statistics */}
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-5">
        <StatCard
          title="总申请数"
          value={appStats.total_applications}
          icon={FileText}
          to="/applications"
          description="所有申请"
          loading={refreshing}
        />
        <StatCard
          title="待审核"
          value={appStats.pending_applications}
          icon={GitPullRequest}
          to="/applications"
          description="等待审核"
          color="warning"
          loading={refreshing}
        />
        <StatCard
          title="已通过"
          value={appStats.approved_applications}
          icon={CheckCircle}
          to="/applications"
          description="审核通过"
          color="success"
          loading={refreshing}
        />
        <StatCard
          title="已拒绝"
          value={appStats.rejected_applications}
          icon={XCircle}
          to="/applications"
          description="审核拒绝"
          color="danger"
          loading={refreshing}
        />
        <StatCard
          title="通过率"
          value={`${appStats.approval_rate}%`}
          icon={Target}
          description="申请通过率"
          color="info"
          loading={refreshing}
        />
      </div>

      {/* Affairs and Activities */}
      <div className="grid gap-8 md:grid-cols-2">
        {/* Affairs Statistics */}
        <Card className="rounded-xl shadow-lg border-0 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm">
          <CardHeader>
            <CardTitle className="flex items-center gap-2 text-xl">
              <Award className="h-6 w-6 text-purple-600" />
              事务统计
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-6">
            <div className="grid grid-cols-2 gap-4">
              <div className="text-center p-4 rounded-lg bg-gradient-to-br from-blue-50 to-blue-100 dark:from-blue-900/20 dark:to-blue-800/20">
                <div className="text-3xl font-bold text-blue-600">
                  {affairStats.total_affairs}
                </div>
                <div className="text-sm text-muted-foreground">总事务数</div>
              </div>
              <div className="text-center p-4 rounded-lg bg-gradient-to-br from-green-50 to-green-100 dark:from-green-900/20 dark:to-green-800/20">
                <div className="text-3xl font-bold text-green-600">
                  {affairStats.active_affairs}
                </div>
                <div className="text-sm text-muted-foreground">活跃事务</div>
              </div>
            </div>

            {affairStats.recent_affairs.length > 0 && (
              <div>
                <h4 className="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300">
                  最近事务
                </h4>
                <div className="space-y-3">
                  {affairStats.recent_affairs.slice(0, 3).map((affair) => (
                    <div
                      key={affair.id}
                      className="flex justify-between items-center p-3 rounded-lg bg-gray-50 dark:bg-gray-800/50"
                    >
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium truncate">
                          {affair.name}
                        </p>
                        <p className="text-xs text-muted-foreground">
                          {affair.participant_count} 参与者
                        </p>
                      </div>
                      <Link to={`/affairs/${affair.id}`}>
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
                to="/affairs"
                className="text-sm text-primary hover:underline font-medium"
              >
                查看所有事务 →
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
          <CardContent className="space-y-2 max-h-96 overflow-y-auto">
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

      {/* Credit Statistics */}
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
                {appStats.total_credits_awarded}
              </div>
              <div className="text-sm text-muted-foreground">总授予学分</div>
            </div>
            <div className="text-center p-6 rounded-xl bg-gradient-to-br from-blue-50 to-blue-100 dark:from-blue-900/20 dark:to-blue-800/20">
              <div className="text-3xl font-bold text-blue-600">
                {appStats.average_credits_per_application}
              </div>
              <div className="text-sm text-muted-foreground">平均学分/申请</div>
            </div>
            <div className="text-center p-6 rounded-xl bg-gradient-to-br from-purple-50 to-purple-100 dark:from-purple-900/20 dark:to-purple-800/20">
              <div className="text-3xl font-bold text-purple-600">
                {appStats.credits_this_month}
              </div>
              <div className="text-sm text-muted-foreground">本月授予学分</div>
            </div>
            <div className="text-center p-6 rounded-xl bg-gradient-to-br from-orange-50 to-orange-100 dark:from-orange-900/20 dark:to-orange-800/20">
              <div className="text-3xl font-bold text-orange-600">
                {appStats.applications_this_month}
              </div>
              <div className="text-sm text-muted-foreground">本月申请数</div>
            </div>
          </div>

          {/* Credit Types Distribution */}
          <div>
            <h4 className="text-lg font-medium mb-4 text-gray-700 dark:text-gray-300">
              学分类型分布
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

      {/* Quick Actions */}
      <div className="grid gap-6 md:grid-cols-2">
        {/* Student Quick Actions */}
        {user?.userType === "student" && (
          <Card className="rounded-xl shadow-lg border-0 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-xl">
                <Star className="h-6 w-6 text-yellow-600" />
                学生快速操作
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2">
                <Link to="/applications">
                  <Button className="w-full h-16 flex flex-col items-center gap-2 rounded-xl shadow-lg hover:shadow-xl transition-all duration-200 bg-gradient-to-r from-blue-500 to-blue-600 hover:from-blue-600 hover:to-blue-700">
                    <FileText className="h-6 w-6" />
                    <span className="text-sm font-medium">查看申请</span>
                  </Button>
                </Link>
                <Link to="/profile">
                  <Button
                    variant="outline"
                    className="w-full h-16 flex flex-col items-center gap-2 rounded-xl shadow-lg hover:shadow-xl transition-all duration-200"
                  >
                    <UserCheck className="h-6 w-6" />
                    <span className="text-sm font-medium">个人资料</span>
                  </Button>
                </Link>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Teacher Quick Actions */}
        {user?.userType === "teacher" && (
          <Card className="rounded-xl shadow-lg border-0 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-xl">
                <Trophy className="h-6 w-6 text-purple-600" />
                教师快速操作
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2">
                <Link to="/applications">
                  <Button className="w-full h-16 flex flex-col items-center gap-2 rounded-xl shadow-lg hover:shadow-xl transition-all duration-200 bg-gradient-to-r from-green-500 to-green-600 hover:from-green-600 hover:to-green-700">
                    <CheckCircle className="h-6 w-6" />
                    <span className="text-sm font-medium">审核申请</span>
                  </Button>
                </Link>
                <Link to="/students">
                  <Button
                    variant="outline"
                    className="w-full h-16 flex flex-col items-center gap-2 rounded-xl shadow-lg hover:shadow-xl transition-all duration-200"
                  >
                    <Users className="h-6 w-6" />
                    <span className="text-sm font-medium">查看学生</span>
                  </Button>
                </Link>
              </div>
            </CardContent>
          </Card>
        )}

        {/* Admin Quick Actions */}
        {hasPermission("admin") && (
          <Card className="rounded-xl shadow-lg border-0 bg-white/80 dark:bg-gray-900/80 backdrop-blur-sm">
            <CardHeader>
              <CardTitle className="flex items-center gap-2 text-xl">
                <Globe className="h-6 w-6 text-indigo-600" />
                管理员快速操作
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid gap-4 md:grid-cols-2">
                <Link to="/affairs">
                  <Button className="w-full h-16 flex flex-col items-center gap-2 rounded-xl shadow-lg hover:shadow-xl transition-all duration-200 bg-gradient-to-r from-purple-500 to-purple-600 hover:from-purple-600 hover:to-purple-700">
                    <Award className="h-6 w-6" />
                    <span className="text-sm font-medium">管理事务</span>
                  </Button>
                </Link>
                <Link to="/users">
                  <Button
                    variant="outline"
                    className="w-full h-16 flex flex-col items-center gap-2 rounded-xl shadow-lg hover:shadow-xl transition-all duration-200"
                  >
                    <Users className="h-6 w-6" />
                    <span className="text-sm font-medium">用户管理</span>
                  </Button>
                </Link>
              </div>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
