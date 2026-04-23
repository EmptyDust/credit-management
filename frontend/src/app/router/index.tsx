import {
  createBrowserRouter,
  Navigate,
  type RouteObject,
} from "react-router-dom";

import AppLayout from "@/app/layouts/AppLayout";
import AuthLayout from "@/app/layouts/AuthLayout";
import AdminLayout from "@/app/layouts/AdminLayout";
import ProtectedRoute from "@/components/ProtectedRoute";
import RoleBasedRoute from "@/components/RoleBasedRoute";
import ActivityDetail from "@/pages/ActivityDetail";
import Dashboard from "@/pages/Dashboard";
import ApplicationsPage from "@/modules/applications/pages/ApplicationsPage";
import ActivitiesPage from "@/modules/activities/pages/ActivitiesPage";
import ApiTester from "@/pages/devtools/ApiTester";
import LogViewer from "@/pages/devtools/LogViewer";
import Login from "@/pages/Login";
import NotFound from "@/pages/NotFound";
import ProfilePage from "@/pages/Profile";
import Register from "@/pages/Register";
import Students from "@/pages/Students";
import Teachers from "@/pages/Teachers";

function AuditLogsPlaceholder() {
  return (
    <div className="rounded-xl border bg-card p-6 shadow-sm">
      <h1 className="text-xl font-semibold">审计日志</h1>
      <p className="mt-2 text-sm text-muted-foreground">
        审计日志页面会在下一阶段迁移到模块化前端，这里先保留路由入口。
      </p>
    </div>
  );
}

export const appRoutes: RouteObject[] = [
  {
    path: "/",
    element: <Navigate to="/dashboard" replace />,
  },
  {
    element: <AuthLayout />,
    children: [
      {
        path: "/login",
        element: <Login />,
      },
      {
        path: "/register",
        element: <Register />,
      },
    ],
  },
  {
    element: <ProtectedRoute />,
    children: [
      {
        element: <AppLayout />,
        children: [
          {
            path: "/dashboard",
            element: <Dashboard />,
          },
          {
            path: "/students",
            element: (
              <RoleBasedRoute allowedRoles={["teacher", "admin"]}>
                <Students />
              </RoleBasedRoute>
            ),
          },
          {
            path: "/teachers",
            element: (
              <RoleBasedRoute allowedRoles={["admin"]}>
                <Teachers />
              </RoleBasedRoute>
            ),
          },
          {
            path: "/activities",
            element: <ActivitiesPage />,
          },
          {
            path: "/activities/:id",
            element: <ActivityDetail />,
          },
          {
            path: "/applications",
            element: <ApplicationsPage />,
          },
          {
            path: "/profile",
            element: <ProfilePage />,
          },
          {
            path: "/devtools",
            element: <AdminLayout />,
            children: [
              {
                path: "api-tester",
                element: <ApiTester />,
              },
              {
                path: "logs",
                element: <LogViewer />,
              },
              {
                path: "audit-logs",
                element: <AuditLogsPlaceholder />,
              },
            ],
          },
        ],
      },
    ],
  },
  {
    path: "*",
    element: <NotFound />,
  },
];

export const router = createBrowserRouter(appRoutes);
