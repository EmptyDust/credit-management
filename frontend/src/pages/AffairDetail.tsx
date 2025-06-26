import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Users,
  Award,
  FileText,
  ArrowLeft,
  User,
  Calendar,
  CheckCircle,
  XCircle,
  Clock,
  TrendingUp,
  FileCheck,
  AlertCircle,
  Star,
  Edit,
  Trash,
  Send,
  RotateCcw,
} from "lucide-react";
import apiClient from "@/lib/api";
import toast from "react-hot-toast";
import { useAuth } from "@/contexts/AuthContext";

interface Activity {
  id: string;
  title: string;
  description: string;
  category: string;
  status: string;
  requirements?: string;
  owner_id: string;
  reviewer_id?: string;
  review_comments?: string;
  reviewed_at?: string;
  start_date?: string;
  end_date?: string;
  created_at: string;
  updated_at: string;
  participants?: Participant[];
  applications?: Application[];
}

interface Participant {
  user_id: string;
  credits: number;
  joined_at: string;
  user_info?: {
    id: string;
    name: string;
    student_id: string;
  };
}

interface Application {
  id: string;
  activity_id: string;
  user_id: string;
  status: string;
  applied_credits: number;
  awarded_credits: number;
  submitted_at: string;
  created_at: string;
  updated_at: string;
  activity?: {
    id: string;
    title: string;
    description: string;
    category: string;
    start_date: string;
    end_date: string;
  };
}

export default function AffairDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [activity, setActivity] = useState<Activity | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!id) return;
    setLoading(true);

    apiClient
      .get(`/activities/${id}`)
      .then((response) => {
        const activityData = response.data.data || response.data;
        setActivity(activityData);
      })
      .catch((error) => {
        console.error("Failed to fetch activity:", error);
        toast.error("获取活动详情失败");
      })
      .finally(() => setLoading(false));
  }, [id]);

  const handleDelete = async () => {
    if (!activity) return;
    if (!window.confirm("确定要删除这个活动吗？此操作不可撤销。")) return;

    try {
      await apiClient.delete(`/activities/${activity.id}`);
      toast.success("活动删除成功");
      navigate("/affairs");
    } catch (err) {
      toast.error("删除活动失败");
    }
  };

  const handleSubmitForReview = async () => {
    if (!activity) return;

    try {
      await apiClient.post(`/activities/${activity.id}/submit`);
      toast.success("活动提交审核成功");
      // 重新获取活动数据
      const response = await apiClient.get(`/activities/${activity.id}`);
      const activityData = response.data.data || response.data;
      setActivity(activityData);
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || "提交审核失败";
      toast.error(errorMessage);
    }
  };

  const handleWithdraw = async () => {
    if (!activity) return;

    try {
      await apiClient.post(`/activities/${activity.id}/withdraw`);
      toast.success("活动撤回成功");
      // 重新获取活动数据
      const response = await apiClient.get(`/activities/${activity.id}`);
      const activityData = response.data.data || response.data;
      setActivity(activityData);
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || "撤回活动失败";
      toast.error(errorMessage);
    }
  };

  // 获取用户ID，兼容多种字段
  const userId =
    user?.id || (user as any)?.user_id || user?.username || user?.studentNumber;
  const isOwner = activity && user && activity.owner_id === userId;

  // 检查是否为教师或管理员
  const isTeacherOrAdmin =
    user && (user.userType === "teacher" || user.userType === "admin");

  // 调试日志
  useEffect(() => {
    console.log("user:", user);
    console.log("activity:", activity);
    console.log("isOwner:", isOwner);
  }, [user, activity, isOwner]);

  if (loading || !user) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="flex items-center gap-2">
          <Clock className="h-8 w-8 animate-spin" />
          <span className="text-lg">加载中...</span>
        </div>
      </div>
    );
  }

  if (!activity) {
    return (
      <div className="flex flex-col items-center mt-16">
        <AlertCircle className="h-16 w-16 text-red-500 mb-4" />
        <h2 className="text-xl font-semibold text-red-500 mb-2">
          未找到该活动
        </h2>
        <Button onClick={() => navigate("/affairs")}>返回活动列表</Button>
      </div>
    );
  }

  // 计算统计数据
  const participants = activity.participants || [];
  const applications = activity.applications || [];
  const approvedApps = applications.filter((app) => app.status === "approved");
  const totalCredits = participants.reduce((sum, p) => sum + p.credits, 0);
  const totalAwardedCredits = approvedApps.reduce(
    (sum, app) => sum + app.awarded_credits,
    0
  );

  // 获取状态显示文本
  const getStatusText = (status: string) => {
    switch (status) {
      case "draft":
        return "草稿";
      case "pending_review":
        return "待审核";
      case "approved":
        return "已通过";
      case "rejected":
        return "已拒绝";
      default:
        return status;
    }
  };

  // 获取状态样式
  const getStatusStyle = (status: string) => {
    switch (status) {
      case "approved":
        return "bg-green-100 text-green-800";
      case "pending_review":
        return "bg-yellow-100 text-yellow-800";
      case "rejected":
        return "bg-red-100 text-red-800";
      case "draft":
        return "bg-gray-100 text-gray-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  // 获取状态图标
  const getStatusIcon = (status: string) => {
    switch (status) {
      case "approved":
        return <CheckCircle className="w-3 h-3 mr-1 inline" />;
      case "pending_review":
        return <AlertCircle className="w-3 h-3 mr-1 inline" />;
      case "rejected":
        return <XCircle className="w-3 h-3 mr-1 inline" />;
      default:
        return <AlertCircle className="w-3 h-3 mr-1 inline" />;
    }
  };

  return (
    <div className="max-w-6xl mx-auto p-4 md:p-8 space-y-8">
      {/* 返回按钮和操作按钮 */}
      <div className="flex items-center justify-between">
        <Button variant="ghost" onClick={() => navigate(-1)} className="mb-2">
          <ArrowLeft className="h-4 w-4 mr-2" /> 返回
        </Button>

        {/* 操作按钮 */}
        <div className="flex items-center gap-2">
          {isOwner && activity.status === "draft" && (
            <Button
              onClick={handleSubmitForReview}
              className="bg-blue-600 hover:bg-blue-700"
            >
              <Send className="h-4 w-4 mr-2" />
              提交审核
            </Button>
          )}

          {(activity.status === "pending_review" ||
            activity.status === "approved" ||
            activity.status === "rejected") &&
            isOwner && (
              <Button onClick={handleWithdraw} variant="outline">
                <RotateCcw className="h-4 w-4 mr-2" />
                撤回活动
              </Button>
            )}

          {isOwner && (
            <Button
              onClick={() => navigate(`/affairs/edit/${activity.id}`)}
              variant="outline"
            >
              <Edit className="h-4 w-4 mr-2" />
              编辑
            </Button>
          )}

          {isOwner && (
            <Button onClick={handleDelete} variant="destructive">
              <Trash className="h-4 w-4 mr-2" />
              删除
            </Button>
          )}
        </div>
      </div>

      {/* 活动基本信息 */}
      <Card className="rounded-xl shadow-lg bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
        <CardHeader>
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-3">
              <div className="p-3 rounded-full bg-primary/10">
                <Award className="h-8 w-8 text-primary" />
              </div>
              <div>
                <CardTitle className="text-3xl font-bold">
                  {activity.title}
                </CardTitle>
                <div className="flex items-center gap-2 mt-2">
                  <Badge className="bg-blue-100 text-blue-800 hover:bg-blue-200">
                    {activity.category}
                  </Badge>
                  <Badge
                    className={`${getStatusStyle(
                      activity.status
                    )} rounded-lg px-2 py-1`}
                  >
                    {getStatusIcon(activity.status)}
                    {getStatusText(activity.status)}
                  </Badge>
                </div>
              </div>
            </div>
            <div className="text-right">{/* 活动ID已隐藏 */}</div>
          </div>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="text-lg leading-relaxed whitespace-pre-line">
            {activity.description}
          </div>

          {activity.requirements && (
            <div className="bg-white/50 dark:bg-gray-800/50 rounded-lg p-4">
              <h3 className="font-semibold mb-2">活动要求</h3>
              <p className="text-gray-700 dark:text-gray-300">
                {activity.requirements}
              </p>
            </div>
          )}

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
              <User className="h-5 w-5 text-blue-600" />
              <div>
                <div className="font-medium">创建人</div>
                <div className="text-sm text-muted-foreground">
                  {activity.owner_id}
                </div>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
              <Star className="h-5 w-5 text-yellow-600" />
              <div>
                <div className="font-medium">总学分</div>
                <div className="text-sm text-muted-foreground">
                  {totalCredits.toFixed(1)} 学分
                </div>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
              <Calendar className="h-5 w-5 text-green-600" />
              <div>
                <div className="font-medium">创建时间</div>
                <div className="text-sm text-muted-foreground">
                  {new Date(activity.created_at).toLocaleDateString()}
                </div>
              </div>
            </div>
          </div>

          {activity.start_date && activity.end_date && (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
                <Calendar className="h-5 w-5 text-green-600" />
                <div>
                  <div className="font-medium">开始时间</div>
                  <div className="text-sm text-muted-foreground">
                    {new Date(activity.start_date).toLocaleDateString()}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
                <Calendar className="h-5 w-5 text-red-600" />
                <div>
                  <div className="font-medium">结束时间</div>
                  <div className="text-sm text-muted-foreground">
                    {new Date(activity.end_date).toLocaleDateString()}
                  </div>
                </div>
              </div>
            </div>
          )}

          {activity.review_comments && (
            <div className="bg-white/50 dark:bg-gray-800/50 rounded-lg p-4">
              <h3 className="font-semibold mb-2">审核意见</h3>
              <p className="text-gray-700 dark:text-gray-300">
                {activity.review_comments}
              </p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 统计卡片 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card className="rounded-xl shadow-lg">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  参与学生
                </p>
                <p className="text-2xl font-bold">{participants.length}</p>
              </div>
              <Users className="h-8 w-8 text-blue-600" />
            </div>
          </CardContent>
        </Card>
        <Card className="rounded-xl shadow-lg">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  总申请数
                </p>
                <p className="text-2xl font-bold">{applications.length}</p>
              </div>
              <FileText className="h-8 w-8 text-green-600" />
            </div>
          </CardContent>
        </Card>
        <Card className="rounded-xl shadow-lg">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  已通过申请
                </p>
                <p className="text-2xl font-bold">{approvedApps.length}</p>
              </div>
              <FileCheck className="h-8 w-8 text-green-600" />
            </div>
          </CardContent>
        </Card>
        <Card className="rounded-xl shadow-lg">
          <CardContent className="p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">
                  已授予学分
                </p>
                <p className="text-2xl font-bold">
                  {totalAwardedCredits.toFixed(1)}
                </p>
              </div>
              <TrendingUp className="h-8 w-8 text-purple-600" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 参与者列表 */}
      {participants.length > 0 && (
        <Card className="rounded-xl shadow-lg">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Users className="h-5 w-5" />
              参与者列表
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {participants.map((participant, index) => (
                <div
                  key={participant.user_id}
                  className="flex items-center justify-between p-3 bg-muted/50 rounded-lg"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
                      <User className="h-4 w-4 text-primary" />
                    </div>
                    <div>
                      <div className="font-medium">
                        {participant.user_info?.name ||
                          `用户 ${participant.user_id}`}
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {participant.user_info?.student_id ||
                          participant.user_id}
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-bold text-primary">
                      {participant.credits} 学分
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {new Date(participant.joined_at).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}

      {/* 申请列表 */}
      {applications.length > 0 && (
        <Card className="rounded-xl shadow-lg">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              申请列表
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {applications.map((application) => (
                <div
                  key={application.id}
                  className="flex items-center justify-between p-3 bg-muted/50 rounded-lg"
                >
                  <div className="flex items-center gap-3">
                    <div className="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
                      <FileCheck className="h-4 w-4 text-green-600" />
                    </div>
                    <div>
                      <div className="font-medium">申请 #{application.id}</div>
                      <div className="text-sm text-muted-foreground">
                        用户: {application.user_id}
                      </div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-bold text-primary">
                      {application.awarded_credits} 学分
                    </div>
                    <Badge
                      className={`${getStatusStyle(
                        application.status
                      )} rounded-lg px-2 py-1`}
                    >
                      {getStatusIcon(application.status)}
                      {getStatusText(application.status)}
                    </Badge>
                    <div className="text-xs text-muted-foreground mt-1">
                      {new Date(application.submitted_at).toLocaleDateString()}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
