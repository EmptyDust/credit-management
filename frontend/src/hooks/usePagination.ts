import { useState, useCallback } from "react";
import apiClient from "@/lib/api";
import { apiHelpers } from "@/lib/api";
import toast from "react-hot-toast";

interface PaginationState {
  currentPage: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
  loading: boolean;
}

interface PaginationActions {
  setCurrentPage: (page: number) => void;
  setPageSize: (size: number) => void;
  setTotalItems: (total: number) => void;
  setTotalPages: (pages: number) => void;
  setLoading: (loading: boolean) => void;
  handlePageChange: (page: number, fetchFunction: (page: number, size: number) => Promise<void>) => void;
  handlePageSizeChange: (size: number, fetchFunction: (page: number, size: number) => Promise<void>) => void;
  resetToFirstPage: (fetchFunction: (page: number, size: number) => Promise<void>) => void;
  fetchData: <T>(
    endpoint: string,
    params: any,
    setData: (data: T[]) => void,
    errorMessage?: string
  ) => Promise<void>;
}

export function usePagination(initialPageSize: number = 10): PaginationState & PaginationActions {
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(initialPageSize);
  const [totalItems, setTotalItems] = useState(0);
  const [totalPages, setTotalPages] = useState(0);
  const [loading, setLoading] = useState(false);

  const handlePageChange = useCallback((
    page: number, 
    fetchFunction: (page: number, size: number) => Promise<void>
  ) => {
    setCurrentPage(page);
    fetchFunction(page, pageSize);
  }, [pageSize]);

  const handlePageSizeChange = useCallback((
    size: number, 
    fetchFunction: (page: number, size: number) => Promise<void>
  ) => {
    setPageSize(size);
    setCurrentPage(1);
    fetchFunction(1, size);
  }, []);

  const resetToFirstPage = useCallback((
    fetchFunction: (page: number, size: number) => Promise<void>
  ) => {
    setCurrentPage(1);
    fetchFunction(1, pageSize);
  }, [pageSize]);

  const fetchData = useCallback(async <T>(
    endpoint: string,
    params: any,
    setData: (data: T[]) => void,
    errorMessage: string = "获取数据失败"
  ) => {
    try {
      setLoading(true);
      const response = await apiClient.get(endpoint, { params });

      // 使用统一的响应处理函数
      const { data, pagination: paginationData } = apiHelpers.processPaginatedResponse(response);

      const list = Array.isArray(data) ? data : [];
      setData(list);
      setTotalItems(paginationData.total ?? list.length);
      setTotalPages(paginationData.total_pages ?? 1);
    } catch (error) {
      console.error(`Failed to fetch data from ${endpoint}:`, error);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  }, []);

  return {
    currentPage,
    pageSize,
    totalItems,
    totalPages,
    loading,
    setCurrentPage,
    setPageSize,
    setTotalItems,
    setTotalPages,
    setLoading,
    handlePageChange,
    handlePageSizeChange,
    resetToFirstPage,
    fetchData,
  };
} 