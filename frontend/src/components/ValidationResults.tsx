import type { ValidationResult } from '../types';

interface ValidationResultsProps {
  result: ValidationResult;
}

export function ValidationResults({ result }: ValidationResultsProps) {
  return (
    <div className="mt-6 space-y-4">
      <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
        <h3 className="text-lg font-semibold text-white mb-4">Validation Results</h3>
        
        <div className="mb-4">
          <div className="flex items-center">
            <span className="text-sm font-medium text-gray-300 mr-2">Status:</span>
            {result.valid ? (
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-900/50 text-green-400 border border-green-700/50">
                Valid
              </span>
            ) : (
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-900/50 text-red-400 border border-red-700/50">
                Invalid
              </span>
            )}
          </div>
        </div>

        {result.errors && result.errors.length > 0 && (
          <div className="mb-4">
            <h4 className="text-sm font-medium text-white mb-2">Validation Errors:</h4>
            <ul className="list-disc list-inside space-y-1">
              {result.errors.map((error, index) => (
                <li key={index} className="text-sm text-red-400">
                  <span className="font-medium">{error.property}:</span> {error.message}
                </li>
              ))}
            </ul>
          </div>
        )}

        <div>
          <h4 className="text-sm font-medium text-white mb-2">Submitted Data:</h4>
          <pre className="bg-gray-900 border border-gray-700 rounded p-3 text-xs overflow-x-auto text-gray-300">
            {JSON.stringify(result.data, null, 2)}
          </pre>
        </div>
      </div>
    </div>
  );
}
