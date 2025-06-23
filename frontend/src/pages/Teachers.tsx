import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { Badge } from '../components/ui/badge';
import { teacherAPI } from '../lib/api';
import { Plus, Search, Filter, Eye, Edit, Trash2, User, Building, Award, BookOpen } from 'lucide-react';
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
    createdAt: string;
    updatedAt: string;
    user: {
        username: string;
        userType: string;
        registerTime: string;
    };
}

const Teachers: React.FC = () => {
    const [teachers, setTeachers] = useState<Teacher[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [departmentFilter, setDepartmentFilter] = useState('all');
    const [titleFilter, setTitleFilter] = useState('all');
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showEditModal, setShowEditModal] = useState(false);
    const [selectedTeacher, setSelectedTeacher] = useState<Teacher | null>(null);
    const [departments, setDepartments] = useState<string[]>([]);
    const [titles, setTitles] = useState<string[]>([]);

    useEffect(() => {
        fetchTeachers();
    }, []);

    const fetchTeachers = async () => {
        try {
            const response = await teacherAPI.getTeachers();
            setTeachers(response.data);
            
            // 提取唯一的院系和职称
            const uniqueDepartments = [...new Set(response.data.map(t => t.department).filter(Boolean))];
            const uniqueTitles = [...new Set(response.data.map(t => t.title).filter(Boolean))];
            setDepartments(uniqueDepartments);
            setTitles(uniqueTitles);
        } catch (error) {
            toast.error('获取教师列表失败');
            console.error('Failed to fetch teachers:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleCreateTeacher = async (teacherData: any) => {
        try {
            await teacherAPI.createTeacher(teacherData);
            toast.success('教师创建成功');
            fetchTeachers();
            setShowCreateModal(false);
        } catch (error) {
            toast.error('创建教师失败');
            console.error('Failed to create teacher:', error);
        }
    };

    const handleUpdateTeacher = async (username: string, teacherData: any) => {
        try {
            await teacherAPI.updateTeacher(username, teacherData);
            toast.success('教师信息更新成功');
            fetchTeachers();
            setShowEditModal(false);
            setSelectedTeacher(null);
        } catch (error) {
            toast.error('更新教师信息失败');
            console.error('Failed to update teacher:', error);
        }
    };

    const handleDeleteTeacher = async (username: string) => {
        if (!confirm('确定要删除这个教师吗？')) return;
        
        try {
            await teacherAPI.deleteTeacher(username);
            toast.success('教师删除成功');
            fetchTeachers();
        } catch (error) {
            toast.error('删除教师失败');
            console.error('Failed to delete teacher:', error);
        }
    };

    const filteredTeachers = teachers.filter(teacher => {
        const matchesSearch = teacher.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            teacher.username.toLowerCase().includes(searchTerm.toLowerCase()) ||
            teacher.specialty.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || teacher.status === statusFilter;
        const matchesDepartment = departmentFilter === 'all' || teacher.department === departmentFilter;
        const matchesTitle = titleFilter === 'all' || teacher.title === titleFilter;
        return matchesSearch && matchesStatus && matchesDepartment && matchesTitle;
    });

    const getStatusBadge = (status: string) => {
        const statusConfig = {
            active: { color: 'bg-green-100 text-green-800', text: '在职' },
            inactive: { color: 'bg-yellow-100 text-yellow-800', text: '离职' },
            retired: { color: 'bg-blue-100 text-blue-800', text: '退休' }
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
                        管理教师信息和档案
                    </p>
                </div>
                <Button onClick={() => setShowCreateModal(true)}>
                    <Plus className="h-4 w-4 mr-2" />
                    添加教师
                </Button>
            </div>

            {/* 搜索和筛选 */}
            <Card>
                <CardContent className="pt-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
                        <div>
                            <Label htmlFor="search">搜索</Label>
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
                            <Label htmlFor="status">状态筛选</Label>
                            <Select value={statusFilter} onValueChange={setStatusFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择状态" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">全部状态</SelectItem>
                                    <SelectItem value="active">在职</SelectItem>
                                    <SelectItem value="inactive">离职</SelectItem>
                                    <SelectItem value="retired">退休</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div>
                            <Label htmlFor="department">院系筛选</Label>
                            <Select value={departmentFilter} onValueChange={setDepartmentFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择院系" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">全部院系</SelectItem>
                                    {departments.map(department => (
                                        <SelectItem key={department} value={department}>
                                            {department}
                                        </SelectItem>
                                    ))}
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
                                    {titles.map(title => (
                                        <SelectItem key={title} value={title}>
                                            {title}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="flex items-end">
                            <Button 
                                variant="outline" 
                                onClick={() => {
                                    setSearchTerm('');
                                    setStatusFilter('all');
                                    setDepartmentFilter('all');
                                    setTitleFilter('all');
                                }}
                            >
                                <Filter className="h-4 w-4 mr-2" />
                                重置筛选
                            </Button>
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
                                        <User className="h-5 w-5 text-primary" />
                                    </div>
                                    <div>
                                        <CardTitle className="text-lg">{teacher.name}</CardTitle>
                                        <CardDescription>
                                            用户名: {teacher.username} | 职称: {teacher.title}
                                        </CardDescription>
                                    </div>
                                </div>
                                <div className="flex items-center space-x-2">
                                    {getStatusBadge(teacher.status)}
                                </div>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
                                <div className="flex items-center space-x-2">
                                    <Building className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm">
                                        <span className="font-medium">院系:</span> {teacher.department || '未设置'}
                                    </span>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <Award className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm">
                                        <span className="font-medium">职称:</span> {teacher.title || '未设置'}
                                    </span>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <BookOpen className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm">
                                        <span className="font-medium">专业领域:</span> {teacher.specialty || '未设置'}
                                    </span>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <User className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm">
                                        <span className="font-medium">状态:</span> {teacher.status}
                                    </span>
                                </div>
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                                <div>
                                    <span className="text-sm font-medium">联系方式:</span>
                                    <p className="text-sm text-muted-foreground">{teacher.contact || '未设置'}</p>
                                </div>
                                <div>
                                    <span className="text-sm font-medium">邮箱:</span>
                                    <p className="text-sm text-muted-foreground">{teacher.email || '未设置'}</p>
                                </div>
                            </div>

                            <div className="flex items-center justify-between">
                                <div className="text-sm text-muted-foreground">
                                    注册时间: {new Date(teacher.user.registerTime).toLocaleDateString('zh-CN')}
                                </div>

                                <div className="flex items-center space-x-2">
                                    <Button 
                                        variant="outline" 
                                        size="sm"
                                        onClick={() => {
                                            setSelectedTeacher(teacher);
                                            setShowEditModal(true);
                                        }}
                                    >
                                        <Eye className="h-4 w-4 mr-1" />
                                        查看详情
                                    </Button>
                                    <Button 
                                        variant="outline" 
                                        size="sm"
                                        onClick={() => {
                                            setSelectedTeacher(teacher);
                                            setShowEditModal(true);
                                        }}
                                    >
                                        <Edit className="h-4 w-4 mr-1" />
                                        编辑
                                    </Button>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() => handleDeleteTeacher(teacher.username)}
                                        className="text-red-600 hover:text-red-700"
                                    >
                                        <Trash2 className="h-4 w-4 mr-1" />
                                        删除
                                    </Button>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                ))}
                            </div>

            {/* 创建教师模态框 */}
            {showCreateModal && (
                <CreateTeacherModal 
                    onClose={() => setShowCreateModal(false)}
                    onSubmit={handleCreateTeacher}
                />
            )}

            {/* 编辑教师模态框 */}
            {showEditModal && selectedTeacher && (
                <EditTeacherModal 
                    teacher={selectedTeacher}
                    onClose={() => {
                        setShowEditModal(false);
                        setSelectedTeacher(null);
                    }}
                    onSubmit={handleUpdateTeacher}
                />
            )}
        </div>
    );
};

// 创建教师模态框组件
interface CreateTeacherModalProps {
    onClose: () => void;
    onSubmit: (data: any) => void;
}

const CreateTeacherModal: React.FC<CreateTeacherModalProps> = ({ onClose, onSubmit }) => {
    const [formData, setFormData] = useState({
        username: '',
        name: '',
        contact: '',
        email: '',
        department: '',
        title: '',
        specialty: ''
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto">
                <h2 className="text-xl font-bold mb-4">添加教师</h2>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label htmlFor="username">用户名 *</Label>
                            <Input
                                id="username"
                                value={formData.username}
                                onChange={(e) => setFormData({...formData, username: e.target.value})}
                                required
                            />
                        </div>
                        <div>
                            <Label htmlFor="name">姓名 *</Label>
                            <Input
                                id="name"
                                value={formData.name}
                                onChange={(e) => setFormData({...formData, name: e.target.value})}
                                required
                            />
                        </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label htmlFor="department">院系</Label>
                            <Input
                                id="department"
                                value={formData.department}
                                onChange={(e) => setFormData({...formData, department: e.target.value})}
                            />
                        </div>
                        <div>
                            <Label htmlFor="title">职称</Label>
                            <Input
                                id="title"
                                value={formData.title}
                                onChange={(e) => setFormData({...formData, title: e.target.value})}
                            />
                        </div>
                    </div>
                    <div>
                        <Label htmlFor="specialty">专业领域</Label>
                        <Input
                            id="specialty"
                            value={formData.specialty}
                            onChange={(e) => setFormData({...formData, specialty: e.target.value})}
                        />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label htmlFor="contact">联系方式</Label>
                            <Input
                                id="contact"
                                value={formData.contact}
                                onChange={(e) => setFormData({...formData, contact: e.target.value})}
                            />
                        </div>
                        <div>
                            <Label htmlFor="email">邮箱</Label>
                            <Input
                                id="email"
                                type="email"
                                value={formData.email}
                                onChange={(e) => setFormData({...formData, email: e.target.value})}
                            />
                        </div>
                    </div>
                    <div className="flex gap-4">
                        <Button type="submit">创建教师</Button>
                        <Button type="button" variant="outline" onClick={onClose}>取消</Button>
                    </div>
                </form>
            </div>
        </div>
    );
};

// 编辑教师模态框组件
interface EditTeacherModalProps {
    teacher: Teacher;
    onClose: () => void;
    onSubmit: (username: string, data: any) => void;
}

const EditTeacherModal: React.FC<EditTeacherModalProps> = ({ teacher, onClose, onSubmit }) => {
    const [formData, setFormData] = useState({
        name: teacher.name,
        contact: teacher.contact,
        email: teacher.email,
        department: teacher.department,
        title: teacher.title,
        specialty: teacher.specialty,
        status: teacher.status
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(teacher.username, formData);
    };

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto">
                <h2 className="text-xl font-bold mb-4">编辑教师信息</h2>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <Label>用户名</Label>
                        <Input value={teacher.username} disabled />
                    </div>
                    <div>
                        <Label htmlFor="name">姓名 *</Label>
                        <Input
                            id="name"
                            value={formData.name}
                            onChange={(e) => setFormData({...formData, name: e.target.value})}
                            required
                        />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label htmlFor="department">院系</Label>
                            <Input
                                id="department"
                                value={formData.department}
                                onChange={(e) => setFormData({...formData, department: e.target.value})}
                            />
                        </div>
                        <div>
                            <Label htmlFor="title">职称</Label>
                            <Input
                                id="title"
                                value={formData.title}
                                onChange={(e) => setFormData({...formData, title: e.target.value})}
                            />
                        </div>
                    </div>
                    <div>
                        <Label htmlFor="specialty">专业领域</Label>
                        <Input
                            id="specialty"
                            value={formData.specialty}
                            onChange={(e) => setFormData({...formData, specialty: e.target.value})}
                        />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label htmlFor="contact">联系方式</Label>
                            <Input
                                id="contact"
                                value={formData.contact}
                                onChange={(e) => setFormData({...formData, contact: e.target.value})}
                            />
                        </div>
                        <div>
                            <Label htmlFor="email">邮箱</Label>
                            <Input
                                id="email"
                                type="email"
                                value={formData.email}
                                onChange={(e) => setFormData({...formData, email: e.target.value})}
                            />
                        </div>
                    </div>
                    <div>
                        <Label htmlFor="status">状态</Label>
                        <Select value={formData.status} onValueChange={(value) => setFormData({...formData, status: value})}>
                            <SelectTrigger>
                                <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="active">在职</SelectItem>
                                <SelectItem value="inactive">离职</SelectItem>
                                <SelectItem value="retired">退休</SelectItem>
                            </SelectContent>
                        </Select>
                    </div>
                    <div className="flex gap-4">
                        <Button type="submit">保存更改</Button>
                        <Button type="button" variant="outline" onClick={onClose}>取消</Button>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default Teachers; 