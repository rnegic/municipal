import { useEffect, useRef, useState, type ChangeEvent } from 'react'

import { Icon16CloseIos, Icon24CloseAndroid, IconButton, Typography, usePlatform } from '@maxhub/max-ui'

import { commonTexts } from '@/shared/config/texts'
import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button'
import s from './PhotoPicker.module.scss'

const BYTES_IN_MEGABYTE = 1024 * 1024

export interface PhotoPickerItem {
  file: File
  previewUrl: string
}

export interface PhotoPickerProps {
  items: PhotoPickerItem[]
  onChange: (items: PhotoPickerItem[]) => void
  maxFiles: number
  maxBytes: number
  acceptedMimeTypes: readonly string[]
  disabled?: boolean
  className?: string
}

const toMegabytes = (bytes: number): number => Math.round(bytes / BYTES_IN_MEGABYTE)

const isSameFile = (left: File, right: File): boolean =>
  left.name === right.name && left.size === right.size && left.lastModified === right.lastModified

export const PhotoPicker = ({
  items,
  onChange,
  maxFiles,
  maxBytes,
  acceptedMimeTypes,
  disabled = false,
  className,
}: PhotoPickerProps) => {
  const inputRef = useRef<HTMLInputElement>(null)
  const [error, setError] = useState<string | null>(null)
  const itemsRef = useRef<PhotoPickerItem[]>(items)
  const platform = usePlatform()

  useEffect(() => {
    itemsRef.current = items
  }, [items])

  useEffect(
    () => () => {
      itemsRef.current.forEach((item) => URL.revokeObjectURL(item.previewUrl))
    },
    [],
  )

  const isLimitReached = items.length >= maxFiles

  const handleSelect = (event: ChangeEvent<HTMLInputElement>) => {
    const selected = Array.from(event.target.files ?? [])
    event.target.value = ''

    if (selected.length === 0) {
      return
    }

    if (selected.some((file) => !acceptedMimeTypes.includes(file.type))) {
      setError(commonTexts.photoPicker.typeError)
      return
    }

    if (selected.some((file) => file.size > maxBytes)) {
      setError(commonTexts.photoPicker.sizeError(toMegabytes(maxBytes)))
      return
    }

    const accepted = selected.filter(
      (file, index) =>
        !items.some((item) => isSameFile(item.file, file)) &&
        selected.findIndex((candidate) => isSameFile(candidate, file)) === index,
    )

    const freeSlots = Math.max(maxFiles - items.length, 0)
    const isOverflowing = accepted.length > freeSlots

    const added = accepted.slice(0, freeSlots).map((file) => ({
      file,
      previewUrl: URL.createObjectURL(file),
    }))

    setError(isOverflowing ? commonTexts.photoPicker.countError(maxFiles) : null)

    if (added.length > 0) {
      onChange([...items, ...added])
    }
  }

  const handleRemove = (index: number) => {
    const removed = items[index]

    if (removed) {
      URL.revokeObjectURL(removed.previewUrl)
    }

    setError(null)
    onChange(items.filter((_item, itemIndex) => itemIndex !== index))
  }

  return (
    <div className={cn(s.root, className)}>
      {items.length > 0 ? (
        <ul className={s.list} aria-label={commonTexts.a11y.photoList}>
          {items.map((item, index) => (
            <li key={item.previewUrl} className={s.item}>
              <img
                className={s.preview}
                src={item.previewUrl}
                alt={commonTexts.a11y.photoPreview(index + 1)}
              />
              <IconButton
                className={s.remove}
                size="small"
                variant="secondary"
                aria-label={commonTexts.photoPicker.remove}
                disabled={disabled}
                onClick={() => handleRemove(index)}
              >
                {platform === 'ios' ? <Icon16CloseIos /> : <Icon24CloseAndroid />}
              </IconButton>
            </li>
          ))}
        </ul>
      ) : null}

      <div className={s.picker}>
        <input
          ref={inputRef}
          className={s.input}
          type="file"
          multiple
          tabIndex={-1}
          aria-hidden="true"
          accept={acceptedMimeTypes.join(',')}
          onClick={(event) => event.stopPropagation()}
          onChange={handleSelect}
        />
        <Button
          type="button"
          tone="secondary"
          disabled={disabled || isLimitReached}
          onClick={() => inputRef.current?.click()}
        >
          {isLimitReached
            ? commonTexts.photoPicker.limitReached
            : items.length > 0
              ? commonTexts.photoPicker.addMore
              : commonTexts.photoPicker.add}
        </Button>
        <Typography.Text variant="note" color="tertiary">
          {commonTexts.photoPicker.hint(maxFiles, toMegabytes(maxBytes))}
        </Typography.Text>
      </div>

      {error ? (
        <span className={s.error} role="alert">
          <Typography.Text variant="note" color="inherit">
            {error}
          </Typography.Text>
        </span>
      ) : null}
    </div>
  )
}
