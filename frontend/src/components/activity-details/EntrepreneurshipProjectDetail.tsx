import React, { useState } from "react";
import type {
  ActivityWithDetails,
  EntrepreneurshipProjectDetail as DetailType,
} from "../../types/activity";
import {
  ActivityBasicInfo,
  ActivityActions,
  ActivityParticipants,
  ActivityApplications,
} from "../activity-common";
import { Card, CardHeader, CardTitle, CardContent } from "../ui/card";
import { Badge } from "../ui/badge";
import { Building2, Award, Star, Users } from "lucide-react";
import { useAuth } from "../../contexts/AuthContext";
import { Input } from "@/components/ui/input";

interface EntrepreneurshipProjectDetailProps {
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

const EntrepreneurshipProjectDetail: React.FC<
  EntrepreneurshipProjectDetailProps
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

  const isOwner = user?.id === activity.owner_id || user?.userType === "admin";
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

      {/* 活动操作 */}
      <ActivityActions
        activity={activity}
        isOwner={isOwner}
        isReviewer={isReviewer}
        onRefresh={handleRefresh}
        onEditModeChange={onEditModeChange}
        onParticipantsModeChange={setIsManagingParticipants}
        onAttachmentsModeChange={setIsManagingAttachments}
        onReviewModeChange={setIsReviewing}
      />

      {/* 创业项目详情 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5 text-blue-600" />
            创业项目详情
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {isEditing ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  项目名称
                </label>
                <Input
                  value={detailInfo.project_name || ""}
                  onChange={(e) =>
                    setDetailInfo({
                      ...detailInfo,
                      project_name: e.target.value,
                    })
                  }
                  placeholder="请输入项目名称"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  项目级别
                </label>
                <Input
                  value={detailInfo.project_level || ""}
                  onChange={(e) =>
                    setDetailInfo({
                      ...detailInfo,
                      project_level: e.target.value,
                    })
                  }
                  placeholder="请输入项目级别"
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  项目排名
                </label>
                <Input
                  value={detailInfo.project_rank || ""}
                  onChange={(e) =>
                    setDetailInfo({
                      ...detailInfo,
                      project_rank: e.target.value,
                    })
                  }
                  placeholder="请输入项目排名"
                />
              </div>
            </div>
          ) : detail ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  项目名称
                </label>
                <p className="text-sm">{detail.project_name || "未填写"}</p>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  项目级别
                </label>
                <div className="flex items-center gap-2">
                  <Award className="h-4 w-4 text-blue-500" />
                  <Badge variant="secondary">
                    {detail.project_level || "未填写"}
                  </Badge>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  项目排名
                </label>
                <div className="flex items-center gap-2">
                  <Star className="h-4 w-4 text-yellow-500" />
                  <p className="text-sm">{detail.project_rank || "未填写"}</p>
                </div>
              </div>
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              <Building2 className="h-12 w-12 mx-auto mb-4 text-gray-300" />
              <p>暂无创业项目详情信息</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 参与者列表 */}
      <ActivityParticipants
        activity={activity}
        isManaging={isManagingParticipants}
        onManagingChange={setIsManagingParticipants}
      />

      {/* 申请列表 */}
      <ActivityApplications activity={activity} />
    </div>
  );
};

export default EntrepreneurshipProjectDetail;
