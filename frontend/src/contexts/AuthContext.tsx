import { createContext, useContext, useState, useEffect } from "react";
import apiClient from "@/lib/api";
import toast from "react-hot-toast";

// Define a comprehensive user interface
interface User {
  id: string;
  username: string;
  userType: 'student' | 'teacher' | 'admin';
  email?: string;
  fullName?: string;
  studentNumber?: string; // For students
  teacherId?: string; // For teachers
  department?: string; // For teachers
  college?: string; // For students
  major?: string; // For students
  class?: string; // For students
  status: 'active' | 'inactive';
  createdAt: string;
  updatedAt: string;
  permissions?: string[];
  roles?: string[];
}

type AuthProviderProps = {
  children: React.ReactNode;
};

type AuthContextType = {
  isAuthenticated: boolean;
  user: User | null;
  loading: boolean;
  login: (token: string, refreshToken: string, user: User) => void;
  logout: () => void;
  refreshUser: () => Promise<void>;
  hasPermission: (permission: string) => boolean;
  hasRole: (role: string) => boolean;
  updateUser: (userData: Partial<User>) => void;
};

const AuthContext = createContext<AuthContextType>({
  isAuthenticated: false,
  user: null,
  loading: true,
  login: () => {},
  logout: () => {},
  refreshUser: async () => {},
  hasPermission: () => false,
  hasRole: () => false,
  updateUser: () => {},
});

export function AuthProvider({ children }: AuthProviderProps) {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  // Check token validity and get user info on mount
  useEffect(() => {
    const initializeAuth = async () => {
      const token = localStorage.getItem("token");
      if (token) {
        try {
          // Validate token and get user info
          const response = await apiClient.get("/users/profile");
          // 兼容后端返回格式
          const userData = response.data.data || response.data;
          const normalizedUser = {
            ...userData,
            userType: userData.userType || userData.user_type,
          };
          setIsAuthenticated(true);
          setUser(normalizedUser);
        } catch (error) {
          // Token is invalid, clear storage
          localStorage.removeItem("token");
          localStorage.removeItem("refreshToken");
          localStorage.removeItem("user");
          setIsAuthenticated(false);
          setUser(null);
        }
      }
      setLoading(false);
    };

    initializeAuth();
  }, []);

  const login = (token: string, refreshToken: string, user: User) => {
    localStorage.setItem("token", token);
    localStorage.setItem("refreshToken", refreshToken);
    localStorage.setItem("user", JSON.stringify(user));
    apiClient.defaults.headers.common['Authorization'] = `Bearer ${token}`;
    setIsAuthenticated(true);
    setUser(user);
    toast.success(`欢迎回来，${user.fullName || user.username}！`);
  };

  const logout = async () => {
    try {
      // Call logout endpoint to invalidate token on server
      await apiClient.post("/auth/logout");
    } catch (error) {
      // Ignore logout errors
      console.error("Logout error:", error);
    } finally {
      // Clear local storage and state
      localStorage.removeItem("token");
      localStorage.removeItem("refreshToken");
      localStorage.removeItem("user");
      delete apiClient.defaults.headers.common['Authorization'];
      setIsAuthenticated(false);
      setUser(null);
      toast.success("已成功退出登录");
    }
  };

  const refreshUser = async () => {
    try {
      const response = await apiClient.get("/users/profile");
      const userData = response.data.data || response.data;
      setUser({
        ...userData,
        userType: userData.userType || userData.user_type,
      });
      localStorage.setItem("user", JSON.stringify(userData));
    } catch (error) {
      console.error("Failed to refresh user data:", error);
      // If refresh fails, logout user
      logout();
    }
  };

  const updateUser = (userData: Partial<User>) => {
    if (user) {
      const updatedUser = { ...user, ...userData };
      setUser(updatedUser);
      localStorage.setItem("user", JSON.stringify(updatedUser));
    }
  };

  const hasPermission = (permission: string): boolean => {
    if (!user) return false;
    
    // Admin has all permissions
    if (user.userType === 'admin') return true;
    
    // Check user's explicit permissions
    if (user.permissions && user.permissions.includes(permission)) {
      return true;
    }
    
    // Check role-based permissions
    if (user.roles) {
      const rolePermissions: Record<string, string[]> = {
        'student': ['view_own_applications', 'create_application', 'view_own_profile', 'edit_own_profile'],
        'teacher': ['view_applications', 'review_applications', 'view_students', 'view_own_profile', 'edit_own_profile'],
        'admin': ['*'] // All permissions
      };
      
      for (const role of user.roles) {
        const permissions = rolePermissions[role] || [];
        if (permissions.includes('*') || permissions.includes(permission)) {
          return true;
        }
      }
    }
    
    // Fallback to user type permissions
    const userTypePermissions: Record<string, string[]> = {
      'student': ['view_own_applications', 'create_application', 'view_own_profile', 'edit_own_profile'],
      'teacher': ['view_applications', 'review_applications', 'view_students', 'view_own_profile', 'edit_own_profile'],
      'admin': ['*'] // All permissions
    };
    
    const permissions = userTypePermissions[user.userType] || [];
    return permissions.includes('*') || permissions.includes(permission);
  };

  const hasRole = (role: string): boolean => {
    if (!user) return false;
    
    // Check explicit roles
    if (user.roles && user.roles.includes(role)) {
      return true;
    }
    
    // Check user type as role
    return user.userType === role;
  };

  return (
    <AuthContext.Provider value={{ 
      isAuthenticated, 
      user, 
      loading,
      login, 
      logout, 
      refreshUser,
      hasPermission,
      hasRole,
      updateUser
    }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => {
  return useContext(AuthContext);
}; 