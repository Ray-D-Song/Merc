import { isMobile } from '@/utils/assert'
import { type ReactNode } from 'react'

export default function MobileSwitcher(mobileComp: ReactNode, elseComp: ReactNode) {
  return isMobile() ? mobileComp : elseComp
}