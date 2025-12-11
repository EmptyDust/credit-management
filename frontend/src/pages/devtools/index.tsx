import { Link, Outlet, useLocation } from "react-router-dom";
import { Terminal, Zap, ScrollText, Code2, Sparkles } from "lucide-react";

const DEVTOOLS_MENU = [
  {
    path: "/devtools/api-tester",
    label: "API 测试器",
    icon: Zap,
    desc: "测试和调试 API 端点",
    gradient: "from-amber-500 to-orange-600",
    iconBg: "bg-amber-500/10",
  },
  {
    path: "/devtools/logs",
    label: "服务日志",
    icon: ScrollText,
    desc: "查看服务运行日志",
    gradient: "from-emerald-500 to-teal-600",
    iconBg: "bg-emerald-500/10",
  },
];

export default function DevToolsLayout() {
  const location = useLocation();
  const isIndex = location.pathname === "/devtools";

  if (isIndex) {
    return (
      <div className="h-[calc(100vh-4rem)] bg-[#0d1117] flex items-center justify-center relative overflow-hidden">
        {/* 背景装饰 */}
        <div className="absolute inset-0 overflow-hidden pointer-events-none">
          <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/5 rounded-full blur-3xl" />
          <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/5 rounded-full blur-3xl" />
          <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPjxnIGZpbGw9IiMyMDIwMjAiIGZpbGwtb3BhY2l0eT0iMC40Ij48cGF0aCBkPSJNMzYgMzRoLTJ2LTRoMnY0em0wLTZoLTJ2LTRoMnY0em0wLTZoLTJ2LTRoMnY0em0wLTZoLTJWOGgydjh6bTAgMThoLTJ2LTRoMnY0em0wIDZoLTJ2LTRoMnY0em0wIDZoLTJ2LTRoMnY0em0wIDZoLTJ2LThoMnY4eiIvPjwvZz48L2c+PC9zdmc+')] opacity-30" />
        </div>

        <div className="text-center relative z-10">
          {/* Logo */}
          <div className="relative inline-block mb-8">
            <div className="absolute inset-0 bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl blur-xl opacity-30 animate-pulse" />
            <div className="relative inline-flex items-center justify-center w-24 h-24 rounded-2xl bg-gradient-to-br from-[#161b22] to-[#0d1117] border border-[#30363d] shadow-2xl">
              <Terminal className="w-12 h-12 text-blue-400" />
              <Sparkles className="absolute -top-2 -right-2 w-6 h-6 text-amber-400" />
            </div>
          </div>

          {/* 标题 */}
          <h1 className="text-3xl font-bold mb-3">
            <span className="bg-gradient-to-r from-blue-400 via-purple-400 to-pink-400 bg-clip-text text-transparent">
              开发者工具
            </span>
          </h1>
          <p className="text-gray-500 mb-10 flex items-center justify-center gap-2">
            <Code2 className="w-4 h-4" />
            调试、测试和监控你的微服务
          </p>

          {/* 工具卡片 */}
          <div className="flex gap-6 justify-center">
            {DEVTOOLS_MENU.map((item) => {
              const Icon = item.icon;
              return (
                <Link
                  key={item.path}
                  to={item.path}
                  className="group relative p-6 rounded-2xl bg-[#161b22]/80 backdrop-blur border border-[#30363d] hover:border-[#484f58] transition-all duration-300 w-56 hover:scale-105 hover:shadow-xl hover:shadow-black/20"
                >
                  {/* 卡片背景光效 */}
                  <div className={`absolute inset-0 rounded-2xl bg-gradient-to-br ${item.gradient} opacity-0 group-hover:opacity-5 transition-opacity duration-300`} />
                  
                  <div className={`inline-flex items-center justify-center w-14 h-14 rounded-xl ${item.iconBg} mb-4 group-hover:scale-110 transition-transform duration-300`}>
                    <Icon className={`w-7 h-7 bg-gradient-to-br ${item.gradient} bg-clip-text`} style={{ color: item.gradient.includes('amber') ? '#f59e0b' : '#10b981' }} />
                  </div>
                  <h3 className="text-lg font-semibold text-gray-100 mb-2">{item.label}</h3>
                  <p className="text-sm text-gray-500">{item.desc}</p>
                  
                  {/* 箭头指示 */}
                  <div className="absolute right-4 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 group-hover:translate-x-1 transition-all duration-300">
                    <svg className="w-5 h-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                    </svg>
                  </div>
                </Link>
              );
            })}
          </div>

          {/* 快捷键提示 */}
          <div className="mt-10 text-xs text-gray-600">
            <span className="px-2 py-1 bg-[#161b22] rounded border border-[#30363d] font-mono">
              仅管理员可访问
            </span>
          </div>
        </div>
      </div>
    );
  }

  return <Outlet />;
}

