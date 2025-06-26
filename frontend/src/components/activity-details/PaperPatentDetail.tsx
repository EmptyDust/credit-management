import React from "react";
import type {
  ActivityWithDetails,
  PaperPatentDetail as DetailType,
} from "../../types/activity";
import {
  ActivityBasicInfo,
  ActivityActions,
  ActivityParticipants,
  ActivityApplications,
} from "../activity-common";
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { Badge } from "../ui/badge";
import { FileText, Award, Hash } from "lucide-react";
import { useAuth } from "../../contexts/AuthContext";

interface PaperPatentDetailProps {
  activity: ActivityWithDetails;
  detail?: DetailType;
}

const PaperPatentDetail: React.FC<PaperPatentDetailProps> = ({
  activity,
  detail,
}) => {
  const { user } = useAuth();
  const isOwner = user?.id === activity.owner_id || user?.userType === "admin";

  const handleRefresh = () => {
    // 这里可以添加刷新逻辑，比如重新获取活动数据
    window.location.reload();
  };

  return (
    <div className="space-y-6">
      {/* 活动基本信息 */}
      <ActivityBasicInfo activity={activity} />

      {/* 活动操作 */}
      <ActivityActions
        activity={activity}
        isOwner={isOwner}
        onRefresh={handleRefresh}
      />

      {/* 论文专利详情 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5 text-purple-600" />
            论文专利详情
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {detail ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  名称
                </label>
                <p className="text-sm">{detail.name || "未填写"}</p>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  类别
                </label>
                <div className="flex items-center gap-2">
                  <FileText className="h-4 w-4 text-purple-500" />
                  <Badge variant="secondary">
                    {detail.category || "未填写"}
                  </Badge>
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  排名
                </label>
                <div className="flex items-center gap-2">
                  <Award className="h-4 w-4 text-yellow-500" />
                  <Badge variant="outline">{detail.rank || "未填写"}</Badge>
                </div>
              </div>
            </div>
          ) : (
            <div className="text-center py-8 text-gray-500">
              <FileText className="h-12 w-12 mx-auto mb-4 text-gray-300" />
              <p>暂无论文专利详情信息</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* 参与者列表 */}
      <ActivityParticipants activity={activity} />

      {/* 申请列表 */}
      <ActivityApplications activity={activity} />
    </div>
  );
};

export default PaperPatentDetail;
