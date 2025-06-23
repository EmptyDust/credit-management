import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Badge } from '../components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { api } from '../lib/api';
import { Plus, Search, Filter, Eye, Edit, Trash2, Users, Calendar, Award } from 'lucide-react';
import toast from 'react-hot-toast';

interface Affair {
    id: number;
    name: string;
    description: string;
    affair_type: string;
    credit_value: number;
    max_participants: number;
    current_participants: number;
    start_date: string;
    end_date: string;
    status: string;
    created_at: string;
    updated_at: string;
}

const Affairs: React.FC = () => {
    const [affairs, setAffairs] = useState<Affair[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [typeFilter, setTypeFilter] = useState('all');
    const [statusFilter, setStatusFilter] = useState('all');

    useEffect(() => {
        fetchAffairs();
    }, []);

    const fetchAffairs = async () => {
        try {
            const response = await api.get('/affairs');
            setAffairs(response.data);
        } catch (error) {
            toast.error('获取事项列表失败');
            console.error('Failed to fetch affairs:', error);
        } finally {
            setLoading(false);
        }
    };

    const filteredAffairs = affairs.filter(affair => {
        const matchesSearch = affair.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            affair.description.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesType = typeFilter === 'all' || affair.affair_type === typeFilter;
        const matchesStatus = statusFilter === 'all' || affair.status === statusFilter;
        return matchesSearch && matchesType && matchesStatus;
    });

    const getStatusBadge = (status: string) => {
        const statusConfig = {
            active: { color: 'bg-green-100 text-green-800', text: '进行中' },
            upcoming: { color: 'bg-blue-100 text-blue-800', text: '即将开始' },
            completed: { color: 'bg-gray-100 text-gray-800', text: '已结束' },
            cancelled: { color: 'bg-red-100 text-red-800', text: '已取消' }
        };

        const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.active;
        return <Badge className={config.color}>{config.text}</Badge>;
    };

    const getTypeBadge = (type: string) => {
        const typeConfig = {
            competition: { color: 'bg-purple-100 text-purple-800', text: '竞赛' },
            project: { color: 'bg-orange-100 text-orange-800', text: '项目' },
            workshop: { color: 'bg-cyan-100 text-cyan-800', text: '工作坊' },
            seminar: { color: 'bg-pink-100 text-pink-800', text: '研讨会' }
        };

        const config = typeConfig[type as keyof typeof typeConfig] || typeConfig.competition;
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
                    <h1 className="text-3xl font-bold text-foreground">事项管理</h1>
                    <p className="text-muted-foreground mt-1">
                        管理创新创业活动和事项
                    </p>
                </div>
                <Button>
                    <Plus className="h-4 w-4 mr-2" />
                    创建事项
                </Button>
            </div>

            {/* 搜索和筛选 */}
            <Card>
                <CardContent className="pt-6">
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                        <div className="md:col-span-2">
                            <Label htmlFor="search">搜索事项</Label>
                            <div className="relative">
                                <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                                <Input
                                    id="search"
                                    placeholder="搜索事项名称或描述..."
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="pl-10"
                                />
                            </div>
                        </div>
                        <div>
                            <Label htmlFor="type">类型筛选</Label>
                            <Select value={typeFilter} onValueChange={setTypeFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择类型" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">全部类型</SelectItem>
                                    <SelectItem value="competition">竞赛</SelectItem>
                                    <SelectItem value="project">项目</SelectItem>
                                    <SelectItem value="workshop">工作坊</SelectItem>
                                    <SelectItem value="seminar">研讨会</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div>
                            <Label htmlFor="status">状态筛选</Label>
                            <Select value={statusFilter} onValueChange={setStatusFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择状态" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">全部状态</SelectItem>
                                    <SelectItem value="active">进行中</SelectItem>
                                    <SelectItem value="upcoming">即将开始</SelectItem>
                                    <SelectItem value="completed">已结束</SelectItem>
                                    <SelectItem value="cancelled">已取消</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* 事项列表 */}
            <div className="grid gap-4">
                {filteredAffairs.map((affair) => (
                    <Card key={affair.id}>
                        <CardHeader>
                            <div className="flex items-center justify-between">
                                <div>
                                    <CardTitle className="text-lg">{affair.name}</CardTitle>
                                    <CardDescription className="mt-1">
                                        {affair.description}
                                    </CardDescription>
                                </div>
                                <div className="flex items-center space-x-2">
                                    {getTypeBadge(affair.affair_type)}
                                    {getStatusBadge(affair.status)}
                                </div>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm mb-4">
                                <div className="flex items-center space-x-2">
                                    <Award className="h-4 w-4 text-muted-foreground" />
                                    <div>
                                        <span className="text-muted-foreground">学分:</span>
                                        <p className="font-medium">{affair.credit_value}</p>
                                    </div>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <Users className="h-4 w-4 text-muted-foreground" />
                                    <div>
                                        <span className="text-muted-foreground">参与人数:</span>
                                        <p className="font-medium">{affair.current_participants}/{affair.max_participants}</p>
                                    </div>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <Calendar className="h-4 w-4 text-muted-foreground" />
                                    <div>
                                        <span className="text-muted-foreground">开始时间:</span>
                                        <p className="font-medium">
                                            {new Date(affair.start_date).toLocaleDateString('zh-CN')}
                                        </p>
                                    </div>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <Calendar className="h-4 w-4 text-muted-foreground" />
                                    <div>
                                        <span className="text-muted-foreground">结束时间:</span>
                                        <p className="font-medium">
                                            {new Date(affair.end_date).toLocaleDateString('zh-CN')}
                                        </p>
                                    </div>
                                </div>
                            </div>

                            <div className="flex items-center justify-between pt-4 border-t">
                                <div className="text-sm text-muted-foreground">
                                    创建时间: {new Date(affair.created_at).toLocaleDateString('zh-CN')}
                                </div>

                                <div className="flex items-center space-x-2">
                                    <Button variant="outline" size="sm">
                                        <Eye className="h-4 w-4 mr-1" />
                                        查看详情
                                    </Button>
                                    <Button variant="outline" size="sm">
                                        <Users className="h-4 w-4 mr-1" />
                                        管理参与者
                                    </Button>
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

                {filteredAffairs.length === 0 && (
                    <Card>
                        <CardContent className="pt-6">
                            <div className="text-center py-8">
                                <p className="text-muted-foreground">暂无事项记录</p>
                            </div>
                        </CardContent>
                    </Card>
                )}
            </div>
        </div>
    );
};

export default Affairs; 