import axios, { type AxiosInstance, type InternalAxiosRequestConfig } from "axios";
import { AUTH_ENDPOINTS } from "@/constants/api";
import { ROUTES } from "@/constants/route";
import { toCamelCase, toSnakeCase } from "./caseTransform";
import { getAccessToken, getRefreshToken, setAccessToken, clearSession } from "./session";

interface RetryableConfig extends InternalAxiosRequestConfig {
  _retry?: boolean;
}

const AUTH_BASE_URL = process.env.NEXT_PUBLIC_AUTH_API_URL ?? "http://localhost:8080/api/v1";

let refreshPromise: Promise<string | null> | null = null;

/** De-duplicated so concurrent 401s from multiple instances share one refresh call. */
function refreshAccessToken(): Promise<string | null> {
  if (refreshPromise) return refreshPromise;

  const refreshToken = getRefreshToken();
  if (!refreshToken) return Promise.resolve(null);

  refreshPromise = axios
    .post(`${AUTH_BASE_URL}${AUTH_ENDPOINTS.REFRESH}`, { refreshToken })
    .then((response) => {
      const accessToken = response.data?.accessToken as string | undefined;
      if (!accessToken) return null;
      setAccessToken(accessToken);
      return accessToken;
    })
    .catch(() => null)
    .finally(() => {
      refreshPromise = null;
    });

  return refreshPromise;
}

function redirectToLogin() {
  if (typeof window === "undefined") return;
  clearSession();
  if (window.location.pathname !== ROUTES.AUTH.LOGIN) {
    window.location.href = ROUTES.AUTH.LOGIN;
  }
}

export interface CreateHttpClientOptions {
  baseURL: string;
  /** false for auth-service, which already returns camelCase JSON. */
  transformCase: boolean;
}

export function createHttpClient({ baseURL, transformCase }: CreateHttpClientOptions): AxiosInstance {
  const instance = axios.create({ baseURL, timeout: 15000 });

  instance.interceptors.request.use((config) => {
    const token = getAccessToken();
    if (token) {
      config.headers.set("Authorization", `Bearer ${token}`);
    }
    if (transformCase) {
      if (config.data) config.data = toSnakeCase(config.data);
      if (config.params) config.params = toSnakeCase(config.params);
    }
    return config;
  });

  instance.interceptors.response.use(
    (response) => {
      if (transformCase && response.data !== undefined) {
        response.data = toCamelCase(response.data);
      }
      return response;
    },
    async (error) => {
      const config = error?.config as RetryableConfig | undefined;
      const isRefreshCall = config?.url?.includes(AUTH_ENDPOINTS.REFRESH);

      if (error?.response?.status === 401 && config && !config._retry && !isRefreshCall) {
        config._retry = true;
        const newToken = await refreshAccessToken();
        if (newToken) {
          config.headers.set("Authorization", `Bearer ${newToken}`);
          return instance(config);
        }
        redirectToLogin();
      }

      return Promise.reject(error);
    }
  );

  return instance;
}
