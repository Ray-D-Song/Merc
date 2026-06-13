import { Breadcrumbs } from '@cloudflare/kumo/components/breadcrumbs'
import { Button } from '@cloudflare/kumo/components/button'
import { DropdownMenu } from '@cloudflare/kumo/components/dropdown'
import { Sidebar } from '@cloudflare/kumo/components/sidebar'
import { useLocation } from 'preact-iso'
import { SignOutIcon, UserIcon } from '@phosphor-icons/react'
import { useMemo, type ReactNode } from 'react'
import { useRouterContext } from '@/contexts/router-context'
import ThemeSwitcher from '@/components/theme-switcher'
import LanguageSwitcher from '@/components/language-switcher'
import GLOBAL_CONFIG from '@/config'
import { useAuth } from '@/hooks/use-auth'
import { useNavigate } from '@/hooks/use-navigate'
import type { NavRoute } from '@/utils/router'

function filterNavRoutes(items: NavRoute[] | undefined, isAdmin: () => boolean): NavRoute[] {
  if (!items) return []

  return items
    .filter(item => item.auth !== 'admin' || isAdmin())
    .map(item => ({
      ...item,
      children: filterNavRoutes(item.children, isAdmin),
    }))
    .filter(item => !item.children || item.children.length > 0 || item.title)
}

export default function MainLayout({ children }: { children: ReactNode }) {
  const location = useLocation()
  const navigate = useNavigate()
  const { logout, isAdmin } = useAuth()
  const { getBreadcrumbs, navRoute } = useRouterContext()

  const filteredNav = useMemo(() => filterNavRoutes(navRoute.children, isAdmin), [navRoute.children, isAdmin])
  const breadcrumbs = getBreadcrumbs(location.path)
  const shouldShowHomeLink = location.path !== '/home'

  const handleLogout = async () => {
    await logout()
    navigate('/login')
  }

  return (
    <Sidebar.Provider defaultOpen collapsible="icon" className="h-svh min-h-svh bg-kumo-canvas">
      <Sidebar className="h-svh min-h-svh">
        <Sidebar.Header>
          <div className="flex w-full items-center justify-center gap-2 px-0 py-2 group-not-data-[state=collapsed]/sidebar:justify-start group-not-data-[state=collapsed]/sidebar:px-2">
            <img src={GLOBAL_CONFIG.LOGO_LIGHT} alt="" className="size-8 rounded" />
            <div className="hidden min-w-0 text-base font-semibold text-kumo-default group-not-data-[state=collapsed]/sidebar:block">
              {GLOBAL_CONFIG.SYSTEM_NAME_TEXT}
            </div>
          </div>
        </Sidebar.Header>
        <Sidebar.Content className="min-h-0 flex-1">
          <Sidebar.Group>
            <Sidebar.GroupLabel>Workspace</Sidebar.GroupLabel>
            <Sidebar.Menu>
              {filteredNav.map(item => (
                <Sidebar.MenuButton
                  key={item.path}
                  icon={item.icon}
                  active={location.path === item.path}
                  tooltip={item.title}
                  onClick={() => navigate(item.path)}
                >
                  {item.title}
                </Sidebar.MenuButton>
              ))}
            </Sidebar.Menu>
          </Sidebar.Group>
        </Sidebar.Content>
        {/* <Sidebar.Footer>
          <div className="flex items-center justify-between gap-2 px-2 py-2 text-xs text-kumo-subtle">
            <Sidebar.Trigger aria-label="Toggle sidebar" />
            <span className="truncate px-2">© 2026</span>
          </div>
        </Sidebar.Footer> */}
        <Sidebar.Rail aria-label="Toggle sidebar" />
      </Sidebar>

      <main className="min-h-svh min-w-0 flex-1 bg-kumo-canvas">
        <header className="sticky top-0 z-10 flex h-[58px] items-center justify-between gap-3 border-b border-kumo-line bg-kumo-canvas/90 px-5 backdrop-blur md:px-8">
          <Breadcrumbs size="sm" className="min-w-0">
            {shouldShowHomeLink ? <Breadcrumbs.Link href="/home">首页</Breadcrumbs.Link> : <></>}
            {breadcrumbs.map((item, index) => (
              <span key={item.path} className="contents">
                {shouldShowHomeLink || index > 0 ? <Breadcrumbs.Separator /> : <></>}
                {index === breadcrumbs.length - 1 ? (
                  <Breadcrumbs.Current>{item.title}</Breadcrumbs.Current>
                ) : (
                  <Breadcrumbs.Link href={item.path}>{item.title}</Breadcrumbs.Link>
                )}
              </span>
            ))}
          </Breadcrumbs>

          <div className="flex shrink-0 items-center gap-2">
            <ThemeSwitcher />
            <LanguageSwitcher />
            <DropdownMenu>
              <DropdownMenu.Trigger>
                <Button type="button" shape="square" size="sm" icon={<UserIcon size={16} />} aria-label="User menu" />
              </DropdownMenu.Trigger>
              <DropdownMenu.Content>
                <DropdownMenu.Item icon={<SignOutIcon size={16} />} onClick={handleLogout}>
                  Logout
                </DropdownMenu.Item>
              </DropdownMenu.Content>
            </DropdownMenu>
          </div>
        </header>
        <div className="mx-auto w-full p-6">
          {children}
        </div>
      </main>
    </Sidebar.Provider>
  )
}
