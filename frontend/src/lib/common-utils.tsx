import toast from "react-hot-toast";
// 导入统一的状态处理函数
export { getStatusBadge } from "./status-utils";

// 文件导入验证
export const validateImportFile = (file: File): boolean => {
  const allowedTypes = [
    "application/vnd.ms-excel",
    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
    "text/csv",
  ];
  
  if (
    !allowedTypes.includes(file.type) &&
    !file.name.toLowerCase().endsWith(".csv")
  ) {
    toast.error("请选择Excel或CSV文件");
    return false;
  }
  return true;
};

// 文件下载工具
export const downloadFile = (response: any, filename: string) => {
  const url = window.URL.createObjectURL(new Blob([response.data]));
  const link = document.createElement("a");
  link.href = url;
  link.setAttribute("download", filename);
  document.body.appendChild(link);
  link.click();
  link.remove();
  window.URL.revokeObjectURL(url);
};

const resolveItemIdentifier = (item: any): string | undefined => {
  if (!item) return undefined;
  if (typeof item === "string") {
    return item;
  }

  return typeof item.uuid === "string" ? item.uuid : undefined;
};

// 通用删除确认处理
export const handleDeleteConfirm = async (
  itemToDelete: any,
  deleteApiCall: (id: string) => Promise<any>,
  successMessage: string,
  errorMessage: string,
  onSuccess?: () => void
) => {
  if (!itemToDelete) return;

  try {
    const identifier = resolveItemIdentifier(itemToDelete);
    if (!identifier) {
      toast.error("无法找到删除项的唯一标识");
      return;
    }

    await deleteApiCall(identifier);
    toast.success(successMessage);
    onSuccess?.();
  } catch (err) {
    toast.error(errorMessage);
  }
};

// 通用导入处理，返回错误列表（如果有）
export const handleImport = async (
  importFile: File | null,
  importApiCall: (formData: FormData) => Promise<any>,
  userType: string,
  onSuccess?: () => void
): Promise<string[] | null> => {
  if (!importFile) return null;

  try {
    const formData = new FormData();
    formData.append("file", importFile);
    formData.append("user_type", userType);

    const response = await importApiCall(formData);

    if (response.data.code === 0) {
      // 成功：仅在右上角给一个成功提示
      toast.success("批量导入成功");
      onSuccess?.();
      return null;
    }

    // 失败：如果有详细错误列表，则交给弹窗展示，不再弹右上角错误 toast
    const errors = response.data.data?.errors || response.data.errors;
    if (Array.isArray(errors) && errors.length > 0) {
      return errors.map((e: any) =>
        typeof e === "string" ? e : JSON.stringify(e)
      );
    }

    // 没有 errors 时才用 toast 提示一条简短错误
    toast.error(response.data.message || "导入失败");
    return null;
  } catch (err: any) {
    const apiErrors =
      err.response?.data?.data?.errors ||
      err.response?.data?.errors ||
      err.response?.data?.error_details;

    if (Array.isArray(apiErrors) && apiErrors.length > 0) {
      // 有详细错误列表：直接交给中心弹窗展示
      const normalized = apiErrors.map((e: any) =>
        typeof e === "string" ? e : JSON.stringify(e)
      );
      return normalized;
    }

    // 没有详细错误时才使用 toast 作为兜底提示
    const errorMessage =
      err.response?.data?.message ||
      err.response?.data?.error ||
      "导入失败";
    toast.error(errorMessage);
    return null;
  }
};

// 通用导出处理
export const handleExport = async (
  exportApiCall: (params?: any) => Promise<any>,
  filename: string,
  params?: any
) => {
  try {
    const response = await exportApiCall(params);
    downloadFile(response, filename);
    toast.success("导出成功");
  } catch (err) {
    toast.error("导出失败");
  }
}; 