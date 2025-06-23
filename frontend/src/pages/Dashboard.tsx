import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { useAuth } from '../contexts/AuthContext';
import { api } from '../lib/api';
import {
    Users,
    FileText,
    UserCheck,
    Settings,
    TrendingUp,
    Clock,
    CheckCircle,
    XCircle
} from 'lucide-react';

interface DashboardStats {
    totalStudents: number;
    totalTeachers: number;
    totalApplications: number;
    totalAffairs: number;
    pendingApplications: number;
    approvedApplications: number;
    rejectedApplications: number;
}

const Dashboard: React.FC = () => {
    const { user } = useAuth();
    const [stats, setStats] = useState<DashboardStats | null>(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchStats = async () => {
            try {
                // 这里应该调用实际的API来获取统计数据
                // 目前使用模拟数据
                const mockStats: DashboardStats = {
                    totalStudents: 1250,
                    totalTeachers: 89,
                    totalApplications: 456,
                    totalAffairs: 23,
                    pendingApplications: 45,
                    approvedApplications: 389,
                    rejectedApplications: 22,
                };
                setStats(mockStats);
            } catch (error) {
                console.error('Failed to fetch dashboard stats:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchStats();
    }, []);

    if (loading) {
        return (
            <div className="flex items-center justify-center h-64">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* 欢迎信息 */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold text-foreground">
                        欢迎回来，{user?.username}！
                    </h1>
                    <p className="text-muted-foreground mt-1">
                        今天是 {new Date().toLocaleDateString('zh-CN')}
                    </p>
                </div>
                <div className="text-right">
                    <p className="text-sm text-muted-foreground">用户类型</p>
                    <p className="text-lg font-semibold capitalize">{user?.user_type}</p>
                </div>
            </div>

            {/* 统计卡片 */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">总学生数</CardTitle>
                        <Users className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats?.totalStudents}</div>
                        <p className="text-xs text-muted-foreground">
                            +12% 较上月
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">总教师数</CardTitle>
                        <UserCheck className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats?.totalTeachers}</div>
                        <p className="text-xs text-muted-foreground">
                            +3% 较上月
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">总申请数</CardTitle>
                        <FileText className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats?.totalApplications}</div>
                        <p className="text-xs text-muted-foreground">
                            +8% 较上月
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">总事项数</CardTitle>
                        <Settings className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{stats?.totalAffairs}</div>
                        <p className="text-xs text-muted-foreground">
                            +2% 较上月
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* 申请状态统计 */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center">
                            <Clock className="h-4 w-4 mr-2 text-yellow-500" />
                            待处理申请
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-3xl font-bold text-yellow-600">
                            {stats?.pendingApplications}
                        </div>
                        <p className="text-sm text-muted-foreground mt-2">
                            需要审核的申请数量
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center">
                            <CheckCircle className="h-4 w-4 mr-2 text-green-500" />
                            已通过申请
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-3xl font-bold text-green-600">
                            {stats?.approvedApplications}
                        </div>
                        <p className="text-sm text-muted-foreground mt-2">
                            审核通过的申请数量
                        </p>
                    </CardContent>
                </Card>

                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center">
                            <XCircle className="h-4 w-4 mr-2 text-red-500" />
                            已拒绝申请
                        </CardTitle>
                    </CardHeader>
                    <CardContent>
                        <div className="text-3xl font-bold text-red-600">
                            {stats?.rejectedApplications}
                        </div>
                        <p className="text-sm text-muted-foreground mt-2">
                            审核拒绝的申请数量
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* 快速操作 */}
            <Card>
                <CardHeader>
                    <CardTitle>快速操作</CardTitle>
                    <CardDescription>
                        常用功能的快捷入口
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
                            学生管理
                        </Button>
                        <Button variant="outline" className="h-20 flex-col">
                            <UserCheck className="h-6 w-6 mb-2" />
                            教师管理
                        </Button>
                        <Button variant="outline" className="h-20 flex-col">
                            <Settings className="h-6 w-6 mb-2" />
                            事项管理
                        </Button>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
};

export default Dashboard; 