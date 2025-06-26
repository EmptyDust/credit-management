import React from "react";
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
import { Card, CardContent, CardHeader, CardTitle } from "../ui/card";
import { Badge } from "../ui/badge";
import { Building2, Trophy, Award } from "lucide-react";
import { useAuth } from "../../contexts/AuthContext";

interface EntrepreneurshipProjectDetailProps {
  activity: ActivityWithDetails;
  detail?: DetailType;
}

const EntrepreneurshipProjectDetail: React.FC<
  EntrepreneurshipProjectDetailProps
> = ({ activity, detail }) => {
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

      {/* 创业项目详情 */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5 text-blue-600" />
            创业项目详情
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {detail ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  项目名称
                </label>
                <p className="text-sm">{detail.project_name || "未填写"}</p>
              </div>

              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-500">
                  项目等级
                </label>
                <div className="flex items-center gap-2">
                  <Trophy className="h-4 w-4 text-yellow-500" />
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
                  <Award className="h-4 w-4 text-green-500" />
                  <Badge variant="outline">
                    {detail.project_rank || "未填写"}
                  </Badge>
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
      <ActivityParticipants activity={activity} />

      {/* 申请列表 */}
      <ActivityApplications activity={activity} />
    </div>
  );
};

export default EntrepreneurshipProjectDetail;
