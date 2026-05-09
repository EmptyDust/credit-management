// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import ApplicationsPage from "@/modules/applications/pages/ApplicationsPage";

const {
  mockExportApplications,
  mockGetApplicationDetail,
  mockGetApplicationStats,
  mockListApplications,
} = vi.hoisted(() => ({
  mockExportApplications: vi.fn(),
  mockGetApplicationDetail: vi.fn(),
  mockGetApplicationStats: vi.fn(),
  mockListApplications: vi.fn(),
}));

vi.mock("@/modules/applications/api/applications", () => ({
  listApplications: mockListApplications,
  getApplicationStats: mockGetApplicationStats,
  getApplicationDetail: mockGetApplicationDetail,
  exportApplications: mockExportApplications,
}));

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });
}

function renderPage() {
  render(
    <QueryClientProvider client={createQueryClient()}>
      <MemoryRouter>
        <ApplicationsPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

describe("ApplicationsPage", () => {
  beforeEach(() => {
    mockListApplications.mockResolvedValue({
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
    });
    mockGetApplicationStats.mockResolvedValue({
      total: 1,
      approved_count: 1,
      total_credits: 2,
    });
    mockGetApplicationDetail.mockResolvedValue({
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
    });
    mockExportApplications.mockResolvedValue(undefined);
  });

  afterEach(() => {
    cleanup();
  });

  it("renders application stats and list rows", async () => {
    renderPage();

    expect(await screen.findByText("Campus Volunteer Day")).toBeTruthy();
    expect(screen.getByText("申请总数")).toBeTruthy();
  });

  it("passes search filters to list stats and export", async () => {
    const user = userEvent.setup();

    renderPage();

    await screen.findByText("Campus Volunteer Day");

    await user.type(screen.getByLabelText("申请搜索关键字"), "Robotics");
    await user.type(screen.getByLabelText("申请学年"), "2025-2026");
    await user.type(screen.getByLabelText("申请学期"), "autumn");

    await waitFor(() => {
      expect(mockListApplications).toHaveBeenLastCalledWith({
        keyword: "Robotics",
        academic_year: "2025-2026",
        semester: "autumn",
      });
      expect(mockGetApplicationStats).toHaveBeenLastCalledWith({
        keyword: "Robotics",
        academic_year: "2025-2026",
        semester: "autumn",
      });
    });

    await user.click(screen.getByRole("button", { name: "导出申请" }));

    expect(mockExportApplications).toHaveBeenCalledWith({
      keyword: "Robotics",
      academic_year: "2025-2026",
      semester: "autumn",
    });
  });
});
