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
import ApplicationsPage from "@/modules/applications/pages/ApplicationsPage";
import ActivitiesPage from "@/modules/activities/pages/ActivitiesPage";
import DashboardPage from "@/modules/dashboard/pages/DashboardPage";
import ApiTesterPage from "@/modules/devtools/pages/ApiTesterPage";
import AuditLogsPage from "@/modules/devtools/pages/AuditLogsPage";
import DevtoolsLogsPage from "@/modules/devtools/pages/DevtoolsLogsPage";
import DevtoolsOverviewPage from "@/modules/devtools/pages/DevtoolsOverviewPage";
import ProfilePage from "@/modules/users/pages/ProfilePage";
import StudentsPage from "@/modules/users/pages/StudentsPage";
import TeachersPage from "@/modules/users/pages/TeachersPage";
import Login from "@/pages/Login";
import NotFound from "@/pages/NotFound";
import Register from "@/pages/Register";

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
                index: true,
                element: <DevtoolsOverviewPage />,
              },
              {
                path: "api-tester",
                element: <ApiTesterPage />,
              },
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

export const router = createBrowserRouter(appRoutes);
