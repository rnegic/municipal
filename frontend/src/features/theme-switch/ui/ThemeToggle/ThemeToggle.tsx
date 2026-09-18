import { IconButton } from '@maxhub/max-ui'

import { IconMoon, IconSun } from '@/shared/assets/icons'
import { themeTexts } from '../../config/texts'
import { useTheme } from '../../model/theme-context'
export const ThemeToggle = () => {
  const { colorScheme, toggleColorScheme } = useTheme()

  return (
    <IconButton
      size="small"
      variant="secondary"
      aria-label={colorScheme === 'dark' ? themeTexts.toLight : themeTexts.toDark}
      onClick={toggleColorScheme}
    >
      {colorScheme === 'dark' ? <IconSun /> : <IconMoon />}
    </IconButton>
  )
}
