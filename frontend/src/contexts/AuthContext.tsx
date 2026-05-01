/* eslint-disable react-refresh/only-export-components */

import { createContext, useContext, useEffect, useState } from "react";
import toast from "react-hot-toast";

import apiClient from "@/lib/api";
import {
  clearStoredSession,
  storeTokens,
  USER_STORAGE_KEY,
} from "@/shared/api/http";
import type { SessionUser, UserRole } from "@/shared/types/api";

export interface User extends SessionUser {
  student_id?: string;
  teacher_id?: string;
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
  logout: () => Promise<void>;
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
  logout: async () => {},
  refreshUser: async () => {},
  hasPermission: () => false,
  hasRole: () => false,
  updateUser: () => {},
});

function normalizeUser(userData: Partial<User> | null | undefined): User | null {
  if (!userData) {
    return null;
  }

  const role = (userData.userType ||
    userData.user_type ||
    "student") as UserRole;
  const id = userData.id || userData.uuid || "";
  if (!id) {
    return null;
  }

  return {
    ...userData,
    id,
    uuid: id,
    userType: role,
    user_type: role,
    fullName: userData.fullName || userData.real_name || "",
    real_name: userData.real_name || userData.fullName || "",
    avatar: userData.avatar || userData.avatar_url,
    avatar_url: userData.avatar_url || userData.avatar,
    status: (userData.status || "active") as User["status"],
  };
}

function storeUser(user: User | null) {
  if (!user) {
    localStorage.removeItem(USER_STORAGE_KEY);
    return;
  }

  localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(user));
}

function loadStoredUser() {
  const raw = localStorage.getItem(USER_STORAGE_KEY);
  if (!raw) {
    return null;
  }

  try {
    return normalizeUser(JSON.parse(raw));
  } catch {
    localStorage.removeItem(USER_STORAGE_KEY);
    return null;
  }
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const initializeAuth = async () => {
      const token = localStorage.getItem("token");
      const storedUser = loadStoredUser();

      if (!token) {
        setUser(null);
        setIsAuthenticated(false);
        setLoading(false);
        return;
      }

      if (storedUser) {
        setUser(storedUser);
        setIsAuthenticated(true);
      }

      try {
        const response = await apiClient.get("/me/profile");
        const normalizedUser = normalizeUser(response.data?.data);

        if (!normalizedUser) {
          throw new Error("invalid profile payload");
        }

        setUser(normalizedUser);
        setIsAuthenticated(true);
        storeUser(normalizedUser);
      } catch {
        clearStoredSession();
        setUser(null);
        setIsAuthenticated(false);
      } finally {
        setLoading(false);
      }
    };

    void initializeAuth();
  }, []);

  const login = (token: string, refreshToken: string, nextUser: User) => {
    const normalizedUser = normalizeUser(nextUser);
    if (!normalizedUser) {
      return;
    }

    storeTokens(token, refreshToken);
    storeUser(normalizedUser);
    setIsAuthenticated(true);
    setUser(normalizedUser);
    toast.success(`欢迎回来，${normalizedUser.fullName || normalizedUser.username}！`);
  };

  const logout = async () => {
    try {
      await apiClient.post("/auth/logout");
    } catch (error) {
      console.error("Logout error:", error);
    } finally {
      clearStoredSession();
      setIsAuthenticated(false);
      setUser(null);
      toast.success("已成功退出登录");
    }
  };

  const refreshUser = async () => {
    try {
      const response = await apiClient.get("/me/profile");
      const normalizedUser = normalizeUser(response.data?.data);
      if (!normalizedUser) {
        throw new Error("invalid profile payload");
      }

      setUser(normalizedUser);
      storeUser(normalizedUser);
    } catch (error) {
      console.error("Failed to refresh user data:", error);
      clearStoredSession();
      setIsAuthenticated(false);
      setUser(null);
    }
  };

  const updateUser = (userData: Partial<User>) => {
    const normalizedUser = normalizeUser({ ...(user ?? {}), ...userData });
    if (!normalizedUser) {
      return;
    }

    setUser(normalizedUser);
    storeUser(normalizedUser);
  };

  const hasPermission = (permission: string): boolean => {
    if (!user) {
      return false;
    }

    if (user.userType === "admin") {
      return true;
    }

    const rolePermissions: Record<UserRole, string[]> = {
      student: [],
      teacher: ["manage_activities", "view_user_stats", "view_students"],
      admin: ["*"],
    };

    const permissions = rolePermissions[user.userType] || [];
    return permissions.includes("*") || permissions.includes(permission);
  };

  const hasRole = (role: string) => {
    if (!user) {
      return false;
    }

    return user.userType === role || user.roles?.includes(role) === true;
  };

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        user,
        loading,
        login,
        logout,
        refreshUser,
        hasPermission,
        hasRole,
        updateUser,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
