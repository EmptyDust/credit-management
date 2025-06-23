import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { useAuth } from "@/contexts/AuthContext";
import apiClient from "@/lib/api";
import { User, Mail, Phone, FileSignature, Edit3, Save } from "lucide-react";

type UserProfile = {
  email: string;
  phone: string;
  real_name: string;
};

export default function ProfilePage() {
  const { user } = useAuth();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  useEffect(() => {
    const fetchProfile = async () => {
      if (!user) return;
      try {
        setLoading(true);
        const response = await apiClient.get(`/users/${user.username}`);
        setProfile(response.data);
      } catch (err) {
        setError("Failed to fetch profile.");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    fetchProfile();
  }, [user]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (profile) {
      setProfile({ ...profile, [e.target.id]: e.target.value });
    }
  };

  const handleSave = async () => {
    if (!profile) return;
    setError("");
    setSuccess("");
    try {
        await apiClient.put(`/users/profile`, profile);
        setSuccess("Profile updated successfully!");
        setIsEditing(false);
    } catch (err) {
        setError("Failed to update profile.");
        console.error(err);
    }
  };

  if (loading) return <div>Loading...</div>;
  if (error) return <div className="text-red-500">{error}</div>;

  return (
    <div className="space-y-6">
        <h1 className="text-3xl font-bold">My Profile</h1>
        <Card>
            <CardHeader className="flex flex-row items-center justify-between">
                <div>
                    <CardTitle>Personal Information</CardTitle>
                    <CardDescription>View and edit your personal details.</CardDescription>
                </div>
                {!isEditing ? (
                    <Button onClick={() => setIsEditing(true)} variant="outline">
                        <Edit3 className="mr-2 h-4 w-4" /> Edit
                    </Button>
                ) : (
                    <Button onClick={handleSave}>
                        <Save className="mr-2 h-4 w-4" /> Save
                    </Button>
                )}
            </CardHeader>
            <CardContent className="space-y-4">
                <div className="flex items-center gap-4">
                    <User className="h-5 w-5 text-muted-foreground" />
                    <Input id="username" value={user?.username ?? ""} disabled />
                </div>
                <div className="flex items-center gap-4">
                    <Mail className="h-5 w-5 text-muted-foreground" />
                    <Input id="email" value={profile?.email ?? ""} onChange={handleChange} disabled={!isEditing} />
                </div>
                <div className="flex items-center gap-4">
                    <Phone className="h-5 w-5 text-muted-foreground" />
                    <Input id="phone" value={profile?.phone ?? ""} onChange={handleChange} disabled={!isEditing} />
                </div>
                <div className="flex items-center gap-4">
                    <FileSignature className="h-5 w-5 text-muted-foreground" />
                    <Input id="real_name" value={profile?.real_name ?? ""} onChange={handleChange} disabled={!isEditing} />
                </div>
                {success && <p className="text-green-500 text-sm">{success}</p>}
                {error && <p className="text-red-500 text-sm">{error}</p>}
            </CardContent>
        </Card>
    </div>
  );
} 