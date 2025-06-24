import { useState, useEffect } from "react";
import * as z from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useAuth } from "@/contexts/AuthContext";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
    DialogFooter,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import apiClient from "@/lib/api";
import { 
    PlusCircle, 
    Search, 
    Edit, 
    Trash, 
    Filter, 
    RefreshCw,
    Users,
    Building,
    Award,
    AlertCircle,
    BookOpen
} from "lucide-react";
import toast from "react-hot-toast";

// Teacher type based on teacher.go
export type Teacher = {
    username: string;
    name: string;
    contact: string;
    email: string;
    department: string;
    title: string;
    specialty: string;
    status: 'active' | 'inactive';
    created_at?: string;
    updated_at?: string;
};

// Form schema for validation
const formSchema = z.object({
  username: z.string().min(1, "用户名不能为空").max(20, "用户名最多20个字符"),
  name: z.string().min(1, "姓名不能为空").max(50, "姓名最多50个字符"),
  contact: z.string().optional(),
  email: z.string().email({ message: "请输入有效的邮箱地址" }).optional().or(z.literal('')),
  department: z.string().min(1, "院系不能为空"),
  title: z.string().min(1, "职称不能为空"),
  specialty: z.string().optional(),
  status: z.enum(['active', 'inactive']).default('active'),
});

export default function TeachersPage() {
    const { hasPermission } = useAuth();
    const [teachers, setTeachers] = useState<Teacher[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [searchQuery, setSearchQuery] = useState("");
    const [departmentFilter, setDepartmentFilter] = useState<string>("all");
    const [statusFilter, setStatusFilter] = useState<string>("all");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingTeacher, setEditingTeacher] = useState<Teacher | null>(null);

    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: { 
            username: "", 
            name: "", 
            contact: "", 
            email: "", 
            department: "", 
            title: "", 
            specialty: "",
            status: 'active'
        },
    });

    const fetchTeachers = async () => {
        try {
            setLoading(true);
            let endpoint = '/teachers';
            const params = new URLSearchParams();
            
            if (searchQuery) {
                params.append('q', searchQuery);
            }
            if (departmentFilter !== 'all') {
                params.append('department', departmentFilter);
            }
            if (statusFilter !== 'all') {
                params.append('status', statusFilter);
            }
            
            if (params.toString()) {
                endpoint += `?${params.toString()}`;
            }
            
            const response = await apiClient.get(endpoint);
            setTeachers(response.data.teachers || []);
        } catch (err) {
            setError("获取教师列表失败");
            console.error(err);
            toast.error("获取教师列表失败");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTeachers();
    }, [searchQuery, departmentFilter, statusFilter]);

    const handleDialogOpen = (teacher: Teacher | null) => {
        setEditingTeacher(teacher);
        if (teacher) {
            form.reset(teacher);
        } else {
            form.reset({ 
                username: "", 
                name: "", 
                contact: "", 
                email: "", 
                department: "", 
                title: "", 
                specialty: "",
                status: 'active'
            });
        }
        setIsDialogOpen(true);
    };

    const handleDelete = async (username: string) => {
        if (!window.confirm("确定要删除这个教师吗？此操作不可撤销。")) return;
        try {
            await apiClient.delete(`/teachers/${username}`);
            fetchTeachers();
            toast.success("教师删除成功");
        } catch (err) {
            toast.error("删除教师失败");
            console.error(err);
        }
    };

    const onSubmit = async (values: z.infer<typeof formSchema>) => {
        setIsSubmitting(true);
        try {
            if (editingTeacher) {
                await apiClient.put(`/teachers/${editingTeacher.username}`, values);
                toast.success("教师信息更新成功");
            } else {
                await apiClient.post("/teachers", values);
                toast.success("教师创建成功");
            }
            setIsDialogOpen(false);
            fetchTeachers();
        } catch (err) {
            toast.error(`教师${editingTeacher ? '更新' : '创建'}失败`);
            console.error(err);
        } finally {
            setIsSubmitting(false);
        }
    };

    const getStatusBadge = (status: string) => {
        const statusConfig = {
            active: { label: "活跃", color: "bg-green-100 text-green-800" },
            inactive: { label: "停用", color: "bg-gray-100 text-gray-800" }
        };
        
        const config = statusConfig[status as keyof typeof statusConfig] || statusConfig.inactive;
        return <Badge className={config.color}>{config.label}</Badge>;
    };

    const getTitleBadge = (title: string) => {
        const titleColors = {
            '教授': 'bg-purple-100 text-purple-800',
            '副教授': 'bg-blue-100 text-blue-800',
            '讲师': 'bg-green-100 text-green-800',
            '助教': 'bg-yellow-100 text-yellow-800'
        };
        
        const color = titleColors[title as keyof typeof titleColors] || 'bg-gray-100 text-gray-800';
        return <Badge className={color}>{title}</Badge>;
    };

    const filteredTeachers = teachers.filter(teacher => {
        const matchesSearch = teacher.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                            teacher.department.toLowerCase().includes(searchQuery.toLowerCase()) ||
                            teacher.specialty?.toLowerCase().includes(searchQuery.toLowerCase());
        const matchesDepartment = departmentFilter === "all" || teacher.department === departmentFilter;
        const matchesStatus = statusFilter === "all" || teacher.status === statusFilter;
        return matchesSearch && matchesDepartment && matchesStatus;
    });

    const departments = Array.from(new Set(teachers.map(t => t.department).filter(Boolean)));
    const canManageTeachers = hasPermission('manage_teachers');

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold">教师管理</h1>
                    <p className="text-muted-foreground">管理教师信息和档案</p>
                </div>
                {canManageTeachers && (
                    <Button onClick={() => handleDialogOpen(null)}>
                        <PlusCircle className="mr-2 h-4 w-4" />
                        添加教师
                    </Button>
                )}
            </div>

            {/* Statistics Cards */}
            <div className="grid gap-4 md:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">总教师数</CardTitle>
                        <Users className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{teachers.length}</div>
                        <p className="text-xs text-muted-foreground">
                            活跃教师: {teachers.filter(t => t.status === 'active').length}
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">院系数量</CardTitle>
                        <Building className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{departments.length}</div>
                        <p className="text-xs text-muted-foreground">
                            不同院系
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">职称分布</CardTitle>
                        <Award className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {Array.from(new Set(teachers.map(t => t.title).filter(Boolean))).length}
                        </div>
                        <p className="text-xs text-muted-foreground">
                            不同职称
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">专业领域</CardTitle>
                        <BookOpen className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {Array.from(new Set(teachers.map(t => t.specialty).filter(Boolean))).length}
                        </div>
                        <p className="text-xs text-muted-foreground">
                            不同专业
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* Filters */}
            <Card>
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Filter className="h-5 w-5" />
                        筛选和搜索
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="flex gap-4">
                        <div className="flex-1">
                            <div className="relative">
                                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                                <Input
                                    placeholder="搜索姓名、院系或专业..."
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    className="pl-10"
                                />
                            </div>
                        </div>
                        <Select value={departmentFilter} onValueChange={setDepartmentFilter}>
                            <SelectTrigger className="w-48">
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
                        <Select value={statusFilter} onValueChange={setStatusFilter}>
                            <SelectTrigger className="w-32">
                                <SelectValue placeholder="状态" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">全部状态</SelectItem>
                                <SelectItem value="active">活跃</SelectItem>
                                <SelectItem value="inactive">停用</SelectItem>
                            </SelectContent>
                        </Select>
                        <Button variant="outline" onClick={fetchTeachers} disabled={loading}>
                            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                        </Button>
                    </div>
                </CardContent>
            </Card>

            {/* Teachers Table */}
            <Card>
                <CardHeader>
                    <CardTitle>教师列表</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="border rounded-md">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>用户名</TableHead>
                                    <TableHead>姓名</TableHead>
                                    <TableHead>院系</TableHead>
                                    <TableHead>职称</TableHead>
                                    <TableHead>专业</TableHead>
                                    <TableHead>状态</TableHead>
                                    <TableHead className="text-right">操作</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {loading ? (
                                    <TableRow>
                                        <TableCell colSpan={7} className="text-center py-8">
                                            <div className="flex items-center justify-center gap-2">
                                                <RefreshCw className="h-4 w-4 animate-spin" />
                                                加载中...
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ) : error ? (
                                    <TableRow>
                                        <TableCell colSpan={7} className="text-center py-8 text-red-500">
                                            {error}
                                        </TableCell>
                                    </TableRow>
                                ) : filteredTeachers.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={7} className="text-center py-8">
                                            <div className="flex flex-col items-center gap-2 text-muted-foreground">
                                                <AlertCircle className="h-8 w-8" />
                                                <p>暂无教师记录</p>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    filteredTeachers.map((teacher) => (
                                        <TableRow key={teacher.username}>
                                            <TableCell className="font-medium">{teacher.username}</TableCell>
                                            <TableCell>
                                                <div>
                                                    <div className="font-medium">{teacher.name}</div>
                                                    <div className="text-sm text-muted-foreground">{teacher.email}</div>
                                                </div>
                                            </TableCell>
                                            <TableCell>{teacher.department}</TableCell>
                                            <TableCell>{getTitleBadge(teacher.title)}</TableCell>
                                            <TableCell>{teacher.specialty || '-'}</TableCell>
                                            <TableCell>{getStatusBadge(teacher.status)}</TableCell>
                                            <TableCell className="text-right space-x-2">
                                                {canManageTeachers && (
                                                    <>
                                                        <Button 
                                                            variant="outline" 
                                                            size="icon" 
                                                            onClick={() => handleDialogOpen(teacher)}
                                                        >
                                                            <Edit className="h-4 w-4" />
                                                        </Button>
                                                        <Button 
                                                            variant="destructive" 
                                                            size="icon" 
                                                            onClick={() => handleDelete(teacher.username)}
                                                        >
                                                            <Trash className="h-4 w-4" />
                                                        </Button>
                                                    </>
                                                )}
                                            </TableCell>
                                        </TableRow>
                                    ))
                                )}
                            </TableBody>
                        </Table>
                    </div>
                </CardContent>
            </Card>
            
            {/* Create/Edit Dialog */}
            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="sm:max-w-[600px]">
                    <DialogHeader>
                        <DialogTitle>
                            {editingTeacher ? "编辑教师" : "添加新教师"}
                        </DialogTitle>
                        <DialogDescription>
                            {editingTeacher ? "修改教师信息" : "填写教师详细信息"}
                        </DialogDescription>
                    </DialogHeader>
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="grid grid-cols-2 gap-4 py-4">
                            <FormField control={form.control} name="username" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>用户名</FormLabel>
                                    <FormControl>
                                        <Input {...field} disabled={!!editingTeacher} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )} />
                            <FormField control={form.control} name="name" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>姓名</FormLabel>
                                    <FormControl>
                                        <Input {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                             )} />
                            <FormField control={form.control} name="email" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>邮箱</FormLabel>
                                    <FormControl>
                                        <Input {...field} type="email" />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )} />
                            <FormField control={form.control} name="department" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>院系</FormLabel>
                                    <FormControl>
                                        <Input {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )} />
                            <FormField control={form.control} name="title" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>职称</FormLabel>
                                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                                        <FormControl>
                                            <SelectTrigger>
                                                <SelectValue placeholder="选择职称" />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem value="教授">教授</SelectItem>
                                            <SelectItem value="副教授">副教授</SelectItem>
                                            <SelectItem value="讲师">讲师</SelectItem>
                                            <SelectItem value="助教">助教</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )} />
                            <FormField control={form.control} name="specialty" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>专业领域</FormLabel>
                                    <FormControl>
                                        <Input {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )} />
                            <FormField control={form.control} name="contact" render={({ field }) => (
                               <FormItem className="col-span-2">
                                   <FormLabel>联系方式</FormLabel>
                                   <FormControl>
                                       <Input {...field} />
                                   </FormControl>
                                   <FormMessage />
                               </FormItem>
                            )} />
                            <FormField control={form.control} name="status" render={({ field }) => (
                                <FormItem className="col-span-2">
                                    <FormLabel>状态</FormLabel>
                                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                                        <FormControl>
                                            <SelectTrigger>
                                                <SelectValue />
                                            </SelectTrigger>
                                        </FormControl>
                                        <SelectContent>
                                            <SelectItem value="active">活跃</SelectItem>
                                            <SelectItem value="inactive">停用</SelectItem>
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                            )} />
                            <DialogFooter className="col-span-2">
                                <Button type="submit" disabled={isSubmitting}>
                                    {isSubmitting ? "保存中..." : "保存"}
                                </Button>
                            </DialogFooter>
                        </form>
                    </Form>
                </DialogContent>
            </Dialog>
        </div>
    );
} 