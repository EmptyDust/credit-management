// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import ActivitiesPage from "@/modules/activities/pages/ActivitiesPage";

const {
  mockConfirmActivityImport,
  mockExportActivities,
  mockGetActivityStats,
  mockListActivities,
  mockNavigate,
  mockPrecheckActivityImport,
} = vi.hoisted(() => ({
  mockConfirmActivityImport: vi.fn().mockResolvedValue({
    job_id: "job-1",
    stage: "completed",
    summary: {
      total_rows: 1,
      valid_rows: 1,
      invalid_rows: 0,
    },
    created_activity_ids: ["activity-2"],
  }),
  mockExportActivities: vi.fn().mockResolvedValue(undefined),
  mockGetActivityStats: vi.fn().mockResolvedValue({
    status_counts: {
      draft: 1,
      pending_review: 1,
      approved: 1,
      rejected: 0,
    },
    category_counts: [
      { key: "volunteer", label: "志愿活动", count: 1 },
      { key: "innovation", label: "创新活动", count: 2 },
    ],
    academic_year_counts: [{ value: "2025-2026", count: 3 }],
    semester_counts: [{ value: "autumn", count: 3 }],
    mine_count: 1,
    participated_count: 2,
  }),
  mockListActivities: vi.fn().mockResolvedValue({
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
  mockNavigate: vi.fn(),
  mockPrecheckActivityImport: vi.fn().mockResolvedValue({
    job_id: "job-1",
    job_type: "activities",
    file_name: "activities.csv",
    file_hash: "hash-1",
    expires_at: "2026-05-10T00:00:00Z",
    summary: {
      total_rows: 1,
      valid_rows: 1,
      invalid_rows: 0,
    },
    rows: [
      {
        row_no: 1,
        status: "valid",
        payload: {
          title: "Imported Activity",
          category: "innovation",
          academic_year: "2025-2026",
          semester: "autumn",
        },
        errors: [],
      },
    ],
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
  listActivities: mockListActivities,
}));

vi.mock("@/modules/activities/api/activity-stats", () => ({
  exportActivities: mockExportActivities,
  getActivityStats: mockGetActivityStats,
}));

vi.mock("@/modules/activities/api/activity-imports", () => ({
  confirmActivityImport: mockConfirmActivityImport,
  precheckActivityImport: mockPrecheckActivityImport,
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
        <ActivitiesPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

describe("ActivitiesPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    cleanup();
  });

  it("renders workspace stats and activities from the modular API layer", async () => {
    renderPage();

    expect(await screen.findByText("Campus Volunteer Day")).toBeTruthy();
    expect(screen.getByText("活动总数")).toBeTruthy();
    expect(screen.getByText("3")).toBeTruthy();
    expect(screen.getByText("我的活动")).toBeTruthy();
    expect(screen.getByText("已参与")).toBeTruthy();
    expect(mockGetActivityStats).toHaveBeenCalled();
  });

  it("routes the new activity button to the schema-aware create page", async () => {
    renderPage();

    await screen.findByText("Campus Volunteer Day");
    await userEvent.click(screen.getByRole("button", { name: "新建活动" }));

    expect(mockNavigate).toHaveBeenCalledWith("/activities/new");
  });

  it("calls the export endpoint with the current filters", async () => {
    renderPage();

    await screen.findByText("Campus Volunteer Day");
    await userEvent.click(screen.getByRole("button", { name: "导出" }));

    await waitFor(() => {
      expect(mockExportActivities).toHaveBeenCalledWith({
        keyword: undefined,
        status: undefined,
        category: undefined,
        academic_year: undefined,
        semester: undefined,
      });
    });
  });

  it("prechecks and confirms an activity import file", async () => {
    renderPage();

    await screen.findByText("Campus Volunteer Day");
    await userEvent.click(screen.getByRole("button", { name: "导入" }));

    const file = new File(["title,category"], "activities.csv", {
      type: "text/csv",
    });
    await userEvent.upload(screen.getByLabelText("选择文件"), file);
    await userEvent.click(screen.getByRole("button", { name: "预校验" }));

    expect(await screen.findByText("预校验完成")).toBeTruthy();
    expect(screen.getByText("有效行")).toBeTruthy();

    await userEvent.click(screen.getByRole("button", { name: "确认导入" }));

    await waitFor(() => {
      expect(mockPrecheckActivityImport).toHaveBeenCalledWith(file);
      expect(mockConfirmActivityImport).toHaveBeenCalledWith({
        job_id: "job-1",
        file_hash: "hash-1",
      });
    });
  });
});
