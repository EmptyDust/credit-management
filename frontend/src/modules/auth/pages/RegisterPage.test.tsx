// @vitest-environment jsdom

import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { beforeEach, describe, expect, it, vi } from "vitest";

import RegisterPage from "@/modules/auth/pages/RegisterPage";

const { mockNavigate, mockPost, mockGetOptions, mockToastSuccess } = vi.hoisted(
  () => ({
    mockNavigate: vi.fn(),
    mockPost: vi.fn(),
    mockGetOptions: vi.fn(),
    mockToastSuccess: vi.fn(),
  })
);

if (!HTMLElement.prototype.hasPointerCapture) {
  HTMLElement.prototype.hasPointerCapture = () => false;
}

if (!HTMLElement.prototype.setPointerCapture) {
  HTMLElement.prototype.setPointerCapture = () => {};
}

if (!HTMLElement.prototype.releasePointerCapture) {
  HTMLElement.prototype.releasePointerCapture = () => {};
}

if (!HTMLElement.prototype.scrollIntoView) {
  HTMLElement.prototype.scrollIntoView = () => {};
}

vi.mock("@/lib/api", () => ({
  default: {
    post: mockPost,
  },
}));

vi.mock("@/lib/options", () => ({
  getOptions: () => mockGetOptions(),
}));

vi.mock("react-hot-toast", () => ({
  default: {
    success: mockToastSuccess,
  },
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

describe("RegisterPage", () => {
  beforeEach(() => {
    mockNavigate.mockReset();
    mockPost.mockReset();
    mockGetOptions.mockReset();
    mockToastSuccess.mockReset();
  });

  it("submits the student registration payload and returns to login", async () => {
    const user = userEvent.setup();

    mockGetOptions.mockResolvedValue({
      colleges: [{ value: "college-1", label: "计算机学院" }],
      majors: {
        "college-1": [{ value: "major-1", label: "软件工程" }],
      },
      classes: {
        "major-1": [{ value: "class-1", label: "软件工程一班" }],
      },
      grades: [{ value: "2024", label: "2024级" }],
    });
    mockPost.mockResolvedValue({
      data: {
        code: "OK",
        message: "OK",
      },
    });

    render(
      <MemoryRouter>
        <RegisterPage />
      </MemoryRouter>
    );

    await user.type(screen.getByLabelText("用户名"), "alice_01");
    await user.type(screen.getByLabelText("真实姓名"), "Alice");
    await user.type(screen.getByLabelText("密码"), "Password123");
    await user.type(screen.getByLabelText("确认密码"), "Password123");
    await user.type(screen.getByLabelText("邮箱"), "alice@example.com");
    await user.type(screen.getByLabelText("手机号码"), "13812345678");
    await user.type(screen.getByLabelText("学号"), "20240001");

    await user.click(screen.getByRole("combobox", { name: "学部" }));
    await user.click(await screen.findByRole("option", { name: "计算机学院" }));

    await user.click(screen.getByRole("combobox", { name: "专业" }));
    await user.click(await screen.findByRole("option", { name: "软件工程" }));

    await user.click(screen.getByRole("combobox", { name: "班级" }));
    await user.click(await screen.findByRole("option", { name: "软件工程一班" }));

    await user.click(screen.getByRole("combobox", { name: "年级" }));
    await user.click(await screen.findByRole("option", { name: "2024级" }));

    await user.click(screen.getByRole("button", { name: "注册学生账号" }));

    await waitFor(() => {
      expect(mockPost).toHaveBeenCalledWith("/auth/register/student", {
        student_id: "20240001",
        username: "alice_01",
        password: "Password123",
        email: "alice@example.com",
        phone: "13812345678",
        real_name: "Alice",
        department_id: "class-1",
        grade: "2024",
      });
    });

    expect(mockToastSuccess).toHaveBeenCalledWith("注册成功！正在跳转到登录页面...");
    await waitFor(
      () => {
        expect(mockNavigate).toHaveBeenCalledWith("/login");
      },
      { timeout: 2500 }
    );
  });
});
