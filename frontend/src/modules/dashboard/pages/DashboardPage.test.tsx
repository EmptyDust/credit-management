// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import DashboardPage from "@/modules/dashboard/pages/DashboardPage";

const { mockGetDashboardOverview } = vi.hoisted(() => ({
  mockGetDashboardOverview: vi.fn(),
}));

vi.mock("@/modules/dashboard/api/dashboard", () => ({
  getDashboardOverview: mockGetDashboardOverview,
}));

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <DashboardPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

describe("DashboardPage", () => {
  beforeEach(() => {
    mockGetDashboardOverview.mockResolvedValue({
      user_count: 4,
      activity_count: 3,
      application_count: 8,
      user_stats: {
        total: 4,
        student_count: 2,
        teacher_count: 1,
        admin_count: 1,
        active_count: 3,
        disabled_count: 1,
        role_counts: { student: 2, teacher: 1, admin: 1 },
        status_counts: { active: 3, disabled: 1 },
      },
      activity_stats: {
        status_counts: {
          draft: 1,
          pending_review: 0,
          approved: 1,
          rejected: 1,
        },
        category_counts: [
          { key: "general", label: "General", count: 1 },
          { key: "innovation", label: "Innovation", count: 2 },
        ],
        academic_year_counts: [{ value: "2025-2026", count: 3 }],
        semester_counts: [{ value: "autumn", count: 3 }],
        mine_count: 1,
        participated_count: 2,
      },
      application_stats: {
        total: 8,
        approved_count: 6,
        total_credits: 12.5,
      },
      category_distribution: [
        { key: "general", label: "General", count: 1 },
        { key: "innovation", label: "Innovation", count: 2 },
      ],
      recent_activities: [
        {
          id: "activity-1",
          title: "Campus Volunteer Day",
          category: "general",
          academic_year: "2025-2026",
          semester: "autumn",
          status: "approved",
          owner_id: "teacher-1",
          participant_count: 1,
        },
      ],
      popular_activities: [
        {
          id: "activity-2",
          title: "Robotics Workshop",
          category: "innovation",
          academic_year: "2025-2026",
          semester: "autumn",
          status: "draft",
          owner_id: "teacher-2",
          participant_count: 3,
        },
      ],
    });
  });

  afterEach(() => {
    cleanup();
  });

  it("renders dashboard aggregate metrics and activity summaries", async () => {
    renderPage();

    expect(await screen.findByText("Campus Volunteer Day")).toBeTruthy();
    expect(screen.getByText("Robotics Workshop")).toBeTruthy();
    expect(screen.getByText("活动状态")).toBeTruthy();
    expect(screen.getByText("分类分布")).toBeTruthy();
    expect(screen.getByText("最近活动")).toBeTruthy();
    expect(screen.getByText("热门活动")).toBeTruthy();
    expect(screen.getByText("学生 / 教师 / 管理员")).toBeTruthy();
    expect(screen.getByText("12.5")).toBeTruthy();
  });
});
