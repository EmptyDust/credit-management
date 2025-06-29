import { useState, useCallback } from "react";
import { usePagination } from "./usePagination";

interface ListPageOptions<T> {
  endpoint: string;
  setData: (data: T[]) => void;
  errorMessage?: string;
}

export function useListPage<T>({ endpoint, setData, errorMessage }: ListPageOptions<T>) {
  const pagination = usePagination(10);
  const [searchQuery, setSearchQuery] = useState("");
  const [filterValue, setFilterValue] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  const fetchList = useCallback(async (page = pagination.currentPage, size = pagination.pageSize) => {
    const params: any = {
      page,
      page_size: size,
    };

    if (searchQuery) {
      params.query = searchQuery;
    }
    if (filterValue !== "all") {
      // 根据不同的页面设置不同的过滤字段名
      if (endpoint.includes("users")) {
        params.college = filterValue;
      } else if (endpoint.includes("activities")) {
        params.category = filterValue;
      }
    }
    if (statusFilter !== "all") {
      params.status = statusFilter;
    }

    await pagination.fetchData(endpoint, params, setData, errorMessage);
  }, [pagination, searchQuery, filterValue, statusFilter, endpoint, setData, errorMessage]);

  const handleSearchAndFilter = useCallback(() => {
    pagination.resetToFirstPage(fetchList);
  }, [pagination, fetchList]);

  const handlePageChange = useCallback((page: number) => {
    pagination.handlePageChange(page, fetchList);
  }, [pagination, fetchList]);

  const handlePageSizeChange = useCallback((size: number) => {
    pagination.handlePageSizeChange(size, fetchList);
  }, [pagination, fetchList]);

  return {
    // 分页相关
    ...pagination,
    fetchList,
    handlePageChange,
    handlePageSizeChange,
    
    // 搜索和过滤
    searchQuery,
    setSearchQuery,
    filterValue,
    setFilterValue,
    statusFilter,
    setStatusFilter,
    handleSearchAndFilter,
  };
} 