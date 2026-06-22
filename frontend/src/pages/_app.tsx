import type { AppProps } from 'next/app'
import Head from 'next/head'
import { AuthProvider } from '@/context/AuthContext'
import { CSRFProvider, useCSRF } from '@/context/CSRFContext'
import { ThemeProvider } from 'next-themes'
import { Toaster } from 'sonner'
import { useEffect } from 'react'
import { setCSRFToken, setCSRFReady } from '@/lib/api'
import AppLayout from '@/components/layout/AppLayout'
import { ErrorBoundary } from '@/components/error-boundary'
import '@/styles/globals.css'

function CSRFSyncComponent() {
  const { token, isReady } = useCSRF();

  useEffect(() => {
    setCSRFToken(token);
    setCSRFReady(isReady);
  }, [token, isReady]);

  return null;
}

export default function App({ Component, pageProps }: AppProps) {
  return (
    <>
      <Head>
        <link rel="icon" href="/assets/servasec-favicon.svg" type="image/svg+xml" />
      </Head>
      <AuthProvider>
        <CSRFProvider>
          <CSRFSyncComponent />
          <ThemeProvider
            attribute="class"
            defaultTheme="light"
            enableSystem
            disableTransitionOnChange
          >
            <AppLayout>
              <ErrorBoundary>
                <Component {...pageProps} />
              </ErrorBoundary>
            </AppLayout>
            <Toaster position="top-right" richColors />
          </ThemeProvider>
        </CSRFProvider>
      </AuthProvider>
    </>
  )
}
