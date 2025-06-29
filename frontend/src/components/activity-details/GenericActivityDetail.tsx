import React from "react";
import type { ActivityWithDetails } from "@/types/activity";
import {
  ActivityBasicInfo,
  ActivityParticipants,
  ActivityAttachments,
} from "../activity-common";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { Badge } from "../ui/badge";
import { Input } from "@/components/ui/input";
import { useAuth } from "@/contexts/AuthContext";
import ReviewActionCard from "../activity-common/ReviewActionCard";
import { activityDetailConfigs } from "@/lib/utils";
import * as Icons from "lucide-react";

interface GenericActivityDetailProps {
  activity: ActivityWithDetails;
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
  detailInfo,
  setDetailInfo,
}) => {
  const { user } = useAuth();
  const isReviewer = user?.userType === "teacher" || user?.userType === "admin";

  const handleRefresh = () => {
    if (onRefresh) onRefresh();
    else window.location.reload();
  };

  // 获取活动配置
  const config = activityDetailConfigs[activity.category as keyof typeof activityDetailConfigs];
  if (!config) {
    return <div>不支持的活动类型: {activity.category}</div>;
  }

  // 动态获取图标组件
  const IconComponent = Icons[config.icon as keyof typeof Icons] as React.ComponentType<any>;

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
          <CardTitle className="flex items-center gap-2">
            <IconComponent className={`h-5 w-5 ${config.color}`} />
            {config.title}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {isEditing ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {config.fields.map((field) => (
                <div key={field.key} className="space-y-2">
                  <label className="text-sm font-medium text-gray-500">
                    {field.label}
                  </label>
                  <Input
                    className="w-full"
                    type={field.type}
                    value={detailInfo[field.key] || ""}
                    onChange={(e) =>
                      setDetailInfo({
                        ...detailInfo,
                        [field.key]: e.target.value,
                      })
                    }
                    placeholder={`请输入${field.label}`}
                  />
                </div>
              ))}
            </div>
          ) : detail ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {config.fields.map((field) => (
                <div key={field.key} className="space-y-2">
                  <label className="text-sm font-medium text-gray-500">
                    {field.label}
                  </label>
                  <div className="flex items-center gap-2">
                    {field.key === "share_percent" && detail[field.key] ? (
                      <Badge variant="secondary">
                        {detail[field.key]}%
                      </Badge>
                    ) : (
                      <p className="text-sm">
                        {detail[field.key] || "未填写"}
                      </p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              <IconComponent className="h-12 w-12 mx-auto mb-4 text-gray-300" />
              <p>暂无{config.title}信息</p>
            </div>
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