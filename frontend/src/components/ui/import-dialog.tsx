import React, { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Upload, Download, RefreshCw } from "lucide-react";
import { validateImportFile } from "@/lib/common-utils";
import { apiHelpers } from "@/lib/api";

interface ImportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  userType: string;
  // 返回错误列表（如果有），用于在弹窗中展示
  onImport: (file: File) => Promise<string[] | null>;
  importing: boolean;
}

export function ImportDialog({
  open,
  onOpenChange,
  title,
  description,
  userType,
  onImport,
  importing,
}: ImportDialogProps) {
  const [importFile, setImportFile] = useState<File | null>(null);
  const [errorLogs, setErrorLogs] = useState<string[]>([]);

  const handleImportFile = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file && validateImportFile(file)) {
      setImportFile(file);
    }
  };

  const handleImport = async () => {
    if (importFile) {
      const errors = await onImport(importFile);
      if (errors && errors.length > 0) {
        setErrorLogs(errors);
      } else {
        setImportFile(null);
        setErrorLogs([]);
        onOpenChange(false);
      }
    }
  };

  const handleClose = () => {
    onOpenChange(false);
    setImportFile(null);
    setErrorLogs([]);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[520px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Upload className="h-5 w-5" />
            {title}
          </DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                apiHelpers.downloadFile(
                  `/users/csv-template?user_type=${userType}`,
                  `${userType}_template.csv`
                );
              }}
            >
              <Download className="mr-2 h-4 w-4" />
              下载CSV模板
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                apiHelpers.downloadFile(
                  `/users/excel-template?user_type=${userType}`,
                  `${userType}_template.xlsx`
                );
              }}
            >
              <Download className="mr-2 h-4 w-4" />
              下载Excel模板
            </Button>
          </div>
          <div>
            <label className="text-sm font-medium">选择文件</label>
            <Input
              type="file"
              accept=".xlsx,.xls,.csv"
              onChange={handleImportFile}
              className="mt-1"
            />
            <p className="text-xs text-muted-foreground mt-1">
              支持Excel (.xlsx, .xls) 和CSV格式，文件大小不超过10MB
            </p>
          </div>
          {importFile && (
            <div className="p-3 bg-muted rounded-lg space-y-1">
              <p className="text-sm font-medium">已选择文件：</p>
              <p className="text-sm text-muted-foreground">
                {importFile.name}
              </p>
              <p className="text-xs text-muted-foreground">
                大小：{(importFile.size / 1024 / 1024).toFixed(2)} MB
              </p>
            </div>
          )}

          {errorLogs.length > 0 && (
            <div className="p-3 border rounded-lg bg-muted/60 max-h-60 overflow-y-auto">
              <p className="text-sm font-medium mb-2">
                导入失败，共 {errorLogs.length} 条错误：
              </p>
              <ul className="space-y-1 text-xs text-red-600">
                {errorLogs.map((msg, idx) => (
                  <li key={idx} className="whitespace-pre-wrap">
                    {msg}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={handleClose}>
            取消
          </Button>
          <Button
            onClick={handleImport}
            disabled={!importFile || importing}
            className="bg-blue-600 hover:bg-blue-700"
          >
            {importing ? (
              <div className="flex items-center gap-2">
                <RefreshCw className="h-4 w-4 animate-spin" />
                导入中...
              </div>
            ) : (
              <>
                <Upload className="mr-2 h-4 w-4" />
                开始导入
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
} 