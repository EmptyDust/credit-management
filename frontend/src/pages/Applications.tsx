import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { Badge } from '../components/ui/badge';
import { api } from '../lib/api';
import { Plus, Search, Filter, Eye, Edit, Trash2, CheckCircle, XCircle } from 'lucide-react';
import toast from 'react-hot-toast';

interface Application {
    id: number;
    student_id: string;
    affair_id: number;
    application_type: string;
    title: string;
    description: string;
    status: string;
    recognized_credits: number;
    created_at: string;
    updated_at: string;
    student?: {
        name: string;
        student_id: string;
    };
    affair?: {
        name: string;
    };
}

const Applications: React.FC = () => {
    const [applications, setApplications] = useState<Application[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');

    useEffect(() => {
        fetchApplications();
    }, []);

    const fetchApplications = async () => {
        try {
            const response = await api.get('/applications');
            setApplications(response.data);
        } catch (error) {
            toast.error('获取申请列表失败');
            console.error('Failed to fetch applications:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleStatusChange = async (applicationId: number, newStatus: string) => {
        try {
            await api.post(`/applications/${applicationId}/review`, {
                status: newStatus,
                review_comment: `状态更新为: ${newStatus}`
            });
            toast.success('申请状态更新成功');
            fetchApplications();
        } catch (error) {
            toast.error('更新申请状态失败');
            console.error('Failed to update application status:', error);
        }
    };

    const filteredApplications = applications.filter(app => {
        const matchesSearch = app.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
            app.student?.name.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || app.status === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const getStatusBadge = (status: string) => {
        const statusConfig = {
            pending: { color: 'bg-yellow-100 text-yellow-800', text: '待审核' },
            approved: { color: 'bg-green-100 text-green-800', text: '已通过' },
            rejected: { color: 'bg-red-100 text-red-800', text: '已拒绝' },
            processing: { color: 'bg-blue-100 text-blue-800', text: '处理中' }
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
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-3xl font-bold text-foreground">申请管理</h1>
                    <p className="text-muted-foreground mt-1">
                        管理学生的创新创业学分申请
                    </p>
                </div>
                <Button>
                    <Plus className="h-4 w-4 mr-2" />
                    新建申请
                </Button>
            </div>

            {/* 搜索和筛选 */}
            <Card>
                <CardContent className="pt-6">
                    <div className="flex flex-col md:flex-row gap-4">
                        <div className="flex-1">
                            <Label htmlFor="search">搜索</Label>
                            <div className="relative">
                                <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                                <Input
                                    id="search"
                                    placeholder="搜索申请标题或学生姓名..."
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="pl-10"
                                />
                            </div>
                        </div>
                        <div className="w-full md:w-48">
                            <Label htmlFor="status">状态筛选</Label>
                            <Select value={statusFilter} onValueChange={setStatusFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择状态" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">全部状态</SelectItem>
                                    <SelectItem value="pending">待审核</SelectItem>
                                    <SelectItem value="processing">处理中</SelectItem>
                                    <SelectItem value="approved">已通过</SelectItem>
                                    <SelectItem value="rejected">已拒绝</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* 申请列表 */}
            <div className="grid gap-4">
                {filteredApplications.map((application) => (
                    <Card key={application.id}>
                        <CardHeader>
                            <div className="flex items-center justify-between">
                                <div>
                                    <CardTitle className="text-lg">{application.title}</CardTitle>
                                    <CardDescription>
                                        申请类型: {application.application_type} |
                                        学生: {application.student?.name} ({application.student?.student_id})
                                    </CardDescription>
                                </div>
                                <div className="flex items-center space-x-2">
                                    {getStatusBadge(application.status)}
                                    <span className="text-sm text-muted-foreground">
                                        学分: {application.recognized_credits}
                                    </span>
                                </div>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <p className="text-sm text-muted-foreground mb-4">
                                {application.description}
                            </p>

                            <div className="flex items-center justify-between">
                                <div className="text-sm text-muted-foreground">
                                    创建时间: {new Date(application.created_at).toLocaleDateString('zh-CN')}
                                </div>

                                <div className="flex items-center space-x-2">
                                    <Button variant="outline" size="sm">
                                        <Eye className="h-4 w-4 mr-1" />
                                        查看详情
                                    </Button>

                                    {application.status === 'pending' && (
                                        <>
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => handleStatusChange(application.id, 'approved')}
                                                className="text-green-600 hover:text-green-700"
                                            >
                                                <CheckCircle className="h-4 w-4 mr-1" />
                                                通过
                                            </Button>
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => handleStatusChange(application.id, 'rejected')}
                                                className="text-red-600 hover:text-red-700"
                                            >
                                                <XCircle className="h-4 w-4 mr-1" />
                                                拒绝
                                            </Button>
                                        </>
                                    )}

                                    <Button variant="outline" size="sm">
                                        <Edit className="h-4 w-4 mr-1" />
                                        编辑
                                    </Button>

                                    <Button variant="outline" size="sm" className="text-red-600 hover:text-red-700">
                                        <Trash2 className="h-4 w-4 mr-1" />
                                        删除
                                    </Button>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                ))}

                {filteredApplications.length === 0 && (
                    <Card>
                        <CardContent className="pt-6">
                            <div className="text-center py-8">
                                <p className="text-muted-foreground">暂无申请记录</p>
                            </div>
                        </CardContent>
                    </Card>
                )}
            </div>
        </div>
    );
};

export default Applications; 