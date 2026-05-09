// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import TeachersPage from "@/modules/users/pages/TeachersPage";

const {
  mockCreateTeacher,
  mockDeleteTeacher,
  mockExportTeachers,
  mockGetTeacherStats,
  mockListTeachers,
  mockResetUserPassword,
  mockUpdateTeacher,
  mockUpdateUserStatus,
} = vi.hoisted(() => ({
  mockCreateTeacher: vi.fn(),
  mockDeleteTeacher: vi.fn(),
  mockExportTeachers: vi.fn(),
  mockGetTeacherStats: vi.fn(),
  mockListTeachers: vi.fn(),
  mockResetUserPassword: vi.fn(),
  mockUpdateTeacher: vi.fn(),
  mockUpdateUserStatus: vi.fn(),
}));

vi.mock("@/modules/imports/components/UserImportDialog", () => ({
  UserImportDialog: () => null,
}));

vi.mock("@/modules/users/api/users", () => ({
  createTeacher: mockCreateTeacher,
  deleteTeacher: mockDeleteTeacher,
  exportTeachers: mockExportTeachers,
  getTeacherStats: mockGetTeacherStats,
  listTeachers: mockListTeachers,
  resetUserPassword: mockResetUserPassword,
  updateTeacher: mockUpdateTeacher,
  updateUserStatus: mockUpdateUserStatus,
}));

vi.mock("react-hot-toast", () => ({
  default: {
    error: vi.fn(),
    success: vi.fn(),
  },
}));

function renderPage() {
  render(
    <QueryClientProvider
      client={
        new QueryClient({
          defaultOptions: {
            queries: { retry: false },
          },
        })
      }
    >
      <TeachersPage />
    </QueryClientProvider>
  );
}

describe("TeachersPage", () => {
  beforeEach(() => {
    mockListTeachers.mockResolvedValue({
      items: [
        {
          id: "teacher-1",
          username: "tina",
          real_name: "Tina",
          user_type: "teacher",
          status: "active",
          teacher_id: "T-001",
          title: "教授",
          department_id: "department-1",
        },
      ],
      total: 1,
      page: 1,
      page_size: 20,
      total_pages: 2,
    });
    mockGetTeacherStats.mockResolvedValue({
      total: 2,
      status_counts: {
        active: 2,
      },
      title_counts: [
        { key: "教授", label: "教授", count: 1 },
        { key: "研究员", label: "研究员", count: 1 },
      ],
    });
    mockCreateTeacher.mockResolvedValue({ id: "teacher-2" });
    mockUpdateTeacher.mockResolvedValue({ id: "teacher-1" });
    mockUpdateUserStatus.mockResolvedValue({ id: "teacher-1", status: "disabled" });
    mockResetUserPassword.mockResolvedValue(undefined);
    mockDeleteTeacher.mockResolvedValue(undefined);
    mockExportTeachers.mockResolvedValue(undefined);
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("creates teachers and manages status password and deletion", async () => {
    renderPage();

    expect(await screen.findByText("Tina")).toBeTruthy();
    expect(await screen.findByText("总教师数")).toBeTruthy();
    expect(screen.getByText("活跃教师")).toBeTruthy();

    fireEvent.click(screen.getByRole("button", { name: "下一页" }));
    await waitFor(() => {
      expect(mockListTeachers).toHaveBeenLastCalledWith(
        expect.objectContaining({ page: 2, page_size: 20 })
      );
    });

    fireEvent.change(screen.getByLabelText("教师用户名"), { target: { value: "tom" } });
    fireEvent.change(screen.getByLabelText("教师密码"), {
      target: { value: "Password123" },
    });
    fireEvent.change(screen.getByLabelText("教师姓名"), { target: { value: "Tom" } });
    fireEvent.change(screen.getByLabelText("教师工号"), { target: { value: "T-002" } });
    fireEvent.change(screen.getByLabelText("教师职称"), { target: { value: "副教授" } });
    fireEvent.click(screen.getByRole("button", { name: "创建教师" }));

    await waitFor(() => {
      expect(mockCreateTeacher).toHaveBeenCalledWith({
        username: "tom",
        password: "Password123",
        real_name: "Tom",
        teacher_id: "T-002",
        title: "副教授",
      });
    });

    fireEvent.click(screen.getByRole("button", { name: "编辑 tina" }));
    fireEvent.change(screen.getByLabelText("编辑教师姓名"), {
      target: { value: "Tina Updated" },
    });
    fireEvent.change(screen.getByLabelText("编辑教师工号"), {
      target: { value: "T-999" },
    });
    fireEvent.change(screen.getByLabelText("编辑教师职称"), {
      target: { value: "研究员" },
    });
    fireEvent.click(screen.getByRole("button", { name: "保存教师" }));

    await waitFor(() => {
      expect(mockUpdateTeacher).toHaveBeenCalledWith("teacher-1", {
        username: "tina",
        real_name: "Tina Updated",
        teacher_id: "T-999",
        title: "研究员",
      });
    });

    fireEvent.click(screen.getByRole("button", { name: "停用 tina" }));
    await waitFor(() => {
      expect(mockUpdateUserStatus).toHaveBeenCalledWith("teacher-1", "disabled");
    });

    fireEvent.change(screen.getByLabelText("新密码"), {
      target: { value: "NewPassword123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "重置 tina 密码" }));
    await waitFor(() => {
      expect(mockResetUserPassword).toHaveBeenCalledWith(
        "teacher-1",
        "NewPassword123"
      );
    });

    fireEvent.click(screen.getByRole("button", { name: "删除 tina" }));
    await waitFor(() => {
      expect(mockDeleteTeacher).toHaveBeenCalledWith("teacher-1");
    });
  });
});
