/* eslint-disable @typescript-eslint/no-explicit-any */
import { useState, useEffect } from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import type { JSONSchema7 } from 'json-schema';
import type { IChangeEvent } from '@rjsf/core';
import { fetchSchema } from '../services/api';
import type { ApiError, ValidationResult } from '../types';
import { ErrorDisplay } from './ErrorDisplay';
import { ValidationResults } from './ValidationResults';

interface SchemaFormProps {
  fullyQualifiedName: string;
}

export function SchemaForm({ fullyQualifiedName }: SchemaFormProps) {
  const [schema, setSchema] = useState<JSONSchema7 | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<ApiError | null>(null);
  const [formData, setFormData] = useState<any>({});
  const [validationResult, setValidationResult] = useState<ValidationResult | null>(null);

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

  return (
    <div>
      <h2 className="text-2xl font-semibold text-white mb-2">
        {schema.title || fullyQualifiedName}
      </h2>
      {schema.description && (
        <p className="text-sm text-gray-400 mb-6">{schema.description}</p>
      )}

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
          <div className="mt-6">
            <button
              type="submit"
              className="px-6 py-2.5 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:ring-offset-gray-800 transition-colors font-medium"
            >
              Validate
            </button>
          </div>
        </Form>
      </div>

      {validationResult && <ValidationResults result={validationResult} />}
    </div>
  );
}
