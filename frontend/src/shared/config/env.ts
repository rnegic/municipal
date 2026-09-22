export const env = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? '/api',
  maxBotName: import.meta.env.VITE_MAX_BOT_NAME ?? 't117_hakaton_max_bot',
  isDev: import.meta.env.DEV,
} as const
