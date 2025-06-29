import { TableCell, TableRow } from "@/components/ui/table";
import { RefreshCw, AlertCircle } from "lucide-react";

interface TableLoadingStatesProps {
  loading: boolean;
  error?: string;
  dataLength: number;
  colSpan: number;
  emptyMessage?: string;
}

export function TableLoadingStates({
  loading,
  error,
  dataLength,
  colSpan,
  emptyMessage = "暂无数据",
}: TableLoadingStatesProps) {
  if (loading) {
    return (
      <TableRow>
        <TableCell colSpan={colSpan} className="text-center py-8">
          <div className="flex items-center justify-center gap-2">
            <RefreshCw className="h-4 w-4 animate-spin" />
            加载中...
          </div>
        </TableCell>
      </TableRow>
    );
  }

  if (error) {
    return (
      <TableRow>
        <TableCell colSpan={colSpan} className="text-center py-8 text-red-500">
          <div className="flex items-center justify-center gap-2">
            <AlertCircle className="h-4 w-4" />
            {error}
          </div>
        </TableCell>
      </TableRow>
    );
  }

  if (dataLength === 0) {
    return (
      <TableRow>
        <TableCell colSpan={colSpan} className="text-center py-8">
          <div className="flex flex-col items-center gap-2 text-muted-foreground">
            <AlertCircle className="w-8 h-8" />
            <p>{emptyMessage}</p>
          </div>
        </TableCell>
      </TableRow>
    );
  }

  return null;
} 