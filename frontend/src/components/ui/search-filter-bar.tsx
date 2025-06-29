import { Search, Filter, RefreshCw } from "lucide-react";
import { Input } from "./input";
import { Button } from "./button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./select";

interface SearchFilterBarProps {
  searchQuery: string;
  onSearchChange: (value: string) => void;
  onSearch: () => void;
  onRefresh: () => void;
  filterOptions?: Array<{ value: string; label: string }>;
  filterValue: string;
  onFilterChange: (value: string) => void;
  filterPlaceholder?: string;
  searchPlaceholder?: string;
  showRefresh?: boolean;
  className?: string;
}

export function SearchFilterBar({
  searchQuery,
  onSearchChange,
  onSearch,
  onRefresh,
  filterOptions = [],
  filterValue,
  onFilterChange,
  filterPlaceholder = "选择过滤条件",
  searchPlaceholder = "搜索...",
  showRefresh = true,
  className = "",
}: SearchFilterBarProps) {
  return (
    <div className={`flex items-center gap-4 ${className}`}>
      {/* 搜索框 */}
      <div className="relative flex-1 max-w-md">
        <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
        <Input
          placeholder={searchPlaceholder}
          value={searchQuery}
          onChange={(e) => onSearchChange(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && onSearch()}
          className="pl-10"
        />
      </div>

      {/* 过滤选择器 */}
      {filterOptions.length > 0 && (
        <Select value={filterValue} onValueChange={onFilterChange}>
          <SelectTrigger className="w-48">
            <SelectValue placeholder={filterPlaceholder} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">全部</SelectItem>
            {filterOptions.map((option) => (
              <SelectItem key={option.value} value={option.value}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}

      {/* 搜索按钮 */}
      <Button onClick={onSearch} variant="outline">
        <Search className="h-4 w-4 mr-2" />
        搜索
      </Button>

      {/* 刷新按钮 */}
      {showRefresh && (
        <Button onClick={onRefresh} variant="outline">
          <RefreshCw className="h-4 w-4" />
        </Button>
      )}
    </div>
  );
} 