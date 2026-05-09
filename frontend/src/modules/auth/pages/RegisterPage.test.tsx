// @vitest-environment jsdom

import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

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

vi.mock("@/components/ui/select", async () => {
  const React = await vi.importActual<typeof import("react")>("react");

  type SelectData = {
    ariaLabel?: string;
    disabled?: boolean;
    placeholder?: string;
    items: Array<{ value: string; label: string }>;
  };

  const SelectTrigger = ({ children }: { children?: React.ReactNode }) => (
    <>{children}</>
  );
  const SelectValue = () => null;
  const SelectContent = ({ children }: { children?: React.ReactNode }) => (
    <>{children}</>
  );
  const SelectItem = ({ children }: { children?: React.ReactNode }) => (
    <>{children}</>
  );

  function collectSelectData(children: React.ReactNode): SelectData {
    const data: SelectData = { items: [] };

    function visit(node: React.ReactNode) {
      React.Children.forEach(node, (child) => {
        if (!React.isValidElement(child)) {
          return;
        }

        if (child.type === SelectTrigger) {
          const props = child.props as {
            "aria-label"?: string;
            disabled?: boolean;
          };
          data.ariaLabel = props["aria-label"];
          data.disabled = props.disabled;
        }

        if (child.type === SelectValue) {
          const props = child.props as { placeholder?: string };
          data.placeholder = props.placeholder;
        }

        if (child.type === SelectItem) {
          const props = child.props as { value: string; children?: React.ReactNode };
          data.items.push({
            value: props.value,
            label: String(props.children ?? props.value),
          });
        }

        const props = child.props as { children?: React.ReactNode };
        visit(props.children);
      });
    }

    visit(children);
    return data;
  }

  return {
    Select: ({
      children,
      defaultValue,
      disabled,
      onValueChange,
      value,
    }: {
      children?: React.ReactNode;
      defaultValue?: string;
      disabled?: boolean;
      onValueChange?: (value: string) => void;
      value?: string;
    }) => {
      const data = collectSelectData(children);

      return (
        <select
          aria-label={data.ariaLabel}
          defaultValue={value === undefined ? defaultValue || "" : undefined}
          disabled={disabled || data.disabled}
          value={value}
          onChange={(event) => onValueChange?.(event.target.value)}
        >
          <option value="" disabled>
            {data.placeholder || "请选择"}
          </option>
          {data.items.map((item) => (
            <option key={item.value} value={item.value}>
              {item.label}
            </option>
          ))}
        </select>
      );
    },
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
  };
});

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

  afterEach(() => {
    cleanup();
  });

  it("submits the student registration payload and returns to login", async () => {
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

    fireEvent.change(screen.getByLabelText("用户名"), {
      target: { value: "alice_01" },
    });
    fireEvent.change(screen.getByLabelText("真实姓名"), { target: { value: "Alice" } });
    fireEvent.change(screen.getByLabelText("密码"), {
      target: { value: "Password123" },
    });
    fireEvent.change(screen.getByLabelText("确认密码"), {
      target: { value: "Password123" },
    });
    fireEvent.change(screen.getByLabelText("邮箱"), {
      target: { value: "alice@example.com" },
    });
    fireEvent.change(screen.getByLabelText("手机号码"), {
      target: { value: "13812345678" },
    });
    fireEvent.change(screen.getByLabelText("学号"), { target: { value: "20240001" } });

    fireEvent.change(await screen.findByLabelText("学部"), {
      target: { value: "college-1" },
    });
    fireEvent.change(await screen.findByLabelText("专业"), {
      target: { value: "major-1" },
    });
    fireEvent.change(await screen.findByLabelText("班级"), {
      target: { value: "class-1" },
    });
    fireEvent.change(await screen.findByLabelText("年级"), { target: { value: "2024" } });

    const originalSetTimeout = globalThis.setTimeout;
    const redirectTimer = vi
      .spyOn(globalThis, "setTimeout")
      .mockImplementation((handler: TimerHandler, timeout?: number) => {
        if (timeout === 2000 && typeof handler === "function") {
          handler();
          return 0 as ReturnType<typeof setTimeout>;
        }

        return originalSetTimeout(handler, timeout);
      });

    try {
      fireEvent.click(screen.getByRole("button", { name: "注册学生账号" }));

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
      expect(mockNavigate).toHaveBeenCalledWith("/login");
    } finally {
      redirectTimer.mockRestore();
    }
  });
});
