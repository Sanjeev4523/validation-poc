/* eslint-disable @typescript-eslint/no-explicit-any */
import { useState, useEffect } from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import type { JSONSchema7 } from 'json-schema';
import type { IChangeEvent } from '@rjsf/core';
import { fetchSchema, validateProto, fetchCommits } from '../services/api';
import type { ApiError, ValidationResult, CommitsResponse } from '../types';
import { ErrorDisplay } from './ErrorDisplay';
import { ValidationResults } from './ValidationResults';

interface SchemaFormProps {
  fullyQualifiedName: string;
  commits: CommitsResponse | null;
}

export function SchemaForm({ fullyQualifiedName, commits: initialCommits }: SchemaFormProps) {
  const [schema, setSchema] = useState<JSONSchema7 | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<ApiError | null>(null);
  const [formData, setFormData] = useState<any>({});
  const [validationResult, setValidationResult] = useState<ValidationResult | null>(null);
  const [validating, setValidating] = useState(false);
  const [commits, setCommits] = useState<CommitsResponse | null>(initialCommits);
  const [pageSize, setPageSize] = useState<number>(10);
  const [selectedCommit, setSelectedCommit] = useState<string>('main');
  const [loadingCommits, setLoadingCommits] = useState(false);

  useEffect(() => {
    let cancelled = false;

    const loadSchema = async () => {
      setLoading(true);
      setError(null);
      setSchema(null);
      setFormData({});
      setValidationResult(null);

      try {
        const fetchedSchema = await fetchSchema(fullyQualifiedName);
        if (!cancelled) {
          setSchema(fetchedSchema);
          setLoading(false);
        }
      } catch (err) {
        if (!cancelled) {
          setError(err as ApiError);
          setLoading(false);
        }
      }
    };

    loadSchema();

    return () => {
      cancelled = true;
    };
  }, [fullyQualifiedName]);

  const handleSubmit = (data: IChangeEvent<any, JSONSchema7, any>, event: any) => {
    // RJSF only calls onSubmit when form is valid
    const result: ValidationResult = {
      valid: true,
      data: data.formData,
    };

    setValidationResult(result);
    event.preventDefault(); // Prevent default form submission
  };

  const handleError = (errors: any) => {
    // Convert RJSF errors to our format
    const validationErrors: Array<{ property: string; message: string }> = [];
    
    if (errors && Array.isArray(errors)) {
      errors.forEach((error: any) => {
        const property = error.property || error.instancePath || 'root';
        const message = error.message || 'Validation error';
        validationErrors.push({
          property: property.replace(/^#\//, ''), // Remove leading #/
          message: message,
        });
      });
    }

    const result: ValidationResult = {
      valid: false,
      data: formData,
      errors: validationErrors.length > 0 ? validationErrors : undefined,
    };

    setValidationResult(result);
  };

  // Manual JSON schema validation
  const validateJsonSchema = () => {
    if (!schema) return;

    setValidating(true);
    try {
      // Use the validator to validate formData against schema
      const validation = validator.validateFormData(formData, schema);
      
      if (!validation.errors || validation.errors.length === 0) {
        const result: ValidationResult = {
          valid: true,
          data: formData,
          validationType: 'json',
        };
        setValidationResult(result);
      } else {
        // Convert RJSF errors to our format
        const validationErrors: Array<{ property: string; message: string }> = [];
        validation.errors.forEach((error: any) => {
          const property = error.property || error.instancePath || 'root';
          const message = error.message || 'Validation error';
          validationErrors.push({
            property: property.replace(/^#\//, ''), // Remove leading #/
            message: message,
          });
        });

        const result: ValidationResult = {
          valid: false,
          data: formData,
          errors: validationErrors,
          validationType: 'json',
        };
        setValidationResult(result);
      }
    } catch (err) {
      const result: ValidationResult = {
        valid: false,
        data: formData,
        errors: [{ property: 'root', message: err instanceof Error ? err.message : 'Validation failed' }],
        validationType: 'json',
      };
      setValidationResult(result);
    } finally {
      setValidating(false);
    }
  };

  // Load commits with custom page size
  const loadCommits = async (size: number) => {
    setLoadingCommits(true);
    try {
      const commitsData = await fetchCommits(size, 'main');
      setCommits(commitsData);
    } catch (err) {
      console.error('Error fetching commits:', err);
    } finally {
      setLoadingCommits(false);
    }
  };

  // Load commits if not provided
  useEffect(() => {
    if (!commits) {
      loadCommits(pageSize);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Handle page size change
  const handlePageSizeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newSize = parseInt(e.target.value, 10);
    if (!isNaN(newSize) && newSize > 0) {
      setPageSize(newSize);
      loadCommits(newSize);
    }
  };

  // Proto validation via API
  const handleValidateProto = async () => {
    if (!schema) return;

    setValidating(true);
    try {
      const commitToUse = selectedCommit === 'main' ? undefined : selectedCommit;
      const response = await validateProto(fullyQualifiedName, formData, commitToUse);
      
      const result: ValidationResult = {
        valid: response.success,
        data: formData,
        apiErrors: response.errors.length > 0 ? response.errors : undefined,
        validationType: 'proto',
        commit: selectedCommit,
      };
      setValidationResult(result);
    } catch (err) {
      const apiError = err as ApiError;
      const errorMessage = apiError.message || 'Failed to validate proto';
      const result: ValidationResult = {
        valid: false,
        data: formData,
        // Use string format for API errors (catchall)
        apiErrors: [errorMessage],
        validationType: 'proto',
        commit: selectedCommit,
      };
      setValidationResult(result);
    } finally {
      setValidating(false);
    }
  };

  // Combined validation: JSON schema first, then proto
  const handleValidateBoth = async () => {
    if (!schema) return;

    setValidating(true);
    try {
      // First, validate JSON schema
      const jsonValidation = validator.validateFormData(formData, schema);
      const jsonValid = !jsonValidation.errors || jsonValidation.errors.length === 0;
      
      let jsonErrors: Array<{ property: string; message: string }> | undefined;
      if (!jsonValid) {
        jsonErrors = [];
        jsonValidation.errors.forEach((error: any) => {
          const property = error.property || error.instancePath || 'root';
          const message = error.message || 'Validation error';
          jsonErrors!.push({
            property: property.replace(/^#\//, ''),
            message: message,
          });
        });
      }

      // If JSON validation fails, return early with JSON errors only
      if (!jsonValid) {
        const result: ValidationResult = {
          valid: false,
          data: formData,
          errors: jsonErrors,
          validationType: 'both',
          jsonValid: false,
        };
        setValidationResult(result);
        setValidating(false);
        return;
      }

      // JSON validation passed, now validate proto
      try {
        const commitToUse = selectedCommit === 'main' ? undefined : selectedCommit;
        const protoResponse = await validateProto(fullyQualifiedName, formData, commitToUse);
        const protoValid = protoResponse.success;

        const result: ValidationResult = {
          valid: jsonValid && protoValid,
          data: formData,
          apiErrors: protoResponse.errors.length > 0 ? protoResponse.errors : undefined,
          validationType: 'both',
          jsonValid: jsonValid,
          protoValid: protoValid,
          commit: selectedCommit,
        };
        setValidationResult(result);
      } catch (protoErr) {
        const apiError = protoErr as ApiError;
        const errorMessage = apiError.message || 'Failed to validate proto';
        const result: ValidationResult = {
          valid: false,
          data: formData,
          // Use string format for API errors (catchall)
          apiErrors: [errorMessage],
          validationType: 'both',
          jsonValid: jsonValid,
          protoValid: false,
          commit: selectedCommit,
        };
        setValidationResult(result);
      }
    } catch (err) {
      const result: ValidationResult = {
        valid: false,
        data: formData,
        errors: [{ property: 'root', message: err instanceof Error ? err.message : 'Validation failed' }],
        validationType: 'both',
        jsonValid: false,
      };
      setValidationResult(result);
    } finally {
      setValidating(false);
    }
  };

  const handleRetry = () => {
    setError(null);
    setLoading(true);
    fetchSchema(fullyQualifiedName)
      .then((fetchedSchema) => {
        setSchema(fetchedSchema);
        setLoading(false);
      })
      .catch((err) => {
        setError(err as ApiError);
        setLoading(false);
      });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="text-center">
          <div className="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
          <p className="mt-4 text-sm text-gray-400">Loading schema...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return <ErrorDisplay error={error} onRetry={handleRetry} />;
  }

  if (!schema) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-400">No schema available</p>
      </div>
    );
  }

  // Prepare commits list with "main" on top
  const commitsList = commits ? [
    { id: 'main', createTime: '', label: 'main (default)' },
    ...commits.values.map((value) => ({
      id: value.commit.id,
      createTime: value.commit.createTime,
      label: `${value.commit.id.substring(0, 8)} - ${new Date(value.commit.createTime).toLocaleString()}`,
    }))
  ] : [{ id: 'main', createTime: '', label: 'main (default)' }];

  return (
    <div>
      <h2 className="text-2xl font-semibold text-white mb-2">
        {schema.title || fullyQualifiedName}
      </h2>
      {schema.description && (
        <p className="text-sm text-gray-400 mb-6">{schema.description}</p>
      )}

      {/* Commits and Page Size Controls */}
      <div className="bg-gray-800 border border-gray-700 rounded-lg p-4 mb-6">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div>
            <label htmlFor="commit-select" className="block text-sm font-medium text-gray-300 mb-2">
              Select Commit
            </label>
            <select
              id="commit-select"
              value={selectedCommit}
              onChange={(e) => setSelectedCommit(e.target.value)}
              className="w-full px-3 py-2 bg-gray-900 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              {commitsList.map((commit) => (
                <option key={commit.id} value={commit.id}>
                  {commit.label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <label htmlFor="page-size" className="block text-sm font-medium text-gray-300 mb-2">
              Page Size
            </label>
            <input
              id="page-size"
              type="number"
              min="1"
              value={pageSize}
              onChange={handlePageSizeChange}
              disabled={loadingCommits}
              className="w-full px-3 py-2 bg-gray-900 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
            />
            {loadingCommits && (
              <p className="text-xs text-gray-400 mt-1">Loading commits...</p>
            )}
          </div>
        </div>
      </div>

      <div className="bg-gray-800 border border-gray-700 rounded-lg p-6">
        <Form
          schema={schema}
          validator={validator}
          formData={formData}
          onChange={({ formData }) => setFormData(formData)}
          onSubmit={handleSubmit}
          onError={handleError}
          liveValidate={false}
        >
          <div></div>
        </Form>
        
        <div className="mt-6 flex gap-3">
          <button
            type="button"
            onClick={validateJsonSchema}
            disabled={validating}
            className="px-6 py-2.5 bg-green-600 text-white rounded-lg hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-green-500 focus:ring-offset-2 focus:ring-offset-gray-800 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {validating ? 'Validating...' : 'Validate JSON Schema'}
          </button>
          <button
            type="button"
            onClick={handleValidateProto}
            disabled={validating}
            className="px-6 py-2.5 bg-purple-600 text-white rounded-lg hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2 focus:ring-offset-gray-800 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {validating ? 'Validating...' : 'Validate Proto'}
          </button>
          <button
            type="button"
            onClick={handleValidateBoth}
            disabled={validating}
            className="px-6 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-800 transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {validating ? 'Validating...' : 'Validate'}
          </button>
        </div>
      </div>

      {validationResult && <ValidationResults result={validationResult} />}
    </div>
  );
}
