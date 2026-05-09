import { useMutation } from "@tanstack/react-query";
import { RefreshCw, Upload } from "lucide-react";
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
  confirmActivityImport,
  precheckActivityImport,
  type ActivityImportPrecheckResult,
} from "@/modules/activities/api/activity-imports";

interface ActivityImportDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCompleted?: () => void;
}

export default function ActivityImportDialog({
  open,
  onOpenChange,
  onCompleted,
}: ActivityImportDialogProps) {
  const [file, setFile] = useState<File | null>(null);
  const [precheck, setPrecheck] = useState<ActivityImportPrecheckResult | null>(
    null
  );

  const invalidRows = useMemo(
    () => precheck?.rows.filter((row) => row.status !== "valid") || [],
    [precheck]
  );

  const reset = () => {
    setFile(null);
    setPrecheck(null);
  };

  const precheckMutation = useMutation({
    mutationFn: (nextFile: File) => precheckActivityImport(nextFile),
    onSuccess: (result) => {
      setPrecheck(result);
      toast.success("活动导入预校验完成");
    },
    onError: (error) => {
      const message =
        error instanceof Error ? error.message : "活动导入预校验失败";
      toast.error(message);
    },
  });

  const confirmMutation = useMutation({
    mutationFn: () =>
      confirmActivityImport({
        job_id: precheck?.job_id || "",
        file_hash: precheck?.file_hash || "",
      }),
    onSuccess: (result) => {
      toast.success(`导入完成，新增 ${result.created_activity_ids.length} 个活动`);
      reset();
      onOpenChange(false);
      onCompleted?.();
    },
    onError: (error) => {
      const message =
        error instanceof Error ? error.message : "活动导入确认失败";
      toast.error(message);
    },
  });

  return (
    <Dialog
      open={open}
      onOpenChange={(nextOpen) => {
        if (!nextOpen) {
          reset();
        }
        onOpenChange(nextOpen);
      }}
    >
      <DialogContent className="sm:max-w-[720px]">
        <DialogHeader>
          <DialogTitle>活动导入</DialogTitle>
          <DialogDescription>
            {"活动导入统一采用“上传文件 -> 预校验 -> 确认导入”的正式流程。"}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div className="space-y-2">
            <label className="text-sm font-medium" htmlFor="activity-import-file">
              选择文件
            </label>
            <Input
              id="activity-import-file"
              accept=".csv"
              type="file"
              onChange={(event) => {
                setFile(event.target.files?.[0] || null);
                setPrecheck(null);
              }}
            />
            <p className="text-xs text-muted-foreground">
              当前后端支持 CSV 文件；确认导入会绑定同一文件哈希并保持幂等。
            </p>
          </div>

          {file ? (
            <div className="rounded-lg border bg-muted/40 p-3 text-sm">
              <div className="font-medium">已选择文件</div>
              <div className="mt-1 text-muted-foreground">{file.name}</div>
            </div>
          ) : null}

          {precheck ? (
            <div className="space-y-4 rounded-lg border p-4">
              <div className="text-sm font-medium">预校验完成</div>
              <div className="grid gap-3 md:grid-cols-3">
                <div>
                  <div className="text-xs text-muted-foreground">总行数</div>
                  <div className="text-2xl font-semibold">
                    {precheck.summary.total_rows}
                  </div>
                </div>
                <div>
                  <div className="text-xs text-muted-foreground">有效行</div>
                  <div className="text-2xl font-semibold">
                    {precheck.summary.valid_rows}
                  </div>
                </div>
                <div>
                  <div className="text-xs text-muted-foreground">无效行</div>
                  <div className="text-2xl font-semibold">
                    {precheck.summary.invalid_rows}
                  </div>
                </div>
              </div>

              {invalidRows.length > 0 ? (
                <div className="max-h-56 space-y-3 overflow-y-auto rounded-md bg-muted/50 p-3 text-sm">
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
            className="gap-2"
            variant="secondary"
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
