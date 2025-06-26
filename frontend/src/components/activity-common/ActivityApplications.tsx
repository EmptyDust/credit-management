import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import {
  FileText,
  FileCheck,
  CheckCircle,
  XCircle,
  Clock,
  AlertCircle,
} from "lucide-react";
import type { Activity } from "@/types/activity";

interface ActivityApplicationsProps {
  activity: Activity;
}

// 获取状态显示文本
const getStatusText = (status: string) => {
  switch (status) {
    case "approved":
      return "已通过";
    case "pending":
      return "待审核";
    case "rejected":
      return "已拒绝";
    default:
      return status;
  }
};

// 获取状态样式
const getStatusStyle = (status: string) => {
  switch (status) {
    case "approved":
      return "bg-green-100 text-green-800";
    case "pending":
      return "bg-yellow-100 text-yellow-800";
    case "rejected":
      return "bg-red-100 text-red-800";
    default:
      return "bg-gray-100 text-gray-800";
  }
};

// 获取状态图标
const getStatusIcon = (status: string) => {
  switch (status) {
    case "approved":
      return <CheckCircle className="w-3 h-3 mr-1 inline" />;
    case "pending":
      return <Clock className="w-3 h-3 mr-1 inline" />;
    case "rejected":
      return <XCircle className="w-3 h-3 mr-1 inline" />;
    default:
      return <AlertCircle className="w-3 h-3 mr-1 inline" />;
  }
};

export default function ActivityApplications({
  activity,
}: ActivityApplicationsProps) {
  const applications = activity.applications || [];

  if (applications.length === 0) {
    return (
      <Card className="rounded-xl shadow-lg">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            申请列表
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            暂无申请记录
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="rounded-xl shadow-lg">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <FileText className="h-5 w-5" />
          申请列表 ({applications.length})
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-3">
          {applications.map((application) => (
            <div
              key={application.id}
              className="flex items-center justify-between p-3 bg-muted/50 rounded-lg"
            >
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
                  <FileCheck className="h-4 w-4 text-green-600" />
                </div>
                <div>
                  <div className="font-medium">申请 #{application.id}</div>
                  <div className="text-sm text-muted-foreground">
                    用户: {application.user_id}
                  </div>
                </div>
              </div>
              <div className="text-right">
                <div className="font-bold text-primary">
                  {application.awarded_credits} 学分
                </div>
                <Badge
                  className={`${getStatusStyle(
                    application.status
                  )} rounded-lg px-2 py-1`}
                >
                  {getStatusIcon(application.status)}
                  {getStatusText(application.status)}
                </Badge>
                <div className="text-xs text-muted-foreground mt-1">
                  {new Date(application.submitted_at).toLocaleDateString()}
                </div>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
