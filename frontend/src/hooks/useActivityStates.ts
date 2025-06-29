import { useState } from "react";

export const useActivityStates = () => {
  const [isManagingParticipants, setIsManagingParticipants] = useState(false);
  const [isManagingAttachments, setIsManagingAttachments] = useState(false);
  const [isReviewing, setIsReviewing] = useState(false);

  return {
    isManagingParticipants,
    setIsManagingParticipants,
    isManagingAttachments,
    setIsManagingAttachments,
    isReviewing,
    setIsReviewing,
  };
}; 