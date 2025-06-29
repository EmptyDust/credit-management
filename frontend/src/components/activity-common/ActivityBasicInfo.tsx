import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Award } from "lucide-react";
import type { Activity } from "@/types/activity";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { getStatusText, getStatusStyle, getStatusIcon, activityCategories } from "@/lib/utils";

interface ActivityBasicInfoProps {
  activity: Activity;
  isEditing?: boolean;
  onEditModeChange?: (isEditing: boolean) => void;
  onRefresh?: () => void;
  basicInfo: any;
  setBasicInfo: React.Dispatch<React.SetStateAction<any>>;
}

export default function ActivityBasicInfo({
  activity,
  isEditing = false,
  basicInfo,
  setBasicInfo,
}: ActivityBasicInfoProps) {
  // 计算总学分
  const participants = activity.participants || [];
  const totalCredits = participants.reduce((sum, p) => sum + p.credits, 0);

  const StatusIcon = getStatusIcon(activity.status);

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
                      <StatusIcon className="w-3 h-3 mr-1 inline" />
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
            <div className="prose prose-lg max-w-none">
              <p className="text-gray-700 dark:text-gray-300 leading-relaxed">
                {activity.description || "暂无描述"}
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm">
                <h4 className="text-sm font-medium text-gray-500 mb-1">
                  开始时间
                </h4>
                <p className="text-lg font-semibold">
                  {activity.start_date
                    ? new Date(activity.start_date).toLocaleDateString()
                    : "未设置"}
                </p>
              </div>
              <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm">
                <h4 className="text-sm font-medium text-gray-500 mb-1">
                  结束时间
                </h4>
                <p className="text-lg font-semibold">
                  {activity.end_date
                    ? new Date(activity.end_date).toLocaleDateString()
                    : "未设置"}
                </p>
              </div>
              <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow-sm">
                <h4 className="text-sm font-medium text-gray-500 mb-1">
                  总学分
                </h4>
                <p className="text-lg font-semibold text-green-600">
                  {totalCredits.toFixed(1)}
                </p>
              </div>
            </div>
          </>
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
