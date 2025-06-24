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
    Bell
} from 'lucide-react';
import { 
    DropdownMenu, 
    DropdownMenuContent, 
    DropdownMenuItem, 
    DropdownMenuLabel, 
    DropdownMenuSeparator, 
    DropdownMenuTrigger 
} from "@/components/ui/dropdown-menu";

const menuItems = [
    { label: "仪表板", icon: Home, path: "/dashboard" },
    { label: "事务管理", icon: Award, path: "/affairs" },
    { label: "申请管理", icon: FileText, path: "/applications" },
    { label: "学生管理", icon: Users, path: "/students" },
    { label: "教师管理", icon: BookUser, path: "/teachers" },
];

export default function Layout() {
    const { logout, user, hasPermission } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    const getUserTypeLabel = (userType: string) => {
        const labels = {
            'student': '学生',
            'teacher': '教师',
            'admin': '管理员'
        };
        return labels[userType as keyof typeof labels] || userType;
    };

    return (
        <div className="flex h-screen w-screen overflow-hidden">
            {/* 左侧栏 */}
            <aside className="fixed left-0 top-0 h-screen w-64 bg-background border-r flex flex-col justify-between z-30">
                <div>
                    <div className="p-6 text-2xl font-bold tracking-tight">学分管理系统</div>
                    <nav className="flex flex-col gap-1 mt-4">
                        {menuItems.map(item => (
                            <Link
                                key={item.path}
                                to={item.path}
                                className={`flex items-center gap-3 px-6 py-3 rounded-lg transition-colors text-base font-medium hover:bg-primary/10 ${location.pathname.startsWith(item.path) ? 'bg-primary/10 text-primary' : 'text-muted-foreground'}`}
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
                        {user?.fullName?.[0] || user?.username?.[0] || 'U'}
                    </div>
                    <div className="flex-1 min-w-0">
                        <div className="font-medium truncate">{user?.fullName || user?.username}</div>
                        <div className="text-xs text-muted-foreground truncate">{getUserTypeLabel(user?.userType || '')}</div>
                    </div>
                </div>
            </aside>
            {/* 右侧主内容区 */}
            <main className="flex-1 ml-64 h-screen overflow-y-auto bg-background">
                {/* Header */}
                <header className="flex h-16 items-center justify-between border-b bg-background px-4">
                    <div className="flex items-center gap-4">
                        <Button
                            variant="ghost"
                            size="icon"
                            className="sm:hidden"
                        >
                            <Menu className="h-4 w-4" />
                        </Button>
                        <div className="hidden sm:block">
                            <h1 className="text-lg font-semibold">学分管理系统</h1>
                        </div>
                    </div>

                    <div className="flex items-center gap-4">
                        {/* Notifications */}
                        <div className="relative">
                            <button className="relative p-2 rounded-full hover:bg-muted transition-colors">
                                <Bell className="h-6 w-6" />
                                <span className="absolute -top-1 -right-1 bg-red-500 text-white text-xs font-bold rounded-full px-1.5 py-0.5 flex items-center justify-center min-w-[18px] min-h-[18px] leading-none" style={{transform: 'translate(50%,-50%)'}}>
                                    3
                                </span>
                            </button>
                        </div>

                        {/* Theme toggle */}
                        <ThemeToggle />

                        {/* User menu */}
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button variant="ghost" className="relative h-8 w-8 rounded-full">
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
                                            {getUserTypeLabel(user?.userType || '')}
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
                                <DropdownMenuItem asChild>
                                    <Link to="/settings">
                                        <Settings className="mr-2 h-4 w-4" />
                                        <span>设置</span>
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