import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { Badge } from '../components/ui/badge';
import { applicationAPI } from '../lib/api';
import { Plus, Search, Filter, Eye, Edit, Trash2, CheckCircle, XCircle, FileText, Upload, Download } from 'lucide-react';
import toast from 'react-hot-toast';

interface ApplicationType {
    id: number;
    name: string;
    description: string;
    category: string;
    maxCredits: number;
    minCredits: number;
    isActive: boolean;
    createdAt: string;
    updatedAt: string;
}

interface Application {
    id: number;
    userID: number;
    typeID: number;
    title: string;
    description: string;
    content: string;
    status: string;
    credits: number;
    approvedCredits: number;
    reviewerID?: number;
    reviewNote?: string;
    reviewTime?: string;
    submitTime: string;
    createdAt: string;
    updatedAt: string;
    type: ApplicationType;
    files: ApplicationFile[];
}

interface ApplicationFile {
    id: number;
    applicationID: number;
    fileName: string;
    originalName: string;
    fileSize: number;
    fileType: string;
    mimeType: string;
    category: string;
    description: string;
    isRequired: boolean;
    downloadURL: string;
    createdAt: string;
}

const Applications: React.FC = () => {
    const [applications, setApplications] = useState<Application[]>([]);
    const [applicationTypes, setApplicationTypes] = useState<ApplicationType[]>([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [statusFilter, setStatusFilter] = useState('all');
    const [typeFilter, setTypeFilter] = useState('all');
    const [showCreateModal, setShowCreateModal] = useState(false);
    const [showReviewModal, setShowReviewModal] = useState(false);
    const [selectedApplication, setSelectedApplication] = useState<Application | null>(null);

    useEffect(() => {
        fetchApplications();
        fetchApplicationTypes();
    }, []);

    const fetchApplications = async () => {
        try {
            const response = await applicationAPI.getApplications();
            setApplications(response.data);
        } catch (error) {
            toast.error('获取申请列表失败');
            console.error('Failed to fetch applications:', error);
        } finally {
            setLoading(false);
        }
    };

    const fetchApplicationTypes = async () => {
        try {
            const response = await applicationAPI.getApplicationTypes();
            setApplicationTypes(response.data);
        } catch (error) {
            console.error('Failed to fetch application types:', error);
        }
    };

    const handleStatusChange = async (applicationId: number, newStatus: string, approvedCredits: number, reviewNote: string) => {
        try {
            await applicationAPI.updateApplicationStatus(applicationId, {
                status: newStatus,
                approvedCredits: approvedCredits,
                reviewNote: reviewNote
            });
            toast.success('申请状态更新成功');
            fetchApplications();
            setShowReviewModal(false);
            setSelectedApplication(null);
        } catch (error) {
            toast.error('更新申请状态失败');
            console.error('Failed to update application status:', error);
        }
    };

    const handleDeleteApplication = async (applicationId: number) => {
        if (!confirm('确定要删除这个申请吗？')) return;
        
        try {
            await applicationAPI.deleteApplication(applicationId);
            toast.success('申请删除成功');
            fetchApplications();
        } catch (error) {
            toast.error('删除申请失败');
            console.error('Failed to delete application:', error);
        }
    };

    const handleDownloadFile = async (fileId: number, fileName: string) => {
        try {
            const response = await applicationAPI.downloadFile(fileId);
            const url = window.URL.createObjectURL(new Blob([response.data]));
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', fileName);
            document.body.appendChild(link);
            link.click();
            link.remove();
            window.URL.revokeObjectURL(url);
        } catch (error) {
            toast.error('文件下载失败');
            console.error('Failed to download file:', error);
        }
    };

    const filteredApplications = applications.filter(app => {
        const matchesSearch = app.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
            app.description.toLowerCase().includes(searchTerm.toLowerCase());
        const matchesStatus = statusFilter === 'all' || app.status === statusFilter;
        const matchesType = typeFilter === 'all' || app.typeID.toString() === typeFilter;
        return matchesSearch && matchesStatus && matchesType;
    });

    const getStatusBadge = (status: string) => {
        const statusConfig = {
            pending: { color: 'bg-yellow-100 text-yellow-800', text: '待审核' },
            approved: { color: 'bg-green-100 text-green-800', text: '已通过' },
            rejected: { color: 'bg-red-100 text-red-800', text: '已拒绝' },
            cancelled: { color: 'bg-gray-100 text-gray-800', text: '已取消' }
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
                <Button onClick={() => setShowCreateModal(true)}>
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
                                    placeholder="搜索申请标题或描述..."
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
                                    <SelectItem value="approved">已通过</SelectItem>
                                    <SelectItem value="rejected">已拒绝</SelectItem>
                                    <SelectItem value="cancelled">已取消</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div className="w-full md:w-48">
                            <Label htmlFor="type">类型筛选</Label>
                            <Select value={typeFilter} onValueChange={setTypeFilter}>
                                <SelectTrigger>
                                    <SelectValue placeholder="选择类型" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="all">全部类型</SelectItem>
                                    {applicationTypes.map(type => (
                                        <SelectItem key={type.id} value={type.id.toString()}>
                                            {type.name}
                                        </SelectItem>
                                    ))}
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
                                        申请类型: {application.type?.name} |
                                        申请学分: {application.credits} |
                                        审核学分: {application.approvedCredits || '-'}
                                    </CardDescription>
                                </div>
                                <div className="flex items-center space-x-2">
                                    {getStatusBadge(application.status)}
                                    <span className="text-sm text-muted-foreground">
                                        用户ID: {application.userID}
                                    </span>
                                </div>
                            </div>
                        </CardHeader>
                        <CardContent>
                            <p className="text-sm text-muted-foreground mb-4">
                                {application.description}
                            </p>

                            {/* 文件列表 */}
                            {application.files && application.files.length > 0 && (
                                <div className="mb-4">
                                    <h4 className="text-sm font-medium mb-2">附件文件:</h4>
                                    <div className="flex flex-wrap gap-2">
                                        {application.files.map(file => (
                                            <Button
                                                key={file.id}
                                                variant="outline"
                                                size="sm"
                                                onClick={() => handleDownloadFile(file.id, file.originalName)}
                                            >
                                                <Download className="h-3 w-3 mr-1" />
                                                {file.originalName}
                                            </Button>
                                        ))}
                                    </div>
                                </div>
                            )}

                            <div className="flex items-center justify-between">
                                <div className="text-sm text-muted-foreground">
                                    提交时间: {new Date(application.submitTime).toLocaleDateString('zh-CN')}
                                    {application.reviewTime && (
                                        <span className="ml-4">
                                            审核时间: {new Date(application.reviewTime).toLocaleDateString('zh-CN')}
                                        </span>
                                    )}
                                </div>

                                <div className="flex items-center space-x-2">
                                    <Button 
                                        variant="outline" 
                                        size="sm"
                                        onClick={() => {
                                            setSelectedApplication(application);
                                            setShowReviewModal(true);
                                        }}
                                    >
                                        <Eye className="h-4 w-4 mr-1" />
                                        查看详情
                                    </Button>

                                    {application.status === 'pending' && (
                                        <>
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => {
                                                    setSelectedApplication(application);
                                                    setShowReviewModal(true);
                                                }}
                                                className="text-green-600 hover:text-green-700"
                                            >
                                                <CheckCircle className="h-4 w-4 mr-1" />
                                                审核
                                            </Button>
                                            <Button
                                                variant="outline"
                                                size="sm"
                                                onClick={() => handleDeleteApplication(application.id)}
                                                className="text-red-600 hover:text-red-700"
                                            >
                                                <Trash2 className="h-4 w-4 mr-1" />
                                                删除
                                            </Button>
                                        </>
                                    )}
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                ))}
            </div>

            {/* 审核模态框 */}
            {showReviewModal && selectedApplication && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
                    <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto">
                        <h2 className="text-xl font-bold mb-4">审核申请</h2>
                        <div className="space-y-4">
                            <div>
                                <Label>申请标题</Label>
                                <p className="text-sm text-muted-foreground">{selectedApplication.title}</p>
                            </div>
                            <div>
                                <Label>申请描述</Label>
                                <p className="text-sm text-muted-foreground">{selectedApplication.description}</p>
                            </div>
                            <div>
                                <Label>申请内容</Label>
                                <p className="text-sm text-muted-foreground whitespace-pre-wrap">{selectedApplication.content}</p>
                            </div>
                            <div>
                                <Label>申请学分</Label>
                                <p className="text-sm text-muted-foreground">{selectedApplication.credits}</p>
                            </div>
                            
                            <div className="flex gap-4">
                                <Button
                                    onClick={() => handleStatusChange(
                                        selectedApplication.id, 
                                        'approved', 
                                        selectedApplication.credits, 
                                        '申请通过'
                                    )}
                                    className="bg-green-600 hover:bg-green-700"
                                >
                                    <CheckCircle className="h-4 w-4 mr-1" />
                                    通过
                                </Button>
                                <Button
                                    onClick={() => handleStatusChange(
                                        selectedApplication.id, 
                                        'rejected', 
                                        0, 
                                        '申请被拒绝'
                                    )}
                                    variant="destructive"
                                >
                                    <XCircle className="h-4 w-4 mr-1" />
                                    拒绝
                                </Button>
                                <Button
                                    variant="outline"
                                    onClick={() => {
                                        setShowReviewModal(false);
                                        setSelectedApplication(null);
                                    }}
                                >
                                    取消
                                </Button>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Applications; 