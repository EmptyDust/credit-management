import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { Badge } from '../components/ui/badge';
import { studentAPI } from '../lib/api';
import { Plus, Search, Filter, Eye, Edit, Trash2, User, GraduationCap, Building, BookOpen } from 'lucide-react';
import toast from 'react-hot-toast';

interface Student {
    username: string;
    studentID: string;
    name: string;
    college: string;
    major: string;
    class: string;
    contact: string;
    email: string;
    grade: string;
    status: string;
    createdAt: string;
    updatedAt: string;
    user: {
        username: string;
        userType: string;
        registerTime: string;
    };
}

const Students: React.FC = () => {
    const [students, setStudents] = useState<Student[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [collegeFilter, setCollegeFilter] = useState('all');
    const [majorFilter, setMajorFilter] = useState('all');
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showEditModal, setShowEditModal] = useState(false);
    const [selectedStudent, setSelectedStudent] = useState<Student | null>(null);
    const [colleges, setColleges] = useState<string[]>([]);
    const [majors, setMajors] = useState<string[]>([]);

    useEffect(() => {
        fetchStudents();
    }, []);

    const fetchStudents = async () => {
        try {
            const response = await studentAPI.getStudents();
            setStudents(response.data);
            
            // 提取唯一的学院和专业
            const uniqueColleges = [...new Set(response.data.map(s => s.college).filter(Boolean))];
            const uniqueMajors = [...new Set(response.data.map(s => s.major).filter(Boolean))];
            setColleges(uniqueColleges);
            setMajors(uniqueMajors);
        } catch (error) {
            toast.error('获取学生列表失败');
            console.error('Failed to fetch students:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleCreateStudent = async (studentData: any) => {
        try {
            await studentAPI.createStudent(studentData);
            toast.success('学生创建成功');
            fetchStudents();
            setShowCreateModal(false);
        } catch (error) {
            toast.error('创建学生失败');
            console.error('Failed to create student:', error);
        }
    };

    const handleUpdateStudent = async (username: string, studentData: any) => {
        try {
            await studentAPI.updateStudent(username, studentData);
            toast.success('学生信息更新成功');
            fetchStudents();
            setShowEditModal(false);
            setSelectedStudent(null);
        } catch (error) {
            toast.error('更新学生信息失败');
            console.error('Failed to update student:', error);
        }
    };

    const handleDeleteStudent = async (username: string) => {
        if (!confirm('确定要删除这个学生吗？')) return;
        
        try {
            await studentAPI.deleteStudent(username);
            toast.success('学生删除成功');
            fetchStudents();
        } catch (error) {
            toast.error('删除学生失败');
            console.error('Failed to delete student:', error);
        }
    };

    const filteredStudents = students.filter(student => {
        const matchesSearch = student.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            student.studentID.toLowerCase().includes(searchTerm.toLowerCase()) ||
            student.username.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || student.status === statusFilter;
        const matchesCollege = collegeFilter === 'all' || student.college === collegeFilter;
        const matchesMajor = majorFilter === 'all' || student.major === majorFilter;
        return matchesSearch && matchesStatus && matchesCollege && matchesMajor;
    });

    const getStatusBadge = (status: string) => {
        const statusConfig = {
            active: { color: 'bg-green-100 text-green-800', text: '在读' },
            inactive: { color: 'bg-yellow-100 text-yellow-800', text: '休学' },
            graduated: { color: 'bg-blue-100 text-blue-800', text: '已毕业' }
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
                    <h1 className="text-3xl font-bold text-foreground">学生管理</h1>
                    <p className="text-muted-foreground mt-1">
                        管理学生信息和档案
                    </p>
                </div>
                <Button onClick={() => setShowCreateModal(true)}>
                    <Plus className="h-4 w-4 mr-2" />
                    添加学生
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
                                    placeholder="搜索姓名、学号或用户名..."
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
                                    <SelectItem value="active">在读</SelectItem>
                                    <SelectItem value="inactive">休学</SelectItem>
                                    <SelectItem value="graduated">已毕业</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div>
                            <Label htmlFor="college">学院筛选</Label>
                            <Select value={collegeFilter} onValueChange={setCollegeFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择学院" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">全部学院</SelectItem>
                                    {colleges.map(college => (
                                        <SelectItem key={college} value={college}>
                                            {college}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                        </div>
                        <div>
                            <Label htmlFor="major">专业筛选</Label>
                            <Select value={majorFilter} onValueChange={setMajorFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择专业" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">全部专业</SelectItem>
                                    {majors.map(major => (
                                        <SelectItem key={major} value={major}>
                                            {major}
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
                                    setCollegeFilter('all');
                                    setMajorFilter('all');
                                }}
                            >
                                <Filter className="h-4 w-4 mr-2" />
                                重置筛选
                            </Button>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* 学生列表 */}
            <div className="grid gap-4">
                {filteredStudents.map((student) => (
                    <Card key={student.username}>
                        <CardHeader>
                            <div className="flex items-center justify-between">
                                <div className="flex items-center space-x-3">
                                    <div className="w-10 h-10 bg-primary/10 rounded-full flex items-center justify-center">
                                        <User className="h-5 w-5 text-primary" />
                                    </div>
                                    <div>
                                        <CardTitle className="text-lg">{student.name}</CardTitle>
                                        <CardDescription>
                                            学号: {student.studentID} | 用户名: {student.username}
                                        </CardDescription>
                                    </div>
                                </div>
                                <div className="flex items-center space-x-2">
                                    {getStatusBadge(student.status)}
                                </div>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-4">
                                <div className="flex items-center space-x-2">
                                    <Building className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm">
                                        <span className="font-medium">学院:</span> {student.college || '未设置'}
                                    </span>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <BookOpen className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm">
                                        <span className="font-medium">专业:</span> {student.major || '未设置'}
                                    </span>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <GraduationCap className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm">
                                        <span className="font-medium">班级:</span> {student.class || '未设置'}
                                    </span>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <User className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-sm">
                                        <span className="font-medium">年级:</span> {student.grade || '未设置'}
                                    </span>
                                </div>
                            </div>

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                                <div>
                                    <span className="text-sm font-medium">联系方式:</span>
                                    <p className="text-sm text-muted-foreground">{student.contact || '未设置'}</p>
                                </div>
                                <div>
                                    <span className="text-sm font-medium">邮箱:</span>
                                    <p className="text-sm text-muted-foreground">{student.email || '未设置'}</p>
                                </div>
                            </div>

                            <div className="flex items-center justify-between">
                                <div className="text-sm text-muted-foreground">
                                    注册时间: {new Date(student.user.registerTime).toLocaleDateString('zh-CN')}
                                </div>

                                <div className="flex items-center space-x-2">
                                    <Button 
                                        variant="outline" 
                                        size="sm"
                                        onClick={() => {
                                            setSelectedStudent(student);
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
                                            setSelectedStudent(student);
                                            setShowEditModal(true);
                                        }}
                                    >
                                        <Edit className="h-4 w-4 mr-1" />
                                        编辑
                                    </Button>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() => handleDeleteStudent(student.username)}
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

            {/* 创建学生模态框 */}
            {showCreateModal && (
                <CreateStudentModal 
                    onClose={() => setShowCreateModal(false)}
                    onSubmit={handleCreateStudent}
                />
            )}

            {/* 编辑学生模态框 */}
            {showEditModal && selectedStudent && (
                <EditStudentModal 
                    student={selectedStudent}
                    onClose={() => {
                        setShowEditModal(false);
                        setSelectedStudent(null);
                    }}
                    onSubmit={handleUpdateStudent}
                />
            )}
        </div>
    );
};

// 创建学生模态框组件
interface CreateStudentModalProps {
    onClose: () => void;
    onSubmit: (data: any) => void;
}

const CreateStudentModal: React.FC<CreateStudentModalProps> = ({ onClose, onSubmit }) => {
    const [formData, setFormData] = useState({
        username: '',
        studentID: '',
        name: '',
        college: '',
        major: '',
        class: '',
        contact: '',
        email: '',
        grade: ''
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto">
                <h2 className="text-xl font-bold mb-4">添加学生</h2>
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
                            <Label htmlFor="studentID">学号 *</Label>
                            <Input
                                id="studentID"
                                value={formData.studentID}
                                onChange={(e) => setFormData({...formData, studentID: e.target.value})}
                                required
                            />
                        </div>
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
                            <Label htmlFor="college">学院</Label>
                            <Input
                                id="college"
                                value={formData.college}
                                onChange={(e) => setFormData({...formData, college: e.target.value})}
                            />
                        </div>
                        <div>
                            <Label htmlFor="major">专业</Label>
                            <Input
                                id="major"
                                value={formData.major}
                                onChange={(e) => setFormData({...formData, major: e.target.value})}
                            />
                        </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label htmlFor="class">班级</Label>
                            <Input
                                id="class"
                                value={formData.class}
                                onChange={(e) => setFormData({...formData, class: e.target.value})}
                            />
                        </div>
                        <div>
                            <Label htmlFor="grade">年级</Label>
                            <Input
                                id="grade"
                                value={formData.grade}
                                onChange={(e) => setFormData({...formData, grade: e.target.value})}
                            />
                        </div>
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
                        <Button type="submit">创建学生</Button>
                        <Button type="button" variant="outline" onClick={onClose}>取消</Button>
                    </div>
                </form>
            </div>
        </div>
    );
};

// 编辑学生模态框组件
interface EditStudentModalProps {
    student: Student;
    onClose: () => void;
    onSubmit: (username: string, data: any) => void;
}

const EditStudentModal: React.FC<EditStudentModalProps> = ({ student, onClose, onSubmit }) => {
    const [formData, setFormData] = useState({
        name: student.name,
        college: student.college,
        major: student.major,
        class: student.class,
        contact: student.contact,
        email: student.email,
        grade: student.grade,
        status: student.status
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(student.username, formData);
    };

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto">
                <h2 className="text-xl font-bold mb-4">编辑学生信息</h2>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label>用户名</Label>
                            <Input value={student.username} disabled />
                        </div>
                        <div>
                            <Label>学号</Label>
                            <Input value={student.studentID} disabled />
                        </div>
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
                            <Label htmlFor="college">学院</Label>
                            <Input
                                id="college"
                                value={formData.college}
                                onChange={(e) => setFormData({...formData, college: e.target.value})}
                            />
                        </div>
                        <div>
                            <Label htmlFor="major">专业</Label>
                            <Input
                                id="major"
                                value={formData.major}
                                onChange={(e) => setFormData({...formData, major: e.target.value})}
                            />
                        </div>
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label htmlFor="class">班级</Label>
                            <Input
                                id="class"
                                value={formData.class}
                                onChange={(e) => setFormData({...formData, class: e.target.value})}
                            />
                        </div>
                        <div>
                            <Label htmlFor="grade">年级</Label>
                            <Input
                                id="grade"
                                value={formData.grade}
                                onChange={(e) => setFormData({...formData, grade: e.target.value})}
                            />
                        </div>
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
                                <SelectItem value="active">在读</SelectItem>
                                <SelectItem value="inactive">休学</SelectItem>
                                <SelectItem value="graduated">已毕业</SelectItem>
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

export default Students; 