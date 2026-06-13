import { useForm, Controller } from 'react-hook-form'
import { Banner } from '@cloudflare/kumo/components/banner'
import { Button } from '@cloudflare/kumo/components/button'
import { Checkbox } from '@cloudflare/kumo/components/checkbox'
import { Input } from '@cloudflare/kumo/components/input'
import { LayerCard } from '@cloudflare/kumo/components/layer-card'
import { SensitiveInput } from '@cloudflare/kumo/components/sensitive-input'
import { useTranslation } from 'react-i18next'
import { useEffect, useState } from 'react'
import isLength from 'validator/lib/isLength'
import type { LoginCredentials } from '@/services/auth'
import { useAuth } from '@/hooks/use-auth'
import LanguageSwitcher from '@/components/language-switcher'
import GLOBAL_CONFIG from '@/config'
import { init } from '@/main'
import { useNavigate } from '@/hooks/use-navigate'

type AuthMode = 'login' | 'register'

interface LoginForm extends LoginCredentials {
  rememberMe: boolean
  confirmPassword?: string
}

const SAVED_CREDENTIALS_KEY = 'savedCredentials'

export default function Login() {
  const navigate = useNavigate()
  const { t } = useTranslation()
  const { login, register, loading, error, clearError } = useAuth()
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [mode, setMode] = useState<AuthMode>('login')

  const getSavedCredentials = (): Partial<LoginForm> => {
    try {
      const saved = localStorage.getItem(SAVED_CREDENTIALS_KEY)
      return saved ? JSON.parse(saved) : { rememberMe: false }
    } catch {
      return { rememberMe: false }
    }
  }

  const handleCredentialsSave = (data: LoginForm) => {
    if (data.rememberMe) {
      localStorage.setItem(SAVED_CREDENTIALS_KEY, JSON.stringify({
        username: data.username,
        password: data.password,
        rememberMe: true
      }))
    } else {
      localStorage.removeItem(SAVED_CREDENTIALS_KEY)
    }
  }

  const {
    control,
    handleSubmit,
    setValue,
    getValues,
    formState: { errors }
  } = useForm<LoginForm>({
    defaultValues: getSavedCredentials()
  })

  useEffect(() => {
    const savedCredentials = getSavedCredentials()
    if (savedCredentials.username) setValue('username', savedCredentials.username)
    if (savedCredentials.password) setValue('password', savedCredentials.password)
    if (savedCredentials.rememberMe) setValue('rememberMe', savedCredentials.rememberMe)
  }, [setValue])

  const onSubmit = async (data: LoginForm) => {
    try {
      setIsSubmitting(true)
      const credentials = {
        username: data.username,
        password: data.password,
      }

      if (mode === 'register') {
        await register(credentials)
      } else {
        await login(credentials)
        handleCredentialsSave(data)
      }

      init().then(() => navigate('/home'))
    } catch (err) {
      console.error(`${mode === 'register' ? 'Registration' : 'Login'} failed:`, err)
    } finally {
      setIsSubmitting(false)
    }
  }

  const submitting = loading || isSubmitting

  return (
    <div className="grid min-h-screen place-items-center bg-kumo-canvas p-6">
      <LayerCard className="w-full max-w-md p-5">
        <form className="space-y-4" onSubmit={handleSubmit(onSubmit)}>
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 space-y-3">
              <div className="flex items-center gap-2">
                <img src={GLOBAL_CONFIG.LOGO_LIGHT} alt="" className="size-8 rounded" />
                <span className="text-base font-semibold text-kumo-default">
                  {GLOBAL_CONFIG.SYSTEM_NAME_TEXT}
                </span>
              </div>
              <h1 className="text-2xl font-semibold text-kumo-default">
                {mode === 'register' ? t('register.title') : t('login.title')}
              </h1>
            </div>
            <LanguageSwitcher />
          </div>

          {error && (
            <Banner
              variant="error"
              title={error}
              action={(
                <Button type="button" size="sm" variant="secondary-destructive" onClick={clearError}>
                  ×
                </Button>
              )}
            />
          )}

          <Controller
            name="username"
            control={control}
            rules={{
              required: t('login.validation.usernameRequired'),
              validate: (value) => isLength(value || '', { min: 3 }) || t('login.validation.usernameMinLength')
            }}
            render={({ field }) => (
              <Input
                {...field}
                label={t('login.username')}
                placeholder={t('login.usernamePlaceholder')}
                autoComplete="username"
                disabled={submitting}
                size="base"
                error={errors.username?.message}
              />
            )}
          />

          <Controller
            name="password"
            control={control}
            rules={{
              required: t('login.validation.passwordRequired'),
              validate: (value) => {
                const minLength = mode === 'register' ? 8 : 6
                return isLength(value || '', { min: minLength }) ||
                  (mode === 'register'
                    ? t('register.validation.passwordMinLength')
                    : t('login.validation.passwordMinLength'))
              }
            }}
            render={({ field }) => (
              <SensitiveInput
                {...field}
                onValueChange={field.onChange}
                label={t('login.password')}
                placeholder={t('login.passwordPlaceholder')}
                autoComplete={mode === 'register' ? 'new-password' : 'current-password'}
                disabled={submitting}
                size="base"
                error={errors.password?.message}
              />
            )}
          />

          {mode === 'register' && (
            <Controller
              name="confirmPassword"
              control={control}
              rules={{
                required: t('register.validation.confirmPasswordRequired'),
                validate: (value) => value === getValues('password') || t('register.validation.passwordMismatch')
              }}
              render={({ field }) => (
                <SensitiveInput
                  {...field}
                  value={field.value || ''}
                  onValueChange={field.onChange}
                  label={t('register.confirmPassword')}
                  placeholder={t('register.confirmPasswordPlaceholder')}
                  autoComplete="new-password"
                  disabled={submitting}
                  size="base"
                  error={errors.confirmPassword?.message}
                />
              )}
            />
          )}

          {mode === 'login' && (
            <Controller
              name="rememberMe"
              control={control}
              defaultValue={false}
              render={({ field }) => (
                <Checkbox
                  label={t('login.rememberMe')}
                  checked={!!field.value}
                  onCheckedChange={field.onChange}
                  disabled={submitting}
                />
              )}
            />
          )}

          <Button
            type="submit"
            variant="primary"
            size="base"
            className="w-full justify-center"
            loading={submitting}
            disabled={submitting}
          >
            {submitting
              ? (mode === 'register' ? t('register.registering') : t('login.loggingIn'))
              : (mode === 'register' ? t('register.registerButton') : t('login.loginButton'))}
          </Button>

          <div className="text-center">
            <button
              type="button"
              className="text-sm font-medium text-kumo-brand hover:underline"
              onClick={() => {
                const newMode = mode === 'login' ? 'register' : 'login'
                setMode(newMode)
                clearError()
                setValue('confirmPassword', '')
              }}
            >
              {mode === 'login' ? t('login.noAccount') : t('register.hasAccount')}
            </button>
          </div>
        </form>
      </LayerCard>
    </div>
  )
}
