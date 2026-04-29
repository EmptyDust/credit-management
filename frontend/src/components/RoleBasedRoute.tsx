import type { ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '@/contexts/AuthContext';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { AlertTriangle } from 'lucide-react';

interface RoleBasedRouteProps {
  children: ReactNode;
  allowedRoles: string[];
  fallbackPath?: string;
}

export default function RoleBasedRoute({ 
  children, 
  allowedRoles}: RoleBasedRouteProps) {
  const { user, isAuthenticated } = useAuth();

  // 如果未认证，重定向到登录页
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  // 如果用户角色不在允许的角色列表中
  if (!user || !allowedRoles.includes(user.userType)) {
    return (
      <div className="flex-1 flex items-center justify-center p-8">
        <Card className="w-full max-w-md">
          <CardHeader className="text-center">
            <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-full bg-red-100 dark:bg-red-900/20">
              <AlertTriangle className="h-6 w-6 text-red-600" />
            </div>
            <CardTitle className="text-xl">权限不足</CardTitle>
          </CardHeader>
          <CardContent className="text-center">
            <p className="text-muted-foreground mb-4">
              您没有权限访问此页面。当前用户类型：{user?.userType}
            </p>
            <p className="text-sm text-muted-foreground">
              需要的权限：{allowedRoles.join(' 或 ')}
            </p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return <>{children}</>;
} 