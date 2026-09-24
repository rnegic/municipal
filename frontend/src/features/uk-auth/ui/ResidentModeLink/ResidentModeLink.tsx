import { Link } from 'react-router-dom'

import { IconChevronLeft } from '@/shared/assets/icons'
import { ROUTES } from '@/shared/config/routes'
import { Button } from '@/shared/ui/button'
import { ukAuthTexts as texts } from '../../config/texts'

export const ResidentModeLink = () => (
  <Button asChild size="small" tone="ghost" iconBefore={<IconChevronLeft size={16} />}>
    <Link to={ROUTES.feed}>{texts.residentMode}</Link>
  </Button>
)
