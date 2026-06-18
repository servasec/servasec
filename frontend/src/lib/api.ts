import axios from "axios";

const instance = axios.create({
  withCredentials: true,
});

let csrfToken: string | null = null;
let csrfReady: boolean = false;

export const setCSRFToken = (token: string | null) => {
  csrfToken = token;
  csrfReady = true;
};

export const setCSRFReady = (ready: boolean) => {
  csrfReady = ready;
};

export const fetchCSRFToken = async () => {
  try {
    const response = await axios.get('/api/csrf-token', { withCredentials: true });
    csrfToken = response.data.csrfToken;
    csrfReady = true;
    return csrfToken;
  } catch (error) {
    csrfReady = true;
    throw error;
  }
};

let isRefreshing = false;
let refreshFailed = false;
let failedQueue: Array<{resolve: Function, reject: Function}> = [];

const processQueue = (error: any, token: string | null = null) => {
  failedQueue.forEach(prom => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
};

instance.interceptors.request.use(
  async (config) => {
    const method = config.method?.toLowerCase();

    if (method && ['post', 'put', 'patch', 'delete'].includes(method)) {
      if (!csrfReady) {
        let attempts = 0;
        while (!csrfReady && attempts < 50) {
          await new Promise(resolve => setTimeout(resolve, 100));
          attempts++;
        }
      }

      if (csrfToken) {
        config.headers['X-CSRF-Token'] = csrfToken;
      }
    }

    return config;
  },
  (error) => Promise.reject(error)
);

instance.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config;

    if (
      originalRequest.url?.includes('/api/login') ||
      originalRequest.url?.includes('/api/logout') ||
      originalRequest.url?.includes('/api/refresh') ||
      originalRequest.url?.includes('/api/me')
    ) {
      throw error;
    }

    if (error.response?.status === 401 && !originalRequest._retry) {
      const isOnLoginPage = typeof window !== 'undefined' && window.location.pathname === '/login';
      if (refreshFailed || isOnLoginPage) {
        throw error;
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then(() => instance(originalRequest))
          .catch((err) => { throw err; });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        await axios.post('/api/refresh', {}, { withCredentials: true });
        processQueue(null, null);
        refreshFailed = false;

        try {
          await fetchCSRFToken();
        } catch (csrfError) {
          // ignore
        }

        return instance(originalRequest);
      } catch (refreshError) {
        refreshFailed = true;
        processQueue(refreshError, null);

        if (typeof window !== 'undefined' && window.location.pathname !== '/login') {
          window.location.href = '/login';
        }
        throw refreshError;
      } finally {
        isRefreshing = false;
      }
    }

    if (error.response?.status === 403) {
      const errorCode = error.response?.data?.error;
      if (errorCode && ['csrf_token_missing', 'csrf_token_invalid', 'csrf_validation_failed'].includes(errorCode)) {
        try {
          await fetchCSRFToken();
          return instance(originalRequest);
        } catch (csrfError) {
          throw error;
        }
      }
    }

    throw error;
  }
);

export default instance;
