import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { api } from '../lib/api';
import toast from 'react-hot-toast';

interface User {
    username: string;
    user_type: string;
    register_time: string;
    created_at: string;
    updated_at: string;
}

interface AuthContextType {
    user: User | null;
    token: string | null;
    login: (username: string, password: string) => Promise<boolean>;
    register: (username: string, password: string, userType: string) => Promise<boolean>;
    logout: () => void;
    isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};

interface AuthProviderProps {
    children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
    const [user, setUser] = useState<User | null>(null);
    const [token, setToken] = useState<string | null>(localStorage.getItem('token'));
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const initAuth = async () => {
            const storedToken = localStorage.getItem('token');
            if (storedToken) {
                try {
                    const response = await api.post('/users/validate-token', {}, {
                        headers: { Authorization: `Bearer ${storedToken}` }
                    });

                    if (response.data.valid) {
                        setToken(storedToken);
                        // 获取用户信息
                        const userResponse = await api.get(`/users/${response.data.username}`, {
                            headers: { Authorization: `Bearer ${storedToken}` }
                        });
                        setUser(userResponse.data);
                    } else {
                        localStorage.removeItem('token');
                        setToken(null);
                    }
                } catch (error) {
                    console.error('Token validation failed:', error);
                    localStorage.removeItem('token');
                    setToken(null);
                }
            }
            setIsLoading(false);
        };

        initAuth();
    }, []);

    const login = async (username: string, password: string): Promise<boolean> => {
        try {
            const response = await api.post('/users/login', { username, password });
            const { token: newToken, user: userData } = response.data;

            localStorage.setItem('token', newToken);
            setToken(newToken);
            setUser(userData);

            toast.success('登录成功！');
            return true;
        } catch (error: any) {
            const message = error.response?.data?.error || '登录失败，请检查用户名和密码';
            toast.error(message);
            return false;
        }
    };

    const register = async (username: string, password: string, userType: string): Promise<boolean> => {
        try {
            const response = await api.post('/users/register', { username, password, user_type: userType });
            const { user: userData } = response.data;

            toast.success('注册成功！请登录');
            return true;
        } catch (error: any) {
            const message = error.response?.data?.error || '注册失败，请稍后重试';
            toast.error(message);
            return false;
        }
    };

    const logout = () => {
        localStorage.removeItem('token');
        setToken(null);
        setUser(null);
        toast.success('已退出登录');
    };

    const value: AuthContextType = {
        user,
        token,
        login,
        register,
        logout,
        isLoading,
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
}; 