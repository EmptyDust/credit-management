import React from "react";
import type { ActivityWithDetails } from "../../types/activity";
import InnovationActivityDetail from "./InnovationActivityDetail";
import CompetitionActivityDetail from "./CompetitionActivityDetail";
import EntrepreneurshipProjectDetail from "./EntrepreneurshipProjectDetail";
import EntrepreneurshipPracticeDetail from "./EntrepreneurshipPracticeDetail";
import PaperPatentDetail from "./PaperPatentDetail";

interface ActivityDetailContainerProps {
  activity: ActivityWithDetails;
}

const ActivityDetailContainer: React.FC<ActivityDetailContainerProps> = ({
  activity,
}) => {
  // 根据活动类别渲染对应的详情组件
  switch (activity.category) {
    case "创新创业实践活动":
      return (
        <InnovationActivityDetail
          activity={activity}
          detail={activity.innovation_detail}
        />
      );

    case "学科竞赛":
      return (
        <CompetitionActivityDetail
          activity={activity}
          detail={activity.competition_detail}
        />
      );

    case "大学生创业项目":
      return (
        <EntrepreneurshipProjectDetail
          activity={activity}
          detail={activity.entrepreneurship_project_detail}
        />
      );

    case "创业实践项目":
      return (
        <EntrepreneurshipPracticeDetail
          activity={activity}
          detail={activity.entrepreneurship_practice_detail}
        />
      );

    case "论文专利":
      return (
        <PaperPatentDetail
          activity={activity}
          detail={activity.paper_patent_detail}
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
};

export default ActivityDetailContainer;
