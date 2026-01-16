import type { JSONSchema7 } from 'json-schema';

export interface ProtoFile {
  name: string;
  description: string;
  fullyQualifiedName: string;
}

export type SchemaResponse = JSONSchema7;

export interface ApiError {
  message: string;
  statusCode?: number;
}

export interface ValidationResult {
  valid: boolean;
  data: any;
  // JSON schema validation errors (property: message format)
  errors?: Array<{
    property: string;
    message: string;
  }>;
  // Proto validation errors (simple string array from API)
  apiErrors?: string[];
  // Track which validation type was run
  validationType?: 'json' | 'proto' | 'both';
  // For combined validation, track individual results
  jsonValid?: boolean;
  protoValid?: boolean;
}
