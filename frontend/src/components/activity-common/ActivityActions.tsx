import { Button } from "@/components/ui/button";
import {
  ArrowLeft,
  Edit,
  Trash,
  Send,
  RotateCcw,
  CheckCircle,
  XCircle,
  Users,
  Paperclip,
} from "lucide-react";
import type { Activity } from "@/types/activity";
import { useNavigate } from "react-router-dom";
import { useState } from "react";
import apiClient from "@/lib/api";
import toast from "react-hot-toast";

interface ActivityActionsProps {
  activity: Activity;
  isOwner: boolean;
  isReviewer: boolean;
  onRefresh: () => void;
  onEditModeChange?: (isEditing: boolean) => void;
  onParticipantsModeChange?: (isManaging: boolean) => void;
  onAttachmentsModeChange?: (isManaging: boolean) => void;
  onReviewModeChange?: (isReviewing: boolean) => void;
}

export default function ActivityActions({
  activity,
  isOwner,
  isReviewer,
  onRefresh,
  onEditModeChange,
  onParticipantsModeChange,
  onAttachmentsModeChange,
  onReviewModeChange,
}: ActivityActionsProps) {
  const navigate = useNavigate();
  const [reviewStatus, setReviewStatus] = useState<"approved" | "rejected">(
    "approved"
  );
  const [reviewComment, setReviewComment] = useState("");

  const handleDelete = async () => {
    if (!window.confirm("确定要删除这个活动吗？此操作不可撤销。")) return;

    try {
      await apiClient.delete(`/activities/${activity.id}`);
      toast.success("活动删除成功");
      navigate("/activities");
    } catch (err) {
      toast.error("删除活动失败");
    }
  };

  const handleSubmitForReview = async () => {
    try {
      await apiClient.post(`/activities/${activity.id}/submit`);
      toast.success("活动提交审核成功");
      onRefresh();
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || "提交审核失败";
      toast.error(errorMessage);
    }
  };

  const handleWithdraw = async () => {
    try {
      await apiClient.post(`/activities/${activity.id}/withdraw`);
      toast.success("活动撤回成功");
      onRefresh();
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || "撤回活动失败";
      toast.error(errorMessage);
    }
  };

  const handleReview = async () => {
    if (!reviewComment.trim()) {
      toast.error("请填写审核评语");
      return;
    }

    try {
      await apiClient.post(`/activities/${activity.id}/review`, {
        status: reviewStatus,
        comment: reviewComment,
      });
      toast.success("审核完成");
      setReviewComment("");
      onReviewModeChange?.(false);
      onRefresh();
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || "审核失败";
      toast.error(errorMessage);
    }
  };

  const handleEdit = () => {
    onEditModeChange?.(true);
  };

  const handleManageParticipants = () => {
    onParticipantsModeChange?.(true);
  };

  const handleManageAttachments = () => {
    onAttachmentsModeChange?.(true);
  };

  const handleReviewClick = () => {
    onReviewModeChange?.(true);
  };

  return (
    <div className="flex items-center justify-end">
      {/* 操作按钮 */}
      <div className="flex items-center gap-2">
        {/* 提交审核按钮 - 只有草稿状态的活动创建者可以看到 */}
        {isOwner && activity.status === "draft" && (
          <Button
            onClick={handleSubmitForReview}
            className="bg-blue-600 hover:bg-blue-700"
          >
            <Send className="h-4 w-4 mr-2" />
            提交审核
          </Button>
        )}

        {/* 审批按钮 - 只有审核者可以看到待审核的活动 */}
        {isReviewer && activity.status === "pending_review" && (
          <Button
            onClick={handleReviewClick}
            className="bg-green-600 hover:bg-green-700"
          >
            <CheckCircle className="h-4 w-4 mr-2" />
            审批
          </Button>
        )}

        {/* 撤回按钮 - 活动创建者可以撤回已提交的活动 */}
        {(activity.status === "pending_review" ||
          activity.status === "approved" ||
          activity.status === "rejected") &&
          isOwner && (
            <Button onClick={handleWithdraw} variant="outline">
              <RotateCcw className="h-4 w-4 mr-2" />
              撤回活动
            </Button>
          )}

        {/* 编辑按钮 - 活动创建者可以编辑草稿状态的活动 */}
        {isOwner && activity.status === "draft" && (
          <Button onClick={handleEdit} variant="outline">
            <Edit className="h-4 w-4 mr-2" />
            编辑
          </Button>
        )}

        {/* 参与者管理按钮 - 活动创建者和管理员可以管理参与者 */}
        {(isOwner || isReviewer) && (
          <Button onClick={handleManageParticipants} variant="outline">
            <Users className="h-4 w-4 mr-2" />
            参与者
          </Button>
        )}

        {/* 附件管理按钮 - 活动创建者和管理员可以管理附件 */}
        {(isOwner || isReviewer) && (
          <Button onClick={handleManageAttachments} variant="outline">
            <Paperclip className="h-4 w-4 mr-2" />
            附件
          </Button>
        )}

        {/* 删除按钮 - 只有活动创建者可以删除草稿状态的活动 */}
        {isOwner && activity.status === "draft" && (
          <Button onClick={handleDelete} variant="destructive">
            <Trash className="h-4 w-4 mr-2" />
            删除
          </Button>
        )}
      </div>
    </div>
  );
}
