import type { Activity, ActivityDetails } from "@/types/activity";
import GenericActivityDetail from "./GenericActivityDetail";

interface ActivityDetailContainerProps {
  activity: Activity;
  detail?: ActivityDetails;
  onRefresh?: () => void;
}

const ActivityDetailContainer = ({ activity, detail, onRefresh }: ActivityDetailContainerProps) => {
  return (
    <GenericActivityDetail
      activity={activity}
      detail={detail}
      onRefresh={onRefresh}
    />
  );
};

export default ActivityDetailContainer;
