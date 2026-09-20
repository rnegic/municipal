import { BottomSheet } from '@/shared/ui/bottom-sheet'
import { incidentReportTexts } from '../../config/texts'
import { IncidentReportForm } from '../IncidentReportForm'

export interface IncidentReportSheetProps {
  open: boolean
  onClose: () => void
}

export const IncidentReportSheet = ({ open, onClose }: IncidentReportSheetProps) => (
  <BottomSheet
    open={open}
    title={incidentReportTexts.title}
    description={incidentReportTexts.description}
    onClose={onClose}
  >
    <IncidentReportForm onSuccess={onClose} onCancel={onClose} />
  </BottomSheet>
)
