import { useState } from 'react';
import type { ValidationResult, ValidationError } from '../types';

interface ValidationResultsProps {
  result: ValidationResult;
}

export function ValidationResults({ result }: ValidationResultsProps) {
  const [expandedErrors, setExpandedErrors] = useState<Set<number>>(new Set());
  const getValidationTypeLabel = () => {
    switch (result.validationType) {
      case 'json':
        return 'JSON Schema Validation';
      case 'proto':
        return 'Proto Validation';
      case 'both':
        return 'Combined Validation';
      default:
        return 'Validation';
    }
  };

  return (
    <div className="mt-6 space-y-4">
      <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-white">Validation Results</h3>
          {result.validationType && (
            <span className="text-xs text-gray-400 px-2 py-1 bg-gray-700/50 rounded">
              {getValidationTypeLabel()}
            </span>
          )}
        </div>
        
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

        {/* Combined validation results */}
        {result.validationType === 'both' && (
          <div className="mb-4 space-y-3">
            <div>
              <div className="flex items-center mb-2">
                <span className="text-sm font-medium text-gray-300 mr-2">JSON Schema:</span>
                {result.jsonValid !== undefined && (
                  result.jsonValid ? (
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-900/50 text-green-400 border border-green-700/50">
                      ✓ Valid
                    </span>
                  ) : (
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-900/50 text-red-400 border border-red-700/50">
                      ✗ Invalid
                    </span>
                  )
                )}
              </div>
            </div>
            <div>
              <div className="flex items-center mb-2">
                <span className="text-sm font-medium text-gray-300 mr-2">Proto:</span>
                {result.protoValid !== undefined && (
                  result.protoValid ? (
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-900/50 text-green-400 border border-green-700/50">
                      ✓ Valid
                    </span>
                  ) : (
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-900/50 text-red-400 border border-red-700/50">
                      ✗ Invalid
                    </span>
                  )
                )}
              </div>
            </div>
          </div>
        )}

        {/* JSON Schema errors */}
        {result.errors && result.errors.length > 0 && (
          <div className="mb-4">
            <h4 className="text-sm font-medium text-white mb-2">
              {result.validationType === 'both' ? 'JSON Schema Errors:' : 'Validation Errors:'}
            </h4>
            <ul className="list-disc list-inside space-y-1">
              {result.errors.map((error, index) => (
                <li key={index} className="text-sm text-red-400">
                  <span className="font-medium">{error.property}:</span> {error.message}
                </li>
              ))}
            </ul>
          </div>
        )}

        {/* Proto validation errors */}
        {result.apiErrors && result.apiErrors.length > 0 && (
          <div className="mb-4">
            <h4 className="text-sm font-medium text-white mb-2">
              {result.validationType === 'both' ? 'Proto Validation Errors:' : 'Validation Errors:'}
            </h4>
            <ul className="list-disc list-inside space-y-2">
              {result.apiErrors.map((error, index) => {
                // Handle both structured errors and legacy string errors
                const isStructured = typeof error === 'object' && 'friendly' in error;
                const friendly = isStructured ? (error as ValidationError).friendly : (error as string);
                const technical = isStructured ? (error as ValidationError).technical : undefined;
                const isExpanded = expandedErrors.has(index);
                const showTechnical = isStructured && technical && technical !== friendly;

                return (
                  <li key={index} className="text-sm text-red-400">
                    <div className="flex flex-col">
                      <div className="flex items-start">
                        <span className="flex-1">{friendly}</span>
                        {showTechnical && (
                          <button
                            type="button"
                            onClick={() => {
                              const newExpanded = new Set(expandedErrors);
                              if (isExpanded) {
                                newExpanded.delete(index);
                              } else {
                                newExpanded.add(index);
                              }
                              setExpandedErrors(newExpanded);
                            }}
                            className="ml-2 text-xs text-red-300 hover:text-red-200 underline transition-colors"
                          >
                            {isExpanded ? 'Show less ▲' : 'Show more ▼'}
                          </button>
                        )}
                      </div>
                      {showTechnical && isExpanded && (
                        <div className="mt-2 pl-4 border-l-2 border-red-700/50">
                          <div className="text-xs text-red-300 font-medium mb-1">Technical details:</div>
                          <pre className="text-xs text-red-400/80 whitespace-pre-wrap break-words">
                            {technical}
                          </pre>
                        </div>
                      )}
                    </div>
                  </li>
                );
              })}
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
