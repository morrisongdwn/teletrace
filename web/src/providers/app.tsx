import { QueryClientProvider } from "@tanstack/react-query";
import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { Suspense } from "react";
import { ErrorBoundary } from "react-error-boundary";
import { HelmetProvider } from "react-helmet-async";
import { BrowserRouter } from "react-router-dom";

import { Loader } from "@/components/Elements/Loader";
import { queryClient } from "@/libs/react-query";

export type AppProvidersProps = {
  children: React.ReactNode;
};

const ErrorFallback = () => (
  <div>Oops, something went wrong! try refreshing</div>
);

export const AppProviders = ({ children }: AppProvidersProps) => (
  <Suspense fallback={<Loader />}>
    <ErrorBoundary FallbackComponent={ErrorFallback}>
      <HelmetProvider>
        <QueryClientProvider client={queryClient}>
          {process.env.NODE_ENV !== "test" && <ReactQueryDevtools />}
          <BrowserRouter>{children}</BrowserRouter>
        </QueryClientProvider>
      </HelmetProvider>
    </ErrorBoundary>
  </Suspense>
);