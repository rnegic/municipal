export interface GeoPoint {
  lat: number
  lon: number
}

export type GeolocationErrorCode = 'unsupported' | 'denied' | 'unavailable' | 'timeout'

export class GeolocationError extends Error {
  readonly code: GeolocationErrorCode

  constructor(code: GeolocationErrorCode, message: string) {
    super(message)
    this.name = 'GeolocationError'
    this.code = code
  }
}

const ERROR_CODE_BY_POSITION_CODE: Record<number, GeolocationErrorCode> = {
  1: 'denied',
  2: 'unavailable',
  3: 'timeout',
}

const GEOLOCATION_OPTIONS: PositionOptions = {
  enableHighAccuracy: true,
  timeout: 10_000,
  maximumAge: 60_000,
}

export const getCurrentPosition = (): Promise<GeoPoint> =>
  new Promise((resolve, reject) => {
    if (typeof navigator === 'undefined' || !navigator.geolocation) {
      reject(new GeolocationError('unsupported', 'Geolocation is not supported'))
      return
    }

    navigator.geolocation.getCurrentPosition(
      ({ coords }) => resolve({ lat: coords.latitude, lon: coords.longitude }),
      (error) =>
        reject(new GeolocationError(ERROR_CODE_BY_POSITION_CODE[error.code] ?? 'unavailable', error.message)),
      GEOLOCATION_OPTIONS,
    )
  })
