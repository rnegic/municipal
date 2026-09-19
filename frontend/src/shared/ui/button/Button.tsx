import { Button as MaxUiButton, type ButtonProps as MaxUiButtonProps } from '@maxhub/max-ui'

export type ButtonTone = 'primary' | 'secondary' | 'ghost' | 'danger'

export type ButtonSize = 'small' | 'medium' | 'large'

export interface ButtonProps extends Omit<MaxUiButtonProps, 'variant' | 'size'> {
  tone?: ButtonTone
  size?: ButtonSize
}

const TONE_TO_VARIANT: Record<ButtonTone, MaxUiButtonProps['variant']> = {
  primary: 'primary',
  secondary: 'secondary',
  ghost: 'ghost',
  danger: 'destructive',
}

export const Button = ({ tone = 'primary', size = 'large', ...rest }: ButtonProps) => (
  <MaxUiButton variant={TONE_TO_VARIANT[tone]} size={size} {...rest} />
)
