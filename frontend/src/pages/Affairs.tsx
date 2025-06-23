import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Badge } from '../components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { affairAPI } from '../lib/api';
import { Plus, Search, Filter, Eye, Edit, Trash2, Users, Calendar, Award } from 'lucide-react';
import toast from 'react-hot-toast';

interface Affair {
    id: number;
    name: string;
    description: string;
    category: string;
    maxCredits: number;
    status: string;
    createdAt: string;
    updatedAt: string;
}

interface AffairStudent {
    affairID: number;
    studentID: string;
    isMainResponsible: boolean;
    createdAt: string;
    affair: Affair;
    student: {
        username: string;
        studentID: string;
        name: string;
        college: string;
        major: string;
        class: string;
    };
}

const Affairs: React.FC = () => {
    const [affairs, setAffairs] = useState<Affair[]>([]);
    const [affairStudents, setAffairStudents] = useState<AffairStudent[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [categoryFilter, setCategoryFilter] = useState('all');
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showEditModal, setShowEditModal] = useState(false);
    const [showStudentModal, setShowStudentModal] = useState(false);
    const [selectedAffair, setSelectedAffair] = useState<Affair | null>(null);
    const [categories, setCategories] = useState<string[]>([]);

    useEffect(() => {
        fetchAffairs();
    }, []);

    const fetchAffairs = async () => {
        try {
            const response = await affairAPI.getAffairs();
            setAffairs(response.data);
            
            // 提取唯一的类别
            const uniqueCategories = [...new Set(response.data.map(a => a.category).filter(Boolean))];
            setCategories(uniqueCategories);
        } catch (error) {
            toast.error('获取事项列表失败');
            console.error('Failed to fetch affairs:', error);
        } finally {
            setLoading(false);
        }
    };

    const fetchAffairStudents = async (affairID: number) => {
        try {
            const response = await affairAPI.getStudentsByAffair(affairID);
            setAffairStudents(response.data);
        } catch (error) {
            console.error('Failed to fetch affair students:', error);
        }
    };

    const handleCreateAffair = async (affairData: any) => {
        try {
            await affairAPI.createAffair(affairData);
            toast.success('事项创建成功');
            fetchAffairs();
            setShowCreateModal(false);
        } catch (error) {
            toast.error('创建事项失败');
            console.error('Failed to create affair:', error);
        }
    };

    const handleUpdateAffair = async (affairId: number, affairData: any) => {
        try {
            await affairAPI.updateAffair(affairId, affairData);
            toast.success('事项更新成功');
            fetchAffairs();
            setShowEditModal(false);
            setSelectedAffair(null);
        } catch (error) {
            toast.error('更新事项失败');
            console.error('Failed to update affair:', error);
        }
    };

    const handleDeleteAffair = async (affairId: number) => {
        if (!confirm('确定要删除这个事项吗？')) return;
        
        try {
            await affairAPI.deleteAffair(affairId);
            toast.success('事项删除成功');
            fetchAffairs();
        } catch (error) {
            toast.error('删除事项失败');
            console.error('Failed to delete affair:', error);
        }
    };

    const handleAddStudentToAffair = async (affairID: number, studentID: string, isMainResponsible: boolean) => {
        try {
            await affairAPI.addStudentToAffair({
                affairID,
                studentID,
                isMainResponsible
            });
            toast.success('学生添加成功');
            fetchAffairStudents(affairID);
        } catch (error) {
            toast.error('添加学生失败');
            console.error('Failed to add student to affair:', error);
        }
    };

    const handleRemoveStudentFromAffair = async (affairID: number, studentID: string) => {
        if (!confirm('确定要移除这个学生吗？')) return;
        
        try {
            await affairAPI.removeStudentFromAffair(affairID, studentID);
            toast.success('学生移除成功');
            fetchAffairStudents(affairID);
        } catch (error) {
            toast.error('移除学生失败');
            console.error('Failed to remove student from affair:', error);
        }
    };

    const filteredAffairs = affairs.filter(affair => {
        const matchesSearch = affair.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
            affair.description.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || affair.status === statusFilter;
        const matchesCategory = categoryFilter === 'all' || affair.category === categoryFilter;
        return matchesSearch && matchesStatus && matchesCategory;
    });

    const getStatusBadge = (status: string) => {
        const statusConfig = {
            active: { color: 'bg-green-100 text-green-800', text: '活跃' },
            inactive: { color: 'bg-gray-100 text-gray-800', text: '非活跃' }
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
                    <h1 className="text-3xl font-bold text-foreground">事项管理</h1>
                    <p className="text-muted-foreground mt-1">
                        管理创新创业事项和学生参与情况
                    </p>
                </div>
                <Button onClick={() => setShowCreateModal(true)}>
                    <Plus className="h-4 w-4 mr-2" />
                    创建事项
                </Button>
            </div>

            {/* 搜索和筛选 */}
            <Card>
                <CardContent className="pt-6">
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
                        <div>
                            <Label htmlFor="search">搜索</Label>
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
                            <Label htmlFor="status">状态筛选</Label>
                            <Select value={statusFilter} onValueChange={setStatusFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择状态" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">全部状态</SelectItem>
                                    <SelectItem value="active">活跃</SelectItem>
                                    <SelectItem value="inactive">非活跃</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div>
                            <Label htmlFor="category">类别筛选</Label>
                            <Select value={categoryFilter} onValueChange={setCategoryFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择类别" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">全部类别</SelectItem>
                                    {categories.map(category => (
                                        <SelectItem key={category} value={category}>
                                            {category}
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
                                    setCategoryFilter('all');
                                }}
                            >
                                <Filter className="h-4 w-4 mr-2" />
                                重置筛选
                            </Button>
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
                                    <CardDescription>
                                        类别: {affair.category} | 最大学分: {affair.maxCredits}
                                    </CardDescription>
                                </div>
                                <div className="flex items-center space-x-2">
                                    {getStatusBadge(affair.status)}
                                </div>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <p className="text-sm text-muted-foreground mb-4">
                                {affair.description}
                            </p>

                            <div className="flex items-center justify-between">
                                <div className="text-sm text-muted-foreground">
                                    创建时间: {new Date(affair.createdAt).toLocaleDateString('zh-CN')}
                                </div>

                                <div className="flex items-center space-x-2">
                                    <Button 
                                        variant="outline" 
                                        size="sm"
                                        onClick={() => {
                                            setSelectedAffair(affair);
                                            fetchAffairStudents(affair.id);
                                            setShowStudentModal(true);
                                        }}
                                    >
                                        <Users className="h-4 w-4 mr-1" />
                                        查看学生
                                    </Button>
                                    <Button 
                                        variant="outline" 
                                        size="sm"
                                        onClick={() => {
                                            setSelectedAffair(affair);
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
                                            setSelectedAffair(affair);
                                            setShowEditModal(true);
                                        }}
                                    >
                                        <Edit className="h-4 w-4 mr-1" />
                                        编辑
                                    </Button>
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() => handleDeleteAffair(affair.id)}
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

            {/* 创建事项模态框 */}
            {showCreateModal && (
                <CreateAffairModal 
                    onClose={() => setShowCreateModal(false)}
                    onSubmit={handleCreateAffair}
                />
            )}

            {/* 编辑事项模态框 */}
            {showEditModal && selectedAffair && (
                <EditAffairModal 
                    affair={selectedAffair}
                    onClose={() => {
                        setShowEditModal(false);
                        setSelectedAffair(null);
                    }}
                    onSubmit={handleUpdateAffair}
                />
            )}

            {/* 学生管理模态框 */}
            {showStudentModal && selectedAffair && (
                <StudentManagementModal 
                    affair={selectedAffair}
                    affairStudents={affairStudents}
                    onClose={() => {
                        setShowStudentModal(false);
                        setSelectedAffair(null);
                    }}
                    onAddStudent={handleAddStudentToAffair}
                    onRemoveStudent={handleRemoveStudentFromAffair}
                />
            )}
        </div>
    );
};

// 创建事项模态框组件
interface CreateAffairModalProps {
    onClose: () => void;
    onSubmit: (data: any) => void;
}

const CreateAffairModal: React.FC<CreateAffairModalProps> = ({ onClose, onSubmit }) => {
    const [formData, setFormData] = useState({
        name: '',
        description: '',
        category: '',
        maxCredits: 0
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(formData);
    };

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto">
                <h2 className="text-xl font-bold mb-4">创建事项</h2>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <Label htmlFor="name">事项名称 *</Label>
                        <Input
                            id="name"
                            value={formData.name}
                            onChange={(e) => setFormData({...formData, name: e.target.value})}
                            required
                        />
                    </div>
                    <div>
                        <Label htmlFor="description">描述</Label>
                        <textarea
                            id="description"
                            value={formData.description}
                            onChange={(e) => setFormData({...formData, description: e.target.value})}
                            className="w-full p-2 border border-gray-300 rounded-md"
                            rows={3}
                        />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label htmlFor="category">类别</Label>
                            <Input
                                id="category"
                                value={formData.category}
                                onChange={(e) => setFormData({...formData, category: e.target.value})}
                            />
                        </div>
                        <div>
                            <Label htmlFor="maxCredits">最大学分</Label>
                            <Input
                                id="maxCredits"
                                type="number"
                                value={formData.maxCredits}
                                onChange={(e) => setFormData({...formData, maxCredits: parseFloat(e.target.value)})}
                            />
                        </div>
                    </div>
                    <div className="flex gap-4">
                        <Button type="submit">创建事项</Button>
                        <Button type="button" variant="outline" onClick={onClose}>取消</Button>
                    </div>
                </form>
            </div>
        </div>
    );
};

// 编辑事项模态框组件
interface EditAffairModalProps {
    affair: Affair;
    onClose: () => void;
    onSubmit: (affairId: number, data: any) => void;
}

const EditAffairModal: React.FC<EditAffairModalProps> = ({ affair, onClose, onSubmit }) => {
    const [formData, setFormData] = useState({
        name: affair.name,
        description: affair.description,
        category: affair.category,
        maxCredits: affair.maxCredits,
        status: affair.status
    });

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();
        onSubmit(affair.id, formData);
    };

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto">
                <h2 className="text-xl font-bold mb-4">编辑事项</h2>
                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <Label htmlFor="name">事项名称 *</Label>
                        <Input
                            id="name"
                            value={formData.name}
                            onChange={(e) => setFormData({...formData, name: e.target.value})}
                            required
                        />
                    </div>
                    <div>
                        <Label htmlFor="description">描述</Label>
                        <textarea
                            id="description"
                            value={formData.description}
                            onChange={(e) => setFormData({...formData, description: e.target.value})}
                            className="w-full p-2 border border-gray-300 rounded-md"
                            rows={3}
                        />
                    </div>
                    <div className="grid grid-cols-2 gap-4">
                        <div>
                            <Label htmlFor="category">类别</Label>
                            <Input
                                id="category"
                                value={formData.category}
                                onChange={(e) => setFormData({...formData, category: e.target.value})}
                            />
                        </div>
                        <div>
                            <Label htmlFor="maxCredits">最大学分</Label>
                            <Input
                                id="maxCredits"
                                type="number"
                                value={formData.maxCredits}
                                onChange={(e) => setFormData({...formData, maxCredits: parseFloat(e.target.value)})}
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
                                <SelectItem value="active">活跃</SelectItem>
                                <SelectItem value="inactive">非活跃</SelectItem>
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

// 学生管理模态框组件
interface StudentManagementModalProps {
    affair: Affair;
    affairStudents: AffairStudent[];
    onClose: () => void;
    onAddStudent: (affairID: number, studentID: string, isMainResponsible: boolean) => void;
    onRemoveStudent: (affairID: number, studentID: string) => void;
}

const StudentManagementModal: React.FC<StudentManagementModalProps> = ({ 
    affair, 
    affairStudents, 
    onClose, 
    onAddStudent, 
    onRemoveStudent 
}) => {
    const [newStudentID, setNewStudentID] = useState('');
    const [isMainResponsible, setIsMainResponsible] = useState(false);

    const handleAddStudent = () => {
        if (!newStudentID.trim()) return;
        onAddStudent(affair.id, newStudentID.trim(), isMainResponsible);
        setNewStudentID('');
        setIsMainResponsible(false);
    };

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 w-full max-w-4xl max-h-[80vh] overflow-y-auto">
                <h2 className="text-xl font-bold mb-4">管理学生 - {affair.name}</h2>
                
                {/* 添加学生 */}
                <div className="mb-6 p-4 border rounded-lg">
                    <h3 className="text-lg font-medium mb-3">添加学生</h3>
                    <div className="flex gap-4 items-end">
                        <div className="flex-1">
                            <Label htmlFor="studentID">学号</Label>
                            <Input
                                id="studentID"
                                value={newStudentID}
                                onChange={(e) => setNewStudentID(e.target.value)}
                                placeholder="输入学号"
                            />
                        </div>
                        <div className="flex items-center space-x-2">
                            <input
                                type="checkbox"
                                id="isMainResponsible"
                                checked={isMainResponsible}
                                onChange={(e) => setIsMainResponsible(e.target.checked)}
                            />
                            <Label htmlFor="isMainResponsible">主要负责人</Label>
                        </div>
                        <Button onClick={handleAddStudent}>添加</Button>
                    </div>
                </div>

                {/* 学生列表 */}
                <div>
                    <h3 className="text-lg font-medium mb-3">参与学生 ({affairStudents.length})</h3>
                    <div className="space-y-2">
                        {affairStudents.map((affairStudent) => (
                            <div key={`${affairStudent.affairID}-${affairStudent.studentID}`} 
                                 className="flex items-center justify-between p-3 border rounded-lg">
                                <div>
                                    <div className="font-medium">{affairStudent.student.name}</div>
                                    <div className="text-sm text-muted-foreground">
                                        学号: {affairStudent.student.studentID} | 
                                        班级: {affairStudent.student.class} | 
                                        专业: {affairStudent.student.major}
                                    </div>
                                </div>
                                <div className="flex items-center space-x-2">
                                    {affairStudent.isMainResponsible && (
                                        <Badge className="bg-blue-100 text-blue-800">主要负责人</Badge>
                                    )}
                                    <Button
                                        variant="outline"
                                        size="sm"
                                        onClick={() => onRemoveStudent(affairStudent.affairID, affairStudent.studentID)}
                                        className="text-red-600 hover:text-red-700"
                                    >
                                        <Trash2 className="h-4 w-4 mr-1" />
                                        移除
                                    </Button>
                                </div>
                            </div>
                        ))}
                        {affairStudents.length === 0 && (
                            <div className="text-center py-8 text-muted-foreground">
                                暂无学生参与此事项
                            </div>
                        )}
                    </div>
                </div>

                <div className="flex justify-end mt-6">
                    <Button onClick={onClose}>关闭</Button>
                </div>
            </div>
        </div>
    );
};

export default Affairs; 