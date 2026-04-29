import { useState, useEffect, useCallback } from "react";
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
  Clock,
  ChevronRight,
  ChevronDown,
  Star,
  StarOff,
  Copy,
  Check,
  Loader2,
  Folder,
  FolderOpen,
  History,
  Trash2,
  RotateCcw,
  Send,
} from "lucide-react";

// API 端点定义
const API_ENDPOINTS = {
  认证服务: [
    { method: "POST", path: "/api/auth/login", desc: "用户登录", needAuth: false },
    { method: "POST", path: "/api/auth/register", desc: "用户注册", needAuth: false },
    { method: "POST", path: "/api/auth/logout", desc: "用户登出", needAuth: true },
    { method: "POST", path: "/api/auth/refresh", desc: "刷新Token", needAuth: true },
  ],
  用户服务: [
    { method: "GET", path: "/api/users/profile", desc: "获取个人资料", needAuth: true },
    { method: "PUT", path: "/api/users/profile", desc: "更新个人资料", needAuth: true },
    { method: "GET", path: "/api/users/stats", desc: "用户统计", needAuth: true },
    { method: "GET", path: "/api/users/:id", desc: "获取用户详情", needAuth: true },
    { method: "POST", path: "/api/users/change_password", desc: "修改密码", needAuth: true },
    { method: "GET", path: "/api/users/stats/students", desc: "学生统计", needAuth: true },
    { method: "GET", path: "/api/users/stats/teachers", desc: "教师统计", needAuth: true },
    { method: "POST", path: "/api/users/teachers", desc: "创建教师", needAuth: true },
    { method: "POST", path: "/api/users/students", desc: "创建学生", needAuth: true },
    { method: "DELETE", path: "/api/users/:id", desc: "删除用户", needAuth: true },
    { method: "POST", path: "/api/users/reset_password", desc: "重置密码", needAuth: true },
  ],
  学生服务: [
    { method: "POST", path: "/api/students/register", desc: "学生注册", needAuth: false },
    { method: "POST", path: "/api/students", desc: "创建学生", needAuth: true },
    { method: "PUT", path: "/api/students/:id", desc: "更新学生", needAuth: true },
    { method: "DELETE", path: "/api/students/:id", desc: "删除学生", needAuth: true },
  ],
  教师服务: [
    { method: "POST", path: "/api/teachers", desc: "创建教师", needAuth: true },
    { method: "PUT", path: "/api/teachers/:id", desc: "更新教师", needAuth: true },
    { method: "DELETE", path: "/api/teachers/:id", desc: "删除教师", needAuth: true },
  ],
  活动服务: [
    { method: "GET", path: "/api/activities", desc: "获取活动列表", needAuth: true },
    { method: "POST", path: "/api/activities", desc: "创建活动", needAuth: true },
    { method: "GET", path: "/api/activities/:id", desc: "获取活动详情", needAuth: true },
    { method: "PUT", path: "/api/activities/:id", desc: "更新活动", needAuth: true },
    { method: "DELETE", path: "/api/activities/:id", desc: "删除活动", needAuth: true },
    { method: "GET", path: "/api/activities/stats", desc: "活动统计", needAuth: true },
    { method: "GET", path: "/api/activities/categories", desc: "活动分类", needAuth: true },
    { method: "GET", path: "/api/activities/templates", desc: "活动模板", needAuth: true },
    { method: "POST", path: "/api/activities/:id/submit", desc: "提交活动", needAuth: true },
    { method: "POST", path: "/api/activities/:id/review", desc: "审核活动", needAuth: true },
    { method: "GET", path: "/api/activities/pending", desc: "待审核活动", needAuth: true },
  ],
  申请服务: [
    { method: "GET", path: "/api/applications", desc: "获取申请列表", needAuth: true },
    { method: "GET", path: "/api/applications/:id", desc: "获取申请详情", needAuth: true },
    { method: "GET", path: "/api/applications/stats", desc: "申请统计", needAuth: true },
    { method: "GET", path: "/api/applications/all", desc: "所有申请(管理)", needAuth: true },
  ],
  搜索服务: [
    { method: "GET", path: "/api/search/users", desc: "搜索用户", needAuth: true },
    { method: "GET", path: "/api/search/activities", desc: "搜索活动", needAuth: true },
    { method: "GET", path: "/api/search/applications", desc: "搜索申请", needAuth: true },
  ],
  系统服务: [
    { method: "GET", path: "/health", desc: "健康检查", needAuth: false },
    { method: "GET", path: "/api/config/options", desc: "配置选项", needAuth: false },
  ],
};

const METHOD_COLORS: Record<string, string> = {
  GET: "bg-emerald-500/20 text-emerald-400 border-emerald-500/30",
  POST: "bg-blue-500/20 text-blue-400 border-blue-500/30",
  PUT: "bg-amber-500/20 text-amber-400 border-amber-500/30",
  DELETE: "bg-rose-500/20 text-rose-400 border-rose-500/30",
  PATCH: "bg-purple-500/20 text-purple-400 border-purple-500/30",
};

interface RequestHistory {
  id: string;
  method: string;
  url: string;
  body?: string;
  headers?: Record<string, string>;
  timestamp: number;
  response?: {
    status: number;
    data: unknown;
    time: number;
  };
}

export default function ApiTester() {
  // 从 localStorage 获取 token
  const getToken = () => localStorage.getItem("token");
  const [method, setMethod] = useState("GET");
  const [url, setUrl] = useState("");
  const [body, setBody] = useState("");
  const [headers, setHeaders] = useState<Record<string, string>>({
    "Content-Type": "application/json",
  });
  const [response, setResponse] = useState<{
    status: number;
    statusText: string;
    data: unknown;
    time: number;
    headers: Record<string, string>;
  } | null>(null);
  const [loading, setLoading] = useState(false);
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({});
  const [favorites, setFavorites] = useState<string[]>(() => {
    const saved = localStorage.getItem("api-tester-favorites");
    return saved ? JSON.parse(saved) : [];
  });
  const [history, setHistory] = useState<RequestHistory[]>(() => {
    const saved = localStorage.getItem("api-tester-history");
    return saved ? JSON.parse(saved) : [];
  });
  const [copied, setCopied] = useState(false);
  const [activeTab, setActiveTab] = useState<"headers" | "body" | "history">("body");
  const [searchEndpoint, setSearchEndpoint] = useState("");

  // 保存收藏
  useEffect(() => {
    localStorage.setItem("api-tester-favorites", JSON.stringify(favorites));
  }, [favorites]);

  // 保存历史
  useEffect(() => {
    localStorage.setItem("api-tester-history", JSON.stringify(history.slice(0, 50)));
  }, [history]);

  const toggleGroup = (group: string) => {
    setExpandedGroups((prev) => ({ ...prev, [group]: !prev[group] }));
  };

  const selectEndpoint = (endpoint: { method: string; path: string }) => {
    setMethod(endpoint.method);
    setUrl(endpoint.path);
    setResponse(null);
  };

  const toggleFavorite = (path: string) => {
    setFavorites((prev) =>
      prev.includes(path) ? prev.filter((p) => p !== path) : [...prev, path]
    );
  };

  const sendRequest = useCallback(async () => {
    if (!url) return;

    setLoading(true);
    const startTime = performance.now();

    try {
      const requestHeaders: Record<string, string> = { ...headers };
      const token = getToken();
      if (token) {
        requestHeaders["Authorization"] = `Bearer ${token}`;
      }

      const options: RequestInit = {
        method,
        headers: requestHeaders,
      };

      if (method !== "GET" && method !== "HEAD" && body) {
        options.body = body;
      }

      const res = await fetch(url, options);
      const endTime = performance.now();
      
      let data: unknown;
      const contentType = res.headers.get("content-type");
      if (contentType?.includes("application/json")) {
        data = await res.json();
      } else {
        data = await res.text();
      }

      const responseHeaders: Record<string, string> = {};
      res.headers.forEach((value, key) => {
        responseHeaders[key] = value;
      });

      const responseData = {
        status: res.status,
        statusText: res.statusText,
        data,
        time: Math.round(endTime - startTime),
        headers: responseHeaders,
      };

      setResponse(responseData);

      // 添加到历史记录
      const historyItem: RequestHistory = {
        id: Date.now().toString(),
        method,
        url,
        body: body || undefined,
        headers: requestHeaders,
        timestamp: Date.now(),
        response: {
          status: res.status,
          data,
          time: responseData.time,
        },
      };
      setHistory((prev) => [historyItem, ...prev]);
    } catch (error) {
      const endTime = performance.now();
      setResponse({
        status: 0,
        statusText: "Network Error",
        data: { error: error instanceof Error ? error.message : "Unknown error" },
        time: Math.round(endTime - startTime),
        headers: {},
      });
    } finally {
      setLoading(false);
    }
  }, [url, method, body, headers]);

  const copyResponse = () => {
    if (response) {
      navigator.clipboard.writeText(JSON.stringify(response.data, null, 2));
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  const getStatusColor = (status: number) => {
    if (status >= 200 && status < 300) return "text-emerald-400";
    if (status >= 300 && status < 400) return "text-blue-400";
    if (status >= 400 && status < 500) return "text-amber-400";
    return "text-rose-400";
  };

  return (
    <div className="h-[calc(100vh-4rem)] bg-[#0d1117] text-gray-100 font-mono overflow-hidden">
      <div className="h-full flex">
        {/* 左侧 - API 端点列表 */}
        <div className="w-80 border-r border-[#30363d] flex flex-col">
          <div className="p-4 border-b border-[#30363d]">
            <h2 className="text-sm font-semibold text-gray-300 mb-3">API 端点</h2>
            <Input
              value={searchEndpoint}
              onChange={(e) => setSearchEndpoint(e.target.value)}
              placeholder="搜索端点..."
              className="bg-[#161b22] border-[#30363d] text-gray-100 placeholder:text-gray-500 h-8 text-sm"
            />
          </div>
          <div className="flex-1 overflow-y-auto">
            {Object.entries(API_ENDPOINTS).map(([group, endpoints]) => {
              // 过滤端点
              const filteredEndpoints = searchEndpoint
                ? endpoints.filter(
                    (e) =>
                      e.path.toLowerCase().includes(searchEndpoint.toLowerCase()) ||
                      e.desc.toLowerCase().includes(searchEndpoint.toLowerCase())
                  )
                : endpoints;
              
              if (filteredEndpoints.length === 0) return null;
              
              return (
              <div key={group} className="border-b border-[#30363d]">
                <button
                  onClick={() => toggleGroup(group)}
                  className="w-full px-4 py-2.5 flex items-center gap-2 hover:bg-[#161b22] transition-colors text-left"
                >
                  {expandedGroups[group] || searchEndpoint ? (
                    <>
                      <ChevronDown className="w-4 h-4 text-gray-500" />
                      <FolderOpen className="w-4 h-4 text-amber-400" />
                    </>
                  ) : (
                    <>
                      <ChevronRight className="w-4 h-4 text-gray-500" />
                      <Folder className="w-4 h-4 text-amber-400" />
                    </>
                  )}
                  <span className="text-sm text-gray-300">{group}</span>
                  <span className="ml-auto text-xs text-gray-500">{filteredEndpoints.length}</span>
                </button>
                {(expandedGroups[group] || searchEndpoint) && (
                  <div className="pb-2">
                    {filteredEndpoints.map((endpoint) => (
                      <div
                        key={endpoint.path + endpoint.method}
                        className="group px-4 py-1.5 flex items-center gap-2 hover:bg-[#161b22] cursor-pointer"
                        onClick={() => selectEndpoint(endpoint)}
                      >
                        <span
                          className={`text-[10px] font-bold px-1.5 py-0.5 rounded border ${METHOD_COLORS[endpoint.method]}`}
                        >
                          {endpoint.method}
                        </span>
                        <span className="text-xs text-gray-400 truncate flex-1" title={endpoint.desc}>
                          {endpoint.path}
                        </span>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            toggleFavorite(endpoint.path);
                          }}
                          className="opacity-0 group-hover:opacity-100 transition-opacity"
                        >
                          {favorites.includes(endpoint.path) ? (
                            <Star className="w-3.5 h-3.5 text-amber-400 fill-amber-400" />
                          ) : (
                            <StarOff className="w-3.5 h-3.5 text-gray-500" />
                          )}
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            );
            })}
          </div>
        </div>

        {/* 右侧 - 请求/响应区域 */}
        <div className="flex-1 flex flex-col">
          {/* 请求构建器 */}
          <div className="p-4 border-b border-[#30363d]">
            <div className="flex gap-2">
              <Select value={method} onValueChange={setMethod}>
                <SelectTrigger className={`w-28 h-10 bg-[#161b22] border-[#30363d] ${METHOD_COLORS[method]} font-bold text-xs`}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent className="bg-[#161b22] border-[#30363d]">
                  {["GET", "POST", "PUT", "DELETE", "PATCH"].map((m) => (
                    <SelectItem key={m} value={m} className={`${METHOD_COLORS[m]} font-bold text-xs`}>
                      {m}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <Input
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                placeholder="输入请求 URL..."
                className="flex-1 h-10 bg-[#161b22] border-[#30363d] text-gray-100 placeholder:text-gray-500"
                onKeyDown={(e) => e.key === "Enter" && sendRequest()}
              />
              <Button
                onClick={sendRequest}
                disabled={loading || !url}
                className="h-10 px-6 bg-[#238636] hover:bg-[#2ea043] text-white border-0"
              >
                {loading ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <>
                    <Play className="w-4 h-4 mr-2" />
                    发送
                  </>
                )}
              </Button>
            </div>

            {/* Tabs */}
            <div className="mt-4">
              <div className="flex gap-4 border-b border-[#30363d]">
                <button
                  onClick={() => setActiveTab("body")}
                  className={`pb-2 text-sm ${
                    activeTab === "body"
                      ? "text-blue-400 border-b-2 border-blue-400"
                      : "text-gray-500 hover:text-gray-300"
                  }`}
                >
                  Body
                </button>
                <button
                  onClick={() => setActiveTab("headers")}
                  className={`pb-2 text-sm ${
                    activeTab === "headers"
                      ? "text-blue-400 border-b-2 border-blue-400"
                      : "text-gray-500 hover:text-gray-300"
                  }`}
                >
                  Headers
                </button>
                <button
                  onClick={() => setActiveTab("history")}
                  className={`pb-2 text-sm flex items-center gap-1 ${
                    activeTab === "history"
                      ? "text-blue-400 border-b-2 border-blue-400"
                      : "text-gray-500 hover:text-gray-300"
                  }`}
                >
                  <History className="w-3 h-3" />
                  历史记录
                  {history.length > 0 && (
                    <span className="ml-1 px-1.5 py-0.5 text-[10px] bg-[#30363d] rounded-full">
                      {history.length}
                    </span>
                  )}
                </button>
              </div>
              <div className="mt-3">
                {activeTab === "body" && (
                  <textarea
                    value={body}
                    onChange={(e) => setBody(e.target.value)}
                    placeholder='{"key": "value"}'
                    className="w-full h-32 p-3 bg-[#161b22] border border-[#30363d] rounded-md text-gray-100 text-sm resize-none focus:outline-none focus:border-blue-500"
                  />
                )}
                {activeTab === "headers" && (
                  <div className="space-y-2">
                    <div className="flex gap-2 text-xs text-gray-500">
                      <span className="w-1/3">Key</span>
                      <span className="flex-1">Value</span>
                    </div>
                    {Object.entries(headers).map(([key, value]) => (
                      <div key={key} className="flex gap-2">
                        <Input
                          value={key}
                          readOnly
                          className="w-1/3 h-8 bg-[#161b22] border-[#30363d] text-gray-300 text-sm"
                        />
                        <Input
                          value={value}
                          onChange={(e) =>
                            setHeaders((prev) => ({ ...prev, [key]: e.target.value }))
                          }
                          className="flex-1 h-8 bg-[#161b22] border-[#30363d] text-gray-300 text-sm"
                        />
                      </div>
                    ))}
                    <div className="text-xs text-gray-500 mt-2">
                      ✓ Authorization header 将自动注入当前用户 Token
                    </div>
                  </div>
                )}
                {activeTab === "history" && (
                  <div className="space-y-1 max-h-32 overflow-y-auto">
                    {history.length === 0 ? (
                      <div className="text-xs text-gray-500 py-4 text-center">
                        暂无历史记录
                      </div>
                    ) : (
                      <>
                        <div className="flex justify-end mb-2">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setHistory([])}
                            className="h-6 text-xs text-gray-500 hover:text-rose-400"
                          >
                            <Trash2 className="w-3 h-3 mr-1" />
                            清空
                          </Button>
                        </div>
                        {history.slice(0, 10).map((item) => (
                          <div
                            key={item.id}
                            className="flex items-center gap-2 px-2 py-1.5 rounded hover:bg-[#161b22] cursor-pointer group"
                            onClick={() => {
                              setMethod(item.method);
                              setUrl(item.url);
                              if (item.body) setBody(item.body);
                              setActiveTab("body");
                            }}
                          >
                            <span
                              className={`text-[10px] font-bold px-1.5 py-0.5 rounded border ${METHOD_COLORS[item.method]}`}
                            >
                              {item.method}
                            </span>
                            <span className="text-xs text-gray-400 truncate flex-1">
                              {item.url}
                            </span>
                            {item.response && (
                              <span
                                className={`text-[10px] ${
                                  item.response.status >= 200 && item.response.status < 300
                                    ? "text-emerald-400"
                                    : item.response.status >= 400
                                    ? "text-rose-400"
                                    : "text-gray-400"
                                }`}
                              >
                                {item.response.status}
                              </span>
                            )}
                            <span className="text-[10px] text-gray-600">
                              {new Date(item.timestamp).toLocaleTimeString()}
                            </span>
                            <RotateCcw className="w-3 h-3 text-gray-600 opacity-0 group-hover:opacity-100" />
                          </div>
                        ))}
                      </>
                    )}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* 响应区域 */}
          <div className="flex-1 flex flex-col overflow-hidden">
            <div className="px-4 py-2 border-b border-[#30363d] flex items-center justify-between">
              <div className="flex items-center gap-4">
                <span className="text-sm text-gray-400">响应</span>
                {response && (
                  <>
                    <span className={`text-sm font-bold ${getStatusColor(response.status)}`}>
                      {response.status} {response.statusText}
                    </span>
                    <span className="text-xs text-gray-500 flex items-center gap-1">
                      <Clock className="w-3 h-3" />
                      {response.time}ms
                    </span>
                  </>
                )}
              </div>
              {response && (
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={copyResponse}
                  className="h-7 text-xs text-gray-400 hover:text-gray-200"
                >
                  {copied ? (
                    <>
                      <Check className="w-3 h-3 mr-1" />
                      已复制
                    </>
                  ) : (
                    <>
                      <Copy className="w-3 h-3 mr-1" />
                      复制
                    </>
                  )}
                </Button>
              )}
            </div>
            <div className="flex-1 overflow-auto p-4 bg-[#0d1117]">
              {response ? (
                <JsonViewer data={response.data} />
              ) : (
                <div className="h-full flex items-center justify-center text-gray-500 text-sm">
                  <div className="text-center">
                    <Send className="w-12 h-12 mx-auto mb-4 opacity-20" />
                    <p>发送请求以查看响应</p>
                    <p className="text-xs mt-1 text-gray-600">按 Enter 快速发送</p>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// JSON 语法高亮组件
function JsonViewer({ data }: { data: unknown }) {
  const jsonString = typeof data === "string" ? data : JSON.stringify(data, null, 2);
  
  // 简单的 JSON 语法高亮
  const highlightJson = (json: string) => {
    return json
      .replace(/(".*?")(?=\s*:)/g, '<span class="text-blue-400">$1</span>') // keys
      .replace(/:\s*(".*?")/g, ': <span class="text-emerald-400">$1</span>') // string values
      .replace(/:\s*(\d+\.?\d*)/g, ': <span class="text-amber-400">$1</span>') // numbers
      .replace(/:\s*(true|false)/g, ': <span class="text-purple-400">$1</span>') // booleans
      .replace(/:\s*(null)/g, ': <span class="text-rose-400">$1</span>'); // null
  };

  return (
    <pre 
      className="text-sm text-gray-300 whitespace-pre-wrap"
      dangerouslySetInnerHTML={{ __html: highlightJson(jsonString) }}
    />
  );
}

