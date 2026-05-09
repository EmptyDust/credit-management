// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { RouterProvider, createMemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import ActivityDetailPage from "@/modules/activities/pages/ActivityDetailPage";

const {
  mockAddActivityParticipants,
  mockBatchDeleteActivityParticipants,
  mockBatchSetActivityParticipantCredits,
  mockCopyActivity,
  mockDeleteActivity,
  mockDeleteActivityFile,
  mockDownloadActivityFile,
  mockExportActivityParticipants,
  mockGetActivityDetail,
  mockGetActivityParticipantStats,
  mockLeaveActivity,
  mockListActivityFiles,
  mockListActivityParticipants,
  mockReviewActivity,
  mockSubmitActivity,
  mockTransferActivityOwner,
  mockUpdateActivity,
  mockUpdateActivityParticipantCredits,
  mockUploadActivityFile,
  mockWithdrawActivity,
} = vi.hoisted(() => ({
  mockAddActivityParticipants: vi.fn(),
  mockBatchDeleteActivityParticipants: vi.fn(),
  mockBatchSetActivityParticipantCredits: vi.fn(),
  mockCopyActivity: vi.fn(),
  mockDeleteActivity: vi.fn(),
  mockDeleteActivityFile: vi.fn(),
  mockDownloadActivityFile: vi.fn(),
  mockExportActivityParticipants: vi.fn(),
  mockGetActivityDetail: vi.fn(),
  mockGetActivityParticipantStats: vi.fn(),
  mockLeaveActivity: vi.fn(),
  mockListActivityFiles: vi.fn(),
  mockListActivityParticipants: vi.fn(),
  mockReviewActivity: vi.fn(),
  mockSubmitActivity: vi.fn(),
  mockTransferActivityOwner: vi.fn(),
  mockUpdateActivity: vi.fn(),
  mockUpdateActivityParticipantCredits: vi.fn(),
  mockUploadActivityFile: vi.fn(),
  mockWithdrawActivity: vi.fn(),
}));

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

vi.mock("react-hot-toast", () => ({
  default: {
    error: vi.fn(),
    success: vi.fn(),
  },
}));

vi.mock("@/modules/activities/api/activities", () => ({
  addActivityParticipants: mockAddActivityParticipants,
  batchDeleteActivityParticipants: mockBatchDeleteActivityParticipants,
  batchSetActivityParticipantCredits: mockBatchSetActivityParticipantCredits,
  copyActivity: mockCopyActivity,
  deleteActivity: mockDeleteActivity,
  deleteActivityFile: mockDeleteActivityFile,
  downloadActivityFile: mockDownloadActivityFile,
  exportActivityParticipants: mockExportActivityParticipants,
  getActivityDetail: mockGetActivityDetail,
  getActivityParticipantStats: mockGetActivityParticipantStats,
  leaveActivity: mockLeaveActivity,
  listActivityFiles: mockListActivityFiles,
  listActivityParticipants: mockListActivityParticipants,
  reviewActivity: mockReviewActivity,
  submitActivity: mockSubmitActivity,
  transferActivityOwner: mockTransferActivityOwner,
  updateActivity: mockUpdateActivity,
  updateActivityParticipantCredits: mockUpdateActivityParticipantCredits,
  uploadActivityFile: mockUploadActivityFile,
  withdrawActivity: mockWithdrawActivity,
}));

function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });
}

function renderPage() {
  const router = createMemoryRouter(
    [
      {
        path: "/activities/:id",
        element: <ActivityDetailPage />,
      },
      {
        path: "/activities",
        element: <div>Activities</div>,
      },
    ],
    {
      initialEntries: ["/activities/activity-1"],
    }
  );

  render(
    <QueryClientProvider client={createQueryClient()}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  );
}

describe("ActivityDetailPage", () => {
  beforeEach(() => {
    mockGetActivityDetail.mockResolvedValue({
      id: "activity-1",
      title: "Robotics Expo",
      description: "Annual showcase",
      category: "innovation",
      academic_year: "2025-2026",
      semester: "autumn",
      status: "draft",
      owner_id: "student-1",
      details: {
        location: "Innovation Lab",
      },
    });
    mockListActivityParticipants.mockResolvedValue([
      {
        id: "participant-1",
        activity_id: "activity-1",
        user_id: "student-1",
        credits: 1.5,
        joined_at: "2026-05-01T00:00:00Z",
      },
    ]);
    mockGetActivityParticipantStats.mockResolvedValue({
      total_count: 1,
      total_credits: 1.5,
      average_credits: 1.5,
      recent_new_count: 1,
    });
    mockListActivityFiles.mockResolvedValue([]);
    mockUpdateActivityParticipantCredits.mockResolvedValue({
      id: "participant-1",
      credits: 4.5,
    });
    mockBatchSetActivityParticipantCredits.mockResolvedValue({ affected_count: 1 });
    mockBatchDeleteActivityParticipants.mockResolvedValue({ affected_count: 1 });
    mockExportActivityParticipants.mockResolvedValue(undefined);
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("renders title from the current backend detail contract", async () => {
    renderPage();

    expect(await screen.findByText("Robotics Expo")).toBeTruthy();
  });

  it("manages participant stats export credits and batch operations", async () => {
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText("Robotics Expo")).toBeTruthy();
    expect(await screen.findByText("参与者总数")).toBeTruthy();
    expect(screen.getByText("1.5 学分")).toBeTruthy();
    expect(screen.getByLabelText("参与者 student-1 学分")).toBeTruthy();

    await user.click(screen.getByRole("button", { name: "导出参与者" }));
    expect(mockExportActivityParticipants).toHaveBeenCalledWith("activity-1");

    await user.clear(screen.getByLabelText("参与者 student-1 学分"));
    await user.type(screen.getByLabelText("参与者 student-1 学分"), "4.5");
    await user.click(screen.getByRole("button", { name: "保存 student-1 学分" }));

    await waitFor(() => {
      expect(mockUpdateActivityParticipantCredits).toHaveBeenCalledWith(
        "activity-1",
        "participant-1",
        4.5
      );
    });

    await user.click(screen.getByLabelText("选择参与者 student-1"));
    await user.clear(screen.getByLabelText("批量设置学分"));
    await user.type(screen.getByLabelText("批量设置学分"), "2");
    await user.click(screen.getByRole("button", { name: "批量设置所选学分" }));

    await waitFor(() => {
      expect(mockBatchSetActivityParticipantCredits).toHaveBeenCalledWith(
        "activity-1",
        ["participant-1"],
        2
      );
    });

    await user.click(screen.getByRole("button", { name: "批量删除所选参与者" }));

    await waitFor(() => {
      expect(mockBatchDeleteActivityParticipants).toHaveBeenCalledWith(
        "activity-1",
        ["participant-1"]
      );
    });
  });
});
