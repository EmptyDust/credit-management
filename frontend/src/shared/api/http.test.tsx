// @vitest-environment jsdom

import { beforeEach, describe, expect, it, vi } from "vitest";

import { clearStoredSession, http } from "@/shared/api/http";

const { mockToastError } = vi.hoisted(() => ({
  mockToastError: vi.fn(),
}));

vi.mock("react-hot-toast", () => ({
  default: {
    error: mockToastError,
  },
}));

describe("http unauthorized handling", () => {
  beforeEach(() => {
    const storage = new Map<string, string>();
    Object.defineProperty(window, "localStorage", {
      configurable: true,
      value: {
        getItem: (key: string) => storage.get(key) ?? null,
        setItem: (key: string, value: string) => {
          storage.set(key, value);
        },
        removeItem: (key: string) => {
          storage.delete(key);
        },
      },
    });
    Object.defineProperty(globalThis, "localStorage", {
      configurable: true,
      value: window.localStorage,
    });

    let currentHref = "http://localhost/register";
    Object.defineProperty(window, "location", {
      configurable: true,
      value: {
        get href() {
          return currentHref;
        },
        set href(value: string) {
          currentHref = value.startsWith("http")
            ? value
            : `http://localhost${value}`;
        },
        get pathname() {
          return new URL(currentHref).pathname;
        },
      },
    });

    clearStoredSession();
    mockToastError.mockReset();
  });

  it("does not force public auth pages back to login when an optional request returns 401", async () => {
    const rejected = (
      http.interceptors.response as unknown as {
        handlers: Array<{ rejected: (error: unknown) => Promise<unknown> }>;
      }
    ).handlers[0].rejected;

    const error = {
      response: {
        status: 401,
        data: {
          message: "unauthorized",
        },
      },
      config: {
        url: "/options/teacher-titles",
        headers: {},
      },
    };

    await rejected(error).catch(() => undefined);

    expect(window.location.href).toBe("http://localhost/register");
  });

  it("redirects authenticated users to login when a protected request returns 401", async () => {
    window.localStorage.setItem("token", "expired-access-token");
    window.location.href = "http://localhost/dashboard";

    const rejected = (
      http.interceptors.response as unknown as {
        handlers: Array<{ rejected: (error: unknown) => Promise<unknown> }>;
      }
    ).handlers[0].rejected;

    const error = {
      response: {
        status: 401,
        data: {
          message: "unauthorized",
        },
      },
      config: {
        url: "/me/profile",
        headers: {},
      },
    };

    await rejected(error).catch(() => undefined);

    expect(window.location.href).toBe("http://localhost/login");
    expect(mockToastError).toHaveBeenCalledWith("登录已过期，请重新登录");
  });
});
