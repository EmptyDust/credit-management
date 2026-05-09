// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import TeachersPage from "@/modules/users/pages/TeachersPage";

const {
  mockCreateTeacher,
  mockDeleteTeacher,
  mockExportTeachers,
  mockListTeachers,
  mockResetUserPassword,
  mockUpdateUserStatus,
} = vi.hoisted(() => ({
  mockCreateTeacher: vi.fn(),
  mockDeleteTeacher: vi.fn(),
  mockExportTeachers: vi.fn(),
  mockListTeachers: vi.fn(),
  mockResetUserPassword: vi.fn(),
  mockUpdateUserStatus: vi.fn(),
}));

vi.mock("@/modules/imports/components/UserImportDialog", () => ({
  UserImportDialog: () => null,
}));

vi.mock("@/modules/users/api/users", () => ({
  createTeacher: mockCreateTeacher,
  deleteTeacher: mockDeleteTeacher,
  exportTeachers: mockExportTeachers,
  listTeachers: mockListTeachers,
  resetUserPassword: mockResetUserPassword,
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
      page_size: 10,
      total_pages: 1,
    });
    mockCreateTeacher.mockResolvedValue({ id: "teacher-2" });
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
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText("Tina")).toBeTruthy();

    await user.type(screen.getByLabelText("教师用户名"), "tom");
    await user.type(screen.getByLabelText("教师密码"), "Password123");
    await user.type(screen.getByLabelText("教师姓名"), "Tom");
    await user.type(screen.getByLabelText("教师工号"), "T-002");
    await user.type(screen.getByLabelText("教师职称"), "副教授");
    await user.click(screen.getByRole("button", { name: "创建教师" }));

    await waitFor(() => {
      expect(mockCreateTeacher).toHaveBeenCalledWith({
        username: "tom",
        password: "Password123",
        real_name: "Tom",
        teacher_id: "T-002",
        title: "副教授",
      });
    });

    await user.click(screen.getByRole("button", { name: "停用 tina" }));
    expect(mockUpdateUserStatus).toHaveBeenCalledWith("teacher-1", "disabled");

    await user.type(screen.getByLabelText("新密码"), "NewPassword123");
    await user.click(screen.getByRole("button", { name: "重置 tina 密码" }));
    expect(mockResetUserPassword).toHaveBeenCalledWith(
      "teacher-1",
      "NewPassword123"
    );

    await user.click(screen.getByRole("button", { name: "删除 tina" }));
    expect(mockDeleteTeacher).toHaveBeenCalledWith("teacher-1");
  });
});
