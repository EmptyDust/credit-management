import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Badge } from '../components/ui/badge';
import { userAPI, studentAPI, teacherAPI, applicationAPI, affairAPI } from '../lib/api';
import { 
    Users, 
    GraduationCap, 
    Award, 
    FileText, 
    Calendar, 
    TrendingUp, 
    CheckCircle, 
    Clock, 
    XCircle,
    Activity,
    BookOpen,
    Target
} from 'lucide-react';
import toast from 'react-hot-toast';

interface DashboardStats {
    totalUsers: number;
    totalStudents: number;
    totalTeachers: number;
    totalApplications: number;
    totalAffairs: number;
    pendingApplications: number;
    approvedApplications: number;
    rejectedApplications: number;
    totalCredits: number;
    approvedCredits: number;
    averageCredits: number;
    applicationsToday: number;
    applicationsWeek: number;
    applicationsMonth: number;
}

interface RecentApplication {
    id: number;
    title: string;
    status: string;
    credits: number;
    createdAt: string;
    userID: number;
}

interface RecentAffair {
    id: number;
    name: string;
    category: string;
    maxCredits: number;
    status: string;
    createdAt: string;
}

const Dashboard: React.FC = () => {
    const [stats, setStats] = useState<DashboardStats>({
        totalUsers: 0,
        totalStudents: 0,
        totalTeachers: 0,
        totalApplications: 0,
        totalAffairs: 0,
        pendingApplications: 0,
        approvedApplications: 0,
        rejectedApplications: 0,
        totalCredits: 0,
        approvedCredits: 0,
        averageCredits: 0,
        applicationsToday: 0,
        applicationsWeek: 0,
        applicationsMonth: 0
    });
    const [recentApplications, setRecentApplications] = useState<RecentApplication[]>([]);
    const [recentAffairs, setRecentAffairs] = useState<RecentAffair[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        fetchDashboardData();
    }, []);

    const fetchDashboardData = async () => {
        try {
            // 并行获取所有数据
            const [
                usersResponse,
                studentsResponse,
                teachersResponse,
                applicationsResponse,
                affairsResponse,
                applicationStatsResponse
            ] = await Promise.all([
                userAPI.getUsers(),
                studentAPI.getStudents(),
                teacherAPI.getTeachers(),
                applicationAPI.getApplications(),
                affairAPI.getAffairs(),
                applicationAPI.getApplicationStats()
            ]);

            // 计算统计数据
            const users = usersResponse.data;
            const students = studentsResponse.data;
            const teachers = teachersResponse.data;
            const applications = applicationsResponse.data;
            const affairs = affairsResponse.data;
            const appStats = applicationStatsResponse.data;

            // 获取最近的申请和事项
            const recentApps = applications
                .sort((a: any, b: any) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
                .slice(0, 5);
            
            const recentAffs = affairs
                .sort((a: any, b: any) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
                .slice(0, 5);

            setStats({
                totalUsers: users.length,
                totalStudents: students.length,
                totalTeachers: teachers.length,
                totalApplications: applications.length,
                totalAffairs: affairs.length,
                pendingApplications: applications.filter((app: any) => app.status === 'pending').length,
                approvedApplications: applications.filter((app: any) => app.status === 'approved').length,
                rejectedApplications: applications.filter((app: any) => app.status === 'rejected').length,
                totalCredits: appStats.totalCredits || 0,
                approvedCredits: appStats.approvedCredits || 0,
                averageCredits: appStats.averageCredits || 0,
                applicationsToday: appStats.applicationsToday || 0,
                applicationsWeek: appStats.applicationsWeek || 0,
                applicationsMonth: appStats.applicationsMonth || 0
            });

            setRecentApplications(recentApps);
            setRecentAffairs(recentAffs);
        } catch (error) {
            toast.error('获取仪表板数据失败');
            console.error('Failed to fetch dashboard data:', error);
        } finally {
            setLoading(false);
        }
    };

    const getStatusBadge = (status: string) => {
        const statusConfig = {
            pending: { color: 'bg-yellow-100 text-yellow-800', text: '待审核' },
            approved: { color: 'bg-green-100 text-green-800', text: '已通过' },
            rejected: { color: 'bg-red-100 text-red-800', text: '已拒绝' },
            active: { color: 'bg-green-100 text-green-800', text: '活跃' },
            inactive: { color: 'bg-gray-100 text-gray-800', text: '非活跃' }
        };

        const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.pending;
        return <Badge className={config.color}>{config.text}</Badge>;
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-64">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* 页面标题 */}
            <div>
                <h1 className="text-3xl font-bold text-foreground">仪表板</h1>
                <p className="text-muted-foreground mt-1">
                    系统概览和统计数据
                </p>
            </div>

            {/* 统计卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                {/* 用户统计 */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">总用户数</CardTitle>
                        <Users className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats.totalUsers}</div>
                        <p className="text-xs text-muted-foreground">
                            学生: {stats.totalStudents} | 教师: {stats.totalTeachers}
                        </p>
                    </CardContent>
                </Card>

                {/* 申请统计 */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">总申请数</CardTitle>
                        <FileText className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats.totalApplications}</div>
                        <p className="text-xs text-muted-foreground">
                            待审核: {stats.pendingApplications} | 已通过: {stats.approvedApplications}
                        </p>
                    </CardContent>
                </Card>

                {/* 事项统计 */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">总事项数</CardTitle>
                        <Activity className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats.totalAffairs}</div>
                        <p className="text-xs text-muted-foreground">
                            活跃事项数量
                        </p>
                    </CardContent>
                </Card>

                {/* 学分统计 */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">总学分</CardTitle>
                        <Target className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats.totalCredits}</div>
                        <p className="text-xs text-muted-foreground">
                            已批准: {stats.approvedCredits} | 平均: {stats.averageCredits.toFixed(1)}
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* 申请状态分布 */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center">
                            <Clock className="h-4 w-4 mr-2" />
                            待审核申请
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-3xl font-bold text-yellow-600">{stats.pendingApplications}</div>
                        <p className="text-sm text-muted-foreground">
                            需要审核的申请数量
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center">
                            <CheckCircle className="h-4 w-4 mr-2" />
                            已通过申请
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-3xl font-bold text-green-600">{stats.approvedApplications}</div>
                        <p className="text-sm text-muted-foreground">
                            审核通过的申请数量
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center">
                            <XCircle className="h-4 w-4 mr-2" />
                            已拒绝申请
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-3xl font-bold text-red-600">{stats.rejectedApplications}</div>
                        <p className="text-sm text-muted-foreground">
                            审核拒绝的申请数量
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* 时间趋势 */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center">
                            <Calendar className="h-4 w-4 mr-2" />
                            今日申请
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats.applicationsToday}</div>
                        <p className="text-sm text-muted-foreground">
                            今天提交的申请数量
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center">
                            <TrendingUp className="h-4 w-4 mr-2" />
                            本周申请
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats.applicationsWeek}</div>
                        <p className="text-sm text-muted-foreground">
                            本周提交的申请数量
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center">
                            <BookOpen className="h-4 w-4 mr-2" />
                            本月申请
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats.applicationsMonth}</div>
                        <p className="text-sm text-muted-foreground">
                            本月提交的申请数量
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* 最近活动 */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                {/* 最近申请 */}
                <Card>
                    <CardHeader>
                        <CardTitle>最近申请</CardTitle>
                        <CardDescription>
                            最近提交的申请列表
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            {recentApplications.map((application) => (
                                <div key={application.id} className="flex items-center justify-between p-3 border rounded-lg">
                                    <div>
                                        <div className="font-medium">{application.title}</div>
                                        <div className="text-sm text-muted-foreground">
                                            学分: {application.credits} | 用户ID: {application.userID}
                                        </div>
                                    </div>
                                    <div className="flex items-center space-x-2">
                                        {getStatusBadge(application.status)}
                                        <span className="text-xs text-muted-foreground">
                                            {new Date(application.createdAt).toLocaleDateString('zh-CN')}
                                        </span>
                                    </div>
                                </div>
                            ))}
                            {recentApplications.length === 0 && (
                                <div className="text-center py-8 text-muted-foreground">
                                    暂无最近申请
                                </div>
                            )}
                        </div>
                    </CardContent>
                </Card>

                {/* 最近事项 */}
                <Card>
                    <CardHeader>
                        <CardTitle>最近事项</CardTitle>
                        <CardDescription>
                            最近创建的事项列表
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="space-y-4">
                            {recentAffairs.map((affair) => (
                                <div key={affair.id} className="flex items-center justify-between p-3 border rounded-lg">
                                    <div>
                                        <div className="font-medium">{affair.name}</div>
                                        <div className="text-sm text-muted-foreground">
                                            类别: {affair.category} | 最大学分: {affair.maxCredits}
                                        </div>
                                    </div>
                                    <div className="flex items-center space-x-2">
                                        {getStatusBadge(affair.status)}
                                        <span className="text-xs text-muted-foreground">
                                            {new Date(affair.createdAt).toLocaleDateString('zh-CN')}
                                        </span>
                                    </div>
                                </div>
                            ))}
                            {recentAffairs.length === 0 && (
                                <div className="text-center py-8 text-muted-foreground">
                                    暂无最近事项
                                </div>
                            )}
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* 快速操作 */}
            <Card>
                <CardHeader>
                    <CardTitle>快速操作</CardTitle>
                    <CardDescription>
                        常用功能的快速入口
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                        <Button variant="outline" className="h-20 flex-col">
                            <FileText className="h-6 w-6 mb-2" />
                            查看申请
                        </Button>
                        <Button variant="outline" className="h-20 flex-col">
                            <Users className="h-6 w-6 mb-2" />
                            管理学生
                        </Button>
                        <Button variant="outline" className="h-20 flex-col">
                            <Award className="h-6 w-6 mb-2" />
                            管理教师
                        </Button>
                        <Button variant="outline" className="h-20 flex-col">
                            <Activity className="h-6 w-6 mb-2" />
                            管理事项
                        </Button>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
};

export default Dashboard; 