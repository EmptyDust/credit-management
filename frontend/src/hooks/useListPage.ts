import { useState, useCallback } from "react";
import { usePagination } from "./usePagination";

interface ListPageOptions<T> {
  endpoint: string;
  setData: (data: T[]) => void;
  errorMessage?: string;
  userType?: "student" | "teacher"; // 添加用户类型参数
}

export function useListPage<T>({ endpoint, setData, errorMessage, userType }: ListPageOptions<T>) {
  const pagination = usePagination(10);
  const {
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
    handlePageChange: paginationHandlePageChange,
    handlePageSizeChange: paginationHandlePageSizeChange,
    resetToFirstPage,
    fetchData,
  } = pagination;
  const [searchQuery, setSearchQuery] = useState("");
  const [filterValue, setFilterValue] = useState<string>("all");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [classFilter, setClassFilter] = useState<string>("all");
  const [gradeFilter, setGradeFilter] = useState<string>("all");

  const fetchList = useCallback(async (page = currentPage, size = pageSize) => {
    const params: any = {
      page,
      page_size: size,
    };

    // 为 /search/users 接口添加必需的 user_type 参数
    if (endpoint === "/search/users" && userType) {
      params.user_type = userType;
    }

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
    if (classFilter !== "all") {
      params.class = classFilter;
    }
    if (gradeFilter !== "all") {
      params.grade = gradeFilter;
    }

    await fetchData(endpoint, params, setData, errorMessage);
  }, [
    currentPage,
    pageSize,
    searchQuery,
    filterValue,
    statusFilter,
    classFilter,
    gradeFilter,
    endpoint,
    setData,
    errorMessage,
    userType,
    fetchData,
  ]);

  const handleSearchAndFilter = useCallback(() => {
    resetToFirstPage(fetchList);
  }, [resetToFirstPage, fetchList]);

  const handlePageChange = useCallback((page: number) => {
    paginationHandlePageChange(page, fetchList);
  }, [paginationHandlePageChange, fetchList]);

  const handlePageSizeChange = useCallback((size: number) => {
    paginationHandlePageSizeChange(size, fetchList);
  }, [paginationHandlePageSizeChange, fetchList]);

  return {
    // 分页相关
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
    classFilter,
    setClassFilter,
    gradeFilter,
    setGradeFilter,
    handleSearchAndFilter,
  };
} 