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
import { useAuth } from "@/contexts/AuthContext";
import apiClient from "@/lib/api";
import { useNavigate, Link } from "react-router-dom";
import { LogIn, User, KeyRound } from "lucide-react";

export default function Login() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      // The API Gateway forwards /api/auth/login to the auth-service
      const response = await apiClient.post("/auth/login", { username, password });
      if (response.data && response.data.token) {
        login(response.data.token, response.data.user);
        navigate("/dashboard");
      }
    } catch (err) {
      setError("Failed to login. Please check your credentials.");
      console.error(err);
    } finally {
        setLoading(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen bg-muted/40">
      <Card className="w-full max-w-sm border-0 shadow-lg sm:border">
        <form onSubmit={handleLogin}>
          <CardHeader className="text-center">
            <CardTitle className="text-3xl font-bold flex items-center justify-center gap-2">
              <LogIn />
              Login
            </CardTitle>
            <CardDescription>
              Enter your credentials to access your account.
            </CardDescription>
          </CardHeader>
          <CardContent className="grid gap-4">
            <div className="relative">
              <User className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
              <Input id="username" type="text" placeholder="Username" required value={username} onChange={(e) => setUsername(e.target.value)} className="pl-10" />
            </div>
            <div className="relative">
              <KeyRound className="absolute left-3 top-1/2 -translate-y-1/2 h-5 w-5 text-muted-foreground" />
              <Input id="password" type="password" placeholder="Password" required value={password} onChange={(e) => setPassword(e.target.value)} className="pl-10" />
            </div>
            <div className="flex items-center justify-between -mt-2">
                <Link to="/register" className="text-sm text-primary hover:underline">
                    Create an account
                </Link>
                <Link to="#" className="text-sm text-primary hover:underline">
                    Forgot password?
                </Link>
            </div>
            {error && <p className="text-red-500 text-sm">{error}</p>}
          </CardContent>
          <CardFooter>
            <Button className="w-full transition-transform duration-200 hover:scale-105" type="submit" disabled={loading}>
                {loading ? "Signing in..." : <>
                    <LogIn className="mr-2 h-4 w-4" />
                    Sign in
                </>}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  );
} 