import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  ArrowLeft,
  Copy,
  Download,
  FileUp,
  RefreshCw,
  Send,
  Trash2,
  UserPlus,
  Users,
} from "lucide-react";
import { startTransition, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import toast from "react-hot-toast";

import { useAuth } from "@/contexts/AuthContext";
import {
  addActivityParticipants,
  copyActivity,
  deleteActivity,
  deleteActivityFile,
  downloadActivityFile,
  getActivityDetail,
  leaveActivity,
  listActivityFiles,
  listActivityParticipants,
  reviewActivity,
  submitActivity,
  transferActivityOwner,
  updateActivity,
  uploadActivityFile,
  withdrawActivity,
  type ActivityDetail,
  type ActivityFile,
  type ActivityParticipant,
  type ActivityMutationInput,
} from "@/modules/activities/api/activities";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";

type ActivityFormState = {
  title: string;
  description: string;
  category: string;
  academic_year: string;
  semester: string;
  start_date: string;
  end_date: string;
  details_text: string;
};

function statusLabel(status: string) {
  switch (status) {
    case "draft":
      return "草稿";
    case "pending_review":
      return "审核中";
    case "approved":
      return "已通过";
    case "rejected":
      return "已拒绝";
    default:
      return status;
  }
}

function createFormState(activity: ActivityDetail): ActivityFormState {
  return {
    title: activity.title || "",
    description: activity.description || "",
    category: activity.category || "",
    academic_year: activity.academic_year || "",
    semester: activity.semester || "",
    start_date: activity.start_date || "",
    end_date: activity.end_date || "",
    details_text: JSON.stringify(activity.details || {}, null, 2),
  };
}

function parseFormState(form: ActivityFormState): ActivityMutationInput {
  return {
    title: form.title,
    description: form.description || undefined,
    category: form.category,
    academic_year: form.academic_year,
    semester: form.semester,
    start_date: form.start_date || undefined,
    end_date: form.end_date || undefined,
    details: form.details_text.trim()
      ? (JSON.parse(form.details_text) as Record<string, unknown>)
      : {},
  };
}

function formatBytes(size: number) {
  if (size < 1024) {
    return `${size} B`;
  }
  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`;
  }

  return `${(size / (1024 * 1024)).toFixed(1)} MB`;
}

function DetailsBlock({ activity }: { activity: ActivityDetail }) {
  const details = activity.details || {};
  const entries = Object.entries(details);

  return (
    <Card>
      <CardHeader>
        <CardTitle>活动信息</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid gap-4 md:grid-cols-2">
          <div>
            <div className="text-sm text-muted-foreground">分类</div>
            <div className="font-medium">{activity.category || "-"}</div>
          </div>
          <div>
            <div className="text-sm text-muted-foreground">学年 / 学期</div>
            <div className="font-medium">
              {activity.academic_year || "-"} / {activity.semester || "-"}
            </div>
          </div>
          <div>
            <div className="text-sm text-muted-foreground">开始日期</div>
            <div className="font-medium">{activity.start_date || "-"}</div>
          </div>
          <div>
            <div className="text-sm text-muted-foreground">结束日期</div>
            <div className="font-medium">{activity.end_date || "-"}</div>
          </div>
        </div>
        <div>
          <div className="text-sm text-muted-foreground">描述</div>
          <div className="mt-1 whitespace-pre-wrap text-sm">
            {activity.description || "暂无描述"}
          </div>
        </div>
        {activity.review_comments ? (
          <div>
            <div className="text-sm text-muted-foreground">最近审核意见</div>
            <div className="mt-1 whitespace-pre-wrap text-sm">
              {activity.review_comments}
            </div>
          </div>
        ) : null}
        <div>
          <div className="text-sm text-muted-foreground">扩展字段</div>
          {entries.length === 0 ? (
            <div className="mt-1 text-sm text-muted-foreground">暂无扩展字段</div>
          ) : (
            <div className="mt-2 grid gap-2 md:grid-cols-2">
              {entries.map(([key, value]) => (
                <div key={key} className="rounded-lg border p-3">
                  <div className="text-xs uppercase tracking-wide text-muted-foreground">
                    {key}
                  </div>
                  <div className="mt-1 text-sm">
                    {typeof value === "string" ? value : JSON.stringify(value)}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function ParticipantsBlock({
  participants,
  canManage,
  canLeave,
  onAdd,
  onLeave,
}: {
  participants: ActivityParticipant[];
  canManage: boolean;
  canLeave: boolean;
  onAdd: (userIDs: string[], credits: number) => Promise<void>;
  onLeave: () => Promise<void>;
}) {
  const [userIDsText, setUserIDsText] = useState("");
  const [creditsText, setCreditsText] = useState("0");

  const handleAdd = async () => {
    const userIDs = userIDsText
      .split(/[,\n]/)
      .map((value) => value.trim())
      .filter(Boolean);
    if (userIDs.length === 0) {
      toast.error("请填写至少一个学生 ID");
      return;
    }

    await onAdd(userIDs, Number(creditsText || "0"));
    setUserIDsText("");
  };

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between">
        <CardTitle className="flex items-center gap-2">
          <Users className="h-4 w-4" />
          参与者
        </CardTitle>
        {canLeave ? (
          <Button variant="outline" size="sm" onClick={() => void onLeave()}>
            退出活动
          </Button>
        ) : null}
      </CardHeader>
      <CardContent className="space-y-4">
        {canManage ? (
          <div className="grid gap-3 rounded-lg border p-4 md:grid-cols-[1fr_160px_auto]">
            <div className="space-y-2">
              <Label htmlFor="participant-ids">学生 ID</Label>
              <Textarea
                id="participant-ids"
                value={userIDsText}
                onChange={(event) => setUserIDsText(event.target.value)}
                placeholder="使用逗号或换行分隔多个学生 ID"
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="participant-credits">学分</Label>
              <Input
                id="participant-credits"
                type="number"
                min="0"
                step="0.5"
                value={creditsText}
                onChange={(event) => setCreditsText(event.target.value)}
              />
            </div>
            <div className="flex items-end">
              <Button className="gap-2" onClick={() => void handleAdd()}>
                <UserPlus className="h-4 w-4" />
                添加参与者
              </Button>
            </div>
          </div>
        ) : null}

        {participants.length === 0 ? (
          <div className="text-sm text-muted-foreground">暂无参与者</div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>用户 ID</TableHead>
                <TableHead>学分</TableHead>
                <TableHead>加入时间</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {participants.map((participant) => (
                <TableRow key={participant.id}>
                  <TableCell>{participant.user_id}</TableCell>
                  <TableCell>{participant.credits}</TableCell>
                  <TableCell>{participant.joined_at || "-"}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        )}
      </CardContent>
    </Card>
  );
}

function FilesBlock({
  files,
  canManage,
  onUpload,
  onDelete,
  onDownload,
}: {
  files: ActivityFile[];
  canManage: boolean;
  onUpload: (file: File) => Promise<void>;
  onDelete: (fileID: string) => Promise<void>;
  onDownload: (fileID: string, filename?: string) => Promise<void>;
}) {
  const [selectedFile, setSelectedFile] = useState<File | null>(null);

  const handleUpload = async () => {
    if (!selectedFile) {
      toast.error("请选择文件");
      return;
    }

    await onUpload(selectedFile);
    setSelectedFile(null);
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>附件</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {canManage ? (
          <div className="grid gap-3 rounded-lg border p-4 md:grid-cols-[1fr_auto]">
            <Input
              type="file"
              onChange={(event) =>
                setSelectedFile(event.target.files?.[0] ?? null)
              }
            />
            <Button className="gap-2" onClick={() => void handleUpload()}>
              <FileUp className="h-4 w-4" />
              上传附件
            </Button>
          </div>
        ) : null}

        {files.length === 0 ? (
          <div className="text-sm text-muted-foreground">暂无附件</div>
        ) : (
          <div className="space-y-3">
            {files.map((file) => (
              <div
                key={file.id}
                className="flex flex-col gap-3 rounded-lg border p-4 md:flex-row md:items-center md:justify-between"
              >
                <div>
                  <div className="font-medium">{file.name}</div>
                  <div className="text-sm text-muted-foreground">
                    {formatBytes(file.size)} · {file.content_type || "unknown"}
                  </div>
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    className="gap-2"
                    onClick={() => void onDownload(file.id, file.name)}
                  >
                    <Download className="h-4 w-4" />
                    下载
                  </Button>
                  {canManage ? (
                    <Button
                      variant="outline"
                      size="sm"
                      className="gap-2"
                      onClick={() => void onDelete(file.id)}
                    >
                      <Trash2 className="h-4 w-4" />
                      删除
                    </Button>
                  ) : null}
                </div>
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  );
}

export default function ActivityDetailPage() {
  const { id = "" } = useParams();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user } = useAuth();

  const [isEditOpen, setIsEditOpen] = useState(false);
  const [isReviewOpen, setIsReviewOpen] = useState(false);
  const [isTransferOpen, setIsTransferOpen] = useState(false);
  const [reviewStatus, setReviewStatus] = useState<"approved" | "rejected">(
    "approved"
  );
  const [reviewComments, setReviewComments] = useState("");
  const [newOwnerID, setNewOwnerID] = useState("");
  const [formState, setFormState] = useState<ActivityFormState>({
    title: "",
    description: "",
    category: "",
    academic_year: "",
    semester: "",
    start_date: "",
    end_date: "",
    details_text: "{}",
  });

  const activityQuery = useQuery({
    queryKey: ["activities", "detail", id],
    queryFn: () => getActivityDetail(id),
    enabled: Boolean(id),
  });

  const participantsQuery = useQuery({
    queryKey: ["activities", "participants", id],
    queryFn: () => listActivityParticipants(id),
    enabled: Boolean(id),
  });

  const filesQuery = useQuery({
    queryKey: ["activities", "files", id],
    queryFn: () => listActivityFiles(id),
    enabled: Boolean(id),
  });

  const activity = activityQuery.data;

  useEffect(() => {
    if (!activity) {
      return;
    }
    startTransition(() => {
      setFormState(createFormState(activity));
    });
  }, [activity]);

  const actorRole = user?.userType;
  const isOwner = Boolean(user && activity && user.id === activity.owner_id);
  const isTeacherOrAdmin = actorRole === "teacher" || actorRole === "admin";
  const canEdit = Boolean(activity && activity.status === "draft" && (isOwner || isTeacherOrAdmin));
  const canSubmit = Boolean(activity && isOwner && activity.status === "draft");
  const canWithdraw = Boolean(
    activity &&
      isOwner &&
      (activity.status === "pending_review" || activity.status === "rejected")
  );
  const canReview = Boolean(
    activity &&
      isTeacherOrAdmin &&
      ["pending_review", "approved", "rejected"].includes(activity.status)
  );
  const canTransfer = Boolean(activity && isTeacherOrAdmin);
  const canDelete = Boolean(
    activity && (isTeacherOrAdmin || (isOwner && activity.status === "draft"))
  );
  const canManageParticipants = canEdit;
  const canManageFiles = canEdit;
  const canLeave = Boolean(activity && actorRole === "student" && activity.status === "draft");

  const invalidateAll = async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["activities", "detail", id] }),
      queryClient.invalidateQueries({ queryKey: ["activities", "participants", id] }),
      queryClient.invalidateQueries({ queryKey: ["activities", "files", id] }),
      queryClient.invalidateQueries({ queryKey: ["activities"] }),
      queryClient.invalidateQueries({ queryKey: ["applications"] }),
    ]);
  };

  const actionMutation = useMutation({
    mutationFn: async (
      action:
        | { type: "submit" }
        | { type: "withdraw" }
        | { type: "copy" }
        | { type: "delete" }
        | { type: "leave" }
        | { type: "review"; target_status: "approved" | "rejected"; review_comments: string }
        | { type: "transfer"; newOwnerID: string }
        | { type: "update"; payload: ActivityMutationInput }
        | { type: "addParticipants"; userIDs: string[]; credits: number }
        | { type: "uploadFile"; file: File }
        | { type: "deleteFile"; fileID: string }
    ) => {
      switch (action.type) {
        case "submit":
          return submitActivity(id);
        case "withdraw":
          return withdrawActivity(id);
        case "copy":
          return copyActivity(id);
        case "delete":
          await deleteActivity(id);
          return null;
        case "leave":
          await leaveActivity(id);
          return null;
        case "review":
          return reviewActivity(id, {
            target_status: action.target_status,
            review_comments: action.review_comments,
          });
        case "transfer":
          return transferActivityOwner(id, action.newOwnerID);
        case "update":
          return updateActivity(id, action.payload);
        case "addParticipants":
          return addActivityParticipants(id, {
            user_ids: action.userIDs,
            credits: action.credits,
          });
        case "uploadFile":
          return uploadActivityFile(id, action.file);
        case "deleteFile":
          await deleteActivityFile(id, action.fileID);
          return null;
        default:
          return null;
      }
    },
    onSuccess: async (_, action) => {
      await invalidateAll();
      switch (action.type) {
        case "submit":
          toast.success("活动已提交审核");
          break;
        case "withdraw":
          toast.success("活动已撤回到草稿");
          break;
        case "copy":
          toast.success("活动已复制为新草稿");
          break;
        case "delete":
          toast.success("活动已删除");
          navigate("/activities");
          return;
        case "leave":
          toast.success("已退出活动");
          break;
        case "review":
          toast.success("审核状态已更新");
          setIsReviewOpen(false);
          setReviewComments("");
          break;
        case "transfer":
          toast.success("活动拥有者已转移");
          setIsTransferOpen(false);
          setNewOwnerID("");
          break;
        case "update":
          toast.success("活动已更新");
          setIsEditOpen(false);
          break;
        case "addParticipants":
          toast.success("参与者已添加");
          break;
        case "uploadFile":
          toast.success("附件已上传");
          break;
        case "deleteFile":
          toast.success("附件已删除");
          break;
      }
    },
    onError: (error) => {
      const message =
        error instanceof Error ? error.message : "操作失败，请稍后重试";
      toast.error(message);
    },
  });

  const isBusy = actionMutation.isPending;

  const pageTitle = useMemo(() => activity?.title || "活动详情", [activity]);

  if (activityQuery.isLoading) {
    return (
      <div className="flex min-h-[320px] items-center justify-center">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <RefreshCw className="h-4 w-4 animate-spin" />
          正在加载活动详情...
        </div>
      </div>
    );
  }

  if (activityQuery.isError || !activity) {
    return (
      <Card>
        <CardContent className="py-12 text-center">
          <div className="text-lg font-medium">活动详情加载失败</div>
          <div className="mt-2 text-sm text-muted-foreground">
            当前活动不存在，或你没有访问权限。
          </div>
          <Button className="mt-6" onClick={() => navigate("/activities")}>
            返回活动列表
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <Card className="border-none bg-gradient-to-r from-slate-900 via-slate-800 to-slate-900 text-slate-50 shadow-xl">
        <CardHeader className="gap-4">
          <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
            <div className="space-y-3">
              <Button
                variant="ghost"
                className="h-auto justify-start px-0 text-slate-200 hover:bg-transparent hover:text-white"
                onClick={() => navigate("/activities")}
              >
                <ArrowLeft className="mr-2 h-4 w-4" />
                返回活动列表
              </Button>
              <div>
                <h1 className="text-3xl font-semibold">{pageTitle}</h1>
                <p className="mt-2 max-w-2xl text-sm text-slate-300">
                  模块化详情页直接面向 `/api/v1/activities/*` 契约，承接编辑、审核、转移拥有者、参与者和附件操作。
                </p>
              </div>
            </div>
            <div className="flex flex-wrap gap-2">
              <Badge variant="secondary" className="bg-white/15 text-white">
                {statusLabel(activity.status)}
              </Badge>
              <Badge variant="outline" className="border-slate-600 text-slate-100">
                {activity.category || "未分类"}
              </Badge>
            </div>
          </div>
          <div className="flex flex-wrap gap-2">
            {canEdit ? (
              <Button variant="secondary" disabled={isBusy} onClick={() => setIsEditOpen(true)}>
                编辑草稿
              </Button>
            ) : null}
            {canSubmit ? (
              <Button className="gap-2" disabled={isBusy} onClick={() => actionMutation.mutate({ type: "submit" })}>
                <Send className="h-4 w-4" />
                提交审核
              </Button>
            ) : null}
            {canWithdraw ? (
              <Button variant="outline" disabled={isBusy} onClick={() => actionMutation.mutate({ type: "withdraw" })}>
                撤回到草稿
              </Button>
            ) : null}
            {canReview ? (
              <Button variant="outline" disabled={isBusy} onClick={() => setIsReviewOpen(true)}>
                审核活动
              </Button>
            ) : null}
            {canTransfer ? (
              <Button variant="outline" disabled={isBusy} onClick={() => setIsTransferOpen(true)}>
                转移拥有者
              </Button>
            ) : null}
            <Button
              variant="outline"
              disabled={isBusy}
              onClick={() => actionMutation.mutate({ type: "copy" })}
            >
              <Copy className="mr-2 h-4 w-4" />
              复制草稿
            </Button>
            {canDelete ? (
              <Button
                variant="destructive"
                disabled={isBusy}
                onClick={() => actionMutation.mutate({ type: "delete" })}
              >
                <Trash2 className="mr-2 h-4 w-4" />
                删除
              </Button>
            ) : null}
          </div>
        </CardHeader>
      </Card>

      <DetailsBlock activity={activity} />

      <ParticipantsBlock
        participants={participantsQuery.data || []}
        canManage={canManageParticipants}
        canLeave={canLeave}
        onAdd={async (userIDs, credits) =>
          actionMutation.mutateAsync({ type: "addParticipants", userIDs, credits }).then(() => undefined)
        }
        onLeave={async () =>
          actionMutation.mutateAsync({ type: "leave" }).then(() => undefined)
        }
      />

      <FilesBlock
        files={filesQuery.data || []}
        canManage={canManageFiles}
        onUpload={async (file) =>
          actionMutation.mutateAsync({ type: "uploadFile", file }).then(() => undefined)
        }
        onDelete={async (fileID) =>
          actionMutation.mutateAsync({ type: "deleteFile", fileID }).then(() => undefined)
        }
        onDownload={async (fileID, filename) => {
          await downloadActivityFile(fileID, filename);
        }}
      />

      <Dialog open={isEditOpen} onOpenChange={setIsEditOpen}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle>编辑活动草稿</DialogTitle>
          </DialogHeader>
          <div className="grid gap-4 py-2 md:grid-cols-2">
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="activity-title">标题</Label>
              <Input
                id="activity-title"
                value={formState.title}
                onChange={(event) =>
                  setFormState((current) => ({ ...current, title: event.target.value }))
                }
              />
            </div>
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="activity-description">描述</Label>
              <Textarea
                id="activity-description"
                value={formState.description}
                onChange={(event) =>
                  setFormState((current) => ({
                    ...current,
                    description: event.target.value,
                  }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="activity-category">分类</Label>
              <Input
                id="activity-category"
                value={formState.category}
                onChange={(event) =>
                  setFormState((current) => ({ ...current, category: event.target.value }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="activity-academic-year">学年</Label>
              <Input
                id="activity-academic-year"
                value={formState.academic_year}
                onChange={(event) =>
                  setFormState((current) => ({
                    ...current,
                    academic_year: event.target.value,
                  }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="activity-semester">学期</Label>
              <Input
                id="activity-semester"
                value={formState.semester}
                onChange={(event) =>
                  setFormState((current) => ({ ...current, semester: event.target.value }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="activity-start-date">开始日期</Label>
              <Input
                id="activity-start-date"
                type="date"
                value={formState.start_date}
                onChange={(event) =>
                  setFormState((current) => ({
                    ...current,
                    start_date: event.target.value,
                  }))
                }
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="activity-end-date">结束日期</Label>
              <Input
                id="activity-end-date"
                type="date"
                value={formState.end_date}
                onChange={(event) =>
                  setFormState((current) => ({
                    ...current,
                    end_date: event.target.value,
                  }))
                }
              />
            </div>
            <div className="space-y-2 md:col-span-2">
              <Label htmlFor="activity-details">扩展字段 JSON</Label>
              <Textarea
                id="activity-details"
                rows={8}
                value={formState.details_text}
                onChange={(event) =>
                  setFormState((current) => ({
                    ...current,
                    details_text: event.target.value,
                  }))
                }
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsEditOpen(false)}>
              取消
            </Button>
            <Button
              disabled={isBusy}
              onClick={() => {
                try {
                  const payload = parseFormState(formState);
                  actionMutation.mutate({ type: "update", payload });
                } catch {
                  toast.error("扩展字段必须是合法 JSON");
                }
              }}
            >
              保存活动
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={isReviewOpen} onOpenChange={setIsReviewOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>审核活动</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label htmlFor="review-status">目标状态</Label>
              <select
                id="review-status"
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={reviewStatus}
                onChange={(event) =>
                  setReviewStatus(event.target.value as "approved" | "rejected")
                }
              >
                <option value="approved">approved</option>
                <option value="rejected">rejected</option>
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="review-comments">审核意见</Label>
              <Textarea
                id="review-comments"
                value={reviewComments}
                onChange={(event) => setReviewComments(event.target.value)}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsReviewOpen(false)}>
              取消
            </Button>
            <Button
              disabled={isBusy}
              onClick={() =>
                actionMutation.mutate({
                  type: "review",
                  target_status: reviewStatus,
                  review_comments: reviewComments,
                })
              }
            >
              提交审核结果
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={isTransferOpen} onOpenChange={setIsTransferOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>转移活动拥有者</DialogTitle>
          </DialogHeader>
          <div className="space-y-2 py-2">
            <Label htmlFor="new-owner-id">新拥有者用户 ID</Label>
            <Input
              id="new-owner-id"
              value={newOwnerID}
              onChange={(event) => setNewOwnerID(event.target.value)}
              placeholder="输入新的用户 ID"
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsTransferOpen(false)}>
              取消
            </Button>
            <Button
              disabled={isBusy}
              onClick={() =>
                actionMutation.mutate({
                  type: "transfer",
                  newOwnerID: newOwnerID.trim(),
                })
              }
            >
              确认转移
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
