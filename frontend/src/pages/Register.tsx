import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
  } from "@/components/ui/select"
import apiClient from "@/lib/api";
import { useNavigate, Link } from "react-router-dom";
import { UserPlus, User, KeyRound, Mail, Phone, FileSignature } from "lucide-react";

export default function Register() {
  const [formData, setFormData] = useState({
    username: "",
    password: "",
    email: "",
    phone: "",
    real_name: "",
    user_type: "student",
  });
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [loading, setLoading] = useState(false);
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});
  const navigate = useNavigate();

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setFormData({ ...formData, [e.target.id]: e.target.value });
    if (formErrors[e.target.id]) {
      setFormErrors({ ...formErrors, [e.target.id]: "" });
    }
  };

  const handleSelectChange = (value: string) => {
    setFormData({ ...formData, user_type: value });
  };

  const validateForm = (): Record<string, string> => {
    const errors: Record<string, string> = {};

    if (formData.username.length < 3 || formData.username.length > 20) {
      errors.username = "Username must be between 3 and 20 characters.";
    }
    if (formData.password.length < 6) {
      errors.password = "Password must be at least 6 characters long.";
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.email)) {
      errors.email = "Please enter a valid email address.";
    }
    if (!formData.real_name) {
      errors.real_name = "Real name is required.";
    }

    return errors;
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setSuccess("");
    setFormErrors({});

    const validationErrors = validateForm();
    if (Object.keys(validationErrors).length > 0) {
      setFormErrors(validationErrors);
      return;
    }

    setLoading(true);
    try {
      await apiClient.post("/users/register", formData);
      setSuccess("Registration successful! You will be redirected to login.");
      setTimeout(() => {
        navigate("/login");
      }, 2000);
    } catch (err: any) {
      const serverError = err.response?.data?.error || "Failed to register. Please try again.";
      setError(serverError);
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-muted/40">
      <Card className="w-full max-w-md">
        <form onSubmit={handleRegister}>
          <CardHeader className="text-center">
            <CardTitle className="text-3xl font-bold flex items-center justify-center gap-2">
              <UserPlus />
              Create an Account
            </CardTitle>
            <CardDescription>
              Enter your details below to create your account.
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4">
            <div>
              <div className="relative">
                  <User className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                  <Input id="username" placeholder="Username" value={formData.username} onChange={handleChange} className="pl-10" />
              </div>
              {formErrors.username && <p className="text-red-500 text-sm pt-1">{formErrors.username}</p>}
            </div>
            <div>
              <div className="relative">
                  <KeyRound className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                  <Input id="password" type="password" placeholder="Password" value={formData.password} onChange={handleChange} className="pl-10" />
              </div>
              {formErrors.password && <p className="text-red-500 text-sm pt-1">{formErrors.password}</p>}
            </div>
            <div>
              <div className="relative">
                  <Mail className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                  <Input id="email" type="email" placeholder="Email" value={formData.email} onChange={handleChange} className="pl-10" />
              </div>
              {formErrors.email && <p className="text-red-500 text-sm pt-1">{formErrors.email}</p>}
            </div>
            <div>
              <div className="relative">
                  <Phone className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                  <Input id="phone" placeholder="Phone (Optional)" value={formData.phone} onChange={handleChange} className="pl-10" />
              </div>
            </div>
            <div>
              <div className="relative">
                  <FileSignature className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
                  <Input id="real_name" placeholder="Real Name" value={formData.real_name} onChange={handleChange} className="pl-10" />
              </div>
              {formErrors.real_name && <p className="text-red-500 text-sm pt-1">{formErrors.real_name}</p>}
            </div>
            <Select onValueChange={handleSelectChange} defaultValue={formData.user_type}>
                <SelectTrigger>
                    <SelectValue placeholder="Select user type" />
                </SelectTrigger>
                <SelectContent>
                    <SelectItem value="student">Student</SelectItem>
                    <SelectItem value="teacher">Teacher</SelectItem>
                </SelectContent>
            </Select>
            {error && <p className="text-red-500 text-sm text-center">{error}</p>}
            {success && <p className="text-green-500 text-sm text-center">{success}</p>}
          </CardContent>
          <CardFooter className="flex flex-col gap-4">
            <Button className="w-full transition-transform duration-200 hover:scale-105" type="submit" disabled={loading}>
                {loading ? "Registering..." : <>
                    <UserPlus className="mr-2 h-4 w-4" />
                    Register
                </>}
            </Button>
            <Link to="/login" className="text-sm text-primary hover:underline">
                Already have an account? Sign in
            </Link>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
} 