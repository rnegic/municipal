import type { IncidentPhoto } from '../../model/types'
import s from './IncidentPhotos.module.scss'

export interface IncidentPhotosProps {
  photos: readonly IncidentPhoto[]
  className?: string
}

const MAX_THUMBNAILS = 3

export const IncidentPhotos = ({ photos, className }: IncidentPhotosProps) => {
  if (photos.length === 0) {
    return null
  }

  const visiblePhotos = photos.slice(0, MAX_THUMBNAILS)
  const remainingCount = photos.length - visiblePhotos.length

  return (
    <div className={`${s.root}${className ? ` ${className}` : ''}`} aria-label="Фото обращения">
      {visiblePhotos.map((photo, index) => (
        <div className={s.item} key={photo.id}>
          <img
            className={s.image}
            src={photo.url}
            alt={`Фото обращения ${index + 1}`}
            loading="lazy"
            decoding="async"
          />
          {index === visiblePhotos.length - 1 && remainingCount > 0 ? (
            <span className={s.more}>+{remainingCount}</span>
          ) : null}
        </div>
      ))}
    </div>
  )
}
