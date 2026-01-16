import type { ApiError } from '../types';

interface ErrorDisplayProps {
  error: ApiError;
  onRetry?: () => void;
}

export function ErrorDisplay({ error, onRetry }: ErrorDisplayProps) {
  const getErrorMessage = () => {
    if (error.statusCode === 404) {
      return 'Schema not found';
    }
    if (error.statusCode === 400) {
      return 'Invalid request';
    }
    if (error.statusCode === 500) {
      return 'Server error';
    }
    return error.message || 'An error occurred';
  };

  return (
    <div className="bg-red-900/20 border border-red-700/50 rounded-lg p-4">
      <div className="flex items-start">
        <div className="flex-shrink-0">
          <svg
            className="h-5 w-5 text-red-400"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
              clipRule="evenodd"
            />
          </svg>
        </div>
        <div className="ml-3 flex-1">
          <h3 className="text-sm font-medium text-red-400">Error</h3>
          <div className="mt-2 text-sm text-red-300">
            <p>{getErrorMessage()}</p>
            {error.statusCode && (
              <p className="mt-1 text-xs text-red-400">Status Code: {error.statusCode}</p>
            )}
          </div>
          {onRetry && (
            <div className="mt-4">
              <button
                type="button"
                onClick={onRetry}
                className="text-sm font-medium text-red-400 hover:text-red-300 underline transition-colors"
              >
                Retry
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
