import { Typography } from '@maxhub/max-ui'

import { cn } from '@/shared/lib/cn'
import type { AddressSuggestion } from '@/shared/lib/dadata'
import { addressBindTexts } from '../../config/texts'
import s from './AddressSuggestions.module.scss'

export interface AddressSuggestionsProps {
  suggestions: AddressSuggestion[]
  activeIndex: number
  listId?: string
  onSelect: (value: string) => void
  onHighlight: (index: number) => void
}

export const AddressSuggestions = ({
  suggestions,
  activeIndex,
  listId,
  onSelect,
  onHighlight,
}: AddressSuggestionsProps) => (
  <ul className={s.root} id={listId} role="listbox" aria-label={addressBindTexts.suggestionsLabel}>
    {suggestions.map((suggestion, index) => (
      <li key={suggestion.value}>
        <button
          type="button"
          role="option"
          aria-selected={index === activeIndex}
          className={cn(s.option, index === activeIndex && s.optionActive)}
          onMouseDown={(event) => event.preventDefault()}
          onMouseEnter={() => onHighlight(index)}
          onClick={() => onSelect(suggestion.value)}
        >
          <Typography.Text variant="description">{suggestion.value}</Typography.Text>
        </button>
      </li>
    ))}
  </ul>
)