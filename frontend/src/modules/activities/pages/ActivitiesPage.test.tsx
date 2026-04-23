// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";

import ActivitiesPage from "@/modules/activities/pages/ActivitiesPage";

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
});
