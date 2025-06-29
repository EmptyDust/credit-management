import type { ActivityWithDetails } from "@/types/activity";
import GenericActivityDetail from "./GenericActivityDetail";
import { useImperativeHandle, forwardRef } from "react";

interface ActivityDetailContainerProps {
  activity: ActivityWithDetails;
  detail?: any;
  isEditing?: boolean;
  onEditModeChange?: (isEditing: boolean) => void;
  onRefresh?: () => void;
  onSave?: (basicInfo: any, detailInfo: any) => Promise<void>;
  basicInfo: any;
  setBasicInfo: React.Dispatch<React.SetStateAction<any>>;
  detailInfo: any;
  setDetailInfo: React.Dispatch<React.SetStateAction<any>>;
}

const ActivityDetailContainer = forwardRef<{ handleSave: () => Promise<void> }, ActivityDetailContainerProps>(({
  activity,
  detail,
  isEditing,
  onEditModeChange,
  onRefresh,
  onSave,
  basicInfo,
  setBasicInfo,
  detailInfo,
  setDetailInfo,
}, ref) => {
  // 暴露保存方法给父组件
  useImperativeHandle(ref, () => ({
    handleSave: async () => {
      if (onSave) {
        await onSave(basicInfo, detailInfo);
      }
    },
  }));

  return (
    <GenericActivityDetail
      activity={activity}
      detail={detail}
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
});

export default ActivityDetailContainer;
