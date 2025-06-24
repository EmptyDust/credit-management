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
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import apiClient from "@/lib/api";
import { 
    PlusCircle, 
    Edit, 
    Trash, 
    Search, 
    Filter, 
    RefreshCw,
    Award,
    Users,
    Calendar,
    AlertCircle,
    CheckCircle,
    XCircle,
    Eye
} from "lucide-react";
import toast from "react-hot-toast";
import { MultiSelect } from "@/components/ui/multi-select";
import type { Option } from "@/components/ui/multi-select";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
import { useNavigate } from "react-router-dom";

// Types
interface Affair {
    id: string;
    name: string;
    description?: string;
    max_credits?: number;
    status: 'active' | 'inactive';
    category?: string;
    created_at: string;
    updated_at: string;
    student_count?: number;
    application_count?: number;
    attachments?: string;
}

type CreateAffairForm = z.infer<typeof affairSchema>;

interface AttachmentInfo {
    name: string;
    url: string;
}

const affairSchema = z.object({
    name: z.string().min(1, "事务名称不能为空").max(100, "事务名称不能超过100个字符"),
    description: z.string().min(10, "描述至少10个字符").max(500, "描述不能超过500个字符"),
    max_credits: z.number().min(0.5, "最大学分至少0.5").max(10, "最大学分最多10"),
    category: z.string().min(1, "请选择类别"),
    status: z.enum(['active', 'inactive']),
    participants: z.array(z.string()).min(1, "请至少选择一名参与学生"),
    attachments: z.array(z.object({ name: z.string(), url: z.string() }))
});

export default function AffairsPage() {
    const { hasPermission, user } = useAuth();
    const [affairs, setAffairs] = useState<Affair[]>([]);
    const [loading, setLoading] = useState(true);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingAffair, setEditingAffair] = useState<Affair | null>(null);
    const [searchTerm, setSearchTerm] = useState("");
    const [categoryFilter, setCategoryFilter] = useState<string>("all");
    const [statusFilter, setStatusFilter] = useState<string>("all");
    const [studentOptions, setStudentOptions] = useState<Option[]>([]);
    const navigate = useNavigate();

    const form = useForm<CreateAffairForm>({
        resolver: zodResolver(affairSchema),
        defaultValues: { 
            name: "", 
            description: "", 
            max_credits: 1,
            category: "",
            status: 'active',
            participants: [],
            attachments: []
        },
    });

    const fetchAffairs = async () => {
        try {
            setLoading(true);
            const response = await apiClient.get('/affairs');
            setAffairs(response.data.affairs || []);
        } catch (err) {
            console.error("Failed to fetch affairs:", err);
            // Fallback to mock data
            setAffairs([
                {
                    id: "1",
                    name: "创新创业项目",
                    description: "参与各类创新创业项目，包括创业计划书撰写、项目路演等",
                    max_credits: 3,
                    status: 'active',
                    category: '创新创业',
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                    student_count: 25,
                    application_count: 18
                },
                {
                    id: "2",
                    name: "学科竞赛",
                    description: "参加各类学科竞赛，如数学建模、程序设计、英语竞赛等",
                    max_credits: 2,
                    status: 'active',
                    category: '学科竞赛',
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                    student_count: 42,
                    application_count: 35
                },
                {
                    id: "3",
                    name: "志愿服务",
                    description: "参与社会志愿服务，包括社区服务、公益活动等",
                    max_credits: 1,
                    status: 'active',
                    category: '志愿服务',
                    created_at: new Date().toISOString(),
                    updated_at: new Date().toISOString(),
                    student_count: 18,
                    application_count: 12
                }
            ]);
        } finally {
            setLoading(false);
        }
    };

    const fetchStudents = async () => {
        try {
            const res = await apiClient.get('/students');
            setStudentOptions(
                (res.data.students || res.data || []).map((stu: any) => ({
                    label: `${stu.name ?? ''}（${stu.studentID || stu.id || stu.username || ''}）`,
                    value: String(stu.studentID || stu.id || stu.username || '')
                }))
            );
        } catch (e) {
            setStudentOptions([]);
        }
    };

    useEffect(() => {
        fetchAffairs();
        fetchStudents();
    }, []);

    // 调试user对象
    useEffect(() => {
        console.log('user:', user);
    }, [user]);

    const handleDialogOpen = (affair: Affair | null) => {
        setEditingAffair(affair);
        if (affair) {
            apiClient.get(`/affairs/${affair.id}/participants`).then(res => {
                let attachments: AttachmentInfo[] = [];
                try {
                    attachments = affair.attachments ? JSON.parse(affair.attachments) : [];
                } catch { attachments = []; }
                form.reset({ ...affair, participants: ((res.data || []).map((p: any) => String(p.student_id ?? '')) ?? []), attachments });
            }).catch(() => {
                form.reset({ ...affair, participants: [], attachments: [] });
            });
        } else {
            form.reset({ 
                name: "", 
                description: "", 
                max_credits: 1,
                category: "",
                status: 'active',
                participants: [],
                attachments: []
            });
        }
        setIsDialogOpen(true);
    };

    const handleDelete = async (id: string) => {
        if (!window.confirm("确定要删除这个事务吗？此操作不可撤销。")) return;
        
        try {
            await apiClient.delete(`/affairs/${id}`);
            fetchAffairs();
            toast.success("事务删除成功");
        } catch (err) {
            toast.error("删除事务失败");
        }
    };

    const onSubmit = async (values: CreateAffairForm) => {
        try {
            const submitData = {
                ...values,
                attachments: JSON.stringify(values.attachments || [])
            };
            if (editingAffair) {
                await apiClient.put(`/affairs/${editingAffair.id}`, submitData);
                toast.success("事务更新成功");
            } else {
                await apiClient.post("/affairs", submitData);
                toast.success("事务创建成功");
            }
            setIsDialogOpen(false);
            fetchAffairs();
        } catch (err) {
            toast.error(`事务${editingAffair ? '更新' : '创建'}失败`);
        }
    };




    const categories = Array.from(new Set(affairs.map(a => a.category).filter(Boolean)));

    // 附件上传处理
    const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
        const files = e.target.files;
        if (!files) return;
        // 模拟上传，实际应调用后端接口
        const uploaded: AttachmentInfo[] = [];
        for (let i = 0; i < files.length; i++) {
            const file = files[i];
            // 这里只做本地URL预览，实际应上传后返回真实URL
            uploaded.push({ name: file.name, url: URL.createObjectURL(file) });
        }
        form.setValue("attachments", [...(form.getValues("attachments") || []), ...uploaded]);
    };
    const handleRemoveAttachment = (idx: number) => {
        const atts = form.getValues("attachments") || [];
        atts.splice(idx, 1);
        form.setValue("attachments", [...atts]);
    };

    return (
        <div className="space-y-8 p-4 md:p-8">
            <div className="flex flex-col md:flex-row md:justify-between md:items-center gap-4">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">事务管理</h1>
                    <p className="text-muted-foreground">管理学分相关事务类型</p>
                </div>
                {(hasPermission('manage_affairs') || user?.userType === 'student') && (
                    <Button onClick={() => handleDialogOpen(null)} className="rounded-lg shadow transition-all duration-200 hover:scale-105 bg-primary text-white font-bold">
                    <PlusCircle className="mr-2 h-4 w-4" />
                        新建事务
                </Button>
                )}
            </div>

            {/* Statistics Cards */}
            <div className="grid gap-4 md:grid-cols-4">
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">总事务数</CardTitle>
                        <Award className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">{affairs.length}</div>
                        <p className="text-xs text-muted-foreground">
                            活跃事务: {affairs.filter(a => a.status === 'active').length}
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">参与学生</CardTitle>
                        <Users className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {affairs.reduce((sum, affair) => sum + (affair.student_count || 0), 0)}
                        </div>
                        <p className="text-xs text-muted-foreground">
                            总参与人次
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">申请数量</CardTitle>
                        <Calendar className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {affairs.reduce((sum, affair) => sum + (affair.application_count || 0), 0)}
                        </div>
                        <p className="text-xs text-muted-foreground">
                            总申请数量
                        </p>
                    </CardContent>
                </Card>
                <Card>
                    <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                        <CardTitle className="text-sm font-medium">平均学分</CardTitle>
                        <CheckCircle className="h-4 w-4 text-muted-foreground" />
                    </CardHeader>
                    <CardContent>
                        <div className="text-2xl font-bold">
                            {(affairs.reduce((sum, affair) => sum + (affair.max_credits || 0), 0) / affairs.length).toFixed(1)}
                        </div>
                        <p className="text-xs text-muted-foreground">
                            每事务平均学分
                        </p>
                    </CardContent>
                </Card>
            </div>

            {/* Filters */}
            <Card className="rounded-xl shadow-lg">
                <CardHeader>
                    <CardTitle className="flex items-center gap-2">
                        <Filter className="h-5 w-5" />
                        筛选和搜索
                    </CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="flex flex-col md:flex-row gap-4 items-center">
                        <div className="relative w-full max-w-xs">
                            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                            <Input
                                placeholder="搜索事务名称..."
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                                className="pl-10 rounded-lg shadow-sm"
                            />
                        </div>
                        <Select value={categoryFilter} onValueChange={setCategoryFilter}>
                            <SelectTrigger className="w-40">
                                <SelectValue placeholder="选择类别" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="all">全部类别</SelectItem>
                                {categories.filter((category): category is string => Boolean(category)).map(category => (
                                    <SelectItem key={category} value={category}>
                                        {category}
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
                        <Button variant="outline" onClick={fetchAffairs} disabled={loading} className="rounded-lg shadow">
                            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
                        </Button>
                    </div>
                </CardContent>
            </Card>

            {/* Affairs Table */}
            <Card className="rounded-xl shadow-lg">
                <CardHeader>
                    <CardTitle>事务列表</CardTitle>
                </CardHeader>
                <CardContent>
                    <div className="border rounded-xl overflow-x-auto">
                        <Table>
                            <TableHeader className="bg-muted/60">
                                <TableRow>
                                    <TableHead className="font-bold text-primary">ID</TableHead>
                                    <TableHead>名称</TableHead>
                                    <TableHead>描述</TableHead>
                                    <TableHead>最大学分</TableHead>
                                    <TableHead>类别</TableHead>
                                    <TableHead>状态</TableHead>
                                    <TableHead>申请数</TableHead>
                                    <TableHead>操作</TableHead>
                                </TableRow>
                            </TableHeader>
                            <TableBody>
                                {loading ? (
                                    <tr>
                                        <td colSpan={8} className="py-8 text-center">
                                            <div className="flex flex-col items-center gap-2">
                                                <RefreshCw className="h-6 w-6 animate-spin text-muted-foreground" />
                                                <span className="text-muted-foreground">加载中...</span>
                                            </div>
                                        </td>
                                    </tr>
                                ) : affairs.length === 0 ? (
                                    <tr>
                                        <td colSpan={8} className="py-12">
                                            <div className="flex flex-col items-center text-muted-foreground">
                                                <AlertCircle className="w-12 h-12 mb-2" />
                                                <p>暂无事务记录</p>
                                            </div>
                                        </td>
                                    </tr>
                                ) : (
                                    affairs.map(affair => (
                                        <TableRow key={affair.id} className="hover:bg-muted/40 transition-colors">
                                            <TableCell className="font-semibold text-primary">#{affair.id}</TableCell>
                                            <TableCell className="font-medium">{affair.name}</TableCell>
                                            <TableCell className="truncate max-w-xs">{affair.description}</TableCell>
                                            <TableCell><span className="font-bold text-blue-600">{affair.max_credits}</span></TableCell>
                                            <TableCell>
                                                <Badge variant="secondary" className="rounded px-2 py-1">
                                                    {affair.category || '-'}
                                                </Badge>
                                            </TableCell>
                                            <TableCell>
                                                <Badge className={affair.status === 'active' ? 'bg-green-100 text-green-800 rounded-lg px-2 py-1' : 'bg-gray-100 text-gray-800 rounded-lg px-2 py-1'}>
                                                    {affair.status === 'active' ? <CheckCircle className="w-3 h-3 mr-1 inline" /> : <XCircle className="w-3 h-3 mr-1 inline" />}
                                                    {affair.status === 'active' ? '活跃' : '停用'}
                                                </Badge>
                                            </TableCell>
                                            <TableCell>{affair.application_count || 0}</TableCell>
                                            <TableCell>
                                                <div className="flex items-center gap-1">
                                                    <Button size="icon" variant="ghost" className="rounded-full hover:bg-primary/10" title="查看详情" onClick={() => navigate(`/affairs/${affair.id}`)}>
                                                        <Eye className="h-4 w-4" />
                                                    </Button>
                                                    <Button size="icon" variant="ghost" className="rounded-full hover:bg-primary/10" title="编辑" onClick={() => handleDialogOpen(affair)}>
                                                        <Edit className="h-4 w-4" />
                                                    </Button>
                                                    <Button size="icon" variant="ghost" className="rounded-full hover:bg-red-100" title="删除" onClick={() => handleDelete(affair.id)}>
                                                        <Trash className="h-4 w-4" />
                                                    </Button>
                                                </div>
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
                <DialogContent className="sm:max-w-[500px]">
                    <DialogHeader>
                        <DialogTitle>
                            {editingAffair ? "编辑事务" : "添加新事务"}
                        </DialogTitle>
                        <DialogDescription>
                            {editingAffair ? "修改事务信息" : "创建新的事务类型"}
                        </DialogDescription>
                    </DialogHeader>
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                            <FormField
                                control={form.control}
                                name="name"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>事务名称</FormLabel>
                                        <FormControl>
                                            <Input placeholder="请输入事务名称" {...field} />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name="category"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>类别</FormLabel>
                                        <Select onValueChange={field.onChange} defaultValue={field.value}>
                                            <FormControl>
                                                <SelectTrigger>
                                                    <SelectValue placeholder="选择类别" />
                                                </SelectTrigger>
                                            </FormControl>
                                            <SelectContent>
                                                <SelectItem value="创新创业">创新创业</SelectItem>
                                                <SelectItem value="学科竞赛">学科竞赛</SelectItem>
                                                <SelectItem value="志愿服务">志愿服务</SelectItem>
                                                <SelectItem value="学术研究">学术研究</SelectItem>
                                                <SelectItem value="文体活动">文体活动</SelectItem>
                                            </SelectContent>
                                        </Select>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name="max_credits"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>最大学分</FormLabel>
                                        <FormControl>
                                            <Input
                                                type="number"
                                                step="0.5"
                                                min="0.5"
                                                max="10"
                                                placeholder="请输入最大学分"
                                                {...field}
                                                onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name="description"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>描述</FormLabel>
                                        <FormControl>
                                            <Textarea
                                                placeholder="请详细描述该事务的内容、要求和标准..."
                                                className="min-h-[100px]"
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name="status"
                                render={({ field }) => (
                                <FormItem>
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
                                )}
                            />
                            <FormField
                                control={form.control}
                                name="participants"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>参与学生</FormLabel>
                                        <FormControl>
                                            <MultiSelect
                                                options={studentOptions}
                                                selected={field.value ?? []}
                                                onChange={field.onChange}
                                                placeholder="请选择参与学生"
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <FormField
                                control={form.control}
                                name="attachments"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>附件</FormLabel>
                                        <FormControl>
                                            <div>
                                                <input type="file" multiple onChange={handleFileChange} className="mb-2" />
                                                <div className="flex flex-wrap gap-2">
                                                    {(field.value || []).map((att: AttachmentInfo, idx: number) => (
                                                        <div key={att.url} className="flex items-center gap-1 bg-muted rounded px-2 py-1">
                                                            <a href={att.url} target="_blank" rel="noopener noreferrer" className="underline text-sm">{att.name}</a>
                                                            <Button size="icon" variant="ghost" onClick={() => handleRemoveAttachment(idx)}><span className="text-lg">×</span></Button>
                                                        </div>
                                                    ))}
                                                </div>
                                            </div>
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />
                            <DialogFooter>
                                <Button type="submit" className="w-full">
                                    {editingAffair ? "更新事务" : "创建事务"}
                            </Button>
                            </DialogFooter>
                        </form>
                    </Form>
                </DialogContent>
            </Dialog>
        </div>
    );
} 