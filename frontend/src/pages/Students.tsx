import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Badge } from '../components/ui/badge';
import { api } from '../lib/api';
import { Plus, Search, Filter, Eye, Edit, Trash2, User } from 'lucide-react';
import toast from 'react-hot-toast';

interface Student {
    id: number;
    student_id: string;
    name: string;
    major: string;
    class_name: string;
    grade: string;
    contact: string;
    email: string;
    status: string;
    created_at: string;
    updated_at: string;
}

const Students: React.FC = () => {
    const [students, setStudents] = useState<Student[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');

    useEffect(() => {
        fetchStudents();
    }, []);

    const fetchStudents = async () => {
        try {
            const response = await api.get('/students');
            setStudents(response.data);
        } catch (error) {
            toast.error('获取学生列表失败');
            console.error('Failed to fetch students:', error);
        } finally {
            setLoading(false);
        }
    };

    const filteredStudents = students.filter(student =>
        student.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        student.student_id.toLowerCase().includes(searchTerm.toLowerCase()) ||
        student.major.toLowerCase().includes(searchTerm.toLowerCase())
    );

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
                        管理学生信息和学分申请
                    </p>
                </div>
                <Button>
                    <Plus className="h-4 w-4 mr-2" />
                    添加学生
                </Button>
            </div>

            {/* 搜索 */}
            <Card>
                <CardContent className="pt-6">
                    <div className="flex items-center space-x-4">
                        <div className="flex-1">
                            <Label htmlFor="search">搜索学生</Label>
                            <div className="relative">
                                <Search className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                                <Input
                                    id="search"
                                    placeholder="搜索姓名、学号或专业..."
                                    value={searchTerm}
                                    onChange={(e) => setSearchTerm(e.target.value)}
                                    className="pl-10"
                                />
                            </div>
                        </div>
                    </div>
                </CardContent>
            </Card>

            {/* 学生列表 */}
            <div className="grid gap-4">
                {filteredStudents.map((student) => (
                    <Card key={student.id}>
                        <CardHeader>
                            <div className="flex items-center justify-between">
                                <div className="flex items-center space-x-3">
                                    <div className="w-10 h-10 bg-primary/10 rounded-full flex items-center justify-center">
                                        <User className="h-5 w-5 text-primary" />
                                    </div>
                                    <div>
                                        <CardTitle className="text-lg">{student.name}</CardTitle>
                                        <CardDescription>
                                            学号: {student.student_id} | 专业: {student.major}
                                        </CardDescription>
                                    </div>
                                </div>
                                <Badge variant={student.status === 'active' ? 'default' : 'secondary'}>
                                    {student.status === 'active' ? '在读' : '毕业'}
                                </Badge>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                                <div>
                                    <span className="text-muted-foreground">班级:</span>
                                    <p className="font-medium">{student.class_name}</p>
                                </div>
                                <div>
                                    <span className="text-muted-foreground">年级:</span>
                                    <p className="font-medium">{student.grade}</p>
                                </div>
                                <div>
                                    <span className="text-muted-foreground">联系方式:</span>
                                    <p className="font-medium">{student.contact}</p>
                                </div>
                                <div>
                                    <span className="text-muted-foreground">邮箱:</span>
                                    <p className="font-medium">{student.email}</p>
                                </div>
                            </div>

                            <div className="flex items-center justify-between mt-4 pt-4 border-t">
                                <div className="text-sm text-muted-foreground">
                                    注册时间: {new Date(student.created_at).toLocaleDateString('zh-CN')}
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

                {filteredStudents.length === 0 && (
                    <Card>
                        <CardContent className="pt-6">
                            <div className="text-center py-8">
                                <p className="text-muted-foreground">暂无学生记录</p>
                            </div>
                        </CardContent>
                    </Card>
                )}
            </div>
        </div>
    );
};

export default Students; 