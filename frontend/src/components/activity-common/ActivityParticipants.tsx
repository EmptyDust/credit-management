import { useState, useEffect } from "react";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Users,
  User,
  Plus,
  Search,
  Trash2,
  Edit3,
  Download,
  BarChart3,
  MoreHorizontal,
  Check,
  X,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { useAuth } from "@/contexts/AuthContext";
import apiClient from "@/lib/api";
import toast from "react-hot-toast";
import type { Activity, Participant, UserInfo } from "@/types/activity";

interface ActivityParticipantsProps {
  activity: Activity;
  onRefresh?: () => void;
}

interface ParticipantWithUserInfo extends Participant {
  user_info?: UserInfo & {
    college?: string;
    major?: string;
    class?: string;
  };
}

interface UserSearchResult {
  user_id: string;
  username: string;
  real_name: string;
  student_id?: string;
  college?: string;
  major?: string;
  class?: string;
}

export default function ActivityParticipants({
  activity,
  onRefresh,
}: ActivityParticipantsProps) {
  const { user } = useAuth();
  const [participants, setParticipants] = useState<ParticipantWithUserInfo[]>(
    []
  );
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [showStatsDialog, setShowStatsDialog] = useState(false);
  const [showBatchDialog, setShowBatchDialog] = useState(false);
  const [selectedParticipants, setSelectedParticipants] = useState<string[]>(
    []
  );
  const [userSearchResults, setUserSearchResults] = useState<
    UserSearchResult[]
  >([]);
  const [userSearchLoading, setUserSearchLoading] = useState(false);
  const [userSearchQuery, setUserSearchQuery] = useState("");
  const [selectedUsers, setSelectedUsers] = useState<string[]>([]);
  const [addDialogCredits, setAddDialogCredits] = useState(1.0);
  const [batchDialogCredits, setBatchDialogCredits] = useState(1.0);
  const [editingCredits, setEditingCredits] = useState<{
    [key: string]: number | undefined;
  }>({});
  const [stats, setStats] = useState<any>(null);

  const isOwner =
    user && (user.user_id === activity.owner_id || user.userType === "admin");
  const isTeacherOrAdmin =
    user && (user.userType === "teacher" || user.userType === "admin");

  // 获取参与者列表
  const fetchParticipants = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get(
        `/activities/${activity.id}/participants`
      );
      setParticipants(response.data.data?.participants || []);
    } catch (error) {
      console.error("Failed to fetch participants:", error);
      toast.error("获取参与者列表失败");
    } finally {
      setLoading(false);
    }
  };

  // 获取参与者统计信息
  const fetchStats = async () => {
    try {
      const response = await apiClient.get(
        `/activities/${activity.id}/participants/stats`
      );
      setStats(response.data.data);
    } catch (error) {
      console.error("Failed to fetch stats:", error);
    }
  };

  // 搜索用户
  const searchUsers = async (query: string) => {
    if (!query.trim()) {
      setUserSearchResults([]);
      return;
    }

    setUserSearchLoading(true);
    try {
      const response = await apiClient.get(`/search/users`, {
        params: {
          query,
          user_type: "student",
          page_size: 20,
        },
      });

      const users = response.data.data?.users || [];
      // 过滤掉已经是参与者的用户
      const filteredUsers = users.filter(
        (user: UserSearchResult) =>
          !participants.some((p) => p.user_id === user.user_id)
      );
      setUserSearchResults(filteredUsers);
    } catch (error) {
      console.error("Failed to search users:", error);
      toast.error("搜索用户失败");
    } finally {
      setUserSearchLoading(false);
    }
  };

  // 添加参与者
  const addParticipants = async () => {
    if (selectedUsers.length === 0) {
      toast.error("请选择要添加的用户");
      return;
    }

    if (addDialogCredits < 0 || addDialogCredits > 10) {
      toast.error("学分值必须在0-10之间");
      return;
    }

    try {
      const response = await apiClient.post(
        `/activities/${activity.id}/participants`,
        {
          user_ids: selectedUsers,
          credits: addDialogCredits,
        }
      );

      toast.success(
        `成功添加 ${response.data.data?.added_count || 0} 名参与者`
      );
      setShowAddDialog(false);
      setSelectedUsers([]);
      setUserSearchResults([]);
      setUserSearchQuery("");
      setAddDialogCredits(1.0);
      fetchParticipants();
      onRefresh?.();
    } catch (error) {
      console.error("Failed to add participants:", error);
    }
  };

  // 删除参与者
  const removeParticipant = async (userId: string) => {
    try {
      await apiClient.delete(
        `/activities/${activity.id}/participants/${userId}`
      );
      toast.success("参与者删除成功");
      fetchParticipants();
      onRefresh?.();
    } catch (error) {
      console.error("Failed to remove participant:", error);
    }
  };

  // 批量删除参与者
  const batchRemoveParticipants = async () => {
    if (selectedParticipants.length === 0) {
      toast.error("请选择要删除的参与者");
      return;
    }

    try {
      await apiClient.post(
        `/activities/${activity.id}/participants/batch-remove`,
        {
          user_ids: selectedParticipants,
        }
      );
      toast.success(`成功删除 ${selectedParticipants.length} 名参与者`);
      setShowBatchDialog(false);
      setSelectedParticipants([]);
      fetchParticipants();
      onRefresh?.();
    } catch (error) {
      console.error("Failed to batch remove participants:", error);
    }
  };

  // 设置单个学分
  const setCredits = async (userId: string, credits: number) => {
    if (credits < 0 || credits > 10) {
      toast.error("学分值必须在0-10之间");
      return;
    }

    try {
      await apiClient.put(
        `/activities/${activity.id}/participants/${userId}/credits`,
        {
          credits,
        }
      );
      toast.success("学分设置成功");
      setEditingCredits((prev) => ({ ...prev, [userId]: undefined }));
      fetchParticipants();
      onRefresh?.();
    } catch (error) {
      console.error("Failed to set credits:", error);
    }
  };

  // 批量设置学分
  const batchSetCredits = async () => {
    if (batchDialogCredits < 0 || batchDialogCredits > 10) {
      toast.error("学分值必须在0-10之间");
      return;
    }

    const creditsMap: { [key: string]: number } = {};
    selectedParticipants.forEach((userId) => {
      creditsMap[userId] = batchDialogCredits;
    });

    try {
      await apiClient.put(
        `/activities/${activity.id}/participants/batch-credits`,
        {
          credits_map: creditsMap,
        }
      );
      toast.success("批量设置学分成功");
      setShowBatchDialog(false);
      setSelectedParticipants([]);
      setBatchDialogCredits(1.0);
      fetchParticipants();
      onRefresh?.();
    } catch (error) {
      console.error("Failed to batch set credits:", error);
    }
  };

  // 导出参与者名单
  const exportParticipants = async () => {
    try {
      const response = await apiClient.get(
        `/activities/${activity.id}/participants/export`,
        {
          params: { format: "json" },
        }
      );

      const data = response.data.data;
      const blob = new Blob([JSON.stringify(data, null, 2)], {
        type: "application/json",
      });
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `参与者名单_${activity.title}_${
        new Date().toISOString().split("T")[0]
      }.json`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      window.URL.revokeObjectURL(url);

      toast.success("导出成功");
    } catch (error) {
      console.error("Failed to export participants:", error);
      toast.error("导出失败");
    }
  };

  // 退出活动（学生功能）
  const leaveActivity = async () => {
    try {
      await apiClient.post(`/activities/${activity.id}/leave`);
      toast.success("退出活动成功");
      onRefresh?.();
    } catch (error) {
      console.error("Failed to leave activity:", error);
    }
  };

  // 过滤参与者
  const filteredParticipants = participants.filter((participant) => {
    if (!searchQuery) return true;
    const userInfo = participant.user_info;
    return (
      userInfo?.name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      userInfo?.student_id?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      userInfo?.username?.toLowerCase().includes(searchQuery.toLowerCase())
    );
  });

  useEffect(() => {
    fetchParticipants();
  }, [activity.id]);

  useEffect(() => {
    if (showStatsDialog) {
      fetchStats();
    }
  }, [showStatsDialog]);

  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (userSearchQuery) {
        searchUsers(userSearchQuery);
      }
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [userSearchQuery]);

  useEffect(() => {
    if (showAddDialog) {
      setAddDialogCredits(1.0);
    }
  }, [showAddDialog]);

  useEffect(() => {
    if (showBatchDialog) {
      setBatchDialogCredits(1.0);
    }
  }, [showBatchDialog]);

  // 全局键盘快捷键
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // 只在对话框打开时处理快捷键
      if (!showAddDialog && !showBatchDialog) return;

      // Ctrl+A: 全选/取消全选
      if (e.ctrlKey && e.key === "a") {
        e.preventDefault();
        if (showAddDialog && userSearchResults.length > 0) {
          const allSelected = selectedUsers.length === userSearchResults.length;
          if (allSelected) {
            setSelectedUsers([]);
          } else {
            setSelectedUsers(userSearchResults.map((u) => u.user_id));
          }
        } else if (showBatchDialog && filteredParticipants.length > 0) {
          const allSelected =
            selectedParticipants.length === filteredParticipants.length;
          if (allSelected) {
            setSelectedParticipants([]);
          } else {
            setSelectedParticipants(filteredParticipants.map((p) => p.user_id));
          }
        }
      }

      // Enter: 确认操作
      if (e.key === "Enter" && !e.ctrlKey) {
        e.preventDefault();
        if (showAddDialog && selectedUsers.length > 0) {
          addParticipants();
        } else if (showBatchDialog && selectedParticipants.length > 0) {
          // 默认执行批量设置学分
          batchSetCredits();
        }
      }

      // Escape: 关闭对话框
      if (e.key === "Escape") {
        if (showAddDialog) {
          setShowAddDialog(false);
        } else if (showBatchDialog) {
          setShowBatchDialog(false);
        }
      }

      // Tab + Shift: 反向循环选择
      if (e.key === "Tab" && e.shiftKey) {
        e.preventDefault();
        if (showAddDialog && userSearchResults.length > 0) {
          const currentIndex = userSearchResults.findIndex(
            (u) =>
              document.activeElement?.getAttribute("data-user-id") === u.user_id
          );
          if (currentIndex > 0) {
            const prevUser = userSearchResults[currentIndex - 1];
            const element = document.querySelector(
              `[data-user-id="${prevUser.user_id}"]`
            ) as HTMLElement;
            element?.focus();
          }
        } else if (showBatchDialog && filteredParticipants.length > 0) {
          const currentIndex = filteredParticipants.findIndex(
            (p) =>
              document.activeElement?.getAttribute("data-participant-id") ===
              p.user_id
          );
          if (currentIndex > 0) {
            const prevParticipant = filteredParticipants[currentIndex - 1];
            const element = document.querySelector(
              `[data-participant-id="${prevParticipant.user_id}"]`
            ) as HTMLElement;
            element?.focus();
          }
        }
      }
    };

    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [
    showAddDialog,
    showBatchDialog,
    selectedUsers,
    selectedParticipants,
    userSearchResults,
    filteredParticipants,
  ]);

  if (participants.length === 0 && !isOwner) {
    return (
      <Card className="rounded-xl shadow-lg">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            参与者列表
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            暂无参与者
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="rounded-xl shadow-lg">
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            参与者列表 ({participants.length})
          </CardTitle>
          <div className="flex items-center gap-2">
            {isOwner && (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowStatsDialog(true)}
                >
                  <BarChart3 className="h-4 w-4 mr-1" />
                  统计
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={exportParticipants}
                >
                  <Download className="h-4 w-4 mr-1" />
                  导出
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowAddDialog(true)}
                >
                  <Plus className="h-4 w-4 mr-1" />
                  添加
                </Button>
              </>
            )}
            {!isOwner &&
              user &&
              participants.some((p) => p.user_id === user.user_id) && (
                <Button variant="outline" size="sm" onClick={leaveActivity}>
                  退出活动
                </Button>
              )}
          </div>
        </div>
      </CardHeader>

      <CardContent>
        {/* 搜索和批量操作 */}
        <div className="flex items-center gap-4 mb-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="搜索参与者..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-10"
            />
          </div>
          {isOwner && selectedParticipants.length > 0 && (
            <div className="flex items-center gap-2">
              <Badge variant="secondary">
                已选择 {selectedParticipants.length} 人
              </Badge>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setShowBatchDialog(true)}
              >
                批量操作
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => setSelectedParticipants([])}
              >
                取消选择
              </Button>
            </div>
          )}
        </div>

        {/* 参与者表格 */}
        <div className="space-y-2">
          {isOwner && (
            <div className="flex items-center justify-between">
              <div className="text-sm text-muted-foreground">
                点击行选择参与者，或使用 Ctrl+A 全选
              </div>
              {selectedParticipants.length > 0 && (
                <div className="text-sm font-medium text-primary">
                  已选择 {selectedParticipants.length} 人
                </div>
              )}
            </div>
          )}
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow>
                  {isOwner && (
                    <TableHead className="w-12">
                      <Checkbox
                        checked={
                          selectedParticipants.length ===
                            filteredParticipants.length &&
                          filteredParticipants.length > 0
                        }
                        onCheckedChange={(checked: boolean) => {
                          if (checked) {
                            setSelectedParticipants(
                              filteredParticipants.map((p) => p.user_id)
                            );
                          } else {
                            setSelectedParticipants([]);
                          }
                        }}
                      />
                    </TableHead>
                  )}
                  <TableHead>用户信息</TableHead>
                  <TableHead>学分</TableHead>
                  <TableHead>加入时间</TableHead>
                  {isOwner && <TableHead className="w-20">操作</TableHead>}
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredParticipants.map((participant) => (
                  <TableRow
                    key={participant.user_id}
                    className={`cursor-pointer hover:bg-muted/50 transition-colors ${
                      selectedParticipants.includes(participant.user_id)
                        ? "bg-primary/10 border-primary/20"
                        : ""
                    }`}
                    onClick={(e) => {
                      // 如果点击的不是checkbox本身，则切换选择状态
                      if (
                        !(e.target as HTMLElement).closest(
                          'input[type="checkbox"]'
                        ) &&
                        isOwner
                      ) {
                        const isSelected = selectedParticipants.includes(
                          participant.user_id
                        );
                        if (isSelected) {
                          setSelectedParticipants((prev) =>
                            prev.filter((id) => id !== participant.user_id)
                          );
                        } else {
                          setSelectedParticipants((prev) => [
                            ...prev,
                            participant.user_id,
                          ]);
                        }
                      }
                    }}
                    onKeyDown={(e) => {
                      if ((e.key === "Enter" || e.key === " ") && isOwner) {
                        e.preventDefault();
                        const isSelected = selectedParticipants.includes(
                          participant.user_id
                        );
                        if (isSelected) {
                          setSelectedParticipants((prev) =>
                            prev.filter((id) => id !== participant.user_id)
                          );
                        } else {
                          setSelectedParticipants((prev) => [
                            ...prev,
                            participant.user_id,
                          ]);
                        }
                      }
                    }}
                    tabIndex={0}
                    role="button"
                    aria-label={`选择参与者 ${
                      participant.user_info?.name || participant.user_id
                    }`}
                    data-participant-id={participant.user_id}
                  >
                    {isOwner && (
                      <TableCell onClick={(e) => e.stopPropagation()}>
                        <Checkbox
                          checked={selectedParticipants.includes(
                            participant.user_id
                          )}
                          onCheckedChange={(checked: boolean) => {
                            if (checked) {
                              setSelectedParticipants((prev) => [
                                ...prev,
                                participant.user_id,
                              ]);
                            } else {
                              setSelectedParticipants((prev) =>
                                prev.filter((id) => id !== participant.user_id)
                              );
                            }
                          }}
                        />
                      </TableCell>
                    )}
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
                          <User className="h-4 w-4 text-primary" />
                        </div>
                        <div>
                          <div className="font-medium">
                            {participant.user_info?.name ||
                              `用户 ${participant.user_id}`}
                          </div>
                          <div className="text-sm text-muted-foreground">
                            {participant.user_info?.student_id ||
                              participant.user_id}
                          </div>
                          {participant.user_info?.college && (
                            <div className="text-xs text-muted-foreground">
                              {participant.user_info.college} -{" "}
                              {participant.user_info.major}
                            </div>
                          )}
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      {editingCredits[participant.user_id] !== undefined ? (
                        <div className="flex items-center gap-2">
                          <Input
                            type="number"
                            step="0.1"
                            min="0"
                            max="10"
                            value={editingCredits[participant.user_id]}
                            onChange={(e) => {
                              const value = e.target.value;
                              if (value === "") {
                                setEditingCredits((prev) => ({
                                  ...prev,
                                  [participant.user_id]: 0,
                                }));
                              } else {
                                const numValue = parseFloat(value);
                                if (!isNaN(numValue) && numValue >= 0) {
                                  setEditingCredits((prev) => ({
                                    ...prev,
                                    [participant.user_id]: numValue,
                                  }));
                                }
                              }
                            }}
                            className="w-20"
                          />
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() =>
                              setCredits(
                                participant.user_id,
                                editingCredits[participant.user_id] || 0
                              )
                            }
                          >
                            <Check className="h-3 w-3" />
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() =>
                              setEditingCredits((prev) => ({
                                ...prev,
                                [participant.user_id]: undefined,
                              }))
                            }
                          >
                            <X className="h-3 w-3" />
                          </Button>
                        </div>
                      ) : (
                        <div className="flex items-center gap-2">
                          <span className="font-bold text-primary">
                            {participant.credits} 学分
                          </span>
                          {isOwner && (
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() =>
                                setEditingCredits((prev) => ({
                                  ...prev,
                                  [participant.user_id]: participant.credits,
                                }))
                              }
                            >
                              <Edit3 className="h-3 w-3" />
                            </Button>
                          )}
                        </div>
                      )}
                    </TableCell>
                    <TableCell>
                      <div className="text-sm text-muted-foreground">
                        {new Date(participant.joined_at).toLocaleDateString()}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {new Date(participant.joined_at).toLocaleTimeString()}
                      </div>
                    </TableCell>
                    {isOwner && (
                      <TableCell>
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="sm">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuLabel>操作</DropdownMenuLabel>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              onClick={() =>
                                setEditingCredits((prev) => ({
                                  ...prev,
                                  [participant.user_id]: participant.credits,
                                }))
                              }
                            >
                              <Edit3 className="h-4 w-4 mr-2" />
                              编辑学分
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() =>
                                removeParticipant(participant.user_id)
                              }
                              className="text-red-600"
                            >
                              <Trash2 className="h-4 w-4 mr-2" />
                              删除
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    )}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </div>

        {filteredParticipants.length === 0 && (
          <div className="text-center py-8 text-muted-foreground">
            {searchQuery ? "没有找到匹配的参与者" : "暂无参与者"}
          </div>
        )}
      </CardContent>

      {/* 添加参与者对话框 */}
      <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>添加参与者</DialogTitle>
            <DialogDescription>
              搜索并选择要添加到活动的学生用户
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label htmlFor="user-search">搜索用户</Label>
              <div className="relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="user-search"
                  placeholder="输入姓名、学号或用户名搜索..."
                  value={userSearchQuery}
                  onChange={(e) => setUserSearchQuery(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>

            <div>
              <Label htmlFor="default-credits">默认学分</Label>
              <Input
                id="default-credits"
                type="number"
                step="0.1"
                min="0"
                max="10"
                value={addDialogCredits}
                onChange={(e) => {
                  const value = e.target.value;
                  if (value === "") {
                    setAddDialogCredits(0);
                  } else {
                    const numValue = parseFloat(value);
                    if (!isNaN(numValue) && numValue >= 0) {
                      setAddDialogCredits(numValue);
                    }
                  }
                }}
                placeholder="请输入学分"
              />
            </div>

            {userSearchLoading && (
              <div className="text-center py-4">
                <div className="w-4 h-4 border-2 border-primary/30 border-t-primary rounded-full animate-spin mx-auto" />
                <p className="text-sm text-muted-foreground mt-2">搜索中...</p>
              </div>
            )}

            {userSearchResults.length > 0 && (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="text-sm text-muted-foreground">
                    点击行选择用户，或使用 Ctrl+A 全选
                  </div>
                  {selectedUsers.length > 0 && (
                    <div className="text-sm font-medium text-primary">
                      已选择 {selectedUsers.length} 人
                    </div>
                  )}
                </div>
                <div className="border rounded-lg max-h-60 overflow-y-auto">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-12">
                          <Checkbox
                            checked={
                              selectedUsers.length === userSearchResults.length
                            }
                            onCheckedChange={(checked: boolean) => {
                              if (checked) {
                                setSelectedUsers(
                                  userSearchResults.map((u) => u.user_id)
                                );
                              } else {
                                setSelectedUsers([]);
                              }
                            }}
                          />
                        </TableHead>
                        <TableHead>用户信息</TableHead>
                        <TableHead>学号</TableHead>
                        <TableHead>学院专业</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {userSearchResults.map((user) => (
                        <TableRow
                          key={user.user_id}
                          className={`cursor-pointer hover:bg-muted/50 transition-colors ${
                            selectedUsers.includes(user.user_id)
                              ? "bg-primary/10 border-primary/20"
                              : ""
                          }`}
                          onClick={(e) => {
                            // 如果点击的不是checkbox本身，则切换选择状态
                            if (
                              !(e.target as HTMLElement).closest(
                                'input[type="checkbox"]'
                              )
                            ) {
                              const isSelected = selectedUsers.includes(
                                user.user_id
                              );
                              if (isSelected) {
                                setSelectedUsers((prev) =>
                                  prev.filter((id) => id !== user.user_id)
                                );
                              } else {
                                setSelectedUsers((prev) => [
                                  ...prev,
                                  user.user_id,
                                ]);
                              }
                            }
                          }}
                          onKeyDown={(e) => {
                            if (e.key === "Enter" || e.key === " ") {
                              e.preventDefault();
                              const isSelected = selectedUsers.includes(
                                user.user_id
                              );
                              if (isSelected) {
                                setSelectedUsers((prev) =>
                                  prev.filter((id) => id !== user.user_id)
                                );
                              } else {
                                setSelectedUsers((prev) => [
                                  ...prev,
                                  user.user_id,
                                ]);
                              }
                            }
                          }}
                          tabIndex={0}
                          role="button"
                          aria-label={`选择用户 ${user.real_name}`}
                          data-user-id={user.user_id}
                        >
                          <TableCell onClick={(e) => e.stopPropagation()}>
                            <Checkbox
                              checked={selectedUsers.includes(user.user_id)}
                              onCheckedChange={(checked: boolean) => {
                                if (checked) {
                                  setSelectedUsers((prev) => [
                                    ...prev,
                                    user.user_id,
                                  ]);
                                } else {
                                  setSelectedUsers((prev) =>
                                    prev.filter((id) => id !== user.user_id)
                                  );
                                }
                              }}
                            />
                          </TableCell>
                          <TableCell>
                            <div>
                              <div className="font-medium">
                                {user.real_name}
                              </div>
                              <div className="text-sm text-muted-foreground">
                                {user.username}
                              </div>
                            </div>
                          </TableCell>
                          <TableCell>{user.student_id || "-"}</TableCell>
                          <TableCell>
                            {user.college && user.major
                              ? `${user.college} - ${user.major}`
                              : "-"}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </div>
            )}

            {userSearchQuery &&
              !userSearchLoading &&
              userSearchResults.length === 0 && (
                <div className="text-center py-4 text-muted-foreground">
                  没有找到匹配的用户
                </div>
              )}
          </div>

          <DialogFooter>
            <div className="flex items-center justify-between w-full">
              <div className="text-xs text-muted-foreground">
                快捷键: Ctrl+A 全选 | Enter 确认 | Esc 取消 | Tab 导航
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  onClick={() => setShowAddDialog(false)}
                >
                  取消
                </Button>
                <Button
                  onClick={addParticipants}
                  disabled={selectedUsers.length === 0}
                >
                  添加 ({selectedUsers.length})
                </Button>
              </div>
            </div>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 批量操作对话框 */}
      <Dialog open={showBatchDialog} onOpenChange={setShowBatchDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>批量操作</DialogTitle>
            <DialogDescription>
              对选中的 {selectedParticipants.length} 名参与者进行批量操作
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label htmlFor="batch-credits">设置学分</Label>
              <Input
                id="batch-credits"
                type="number"
                step="0.1"
                min="0"
                max="10"
                value={batchDialogCredits}
                onChange={(e) => {
                  const value = e.target.value;
                  if (value === "") {
                    setBatchDialogCredits(0);
                  } else {
                    const numValue = parseFloat(value);
                    if (!isNaN(numValue) && numValue >= 0) {
                      setBatchDialogCredits(numValue);
                    }
                  }
                }}
                placeholder="请输入学分"
              />
            </div>
          </div>

          <DialogFooter>
            <div className="flex items-center justify-between w-full">
              <div className="text-xs text-muted-foreground">
                快捷键: Ctrl+A 全选 | Enter 确认 | Esc 取消 | Tab 导航
              </div>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  onClick={() => setShowBatchDialog(false)}
                >
                  取消
                </Button>
                <Button variant="destructive" onClick={batchRemoveParticipants}>
                  批量删除
                </Button>
                <Button onClick={batchSetCredits}>批量设置学分</Button>
              </div>
            </div>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 统计信息对话框 */}
      <Dialog open={showStatsDialog} onOpenChange={setShowStatsDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>参与者统计</DialogTitle>
            <DialogDescription>活动的参与者统计信息</DialogDescription>
          </DialogHeader>

          {stats && (
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <div className="text-sm font-medium">总参与者</div>
                <div className="text-2xl font-bold text-primary">
                  {stats.total_participants}
                </div>
              </div>
              <div className="space-y-2">
                <div className="text-sm font-medium">总学分</div>
                <div className="text-2xl font-bold text-green-600">
                  {stats.total_credits}
                </div>
              </div>
              <div className="space-y-2">
                <div className="text-sm font-medium">平均学分</div>
                <div className="text-lg font-semibold">
                  {stats.avg_credits?.toFixed(1) || 0}
                </div>
              </div>
              <div className="space-y-2">
                <div className="text-sm font-medium">最近加入</div>
                <div className="text-lg font-semibold">
                  {stats.recent_participants} 人
                </div>
              </div>
              <div className="space-y-2">
                <div className="text-sm font-medium">最高学分</div>
                <div className="text-lg font-semibold text-orange-600">
                  {stats.max_credits || 0}
                </div>
              </div>
              <div className="space-y-2">
                <div className="text-sm font-medium">最低学分</div>
                <div className="text-lg font-semibold text-blue-600">
                  {stats.min_credits || 0}
                </div>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button onClick={() => setShowStatsDialog(false)}>关闭</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </Card>
  );
}
