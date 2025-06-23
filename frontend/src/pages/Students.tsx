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
    DialogTrigger,
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
import apiClient from "@/lib/api";
import { PlusCircle, Search, Edit, Trash } from "lucide-react";

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
    status: string;
};

// Form schema for validation
const formSchema = z.object({
  username: z.string().min(1, "Username is required"),
  student_id: z.string().min(1, "Student ID is required"),
  name: z.string().min(1, "Name is required"),
  college: z.string().optional(),
  major: z.string().optional(),
  class: z.string().optional(),
  contact: z.string().optional(),
  email: z.string().email({ message: "Invalid email address" }).optional().or(z.literal('')),
  grade: z.string().optional(),
});

export default function StudentsPage() {
    const [students, setStudents] = useState<Student[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [searchQuery, setSearchQuery] = useState("");
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
        },
    });

    const fetchStudents = async () => {
        try {
            setLoading(true);
            const endpoint = searchQuery ? `/students/search?q=${searchQuery}` : '/students';
            const response = await apiClient.get(endpoint);
            setStudents(response.data.students || []);
        } catch (err) {
            setError("Failed to fetch students.");
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStudents();
    }, []);

    const handleSearch = (e: React.FormEvent) => {
        e.preventDefault();
        fetchStudents();
    };

    const handleDialogOpen = (student: Student | null) => {
        setEditingStudent(student);
        if (student) {
            form.reset(student);
        } else {
            form.reset({
                username: "", student_id: "", name: "", college: "",
                major: "", class: "", contact: "", email: "", grade: "",
            });
        }
        setIsDialogOpen(true);
    };

    const handleDelete = async (studentId: string) => {
        if (!window.confirm("Are you sure you want to delete this student?")) return;
        try {
            await apiClient.delete(`/students/${studentId}`);
            fetchStudents(); // Refresh list
        } catch (err) {
            alert("Failed to delete student.");
            console.error(err);
        }
    };

    const onSubmit = async (values: z.infer<typeof formSchema>) => {
        setIsSubmitting(true);
        try {
            if (editingStudent) {
                await apiClient.put(`/students/${editingStudent.student_id}`, values);
            } else {
                await apiClient.post("/students", values);
            }
            setIsDialogOpen(false);
            fetchStudents();
        } catch (err) {
            alert(`Failed to ${editingStudent ? 'update' : 'create'} student.`);
            console.error(err);
        } finally {
            setIsSubmitting(false);
        }
    };

  return (
        <div className="space-y-4">
            <h1 className="text-3xl font-bold">Students Management</h1>
            <div className="flex justify-between items-center">
                <form onSubmit={handleSearch} className="flex items-center gap-2">
                    <Input
                        placeholder="Search by name, ID, major..."
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="w-80"
                    />
                    <Button type="submit" variant="outline" size="icon">
                        <Search className="h-4 w-4" />
                    </Button>
                </form>
                <Button onClick={() => handleDialogOpen(null)}>
                    <PlusCircle className="mr-2 h-4 w-4" />
                    Add Student
                </Button>
            </div>
            
            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="sm:max-w-[600px]">
                    <DialogHeader>
                        <DialogTitle>{editingStudent ? "Edit Student" : "Add New Student"}</DialogTitle>
                        <DialogDescription>
                            Fill in the details for the student record.
                        </DialogDescription>
                    </DialogHeader>
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="grid grid-cols-2 gap-4 py-4">
                            <FormField control={form.control} name="username" render={({ field }) => (
                                <FormItem><FormLabel>Username</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <FormField control={form.control} name="student_id" render={({ field }) => (
                                <FormItem><FormLabel>Student ID</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <FormField control={form.control} name="name" render={({ field }) => (
                                <FormItem><FormLabel>Full Name</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                             )} />
                            <FormField control={form.control} name="email" render={({ field }) => (
                                <FormItem><FormLabel>Email</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <FormField control={form.control} name="college" render={({ field }) => (
                                <FormItem><FormLabel>College</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <FormField control={form.control} name="major" render={({ field }) => (
                                <FormItem><FormLabel>Major</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <FormField control={form.control} name="class" render={({ field }) => (
                                <FormItem><FormLabel>Class</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                             <FormField control={form.control} name="grade" render={({ field }) => (
                                <FormItem><FormLabel>Grade</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <FormField control={form.control} name="contact" render={({ field }) => (
                               <FormItem className="col-span-2"><FormLabel>Contact</FormLabel><FormControl><Input {...field} /></FormControl><FormMessage /></FormItem>
                            )} />
                            <DialogFooter className="col-span-2">
                                <Button type="submit" disabled={isSubmitting}>
                                    {isSubmitting ? "Saving..." : "Save"}
                                </Button>
                            </DialogFooter>
                        </form>
                    </Form>
                </DialogContent>
            </Dialog>

            {/* Table Display */}
            <div className="border rounded-md">
                 <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>Student ID</TableHead>
                            <TableHead>Name</TableHead>
                            <TableHead>College</TableHead>
                            <TableHead>Major</TableHead>
                            <TableHead>Status</TableHead>
                            <TableHead>Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {loading ? (
                            <TableRow><TableCell colSpan={6} className="text-center">Loading...</TableCell></TableRow>
                        ) : error ? (
                            <TableRow><TableCell colSpan={6} className="text-center text-red-500">{error}</TableCell></TableRow>
                        ) : students.map((student) => (
                            <TableRow key={student.student_id}>
                                <TableCell>{student.student_id}</TableCell>
                                <TableCell>{student.name}</TableCell>
                                <TableCell>{student.college}</TableCell>
                                <TableCell>{student.major}</TableCell>
                                <TableCell>{student.status}</TableCell>
                                <TableCell className="space-x-2">
                                    <Button variant="outline" size="icon" onClick={() => handleDialogOpen(student)}>
                                        <Edit className="h-4 w-4" />
                                    </Button>
                                    <Button variant="destructive" size="icon" onClick={() => handleDelete(student.student_id)}>
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