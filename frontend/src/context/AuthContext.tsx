import React, { createContext, useContext, useState, useEffect, useCallback, useRef, ReactNode } from "react";
import api from "@/lib/api";

interface AuthContextType {
  loggedIn: boolean;
  user: any | null;
  login: (user: any) => void;
  logout: (redirect?: boolean) => Promise<void>;
  checkAuth: () => Promise<void>;
  authChecked: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: ReactNode }) => {
  const [loggedIn, setLoggedIn] = useState(false);
  const [user, setUser] = useState<any | null>(null);
  const [authChecked, setAuthChecked] = useState(false);
  const authCheckedRef = useRef(false);
  const isCheckingRef = useRef(false);

  const login = (userData: any) => {
    setLoggedIn(true);
    setUser(userData);
  };

  const logout = async (redirect = true) => {
    try {
      await api.post("/api/logout");
    } catch (error) {
      // ignore
    }
    setLoggedIn(false);
    setUser(null);

    if (redirect && typeof window !== "undefined" && window.location.pathname !== "/login") {
      window.location.href = "/login";
    }
  };

  const checkAuth = useCallback(async () => {
    if (authCheckedRef.current || isCheckingRef.current) {
      return;
    }

    isCheckingRef.current = true;

    try {
      const res = await api.get("/api/me");
      setLoggedIn(true);
      setUser(res.data);
    } catch (err: any) {
      if (err?.response?.status !== 401 && err?.response?.status !== 403) {
        console.error('Auth check failed:', err);
      }
      setLoggedIn(false);
      setUser(null);
    } finally {
      authCheckedRef.current = true;
      isCheckingRef.current = false;
      setAuthChecked(true);
    }
  }, []);

  useEffect(() => {
    checkAuth();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <AuthContext.Provider
      value={{ loggedIn, user, login, logout, checkAuth, authChecked }}
    >
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};
