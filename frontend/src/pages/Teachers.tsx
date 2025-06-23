import { useState, useEffect } from "react";
import * as z from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
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
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import apiClient from "@/lib/api";
import { PlusCircle, Search, Edit, Trash } from "lucide-react";

// Teacher type based on teacher.go
export type Teacher = {
    username: string;
    name: string;
    contact: string;
    email: string;
    department: string;
    title: string;
    specialty: string;
    status: string;
};

// Form schema for validation
const formSchema = z.object({
  username: z.string().min(1, "Username is required"),
  name: z.string().min(1, "Name is required"),
  contact: z.string().optional(),
  email: z.string().email({ message: "Invalid email address" }).optional().or(z.literal('')),
  department: z.string().optional(),
  title: z.string().optional(),
  specialty: z.string().optional(),
});

export default function TeachersPage() {
    const [teachers, setTeachers] = useState<Teacher[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [searchQuery, setSearchQuery] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingTeacher, setEditingTeacher] = useState<Teacher | null>(null);

    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: { username: "", name: "", contact: "", email: "", department: "", title: "", specialty: "" },
    });

    const fetchTeachers = async () => {
        try {
            setLoading(true);
            const endpoint = searchQuery ? `/teachers/search?q=${searchQuery}` : '/teachers';
            const response = await apiClient.get(endpoint);
            setTeachers(response.data.teachers || []);
        } catch (err) {
            setError("Failed to fetch teachers.");
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchTeachers();
    }, []);

    const handleSearch = (e: React.FormEvent) => {
        e.preventDefault();
        fetchTeachers();
    };

    const handleDialogOpen = (teacher: Teacher | null) => {
        setEditingTeacher(teacher);
        if (teacher) {
            form.reset(teacher);
        } else {
            form.reset({ username: "", name: "", contact: "", email: "", department: "", title: "", specialty: "" });
        }
        setIsDialogOpen(true);
    };

    const handleDelete = async (username: string) => {
        if (!window.confirm("Are you sure you want to delete this teacher?")) return;
        try {
            await apiClient.delete(`/teachers/${username}`);
            fetchTeachers();
        } catch (err) {
            alert("Failed to delete teacher.");
        }
    };

    const onSubmit = async (values: z.infer<typeof formSchema>) => {
        setIsSubmitting(true);
        try {
            if (editingTeacher) {
                await apiClient.put(`/teachers/${editingTeacher.username}`, values);
            } else {
                await apiClient.post("/teachers", values);
            }
            setIsDialogOpen(false);
            fetchTeachers();
        } catch (err) {
            alert(`Failed to ${editingTeacher ? 'update' : 'create'} teacher.`);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="space-y-4">
            <h1 className="text-3xl font-bold">Teachers Management</h1>
            <div className="flex justify-between items-center">
                <form onSubmit={handleSearch} className="flex items-center gap-2">
                    <Input
                        placeholder="Search by name, department..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-80"
                    />
                    <Button type="submit" variant="outline" size="icon"><Search className="h-4 w-4" /></Button>
                </form>
                <Button onClick={() => handleDialogOpen(null)}>
                    <PlusCircle className="mr-2 h-4 w-4" />
                    Add Teacher
                </Button>
            </div>
            
            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="sm:max-w-[600px]">
                    <DialogHeader>
                        <DialogTitle>{editingTeacher ? "Edit Teacher" : "Add New Teacher"}</DialogTitle>
                        <DialogDescription>Fill in the details for the teacher record.</DialogDescription>
                    </DialogHeader>
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="grid grid-cols-2 gap-4 py-4">
                            <FormField control={form.control} name="username" render={({ field }) => (
                                <FormItem><FormLabel>Username</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <FormField control={form.control} name="name" render={({ field }) => (
                                <FormItem><FormLabel>Full Name</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                             )} />
                            <FormField control={form.control} name="email" render={({ field }) => (
                                <FormItem><FormLabel>Email</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <FormField control={form.control} name="department" render={({ field }) => (
                                <FormItem><FormLabel>Department</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <FormField control={form.control} name="title" render={({ field }) => (
                                <FormItem><FormLabel>Title</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <FormField control={form.control} name="specialty" render={({ field }) => (
                                <FormItem><FormLabel>Specialty</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <FormField control={form.control} name="contact" render={({ field }) => (
                               <FormItem className="col-span-2"><FormLabel>Contact</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <div className="col-span-2">
                                <Button type="submit" disabled={isSubmitting} className="w-full">
                                    {isSubmitting ? "Saving..." : "Save"}
                                </Button>
                            </div>
                        </form>
                    </Form>
                </DialogContent>
            </Dialog>

            <div className="border rounded-md">
                 <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Username</TableHead>
                            <TableHead>Name</TableHead>
                            <TableHead>Department</TableHead>
                            <TableHead>Title</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {loading ? (
                            <TableRow><TableCell colSpan={6} className="text-center">Loading...</TableCell></TableRow>
                        ) : error ? (
                            <TableRow><TableCell colSpan={6} className="text-center text-red-500">{error}</TableCell></TableRow>
                        ) : teachers.map((teacher) => (
                            <TableRow key={teacher.username}>
                                <TableCell>{teacher.username}</TableCell>
                                <TableCell>{teacher.name}</TableCell>
                                <TableCell>{teacher.department}</TableCell>
                                <TableCell>{teacher.title}</TableCell>
                                <TableCell>{teacher.status}</TableCell>
                                <TableCell className="space-x-2">
                                    <Button variant="outline" size="icon" onClick={() => handleDialogOpen(teacher)}>
                                        <Edit className="h-4 w-4" />
                                    </Button>
                                    <Button variant="destructive" size="icon" onClick={() => handleDelete(teacher.username)}>
                                        <Trash className="h-4 w-4" />
                                    </Button>
                                </TableCell>
                            </TableRow>
                        ))}
                    </TableBody>
                </Table>
            </div>
        </div>
    );
} 