import {
  createBrowserRouter,
  Navigate,
  type RouteObject,
} from "react-router-dom";

import { isApiTesterEnabled } from "@/app/config/devtools";
import AppLayout from "@/app/layouts/AppLayout";
import AuthLayout from "@/app/layouts/AuthLayout";
import AdminLayout from "@/app/layouts/AdminLayout";
import ProtectedRoute from "@/components/ProtectedRoute";
import RoleBasedRoute from "@/components/RoleBasedRoute";
import ActivityDetailPage from "@/modules/activities/pages/ActivityDetailPage";
import ActivityCreatePage from "@/modules/activities/pages/ActivityCreatePage";
import ActivityEditPage from "@/modules/activities/pages/ActivityEditPage";
import ApplicationsPage from "@/modules/applications/pages/ApplicationsPage";
import ActivitiesPage from "@/modules/activities/pages/ActivitiesPage";
import LoginPage from "@/modules/auth/pages/LoginPage";
import RegisterPage from "@/modules/auth/pages/RegisterPage";
import DashboardPage from "@/modules/dashboard/pages/DashboardPage";
import ApiTesterPage from "@/modules/devtools/pages/ApiTesterPage";
import AuditLogsPage from "@/modules/devtools/pages/AuditLogsPage";
import DevtoolsLogsPage from "@/modules/devtools/pages/DevtoolsLogsPage";
import DevtoolsOverviewPage from "@/modules/devtools/pages/DevtoolsOverviewPage";
import ProfilePage from "@/modules/users/pages/ProfilePage";
import StudentsPage from "@/modules/users/pages/StudentsPage";
import TeachersPage from "@/modules/users/pages/TeachersPage";
import NotFound from "@/pages/NotFound";

type AppRoutesOptions = {
  enableApiTester?: boolean;
};

export function createAppRoutes(options: AppRoutesOptions = {}): RouteObject[] {
  const enableApiTester = options.enableApiTester ?? false;

  return [
  {
    path: "/",
    element: <Navigate to="/dashboard" replace />,
  },
  {
    element: <AuthLayout />,
    children: [
      {
        path: "/login",
        element: <LoginPage />,
      },
      {
        path: "/register",
        element: <RegisterPage />,
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
            element: <DashboardPage />,
          },
          {
            path: "/students",
            element: (
              <RoleBasedRoute allowedRoles={["teacher", "admin"]}>
                <StudentsPage />
              </RoleBasedRoute>
            ),
          },
          {
            path: "/teachers",
            element: (
              <RoleBasedRoute allowedRoles={["admin"]}>
                <TeachersPage />
              </RoleBasedRoute>
            ),
          },
          {
            path: "/activities",
            element: <ActivitiesPage />,
          },
          {
            path: "/activities/new",
            element: <ActivityCreatePage />,
          },
          {
            path: "/activities/:id/edit",
            element: <ActivityEditPage />,
          },
          {
            path: "/activities/:id",
            element: <ActivityDetailPage />,
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
                index: true,
                element: <DevtoolsOverviewPage />,
              },
              ...(enableApiTester
                ? [
                    {
                      path: "api-tester",
                      element: <ApiTesterPage />,
                    },
                  ]
                : []),
              {
                path: "logs",
                element: <DevtoolsLogsPage />,
              },
              {
                path: "audit-logs",
                element: <AuditLogsPage />,
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
}

export const appRoutes: RouteObject[] = createAppRoutes({
  enableApiTester: isApiTesterEnabled(),
});

export const router = createBrowserRouter(appRoutes);
