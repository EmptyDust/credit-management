// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

import ActivityCreatePage from "@/modules/activities/pages/ActivityCreatePage";

const {
  mockCreateActivity,
  mockGetActivityCategorySchema,
  mockListActivityCategories,
  mockNavigate,
} = vi.hoisted(() => ({
  mockCreateActivity: vi.fn().mockResolvedValue({
    id: "activity-1",
    title: "机器人展示活动",
    category: "innovation",
    academic_year: "2025-2026",
    semester: "autumn",
    status: "draft",
    owner_id: "student-1",
  }),
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
  mockListActivityCategories: vi.fn().mockResolvedValue([
    { key: "innovation", label: "创新活动", sort_order: 1 },
    { key: "volunteer", label: "志愿活动", sort_order: 2 },
  ]),
  mockNavigate: vi.fn(),
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

vi.mock("@/modules/activities/api/activity-categories", () => ({
  getActivityCategorySchema: mockGetActivityCategorySchema,
  listActivityCategories: mockListActivityCategories,
}));

vi.mock("@/modules/activities/api/activities", () => ({
  createActivity: mockCreateActivity,
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
        <ActivityCreatePage />
      </MemoryRouter>
    </QueryClientProvider>
  );
}

describe("ActivityCreatePage", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("loads the selected category schema and redirects to the edit page after draft creation", async () => {
    renderPage();

    expect(await screen.findByRole("heading", { name: "新建活动" })).toBeTruthy();
    expect(await screen.findByRole("option", { name: "创新活动" })).toBeTruthy();

    await userEvent.selectOptions(screen.getByLabelText("活动分类"), "innovation");

    expect(await screen.findByLabelText(/活动地点/)).toBeTruthy();
    expect(mockGetActivityCategorySchema).toHaveBeenCalledWith("innovation");

    await userEvent.type(screen.getByLabelText("活动标题"), "机器人展示活动");
    await userEvent.type(screen.getByLabelText("活动描述"), "年度机器人展示");
    await userEvent.type(screen.getByLabelText("学年"), "2025-2026");
    await userEvent.selectOptions(screen.getByLabelText("学期"), "autumn");
    await userEvent.type(screen.getByLabelText(/活动地点/), "A101");
    await userEvent.click(screen.getByRole("button", { name: "保存草稿" }));

    await waitFor(() => {
      expect(mockCreateActivity).toHaveBeenCalledWith({
        title: "机器人展示活动",
        description: "年度机器人展示",
        category: "innovation",
        academic_year: "2025-2026",
        semester: "autumn",
        start_date: undefined,
        end_date: undefined,
        details: {
          location: "A101",
        },
      });
    });

    expect(mockNavigate).toHaveBeenCalledWith("/activities/activity-1/edit");
  });
});
