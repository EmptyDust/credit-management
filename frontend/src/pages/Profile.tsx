import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useAuth } from "@/contexts/AuthContext";
import apiClient from "@/lib/api";
import { 
    User, 
    Mail, 
    Phone, 
    FileSignature, 
    Edit3, 
    Save, 
    X,
    Shield,
    Calendar,
    MapPin,
    GraduationCap,
    Building,
    Award,
    Clock
} from "lucide-react";
import * as z from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "@/components/ui/form";
import toast from "react-hot-toast";

const profileSchema = z.object({
    email: z.string().email("请输入有效的邮箱地址"),
    phone: z.string().optional(),
    real_name: z.string().min(1, "真实姓名不能为空").max(50, "真实姓名最多50个字符"),
});

type ProfileForm = z.infer<typeof profileSchema>;

type UserProfile = {
    email: string;
    phone: string;
    real_name: string;
    username: string;
    user_type: string;
    status: string;
    created_at: string;
    updated_at: string;
    // Student specific fields
    student_id?: string;
    college?: string;
    major?: string;
    class?: string;
    grade?: string;
    // Teacher specific fields
    department?: string;
    title?: string;
    specialty?: string;
};

export default function ProfilePage() {
    const { user, updateUser } = useAuth();
    const [profile, setProfile] = useState<UserProfile | null>(null);
    const [isEditing, setIsEditing] = useState(false);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState("");

    const form = useForm<ProfileForm>({
        resolver: zodResolver(profileSchema),
        defaultValues: {
            email: "",
            phone: "",
            real_name: "",
        },
    });

    useEffect(() => {
        const fetchProfile = async () => {
            if (!user) return;
            try {
                setLoading(true);
                const response = await apiClient.get(`/users/profile`);
                const profileData = response.data;
                setProfile(profileData);
                form.reset({
                    email: profileData.email || "",
                    phone: profileData.phone || "",
                    real_name: profileData.real_name || "",
                });
            } catch (err) {
                setError("获取个人资料失败");
                console.error(err);
                toast.error("获取个人资料失败");
            } finally {
                setLoading(false);
            }
        };
        fetchProfile();
    }, [user, form]);

    const handleSave = async (values: ProfileForm) => {
        if (!profile) return;
        setError("");
        setSaving(true);
        try {
            const updatedProfile = { ...profile, ...values };
            const { status, ...profileWithoutStatus } = updatedProfile;
            await apiClient.put(`/users/profile`, profileWithoutStatus);
            setProfile(updatedProfile);
            updateUser({
                ...user,
                ...profileWithoutStatus
            });
            setIsEditing(false);
            toast.success("个人资料更新成功！");
        } catch (err) {
            setError("更新个人资料失败");
            console.error(err);
            toast.error("更新个人资料失败");
        } finally {
            setSaving(false);
        }
    };

    const handleCancel = () => {
        if (profile) {
            form.reset({
                email: profile.email || "",
                phone: profile.phone || "",
                real_name: profile.real_name || "",
            });
        }
        setIsEditing(false);
        setError("");
    };

    const getStatusBadge = (status: string) => {
        const statusConfig = {
            active: { label: "活跃", color: "bg-green-100 text-green-800" },
            inactive: { label: "停用", color: "bg-gray-100 text-gray-800" }
        };
        
        const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.inactive;
        return <Badge className={config.color}>{config.label}</Badge>;
    };

    const getUserTypeLabel = (userType: string) => {
        const labels = {
            'student': '学生',
            'teacher': '教师',
            'admin': '管理员'
        };
        return labels[userType as keyof typeof labels] || userType;
    };

    if (loading) {
        return (
            <div className="flex justify-center items-center h-64">
                <div className="flex items-center gap-2">
                    <Clock className="h-8 w-8 animate-spin" />
                    <span>加载中...</span>
                </div>
            </div>
        );
    }

    if (error && !profile) {
        return (
            <div className="flex justify-center items-center h-64">
                <div className="text-center">
                    <div className="text-red-500 mb-2">{error}</div>
                    <Button onClick={() => window.location.reload()}>重试</Button>
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-8 p-4 md:p-8">
            <div>
                <h1 className="text-3xl font-bold">个人资料</h1>
                <p className="text-muted-foreground">查看和管理您的个人信息</p>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {/* Basic Information */}
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between">
                        <div>
                            <CardTitle>基本信息</CardTitle>
                            <CardDescription>您的账户基本信息</CardDescription>
                        </div>
                        {!isEditing ? (
                            <Button onClick={() => setIsEditing(true)} variant="outline">
                                <Edit3 className="mr-2 h-4 w-4" /> 编辑
                            </Button>
                        ) : (
                            <div className="flex gap-2">
                                <Button onClick={handleCancel} variant="outline" size="sm">
                                    <X className="h-4 w-4" />
                                </Button>
                                <Button onClick={form.handleSubmit(handleSave)} size="sm" disabled={saving}>
                                    <Save className="mr-2 h-4 w-4" />
                                    {saving ? "保存中..." : "保存"}
                                </Button>
                            </div>
                        )}
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <Form {...form}>
                            <form onSubmit={form.handleSubmit(handleSave)} className="space-y-4">
                                <div className="flex items-center gap-4">
                                    <User className="h-5 w-5 text-muted-foreground" />
                                    <div className="flex-1">
                                        <label className="text-sm font-medium">用户名</label>
                                        <Input value={profile?.username ?? ""} disabled className="mt-1" />
                                    </div>
                                </div>
                                
                                <FormField
                                    control={form.control}
                                    name="real_name"
                                    render={({ field }) => (
                                        <FormItem>
                                            <div className="flex items-center gap-4">
                                                <FileSignature className="h-5 w-5 text-muted-foreground" />
                                                <div className="flex-1">
                                                    <FormLabel>真实姓名</FormLabel>
                                                    <FormControl>
                                                        <Input {...field} disabled={!isEditing} className="mt-1" />
                                                    </FormControl>
                                                    <FormMessage />
                                                </div>
                                            </div>
                                        </FormItem>
                                    )}
                                />
                                
                                <FormField
                                    control={form.control}
                                    name="email"
                                    render={({ field }) => (
                                        <FormItem>
                                            <div className="flex items-center gap-4">
                                                <Mail className="h-5 w-5 text-muted-foreground" />
                                                <div className="flex-1">
                                                    <FormLabel>邮箱地址</FormLabel>
                                                    <FormControl>
                                                        <Input {...field} type="email" disabled={!isEditing} className="mt-1" />
                                                    </FormControl>
                                                    <FormMessage />
                                                </div>
                                            </div>
                                        </FormItem>
                                    )}
                                />
                                
                                <FormField
                                    control={form.control}
                                    name="phone"
                                    render={({ field }) => (
                                        <FormItem>
                                            <div className="flex items-center gap-4">
                                                <Phone className="h-5 w-5 text-muted-foreground" />
                                                <div className="flex-1">
                                                    <FormLabel>手机号码</FormLabel>
                                                    <FormControl>
                                                        <Input {...field} disabled={!isEditing} className="mt-1" />
                                                    </FormControl>
                                                    <FormMessage />
                                                </div>
                                            </div>
                                        </FormItem>
                                    )}
                                />
                            </form>
                        </Form>
                    </CardContent>
                </Card>

                {/* Account Information */}
                <Card>
                    <CardHeader>
                        <CardTitle>账户信息</CardTitle>
                        <CardDescription>您的账户状态和类型</CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="flex items-center gap-4">
                            <Shield className="h-5 w-5 text-muted-foreground" />
                            <div className="flex-1">
                                <label className="text-sm font-medium">用户类型</label>
                                <div className="mt-1">
                                    <Badge variant="outline">{getUserTypeLabel(profile?.user_type || '')}</Badge>
                                </div>
                            </div>
                        </div>
                        
                        <div className="flex items-center gap-4">
                            <Shield className="h-5 w-5 text-muted-foreground" />
                            <div className="flex-1">
                                <label className="text-sm font-medium">账户状态</label>
                                <div className="mt-1">
                                    {getStatusBadge(profile?.status || '')}
                                </div>
                            </div>
                        </div>
                        
                        <div className="flex items-center gap-4">
                            <Calendar className="h-5 w-5 text-muted-foreground" />
                            <div className="flex-1">
                                <label className="text-sm font-medium">注册时间</label>
                                <p className="text-sm text-muted-foreground mt-1">
                                    {profile?.created_at ? new Date(profile.created_at).toLocaleDateString() : '未知'}
                                </p>
                            </div>
                        </div>
                        
                        <div className="flex items-center gap-4">
                            <Clock className="h-5 w-5 text-muted-foreground" />
                            <div className="flex-1">
                                <label className="text-sm font-medium">最后更新</label>
                                <p className="text-sm text-muted-foreground mt-1">
                                    {profile?.updated_at ? new Date(profile.updated_at).toLocaleDateString() : '未知'}
                                </p>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            </div>

            {/* Student Specific Information */}
            {profile?.user_type === 'student' && (
                <Card>
                    <CardHeader>
                        <CardTitle>学生信息</CardTitle>
                        <CardDescription>您的学生档案信息</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="grid gap-4 md:grid-cols-2">
                            <div className="flex items-center gap-4">
                                <GraduationCap className="h-5 w-5 text-muted-foreground" />
                                <div className="flex-1">
                                    <label className="text-sm font-medium">学号</label>
                                    <p className="text-sm text-muted-foreground mt-1">{profile.student_id || '未设置'}</p>
                                </div>
                            </div>
                            
                            <div className="flex items-center gap-4">
                                <MapPin className="h-5 w-5 text-muted-foreground" />
                                <div className="flex-1">
                                    <label className="text-sm font-medium">学院</label>
                                    <p className="text-sm text-muted-foreground mt-1">{profile.college || '未设置'}</p>
                                </div>
                            </div>
                            
                            <div className="flex items-center gap-4">
                                <GraduationCap className="h-5 w-5 text-muted-foreground" />
                                <div className="flex-1">
                                    <label className="text-sm font-medium">专业</label>
                                    <p className="text-sm text-muted-foreground mt-1">{profile.major || '未设置'}</p>
                                </div>
                            </div>
                            
                            <div className="flex items-center gap-4">
                                <Building className="h-5 w-5 text-muted-foreground" />
                                <div className="flex-1">
                                    <label className="text-sm font-medium">班级</label>
                                    <p className="text-sm text-muted-foreground mt-1">{profile.class || '未设置'}</p>
                                </div>
                            </div>
                            
                            <div className="flex items-center gap-4">
                                <Award className="h-5 w-5 text-muted-foreground" />
                                <div className="flex-1">
                                    <label className="text-sm font-medium">年级</label>
                                    <p className="text-sm text-muted-foreground mt-1">{profile.grade || '未设置'}</p>
                                </div>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            )}

            {/* Teacher Specific Information */}
            {profile?.user_type === 'teacher' && (
                <Card>
                    <CardHeader>
                        <CardTitle>教师信息</CardTitle>
                        <CardDescription>您的教师档案信息</CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="grid gap-4 md:grid-cols-2">
                            <div className="flex items-center gap-4">
                                <Building className="h-5 w-5 text-muted-foreground" />
                                <div className="flex-1">
                                    <label className="text-sm font-medium">院系</label>
                                    <p className="text-sm text-muted-foreground mt-1">{profile.department || '未设置'}</p>
                                </div>
                            </div>
                            
                            <div className="flex items-center gap-4">
                                <Award className="h-5 w-5 text-muted-foreground" />
                                <div className="flex-1">
                                    <label className="text-sm font-medium">职称</label>
                                    <p className="text-sm text-muted-foreground mt-1">{profile.title || '未设置'}</p>
                                </div>
                            </div>
                            
                            <div className="flex items-center gap-4 md:col-span-2">
                                <GraduationCap className="h-5 w-5 text-muted-foreground" />
                                <div className="flex-1">
                                    <label className="text-sm font-medium">专业领域</label>
                                    <p className="text-sm text-muted-foreground mt-1">{profile.specialty || '未设置'}</p>
                                </div>
                            </div>
                        </div>
                    </CardContent>
                </Card>
            )}
        </div>
    );
} 