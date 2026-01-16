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
  errors?: Array<{
    property: string;
    message: string;
  }>;
}
