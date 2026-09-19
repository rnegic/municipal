import type { CSSProperties } from 'react'

export type CssVarName = `--${string}`

export type CssVarValue = string | number | undefined
export const cssVars = (vars: Partial<Record<CssVarName, CssVarValue>>): CSSProperties => {
  const style: Record<string, string | number> = {}

  for (const [name, value] of Object.entries(vars)) {
    if (value !== undefined) {
      style[name] = value
    }
  }

  return style as CSSProperties
}
