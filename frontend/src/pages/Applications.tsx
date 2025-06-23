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
import apiClient from "@/lib/api";
import { PlusCircle, Search, Edit, Eye } from "lucide-react";

// Types based on application.go
type Application = {
    id: number;
    affair_id: number;
    student_number: string;
    submission_time: string;
    status: string;
    reviewer_id?: number;
    review_comment?: string;
    applied_credits: number;
    approved_credits: number;
};

type Affair = {
    id: number;
    name: string;
};

const formSchema = z.object({
  affair_id: z.string().min(1, "Please select an affair type."),
  details: z.string().min(1, "Details are required."), // Simple text for now
});

export default function ApplicationsPage() {
    const { user } = useAuth();
    const [applications, setApplications] = useState<Application[]>([]);
    const [affairs, setAffairs] = useState<Affair[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isCreateDialogOpen, setCreateDialogOpen] = useState(false);
    const [isDetailDialogOpen, setDetailDialogOpen] = useState(false);
    const [selectedApp, setSelectedApp] = useState<Application | null>(null);

    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: { affair_id: "", details: "" },
    });

    const fetchApplications = async () => {
        try {
            setLoading(true);
            const endpoint = user?.userType === 'student'
                ? `/applications/user/${user.username}`
                : '/applications';
            const response = await apiClient.get(endpoint);
            setApplications(response.data.applications || []);
        } catch (err) {
            setError("Failed to fetch applications.");
        } finally {
            setLoading(false);
        }
    };
    
    const fetchAffairs = async () => {
        try {
            const response = await apiClient.get('/affairs');
            setAffairs(response.data.affairs || []);
        } catch (err) {
            console.error("Failed to fetch affairs for form.");
        }
    };

    useEffect(() => {
        fetchApplications();
        if(user?.userType === 'student') {
            fetchAffairs();
        }
    }, [user]);

    const handleStatusUpdate = async (id: number, status: string) => {
        try {
            await apiClient.put(`/applications/${id}/status`, { status });
            fetchApplications();
        } catch (err) {
            alert("Failed to update status.");
        }
    };
    
    const handleDetailOpen = (app: Application) => {
        setSelectedApp(app);
        setDetailDialogOpen(true);
    }

    const onSubmit = async (values: z.infer<typeof formSchema>) => {
        setIsSubmitting(true);
        try {
            await apiClient.post("/applications", {
                affair_id: parseInt(values.affair_id),
                student_number: user?.username,
                details: JSON.parse(values.details) // Assuming details are JSON string
            });
            setCreateDialogOpen(false);
            fetchApplications();
        } catch (err) {
            alert(`Failed to create application.`);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="space-y-4">
            <div className="flex justify-between items-center">
                <h1 className="text-3xl font-bold">Applications Management</h1>
                {user?.userType === 'student' && (
                    <Button onClick={() => setCreateDialogOpen(true)}>
                        <PlusCircle className="mr-2 h-4 w-4" />
                        New Application
                    </Button>
                )}
            </div>

            {/* Create Application Dialog */}
            <Dialog open={isCreateDialogOpen} onOpenChange={setCreateDialogOpen}>
                 <DialogContent>
                    <DialogHeader>
                        <DialogTitle>Create New Application</DialogTitle>
                        <DialogDescription>Select affair type and provide details.</DialogDescription>
                    </DialogHeader>
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
                            <FormField control={form.control} name="affair_id" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Affair Type</FormLabel>
                                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                                        <FormControl><SelectTrigger><SelectValue placeholder="Select an affair type" /></SelectTrigger></FormControl>
                                        <SelectContent>
                                            {affairs.map(affair => <SelectItem key={affair.id} value={String(affair.id)}>{affair.name}</SelectItem>)}
                                        </SelectContent>
                                    </Select>
                                    <FormMessage />
                                </FormItem>
                             )} />
                             <FormField control={form.control} name="details" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Details (JSON format)</FormLabel>
                                    <FormControl><Input {...field} /></FormControl>
                                    <FormMessage />
                                </FormItem>
                             )} />
                            <Button type="submit" disabled={isSubmitting} className="w-full">
                                {isSubmitting ? "Submitting..." : "Submit Application"}
                            </Button>
                        </form>
                    </Form>
                 </DialogContent>
            </Dialog>
            
            {/* Details Dialog */}
            <Dialog open={isDetailDialogOpen} onOpenChange={setDetailDialogOpen}>
                <DialogContent>
                    <DialogHeader><DialogTitle>Application Details</DialogTitle></DialogHeader>
                    <div>ID: {selectedApp?.id}</div>
                    <div>Student: {selectedApp?.student_number}</div>
                    <div>Status: {selectedApp?.status}</div>
                    <div>Submitted: {selectedApp?.submission_time}</div>
                    {/* Add more details here */}
                </DialogContent>
            </Dialog>

            <div className="border rounded-md">
                 <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>ID</TableHead>
                            {user?.userType !== 'student' && <TableHead>Student</TableHead>}
                            <TableHead>Affair ID</TableHead>
                            <TableHead>Submission Time</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {loading ? (
                            <TableRow><TableCell colSpan={6} className="text-center">Loading...</TableCell></TableRow>
                        ) : error ? (
                            <TableRow><TableCell colSpan={6} className="text-center text-red-500">{error}</TableCell></TableRow>
                        ) : applications.map((app) => (
                            <TableRow key={app.id}>
                                <TableCell>{app.id}</TableCell>
                                {user?.userType !== 'student' && <TableCell>{app.student_number}</TableCell>}
                                <TableCell>{app.affair_id}</TableCell>
                                <TableCell>{new Date(app.submission_time).toLocaleString()}</TableCell>
                                <TableCell>{app.status}</TableCell>
                                <TableCell className="space-x-2">
                                    <Button variant="outline" size="icon" onClick={() => handleDetailOpen(app)}>
                                        <Eye className="h-4 w-4" />
                                    </Button>
                                    {user?.userType !== 'student' && (
                                        <>
                                            <Button size="sm" onClick={() => handleStatusUpdate(app.id, 'Approved')}>Approve</Button>
                                            <Button size="sm" variant="destructive" onClick={() => handleStatusUpdate(app.id, 'Rejected')}>Reject</Button>
                                        </>
                                    )}
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
} 