import type { ComponentType } from 'react'

import {
  BuildingStructureCategoryIcon,
  CityTerritoryCategoryIcon,
  CleaningYardCategoryIcon,
  DefaultCategoryIcon,
  ElectricityCategoryIcon,
  ElevatorCategoryIcon,
  WaterCategoryIcon,
} from '@/shared/assets/category-icons'
import { cn } from '@/shared/lib/cn'
import type { IncidentCategory } from '../../model/types'
import s from './CategoryIcon.module.scss'

export interface CategoryIconProps {
  category?: IncidentCategory | null
  size?: number
  className?: string
}

const CATEGORY_TO_ICON: Record<IncidentCategory, ComponentType> = {
  WATER_HEAT: WaterCategoryIcon,
  ELECTRICITY: ElectricityCategoryIcon,
  ELEVATOR: ElevatorCategoryIcon,
  CLEANING_YARD: CleaningYardCategoryIcon,
  BUILDING_STRUCTURE: BuildingStructureCategoryIcon,
  CITY_TERRITORY: CityTerritoryCategoryIcon,
}

export const CategoryIcon = ({ category, size = 44, className }: CategoryIconProps) => {
  const Icon = (category && CATEGORY_TO_ICON[category]) || DefaultCategoryIcon

  return (
    <span
      className={cn(s.root, className)}
      style={{ width: size, height: size }}
      aria-hidden="true"
    >
      <Icon />
    </span>
  )
}
