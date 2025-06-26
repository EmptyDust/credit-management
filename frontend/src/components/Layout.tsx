import { Link, Outlet, useNavigate, useLocation } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/ThemeToggle";
import {
  Home,
  Users,
  BookUser,
  FileText,
  LogOut,
  User as UserIcon,
  Settings,
  Menu,
  Award,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

// 定义菜单项配置
const getMenuItems = (userType: string) => {
  const baseItems = [
    { label: "仪表板", icon: Home, path: "/dashboard" },
    { label: "活动列表", icon: Award, path: "/activities" },
    { label: "申请列表", icon: FileText, path: "/applications" },
  ];

  // 只有教师和管理员可以看到学生管理
  if (userType === "teacher" || userType === "admin") {
    baseItems.push({ label: "学生列表", icon: Users, path: "/students" });
  }

  // 只有管理员可以看到教师管理
  if (userType === "admin") {
    baseItems.push({ label: "教师列表", icon: BookUser, path: "/teachers" });
  }

  return baseItems;
};

export default function Layout() {
  const { logout, user } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

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
              <Link
                key={item.path}
                to={item.path}
                className={`flex items-center gap-3 px-6 py-3 rounded-lg transition-colors text-base font-medium hover:bg-primary/10 ${
                  location.pathname.startsWith(item.path)
                    ? "bg-primary/10 text-primary"
                    : "text-muted-foreground"
                }`}
              >
                <item.icon className="h-5 w-5" />
                {item.label}
              </Link>
            ))}
          </nav>
        </div>
        {/* 用户卡片始终贴底 */}
        <div className="p-4 border-t flex items-center gap-3 bg-background">
          <div className="rounded-full bg-muted w-10 h-10 flex items-center justify-center font-bold text-lg">
            {user?.fullName?.[0] || user?.username?.[0] || "U"}
          </div>
          <div className="flex-1 min-w-0">
            <div className="font-medium truncate">
              {user?.fullName || user?.username}
            </div>
            <div className="text-xs text-muted-foreground truncate">
              {getUserTypeLabel(user?.userType || "")}
            </div>
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
            <div className="hidden sm:block">
              <h1 className="text-lg font-semibold">双创分申请平台</h1>
            </div>
          </div>

          <div className="flex items-center gap-4">
            {/* Theme toggle */}
            <ThemeToggle />

            {/* User menu */}
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button
                  variant="ghost"
                  className="relative h-8 w-8 rounded-full"
                >
                  <UserIcon className="h-5 w-5" />
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
        </header>

        {/* Main content area */}
        <main className="flex-1 overflow-auto">
          <Outlet />
        </main>
      </main>
    </div>
  );
}
