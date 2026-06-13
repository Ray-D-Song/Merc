/* eslint-disable react-refresh/only-export-components */
import { type ComponentType, type ErrorInfo, type ReactNode } from 'react'
import { ErrorBoundary, type FallbackProps } from 'react-error-boundary'
import { Button } from '@cloudflare/kumo/components/button'
import { LayerCard } from '@cloudflare/kumo/components/layer-card'
import { useTranslation } from 'react-i18next'
import { useNavigate } from '@/hooks/use-navigate'

interface Props {
  children: ReactNode
  pageName?: string
  fallback?: ComponentType<PageErrorFallbackProps>
}

export interface PageErrorFallbackProps {
  error: unknown
  resetErrorBoundary: () => void
  pageName?: string
}

function normalizeError(error: unknown) {
  return error instanceof Error ? error : new Error(String(error))
}

function DefaultPageErrorFallback({ error, resetErrorBoundary, pageName }: PageErrorFallbackProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const normalizedError = normalizeError(error)
  const isDev = import.meta.env.DEV

  return (
    <LayerCard className="mx-auto max-w-4xl p-6">
      <div className="space-y-5 text-center">
        <div className="space-y-2">
          <div className="text-4xl font-semibold text-kumo-warning">!</div>
          <h1 className="text-2xl font-semibold text-kumo-default">{t('error.page.title')}</h1>
          <p className="text-kumo-subtle">
            {pageName
              ? t('error.page.subtitleWithPage', { pageName })
              : t('error.page.subtitle')}
          </p>
        </div>
        <div className="flex flex-wrap justify-center gap-2">
          <Button type="button" variant="primary" onClick={resetErrorBoundary}>
            {t('error.page.retry')}
          </Button>
          <Button type="button" onClick={() => navigate(-1)}>
            {t('error.page.goBack')}
          </Button>
          <Button type="button" onClick={() => window.location.reload()}>
            {t('error.page.reload')}
          </Button>
        </div>
        {isDev && (
          <div className="space-y-3 rounded-lg border border-kumo-danger/30 bg-kumo-danger/5 p-4 text-left">
            <div>
              <div className="text-sm font-semibold text-kumo-danger">Debug Info</div>
              <code className="text-sm text-kumo-danger">{normalizedError.message}</code>
            </div>
            {normalizedError.stack && (
              <pre className="max-h-60 overflow-auto rounded bg-kumo-base p-3 text-xs text-kumo-default">
                {normalizedError.stack}
              </pre>
            )}
          </div>
        )}
      </div>
    </LayerCard>
  )
}

const handlePageError = (error: unknown, errorInfo: ErrorInfo, pageName?: string) => {
  console.error(`Page Error Boundary (${pageName || 'Unknown'}) caught an error:`, error, errorInfo)
}

export default function PageErrorBoundary({ children, pageName, fallback: CustomFallback }: Props) {
  const FallbackComponent = CustomFallback || DefaultPageErrorFallback

  return (
    <ErrorBoundary
      FallbackComponent={(props: FallbackProps) => (
        <FallbackComponent {...props} pageName={pageName} />
      )}
      onError={(error, errorInfo) => handlePageError(error, errorInfo, pageName)}
    >
      {children}
    </ErrorBoundary>
  )
}

export function usePageErrorBoundary() {
  const createErrorBoundary = (pageName?: string, fallback?: ComponentType<PageErrorFallbackProps>) => {
    return function ErrorBoundaryWrapper({ children }: { children: ReactNode }) {
      return (
        <PageErrorBoundary pageName={pageName} fallback={fallback}>
          {children}
        </PageErrorBoundary>
      )
    }
  }

  return { createErrorBoundary }
}
