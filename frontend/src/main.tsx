import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'

import { App } from '@/app'
import { normalizeLaunchHash } from '@/shared/lib/max'
import '@/app/styles/index.scss'

normalizeLaunchHash()

const container = document.getElementById('root')

if (!container) {
  throw new Error('Root container #root is missing in index.html')
}

createRoot(container).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
