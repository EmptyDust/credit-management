// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

import ActivitiesPage from "@/modules/activities/pages/ActivitiesPage";

const { mockNavigate, mockCreateActivity } = vi.hoisted(() => ({
  mockNavigate: vi.fn(),
  mockCreateActivity: vi.fn().mockResolvedValue({
    id: "activity-2",
    title: "New Draft Activity",
    category: "innovation",
    academic_year: "2025-2026",
    semester: "autumn",
    status: "draft",
    owner_id: "student-1",
  }),
}));

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual<typeof import("react-router-dom")>(
    "react-router-dom"
  );

  return {
    ...actual,
    useNavigate: () => mockNavigate,
  };
});

vi.mock("@/modules/activities/api/activities", () => ({
  listActivities: vi.fn().mockResolvedValue({
    items: [
      {
        id: "activity-1",
        title: "Campus Volunteer Day",
        category: "volunteer",
        academic_year: "2025-2026",
        semester: "autumn",
        status: "approved",
        owner_id: "teacher-1",
      },
    ],
    total: 1,
    page: 1,
    page_size: 10,
    total_pages: 1,
  }),
  createActivity: mockCreateActivity,
}));

describe("ActivitiesPage", () => {
  it("renders activities from the modular API layer", async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <ActivitiesPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(await screen.findByText("Campus Volunteer Day")).toBeTruthy();
    expect(screen.getByText("活动列表")).toBeTruthy();
  });

  it("creates a draft activity from the modular list page", async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <ActivitiesPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    const createButtons = await screen.findAllByRole("button", { name: "新建活动" });
    await userEvent.click(createButtons[0]);
    await userEvent.type(screen.getByLabelText("标题"), "New Draft Activity");
    await userEvent.type(screen.getByLabelText("分类"), "innovation");
    await userEvent.type(screen.getByLabelText("学年"), "2025-2026");
    await userEvent.type(screen.getByLabelText("学期"), "autumn");
    await userEvent.click(screen.getByRole("button", { name: "创建草稿" }));

    await waitFor(() => {
      expect(mockCreateActivity).toHaveBeenCalledWith({
        title: "New Draft Activity",
        description: undefined,
        category: "innovation",
        academic_year: "2025-2026",
        semester: "autumn",
        start_date: undefined,
        end_date: undefined,
        details: {},
      });
    });

    expect(mockNavigate).toHaveBeenCalledWith("/activities/activity-2");
  });
});
