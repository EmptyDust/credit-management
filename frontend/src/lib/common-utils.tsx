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
    await deleteApiCall(itemToDelete.user_id || itemToDelete.id);
    toast.success(successMessage);
    onSuccess?.();
  } catch (err) {
    toast.error(errorMessage);
  }
};

// 通用导入处理
export const handleImport = async (
  importFile: File | null,
  importApiCall: (formData: FormData) => Promise<any>,
  userType: string,
  onSuccess?: () => void
) => {
  if (!importFile) return;

  try {
    const formData = new FormData();
    formData.append("file", importFile);
    formData.append("user_type", userType);

    const response = await importApiCall(formData);

    if (response.data.code === 0) {
      toast.success("批量导入成功");
      onSuccess?.();
    } else {
      toast.error(response.data.message || "导入失败");
    }
  } catch (err: any) {
    const errorMessage = err.response?.data?.message || "导入失败";
    toast.error(errorMessage);
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