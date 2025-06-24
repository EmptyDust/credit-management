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
    School,
    GraduationCap,
    AlertCircle,
    UserCheck,
} from "lucide-react";
import toast from "react-hot-toast";

// Updated Student type based on student.go
export type Student = {
    username: string;
    student_id: string;
    name: string;
    college: string;
    major: string;
    class: string;
    contact: string;
    email: string;
    grade: string;
    status: 'active' | 'inactive';
    created_at?: string;
    updated_at?: string;
};

// Form schema for validation
const formSchema = z.object({
  username: z.string().min(1, "用户名不能为空").max(20, "用户名最多20个字符"),
  student_id: z.string().min(1, "学号不能为空").max(20, "学号最多20个字符"),
  name: z.string().min(1, "姓名不能为空").max(50, "姓名最多50个字符"),
  college: z.string().min(1, "学院不能为空"),
  major: z.string().min(1, "专业不能为空"),
  class: z.string().min(1, "班级不能为空"),
  contact: z.string().optional(),
  email: z.string().email({ message: "请输入有效的邮箱地址" }).optional().or(z.literal('')),
  grade: z.string().min(1, "年级不能为空"),
  status: z.enum(['active', 'inactive']).default('active'),
});

export default function StudentsPage() {
    const { hasPermission } = useAuth();
    const [students, setStudents] = useState<Student[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [searchQuery, setSearchQuery] = useState("");
    const [collegeFilter, setCollegeFilter] = useState<string>("all");
    const [statusFilter, setStatusFilter] = useState<string>("all");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingStudent, setEditingStudent] = useState<Student | null>(null);

    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            username: "",
            student_id: "",
            name: "",
            college: "",
            major: "",
            class: "",
            contact: "",
            email: "",
            grade: "",
            status: 'active',
        },
    });

    const fetchStudents = async () => {
        try {
            setLoading(true);
            let endpoint = '/students';
            const params = new URLSearchParams();
            
            if (searchQuery) {
                params.append('q', searchQuery);
            }
            if (collegeFilter !== 'all') {
                params.append('college', collegeFilter);
            }
            if (statusFilter !== 'all') {
                params.append('status', statusFilter);
            }
            
            if (params.toString()) {
                endpoint += `?${params.toString()}`;
            }
            
            const response = await apiClient.get(endpoint);
            setStudents(response.data.students || []);
        } catch (err) {
            setError("获取学生列表失败");
            console.error(err);
            toast.error("获取学生列表失败");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStudents();
    }, [searchQuery, collegeFilter, statusFilter]);

    const handleDialogOpen = (student: Student | null) => {
        setEditingStudent(student);
        if (student) {
            form.reset(student);
        } else {
            form.reset({
                username: "", 
                student_id: "", 
                name: "", 
                college: "",
                major: "", 
                class: "", 
                contact: "", 
                email: "", 
                grade: "",
                status: 'active',
            });
        }
        setIsDialogOpen(true);
    };

    const handleDelete = async (studentId: string) => {
        if (!window.confirm("确定要删除这个学生吗？此操作不可撤销。")) return;
        try {
            await apiClient.delete(`/students/${studentId}`);
            fetchStudents();
            toast.success("学生删除成功");
        } catch (err) {
            toast.error("删除学生失败");
            console.error(err);
        }
    };

    const onSubmit = async (values: z.infer<typeof formSchema>) => {
        setIsSubmitting(true);
        try {
            if (editingStudent) {
                await apiClient.put(`/students/${editingStudent.student_id}`, values);
                toast.success("学生信息更新成功");
            } else {
                await apiClient.post("/students", values);
                toast.success("学生创建成功");
            }
            setIsDialogOpen(false);
            fetchStudents();
        } catch (err) {
            toast.error(`学生${editingStudent ? '更新' : '创建'}失败`);
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

    const filteredStudents = students.filter(student => {
        const matchesSearch = student.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                            student.student_id.toLowerCase().includes(searchQuery.toLowerCase()) ||
                            student.major.toLowerCase().includes(searchQuery.toLowerCase());
        const matchesCollege = collegeFilter === "all" || student.college === collegeFilter;
        const matchesStatus = statusFilter === "all" || student.status === statusFilter;
        return matchesSearch && matchesCollege && matchesStatus;
    });

    const colleges = Array.from(new Set(students.map(s => s.college).filter(Boolean)));
    const canManageStudents = hasPermission('manage_students');

    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <div>
                    <h1 className="text-3xl font-bold">学生管理</h1>
                    <p className="text-muted-foreground">管理学生信息和档案</p>
                </div>
                {canManageStudents && (
                    <Button onClick={() => handleDialogOpen(null)}>
                        <PlusCircle className="mr-2 h-4 w-4" />
                        添加学生
                    </Button>
                )}
            </div>

            {/* Statistics Cards */}
            <div className="grid gap-4 md:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">总学生数</CardTitle>
                        <Users className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{students.length}</div>
                        <p className="text-xs text-muted-foreground">
                            活跃学生: {students.filter(s => s.status === 'active').length}
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">学院数量</CardTitle>
                        <School className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{colleges.length}</div>
                        <p className="text-xs text-muted-foreground">
                            不同学院
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">专业数量</CardTitle>
                        <GraduationCap className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {Array.from(new Set(students.map(s => s.major).filter(Boolean))).length}
                        </div>
                        <p className="text-xs text-muted-foreground">
                            不同专业
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">年级分布</CardTitle>
                        <UserCheck className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {Array.from(new Set(students.map(s => s.grade).filter(Boolean))).length}
                        </div>
                        <p className="text-xs text-muted-foreground">
                            不同年级
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
                                    placeholder="搜索姓名、学号或专业..."
                                    value={searchQuery}
                                    onChange={(e) => setSearchQuery(e.target.value)}
                                    className="pl-10"
                                />
                            </div>
                        </div>
                        <Select value={collegeFilter} onValueChange={setCollegeFilter}>
                            <SelectTrigger className="w-48">
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
                        <Button variant="outline" onClick={fetchStudents} disabled={loading}>
                            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                        </Button>
                    </div>
                </CardContent>
            </Card>

            {/* Students Table */}
            <Card>
                <CardHeader>
                    <CardTitle>学生列表</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="border rounded-md">
                        <Table>
                            <TableHeader>
                                <TableRow>
                                    <TableHead>学号</TableHead>
                                    <TableHead>姓名</TableHead>
                                    <TableHead>学院</TableHead>
                                    <TableHead>专业</TableHead>
                                    <TableHead>班级</TableHead>
                                    <TableHead>年级</TableHead>
                                    <TableHead>状态</TableHead>
                                    <TableHead className="text-right">操作</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {loading ? (
                                    <TableRow>
                                        <TableCell colSpan={8} className="text-center py-8">
                                            <div className="flex items-center justify-center gap-2">
                                                <RefreshCw className="h-4 w-4 animate-spin" />
                                                加载中...
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ) : error ? (
                                    <TableRow>
                                        <TableCell colSpan={8} className="text-center py-8 text-red-500">
                                            {error}
                                        </TableCell>
                                    </TableRow>
                                ) : filteredStudents.length === 0 ? (
                                    <TableRow>
                                        <TableCell colSpan={8} className="text-center py-8">
                                            <div className="flex flex-col items-center gap-2 text-muted-foreground">
                                                <AlertCircle className="h-8 w-8" />
                                                <p>暂无学生记录</p>
                                            </div>
                                        </TableCell>
                                    </TableRow>
                                ) : (
                                    filteredStudents.map((student) => (
                                        <TableRow key={student.student_id}>
                                            <TableCell className="font-medium">{student.student_id}</TableCell>
                                            <TableCell>
                                                <div>
                                                    <div className="font-medium">{student.name}</div>
                                                    <div className="text-sm text-muted-foreground">{student.username}</div>
                                                </div>
                                            </TableCell>
                                            <TableCell>{student.college}</TableCell>
                                            <TableCell>{student.major}</TableCell>
                                            <TableCell>{student.class}</TableCell>
                                            <TableCell>{student.grade}</TableCell>
                                            <TableCell>{getStatusBadge(student.status)}</TableCell>
                                            <TableCell className="text-right space-x-2">
                                                {canManageStudents && (
                                                    <>
                                                        <Button 
                                                            variant="outline" 
                                                            size="icon" 
                                                            onClick={() => handleDialogOpen(student)}
                                                        >
                                                            <Edit className="h-4 w-4" />
                                                        </Button>
                                                        <Button 
                                                            variant="destructive" 
                                                            size="icon" 
                                                            onClick={() => handleDelete(student.student_id)}
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
                            {editingStudent ? "编辑学生" : "添加新学生"}
                        </DialogTitle>
                        <DialogDescription>
                            {editingStudent ? "修改学生信息" : "填写学生详细信息"}
                        </DialogDescription>
                    </DialogHeader>
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="grid grid-cols-2 gap-4 py-4">
                            <FormField control={form.control} name="username" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>用户名</FormLabel>
                                    <FormControl>
                                        <Input {...field} disabled={!!editingStudent} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )} />
                            <FormField control={form.control} name="student_id" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>学号</FormLabel>
                                    <FormControl>
                                        <Input {...field} disabled={!!editingStudent} />
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
                            <FormField control={form.control} name="college" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>学院</FormLabel>
                                    <FormControl>
                                        <Input {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )} />
                            <FormField control={form.control} name="major" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>专业</FormLabel>
                                    <FormControl>
                                        <Input {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )} />
                            <FormField control={form.control} name="class" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>班级</FormLabel>
                                    <FormControl>
                                        <Input {...field} />
                                    </FormControl>
                                    <FormMessage />
                                </FormItem>
                            )} />
                             <FormField control={form.control} name="grade" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>年级</FormLabel>
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