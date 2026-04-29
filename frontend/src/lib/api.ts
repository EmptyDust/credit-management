import { http } from "@/shared/api/http";

const apiClient = http;

export const apiHelpers = {
  uploadFile: async (
    url: string,
    file: File,
    onProgress?: (progress: number) => void
  ) => {
    const formData = new FormData();
    formData.append("file", file);

    return apiClient.post(url, formData, {
      onUploadProgress: (progressEvent) => {
        if (onProgress && progressEvent.total) {
          const progress = Math.round(
            (progressEvent.loaded * 100) / progressEvent.total
          );
          onProgress(progress);
        }
      },
    });
  },

  downloadFile: async (url: string, filename?: string) => {
    const response = await apiClient.get(url, {
      responseType: "blob",
    });

    const blob = new Blob([response.data]);
    const downloadUrl = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = downloadUrl;
    link.download = filename || "download";
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(downloadUrl);
  },

  getPaginated: async (
    url: string,
    page = 1,
    limit = 10,
    filters?: Record<string, string | number | undefined>
  ) => {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
    });

    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined && value !== "") {
          params.set(key, String(value));
        }
      });
    }

    return apiClient.get(`${url}?${params.toString()}`);
  },

  processPaginatedResponse: (response: {
    data?: {
      data?: {
        items?: unknown[];
        data?: unknown[];
        total?: number;
        page?: number;
        page_size?: number;
        total_pages?: number;
        users?: unknown[];
      };
    };
  }) => {
    const payload = response.data?.data;
    const rows = Array.isArray(payload?.items)
      ? payload.items
      : Array.isArray(payload?.data)
        ? payload.data
        : Array.isArray(payload?.users)
          ? payload.users
          : [];

    return {
      data: rows,
      pagination: {
        total: payload?.total || rows.length,
        page: payload?.page || 1,
        page_size: payload?.page_size || rows.length || 10,
        total_pages: payload?.total_pages || 1,
      },
    };
  },
};

export default apiClient;
