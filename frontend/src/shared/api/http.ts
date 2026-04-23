import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";
import toast from "react-hot-toast";

import type {
  ApiErrorResponse,
  ApiSuccessResponse,
  AuthPayload,
} from "@/shared/types/api";

export const ACCESS_TOKEN_KEY = "token";
export const REFRESH_TOKEN_KEY = "refreshToken";
export const USER_STORAGE_KEY = "user";

type RetryableRequestConfig = InternalAxiosRequestConfig & {
  _retry?: boolean;
};

export const http = axios.create({
  baseURL: "/api/v1",
  timeout: 15000,
  headers: {
    "Content-Type": "application/json",
  },
});

function getAccessToken() {
  return localStorage.getItem(ACCESS_TOKEN_KEY);
}

function getRefreshToken() {
  return localStorage.getItem(REFRESH_TOKEN_KEY);
}

export function storeTokens(accessToken: string, refreshToken?: string) {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
  if (refreshToken) {
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  }
  http.defaults.headers.common.Authorization = `Bearer ${accessToken}`;
}

export function clearStoredSession() {
  localStorage.removeItem(ACCESS_TOKEN_KEY);
  localStorage.removeItem(REFRESH_TOKEN_KEY);
  localStorage.removeItem(USER_STORAGE_KEY);
  delete http.defaults.headers.common.Authorization;
}

let refreshPromise: Promise<string | null> | null = null;

async function refreshAccessToken() {
  const refreshToken = getRefreshToken();
  if (!refreshToken) {
    return null;
  }

  if (!refreshPromise) {
    refreshPromise = axios
      .post<ApiSuccessResponse<AuthPayload>>("/api/v1/auth/refresh", {
        refresh_token: refreshToken,
      })
      .then((response) => {
        const payload = response.data.data;
        storeTokens(payload.access_token, payload.refresh_token);
        return payload.access_token;
      })
      .catch(() => {
        clearStoredSession();
        return null;
      })
      .finally(() => {
        refreshPromise = null;
      });
  }

  return refreshPromise;
}

http.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  if (config.data instanceof FormData) {
    delete config.headers["Content-Type"];
  }

  return config;
});

http.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiErrorResponse>) => {
    const response = error.response;
    const config = error.config as RetryableRequestConfig | undefined;
    const requestPath = config?.url ?? "";

    if (
      response?.status === 401 &&
      config &&
      !config._retry &&
      requestPath !== "/auth/login" &&
      requestPath !== "/auth/refresh"
    ) {
      config._retry = true;
      const nextToken = await refreshAccessToken();
      if (nextToken) {
        config.headers.Authorization = `Bearer ${nextToken}`;
        return http(config);
      }
    }

    if (response) {
      const message = response.data?.message ?? "请求失败";

      switch (response.status) {
        case 401:
          clearStoredSession();
          if (window.location.pathname !== "/login") {
            window.location.href = "/login";
          }
          toast.error("登录已过期，请重新登录");
          break;
        case 403:
          toast.error("权限不足，无法执行此操作");
          break;
        case 404:
          toast.error("请求的资源不存在");
          break;
        case 409:
          toast.error(message);
          break;
        case 422:
          toast.error(message);
          break;
        case 429:
          toast.error("请求过于频繁，请稍后再试");
          break;
        case 500:
          toast.error("服务器内部错误，请稍后再试");
          break;
        default:
          toast.error(message);
      }
    } else if (error.request) {
      toast.error("网络连接失败，请检查网络设置");
    } else {
      toast.error("请求配置错误");
    }

    return Promise.reject(error);
  }
);
