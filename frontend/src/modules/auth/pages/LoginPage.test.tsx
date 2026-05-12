// @vitest-environment jsdom

import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi, beforeEach } from "vitest";

import { memoryRouterFuture } from "@/app/router/future";
import LoginPage from "@/modules/auth/pages/LoginPage";

const { mockLogin, mockNavigate, mockPost } = vi.hoisted(() => ({
  mockLogin: vi.fn(),
  mockNavigate: vi.fn(),
  mockPost: vi.fn(),
}));

vi.mock("@/contexts/AuthContext", () => ({
  useAuth: () => ({
    login: mockLogin,
  }),
}));

vi.mock("@/lib/api", () => ({
  default: {
    post: mockPost,
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

describe("LoginPage", () => {
  beforeEach(() => {
    mockLogin.mockReset();
    mockNavigate.mockReset();
    mockPost.mockReset();
  });

it("submits the login payload and redirects to the dashboard", async () => {
    mockPost.mockResolvedValue({
      data: {
        data: {
          access_token: "access-token",
          refresh_token: "refresh-token",
          user: {
            id: "user-1",
            uuid: "user-1",
            username: "alice",
            userType: "student",
            user_type: "student",
            real_name: "Alice",
            avatar_url: "https://example.com/avatar.png",
            status: "active",
          },
        },
      },
    });

    render(
      <MemoryRouter future={memoryRouterFuture}>
        <LoginPage />
      </MemoryRouter>
    );

    await userEvent.type(screen.getByLabelText("用户名"), "alice");
    await userEvent.type(screen.getByLabelText("密码"), "Password123");
    await userEvent.click(screen.getByRole("button", { name: "登录" }));

    await waitFor(() => {
      expect(mockPost).toHaveBeenCalledWith("/auth/login", {
        login_id: "alice",
        password: "Password123",
      });
    });

    expect(mockLogin).toHaveBeenCalledWith("access-token", "refresh-token", {
      id: "user-1",
      uuid: "user-1",
      username: "alice",
      userType: "student",
      user_type: "student",
      real_name: "Alice",
      avatar_url: "https://example.com/avatar.png",
      status: "active",
      fullName: "Alice",
      avatar: "https://example.com/avatar.png",
    });
    expect(mockNavigate).toHaveBeenCalledWith("/dashboard");
  });
});
