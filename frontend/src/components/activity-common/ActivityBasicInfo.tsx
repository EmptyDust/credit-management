import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import {
  Award,
  User,
  Calendar,
  CheckCircle,
  XCircle,
  Clock,
  AlertCircle,
  Star,
} from "lucide-react";
import type { Activity } from "@/types/activity";
import { useState } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import apiClient from "@/lib/api";
import toast from "react-hot-toast";

interface ActivityBasicInfoProps {
  activity: Activity;
  isEditing?: boolean;
  onEditModeChange?: (isEditing: boolean) => void;
  onRefresh?: () => void;
  basicInfo: any;
  setBasicInfo: React.Dispatch<React.SetStateAction<any>>;
}

// 获取状态显示文本
const getStatusText = (status: string) => {
  switch (status) {
    case "draft":
      return "草稿";
    case "pending_review":
      return "待审核";
    case "approved":
      return "已通过";
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
    case "pending_review":
      return "bg-yellow-100 text-yellow-800";
    case "rejected":
      return "bg-red-100 text-red-800";
    case "draft":
      return "bg-gray-100 text-gray-800";
    default:
      return "bg-gray-100 text-gray-800";
  }
};

// 获取状态图标
const getStatusIcon = (status: string) => {
  switch (status) {
    case "approved":
      return <CheckCircle className="w-3 h-3 mr-1 inline" />;
    case "pending_review":
      return <AlertCircle className="w-3 h-3 mr-1 inline" />;
    case "rejected":
      return <XCircle className="w-3 h-3 mr-1 inline" />;
    default:
      return <Clock className="w-3 h-3 mr-1 inline" />;
  }
};

const activityCategories = [
  { value: "创新创业实践活动", label: "创新创业实践活动" },
  { value: "学科竞赛", label: "学科竞赛" },
  { value: "大学生创业项目", label: "大学生创业项目" },
  { value: "创业实践项目", label: "创业实践项目" },
  { value: "论文专利", label: "论文专利" },
];

export default function ActivityBasicInfo({
  activity,
  isEditing = false,
  onEditModeChange,
  onRefresh,
  basicInfo,
  setBasicInfo,
}: ActivityBasicInfoProps) {
  const [saving, setSaving] = useState(false);

  const handleSave = async (values: any) => {
    setSaving(true);
    try {
      await apiClient.put(`/activities/${activity.id}`, values);
      toast.success("活动信息更新成功");
      onEditModeChange?.(false);
      onRefresh?.();
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || "更新失败";
      toast.error(errorMessage);
    } finally {
      setSaving(false);
    }
  };

  const handleCancel = () => {
    onEditModeChange?.(false);
  };

  // 计算总学分
  const participants = activity.participants || [];
  const totalCredits = participants.reduce((sum, p) => sum + p.credits, 0);

  return (
    <Card className="rounded-xl shadow-lg bg-gradient-to-br from-blue-50 to-indigo-100 dark:from-gray-900 dark:to-gray-800">
      <CardHeader>
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="p-3 rounded-full bg-primary/10">
              <Award className="h-8 w-8 text-primary" />
            </div>
            <div className="flex-1">
              {isEditing ? (
                <form
                  className="space-y-4"
                  id="activity-edit-form"
                  autoComplete="off"
                >
                  <div className="mb-4">
                    <label className="block text-sm font-medium mb-1">
                      活动标题
                    </label>
                    <Input
                      className="text-2xl font-bold"
                      value={basicInfo.title}
                      onChange={(e) =>
                        setBasicInfo({ ...basicInfo, title: e.target.value })
                      }
                      placeholder="请输入活动标题"
                    />
                  </div>
                  <div className="mb-4">
                    <label className="block text-sm font-medium mb-1">
                      活动类别
                    </label>
                    <Select
                      value={basicInfo.category}
                      onValueChange={(val) =>
                        setBasicInfo({ ...basicInfo, category: val })
                      }
                    >
                      <SelectTrigger>
                        <SelectValue placeholder="选择活动类别" />
                      </SelectTrigger>
                      <SelectContent>
                        {activityCategories.map((category) => (
                          <SelectItem
                            key={category.value}
                            value={category.value}
                          >
                            {category.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                </form>
              ) : (
                <>
                  <CardTitle className="text-3xl font-bold">
                    {activity.title}
                  </CardTitle>
                  <div className="flex items-center gap-2 mt-2">
                    <Badge className="bg-blue-100 text-blue-800 hover:bg-blue-200">
                      {activity.category}
                    </Badge>
                    <Badge
                      className={`${getStatusStyle(
                        activity.status
                      )} rounded-lg px-2 py-1`}
                    >
                      {getStatusIcon(activity.status)}
                      {getStatusText(activity.status)}
                    </Badge>
                  </div>
                </>
              )}
            </div>
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-6">
        {isEditing ? (
          <form
            className="space-y-4"
            id="activity-edit-form-details"
            autoComplete="off"
          >
            <div className="mb-4">
              <label className="block text-sm font-medium mb-1">活动描述</label>
              <Textarea
                className="min-h-[120px] text-lg leading-relaxed"
                placeholder="请详细描述活动内容..."
                value={basicInfo.description}
                onChange={(e) =>
                  setBasicInfo({ ...basicInfo, description: e.target.value })
                }
              />
            </div>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium mb-1">
                  开始时间
                </label>
                <Input
                  type="date"
                  value={basicInfo.start_date}
                  onChange={(e) =>
                    setBasicInfo({ ...basicInfo, start_date: e.target.value })
                  }
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">
                  结束时间
                </label>
                <Input
                  type="date"
                  value={basicInfo.end_date}
                  onChange={(e) =>
                    setBasicInfo({ ...basicInfo, end_date: e.target.value })
                  }
                />
              </div>
            </div>
          </form>
        ) : (
          <>
            <div className="text-lg leading-relaxed whitespace-pre-line break-words max-w-full">
              {activity.description}
            </div>
          </>
        )}

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
            <User className="h-5 w-5 text-blue-600" />
            <div>
              <div className="font-medium">创建人</div>
              <div className="text-sm text-muted-foreground">
                {activity.owner_info?.name || activity.owner_id}
              </div>
            </div>
          </div>
          <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
            <Star className="h-5 w-5 text-yellow-600" />
            <div>
              <div className="font-medium">总学分</div>
              <div className="text-sm text-muted-foreground">
                {totalCredits.toFixed(1)} 学分
              </div>
            </div>
          </div>
          <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
            <Calendar className="h-5 w-5 text-green-600" />
            <div>
              <div className="font-medium">创建时间</div>
              <div className="text-sm text-muted-foreground">
                {new Date(activity.created_at).toLocaleDateString()}
              </div>
            </div>
          </div>
        </div>

        {activity.start_date && activity.end_date && !isEditing && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
              <Calendar className="h-5 w-5 text-green-600" />
              <div>
                <div className="font-medium">开始时间</div>
                <div className="text-sm text-muted-foreground">
                  {new Date(activity.start_date).toLocaleDateString()}
                </div>
              </div>
            </div>
            <div className="flex items-center gap-3 p-3 bg-white/50 dark:bg-gray-800/50 rounded-lg">
              <Calendar className="h-5 w-5 text-red-600" />
              <div>
                <div className="font-medium">结束时间</div>
                <div className="text-sm text-muted-foreground">
                  {new Date(activity.end_date).toLocaleDateString()}
                </div>
              </div>
            </div>
          </div>
        )}

        {activity.review_comments && (
          <div className="bg-white/50 dark:bg-gray-800/50 rounded-lg p-4">
            <h3 className="font-semibold mb-2">审核意见</h3>
            <p className="text-gray-700 dark:text-gray-300">
              {activity.review_comments}
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
