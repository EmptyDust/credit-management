// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

import ApplicationsPage from "@/modules/applications/pages/ApplicationsPage";

vi.mock("@/modules/applications/api/applications", () => ({
  listApplications: vi.fn().mockResolvedValue({
    items: [
      {
        id: "application-1",
        status: "approved",
        applied_credits: 2,
        awarded_credits: 2,
        academic_year: "2025-2026",
        semester: "autumn",
        activity: {
          id: "activity-1",
          title: "Campus Volunteer Day",
          category: "volunteer",
        },
      },
    ],
    total: 1,
    page: 1,
    page_size: 10,
    total_pages: 1,
  }),
  getApplicationStats: vi.fn().mockResolvedValue({
    total: 1,
    approved_count: 1,
    total_credits: 2,
  }),
  getApplicationDetail: vi.fn().mockResolvedValue({
    id: "application-1",
    status: "approved",
    applied_credits: 2,
    awarded_credits: 2,
    academic_year: "2025-2026",
    semester: "autumn",
    activity: {
      id: "activity-1",
      title: "Campus Volunteer Day",
      category: "volunteer",
    },
  }),
}));

describe("ApplicationsPage", () => {
  it("renders application stats and list rows", async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
      },
    });

    render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter>
          <ApplicationsPage />
        </MemoryRouter>
      </QueryClientProvider>
    );

    expect(await screen.findByText("Campus Volunteer Day")).toBeTruthy();
    expect(screen.getByText("申请总数")).toBeTruthy();
  });
});
