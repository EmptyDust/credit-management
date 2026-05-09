// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import StudentsPage from "@/modules/users/pages/StudentsPage";

const {
  mockCreateStudent,
  mockDeleteStudent,
  mockExportStudents,
  mockListStudents,
  mockResetUserPassword,
  mockUpdateStudent,
  mockUpdateUserStatus,
} = vi.hoisted(() => ({
  mockCreateStudent: vi.fn(),
  mockDeleteStudent: vi.fn(),
  mockExportStudents: vi.fn(),
  mockListStudents: vi.fn(),
  mockResetUserPassword: vi.fn(),
  mockUpdateStudent: vi.fn(),
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
  updateStudent: mockUpdateStudent,
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
    mockUpdateStudent.mockResolvedValue({ id: "student-1" });
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
    renderPage();

    expect(await screen.findByText("Alice")).toBeTruthy();

    fireEvent.change(screen.getByLabelText("学生用户名"), { target: { value: "bob" } });
    fireEvent.change(screen.getByLabelText("学生密码"), {
      target: { value: "Password123" },
    });
    fireEvent.change(screen.getByLabelText("学生姓名"), { target: { value: "Bob" } });
    fireEvent.change(screen.getByLabelText("学生学号"), {
      target: { value: "20240002" },
    });
    fireEvent.change(screen.getByLabelText("学生年级"), { target: { value: "2024" } });
    fireEvent.click(screen.getByRole("button", { name: "创建学生" }));

    await waitFor(() => {
      expect(mockCreateStudent).toHaveBeenCalledWith({
        username: "bob",
        password: "Password123",
        real_name: "Bob",
        student_id: "20240002",
        grade: "2024",
      });
    });

    fireEvent.click(screen.getByRole("button", { name: "编辑 alice" }));
    fireEvent.change(screen.getByLabelText("编辑学生姓名"), {
      target: { value: "Alice Updated" },
    });
    fireEvent.change(screen.getByLabelText("编辑学生学号"), {
      target: { value: "20249999" },
    });
    fireEvent.change(screen.getByLabelText("编辑学生年级"), {
      target: { value: "2025" },
    });
    fireEvent.click(screen.getByRole("button", { name: "保存学生" }));

    await waitFor(() => {
      expect(mockUpdateStudent).toHaveBeenCalledWith("student-1", {
        username: "alice",
        real_name: "Alice Updated",
        student_id: "20249999",
        grade: "2025",
      });
    });

    fireEvent.click(screen.getByRole("button", { name: "停用 alice" }));
    await waitFor(() => {
      expect(mockUpdateUserStatus).toHaveBeenCalledWith("student-1", "disabled");
    });

    fireEvent.change(screen.getByLabelText("新密码"), {
      target: { value: "NewPassword123" },
    });
    fireEvent.click(screen.getByRole("button", { name: "重置 alice 密码" }));
    await waitFor(() => {
      expect(mockResetUserPassword).toHaveBeenCalledWith(
        "student-1",
        "NewPassword123"
      );
    });

    fireEvent.click(screen.getByRole("button", { name: "删除 alice" }));
    await waitFor(() => {
      expect(mockDeleteStudent).toHaveBeenCalledWith("student-1");
    });
  });
});
