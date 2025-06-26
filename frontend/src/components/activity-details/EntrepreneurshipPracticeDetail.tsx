import React, { useState } from "react";
import type {
  ActivityWithDetails,
  EntrepreneurshipPracticeDetail as DetailType,
} from "../../types/activity";
import {
  ActivityBasicInfo,
  ActivityParticipants,
  ActivityAttachments,
} from "../activity-common";
import { Card, CardHeader, CardTitle, CardContent } from "../ui/card";
import { Badge } from "../ui/badge";
import { Building2, User, Percent, Users } from "lucide-react";
import { useAuth } from "../../contexts/AuthContext";
import { Input } from "@/components/ui/input";
import ReviewActionCard from "../activity-common/ReviewActionCard";

interface EntrepreneurshipPracticeDetailProps {
  activity: ActivityWithDetails;
  detail?: DetailType;
  isEditing?: boolean;
  onEditModeChange?: (isEditing: boolean) => void;
  onRefresh?: () => void;
  onSave?: (basicInfo: any, detailInfo: any) => Promise<void>;
  basicInfo: any;
  setBasicInfo: React.Dispatch<React.SetStateAction<any>>;
  detailInfo: any;
  setDetailInfo: React.Dispatch<React.SetStateAction<any>>;
}

const EntrepreneurshipPracticeDetail: React.FC<
  EntrepreneurshipPracticeDetailProps
> = ({
  activity,
  detail,
  isEditing,
  onEditModeChange,
  onRefresh,
  onSave,
  basicInfo,
  setBasicInfo,
  detailInfo,
  setDetailInfo,
}) => {
  const { user } = useAuth();
  const [isManagingParticipants, setIsManagingParticipants] = useState(false);
  const [isManagingAttachments, setIsManagingAttachments] = useState(false);
  const [isReviewing, setIsReviewing] = useState(false);

  const isOwner =
    user?.user_id === activity.owner_id || user?.userType === "admin";
  const isReviewer = user?.userType === "teacher" || user?.userType === "admin";

  const handleRefresh = () => {
    if (onRefresh) onRefresh();
    else window.location.reload();
  };

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

      {/* 创业实践详情 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5 text-green-600" />
            创业实践详情
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {isEditing ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  公司名称
                </label>
                <Input
                  value={detailInfo.company_name || ""}
                  onChange={(e) =>
                    setDetailInfo({
                      ...detailInfo,
                      company_name: e.target.value,
                    })
                  }
                  placeholder="请输入公司名称"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  法人代表
                </label>
                <Input
                  value={detailInfo.legal_person || ""}
                  onChange={(e) =>
                    setDetailInfo({
                      ...detailInfo,
                      legal_person: e.target.value,
                    })
                  }
                  placeholder="请输入法人代表"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  持股比例
                </label>
                <Input
                  type="number"
                  value={detailInfo.share_percent || ""}
                  onChange={(e) =>
                    setDetailInfo({
                      ...detailInfo,
                      share_percent: e.target.value,
                    })
                  }
                  placeholder="请输入持股比例"
                />
              </div>
            </div>
          ) : detail ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  公司名称
                </label>
                <div className="flex items-center gap-2">
                  <Building2 className="h-4 w-4 text-blue-500" />
                  <p className="text-sm">{detail.company_name || "未填写"}</p>
                </div>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  法人代表
                </label>
                <div className="flex items-center gap-2">
                  <User className="h-4 w-4 text-green-500" />
                  <p className="text-sm">{detail.legal_person || "未填写"}</p>
                </div>
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  持股比例
                </label>
                <div className="flex items-center gap-2">
                  <Percent className="h-4 w-4 text-purple-500" />
                  <Badge variant="secondary">
                    {detail.share_percent
                      ? `${detail.share_percent}%`
                      : "未填写"}
                  </Badge>
                </div>
              </div>
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              <Building2 className="h-12 w-12 mx-auto mb-4 text-gray-300" />
              <p>暂无创业实践详情信息</p>
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

export default EntrepreneurshipPracticeDetail;
