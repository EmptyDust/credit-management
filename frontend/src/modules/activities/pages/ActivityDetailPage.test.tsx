// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { Outlet, RouterProvider, createMemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

import { appRoutes } from "@/app/router";

vi.mock("@/contexts/AuthContext", () => ({
  useAuth: () => ({
    isAuthenticated: true,
    loading: false,
    user: {
      id: "teacher-1",
      uuid: "teacher-1",
      username: "teacher",
      userType: "teacher",
      user_type: "teacher",
      fullName: "Teacher One",
      real_name: "Teacher One",
      status: "active",
    },
  }),
}));

vi.mock("@/app/layouts/AppLayout", () => ({
  default: () => <Outlet />,
}));

vi.mock("@/app/layouts/AuthLayout", () => ({
  default: () => <Outlet />,
}));

vi.mock("@/app/layouts/AdminLayout", () => ({
  default: () => <Outlet />,
}));

vi.mock("@/pages/Login", () => ({
  default: () => <div>Login</div>,
}));

vi.mock("@/pages/Register", () => ({
  default: () => <div>Register</div>,
}));

vi.mock("@/pages/NotFound", () => ({
  default: () => <div>NotFound</div>,
}));

vi.mock("@/components/activity-details/index", () => ({
  ActivityDetailContainer: ({
    activity,
  }: {
    activity: { title?: string; status?: string };
  }) => (
    <div>
      <div>{activity.title}</div>
      <div>{activity.status}</div>
    </div>
  ),
}));

vi.mock("@/components/activity-common", () => ({
  ActivityEditDialog: () => null,
}));

vi.mock("@/components/ui/top-progress-bar", () => ({
  TopProgressBar: () => null,
}));

vi.mock("@/lib/options", () => ({
  getActivityOptions: vi.fn().mockResolvedValue({
    review_actions: [],
  }),
}));

vi.mock("@/lib/api", () => ({
  default: {
    get: vi.fn().mockImplementation((url: string) => {
      if (url === "/activities/activity-1") {
        return Promise.resolve({
          data: {
            data: {
              activity: {
                id: "activity-1",
                title: "Robotics Expo",
                description: "Annual showcase",
                category: "innovation",
                academic_year: "2025-2026",
                semester: "autumn",
                status: "pending_review",
                owner_id: "student-1",
              },
            },
          },
        });
      }

      return Promise.reject(new Error(`unexpected GET ${url}`));
    }),
    post: vi.fn(),
    delete: vi.fn(),
  },
}));

describe("activity detail route", () => {
  it("renders title from the current backend detail contract", async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });

    const router = createMemoryRouter(appRoutes, {
      initialEntries: ["/activities/activity-1"],
    });

    render(
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    );

    expect(await screen.findByText("Robotics Expo")).toBeTruthy();
  });
});
