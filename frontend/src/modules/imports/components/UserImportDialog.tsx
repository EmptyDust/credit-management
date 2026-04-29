import { useMutation } from "@tanstack/react-query";
import { Download, RefreshCw, Upload } from "lucide-react";
import { useMemo, useState } from "react";
import toast from "react-hot-toast";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import {
  confirmUserImport,
  downloadUserImportTemplate,
  precheckUserImport,
  type ImportPrecheckResult,
} from "@/modules/imports/api/user-imports";

interface UserImportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  userType: "student" | "teacher";
  onCompleted?: () => void;
}

export function UserImportDialog({
  open,
  onOpenChange,
  userType,
  onCompleted,
}: UserImportDialogProps) {
  const [file, setFile] = useState<File | null>(null);
  const [precheck, setPrecheck] = useState<ImportPrecheckResult | null>(null);

  const precheckMutation = useMutation({
    mutationFn: (nextFile: File) => precheckUserImport(userType, nextFile),
    onSuccess: (data) => {
      setPrecheck(data);
      toast.success("预校验完成");
    },
  });

  const confirmMutation = useMutation({
    mutationFn: () =>
      confirmUserImport(userType, {
        job_id: precheck?.job_id || "",
        file_hash: precheck?.file_hash || "",
      }),
    onSuccess: (data) => {
      toast.success(`导入完成，新增 ${data.created_user_ids.length} 条记录`);
      setFile(null);
      setPrecheck(null);
      onOpenChange(false);
      onCompleted?.();
    },
  });

  const invalidRows = useMemo(
    () => precheck?.rows.filter((row) => row.status !== "valid") || [],
    [precheck]
  );

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen) {
          setFile(null);
          setPrecheck(null);
        }
        onOpenChange(nextOpen);
      }}
    >
      <DialogContent className="sm:max-w-[640px]">
        <DialogHeader>
          <DialogTitle>{userType === "student" ? "学生导入" : "教师导入"}</DialogTitle>
          <DialogDescription>
            {"统一采用“预校验 -> 明确确认 -> 事务导入”的两步流程。"}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="flex gap-2">
            <Button
              variant="outline"
              className="gap-2"
              onClick={() => void downloadUserImportTemplate(userType)}
            >
              <Download className="h-4 w-4" />
              下载模板
            </Button>
          </div>

          <div className="space-y-2">
            <Input
              type="file"
              accept=".csv"
              onChange={(event) => {
                setFile(event.target.files?.[0] || null);
                setPrecheck(null);
              }}
            />
            <p className="text-xs text-muted-foreground">
              当前后端支持 CSV 模板上传，确认阶段必须使用同一文件。
            </p>
          </div>

          {precheck ? (
            <div className="space-y-3 rounded-lg border p-4">
              <div className="grid gap-3 md:grid-cols-3">
                <div>
                  <div className="text-xs text-muted-foreground">总行数</div>
                  <div className="text-xl font-semibold">{precheck.summary.total_rows}</div>
                </div>
                <div>
                  <div className="text-xs text-muted-foreground">有效行</div>
                  <div className="text-xl font-semibold">{precheck.summary.valid_rows}</div>
                </div>
                <div>
                  <div className="text-xs text-muted-foreground">无效行</div>
                  <div className="text-xl font-semibold">{precheck.summary.invalid_rows}</div>
                </div>
              </div>
              {invalidRows.length ? (
                <div className="max-h-48 space-y-2 overflow-y-auto rounded-md bg-muted/50 p-3 text-sm">
                  {invalidRows.map((row) => (
                    <div key={row.row_no}>
                      <div className="font-medium">第 {row.row_no} 行</div>
                      <ul className="list-disc pl-5 text-muted-foreground">
                        {row.errors.map((error) => (
                          <li key={`${row.row_no}-${error.field}-${error.code}`}>
                            {error.message}
                          </li>
                        ))}
                      </ul>
                    </div>
                  ))}
                </div>
              ) : null}
            </div>
          ) : null}
        </div>

        <DialogFooter className="gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            关闭
          </Button>
          <Button
            variant="secondary"
            className="gap-2"
            disabled={!file || precheckMutation.isPending}
            onClick={() => file && precheckMutation.mutate(file)}
          >
            {precheckMutation.isPending ? (
              <RefreshCw className="h-4 w-4 animate-spin" />
            ) : (
              <Upload className="h-4 w-4" />
            )}
            预校验
          </Button>
          <Button
            className="gap-2"
            disabled={
              !precheck ||
              precheck.summary.invalid_rows > 0 ||
              confirmMutation.isPending
            }
            onClick={() => confirmMutation.mutate()}
          >
            {confirmMutation.isPending ? (
              <RefreshCw className="h-4 w-4 animate-spin" />
            ) : (
              <Upload className="h-4 w-4" />
            )}
            确认导入
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
