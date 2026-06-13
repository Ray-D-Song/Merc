import { responseHandler } from './utils/response'

const GLOBAL_CONFIG = {
  SYSTEM_NAME_TEXT: 'Merc',
  AUTHER: 'Ray-D-Song',
  LOGO_LIGHT: '/logo.svg',
  LOGO_DARK: '/logo.svg',

  // request
  BASE_URL: '/api',
  RESPONSE_HANDLER: responseHandler

} as const

export default GLOBAL_CONFIG
