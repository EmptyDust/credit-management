import { useState, useCallback } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import toast from "react-hot-toast";
import apiClient from "@/lib/api";
import { 
  handleDeleteConfirm as handleDeleteConfirmUtil,
  handleImport as handleImportUtil,
  handleExport as handleExportUtil
} from "@/lib/common-utils";

interface UserManagementOptions {
  userType: string;
  formSchema: z.ZodSchema<any>;
  defaultValues: any;
  fetchFunction: (page: number, size: number) => Promise<void>;
  onSuccess?: () => void;
}

export function useUserManagement<T extends { uuid?: string; real_name?: string }>({
  userType,
  formSchema,
  defaultValues,
  fetchFunction,
  onSuccess,
}: UserManagementOptions) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [searchQuery, setSearchQuery] = useState("");
  const [filterValue, setFilterValue] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<T | null>(null);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [itemToDelete, setItemToDelete] = useState<T | null>(null);
  const [isImportDialogOpen, setIsImportDialogOpen] = useState(false);
  const [importing, setImporting] = useState(false);

  const getUserIdentifier = useCallback((item: T | null) => {
    if (!item) return undefined;
    return typeof (item as any).uuid === "string" ? (item as any).uuid : undefined;
  }, []);

  const form = useForm({
    resolver: zodResolver(formSchema as any),
    defaultValues,
  });

  const handleDialogOpen = useCallback((item: T | null) => {
    setEditingItem(item);
    if (item) {
      form.reset(item);
    } else {
      form.reset(defaultValues);
    }
    setIsDialogOpen(true);
  }, [form, defaultValues]);

  const onSubmit = useCallback(async (values: z.infer<typeof formSchema>) => {
    setIsSubmitting(true);
    try {
      if (editingItem) {
        const userIdentifier = getUserIdentifier(editingItem);
        if (!userIdentifier) {
          toast.error("无法找到用户UUID");
          return;
        }
        await apiClient.put(`/users/${userIdentifier}`, values);
        toast.success("用户信息更新成功");
      } else {
        const createData = {
          ...values,
          password: values.password || "Password123",
          user_type: userType,
        };
        await apiClient.post(`/users/${userType}s`, createData);
        toast.success("用户创建成功");
      }
      setIsDialogOpen(false);
      fetchFunction(1, 10);
      onSuccess?.();
    } catch (err: any) {
      if (!err.response || err.response.status !== 409) {
        toast.error(`用户${editingItem ? "更新" : "创建"}失败`);
      }
      console.error(err);
    } finally {
      setIsSubmitting(false);
    }
  }, [editingItem, userType, fetchFunction, onSuccess, getUserIdentifier]);

  const handleDeleteConfirm = useCallback(async () => {
    await handleDeleteConfirmUtil(
      itemToDelete,
      (id) => apiClient.delete(`/users/${id}`),
      "用户删除成功",
      "删除用户失败",
      () => {
        setDeleteDialogOpen(false);
        setItemToDelete(null);
        fetchFunction(1, 10);
      }
    );
  }, [itemToDelete, fetchFunction]);

  const handleImport = useCallback(async (file: File) => {
    setImporting(true);
    await handleImportUtil(
      file,
      (formData) => apiClient.post("/users/import", formData, {
        headers: {
          // Remove Content-Type header to let browser set it with boundary
        },
      }),
      userType,
      () => {
        setIsImportDialogOpen(false);
        fetchFunction(1, 10);
      }
    );
    setImporting(false);
  }, [userType, fetchFunction]);

  const handleExport = useCallback(async () => {
    await handleExportUtil(
      (params) => apiClient.get("/users/export", {
        params: { user_type: userType, ...params },
        responseType: "blob",
      }),
      `${userType}s_${new Date().toISOString().split("T")[0]}.xlsx`
    );
  }, [userType]);

  const handleSearchAndFilter = useCallback(() => {
    fetchFunction(1, 10);
  }, [fetchFunction]);

  return {
    // State
    loading,
    setLoading,
    error,
    setError,
    searchQuery,
    setSearchQuery,
    filterValue,
    setFilterValue,
    statusFilter,
    setStatusFilter,
    isSubmitting,
    isDialogOpen,
    editingItem,
    deleteDialogOpen,
    itemToDelete,
    isImportDialogOpen,
    importing,
    
    // Form
    form,
    
    // Actions
    handleDialogOpen,
    onSubmit,
    handleDeleteConfirm,
    handleImport,
    handleExport,
    handleSearchAndFilter,
    
    // Setters
    setIsDialogOpen,
    setDeleteDialogOpen,
    setItemToDelete,
    setIsImportDialogOpen,
  };
} 