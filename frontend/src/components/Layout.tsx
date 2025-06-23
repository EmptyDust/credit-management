import { Link, Outlet, useNavigate } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import { Button } from "@/components/ui/button";
import { ThemeToggle } from "@/components/ThemeToggle";
import { Home, Users, BookUser, FileText, FileCheck, LogOut } from 'lucide-react';

const navItems = [
    { href: "/dashboard", label: "Dashboard", icon: <Home className="h-4 w-4" /> },
    { href: "/students", label: "Students", icon: <Users className="h-4 w-4" /> },
    { href: "/teachers", label: "Teachers", icon: <BookUser className="h-4 w-4" /> },
    { href: "/affairs", label: "Affairs", icon: <FileText className="h-4 w-4" /> },
    { href: "/applications", label: "Applications", icon: <FileCheck className="h-4 w-4" /> },
]

export default function Layout() {
    const { logout, user } = useAuth();
    const navigate = useNavigate();

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <div className="flex min-h-screen w-full">
            <aside className="hidden w-64 flex-col border-r bg-background p-4 sm:flex">
                <nav className="flex flex-col gap-2">
                    <h2 className="mb-4 text-lg font-semibold tracking-tight">
                        Credit Management
                    </h2>
                    {navItems.map((item) => (
                        <Button asChild key={item.href} variant="ghost" className="justify-start">
                            <Link to={item.href} className="flex items-center gap-2">
                                {item.icon}
                                {item.label}
                            </Link>
                        </Button>
                    ))}
                </nav>
            </aside>
            <div className="flex flex-1 flex-col">
                <header className="flex h-16 items-center justify-between border-b bg-background px-4">
                    <div className="text-lg font-semibold">Welcome, {user?.username} ({user?.userType})</div>
                    <div className="flex items-center gap-4">
                        <ThemeToggle />
                        <Button onClick={handleLogout} variant="outline">
                            <LogOut className="mr-2 h-4 w-4" />
                            Logout
                        </Button>
                    </div>
                </header>
                <main className="flex-1 p-6">
                    <Outlet />
                </main>
            </div>
        </div>
    );
} 