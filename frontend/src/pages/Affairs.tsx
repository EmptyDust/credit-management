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
import { PlusCircle, Edit, Trash } from "lucide-react";

// Affair type based on affair.go
export type Affair = {
    id: number;
    name: string;
};

// Form schema for validation
const formSchema = z.object({
  name: z.string().min(1, "Affair name is required"),
});

export default function AffairsPage() {
    const [affairs, setAffairs] = useState<Affair[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [isDialogOpen, setIsDialogOpen] = useState(false);
    const [editingAffair, setEditingAffair] = useState<Affair | null>(null);

    const form = useForm<z.infer<typeof formSchema>>({
        resolver: zodResolver(formSchema),
        defaultValues: { name: "" },
    });

    const fetchAffairs = async () => {
        try {
            setLoading(true);
            const response = await apiClient.get('/affairs');
            setAffairs(response.data.affairs || []);
        } catch (err) {
            setError("Failed to fetch affairs.");
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchAffairs();
    }, []);

    const handleDialogOpen = (affair: Affair | null) => {
        setEditingAffair(affair);
        if (affair) {
            form.reset(affair);
        } else {
            form.reset({ name: "" });
        }
        setIsDialogOpen(true);
    };

    const handleDelete = async (id: number) => {
        if (!window.confirm("Are you sure you want to delete this affair?")) return;
        try {
            await apiClient.delete(`/affairs/${id}`);
            fetchAffairs();
        } catch (err) {
            alert("Failed to delete affair.");
        }
    };

    const onSubmit = async (values: z.infer<typeof formSchema>) => {
        setIsSubmitting(true);
        try {
            if (editingAffair) {
                await apiClient.put(`/affairs/${editingAffair.id}`, values);
            } else {
                await apiClient.post("/affairs", values);
            }
            setIsDialogOpen(false);
            fetchAffairs();
        } catch (err) {
            alert(`Failed to ${editingAffair ? 'update' : 'create'} affair.`);
        } finally {
            setIsSubmitting(false);
        }
    };

    return (
        <div className="space-y-4">
            <h1 className="text-3xl font-bold">Affairs Management</h1>
            <div className="flex justify-end">
                <Button onClick={() => handleDialogOpen(null)}>
                    <PlusCircle className="mr-2 h-4 w-4" />
                    Add Affair Type
                </Button>
            </div>
            
            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="sm:max-w-[425px]">
                    <DialogHeader>
                        <DialogTitle>{editingAffair ? "Edit Affair Type" : "Add New Affair Type"}</DialogTitle>
                        <DialogDescription>Enter the name for the new affair type.</DialogDescription>
                    </DialogHeader>
                    <Form {...form}>
                        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
                            <FormField control={form.control} name="name" render={({ field }) => (
                                <FormItem>
                                    <FormLabel>Affair Name</FormLabel>
                                    <FormControl><Input {...field} /></FormControl>
                                    <FormMessage />
                                </FormItem>
                             )} />
                            <Button type="submit" disabled={isSubmitting} className="w-full">
                                {isSubmitting ? "Saving..." : "Save"}
                            </Button>
                        </form>
                    </Form>
                </DialogContent>
            </Dialog>

            <div className="border rounded-md">
                 <Table>
                    <TableHeader>
                        <TableRow>
                            <TableHead>ID</TableHead>
                            <TableHead>Name</TableHead>
                            <TableHead className="text-right">Actions</TableHead>
                        </TableRow>
                    </TableHeader>
                    <TableBody>
                        {loading ? (
                            <TableRow><TableCell colSpan={3} className="text-center">Loading...</TableCell></TableRow>
                        ) : error ? (
                            <TableRow><TableCell colSpan={3} className="text-center text-red-500">{error}</TableCell></TableRow>
                        ) : affairs.map((affair) => (
                            <TableRow key={affair.id}>
                                <TableCell>{affair.id}</TableCell>
                                <TableCell className="font-medium">{affair.name}</TableCell>
                                <TableCell className="text-right space-x-2">
                                    <Button variant="outline" size="icon" onClick={() => handleDialogOpen(affair)}>
                                        <Edit className="h-4 w-4" />
                                    </Button>
                                    <Button variant="destructive" size="icon" onClick={() => handleDelete(affair.id)}>
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