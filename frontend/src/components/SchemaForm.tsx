/* eslint-disable @typescript-eslint/no-explicit-any */
import { useState, useEffect } from 'react';
import Form from '@rjsf/core';
import validator from '@rjsf/validator-ajv8';
import type { JSONSchema7 } from 'json-schema';
import type { IChangeEvent } from '@rjsf/core';
import { fetchSchema, validateProto } from '../services/api';
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
  const [validating, setValidating] = useState(false);

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

  // Proto validation via API
  const handleValidateProto = async () => {
    if (!schema) return;

    setValidating(true);
    try {
      const response = await validateProto(fullyQualifiedName, formData);
      
      const result: ValidationResult = {
        valid: response.success,
        data: formData,
        apiErrors: response.errors.length > 0 ? response.errors : undefined,
        validationType: 'proto',
      };
      setValidationResult(result);
    } catch (err) {
      const apiError = err as ApiError;
      const result: ValidationResult = {
        valid: false,
        data: formData,
        apiErrors: [apiError.message || 'Failed to validate proto'],
        validationType: 'proto',
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
        const protoResponse = await validateProto(fullyQualifiedName, formData);
        const protoValid = protoResponse.success;

        const result: ValidationResult = {
          valid: jsonValid && protoValid,
          data: formData,
          apiErrors: protoResponse.errors.length > 0 ? protoResponse.errors : undefined,
          validationType: 'both',
          jsonValid: jsonValid,
          protoValid: protoValid,
        };
        setValidationResult(result);
      } catch (protoErr) {
        const apiError = protoErr as ApiError;
        const result: ValidationResult = {
          valid: false,
          data: formData,
          apiErrors: [apiError.message || 'Failed to validate proto'],
          validationType: 'both',
          jsonValid: jsonValid,
          protoValid: false,
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
