import { Button } from '@cloudflare/kumo/components/button'
import { MoonIcon, SunIcon } from '@phosphor-icons/react'
import { useTheme } from '@/contexts/theme-context'

export default function ThemeSwitcher() {
  const { theme, toggleTheme } = useTheme()

  return (
    <Button
      type="button"
      size="sm"
      shape="square"
      variant="secondary"
      onClick={toggleTheme}
      icon={theme === 'light' ? <MoonIcon size={16} /> : <SunIcon size={16} />}
      aria-label="Toggle theme"
    />
  )
}
