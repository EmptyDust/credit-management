import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
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
import type { Activity } from "@/types/activity";

// 导入活动详情组件
import { ActivityDetailContainer } from "@/components/activity-details/index";
import { ActivityEditDialog } from "@/components/activity-common";
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
import { getActivityOptions } from "@/lib/options";

export default function ActivityDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [activity, setActivity] = useState<Activity | null>(null);
  const [loading, setLoading] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [showReviewDialog, setShowReviewDialog] = useState(false);
  const [reviewStatus, setReviewStatus] = useState<"approved" | "rejected">(
    "approved"
  );
  const [reviewComment, setReviewComment] = useState("");
  const [reviewActions, setReviewActions] = useState<{ value: string; label: string }[]>([]);
  const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);

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
    (async () => {
      try {
        const opts = await getActivityOptions();
        setReviewActions(opts.review_actions || []);
      } catch (e) {
        console.error("Failed to load review actions", e);
      }
    })();
  }, [id]);

  const isOwner = user && activity && (user.uuid === activity.owner_id || user.userType === "admin");
  const isReviewer = user?.userType === "teacher" || user?.userType === "admin";
  const canEdit = isOwner && activity?.status === "draft";
  const canSubmitForReview =
    user &&
    activity &&
    user.uuid === activity.owner_id &&
    activity?.status === "draft";
  const canWithdraw =
    user &&
    activity &&
    user.uuid === activity.owner_id &&
    activity?.status !== "draft";
  const canReview =
    isReviewer &&
    (activity?.status === "pending_review" ||
      activity?.status === "approved" ||
      activity?.status === "rejected");
  const canDelete =
    user?.userType === "admin" ||
    (user?.uuid === activity?.owner_id && activity?.status === "draft");

  const handleEdit = () => {
    if (!activity) return;
    setIsEditDialogOpen(true);
  };

  const handleSubmitForReview = async () => {
    if (!activity) return;
    setSubmitting(true);
    try {
      await apiClient.post(`/activities/${activity.id}/submit`);
      toast.success("活动提交审核成功");
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
          {canEdit && (
            <>
              <Button
                onClick={handleEdit}
                size="lg"
                variant="outline"
                className="font-bold px-6"
                disabled={submitting}
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
          )}
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

      {/* 活动详情容器（仅展示，不在此处编辑基础信息） */}
      <ActivityDetailContainer activity={activity} />

      {/* 基础信息编辑弹窗（与列表页共用 ActivityEditDialog） */}
      <ActivityEditDialog
        open={isEditDialogOpen}
        onOpenChange={setIsEditDialogOpen}
        activity={activity}
        onSuccess={fetchActivity}
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
                  {reviewActions.map((a) => (
                    <SelectItem key={a.value} value={a.value}>{a.label}</SelectItem>
                  ))}
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
