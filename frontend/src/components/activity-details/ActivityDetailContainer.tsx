import { useState, useImperativeHandle, forwardRef } from "react";
import type { ActivityWithDetails } from "../../types/activity";
import InnovationActivityDetail from "./InnovationActivityDetail";
import CompetitionActivityDetail from "./CompetitionActivityDetail";
import { EntrepreneurshipProjectDetail } from "./index";
import EntrepreneurshipPracticeDetail from "./EntrepreneurshipPracticeDetail";
import PaperPatentDetail from "./PaperPatentDetail";

interface ActivityDetailContainerProps {
  activity: ActivityWithDetails;
  isEditing?: boolean;
  onEditModeChange?: (isEditing: boolean) => void;
  onRefresh?: () => void;
  onSave?: (basicInfo: any, detailInfo: any) => Promise<void>;
}

export interface ActivityDetailContainerRef {
  handleSave: () => Promise<void>;
}

// 辅助函数：将ISO日期字符串转换为yyyy-MM-dd格式（用于date input）
const formatDateForInput = (dateString: string): string => {
  if (
    !dateString ||
    dateString.trim() === "" ||
    dateString === "0001-01-01T00:00:00Z"
  ) {
    return "";
  }
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      return "";
    }
    return date.toISOString().split("T")[0]; // 返回 yyyy-MM-dd 格式
  } catch (error) {
    console.error("Date parsing error:", error);
    return "";
  }
};

const ActivityDetailContainer = forwardRef<
  ActivityDetailContainerRef,
  ActivityDetailContainerProps
>(({ activity, isEditing, onEditModeChange, onRefresh, onSave }, ref) => {
  // 提升主表单状态
  const [basicInfo, setBasicInfo] = useState({
    title: activity.title || "",
    description: activity.description || "",
    category: activity.category || "",
    requirements: activity.requirements || "",
    start_date: formatDateForInput(activity.start_date || ""),
    end_date: formatDateForInput(activity.end_date || ""),
  });

  // 提升详细信息状态
  const [detailInfo, setDetailInfo] = useState(() => {
    switch (activity.category) {
      case "创新创业实践活动":
        return (
          activity.innovation_detail || {
            item: "",
            company: "",
            project_no: "",
            issuer: "",
            date: formatDateForInput(
              (activity.innovation_detail as any)?.date || ""
            ),
            total_hours: "",
          }
        );
      case "学科竞赛":
        return (
          activity.competition_detail || {
            competition: "",
            level: "",
            award_level: "",
            rank: "",
          }
        );
      case "大学生创业项目":
        return (
          activity.entrepreneurship_project_detail || {
            project_name: "",
            project_level: "",
            project_rank: "",
          }
        );
      case "创业实践项目":
        return (
          activity.entrepreneurship_practice_detail || {
            company_name: "",
            legal_person: "",
            share_percent: "",
          }
        );
      case "论文专利":
        return (
          activity.paper_patent_detail || {
            name: "",
            category: "",
            rank: "",
            publication_date: "",
          }
        );
      default:
        return {};
    }
  });

  // 暴露保存方法给父组件
  useImperativeHandle(ref, () => ({
    handleSave: async () => {
      console.log("ActivityDetailContainer handleSave called");
      console.log("Current basicInfo:", basicInfo);
      console.log("Current detailInfo:", detailInfo);
      if (onSave) {
        await onSave(basicInfo, detailInfo);
      }
    },
  }));

  // 根据活动类别渲染对应的详情组件
  switch (activity.category) {
    case "创新创业实践活动":
      return (
        <InnovationActivityDetail
          activity={activity}
          detail={activity.innovation_detail}
          isEditing={isEditing}
          onEditModeChange={onEditModeChange}
          onRefresh={onRefresh}
          onSave={onSave}
          basicInfo={basicInfo}
          setBasicInfo={setBasicInfo}
          detailInfo={detailInfo}
          setDetailInfo={setDetailInfo}
        />
      );

    case "学科竞赛":
      return (
        <CompetitionActivityDetail
          activity={activity}
          detail={activity.competition_detail}
          isEditing={isEditing}
          onEditModeChange={onEditModeChange}
          onRefresh={onRefresh}
          onSave={onSave}
          basicInfo={basicInfo}
          setBasicInfo={setBasicInfo}
          detailInfo={detailInfo}
          setDetailInfo={setDetailInfo}
        />
      );

    case "大学生创业项目":
      return (
        <EntrepreneurshipProjectDetail
          activity={activity}
          detail={activity.entrepreneurship_project_detail}
          isEditing={isEditing}
          onEditModeChange={onEditModeChange}
          onRefresh={onRefresh}
          onSave={onSave}
          basicInfo={basicInfo}
          setBasicInfo={setBasicInfo}
          detailInfo={detailInfo}
          setDetailInfo={setDetailInfo}
        />
      );

    case "创业实践项目":
      return (
        <EntrepreneurshipPracticeDetail
          activity={activity}
          detail={activity.entrepreneurship_practice_detail}
          isEditing={isEditing}
          onEditModeChange={onEditModeChange}
          onRefresh={onRefresh}
          onSave={onSave}
          basicInfo={basicInfo}
          setBasicInfo={setBasicInfo}
          detailInfo={detailInfo}
          setDetailInfo={setDetailInfo}
        />
      );

    case "论文专利":
      return (
        <PaperPatentDetail
          activity={activity}
          detail={activity.paper_patent_detail}
          isEditing={isEditing}
          onEditModeChange={onEditModeChange}
          onRefresh={onRefresh}
          onSave={onSave}
          basicInfo={basicInfo}
          setBasicInfo={setBasicInfo}
          detailInfo={detailInfo}
          setDetailInfo={setDetailInfo}
        />
      );

    default:
      return (
        <div className="p-6 bg-white rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-4">活动详情</h2>
          <p className="text-gray-600">未知的活动类型：{activity.category}</p>
        </div>
      );
  }
});

ActivityDetailContainer.displayName = "ActivityDetailContainer";

export default ActivityDetailContainer;
