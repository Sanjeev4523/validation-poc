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

// Structured validation error with friendly and technical messages
export interface ValidationError {
  friendly: string;  // Human-readable message
  technical: string;  // Original technical error
}

export interface ValidationResult {
  valid: boolean;
  data: any;
  // JSON schema validation errors (property: message format)
  errors?: Array<{
    property: string;
    message: string;
  }>;
  // Proto validation errors (structured format with friendly and technical)
  // Also supports legacy string format for backward compatibility
  apiErrors?: (ValidationError | string)[];
  // Track which validation type was run
  validationType?: 'json' | 'proto' | 'both';
  // For combined validation, track individual results
  jsonValid?: boolean;
  protoValid?: boolean;
  // Track commit used for validation
  commit?: string;
}

export interface Commit {
  id: string;
  createTime: string;
  ownerId: string;
  moduleId: string;
  digest: {
    type: string;
    value: string;
  };
  createdByUserId: string;
}

export interface CommitCheckState {
  status: string;
  updateTime: string;
}

export interface LabelHistoryValue {
  commit: Commit;
  commitCheckState: CommitCheckState;
}

export interface CommitsResponse {
  nextPageToken?: string;
  values: LabelHistoryValue[];
}
