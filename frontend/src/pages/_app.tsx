import type { AppProps } from 'next/app'
import Head from 'next/head'
import { useState, useEffect } from 'react'
import { AuthProvider, useAuth } from '@/context/AuthContext'
import { CSRFProvider, useCSRF } from '@/context/CSRFContext'
import { ThemeProvider } from 'next-themes'
import { Toaster } from 'sonner'
import { setCSRFToken, setCSRFReady } from '@/lib/api'
import AppLayout from '@/components/layout/AppLayout'
import { ErrorBoundary } from '@/components/error-boundary'
import { OnboardingModal } from '@/components/onboarding/onboarding-modal'
import { PageTransition } from '@/components/page-transition'
import '@/styles/globals.css'
import type { User } from '@/lib/types'

function CSRFSyncComponent() {
  const { token, isReady } = useCSRF();

  useEffect(() => {
    setCSRFToken(token);
    setCSRFReady(isReady);
  }, [token, isReady]);

  return null;
}

function OnboardingGate() {
  const { loggedIn, user, authChecked } = useAuth();
  const [show, setShow] = useState(false);

  useEffect(() => {
    if (authChecked && loggedIn && user && !user.hasSeenOnboarding) {
      setShow(true);
    }
  }, [authChecked, loggedIn, user]);

  if (!authChecked || !loggedIn || !user) return null;

  return (
    <OnboardingModal
      open={show}
      user={user}
      onComplete={(updatedUser: User) => {
        setShow(false);
      }}
    />
  );
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
            defaultTheme="dark"
            storageKey="servasec-theme"
            themes={["light", "dark", "catppuccin", "atom-one", "nord"]}
          >
            <AppLayout>
              <ErrorBoundary>
                <PageTransition>
                  <Component {...pageProps} />
                </PageTransition>
              </ErrorBoundary>
            </AppLayout>
            <OnboardingGate />
            <Toaster
              position="top-right"
              toastOptions={{
                duration: 4000,
                className: "!bg-card !border !border-border !text-foreground !shadow-sm !text-xs",
              }}
            />
          </ThemeProvider>
        </CSRFProvider>
      </AuthProvider>
    </>
  )
}
