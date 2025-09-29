import React from "react";
import type { Activity } from "@/types/activity";
import {
  ActivityBasicInfo,
  ActivityParticipants,
  ActivityAttachments,
} from "../activity-common";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { Badge } from "../ui/badge";
import { useAuth } from "@/contexts/AuthContext";
import ReviewActionCard from "../activity-common/ReviewActionCard";

interface GenericActivityDetailProps {
  activity: Activity;
  detail?: any;
  isEditing?: boolean;
  onEditModeChange?: (isEditing: boolean) => void;
  onRefresh?: () => void;
  onSave?: (basicInfo: any, detailInfo: any) => Promise<void>;
  basicInfo: any;
  setBasicInfo: React.Dispatch<React.SetStateAction<any>>;
  detailInfo: any;
  setDetailInfo: React.Dispatch<React.SetStateAction<any>>;
}

const GenericActivityDetail: React.FC<GenericActivityDetailProps> = ({
  activity,
  detail,
  isEditing,
  onEditModeChange,
  onRefresh,
  basicInfo,
  setBasicInfo,
}) => {
  const { user } = useAuth();
  const isReviewer = user?.userType === "teacher" || user?.userType === "admin";

  const handleRefresh = () => {
    if (onRefresh) onRefresh();
    else window.location.reload();
  };

  // 通用详情（已改为配置驱动，前端不再硬编码）
  const detailsData: Record<string, any> = detail ?? (activity as any).details ?? {};

  return (
    <div className="space-y-6">
      {/* 活动基本信息 */}
      <ActivityBasicInfo
        activity={activity}
        isEditing={isEditing}
        onEditModeChange={onEditModeChange}
        onRefresh={handleRefresh}
        basicInfo={basicInfo}
        setBasicInfo={setBasicInfo}
      />

      {/* 活动详情 */}
      <Card>
        <CardHeader>
          <CardTitle>活动详情</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {Object.keys(detailsData).length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {Object.entries(detailsData).map(([k, v]) => (
                <div key={k} className="space-y-1">
                  <label className="text-sm font-medium text-gray-500">
                    {k}
                  </label>
                  <div className="flex items-center gap-2">
                    {k === "share_percent" && v ? (
                      <Badge variant="secondary">{String(v)}%</Badge>
                    ) : (
                      <p className="text-sm break-words">{String(v ?? "未填写")}</p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">暂无活动详情</div>
          )}
        </CardContent>
      </Card>

      {/* 参与者列表 */}
      <ActivityParticipants activity={activity} onRefresh={handleRefresh} />

      {/* 附件 */}
      <ActivityAttachments activity={activity} onRefresh={handleRefresh} />

      {/* 审批意见卡片 */}
      <ReviewActionCard
        activityId={activity.id}
        activityStatus={activity.status}
        isReviewer={isReviewer}
        onSuccess={handleRefresh}
      />
    </div>
  );
};

export default GenericActivityDetail; 