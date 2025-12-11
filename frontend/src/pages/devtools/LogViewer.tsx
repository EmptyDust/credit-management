import { useState, useEffect, useRef, useCallback } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Play,
  Pause,
  Trash2,
  Download,
  Search,
  Server,
  Database,
  Shield,
  Activity,
  Users,
  Globe,
  RefreshCw,
  AlertCircle,
  Wifi,
  WifiOff,
} from "lucide-react";

// 服务配置 - 使用 Docker 容器名称
const SERVICES = [
  { id: "credit_management_gateway", name: "API Gateway", icon: Globe, color: "text-blue-400" },
  { id: "credit_management_auth", name: "Auth Service", icon: Shield, color: "text-emerald-400" },
  { id: "credit_management_user", name: "User Service", icon: Users, color: "text-purple-400" },
  { id: "credit_management_credit_activity", name: "Activity Service", icon: Activity, color: "text-amber-400" },
  { id: "credit_management_postgres", name: "PostgreSQL", icon: Database, color: "text-cyan-400" },
  { id: "credit_management_redis", name: "Redis", icon: Server, color: "text-rose-400" },
  { id: "credit_management_frontend", name: "Frontend", icon: Globe, color: "text-pink-400" },
];

// 日志级别配置
const LOG_LEVELS = {
  DEBUG: { color: "text-gray-400", bg: "bg-gray-500/10" },
  INFO: { color: "text-blue-400", bg: "bg-blue-500/10" },
  WARN: { color: "text-amber-400", bg: "bg-amber-500/10" },
  ERROR: { color: "text-rose-400", bg: "bg-rose-500/10" },
};

interface LogEntry {
  id: string;
  timestamp: string;
  level: keyof typeof LOG_LEVELS;
  service: string;
  message: string;
}

interface ServiceStatus {
  id: string;
  name: string;
  status: string;
}

// 格式化时间戳
const formatTimestamp = (timestamp: string): string => {
  try {
    // 处理 Docker 时间戳格式
    const cleanTimestamp = timestamp.replace(/Z$/, "").substring(0, 23);
    const date = new Date(cleanTimestamp + "Z");
    if (isNaN(date.getTime())) {
      return timestamp.substring(11, 19); // 只取时间部分
    }
    return date.toLocaleTimeString("zh-CN", {
      hour12: false,
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    });
  } catch {
    return timestamp.substring(11, 19);
  }
};

export default function LogViewer() {
  // 从 localStorage 获取 token
  const getToken = () => localStorage.getItem("token");
  const [selectedService, setSelectedService] = useState<string>("credit_management_gateway");
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isPaused, setIsPaused] = useState(false);
  const [filter, setFilter] = useState("");
  const [levelFilter, setLevelFilter] = useState<string>("ALL");
  const [serviceStatus, setServiceStatus] = useState<Record<string, string>>({});
  const [connectionStatus, setConnectionStatus] = useState<"connecting" | "connected" | "disconnected" | "error">("disconnected");
  const [errorMessage, setErrorMessage] = useState<string>("");
  const logContainerRef = useRef<HTMLDivElement>(null);
  const [autoScroll, setAutoScroll] = useState(true);
  const eventSourceRef = useRef<EventSource | null>(null);

  // 获取服务状态
  useEffect(() => {
    const fetchServiceStatus = async () => {
      try {
        const token = getToken();
        const response = await fetch("/api/devtools/services", {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });
        if (response.ok) {
          const result = await response.json();
          if (result.data) {
            const statusMap: Record<string, string> = {};
            result.data.forEach((svc: ServiceStatus) => {
              statusMap[svc.id] = svc.status;
            });
            setServiceStatus(statusMap);
          }
        }
      } catch (error) {
        console.error("Failed to fetch service status:", error);
      }
    };

    fetchServiceStatus();
    const interval = setInterval(fetchServiceStatus, 30000);
    return () => clearInterval(interval);
  }, []);

  // 连接日志流
  useEffect(() => {
    if (isPaused) {
      // 暂停时关闭连接
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
      setConnectionStatus("disconnected");
      return;
    }

    // 关闭旧连接
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    setConnectionStatus("connecting");
    setErrorMessage("");

    // 创建新的 SSE 连接
    const token = getToken();
    const url = `/api/devtools/logs/${selectedService}?tail=100&follow=true&token=${token}`;
    const eventSource = new EventSource(url);
    eventSourceRef.current = eventSource;

    eventSource.onopen = () => {
      setConnectionStatus("connected");
    };

    eventSource.addEventListener("log", (event) => {
      try {
        const logEntry: LogEntry = JSON.parse(event.data);
        // 确保 level 是有效的
        if (!LOG_LEVELS[logEntry.level]) {
          logEntry.level = "INFO";
        }
        setLogs((prev) => [...prev.slice(-500), logEntry]);
      } catch (error) {
        console.error("Failed to parse log entry:", error);
      }
    });

    eventSource.addEventListener("error", (event) => {
      try {
        const errorData = JSON.parse((event as MessageEvent).data);
        setErrorMessage(errorData.message || "连接错误");
      } catch {
        setErrorMessage("日志流连接失败");
      }
      setConnectionStatus("error");
    });

    eventSource.onerror = () => {
      setConnectionStatus("error");
      setErrorMessage("SSE 连接断开，请检查服务状态");
    };

    return () => {
      eventSource.close();
      eventSourceRef.current = null;
    };
  }, [selectedService, isPaused]);

  // 自动滚动
  useEffect(() => {
    if (autoScroll && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight;
    }
  }, [logs, autoScroll]);

  // 处理滚动
  const handleScroll = useCallback(() => {
    if (!logContainerRef.current) return;
    const { scrollTop, scrollHeight, clientHeight } = logContainerRef.current;
    const isAtBottom = scrollHeight - scrollTop - clientHeight < 50;
    setAutoScroll(isAtBottom);
  }, []);

  const clearLogs = () => setLogs([]);

  const downloadLogs = () => {
    const content = logs
      .map((log) => `[${log.timestamp}] [${log.level}] ${log.message}`)
      .join("\n");
    const blob = new Blob([content], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${selectedService}-logs-${new Date().toISOString().split("T")[0]}.txt`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const reconnect = () => {
    setIsPaused(false);
    // 触发重新连接
    setLogs([]);
  };

  const filteredLogs = logs.filter((log) => {
    if (levelFilter !== "ALL" && log.level !== levelFilter) return false;
    if (filter && !log.message.toLowerCase().includes(filter.toLowerCase())) return false;
    return true;
  });

  const getServiceStatus = (serviceId: string): "running" | "stopped" | "unknown" => {
    const status = serviceStatus[serviceId];
    if (status === "running") return "running";
    if (status === "exited" || status === "stopped") return "stopped";
    return "unknown";
  };

  const getConnectionStatusIcon = () => {
    switch (connectionStatus) {
      case "connecting":
        return <RefreshCw className="w-4 h-4 animate-spin text-blue-400" />;
      case "connected":
        return <Wifi className="w-4 h-4 text-emerald-400" />;
      case "error":
        return <AlertCircle className="w-4 h-4 text-rose-400" />;
      default:
        return <WifiOff className="w-4 h-4 text-gray-500" />;
    }
  };

  return (
    <div className="h-[calc(100vh-4rem)] bg-[#0d1117] text-gray-100 font-mono overflow-hidden">
      <div className="h-full flex">
        {/* 左侧 - 服务列表 */}
        <div className="w-56 border-r border-[#30363d] flex flex-col">
          <div className="p-4 border-b border-[#30363d]">
            <h2 className="text-sm font-semibold text-gray-300 mb-1">服务列表</h2>
            <p className="text-xs text-gray-500">选择要查看的服务</p>
          </div>
          <div className="flex-1 overflow-y-auto py-2">
            {SERVICES.map((service) => {
              const Icon = service.icon;
              const status = getServiceStatus(service.id);
              const isSelected = selectedService === service.id;

              return (
                <button
                  key={service.id}
                  onClick={() => {
                    setSelectedService(service.id);
                    setLogs([]);
                  }}
                  className={`w-full px-4 py-2.5 flex items-center gap-3 transition-colors ${
                    isSelected
                      ? "bg-[#161b22] border-l-2 border-blue-500"
                      : "hover:bg-[#161b22] border-l-2 border-transparent"
                  }`}
                >
                  <div className="relative">
                    <Icon className={`w-4 h-4 ${service.color}`} />
                    <span
                      className={`absolute -bottom-0.5 -right-0.5 w-2 h-2 rounded-full border border-[#0d1117] ${
                        status === "running"
                          ? "bg-emerald-400"
                          : status === "stopped"
                          ? "bg-rose-400"
                          : "bg-gray-500"
                      }`}
                    />
                  </div>
                  <span className={`text-sm ${isSelected ? "text-gray-100" : "text-gray-400"}`}>
                    {service.name}
                  </span>
                </button>
              );
            })}
          </div>
          <div className="p-4 border-t border-[#30363d]">
            <div className="flex items-center gap-2 text-xs text-gray-500">
              <span className="w-2 h-2 rounded-full bg-emerald-400" />
              <span>运行中</span>
              <span className="w-2 h-2 rounded-full bg-rose-400 ml-2" />
              <span>已停止</span>
            </div>
          </div>
        </div>

        {/* 右侧 - 日志流 */}
        <div className="flex-1 flex flex-col">
          {/* 工具栏 */}
          <div className="p-3 border-b border-[#30363d] flex items-center gap-3">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
              <Input
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
                placeholder="搜索日志..."
                className="pl-9 h-8 bg-[#161b22] border-[#30363d] text-gray-100 placeholder:text-gray-500 text-sm"
              />
            </div>
            <Select value={levelFilter} onValueChange={setLevelFilter}>
              <SelectTrigger className="w-28 h-8 bg-[#161b22] border-[#30363d] text-gray-300 text-sm">
                <SelectValue placeholder="级别" />
              </SelectTrigger>
              <SelectContent className="bg-[#161b22] border-[#30363d]">
                <SelectItem value="ALL">所有级别</SelectItem>
                <SelectItem value="DEBUG" className="text-gray-400">DEBUG</SelectItem>
                <SelectItem value="INFO" className="text-blue-400">INFO</SelectItem>
                <SelectItem value="WARN" className="text-amber-400">WARN</SelectItem>
                <SelectItem value="ERROR" className="text-rose-400">ERROR</SelectItem>
              </SelectContent>
            </Select>
            <div className="flex items-center gap-1 ml-auto">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setIsPaused(!isPaused)}
                className={`h-8 px-3 ${isPaused ? "text-amber-400" : "text-gray-400"} hover:text-gray-200`}
              >
                {isPaused ? (
                  <>
                    <Play className="w-4 h-4 mr-1" />
                    继续
                  </>
                ) : (
                  <>
                    <Pause className="w-4 h-4 mr-1" />
                    暂停
                  </>
                )}
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={clearLogs}
                className="h-8 px-3 text-gray-400 hover:text-gray-200"
              >
                <Trash2 className="w-4 h-4 mr-1" />
                清空
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={downloadLogs}
                className="h-8 px-3 text-gray-400 hover:text-gray-200"
              >
                <Download className="w-4 h-4 mr-1" />
                导出
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setAutoScroll(true)}
                className={`h-8 px-3 ${autoScroll ? "text-blue-400" : "text-gray-400"} hover:text-gray-200`}
              >
                <RefreshCw className={`w-4 h-4 ${autoScroll && connectionStatus === "connected" ? "animate-spin" : ""}`} />
              </Button>
            </div>
          </div>

          {/* 错误提示 */}
          {connectionStatus === "error" && (
            <div className="px-4 py-2 bg-rose-500/10 border-b border-rose-500/20 flex items-center justify-between">
              <div className="flex items-center gap-2 text-rose-400 text-sm">
                <AlertCircle className="w-4 h-4" />
                <span>{errorMessage || "连接失败"}</span>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={reconnect}
                className="h-7 text-xs text-rose-400 hover:text-rose-300"
              >
                重新连接
              </Button>
            </div>
          )}

          {/* 日志内容 */}
          <div
            ref={logContainerRef}
            onScroll={handleScroll}
            className="flex-1 overflow-auto p-2 bg-[#0d1117]"
          >
            {filteredLogs.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center text-gray-500">
                <Server className="w-12 h-12 mb-4 opacity-30" />
                <p className="text-sm">
                  {connectionStatus === "connecting"
                    ? "正在连接日志流..."
                    : connectionStatus === "error"
                    ? "连接失败，请检查服务状态"
                    : isPaused
                    ? "日志流已暂停"
                    : "等待日志..."}
                </p>
                <p className="text-xs mt-1">
                  {connectionStatus === "connected" && "日志将实时显示在此处"}
                </p>
              </div>
            ) : (
              <div className="space-y-0.5">
                {filteredLogs.map((log) => (
                  <div
                    key={log.id}
                    className={`px-3 py-1 rounded text-xs flex items-start gap-3 hover:bg-[#161b22] transition-colors ${
                      LOG_LEVELS[log.level]?.bg || "bg-gray-500/10"
                    }`}
                  >
                    <span className="text-gray-500 shrink-0 w-[70px]">
                      {formatTimestamp(log.timestamp)}
                    </span>
                    <span
                      className={`shrink-0 w-12 font-bold ${
                        LOG_LEVELS[log.level]?.color || "text-gray-400"
                      }`}
                    >
                      {log.level}
                    </span>
                    <span className="text-gray-300 break-all">
                      {filter ? (
                        <HighlightText text={log.message} highlight={filter} />
                      ) : (
                        log.message
                      )}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* 状态栏 */}
          <div className="px-4 py-2 border-t border-[#30363d] flex items-center justify-between text-xs text-gray-500">
            <div className="flex items-center gap-4">
              <span>
                显示 {filteredLogs.length} / {logs.length} 条日志
              </span>
              {isPaused && (
                <span className="text-amber-400 flex items-center gap-1">
                  <Pause className="w-3 h-3" />
                  已暂停
                </span>
              )}
            </div>
            <div className="flex items-center gap-3">
              <span>服务: {SERVICES.find((s) => s.id === selectedService)?.name}</span>
              <div className="flex items-center gap-1">
                {getConnectionStatusIcon()}
                <span
                  className={
                    connectionStatus === "connected"
                      ? "text-emerald-400"
                      : connectionStatus === "error"
                      ? "text-rose-400"
                      : connectionStatus === "connecting"
                      ? "text-blue-400"
                      : "text-gray-500"
                  }
                >
                  {connectionStatus === "connected"
                    ? "已连接"
                    : connectionStatus === "connecting"
                    ? "连接中"
                    : connectionStatus === "error"
                    ? "连接错误"
                    : "未连接"}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// 高亮文本组件
function HighlightText({ text, highlight }: { text: string; highlight: string }) {
  if (!highlight.trim()) return <>{text}</>;

  const regex = new RegExp(`(${highlight.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")})`, "gi");
  const parts = text.split(regex);
  return (
    <>
      {parts.map((part, i) =>
        part.toLowerCase() === highlight.toLowerCase() ? (
          <span key={i} className="bg-amber-500/30 text-amber-200 px-0.5 rounded">
            {part}
          </span>
        ) : (
          part
        )
      )}
    </>
  );
}
