// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { memoryRouterFuture } from "@/app/router/future";
import ActivityEditPage from "@/modules/activities/pages/ActivityEditPage";

const {
  mockGetActivityCategorySchema,
  mockGetActivityDetail,
  mockNavigate,
  mockSubmitActivity,
  mockUpdateActivity,
} = vi.hoisted(() => ({
  mockGetActivityCategorySchema: vi.fn().mockResolvedValue({
    category: { key: "innovation", label: "创新活动" },
    schema: {
      version: 1,
      fields: [
        {
          key: "location",
          label: "活动地点",
          type: "text",
          required: true,
          display_order: 1,
        },
      ],
    },
  }),
  mockGetActivityDetail: vi.fn().mockResolvedValue({
    id: "activity-1",
    title: "机器人展示活动",
    description: "年度机器人展示",
    category: "innovation",
    academic_year: "2025-2026",
    semester: "autumn",
    status: "draft",
    owner_id: "student-1",
    details: {
      location: "A101",
    },
  }),
  mockNavigate: vi.fn(),
  mockSubmitActivity: vi.fn().mockResolvedValue({
    id: "activity-1",
    status: "pending_review",
  }),
  mockUpdateActivity: vi.fn().mockResolvedValue({
    id: "activity-1",
    title: "机器人展示活动",
    category: "innovation",
    academic_year: "2025-2026",
    semester: "autumn",
    status: "draft",
    owner_id: "student-1",
  }),
}));

const draftActivity = {
  id: "activity-1",
  title: "机器人展示活动",
  description: "年度机器人展示",
  category: "innovation",
  academic_year: "2025-2026",
  semester: "autumn",
  status: "draft",
  owner_id: "student-1",
  details: {
    location: "A101",
  },
};

vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual<typeof import("react-router-dom")>(
    "react-router-dom"
  );

  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ id: "activity-1" }),
  };
});

vi.mock("@/modules/activities/api/activity-categories", () => ({
  getActivityCategorySchema: mockGetActivityCategorySchema,
}));

vi.mock("@/modules/activities/api/activities", () => ({
  getActivityDetail: mockGetActivityDetail,
  submitActivity: mockSubmitActivity,
  updateActivity: mockUpdateActivity,
}));

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter future={memoryRouterFuture}>
        <ActivityEditPage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

describe("ActivityEditPage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockGetActivityDetail.mockResolvedValue(draftActivity);
    mockSubmitActivity.mockResolvedValue({
      ...draftActivity,
      status: "pending_review",
    });
    mockUpdateActivity.mockResolvedValue({
      ...draftActivity,
      details: {
        location: "B202",
      },
    });
  });

  afterEach(() => {
    cleanup();
  });

  it("hydrates fixed and schema-driven fields, saves the draft, and can submit for review", async () => {
    renderPage();

    expect(await screen.findByRole("heading", { name: "编辑活动" })).toBeTruthy();
    expect((screen.getByLabelText("活动标题") as HTMLInputElement).value).toBe(
      "机器人展示活动"
    );
    const locationInput = (await screen.findByLabelText(
      /活动地点/
    )) as HTMLInputElement;
    expect(locationInput.value).toBe("A101");
    expect(mockGetActivityCategorySchema).toHaveBeenCalledWith("innovation");

    await userEvent.clear(screen.getByLabelText(/活动地点/));
    await userEvent.type(screen.getByLabelText(/活动地点/), "B202");
    await userEvent.click(screen.getByRole("button", { name: "保存修改" }));

    await waitFor(() => {
      expect(mockUpdateActivity).toHaveBeenCalledWith("activity-1", {
        title: "机器人展示活动",
        description: "年度机器人展示",
        category: "innovation",
        academic_year: "2025-2026",
        semester: "autumn",
        start_date: undefined,
        end_date: undefined,
        details: {
          location: "B202",
        },
      });
    });

    await userEvent.click(screen.getByRole("button", { name: "提交审核" }));

    await waitFor(() => {
      expect(mockSubmitActivity).toHaveBeenCalledWith("activity-1");
    });

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith("/activities/activity-1");
    });

    await waitFor(() => {
      expect(screen.queryByRole("button", { name: "提交审核" })).toBeNull();
    });
  });

  it("does not show submit for review outside the draft flow", async () => {
    mockGetActivityDetail.mockResolvedValue({
      ...draftActivity,
      status: "pending_review",
    });

    renderPage();

    expect(await screen.findByRole("heading", { name: "编辑活动" })).toBeTruthy();
    expect(screen.queryByRole("button", { name: "提交审核" })).toBeNull();
  });
});
