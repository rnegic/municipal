import {
  useCallback,
  useEffect,
  useId,
  useRef,
  useState,
  type CSSProperties,
  type KeyboardEvent,
  type ReactNode,
} from 'react'

import { Icon16Chevron, Typography } from '@maxhub/max-ui'

import { commonTexts } from '@/shared/config/texts'
import { cn } from '@/shared/lib/cn'
import s from './Select.module.scss'

export interface SelectOption<TValue extends string = string> {
  value: TValue
  label: string
  description?: string
}

export interface SelectProps<TValue extends string = string> {
  label: string
  options: ReadonlyArray<SelectOption<TValue>>
  value: TValue | ''
  onValueChange: (value: TValue) => void
  placeholder?: string
  hint?: ReactNode
  invalid?: boolean
  disabled?: boolean
  name?: string
  className?: string
}

interface ListCoords {
  top: number
  left: number
  width: number
}

const LIST_FLIP_THRESHOLD = 280
const LIST_MAX_HEIGHT = 320
const LIST_MIN_HEIGHT = 180
const LIST_GAP = 8
const LIST_EDGE_MARGIN = 12

export const Select = <TValue extends string = string>({
  label,
  options,
  value,
  onValueChange,
  placeholder = commonTexts.select.placeholder,
  hint,
  invalid = false,
  disabled = false,
  className,
}: SelectProps<TValue>) => {
  const [open, setOpen] = useState(false)
  const [activeIndex, setActiveIndex] = useState(-1)
  const [maxListHeight, setMaxListHeight] = useState(LIST_MAX_HEIGHT)
  const [coords, setCoords] = useState<ListCoords>({ top: 0, left: 0, width: 0 })

  const anchorRef = useRef<HTMLDivElement>(null)
  const triggerRef = useRef<HTMLButtonElement>(null)
  const listRef = useRef<HTMLUListElement>(null)

  const labelId = useId()
  const listboxId = useId()
  const hintId = useId()

  const selectedIndex = options.findIndex((option) => option.value === value)
  const selectedOption = selectedIndex >= 0 ? options[selectedIndex] : undefined

  const updatePosition = useCallback(() => {
    const trigger = triggerRef.current

    if (!trigger) {
      return
    }

    const rect = trigger.getBoundingClientRect()
    const spaceBelow = window.innerHeight - rect.bottom - LIST_GAP - LIST_EDGE_MARGIN
    const spaceAbove = rect.top - LIST_GAP - LIST_EDGE_MARGIN
    const openUpwards = spaceBelow < LIST_FLIP_THRESHOLD && spaceAbove > spaceBelow
    const availableSpace = openUpwards ? spaceAbove : spaceBelow
    const height = Math.max(LIST_MIN_HEIGHT, Math.min(LIST_MAX_HEIGHT, availableSpace))

    setMaxListHeight(height)
    setCoords({
      top: openUpwards ? rect.top - LIST_GAP - height : rect.bottom + LIST_GAP,
      left: rect.left,
      width: rect.width,
    })
  }, [])

  const openList = () => {
    updatePosition()
    setActiveIndex(selectedIndex >= 0 ? selectedIndex : 0)
    setOpen(true)
  }

  const closeList = (restoreFocus = false) => {
    setOpen(false)

    if (restoreFocus) {
      triggerRef.current?.focus()
    }
  }

  const selectOption = (index: number) => {
    const option = options[index]
    if (!option) {
      return
    }

    onValueChange(option.value)
    closeList(true)
  }

  useEffect(() => {
    if (!open) {
      return undefined
    }

    const handlePointerDown = (event: PointerEvent) => {
      if (!anchorRef.current?.contains(event.target as Node)) {
        setOpen(false)
      }
    }

    document.addEventListener('pointerdown', handlePointerDown)

    return () => document.removeEventListener('pointerdown', handlePointerDown)
  }, [open])

  useEffect(() => {
    if (!open) {
      return undefined
    }

    const handleReposition = () => updatePosition()

    window.addEventListener('resize', handleReposition)
    window.addEventListener('scroll', handleReposition, true)

    return () => {
      window.removeEventListener('resize', handleReposition)
      window.removeEventListener('scroll', handleReposition, true)
    }
  }, [open, updatePosition])

  useEffect(() => {
    if (!open) {
      return
    }

    const list = listRef.current
    const active = list?.querySelector<HTMLElement>(`[data-index="${activeIndex}"]`)

    if (!list || !active) {
      return
    }

    const listRect = list.getBoundingClientRect()
    const activeRect = active.getBoundingClientRect()

    if (activeRect.top < listRect.top) {
      list.scrollTop -= listRect.top - activeRect.top
    } else if (activeRect.bottom > listRect.bottom) {
      list.scrollTop += activeRect.bottom - listRect.bottom
    }
  }, [open, activeIndex])

  const handleTriggerKeyDown = (event: KeyboardEvent<HTMLButtonElement>) => {
    switch (event.key) {
      case 'ArrowDown':
      case 'ArrowUp': {
        event.preventDefault()

        if (!open) {
          openList()
          return
        }

        const step = event.key === 'ArrowDown' ? 1 : -1
        setActiveIndex((prev) => (prev + step + options.length) % options.length)
        return
      }
      case 'Home':
        if (open) {
          event.preventDefault()
          setActiveIndex(0)
        }

        return
      case 'End':
        if (open) {
          event.preventDefault()
          setActiveIndex(options.length - 1)
        }

        return
      case 'Enter':
      case ' ': {
        event.preventDefault()

        if (open) {
          selectOption(activeIndex)
        } else {
          openList()
        }

        return
      }
      case 'Escape':
        if (open) {
          event.preventDefault()
          closeList()
        }

        return
      case 'Tab':
        if (open) {
          setOpen(false)
        }

        return
      default:
    }
  }

  const activeOptionId = open && activeIndex >= 0 ? `${listboxId}-option-${activeIndex}` : undefined
  const listStyle: CSSProperties = {
    top: coords.top,
    left: coords.left,
    width: coords.width,
    maxHeight: maxListHeight,
  }

  return (
    <div className={cn(s.root, className)}>
      <span id={labelId} className={s.label}>
        <Typography.Text variant="description" color="secondary">
          {label}
        </Typography.Text>
      </span>

      <div ref={anchorRef} className={s.anchor}>
        <button
          ref={triggerRef}
          type="button"
          role="combobox"
          className={cn(s.control, open && s.controlOpen, invalid && s.invalid)}
          aria-haspopup="listbox"
          aria-expanded={open}
          aria-controls={open ? listboxId : undefined}
          aria-activedescendant={activeOptionId}
          aria-labelledby={labelId}
          aria-invalid={invalid}
          aria-describedby={hint ? hintId : undefined}
          disabled={disabled}
          onClick={() => (open ? closeList() : openList())}
          onKeyDown={handleTriggerKeyDown}
        >
          <span className={cn(s.value, !selectedOption && s.placeholder)}>
            {selectedOption ? selectedOption.label : placeholder}
          </span>
          <span className={cn(s.chevron, open && s.chevronOpen)} aria-hidden="true">
            <Icon16Chevron />
          </span>
        </button>

        {open ? (
          <ul
            ref={listRef}
            id={listboxId}
            role="listbox"
            className={s.list}
            style={listStyle}
            aria-labelledby={labelId}
            tabIndex={-1}
          >
            {options.map((option, index) => {
              const isSelected = option.value === value
              const isActive = index === activeIndex

              return (
                <li
                  key={option.value}
                  id={`${listboxId}-option-${index}`}
                  role="option"
                  aria-selected={isSelected}
                  data-index={index}
                  className={cn(s.option, isActive && s.optionActive, isSelected && s.optionSelected)}
                  onPointerDown={(event) => event.preventDefault()}
                  onMouseEnter={() => setActiveIndex(index)}
                  onClick={() => selectOption(index)}
                >
                  <Typography.Text variant="body">{option.label}</Typography.Text>
                  {option.description ? (
                    <Typography.Text variant="note" color="tertiary">
                      {option.description}
                    </Typography.Text>
                  ) : null}
                </li>
              )
            })}
          </ul>
        ) : null}
      </div>

      {hint ? (
        <span id={hintId} className={cn(s.hint, invalid && s.hintInvalid)}>
          <Typography.Text variant="note" color="inherit">
            {hint}
          </Typography.Text>
        </span>
      ) : null}
    </div>
  )
}
