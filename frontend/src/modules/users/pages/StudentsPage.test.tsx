// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import StudentsPage from "@/modules/users/pages/StudentsPage";

const {
  mockCreateStudent,
  mockDeleteStudent,
  mockExportStudents,
  mockListStudents,
  mockResetUserPassword,
  mockUpdateUserStatus,
} = vi.hoisted(() => ({
  mockCreateStudent: vi.fn(),
  mockDeleteStudent: vi.fn(),
  mockExportStudents: vi.fn(),
  mockListStudents: vi.fn(),
  mockResetUserPassword: vi.fn(),
  mockUpdateUserStatus: vi.fn(),
}));

vi.mock("@/modules/imports/components/UserImportDialog", () => ({
  UserImportDialog: () => null,
}));

vi.mock("@/modules/users/api/users", () => ({
  createStudent: mockCreateStudent,
  deleteStudent: mockDeleteStudent,
  exportStudents: mockExportStudents,
  listStudents: mockListStudents,
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
      <StudentsPage />
    </QueryClientProvider>
  );
}

describe("StudentsPage", () => {
  beforeEach(() => {
    mockListStudents.mockResolvedValue({
      items: [
        {
          id: "student-1",
          username: "alice",
          real_name: "Alice",
          user_type: "student",
          status: "active",
          student_id: "20240001",
          grade: "2024",
          department_id: "class-1",
        },
      ],
      total: 1,
      page: 1,
      page_size: 10,
      total_pages: 1,
    });
    mockCreateStudent.mockResolvedValue({ id: "student-2" });
    mockUpdateUserStatus.mockResolvedValue({ id: "student-1", status: "disabled" });
    mockResetUserPassword.mockResolvedValue(undefined);
    mockDeleteStudent.mockResolvedValue(undefined);
    mockExportStudents.mockResolvedValue(undefined);
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("creates students and manages status password and deletion", async () => {
    const user = userEvent.setup();

    renderPage();

    expect(await screen.findByText("Alice")).toBeTruthy();

    await user.type(screen.getByLabelText("学生用户名"), "bob");
    await user.type(screen.getByLabelText("学生密码"), "Password123");
    await user.type(screen.getByLabelText("学生姓名"), "Bob");
    await user.type(screen.getByLabelText("学生学号"), "20240002");
    await user.type(screen.getByLabelText("学生年级"), "2024");
    await user.click(screen.getByRole("button", { name: "创建学生" }));

    await waitFor(() => {
      expect(mockCreateStudent).toHaveBeenCalledWith({
        username: "bob",
        password: "Password123",
        real_name: "Bob",
        student_id: "20240002",
        grade: "2024",
      });
    });

    await user.click(screen.getByRole("button", { name: "停用 alice" }));
    expect(mockUpdateUserStatus).toHaveBeenCalledWith("student-1", "disabled");

    await user.type(screen.getByLabelText("新密码"), "NewPassword123");
    await user.click(screen.getByRole("button", { name: "重置 alice 密码" }));
    expect(mockResetUserPassword).toHaveBeenCalledWith(
      "student-1",
      "NewPassword123"
    );

    await user.click(screen.getByRole("button", { name: "删除 alice" }));
    expect(mockDeleteStudent).toHaveBeenCalledWith("student-1");
  });
});
