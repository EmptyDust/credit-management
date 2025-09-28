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
import { Label } from "@/components/ui/label";
import { Checkbox } from "@/components/ui/checkbox";
import { useAuth } from "@/contexts/AuthContext";
import apiClient from "@/lib/api";
import userService from "@/lib/userService";
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
  id: string;
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
    [key: string]: number | "" | undefined;
  }>({});
  const [stats, setStats] = useState<any>(null);

  const isOwner =
    user && (user.id === activity.owner_id || user.userType === "admin");

  // 添加活动状态检查：只有草稿状态的活动才能编辑参与者
  const canEditParticipants = isOwner && activity.status === "draft";

  // 获取参与者列表
  const fetchParticipants = async () => {
    setLoading(true);
    try {
      const response = await apiClient.get(
        `/activities/${activity.id}/participants`
      );
      
      // 调试：打印响应数据
      console.log("Participants API Response:", response.data);
      
      // 处理分页响应数据结构
      let participantsData = [];
      if (response.data.code === 0 && response.data.data) {
        if (response.data.data.data && Array.isArray(response.data.data.data)) {
          // 分页数据结构
          participantsData = response.data.data.data;
        } else if (response.data.data.participants && Array.isArray(response.data.data.participants)) {
          // 非分页数据结构
          participantsData = response.data.data.participants;
        } else {
          // 如果没有数据或数据不是数组，使用空数组
          participantsData = [];
        }
      }
      
      // 调试：打印处理后的数据
      console.log("Participants Data:", participantsData);
      console.log("Is Array:", Array.isArray(participantsData));
      
      setParticipants(participantsData);
    } catch (error) {
      console.error("Failed to fetch participants:", error);
      toast.error("获取参与者列表失败");
      setParticipants([]);
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
      // 使用userService的searchUsers方法
      const response = await userService.searchUsers({
        query: query.trim(),
        user_type: "student",
        page: 1,
        page_size: 20,
      });

      const users = response.users || [];
      // 过滤掉已经是参与者的用户
      const filteredUsers = users.filter(
        (user: UserSearchResult) =>
          !participants.some((p) => p.id === user.id)
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
          ids: selectedUsers,
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
          ids: selectedParticipants,
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
  const safeParticipants = Array.isArray(participants) ? participants : [];
  const filteredParticipants = safeParticipants.filter((participant) => {
    if (!searchQuery) return true;
    const userInfo = participant.user_info;
    return (
      userInfo?.real_name?.toLowerCase().includes(searchQuery.toLowerCase()) ||
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
            setSelectedUsers(userSearchResults.map((u) => u.id));
          }
        } else if (showBatchDialog && filteredParticipants.length > 0) {
          const allSelected =
            selectedParticipants.length === filteredParticipants.length;
          if (allSelected) {
            setSelectedParticipants([]);
          } else {
            setSelectedParticipants(filteredParticipants.map((p) => p.id));
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
              document.activeElement?.getAttribute("data-user-id") === u.id
          );
          if (currentIndex > 0) {
            const prevUser = userSearchResults[currentIndex - 1];
            const element = document.querySelector(
              `[data-user-id="${prevUser.id}"]`
            ) as HTMLElement;
            element?.focus();
          }
        } else if (showBatchDialog && filteredParticipants.length > 0) {
          const currentIndex = filteredParticipants.findIndex(
            (p) =>
              document.activeElement?.getAttribute("data-participant-id") ===
              p.id
          );
          if (currentIndex > 0) {
            const prevParticipant = filteredParticipants[currentIndex - 1];
            const element = document.querySelector(
              `[data-participant-id="${prevParticipant.id}"]`
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

  if ((!Array.isArray(participants) || participants.length === 0) && !isOwner) {
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
            参与者列表 ({Array.isArray(participants) ? participants.length : 0})
            {!canEditParticipants && isOwner && (
              <Badge variant="secondary" className="ml-2">
                仅查看模式
              </Badge>
            )}
          </CardTitle>
          <div className="flex items-center gap-2">
            {canEditParticipants && (
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
              participants.some((p) => p.id === user.id) && (
                <Button variant="outline" size="sm" onClick={leaveActivity}>
                  退出活动
                </Button>
              )}
          </div>
        </div>
      </CardHeader>

      <CardContent>
        <div className="space-y-4">
          {/* 搜索和批量操作 */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              <div className="relative">
                <Search className="absolute left-2 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="搜索参与者..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-8 w-64"
                />
              </div>
            </div>
            {canEditParticipants && filteredParticipants.length > 0 && (
              <div className="flex items-center gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setShowBatchDialog(true)}
                  disabled={selectedParticipants.length === 0}
                >
                  <MoreHorizontal className="h-4 w-4 mr-1" />
                  批量操作 ({selectedParticipants.length})
                </Button>
              </div>
            )}
          </div>

          {/* 参与者表格 */}
          <div className="border rounded-lg">
            <Table>
              <TableHeader>
                <TableRow>
                  {canEditParticipants && (
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
                              filteredParticipants.map((p) => p.id)
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
                  {canEditParticipants && (
                    <TableHead className="w-12">操作</TableHead>
                  )}
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredParticipants.map((participant) => (
                  <TableRow
                    key={participant.id}
                    onClick={() => {
                      if (canEditParticipants) {
                        if (
                          selectedParticipants.includes(participant.id)
                        ) {
                          setSelectedParticipants((prev) =>
                            prev.filter((id) => id !== participant.id)
                          );
                        } else {
                          setSelectedParticipants((prev) => [
                            ...prev,
                            participant.id,
                          ]);
                        }
                      }
                    }}
                    tabIndex={0}
                    role="button"
                    aria-label={`选择参与者 ${
                      participant.user_info?.real_name || participant.id
                    }`}
                    data-participant-id={participant.id}
                  >
                    {canEditParticipants && (
                      <TableCell onClick={(e) => e.stopPropagation()}>
                        <Checkbox
                          checked={selectedParticipants.includes(
                            participant.id
                          )}
                          onCheckedChange={(checked: boolean) => {
                            if (checked) {
                              setSelectedParticipants((prev) => [
                                ...prev,
                                participant.id,
                              ]);
                            } else {
                              setSelectedParticipants((prev) =>
                                prev.filter((id) => id !== participant.id)
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
                            {participant.user_info?.real_name ||
                              participant.user_info?.username ||
                              "未知用户"}
                          </div>
                          <div className="text-sm text-muted-foreground">
                            {participant.user_info?.username}
                          </div>
                        </div>
                      </div>
                    </TableCell>
                    <TableCell>
                      {editingCredits[participant.id] !== undefined ? (
                        <div className="flex items-center gap-2">
                          <Input
                            type="number"
                            step="0.1"
                            min="0"
                            max="10"
                            value={
                              editingCredits[participant.id] === undefined
                                ? ""
                                : editingCredits[participant.id]
                            }
                            onChange={(e) => {
                              const value = e.target.value;
                              setEditingCredits((prev) => ({
                                ...prev,
                                [participant.id]:
                                  value === "" ? "" : parseFloat(value),
                              }));
                            }}
                            className="w-20"
                          />
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => {
                              const val = editingCredits[participant.id];
                              if (
                                val === "" ||
                                val === undefined ||
                                isNaN(Number(val))
                              ) {
                                toast.error("请输入有效的学分");
                                return;
                              }
                              setCredits(participant.id, Number(val));
                            }}
                          >
                            <Check className="h-3 w-3" />
                          </Button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() =>
                              setEditingCredits((prev) => ({
                                ...prev,
                                [participant.id]: undefined,
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
                          {canEditParticipants && (
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={() =>
                                setEditingCredits((prev) => ({
                                  ...prev,
                                  [participant.id]: participant.credits,
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
                    {canEditParticipants && (
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
                                  [participant.id]: participant.credits,
                                }))
                              }
                            >
                              <Edit3 className="h-4 w-4 mr-2" />
                              编辑学分
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() =>
                                removeParticipant(participant.id)
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
                <Search className="absolute left-2 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  id="user-search"
                  placeholder="输入用户名、真实姓名、学号或UUID进行搜索..."
                  value={userSearchQuery}
                  onChange={(e) => setUserSearchQuery(e.target.value)}
                  className="pl-8"
                />
              </div>
            </div>

            <div>
              <Label htmlFor="credits">学分</Label>
              <Input
                id="credits"
                type="number"
                step="0.1"
                min="0"
                max="10"
                value={addDialogCredits}
                onChange={(e) =>
                  setAddDialogCredits(parseFloat(e.target.value) || 0)
                }
              />
            </div>

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
                                  userSearchResults.map((u) => u.id)
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
                          key={user.id}
                          onClick={() => {
                            if (selectedUsers.includes(user.id)) {
                              setSelectedUsers((prev) =>
                                prev.filter((id) => id !== user.id)
                              );
                            } else {
                              setSelectedUsers((prev) => [
                                ...prev,
                                user.id,
                              ]);
                            }
                          }}
                          className="cursor-pointer hover:bg-muted/50"
                          tabIndex={0}
                          role="button"
                          aria-label={`选择用户 ${user.real_name}`}
                          data-user-id={user.id}
                        >
                          <TableCell onClick={(e) => e.stopPropagation()}>
                            <Checkbox
                              checked={selectedUsers.includes(user.id)}
                              onCheckedChange={(checked: boolean) => {
                                if (checked) {
                                  setSelectedUsers((prev) => [
                                    ...prev,
                                    user.id,
                                  ]);
                                } else {
                                  setSelectedUsers((prev) =>
                                    prev.filter((id) => id !== user.id)
                                  );
                                }
                              }}
                            />
                          </TableCell>
                          <TableCell>
                            <div className="flex flex-col">
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
                              : user.college || user.major || "-"}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </div>
              </div>
            )}

            {userSearchLoading && (
              <div className="text-center py-4 text-muted-foreground">
                搜索中...
              </div>
            )}

            {!userSearchLoading &&
              userSearchQuery &&
              userSearchResults.length === 0 && (
                <div className="text-center py-4 text-muted-foreground">
                  没有找到匹配的用户
                </div>
              )}
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAddDialog(false)}>
              取消
            </Button>
            <Button
              onClick={addParticipants}
              disabled={selectedUsers.length === 0}
            >
              添加参与者
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 批量操作对话框 */}
      <Dialog open={showBatchDialog} onOpenChange={setShowBatchDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>批量操作</DialogTitle>
            <DialogDescription>
              对选中的 {selectedParticipants.length} 名参与者进行操作
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label htmlFor="batch-credits">批量设置学分</Label>
              <Input
                id="batch-credits"
                type="number"
                step="0.1"
                min="0"
                max="10"
                value={batchDialogCredits}
                onChange={(e) =>
                  setBatchDialogCredits(parseFloat(e.target.value) || 0)
                }
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowBatchDialog(false)}>
              取消
            </Button>
            <Button onClick={batchSetCredits}>设置学分</Button>
            <Button variant="destructive" onClick={batchRemoveParticipants}>
              删除参与者
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* 统计信息对话框 */}
      <Dialog open={showStatsDialog} onOpenChange={setShowStatsDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>参与者统计</DialogTitle>
          </DialogHeader>

          {stats && (
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                <div className="text-center p-4 bg-muted rounded-lg">
                  <div className="text-2xl font-bold text-primary">
                    {stats.total_participants}
                  </div>
                  <div className="text-sm text-muted-foreground">总参与者</div>
                </div>
                <div className="text-center p-4 bg-muted rounded-lg">
                  <div className="text-2xl font-bold text-primary">
                    {stats.total_credits}
                  </div>
                  <div className="text-sm text-muted-foreground">总学分</div>
                </div>
                <div className="text-center p-4 bg-muted rounded-lg">
                  <div className="text-2xl font-bold text-primary">
                    {stats.avg_credits?.toFixed(1) || "0"}
                  </div>
                  <div className="text-sm text-muted-foreground">平均学分</div>
                </div>
                <div className="text-center p-4 bg-muted rounded-lg">
                  <div className="text-2xl font-bold text-primary">
                    {stats.recent_participants}
                  </div>
                  <div className="text-sm text-muted-foreground">最近加入</div>
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
