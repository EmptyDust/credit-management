import { useState } from "react";

interface DialogStates {
  isDialogOpen: boolean;
  isDeleteDialogOpen: boolean;
  isImportDialogOpen: boolean;
  isSubmitting: boolean;
  importing: boolean;
  setDialogOpen: (open: boolean) => void;
  setDeleteDialogOpen: (open: boolean) => void;
  setImportDialogOpen: (open: boolean) => void;
  setSubmitting: (submitting: boolean) => void;
  setImporting: (importing: boolean) => void;
}

export function useDialogStates(): DialogStates {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [importing, setImporting] = useState(false);

  return {
    isDialogOpen,
    isDeleteDialogOpen,
    isImportDialogOpen,
    isSubmitting,
    importing,
    setDialogOpen: setIsDialogOpen,
    setDeleteDialogOpen: setIsDeleteDialogOpen,
    setImportDialogOpen: setIsImportDialogOpen,
    setSubmitting: setIsSubmitting,
    setImporting,
  };
} 