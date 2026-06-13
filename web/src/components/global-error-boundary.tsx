import { type ErrorInfo, type ReactNode } from 'react'
import { ErrorBoundary, type FallbackProps } from 'react-error-boundary'
import { Button } from '@cloudflare/kumo/components/button'
import { LayerCard } from '@cloudflare/kumo/components/layer-card'
import { useTranslation } from 'react-i18next'

function formatError(error: unknown) {
  return error instanceof Error ? error : new Error(String(error))
}

function GlobalErrorFallback({ error }: FallbackProps) {
  const { t } = useTranslation()
  const normalizedError = formatError(error)
  const isDev = import.meta.env.DEV

  return (
    <div className="flex min-h-screen items-center justify-center bg-kumo-base p-6">
      <LayerCard className="w-full max-w-2xl p-6">
        <div className="space-y-5 text-center">
          <div className="space-y-2">
            <div className="text-5xl font-semibold text-kumo-danger">500</div>
            <h1 className="text-2xl font-semibold text-kumo-default">{t('error.global.title')}</h1>
            <p className="text-kumo-subtle">{t('error.global.subtitle')}</p>
          </div>
          <div className="flex flex-wrap justify-center gap-2">
            <Button type="button" variant="primary" onClick={() => window.location.reload()}>
              {t('error.global.reload')}
            </Button>
            <Button type="button" onClick={() => { window.location.href = '/home' }}>
              {t('error.global.goHome')}
            </Button>
          </div>
          {isDev && (
            <div className="space-y-3 rounded-lg border border-kumo-danger/30 bg-kumo-danger/5 p-4 text-left">
              <div>
                <div className="text-sm font-semibold text-kumo-danger">Error Message</div>
                <code className="text-sm text-kumo-danger">{normalizedError.message}</code>
              </div>
              {normalizedError.stack && (
                <pre className="max-h-48 overflow-auto rounded bg-kumo-base p-3 text-xs text-kumo-default">
                  {normalizedError.stack}
                </pre>
              )}
            </div>
          )}
        </div>
      </LayerCard>
    </div>
  )
}

const handleGlobalError = (error: unknown, errorInfo: ErrorInfo) => {
  console.error('Global Error Boundary caught an error:', error, errorInfo)
}

export default function GlobalErrorBoundary({ children }: { children: ReactNode }) {
  return (
    <ErrorBoundary FallbackComponent={GlobalErrorFallback} onError={handleGlobalError}>
      {children}
    </ErrorBoundary>
  )
}
