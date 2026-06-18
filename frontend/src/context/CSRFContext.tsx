import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';
import axios from 'axios';

interface CSRFContextType {
  token: string | null;
  isReady: boolean;
  refetch: () => Promise<string | null>;
}

const CSRFContext = createContext<CSRFContextType>({
  token: null,
  isReady: false,
  refetch: async () => null,
});

export const useCSRF = () => useContext(CSRFContext);

interface CSRFProviderProps {
  children: ReactNode;
}

export const CSRFProvider: React.FC<CSRFProviderProps> = ({ children }) => {
  const [token, setToken] = useState<string | null>(null);
  const [isReady, setIsReady] = useState(false);

  const refetch = useCallback(async (): Promise<string | null> => {
    try {
      const response = await axios.get('/api/csrf-token', { withCredentials: true });
      const newToken = response.data.csrfToken;
      setToken(newToken);
      setIsReady(true);
      return newToken;
    } catch (error) {
      setIsReady(true);
      return null;
    }
  }, []);

  useEffect(() => {
    refetch();
  }, [refetch]);

  const value: CSRFContextType = {
    token,
    isReady,
    refetch,
  };

  return <CSRFContext.Provider value={value}>{children}</CSRFContext.Provider>;
};
