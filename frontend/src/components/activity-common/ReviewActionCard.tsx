import { useState } from "react";
import { Textarea } from "../ui/textarea";
import { Button } from "../ui/button";
import apiClient from "@/lib/api";

interface ReviewActionCardProps {
  activityId: string;
  activityStatus: string;
  isReviewer: boolean;
  onSuccess?: () => void;
}

export default function ReviewActionCard({
  activityId,
  activityStatus,
  isReviewer,
  onSuccess,
}: ReviewActionCardProps) {
  const [reviewComment, setReviewComment] = useState("");
  const [loading, setLoading] = useState(false);

  const canReview = isReviewer && (
    activityStatus === "pending_review" || 
    activityStatus === "approved" || 
    activityStatus === "rejected"
  );

  const handleReview = async (status: "approved" | "rejected") => {
    const isModifying = activityStatus !== "pending_review";
    if (!reviewComment.trim()) {
      alert(isModifying ? "请填写修改原因" : "请填写审批意见");
      return;
    }
    setLoading(true);
    try {
      await apiClient.post(`/activities/${activityId}/review`, {
        status,
        review_comments: reviewComment,
      });
      alert(isModifying ? "审核状态修改成功" : "审批提交成功");
      setReviewComment("");
      onSuccess?.();
    } catch (e) {
      alert(isModifying ? "修改失败" : "审批失败");
    } finally {
      setLoading(false);
    }
  };

  if (!canReview) return null;

  const isModifying = activityStatus !== "pending_review";

  return (
    <div className="mt-8 p-6 bg-gray-50 dark:bg-gray-800 rounded-xl shadow space-y-4">
      <label className="block text-lg font-medium mb-2">
        {isModifying ? "修改审核状态" : "审批意见"}
      </label>
      <Textarea
        className="w-full min-h-[100px]"
        value={reviewComment}
        onChange={(e) => setReviewComment(e.target.value)}
        placeholder={isModifying ? "请输入修改原因" : "请输入审批意见"}
        disabled={loading}
      />
      <div className="flex gap-4 mt-2">
        <Button
          onClick={() => handleReview("approved")}
          disabled={loading}
          className="bg-green-600 hover:bg-green-700 text-white"
        >
          {isModifying ? "改为通过" : "同意审批"}
        </Button>
        <Button
          onClick={() => handleReview("rejected")}
          disabled={loading}
          className="bg-red-600 hover:bg-red-700 text-white"
        >
          {isModifying ? "改为拒绝" : "拒绝审批"}
        </Button>
      </div>
    </div>
  );
}
