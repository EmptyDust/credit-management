import React, { useState } from 'react';
import { Outlet, Link, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { useTheme } from '../contexts/ThemeContext';
import { Button } from './ui/button';
import { Card } from './ui/card';
import {
    LayoutDashboard,
    FileText,
    Users,
    UserCheck,
    Settings,
    LogOut,
    Menu,
    X,
    Sun,
    Moon,
    Monitor
} from 'lucide-react';

const Layout: React.FC = () => {
    const { user, logout } = useAuth();
    const { theme, setTheme } = useTheme();
    const location = useLocation();
    const navigate = useNavigate();
    const [sidebarOpen, setSidebarOpen] = useState(false);

    const navigation = [
        { name: '仪表板', href: '/dashboard', icon: LayoutDashboard },
        { name: '申请管理', href: '/applications', icon: FileText },
        { name: '学生管理', href: '/students', icon: Users },
        { name: '教师管理', href: '/teachers', icon: UserCheck },
        { name: '事项管理', href: '/affairs', icon: Settings },
    ];

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    const isActive = (href: string) => {
        return location.pathname === href;
    };

    return (
        <div className="min-h-screen bg-background">
            {/* 移动端侧边栏遮罩 */}
            {sidebarOpen && (
                <div
                    className="fixed inset-0 z-40 bg-black/50 lg:hidden"
                    onClick={() => setSidebarOpen(false)}
                />
            )}

            {/* 侧边栏 */}
            <div className={`fixed inset-y-0 left-0 z-50 w-64 bg-card border-r border-border transform transition-transform duration-200 ease-in-out lg:translate-x-0 lg:static lg:inset-0 ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'
                }`}>
                <div className="flex flex-col h-full">
                    {/* Logo */}
                    <div className="flex items-center justify-between h-16 px-6 border-b border-border">
                        <h1 className="text-xl font-bold text-foreground">学分管理系统</h1>
                        <Button
                            variant="ghost"
                            size="sm"
                            className="lg:hidden"
                            onClick={() => setSidebarOpen(false)}
                        >
                            <X className="h-4 w-4" />
                        </Button>
                    </div>

                    {/* 导航菜单 */}
                    <nav className="flex-1 px-4 py-6 space-y-2">
                        {navigation.map((item) => {
                            const Icon = item.icon;
                            return (
                                <Link
                                    key={item.name}
                                    to={item.href}
                                    className={`flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors ${isActive(item.href)
                                            ? 'bg-primary text-primary-foreground'
                                            : 'text-muted-foreground hover:text-foreground hover:bg-muted'
                                        }`}
                                    onClick={() => setSidebarOpen(false)}
                                >
                                    <Icon className="mr-3 h-4 w-4" />
                                    {item.name}
                                </Link>
                            );
                        })}
                    </nav>

                    {/* 底部用户信息和设置 */}
                    <div className="p-4 border-t border-border">
                        {/* 主题切换 */}
                        <div className="flex items-center justify-between mb-4">
                            <span className="text-sm text-muted-foreground">主题</span>
                            <div className="flex items-center space-x-1">
                                <Button
                                    variant={theme === 'light' ? 'default' : 'ghost'}
                                    size="sm"
                                    onClick={() => setTheme('light')}
                                >
                                    <Sun className="h-4 w-4" />
                                </Button>
                                <Button
                                    variant={theme === 'dark' ? 'default' : 'ghost'}
                                    size="sm"
                                    onClick={() => setTheme('dark')}
                                >
                                    <Moon className="h-4 w-4" />
                                </Button>
                                <Button
                                    variant={theme === 'system' ? 'default' : 'ghost'}
                                    size="sm"
                                    onClick={() => setTheme('system')}
                                >
                                    <Monitor className="h-4 w-4" />
                                </Button>
                            </div>
                        </div>

                        {/* 用户信息 */}
                        <Card className="p-3">
                            <div className="flex items-center justify-between">
                                <div className="flex-1 min-w-0">
                                    <p className="text-sm font-medium text-foreground truncate">
                                        {user?.username}
                                    </p>
                                    <p className="text-xs text-muted-foreground capitalize">
                                        {user?.user_type}
                                    </p>
                                </div>
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={handleLogout}
                                    className="ml-2"
                                >
                                    <LogOut className="h-4 w-4" />
                                </Button>
                            </div>
                        </Card>
                    </div>
                </div>
            </div>

            {/* 主内容区域 */}
            <div className="lg:pl-64">
                {/* 顶部导航栏 */}
                <header className="sticky top-0 z-30 flex items-center justify-between h-16 px-6 bg-background border-b border-border">
                    <Button
                        variant="ghost"
                        size="sm"
                        className="lg:hidden"
                        onClick={() => setSidebarOpen(true)}
                    >
                        <Menu className="h-4 w-4" />
                    </Button>

                    <div className="flex items-center space-x-4">
                        <h2 className="text-lg font-semibold text-foreground">
                            {navigation.find(item => isActive(item.href))?.name || '仪表板'}
                        </h2>
                    </div>

                    <div className="flex items-center space-x-2">
                        <Link to="/profile">
                            <Button variant="ghost" size="sm">
                                <Users className="h-4 w-4 mr-2" />
                                个人资料
                            </Button>
                        </Link>
                    </div>
                </header>

                {/* 页面内容 */}
                <main className="p-6">
                    <Outlet />
                </main>
            </div>
        </div>
    );
};

export default Layout; 