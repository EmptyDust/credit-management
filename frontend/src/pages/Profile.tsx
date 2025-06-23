import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Badge } from '../components/ui/badge';
import { useAuth } from '../contexts/AuthContext';
import { User, Mail, Calendar, Shield, Settings, Key } from 'lucide-react';
import toast from 'react-hot-toast';

const Profile: React.FC = () => {
    const { user, logout } = useAuth();
    const [isEditing, setIsEditing] = useState(false);
    const [formData, setFormData] = useState({
        name: user?.username || '',
        email: 'user@example.com',
        phone: '13800138000',
        department: '计算机科学学院',
    });

    const handleSave = () => {
        // 这里应该调用API更新用户信息
        toast.success('个人信息更新成功');
        setIsEditing(false);
    };

    const handleCancel = () => {
        setIsEditing(false);
        // 重置表单数据
        setFormData({
            name: user?.username || '',
            email: 'user@example.com',
            phone: '13800138000',
            department: '计算机科学学院',
        });
    };

    if (!user) {
        return (
            <div className="flex items-center justify-center h-64">
                <p className="text-muted-foreground">请先登录</p>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* 页面标题 */}
            <div>
                <h1 className="text-3xl font-bold text-foreground">个人资料</h1>
                <p className="text-muted-foreground mt-1">
                    管理您的个人信息和账户设置
                </p>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* 个人信息卡片 */}
                <div className="lg:col-span-2">
                    <Card>
                        <CardHeader>
                            <div className="flex items-center justify-between">
                                <div>
                                    <CardTitle>个人信息</CardTitle>
                                    <CardDescription>
                                        更新您的基本信息
                                    </CardDescription>
                                </div>
                                <Button
                                    variant="outline"
                                    onClick={() => setIsEditing(!isEditing)}
                                >
                                    {isEditing ? '取消' : '编辑'}
                                </Button>
                            </div>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                                <div>
                                    <Label htmlFor="name">用户名</Label>
                                    <Input
                                        id="name"
                                        value={formData.name}
                                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                        disabled={!isEditing}
                                    />
                                </div>
                                <div>
                                    <Label htmlFor="email">邮箱</Label>
                                    <Input
                                        id="email"
                                        type="email"
                                        value={formData.email}
                                        onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                                        disabled={!isEditing}
                                    />
                                </div>
                                <div>
                                    <Label htmlFor="phone">手机号</Label>
                                    <Input
                                        id="phone"
                                        value={formData.phone}
                                        onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                                        disabled={!isEditing}
                                    />
                                </div>
                                <div>
                                    <Label htmlFor="department">所属院系</Label>
                                    <Input
                                        id="department"
                                        value={formData.department}
                                        onChange={(e) => setFormData({ ...formData, department: e.target.value })}
                                        disabled={!isEditing}
                                    />
                                </div>
                            </div>

                            {isEditing && (
                                <div className="flex items-center space-x-2 pt-4">
                                    <Button onClick={handleSave}>
                                        保存更改
                                    </Button>
                                    <Button variant="outline" onClick={handleCancel}>
                                        取消
                                    </Button>
                                </div>
                            )}
                        </CardContent>
                    </Card>
                </div>

                {/* 侧边栏信息 */}
                <div className="space-y-6">
                    {/* 用户信息摘要 */}
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center">
                                <User className="h-5 w-5 mr-2" />
                                账户信息
                            </CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex items-center space-x-3">
                                <div className="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center">
                                    <User className="h-6 w-6 text-primary" />
                                </div>
                                <div>
                                    <p className="font-medium">{user.username}</p>
                                    <Badge variant="outline" className="capitalize">
                                        {user.user_type}
                                    </Badge>
                                </div>
                            </div>

                            <div className="space-y-2 text-sm">
                                <div className="flex items-center space-x-2">
                                    <Mail className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-muted-foreground">邮箱:</span>
                                    <span>{formData.email}</span>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <Calendar className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-muted-foreground">注册时间:</span>
                                    <span>{new Date(user.register_time).toLocaleDateString('zh-CN')}</span>
                                </div>
                                <div className="flex items-center space-x-2">
                                    <Shield className="h-4 w-4 text-muted-foreground" />
                                    <span className="text-muted-foreground">账户状态:</span>
                                    <Badge className="bg-green-100 text-green-800">活跃</Badge>
                                </div>
                            </div>
                        </CardContent>
                    </Card>

                    {/* 快速操作 */}
                    <Card>
                        <CardHeader>
                            <CardTitle>快速操作</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-2">
                            <Button variant="outline" className="w-full justify-start">
                                <Key className="h-4 w-4 mr-2" />
                                修改密码
                            </Button>
                            <Button variant="outline" className="w-full justify-start">
                                <Settings className="h-4 w-4 mr-2" />
                                账户设置
                            </Button>
                            <Button
                                variant="outline"
                                className="w-full justify-start text-red-600 hover:text-red-700"
                                onClick={logout}
                            >
                                退出登录
                            </Button>
                        </CardContent>
                    </Card>

                    {/* 统计信息 */}
                    <Card>
                        <CardHeader>
                            <CardTitle>统计信息</CardTitle>
                        </CardHeader>
                        <CardContent className="space-y-4">
                            <div className="flex items-center justify-between">
                                <span className="text-sm text-muted-foreground">提交申请</span>
                                <span className="font-medium">12</span>
                            </div>
                            <div className="flex items-center justify-between">
                                <span className="text-sm text-muted-foreground">已通过</span>
                                <span className="font-medium text-green-600">8</span>
                            </div>
                            <div className="flex items-center justify-between">
                                <span className="text-sm text-muted-foreground">待审核</span>
                                <span className="font-medium text-yellow-600">3</span>
                            </div>
                            <div className="flex items-center justify-between">
                                <span className="text-sm text-muted-foreground">已拒绝</span>
                                <span className="font-medium text-red-600">1</span>
                            </div>
                            <div className="pt-2 border-t">
                                <div className="flex items-center justify-between">
                                    <span className="text-sm text-muted-foreground">累计学分</span>
                                    <span className="font-medium text-primary">24.5</span>
                                </div>
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>
        </div>
    );
};

export default Profile; 