import { IconShield } from '@/shared/assets/icons'
import { Button } from '@/shared/ui/button'
import { ukAuthTexts as texts } from '../../config/texts'
import s from './UkAuthEntryLink.module.scss'

export interface UkAuthEntryLinkProps {
  onClick: () => void
}

export const UkAuthEntryLink = ({ onClick }: UkAuthEntryLinkProps) => (
  <Button
    className={s.root}
    type="button"
    size="small"
    tone="ghost"
    iconBefore={<IconShield size={14} />}
    onClick={onClick}
  >
    {texts.entry}
  </Button>
)
