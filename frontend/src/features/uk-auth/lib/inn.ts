const WEIGHTS_10 = [2, 4, 10, 3, 5, 9, 4, 6, 8] as const
const WEIGHTS_12_FIRST = [7, 2, 4, 10, 3, 5, 9, 4, 6, 8] as const
const WEIGHTS_12_SECOND = [3, 7, 2, 4, 10, 3, 5, 9, 4, 6, 8] as const

const checksum = (digits: readonly number[], weights: readonly number[]): number =>
  (weights.reduce((sum, weight, index) => sum + weight * digits[index], 0) % 11) % 10

export const normalizeInn = (value: string): string => value.replace(/\D/g, '').slice(0, 12)

export const isValidInn = (value: string): boolean => {
  const inn = normalizeInn(value)

  if (inn.length !== 10 && inn.length !== 12) {
    return false
  }

  const digits = [...inn].map(Number)

  if (inn.length === 10) {
    return checksum(digits, WEIGHTS_10) === digits[9]
  }

  return (
    checksum(digits, WEIGHTS_12_FIRST) === digits[10] &&
    checksum(digits, WEIGHTS_12_SECOND) === digits[11]
  )
}

export const isOrganizationInn = (value: string): boolean =>
  normalizeInn(value).length === 10 && isValidInn(value)
