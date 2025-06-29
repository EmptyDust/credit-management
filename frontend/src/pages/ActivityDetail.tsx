import { useEffect, useState, useRef } from "react";
import { useParams, useNavigate, useSearchParams } from "react-router-dom";
import { Card, CardContent } from "@/components/ui/card";
import {
  Clock,
  AlertCircle,
  ArrowLeft,
  Edit3,
  Send,
  Trash2,
} from "lucide-react";
import apiClient from "@/lib/api";
import toast from "react-hot-toast";
import { useAuth } from "@/contexts/AuthContext";
import type { ActivityWithDetails } from "@/types/activity";

// 导入活动详情组件
import { ActivityDetailContainer } from "@/components/activity-details/index";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Textarea } from "@/components/ui/textarea";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

export default function ActivityDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { user } = useAuth();
  const [activity, setActivity] = useState<ActivityWithDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [isEditing, setIsEditing] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [showReviewDialog, setShowReviewDialog] = useState(false);
  const [reviewStatus, setReviewStatus] = useState<"approved" | "rejected">(
    "approved"
  );
  const [reviewComment, setReviewComment] = useState("");
  const [basicInfo, setBasicInfo] = useState<any>({});
  const [detailInfo, setDetailInfo] = useState<any>({});
  const detailContainerRef = useRef<any>(null);

  const fetchActivity = async () => {
    if (!id) return;
    setLoading(true);

    try {
      const response = await apiClient.get(`/activities/${id}`);
      const activityData = response.data.data || response.data;
      setActivity(activityData);
    } catch (error) {
      console.error("Failed to fetch activity:", error);
      toast.error("获取活动详情失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchActivity();
    // 检查URL参数，自动进入编辑模式
    const editParam = searchParams.get("edit");
    if (editParam === "1") {
      setIsEditing(true);
    }
  }, [id, searchParams]);

  const isOwner = user && activity && (user.user_id === activity.owner_id || user.userType === "admin");
  const isReviewer = user?.userType === "teacher" || user?.userType === "admin";
  const canEdit = isOwner && activity?.status === "draft";
  const canSubmitForReview =
    user &&
    activity &&
    user.user_id === activity.owner_id &&
    activity?.status === "draft";
  const canWithdraw =
    user &&
    activity &&
    user.user_id === activity.owner_id &&
    activity?.status !== "draft";
  const canReview =
    isReviewer &&
    (activity?.status === "pending_review" ||
      activity?.status === "approved" ||
      activity?.status === "rejected");
  const canDelete =
    user?.userType === "admin" ||
    (user?.user_id === activity?.owner_id && activity?.status === "draft");

  const handleEdit = () => {
    setIsEditing(true);
  };

  // 辅助函数：将ISO日期字符串转换为yyyy-MM-dd格式
  const formatDateForAPI = (dateString: string): string | null => {
    if (
      !dateString ||
      dateString.trim() === "" ||
      dateString === "0001-01-01T00:00:00Z"
    ) {
      return null;
    }
    try {
      const date = new Date(dateString);
      if (isNaN(date.getTime())) {
        return null;
      }
      return date.toISOString().split("T")[0]; // 返回 yyyy-MM-dd 格式
    } catch (error) {
      console.error("Date parsing error:", error);
      return null;
    }
  };

  const handleSave = async (basicInfo: any, detailInfo: any) => {
    if (!activity) return;

    // 验证日期
    if (basicInfo.start_date && basicInfo.end_date) {
      const startDate = new Date(basicInfo.start_date);
      const endDate = new Date(basicInfo.end_date);

      if (endDate <= startDate) {
        toast.error("结束时间必须在开始时间之后");
        return;
      }
    }

    setSubmitting(true);

    try {
      // 构建更新请求数据
      const updateData: any = {
        title: basicInfo.title || activity.title,
        description: basicInfo.description || activity.description,
        category: basicInfo.category || activity.category,
      };

      // 处理日期字段，转换为正确的格式
      const startDate = formatDateForAPI(
        basicInfo.start_date || activity.start_date
      );
      if (startDate) {
        updateData.start_date = startDate;
      }

      const endDate = formatDateForAPI(basicInfo.end_date || activity.end_date);
      if (endDate) {
        updateData.end_date = endDate;
      }

      // 根据活动类别添加对应的详情数据，并处理数据类型转换
      switch (activity.category) {
        case "创新创业实践活动":
          const innovationDetail: any = {
            item: detailInfo.item || "",
            company: detailInfo.company || "",
            project_no: detailInfo.project_no || "",
            issuer: detailInfo.issuer || "",
          };
          if (
            detailInfo.total_hours !== undefined &&
            detailInfo.total_hours !== null
          ) {
            innovationDetail.total_hours = detailInfo.total_hours
              ? parseFloat(detailInfo.total_hours)
              : 0;
          }
          if (detailInfo.date !== undefined && detailInfo.date !== null) {
            const formattedDate = formatDateForAPI(detailInfo.date);
            innovationDetail.date = formattedDate || "";
          }
          updateData.innovation_detail = innovationDetail;
          break;
        case "学科竞赛":
          const competitionDetail: any = {
            competition: detailInfo.competition || "",
            level: detailInfo.level || "",
            award_level: detailInfo.award_level || "",
            rank: detailInfo.rank || "",
          };
          updateData.competition_detail = competitionDetail;
          break;
        case "大学生创业项目":
          const projectDetail: any = {
            project_name: detailInfo.project_name || "",
            project_level: detailInfo.project_level || "",
            project_rank: detailInfo.project_rank || "",
          };
          updateData.entrepreneurship_project_detail = projectDetail;
          break;
        case "创业实践项目":
          const practiceDetail: any = {
            company_name: detailInfo.company_name || "",
            legal_person: detailInfo.legal_person || "",
          };
          if (
            detailInfo.share_percent !== undefined &&
            detailInfo.share_percent !== null
          ) {
            practiceDetail.share_percent = detailInfo.share_percent
              ? parseFloat(detailInfo.share_percent)
              : 0;
          }
          updateData.entrepreneurship_practice_detail = practiceDetail;
          break;
        case "论文专利":
          const patentDetail: any = {
            name: detailInfo.name || "",
            category: detailInfo.category || "",
            rank: detailInfo.rank || "",
          };
          updateData.paper_patent_detail = patentDetail;
          break;
      }

      await apiClient.put(`/activities/${activity.id}`, updateData);
      toast.success("活动保存成功");
      setIsEditing(false);
      fetchActivity();
    } catch (err: any) {
      console.error("Save error:", err);
      const errorMessage = err.response?.data?.message || "保存失败";
      toast.error(errorMessage);
    } finally {
      setSubmitting(false);
    }
  };

  const handleSaveClick = async () => {
    if (detailContainerRef.current) {
      await detailContainerRef.current.handleSave();
    }
  };

  const handleSubmitForReview = async () => {
    if (!activity) return;
    setSubmitting(true);
    try {
      await apiClient.post(`/activities/${activity.id}/submit`);
      toast.success("活动提交审核成功");
      setIsEditing(false);
      fetchActivity();
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || "提交审核失败";
      toast.error(errorMessage);
    } finally {
      setSubmitting(false);
    }
  };

  const handleWithdraw = async () => {
    if (!activity) return;
    setSubmitting(true);
    try {
      await apiClient.post(`/activities/${activity.id}/withdraw`);
      toast.success("活动已撤回");
      fetchActivity();
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || "撤回失败";
      toast.error(errorMessage);
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!activity) return;
    if (!window.confirm("确定要删除这个活动吗？此操作不可撤销。")) return;

    setSubmitting(true);
    try {
      await apiClient.delete(`/activities/${activity.id}`);
      toast.success("活动删除成功");
      navigate("/activities");
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || "删除失败";
      toast.error(errorMessage);
    } finally {
      setSubmitting(false);
    }
  };

  const handleReview = async () => {
    if (!activity) return;
    if (!reviewComment.trim()) {
      toast.error(
        activity.status === "pending_review"
          ? "请填写审核评语"
          : "请填写修改原因"
      );
      return;
    }

    setSubmitting(true);
    try {
      await apiClient.post(`/activities/${activity.id}/review`, {
        status: reviewStatus,
        review_comments: reviewComment,
      });
      toast.success(
        activity.status === "pending_review" ? "审核完成" : "审核状态修改完成"
      );
      setReviewComment("");
      setShowReviewDialog(false);
      fetchActivity();
    } catch (err: any) {
      const errorMessage =
        err.response?.data?.message ||
        (activity.status === "pending_review" ? "审核失败" : "修改失败");
      toast.error(errorMessage);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading || !user) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex justify-center items-center h-64">
          <div className="flex items-center gap-2">
            <Clock className="h-8 w-8 animate-spin" />
            <span className="text-lg">加载中...</span>
          </div>
        </div>
      </div>
    );
  }

  if (!activity) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Card>
          <CardContent className="flex flex-col items-center justify-center min-h-[400px]">
            <AlertCircle className="h-16 w-16 text-red-500 mb-4" />
            <h2 className="text-xl font-semibold text-red-500 mb-2">
              未找到该活动
            </h2>
            <button
              onClick={() => navigate("/activities")}
              className="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
            >
              返回活动列表
            </button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 顶部按钮区 */}
      <div className="mb-6 flex items-center justify-between">
        <button
          onClick={() => navigate("/activities")}
          className="flex items-center gap-2 px-4 py-2 text-gray-600 hover:text-gray-800"
        >
          <ArrowLeft className="h-4 w-4" />
          返回活动列表
        </button>
        <div className="flex gap-3 items-center">
          {canEdit &&
            (isEditing ? (
              <>
                <Button
                  onClick={() => setIsEditing(false)}
                  size="lg"
                  variant="outline"
                  className="font-bold px-6"
                  disabled={submitting}
                >
                  取消
                </Button>
                <Button
                  onClick={handleSaveClick}
                  size="lg"
                  className="bg-blue-600 hover:bg-blue-700 font-bold px-8 shadow-lg"
                  disabled={submitting}
                >
                  {submitting ? "保存中..." : "保存"}
                </Button>
              </>
            ) : (
              <>
                <Button
                  onClick={handleEdit}
                  size="lg"
                  variant="outline"
                  className="font-bold px-6"
                  disabled={isEditing}
                >
                  <Edit3 className="h-4 w-4 mr-2" />
                  编辑
                </Button>
                {canSubmitForReview && (
                  <Button
                    onClick={handleSubmitForReview}
                    size="lg"
                    className="bg-blue-600 hover:bg-blue-700 font-bold px-8 shadow-lg"
                    disabled={submitting}
                  >
                    <Send className="h-4 w-4 mr-2" />
                    {submitting ? "提交中..." : "提交审核"}
                  </Button>
                )}
              </>
            ))}
          {canWithdraw && (
            <Button
              onClick={handleWithdraw}
              size="lg"
              variant="destructive"
              className="font-bold px-6"
              disabled={submitting}
            >
              撤回活动
            </Button>
          )}

          {canReview && (
            <Button
              onClick={() => setShowReviewDialog(true)}
              size="lg"
              className="bg-green-600 hover:bg-green-700 font-bold px-6"
              disabled={submitting}
            >
              {activity?.status === "pending_review" ? "审批" : "修改审核状态"}
            </Button>
          )}
          {canDelete && (
            <Button
              onClick={handleDelete}
              size="lg"
              variant="destructive"
              className="font-bold px-6"
              disabled={submitting}
            >
              <Trash2 className="h-4 w-4 mr-2" />
              删除
            </Button>
          )}
        </div>
      </div>

      {/* 活动详情容器 */}
      <ActivityDetailContainer
        ref={detailContainerRef}
        activity={activity}
        isEditing={isEditing}
        onEditModeChange={setIsEditing}
        onRefresh={fetchActivity}
        onSave={handleSave}
        basicInfo={basicInfo}
        setBasicInfo={setBasicInfo}
        detailInfo={detailInfo}
        setDetailInfo={setDetailInfo}
      />

      {/* 审批弹窗 */}
      <Dialog open={showReviewDialog} onOpenChange={setShowReviewDialog}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>
              {activity?.status === "pending_review"
                ? "活动审批"
                : "修改审核状态"}
            </DialogTitle>
            <DialogDescription>
              {activity?.status === "pending_review"
                ? "请填写审批意见并选择审批结果"
                : "请填写修改原因并选择新的审核状态"}
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-4 py-4">
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="review-status" className="text-right">
                {activity?.status === "pending_review"
                  ? "审批结果"
                  : "审核状态"}
              </Label>
              <Select
                value={reviewStatus}
                onValueChange={(value: "approved" | "rejected") =>
                  setReviewStatus(value)
                }
              >
                <SelectTrigger className="col-span-3">
                  <SelectValue
                    placeholder={
                      activity?.status === "pending_review"
                        ? "选择审批结果"
                        : "选择审核状态"
                    }
                  />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="approved">通过</SelectItem>
                  <SelectItem value="rejected">拒绝</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <Label htmlFor="review-comment" className="text-right">
                {activity?.status === "pending_review"
                  ? "审批评语"
                  : "修改原因"}
              </Label>
              <Textarea
                id="review-comment"
                value={reviewComment}
                onChange={(e) => setReviewComment(e.target.value)}
                placeholder={
                  activity?.status === "pending_review"
                    ? "请输入审批评语..."
                    : "请输入修改原因..."
                }
                className="col-span-3"
                rows={4}
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setShowReviewDialog(false)}
              disabled={submitting}
            >
              取消
            </Button>
            <Button type="button" onClick={handleReview} disabled={submitting}>
              {submitting
                ? "提交中..."
                : activity?.status === "pending_review"
                ? "提交审批"
                : "提交修改"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
