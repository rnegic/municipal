import type { IncidentCategory } from '../model/types'

export interface IncidentLegalTerm {
  days: number
  businessDays: boolean
  scope: string
  law: string
}

export const INCIDENT_LEGAL_TERMS: Record<IncidentCategory, IncidentLegalTerm> = {
  WATER_HEAT: {
    days: 3,
    businessDays: false,
    scope: 'На устранение аварии с водой или отоплением',
    law: 'Постановление Правительства РФ № 416, п. 18',
  },
  ELECTRICITY: {
    days: 3,
    businessDays: false,
    scope: 'На устранение аварии в электросети дома',
    law: 'Постановление Правительства РФ № 416, п. 18',
  },
  ELEVATOR: {
    days: 1,
    businessDays: false,
    scope: 'На ремонт неисправного лифта',
    law: 'Правила № 170, приложение 2',
  },
  BUILDING_STRUCTURE: {
    days: 5,
    businessDays: false,
    scope: 'На ремонт кровли, фасада и водостоков',
    law: 'Правила № 170, приложение 2',
  },
  CLEANING_YARD: {
    days: 10,
    businessDays: true,
    scope: 'На ответ управляющей компании по обращению',
    law: 'Постановление Правительства РФ № 416',
  },
  CITY_TERRITORY: {
    days: 30,
    businessDays: false,
    scope: 'На ответ ведомства по обращению жителя',
    law: 'Федеральный закон № 59-ФЗ, ст. 12',
  },
}

export const INCIDENT_LEGAL_TERM_DEFAULT: IncidentLegalTerm = {
  days: 10,
  businessDays: true,
  scope: 'На ответ управляющей компании по обращению',
  law: 'Постановление Правительства РФ № 416',
}
