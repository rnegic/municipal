import type { House } from '@/entities/house'
import { BottomSheet } from '@/shared/ui/bottom-sheet'
import { eventCreateTexts } from '../../config/texts'
import { EventCreateForm } from '../EventCreateForm'

export interface EventCreateSheetProps {
  open: boolean
  onClose: () => void
  houses: readonly House[]
}

export const EventCreateSheet = ({ open, onClose, houses }: EventCreateSheetProps) => (
  <BottomSheet
    open={open}
    title={eventCreateTexts.title}
    description={eventCreateTexts.description}
    onClose={onClose}
  >
    <EventCreateForm houses={houses} onSuccess={onClose} onCancel={onClose} />
  </BottomSheet>
)
