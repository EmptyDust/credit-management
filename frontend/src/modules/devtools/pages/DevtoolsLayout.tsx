import { Code2, ScrollText, Sparkles, Terminal, Zap } from "lucide-react";
import { Link, Outlet, useLocation } from "react-router-dom";

import { isApiTesterEnabled } from "@/app/config/devtools";

const DEVTOOLS_MENU = [
  ...(isApiTesterEnabled()
    ? [
        {
          path: "/devtools/api-tester",
          label: "API 测试器",
          icon: Zap,
          desc: "测试和调试 API 端点",
          gradient: "from-amber-500 to-orange-600",
          iconBg: "bg-amber-500/10",
        },
      ]
    : []),
  {
    path: "/devtools/logs",
    label: "服务日志",
    icon: ScrollText,
    desc: "查看服务运行日志",
    gradient: "from-emerald-500 to-teal-600",
    iconBg: "bg-emerald-500/10",
  },
];

export default function DevtoolsLayout() {
  const location = useLocation();
  const isIndex = location.pathname === "/devtools";

  if (!isIndex) {
    return <Outlet />;
  }

  return (
    <div className="relative flex h-[calc(100vh-4rem)] items-center justify-center overflow-hidden bg-[#0d1117]">
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div className="absolute left-1/4 top-1/4 h-96 w-96 rounded-full bg-blue-500/5 blur-3xl" />
        <div className="absolute bottom-1/4 right-1/4 h-96 w-96 rounded-full bg-cyan-500/5 blur-3xl" />
      </div>

      <div className="relative z-10 text-center">
        <div className="relative mb-8 inline-block">
          <div className="absolute inset-0 animate-pulse rounded-2xl bg-gradient-to-br from-blue-500 to-cyan-600 opacity-30 blur-xl" />
          <div className="relative inline-flex h-24 w-24 items-center justify-center rounded-2xl border border-[#30363d] bg-gradient-to-br from-[#161b22] to-[#0d1117] shadow-2xl">
            <Terminal className="h-12 w-12 text-blue-400" />
            <Sparkles className="absolute -right-2 -top-2 h-6 w-6 text-amber-400" />
          </div>
        </div>

        <h1 className="mb-3 text-3xl font-bold">
          <span className="bg-gradient-to-r from-blue-400 via-cyan-300 to-emerald-300 bg-clip-text text-transparent">
            开发者工具
          </span>
        </h1>
        <p className="mb-10 flex items-center justify-center gap-2 text-gray-500">
          <Code2 className="h-4 w-4" />
          调试、测试和监控当前应用服务
        </p>

        <div className="flex justify-center gap-6">
          {DEVTOOLS_MENU.map((item) => {
            const Icon = item.icon;

            return (
              <Link
                key={item.path}
                to={item.path}
                className="group relative w-56 rounded-2xl border border-[#30363d] bg-[#161b22]/80 p-6 backdrop-blur transition-all duration-300 hover:scale-105 hover:border-[#484f58] hover:shadow-xl hover:shadow-black/20"
              >
                <div
                  className={`absolute inset-0 rounded-2xl bg-gradient-to-br ${item.gradient} opacity-0 transition-opacity duration-300 group-hover:opacity-5`}
                />
                <div
                  className={`mb-4 inline-flex h-14 w-14 items-center justify-center rounded-xl ${item.iconBg} transition-transform duration-300 group-hover:scale-110`}
                >
                  <Icon className="h-7 w-7 text-gray-100" />
                </div>
                <h3 className="mb-2 text-lg font-semibold text-gray-100">
                  {item.label}
                </h3>
                <p className="text-sm text-gray-500">{item.desc}</p>
              </Link>
            );
          })}
        </div>

        <div className="mt-10 text-xs text-gray-600">
          <span className="rounded border border-[#30363d] bg-[#161b22] px-2 py-1 font-mono">
            仅管理员可访问
          </span>
        </div>
      </div>
    </div>
  );
}
