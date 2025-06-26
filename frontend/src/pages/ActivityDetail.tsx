import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Card, CardContent } from "@/components/ui/card";
import { Users, FileText, Clock, AlertCircle, ArrowLeft } from "lucide-react";
import apiClient from "@/lib/api";
import toast from "react-hot-toast";
import { useAuth } from "@/contexts/AuthContext";
import type { ActivityWithDetails } from "@/types/activity";

// 导入活动详情组件
import { ActivityDetailContainer } from "@/components/activity-details/index";

export default function ActivityDetailPage() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [activity, setActivity] = useState<ActivityWithDetails | null>(null);
  const [loading, setLoading] = useState(true);

  const fetchActivity = async () => {
    if (!id) return;
    setLoading(true);

    try {
      const response = await apiClient.get(`/activities/${id}`);
      const activityData = response.data.data || response.data;
      setActivity(activityData);
    } catch (error) {
      console.error("Failed to fetch activity:", error);
      toast.error("获取活动详情失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchActivity();
  }, [id]);

  if (loading || !user) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="flex justify-center items-center h-64">
          <div className="flex items-center gap-2">
            <Clock className="h-8 w-8 animate-spin" />
            <span className="text-lg">加载中...</span>
          </div>
        </div>
      </div>
    );
  }

  if (!activity) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Card>
          <CardContent className="flex flex-col items-center justify-center min-h-[400px]">
            <AlertCircle className="h-16 w-16 text-red-500 mb-4" />
            <h2 className="text-xl font-semibold text-red-500 mb-2">
              未找到该活动
            </h2>
            <button
              onClick={() => navigate("/activities")}
              className="mt-4 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
            >
              返回活动列表
            </button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* 返回按钮 */}
      <div className="mb-6">
        <button
          onClick={() => navigate("/activities")}
          className="flex items-center gap-2 px-4 py-2 text-gray-600 hover:text-gray-800"
        >
          <ArrowLeft className="h-4 w-4" />
          返回活动列表
        </button>
      </div>

      {/* 活动详情容器 */}
      <ActivityDetailContainer activity={activity} />
    </div>
  );
}
