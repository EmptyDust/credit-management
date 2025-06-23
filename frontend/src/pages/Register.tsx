import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '../components/ui/select';
import { authAPI } from '../lib/api';
import { User, Lock, Eye, EyeOff, Loader2, Mail, UserPlus } from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';

const Register: React.FC = () => {
    const navigate = useNavigate();
    const [formData, setFormData] = useState({
        username: '',
        password: '',
        confirmPassword: '',
        userType: ''
    });
    const [showPassword, setShowPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        
        // 验证表单
        if (!formData.username || !formData.password || !formData.confirmPassword || !formData.userType) {
            toast.error('请填写所有必填字段');
            return;
        }

        if (formData.password !== formData.confirmPassword) {
            toast.error('两次输入的密码不一致');
            return;
        }

        if (formData.password.length < 6) {
            toast.error('密码长度至少为6位');
            return;
        }

        setLoading(true);
        try {
            await authAPI.register({
                username: formData.username,
                password: formData.password,
                userType: formData.userType
            });
            
            toast.success('注册成功，请登录');
            navigate('/login');
        } catch (error: any) {
            const errorMessage = error.response?.data?.error || '注册失败，请稍后重试';
            toast.error(errorMessage);
            console.error('Register error:', error);
        } finally {
            setLoading(false);
        }
    };

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        setFormData({
            ...formData,
            [e.target.name]: e.target.value
        });
    };

    const handleSelectChange = (value: string) => {
        setFormData({
            ...formData,
            userType: value
        });
    };

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-md w-full space-y-8">
                {/* Logo和标题 */}
                <div className="text-center">
                    <div className="mx-auto h-12 w-12 bg-primary rounded-full flex items-center justify-center">
                        <UserPlus className="h-6 w-6 text-white" />
                    </div>
                    <h2 className="mt-6 text-3xl font-bold text-gray-900">
                        创建账户
                    </h2>
                    <p className="mt-2 text-sm text-gray-600">
                        注册新账户以开始使用系统
                    </p>
                </div>

                {/* 注册表单 */}
                <Card>
                    <CardHeader>
                        <CardTitle>用户注册</CardTitle>
                        <CardDescription>
                            请填写以下信息完成注册
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div>
                                <Label htmlFor="username">用户名 *</Label>
                                <div className="relative">
                                    <User className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                                    <Input
                                        id="username"
                                        name="username"
                                        type="text"
                                        value={formData.username}
                                        onChange={handleInputChange}
                                        placeholder="请输入用户名"
                                        className="pl-10"
                                        required
                                    />
                                </div>
                            </div>

                            <div>
                                <Label htmlFor="userType">用户类型 *</Label>
                                <Select value={formData.userType} onValueChange={handleSelectChange}>
                                    <SelectTrigger>
                                        <SelectValue placeholder="选择用户类型" />
                                    </SelectTrigger>
                                    <SelectContent>
                                        <SelectItem value="student">学生</SelectItem>
                                        <SelectItem value="teacher">教师</SelectItem>
                                        <SelectItem value="admin">管理员</SelectItem>
                                    </SelectContent>
                                </Select>
                            </div>

                            <div>
                                <Label htmlFor="password">密码 *</Label>
                                <div className="relative">
                                    <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                                    <Input
                                        id="password"
                                        name="password"
                                        type={showPassword ? 'text' : 'password'}
                                        value={formData.password}
                                        onChange={handleInputChange}
                                        placeholder="请输入密码（至少6位）"
                                        className="pl-10 pr-10"
                                        required
                                    />
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="sm"
                                        className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                                        onClick={() => setShowPassword(!showPassword)}
                                    >
                                        {showPassword ? (
                                            <EyeOff className="h-4 w-4" />
                                        ) : (
                                            <Eye className="h-4 w-4" />
                                        )}
                                    </Button>
                                </div>
                            </div>

                            <div>
                                <Label htmlFor="confirmPassword">确认密码 *</Label>
                                <div className="relative">
                                    <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                                    <Input
                                        id="confirmPassword"
                                        name="confirmPassword"
                                        type={showConfirmPassword ? 'text' : 'password'}
                                        value={formData.confirmPassword}
                                        onChange={handleInputChange}
                                        placeholder="请再次输入密码"
                                        className="pl-10 pr-10"
                                        required
                                    />
                                    <Button
                                        type="button"
                                        variant="ghost"
                                        size="sm"
                                        className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                                        onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                    >
                                        {showConfirmPassword ? (
                                            <EyeOff className="h-4 w-4" />
                                        ) : (
                                            <Eye className="h-4 w-4" />
                                        )}
                                    </Button>
                                </div>
                            </div>

                            <Button
                                type="submit"
                                className="w-full"
                                disabled={loading}
                            >
                                {loading ? (
                                    <>
                                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                        注册中...
                                    </>
                                ) : (
                                    '注册'
                                )}
                            </Button>
                        </form>

                        {/* 其他选项 */}
                        <div className="mt-6">
                            <div className="relative">
                                <div className="absolute inset-0 flex items-center">
                                    <span className="w-full border-t" />
                                </div>
                                <div className="relative flex justify-center text-xs uppercase">
                                    <span className="bg-background px-2 text-muted-foreground">
                                        或者
                                    </span>
                                </div>
                            </div>

                            <div className="mt-6 text-center">
                                <p className="text-sm text-muted-foreground">
                                    已有账户？{' '}
                                    <Link
                                        to="/login"
                                        className="font-medium text-primary hover:text-primary/80"
                                    >
                                        立即登录
                                    </Link>
                                </p>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* 注册说明 */}
                <Card>
                    <CardHeader>
                        <CardTitle className="text-sm">注册说明</CardTitle>
                    </CardHeader>
                    <CardContent className="text-xs text-muted-foreground space-y-2">
                        <p>• 用户名必须唯一，不能与其他用户重复</p>
                        <p>• 密码长度至少为6位，建议包含字母和数字</p>
                        <p>• 学生用户注册后需要完善个人信息</p>
                        <p>• 教师用户需要管理员审核后才能使用</p>
                        <p>• 注册即表示同意系统的使用条款</p>
                    </CardContent>
                </Card>

                {/* 帮助信息 */}
                <div className="text-center">
                    <p className="text-xs text-muted-foreground">
                        如果您遇到注册问题，请联系系统管理员
                    </p>
                </div>
            </div>
        </div>
    );
};

export default Register; 