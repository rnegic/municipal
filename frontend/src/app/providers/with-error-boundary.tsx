import { Component, type ErrorInfo } from 'react'

import { ErrorIllustration } from '@/shared/assets/illustrations'
import { commonTexts } from '@/shared/config/texts'
import { Button } from '@/shared/ui/button'
import { EmptyState } from '@/shared/ui/empty-state'
import { PageLayout } from '@/shared/ui/page-layout'

import { createProvider, type ProviderWrapperProps } from './create-provider'

interface ErrorBoundaryState {
  hasError: boolean
}

class ErrorBoundary extends Component<ProviderWrapperProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = { hasError: false }

  static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true }
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    console.error(error, info.componentStack)
  }

  private readonly handleRetry = (): void => {
    this.setState({ hasError: false })
  }

  render() {
    if (!this.state.hasError) {
      return this.props.children
    }

    return (
      <PageLayout>
        <EmptyState
          tone="danger"
          illustration={<ErrorIllustration />}
          title={commonTexts.errors.unexpectedTitle}
          description={commonTexts.errors.unexpectedDescription}
          action={
            <Button tone="secondary" onClick={this.handleRetry}>
              {commonTexts.actions.tryAgain}
            </Button>
          }
        />
      </PageLayout>
    )
  }
}

export const WithErrorBoundary = createProvider(ErrorBoundary)
