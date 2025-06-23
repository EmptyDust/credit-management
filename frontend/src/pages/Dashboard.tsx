import { useEffect, useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import apiClient from '@/lib/api';
import { Users, UserCheck, School, Briefcase, FileText, GitPullRequest, Hourglass } from 'lucide-react';
import { Link } from 'react-router-dom';

interface UserStats {
  total_users: number;
  student_users: number;
  teacher_users: number;
  admin_users: number;
}

// A placeholder for application stats, as the API is not yet available
interface AppStats {
    total_applications: number;
    pending_applications: number;
    approved_applications: number;
}

const StatCard = ({ title, value, icon: Icon, to, description }: { title: string, value: string | number, icon: React.ElementType, to?: string, description?: string }) => {
    const content = (
        <Card>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">{title}</CardTitle>
                <Icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
                <div className="text-2xl font-bold">{value}</div>
                {description && <p className="text-xs text-muted-foreground">{description}</p>}
            </CardContent>
        </Card>
    );

    return to ? <Link to={to} className="hover:shadow-lg transition-shadow duration-300">{content}</Link> : content;
};


export default function Dashboard() {
    const [userStats, setUserStats] = useState<UserStats | null>(null);
    const [appStats, setAppStats] = useState<AppStats>({ total_applications: 0, pending_applications: 0, approved_applications: 0 }); // Placeholder
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchStats = async () => {
            try {
                setLoading(true);
                // The API Gateway forwards /api/users/stats to the user-management-service
                const response = await apiClient.get('/users/stats');
                setUserStats(response.data);

                // TODO: Fetch application stats when the API is available
                // For now, using mock data
                setAppStats({ total_applications: 152, pending_applications: 25, approved_applications: 127 });

            } catch (err) {
                setError('Failed to fetch dashboard data.');
                console.error(err);
            } finally {
                setLoading(false);
            }
        };

        fetchStats();
    }, []);

    if (loading) {
        return <div className="flex justify-center items-center h-full"><Hourglass className="h-8 w-8 animate-spin" /></div>;
    }

    if (error) {
        return <div className="text-red-500 text-center">{error}</div>;
    }

  return (
    <div className="flex-1 space-y-4 p-8 pt-6">
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">
            Welcome back! Here's a summary of the system status.
        </p>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            <StatCard title="Total Users" value={userStats?.total_users ?? 'N/A'} icon={Users} to="/students" description="All registered users" />
            <StatCard title="Total Students" value={userStats?.student_users ?? 'N/A'} icon={School} to="/students" description="All student accounts" />
            <StatCard title="Total Teachers" value={userStats?.teacher_users ?? 'N/A'} icon={Briefcase} to="/teachers" description="All teacher accounts" />
            <StatCard title="Admin Users" value={userStats?.admin_users ?? 'N/A'} icon={UserCheck} description="System administrators" />
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
             <StatCard title="Total Applications" value={appStats.total_applications} icon={FileText} to="/applications" />
             <StatCard title="Pending Applications" value={appStats.pending_applications} icon={GitPullRequest} to="/applications" />
             <StatCard title="Approved Applications" value={appStats.approved_applications} icon={UserCheck} to="/applications" />
        </div>
    </div>
  );
} 