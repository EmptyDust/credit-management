import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Badge } from '../components/ui/badge';
import { useAuth } from '../contexts/AuthContext';
import { userAPI } from '../lib/api';
import { User, Mail, Phone, Calendar, Edit, Save, X, Bell, Trash2 } from 'lucide-react';
import toast from 'react-hot-toast';

interface UserProfile {
    id: number;
    username: string;
    name: string;
    contact: string;
    email: string;
    userType: string;
    status: string;
    registerTime: string;
    lastLoginTime?: string;
}

interface Notification {
    id: number;
    userID: number;
    title: string;
    content: string;
    type: string;
    isRead: boolean;
    createdAt: string;
}

const Profile: React.FC = () => {
    const { user } = useAuth();
    const [profile, setProfile] = useState<UserProfile | null>(null);
    const [notifications, setNotifications] = useState<Notification[]>([]);
    const [loading, setLoading] = useState(true);
    const [editing, setEditing] = useState(false);
    const [editData, setEditData] = useState({
        name: '',
        contact: '',
        email: ''
    });

    useEffect(() => {
        fetchProfile();
        fetchNotifications();
    }, []);

    const fetchProfile = async () => {
        try {
            const response = await userAPI.getProfile();
            setProfile(response.data);
            setEditData({
                name: response.data.name,
                contact: response.data.contact,
                email: response.data.email
            });
        } catch (error) {
            toast.error('获取用户信息失败');
            console.error('Failed to fetch profile:', error);
        } finally {
            setLoading(false);
        }
    };

    const fetchNotifications = async () => {
        try {
            const response = await userAPI.getNotifications();
            setNotifications(response.data);
        } catch (error) {
            console.error('Failed to fetch notifications:', error);
        }
    };

    const handleUpdateProfile = async () => {
        try {
            await userAPI.updateProfile(editData);
            toast.success('个人信息更新成功');
            setEditing(false);
            fetchProfile();
        } catch (error) {
            toast.error('更新个人信息失败');
            console.error('Failed to update profile:', error);
        }
    };

    const handleMarkNotificationRead = async (notificationId: number) => {
        try {
            await userAPI.markNotificationRead(notificationId);
            toast.success('通知已标记为已读');
            fetchNotifications();
        } catch (error) {
            toast.error('标记通知失败');
            console.error('Failed to mark notification read:', error);
        }
    };

    const handleDeleteNotification = async (notificationId: number) => {
        if (!confirm('确定要删除这个通知吗？')) return;
        
        try {
            await userAPI.deleteNotification(notificationId);
            toast.success('通知删除成功');
            fetchNotifications();
        } catch (error) {
            toast.error('删除通知失败');
            console.error('Failed to delete notification:', error);
        }
    };

    const getStatusBadge = (status: string) => {
        const statusConfig = {
            active: { color: 'bg-green-100 text-green-800', text: '活跃' },
            inactive: { color: 'bg-yellow-100 text-yellow-800', text: '非活跃' },
            suspended: { color: 'bg-red-100 text-red-800', text: '已暂停' }
        };

        const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.active;
        return <Badge className={config.color}>{config.text}</Badge>;
    };

    const getUserTypeText = (userType: string) => {
        const typeConfig = {
            student: '学生',
            teacher: '教师',
            admin: '管理员'
        };
        return typeConfig[userType as keyof typeof typeConfig] || userType;
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center h-64">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            </div>
        );
    }

    if (!profile) {
        return (
            <div className="text-center py-8">
                <p className="text-muted-foreground">无法加载用户信息</p>
            </div>
        );
    }

    return (
        <div className="space-y-6">
            {/* 页面标题 */}
            <div>
                <h1 className="text-3xl font-bold text-foreground">个人资料</h1>
                <p className="text-muted-foreground mt-1">
                    管理您的个人信息和通知
                </p>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
                {/* 个人信息 */}
                <div className="lg:col-span-2">
                    <Card>
                        <CardHeader>
                            <div className="flex items-center justify-between">
                                <div>
                                    <CardTitle>个人信息</CardTitle>
                                    <CardDescription>
                                        查看和编辑您的个人信息
                                    </CardDescription>
                                </div>
                                {!editing ? (
                                    <Button variant="outline" onClick={() => setEditing(true)}>
                                        <Edit className="h-4 w-4 mr-2" />
                                        编辑
                                    </Button>
                                ) : (
                                    <div className="flex space-x-2">
                                        <Button onClick={handleUpdateProfile}>
                                            <Save className="h-4 w-4 mr-2" />
                                            保存
                                        </Button>
                                        <Button variant="outline" onClick={() => setEditing(false)}>
                                            <X className="h-4 w-4 mr-2" />
                                            取消
                                        </Button>
                                    </div>
                                )}
                            </div>
                        </CardHeader>
                        <CardContent className="space-y-6">
                            {/* 头像和基本信息 */}
                            <div className="flex items-center space-x-4">
                                <div className="w-16 h-16 bg-primary/10 rounded-full flex items-center justify-center">
                                    <User className="h-8 w-8 text-primary" />
                                </div>
                                <div>
                                    <h3 className="text-lg font-semibold">{profile.name}</h3>
                                    <p className="text-muted-foreground">@{profile.username}</p>
                                    <div className="flex items-center space-x-2 mt-1">
                                        {getStatusBadge(profile.status)}
                                        <Badge variant="outline">{getUserTypeText(profile.userType)}</Badge>
                                    </div>
                                </div>
                            </div>

                            {/* 详细信息 */}
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                                <div>
                                    <Label htmlFor="name">姓名</Label>
                                    {editing ? (
                                        <Input
                                            id="name"
                                            value={editData.name}
                                            onChange={(e) => setEditData({...editData, name: e.target.value})}
                                        />
                                    ) : (
                                        <p className="text-sm text-muted-foreground mt-1">{profile.name}</p>
                                    )}
                                </div>

                                <div>
                                    <Label htmlFor="username">用户名</Label>
                                    <p className="text-sm text-muted-foreground mt-1">{profile.username}</p>
                                </div>

                                <div>
                                    <Label htmlFor="contact">联系方式</Label>
                                    {editing ? (
                                        <Input
                                            id="contact"
                                            value={editData.contact}
                                            onChange={(e) => setEditData({...editData, contact: e.target.value})}
                                        />
                                    ) : (
                                        <p className="text-sm text-muted-foreground mt-1 flex items-center">
                                            <Phone className="h-3 w-3 mr-1" />
                                            {profile.contact || '未设置'}
                                        </p>
                                    )}
                                </div>

                                <div>
                                    <Label htmlFor="email">邮箱</Label>
                                    {editing ? (
                                        <Input
                                            id="email"
                                            type="email"
                                            value={editData.email}
                                            onChange={(e) => setEditData({...editData, email: e.target.value})}
                                        />
                                    ) : (
                                        <p className="text-sm text-muted-foreground mt-1 flex items-center">
                                            <Mail className="h-3 w-3 mr-1" />
                                            {profile.email || '未设置'}
                                        </p>
                                    )}
                                </div>

                                <div>
                                    <Label>用户类型</Label>
                                    <p className="text-sm text-muted-foreground mt-1">{getUserTypeText(profile.userType)}</p>
                                </div>

                                <div>
                                    <Label>账户状态</Label>
                                    <div className="mt-1">
                                        {getStatusBadge(profile.status)}
                                    </div>
                                </div>
                            </div>

                            {/* 时间信息 */}
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 pt-4 border-t">
                                <div>
                                    <Label>注册时间</Label>
                                    <p className="text-sm text-muted-foreground mt-1 flex items-center">
                                        <Calendar className="h-3 w-3 mr-1" />
                                        {new Date(profile.registerTime).toLocaleDateString('zh-CN')}
                                    </p>
                                </div>

                                {profile.lastLoginTime && (
                                    <div>
                                        <Label>最后登录</Label>
                                        <p className="text-sm text-muted-foreground mt-1 flex items-center">
                                            <Calendar className="h-3 w-3 mr-1" />
                                            {new Date(profile.lastLoginTime).toLocaleDateString('zh-CN')}
                                        </p>
                                    </div>
                                )}
                            </div>
                        </CardContent>
                    </Card>
                </div>

                {/* 通知 */}
                <div>
                    <Card>
                        <CardHeader>
                            <CardTitle className="flex items-center">
                                <Bell className="h-4 w-4 mr-2" />
                                通知
                                {notifications.filter(n => !n.isRead).length > 0 && (
                                    <Badge className="ml-2" variant="destructive">
                                        {notifications.filter(n => !n.isRead).length}
                                    </Badge>
                                )}
                            </CardTitle>
                            <CardDescription>
                                系统通知和消息
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            <div className="space-y-3">
                                {notifications.length > 0 ? (
                                    notifications.map((notification) => (
                                        <div
                                            key={notification.id}
                                            className={`p-3 border rounded-lg ${
                                                !notification.isRead ? 'bg-blue-50 border-blue-200' : ''
                                            }`}
                                        >
                                            <div className="flex items-start justify-between">
                                                <div className="flex-1">
                                                    <div className="font-medium text-sm">{notification.title}</div>
                                                    <div className="text-xs text-muted-foreground mt-1">
                                                        {notification.content}
                                                    </div>
                                                    <div className="text-xs text-muted-foreground mt-2">
                                                        {new Date(notification.createdAt).toLocaleDateString('zh-CN')}
                                                    </div>
                                                </div>
                                                <div className="flex items-center space-x-1 ml-2">
                                                    {!notification.isRead && (
                                                        <Button
                                                            variant="ghost"
                                                            size="sm"
                                                            onClick={() => handleMarkNotificationRead(notification.id)}
                                                        >
                                                            <Bell className="h-3 w-3" />
                                                        </Button>
                                                    )}
                                                    <Button
                                                        variant="ghost"
                                                        size="sm"
                                                        onClick={() => handleDeleteNotification(notification.id)}
                                                        className="text-red-600 hover:text-red-700"
                                                    >
                                                        <Trash2 className="h-3 w-3" />
                                                    </Button>
                                                </div>
                                            </div>
                                        </div>
                                    ))
                                ) : (
                                    <div className="text-center py-8 text-muted-foreground">
                                        暂无通知
                                    </div>
                                )}
                            </div>
                        </CardContent>
                    </Card>
                </div>
            </div>

            {/* 账户安全 */}
            <Card>
                <CardHeader>
                    <CardTitle>账户安全</CardTitle>
                    <CardDescription>
                        管理您的账户安全设置
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                        <div>
                            <h4 className="font-medium mb-2">密码</h4>
                            <p className="text-sm text-muted-foreground mb-3">
                                定期更改密码以确保账户安全
                            </p>
                            <Button variant="outline">更改密码</Button>
                        </div>
                        <div>
                            <h4 className="font-medium mb-2">登录历史</h4>
                            <p className="text-sm text-muted-foreground mb-3">
                                查看最近的登录活动
                            </p>
                            <Button variant="outline">查看历史</Button>
                        </div>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
};

export default Profile; 