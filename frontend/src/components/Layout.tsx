import { useState, useEffect } from "react";
import { Link, Outlet, useNavigate, useLocation } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/ThemeToggle";
import { HelpMeWidget } from "@/components/HelpMeWidget";
import {
  Home,
  Users,
  BookUser,
  FileText,
  LogOut,
  User as UserIcon,
  Menu,
  Award,
  Terminal,
  Zap,
  ScrollText,
  ChevronDown,
  ChevronRight,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

interface MenuItem {
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  path: string;
  children?: MenuItem[];
}

const getMenuItems = (userType: string): MenuItem[] => {
  const baseItems: MenuItem[] = [
    { label: "仪表板", icon: Home, path: "/dashboard" },
    { label: "活动列表", icon: Award, path: "/activities" },
    { label: "申请列表", icon: FileText, path: "/applications" },
  ];

  if (userType === "teacher" || userType === "admin") {
    baseItems.push({ label: "学生列表", icon: Users, path: "/students" });
  }

  if (userType === "admin") {
    baseItems.push({ label: "教师列表", icon: BookUser, path: "/teachers" });
    // 管理员专属：开发者工具
    baseItems.push({
      label: "开发者工具",
      icon: Terminal,
      path: "/devtools",
      children: [
        { label: "API 测试器", icon: Zap, path: "/devtools/api-tester" },
        { label: "服务日志", icon: ScrollText, path: "/devtools/logs" },
      ],
    });
  }

  return baseItems;
};

export default function Layout() {
  const { logout, user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const [expandedMenus, setExpandedMenus] = useState<Record<string, boolean>>({});
  const [avatarError, setAvatarError] = useState(false);

  // Reset avatar error when user changes
  useEffect(() => {
    setAvatarError(false);
  }, [user?.avatar]);

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const getUserTypeLabel = (userType: string) => {
    const labels = {
      student: "学生",
      teacher: "教师",
      admin: "管理员",
    };
    return labels[userType as keyof typeof labels] || userType;
  };

  const toggleMenu = (path: string) => {
    setExpandedMenus((prev) => ({ ...prev, [path]: !prev[path] }));
  };

  const isMenuActive = (item: MenuItem): boolean => {
    if (location.pathname === item.path) return true;
    if (location.pathname.startsWith(item.path + "/")) return true;
    if (item.children) {
      return item.children.some((child) => isMenuActive(child));
    }
    return false;
  };

  // 根据用户类型获取菜单项
  const menuItems = getMenuItems(user?.userType || "");

  return (
    <div className="flex h-screen w-screen overflow-hidden">
      {/* 左侧栏 */}
      <aside className="fixed left-0 top-0 h-screen w-64 bg-background border-r flex flex-col justify-between z-30">
        <div>
          <div className="p-6 text-2xl font-bold tracking-tight">
            双创分申请平台
          </div>
          <nav className="flex flex-col gap-1 mt-4">
            {menuItems.map((item) => (
              <div key={item.path}>
                {item.children ? (
                  // 有子菜单的项
                  <>
                    <button
                      onClick={() => toggleMenu(item.path)}
                      className={`w-full flex items-center gap-3 px-6 py-3 rounded-lg transition-colors text-base font-medium hover:bg-primary/10 ${
                        isMenuActive(item)
                          ? "bg-primary/10 text-primary"
                          : "text-muted-foreground"
                      }`}
                    >
                      <item.icon className="h-5 w-5" />
                      <span className="flex-1 text-left">{item.label}</span>
                      {expandedMenus[item.path] || isMenuActive(item) ? (
                        <ChevronDown className="h-4 w-4" />
                      ) : (
                        <ChevronRight className="h-4 w-4" />
                      )}
                    </button>
                    {(expandedMenus[item.path] || isMenuActive(item)) && (
                      <div className="ml-6 mt-1 space-y-1">
                        {item.children.map((child) => (
                          <Link
                            key={child.path}
                            to={child.path}
                            className={`flex items-center gap-3 px-6 py-2 rounded-lg transition-colors text-sm font-medium hover:bg-primary/10 ${
                              location.pathname === child.path
                                ? "bg-primary/10 text-primary"
                                : "text-muted-foreground"
                            }`}
                          >
                            <child.icon className="h-4 w-4" />
                            {child.label}
                          </Link>
                        ))}
                      </div>
                    )}
                  </>
                ) : (
                  // 普通菜单项
                  <Link
                    to={item.path}
                    className={`flex items-center gap-3 px-6 py-3 rounded-lg transition-colors text-base font-medium hover:bg-primary/10 ${
                      isMenuActive(item)
                        ? "bg-primary/10 text-primary"
                        : "text-muted-foreground"
                    }`}
                  >
                    <item.icon className="h-5 w-5" />
                    {item.label}
                  </Link>
                )}
              </div>
            ))}
          </nav>
        </div>
        {/* 用户卡片和主题切换始终贴底 */}
        <div className="p-4 border-t bg-background">
          <div className="flex items-center gap-3 mb-3">
            <div className="rounded-full bg-muted w-10 h-10 flex items-center justify-center font-bold text-lg overflow-hidden">
              {user?.avatar && !avatarError ? (
                <img
                  src={user.avatar}
                  alt=""
                  className="w-full h-full object-cover"
                  onError={() => setAvatarError(true)}
                />
              ) : (
                user?.fullName?.[0] || user?.username?.[0] || "U"
              )}
            </div>
            <div className="flex-1 min-w-0">
              <div className="font-medium truncate">
                {user?.fullName || user?.username}
              </div>
              <div className="text-xs text-muted-foreground truncate">
                {getUserTypeLabel(user?.userType || "")}
              </div>
            </div>
            {/* User menu dropdown */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                >
                  <Menu className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56" align="end" forceMount>
                <DropdownMenuLabel className="font-normal">
                  <div className="flex flex-col space-y-1">
                    <p className="text-sm font-medium leading-none">
                      {user?.fullName || user?.username}
                    </p>
                    <p className="text-xs leading-none text-muted-foreground">
                      {getUserTypeLabel(user?.userType || "")}
                    </p>
                    {user?.email && (
                      <p className="text-xs leading-none text-muted-foreground">
                        {user.email}
                      </p>
                    )}
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem asChild>
                  <Link to="/profile">
                    <UserIcon className="mr-2 h-4 w-4" />
                    <span>个人资料</span>
                  </Link>
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={handleLogout}>
                  <LogOut className="mr-2 h-4 w-4" />
                  <span>退出登录</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
          {/* Theme toggle */}
          <div className="flex items-center justify-center pt-3 border-t">
            <ThemeToggle />
          </div>
        </div>
      </aside>
      {/* 右侧主内容区 */}
      <main className="flex-1 ml-64 h-screen overflow-y-auto bg-background">
        {/* Header */}
        <header className="flex h-16 items-center justify-between border-b bg-background px-4">
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="icon" className="sm:hidden">
              <Menu className="h-4 w-4" />
            </Button>
          </div>

          {/* Empty right side - components moved to sidebar */}
          <div className="flex items-center gap-4">
          </div>
        </header>

        {/* Main content area */}
        <main className="flex-1 overflow-auto">
          <Outlet />
        </main>
      </main>

      {/* HelpMe Widget - 全局帮助按钮 */}
      <HelpMeWidget />
    </div>
  );
}
