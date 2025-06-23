import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Badge } from '../components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { api } from '../lib/api';
import { Plus, Search, Filter, Eye, Edit, Trash2, UserCheck } from 'lucide-react';
import toast from 'react-hot-toast';

interface Teacher {
    username: string;
    name: string;
    contact: string;
    email: string;
    department: string;
    title: string;
    specialty: string;
    status: string;
    created_at: string;
    updated_at: string;
}

const Teachers: React.FC = () => {
    const [teachers, setTeachers] = useState<Teacher[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [departmentFilter, setDepartmentFilter] = useState('all');
    const [titleFilter, setTitleFilter] = useState('all');

    useEffect(() => {
        fetchTeachers();
    }, []);

    const fetchTeachers = async () => {
        try {
            const response = await api.get('/teachers');
            setTeachers(response.data);
        } catch (error) {
            toast.error('获取教师列表失败');
            console.error('Failed to fetch teachers:', error);
        } finally {
            setLoading(false);
        }
    };

    const filteredTeachers = teachers.filter(teacher => {
        const matchesSearch = teacher.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            teacher.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
            teacher.specialty.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesDepartment = departmentFilter === 'all' || teacher.department === departmentFilter;
        const matchesTitle = titleFilter === 'all' || teacher.title === titleFilter;
        return matchesSearch && matchesDepartment && matchesTitle;
    });

    const getStatusBadge = (status: string) => {
        const statusConfig = {
            active: { color: 'bg-green-100 text-green-800', text: '在职' },
            inactive: { color: 'bg-gray-100 text-gray-800', text: '离职' },
            retired: { color: 'bg-orange-100 text-orange-800', text: '退休' }
        };

        const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.active;
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
                    <h1 className="text-3xl font-bold text-foreground">教师管理</h1>
                    <p className="text-muted-foreground mt-1">
                        管理教师信息和审核权限
                    </p>
                </div>
                <Button>
                    <Plus className="h-4 w-4 mr-2" />
                    添加教师
                </Button>
            </div>

            {/* 搜索和筛选 */}
            <Card>
                <CardContent className="pt-6">
                    <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                        <div className="md:col-span-2">
                            <Label htmlFor="search">搜索教师</Label>
                            <div className="relative">
                                <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                                <Input
                                    id="search"
                                    placeholder="搜索姓名、用户名或专业领域..."
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="pl-10"
                                />
                            </div>
                        </div>
                        <div>
                            <Label htmlFor="department">院系筛选</Label>
                            <Select value={departmentFilter} onValueChange={setDepartmentFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择院系" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">全部院系</SelectItem>
                                    <SelectItem value="计算机科学学院">计算机科学学院</SelectItem>
                                    <SelectItem value="数学学院">数学学院</SelectItem>
                                    <SelectItem value="物理学院">物理学院</SelectItem>
                                    <SelectItem value="化学学院">化学学院</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div>
                            <Label htmlFor="title">职称筛选</Label>
                            <Select value={titleFilter} onValueChange={setTitleFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择职称" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">全部职称</SelectItem>
                                    <SelectItem value="教授">教授</SelectItem>
                                    <SelectItem value="副教授">副教授</SelectItem>
                                    <SelectItem value="讲师">讲师</SelectItem>
                                    <SelectItem value="助教">助教</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* 教师列表 */}
            <div className="grid gap-4">
                {filteredTeachers.map((teacher) => (
                    <Card key={teacher.username}>
                        <CardHeader>
                            <div className="flex items-center justify-between">
                                <div className="flex items-center space-x-3">
                                    <div className="w-10 h-10 bg-primary/10 rounded-full flex items-center justify-center">
                                        <UserCheck className="h-5 w-5 text-primary" />
                                    </div>
                                    <div>
                                        <CardTitle className="text-lg">{teacher.name}</CardTitle>
                                        <CardDescription>
                                            用户名: {teacher.username} | 院系: {teacher.department}
                                        </CardDescription>
                                    </div>
                                </div>
                                <div className="flex items-center space-x-2">
                                    {getStatusBadge(teacher.status)}
                                    <Badge variant="outline">{teacher.title}</Badge>
                                </div>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                                <div>
                                    <span className="text-muted-foreground">专业领域:</span>
                                    <p className="font-medium">{teacher.specialty}</p>
                                </div>
                                <div>
                                    <span className="text-muted-foreground">联系方式:</span>
                                    <p className="font-medium">{teacher.contact}</p>
                                </div>
                                <div>
                                    <span className="text-muted-foreground">邮箱:</span>
                                    <p className="font-medium">{teacher.email}</p>
                                </div>
                                <div>
                                    <span className="text-muted-foreground">注册时间:</span>
                                    <p className="font-medium">
                                        {new Date(teacher.created_at).toLocaleDateString('zh-CN')}
                                    </p>
                                </div>
                            </div>

                            <div className="flex items-center justify-between mt-4 pt-4 border-t">
                                <div className="text-sm text-muted-foreground">
                                    最后更新: {new Date(teacher.updated_at).toLocaleDateString('zh-CN')}
                                </div>

                                <div className="flex items-center space-x-2">
                                    <Button variant="outline" size="sm">
                                        <Eye className="h-4 w-4 mr-1" />
                                        查看详情
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

                {filteredTeachers.length === 0 && (
                    <Card>
                        <CardContent className="pt-6">
                            <div className="text-center py-8">
                                <p className="text-muted-foreground">暂无教师记录</p>
                            </div>
                        </CardContent>
                    </Card>
                )}
            </div>
        </div>
    );
};

export default Teachers; 