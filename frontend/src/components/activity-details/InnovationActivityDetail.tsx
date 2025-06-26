import React, { useState } from "react";
import type {
  ActivityWithDetails,
  InnovationActivityDetail as DetailType,
} from "../../types/activity";
import {
  ActivityBasicInfo,
  ActivityActions,
  ActivityParticipants,
  ActivityApplications,
} from "../activity-common";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { Badge } from "../ui/badge";
import { Lightbulb, Building2, FileText, Calendar, Clock } from "lucide-react";
import { useAuth } from "../../contexts/AuthContext";

interface InnovationActivityDetailProps {
  activity: ActivityWithDetails;
  detail?: DetailType;
}

const InnovationActivityDetail: React.FC<InnovationActivityDetailProps> = ({
  activity,
  detail,
}) => {
  const { user } = useAuth();
  const [isEditing, setIsEditing] = useState(false);
  const [isManagingParticipants, setIsManagingParticipants] = useState(false);
  const [isManagingAttachments, setIsManagingAttachments] = useState(false);
  const [isReviewing, setIsReviewing] = useState(false);

  const isOwner = user?.id === activity.owner_id || user?.userType === "admin";
  const isReviewer = user?.userType === "teacher" || user?.userType === "admin";

  const handleRefresh = () => {
    // 这里可以添加刷新逻辑，比如重新获取活动数据
    window.location.reload();
  };

  return (
    <div className="space-y-6">
      {/* 活动基本信息 */}
      <ActivityBasicInfo
        activity={activity}
        isEditing={isEditing}
        onEditModeChange={setIsEditing}
        onRefresh={handleRefresh}
      />

      {/* 活动操作 */}
      <ActivityActions
        activity={activity}
        isOwner={isOwner}
        isReviewer={isReviewer}
        onRefresh={handleRefresh}
        onEditModeChange={setIsEditing}
        onParticipantsModeChange={setIsManagingParticipants}
        onAttachmentsModeChange={setIsManagingAttachments}
        onReviewModeChange={setIsReviewing}
      />

      {/* 创新创业详情 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Lightbulb className="h-5 w-5 text-yellow-600" />
            创新创业详情
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {detail ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  实践事项
                </label>
                <p className="text-sm">{detail.item || "未填写"}</p>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  实习公司
                </label>
                <div className="flex items-center gap-2">
                  <Building2 className="h-4 w-4 text-blue-500" />
                  <p className="text-sm">{detail.company || "未填写"}</p>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  课题编号
                </label>
                <div className="flex items-center gap-2">
                  <FileText className="h-4 w-4 text-green-500" />
                  <p className="text-sm">{detail.project_no || "未填写"}</p>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  发证机构
                </label>
                <p className="text-sm">{detail.issuer || "未填写"}</p>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  实践日期
                </label>
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-red-500" />
                  <p className="text-sm">{detail.date || "未填写"}</p>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  累计学时
                </label>
                <div className="flex items-center gap-2">
                  <Clock className="h-4 w-4 text-purple-500" />
                  <Badge variant="secondary">
                    {detail.total_hours
                      ? `${detail.total_hours}小时`
                      : "未填写"}
                  </Badge>
                </div>
              </div>
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              <Lightbulb className="h-12 w-12 mx-auto mb-4 text-gray-300" />
              <p>暂无创新创业详情信息</p>
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

export default InnovationActivityDetail;
