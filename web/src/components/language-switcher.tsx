import { Button } from '@cloudflare/kumo/components/button'
import { DropdownMenu } from '@cloudflare/kumo/components/dropdown'
import { useTranslation } from 'react-i18next'

export default function LanguageSwitcher() {
  const { i18n } = useTranslation()

  const changeLanguage = (lng: string) => {
    i18n.changeLanguage(lng)
  }

  const currentLanguage = i18n.language === 'zh' ? '中文' : 'English'

  return (
    <DropdownMenu>
      <DropdownMenu.Trigger>
        <Button type="button" size="sm" variant="secondary">
          {currentLanguage}
        </Button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Content>
        <DropdownMenu.Item selected={i18n.language !== 'zh'} onClick={() => changeLanguage('en')}>
          English
        </DropdownMenu.Item>
        <DropdownMenu.Item selected={i18n.language === 'zh'} onClick={() => changeLanguage('zh')}>
          中文
        </DropdownMenu.Item>
      </DropdownMenu.Content>
    </DropdownMenu>
  )
}
