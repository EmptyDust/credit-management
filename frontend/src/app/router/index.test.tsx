// @vitest-environment jsdom

import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { createMemoryRouter } from "react-router-dom";

import { AppProviders } from "@/app/providers/AppProviders";
import { appRoutes } from "@/app/router";

describe("app router", () => {
  beforeEach(() => {
    const storage = {
      store: {} as Record<string, string>,
      getItem(key: string) {
        return this.store[key] ?? null;
      },
      setItem(key: string, value: string) {
        this.store[key] = value;
      },
      removeItem(key: string) {
        delete this.store[key];
      },
      clear() {
        this.store = {};
      },
    };

    Object.defineProperty(window, "localStorage", {
      configurable: true,
      value: storage,
    });

    class MockRequest {
      url: string;
      signal?: AbortSignal;

      constructor(input: string | URL, init?: { signal?: AbortSignal }) {
        this.url = String(input);
        this.signal = init?.signal;
      }
    }

    vi.stubGlobal("Request", MockRequest);

    window.localStorage.clear();
    window.history.pushState({}, "", "/");

    Object.defineProperty(window, "matchMedia", {
      configurable: true,
      writable: true,
      value: vi.fn().mockImplementation((query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
    });
  });

  it("redirects anonymous users to login", async () => {
    const router = createMemoryRouter(appRoutes, {
      initialEntries: ["/"],
    });

    render(<AppProviders routerOverride={router} />);

    expect(await screen.findByRole("button", { name: "登录" })).toBeTruthy();
  });
});
