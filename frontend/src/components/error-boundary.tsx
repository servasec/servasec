"use client"

import { Component, type ErrorInfo, type ReactNode } from "react"
import { ShieldAlert } from "lucide-react"
import { Button } from "@/components/ui/button"

interface Props {
  children: ReactNode
  fallback?: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error("ErrorBoundary caught:", error, errorInfo)
  }

  handleRetry = () => {
    this.setState({ hasError: false, error: null })
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <ShieldAlert className="h-10 w-10 mb-3 text-destructive opacity-60" />
          <p className="text-sm font-medium text-foreground">Something went wrong</p>
          <p className="text-xs text-muted-foreground mt-1 mb-4 max-w-sm">
            {this.state.error?.message || "An unexpected error occurred"}
          </p>
          <Button variant="outline" size="sm" onClick={this.handleRetry}>
            Try again
          </Button>
        </div>
      )
    }

    return this.props.children
  }
}
