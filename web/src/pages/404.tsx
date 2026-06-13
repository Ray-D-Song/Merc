import { Button } from '@cloudflare/kumo/components/button'
import { LayerCard } from '@cloudflare/kumo/components/layer-card'
import { useTranslation } from 'react-i18next'
import { useNavigate } from '@/hooks/use-navigate'

export default function NotFound() {
  const navigate = useNavigate()
  const { t } = useTranslation()

  return (
    <div className="flex min-h-[calc(100svh-154px)] items-center justify-center bg-kumo-base p-6">
      <LayerCard className="w-full max-w-lg p-6 text-center">
        <div className="space-y-5">
          <div className="space-y-2">
            <h1 className="text-5xl font-semibold text-kumo-info">{t('error.404.title')}</h1>
            <p className="text-kumo-subtle">{t('error.404.subtitle')}</p>
          </div>
          <div className="flex justify-center gap-2">
            <Button type="button" variant="primary" onClick={() => navigate('/home')}>
              {t('error.404.goHome')}
            </Button>
            <Button type="button" onClick={() => navigate(-1)}>
              {t('error.404.goBack')}
            </Button>
          </div>
        </div>
      </LayerCard>
    </div>
  )
}
