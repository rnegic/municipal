import { useReducer } from 'react'

import {
  INCIDENT_DESCRIPTION_MAX_LENGTH,
  INCIDENT_DESCRIPTION_MIN_LENGTH,
  INCIDENT_TITLE_MAX_LENGTH,
  INCIDENT_TITLE_MIN_LENGTH,
  isPhotoRequiredForCategory,
  isUkAuthority,
  useAnalyzeIncidentMutation,
  type AnalyzeIncidentResult,
  type IncidentAuthority,
  type IncidentCategory,
} from '@/entities/incident'
import type { PhotoPickerItem } from '@/shared/ui/photo-picker'

import { useSubmitIncidentReportMutation } from '../api'
import { incidentReportTexts as texts } from '../config/texts'
import type { IncidentReportStep, IncidentRoutingDecision } from './types'

interface WizardState {
  step: IncidentReportStep
  title: string
  description: string
  entrance: string
  floorZone: string
  photos: PhotoPickerItem[]
  autoDecision: IncidentRoutingDecision | null
  manualAuthority: IncidentAuthority | ''
  manualCategory: IncidentCategory | ''
  isManualRouting: boolean
  validationError: string | null
}

type WizardAction =
  | { type: 'setTitle'; value: string }
  | { type: 'setDescription'; value: string }
  | { type: 'setEntrance'; value: string }
  | { type: 'setFloorZone'; value: string }
  | { type: 'setPhotos'; value: PhotoPickerItem[] }
  | { type: 'setManualAuthority'; value: IncidentAuthority }
  | { type: 'setManualCategory'; value: IncidentCategory }
  | { type: 'reject'; message: string }
  | { type: 'resolveRouting'; decision: IncidentRoutingDecision }
  | { type: 'fallbackRouting' }
  | { type: 'goToDescription' }
  | { type: 'goToRouting' }
  | { type: 'goToPhoto' }

const initialState: WizardState = {
  step: 'description',
  title: '',
  description: '',
  entrance: '',
  floorZone: '',
  photos: [],
  autoDecision: null,
  manualAuthority: '',
  manualCategory: '',
  isManualRouting: false,
  validationError: null,
}

const wizardReducer = (state: WizardState, action: WizardAction): WizardState => {
  switch (action.type) {
    case 'setTitle':
      return { ...state, title: action.value, validationError: null }
    case 'setDescription':
      return { ...state, description: action.value, validationError: null }
    case 'setEntrance':
      return { ...state, entrance: action.value }
    case 'setFloorZone':
      return { ...state, floorZone: action.value }
    case 'setPhotos':
      return { ...state, photos: action.value, validationError: null }
    case 'setManualAuthority':
      return { ...state, manualAuthority: action.value, validationError: null }
    case 'setManualCategory':
      return { ...state, manualCategory: action.value, validationError: null }
    case 'reject':
      return { ...state, validationError: action.message }
    case 'resolveRouting':
      return {
        ...state,
        step: 'routing',
        autoDecision: action.decision,
        isManualRouting: false,
        validationError: null,
      }
    case 'fallbackRouting':
      return {
        ...state,
        step: 'routing',
        autoDecision: null,
        isManualRouting: true,
        validationError: null,
      }
    case 'goToDescription':
      return { ...state, step: 'description', validationError: null }
    case 'goToRouting':
      return { ...state, step: 'routing', validationError: null }
    case 'goToPhoto':
      return { ...state, step: 'photo', validationError: null }
    default:
      return state
  }
}

const toAutoDecision = (result: AnalyzeIncidentResult): IncidentRoutingDecision => ({
  category: result.category,
  authority: result.authority,
  isUkResponsibility: isUkAuthority(result.authority) && result.isUkResponsibility,
  photoRequired: result.photoRequired || isPhotoRequiredForCategory(result.category),
  reasoningText: result.reasoningText,
  source: 'auto',
})

const toManualDecision = (
  authority: IncidentAuthority,
  category: IncidentCategory,
): IncidentRoutingDecision => ({
  category,
  authority,
  isUkResponsibility: isUkAuthority(authority),
  photoRequired: isPhotoRequiredForCategory(category),
  reasoningText: null,
  source: 'manual',
})

export interface UseIncidentReportWizardOptions {
  onSuccess: () => void
}

export const useIncidentReportWizard = ({ onSuccess }: UseIncidentReportWizardOptions) => {
  const [state, dispatch] = useReducer(wizardReducer, initialState)
  const analyzeMutation = useAnalyzeIncidentMutation()
  const submitMutation = useSubmitIncidentReportMutation()

  const decision: IncidentRoutingDecision | null =
    state.autoDecision ??
    (state.manualAuthority && state.manualCategory
      ? toManualDecision(state.manualAuthority, state.manualCategory)
      : null)

  const analyze = () => {
    const title = state.title.trim()
    const description = state.description.trim()

    if (title.length < INCIDENT_TITLE_MIN_LENGTH) {
      dispatch({
        type: 'reject',
        message: texts.errors.titleTooShort(INCIDENT_TITLE_MIN_LENGTH),
      })
      return
    }

    if (title.length > INCIDENT_TITLE_MAX_LENGTH) {
      dispatch({
        type: 'reject',
        message: texts.errors.titleTooLong(INCIDENT_TITLE_MAX_LENGTH),
      })
      return
    }

    if (description.length < INCIDENT_DESCRIPTION_MIN_LENGTH) {
      dispatch({
        type: 'reject',
        message: texts.errors.descriptionTooShort(INCIDENT_DESCRIPTION_MIN_LENGTH),
      })
      return
    }

    if (description.length > INCIDENT_DESCRIPTION_MAX_LENGTH) {
      dispatch({
        type: 'reject',
        message: texts.errors.descriptionTooLong(INCIDENT_DESCRIPTION_MAX_LENGTH),
      })
      return
    }

    analyzeMutation.mutate(
      { description },
      {
        onSuccess: (result) =>
          dispatch({ type: 'resolveRouting', decision: toAutoDecision(result) }),
        onError: () => dispatch({ type: 'fallbackRouting' }),
      },
    )
  }

  const goToPhoto = () => {
    if (!decision) {
      dispatch({ type: 'reject', message: texts.errors.routingRequired })
      return
    }

    if (!decision.isUkResponsibility) {
      return
    }

    dispatch({ type: 'goToPhoto' })
  }

  const submit = () => {
    if (!decision?.isUkResponsibility) {
      return
    }

    if (decision.photoRequired && state.photos.length === 0) {
      dispatch({ type: 'reject', message: texts.errors.photoRequired })
      return
    }

    submitMutation.mutate(
      {
        title: state.title.trim(),
        description: state.description.trim(),
        category: decision.category,
        entrance: state.entrance.trim() || undefined,
        floorZone: state.floorZone.trim() || undefined,
        photos: state.photos.map((item) => item.file),
      },
      { onSuccess },
    )
  }

  return {
    step: state.step,
    title: state.title,
    description: state.description,
    entrance: state.entrance,
    floorZone: state.floorZone,
    photos: state.photos,
    manualAuthority: state.manualAuthority,
    manualCategory: state.manualCategory,
    isManualRouting: state.isManualRouting,
    validationError: state.validationError,
    decision,
    isAnalyzing: analyzeMutation.isPending,
    isSubmitting: submitMutation.isPending,
    submitError: submitMutation.error,
    setTitle: (value: string) => dispatch({ type: 'setTitle', value }),
    setDescription: (value: string) => dispatch({ type: 'setDescription', value }),
    setEntrance: (value: string) => dispatch({ type: 'setEntrance', value }),
    setFloorZone: (value: string) => dispatch({ type: 'setFloorZone', value }),
    setPhotos: (value: PhotoPickerItem[]) => dispatch({ type: 'setPhotos', value }),
    setManualAuthority: (value: IncidentAuthority) =>
      dispatch({ type: 'setManualAuthority', value }),
    setManualCategory: (value: IncidentCategory) => dispatch({ type: 'setManualCategory', value }),
    analyze,
    goToPhoto,
    goToDescription: () => dispatch({ type: 'goToDescription' }),
    goToRouting: () => dispatch({ type: 'goToRouting' }),
    submit,
  }
}

export type IncidentReportWizard = ReturnType<typeof useIncidentReportWizard>
