import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { useAuth } from '../contexts/AuthContext';
import { authAPI } from '../lib/api';
import { User, Lock, Eye, EyeOff, Loader2 } from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import toast from 'react-hot-toast';

const Login: React.FC = () => {
    const { login } = useAuth();
    const navigate = useNavigate();
    const [formData, setFormData] = useState({
        username: '',
        password: ''
    });
    const [showPassword, setShowPassword] = useState(false);
    const [loading, setLoading] = useState(false);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        
        if (!formData.username || !formData.password) {
            toast.error('请填写用户名和密码');
            return;
        }

        setLoading(true);
        try {
            const response = await authAPI.login(formData);
            const { token, refresh_token, user } = response.data;
            
            // 保存token到本地存储
            localStorage.setItem('token', token);
            localStorage.setItem('refresh_token', refresh_token);
            
            // 更新认证上下文
            await login(token, user);
            
            toast.success('登录成功');
            navigate('/dashboard');
        } catch (error: any) {
            const errorMessage = error.response?.data?.error || '登录失败，请检查用户名和密码';
            toast.error(errorMessage);
            console.error('Login error:', error);
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

    return (
        <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
            <div className="max-w-md w-full space-y-8">
                {/* Logo和标题 */}
                <div className="text-center">
                    <div className="mx-auto h-12 w-12 bg-primary rounded-full flex items-center justify-center">
                        <User className="h-6 w-6 text-white" />
                    </div>
                    <h2 className="mt-6 text-3xl font-bold text-gray-900">
                        欢迎回来
                    </h2>
                    <p className="mt-2 text-sm text-gray-600">
                        登录您的账户以继续
                    </p>
                </div>

                {/* 登录表单 */}
                <Card>
                    <CardHeader>
                        <CardTitle>用户登录</CardTitle>
                        <CardDescription>
                            请输入您的用户名和密码
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleSubmit} className="space-y-4">
                            <div>
                                <Label htmlFor="username">用户名</Label>
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
                                <Label htmlFor="password">密码</Label>
                                <div className="relative">
                                    <Lock className="absolute left-3 top-3 h-4 w-4 text-muted-foreground" />
                                    <Input
                                        id="password"
                                        name="password"
                                        type={showPassword ? 'text' : 'password'}
                                        value={formData.password}
                                        onChange={handleInputChange}
                                        placeholder="请输入密码"
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

                            <Button
                                type="submit"
                                className="w-full"
                                disabled={loading}
                            >
                                {loading ? (
                                    <>
                                        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                        登录中...
                                    </>
                                ) : (
                                    '登录'
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
                                    还没有账户？{' '}
                                    <Link
                                        to="/register"
                                        className="font-medium text-primary hover:text-primary/80"
                                    >
                                        立即注册
                                    </Link>
                                </p>
                            </div>
                        </div>
                    </CardContent>
                </Card>

                {/* 帮助信息 */}
                <div className="text-center">
                    <p className="text-xs text-muted-foreground">
                        如果您遇到登录问题，请联系系统管理员
                    </p>
                </div>
            </div>
        </div>
    );
};

export default Login; 