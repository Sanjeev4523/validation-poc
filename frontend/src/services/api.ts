import type { JSONSchema7 } from 'json-schema';
import type { ApiError } from '../types';

const API_BASE_URL = 'http://localhost:8080/api';

export async function fetchSchema(fullyQualifiedName: string): Promise<JSONSchema7> {
  const url = `${API_BASE_URL}/v1/schema/${encodeURIComponent(fullyQualifiedName)}`;
  
  try {
    const response = await fetch(url);
    
    if (!response.ok) {
      const errorText = await response.text().catch(() => 'Unknown error');
      const error: ApiError = {
        message: errorText || `HTTP ${response.status}: ${response.statusText}`,
        statusCode: response.status,
      };
      throw error;
    }
    
    const data = await response.json();
    
    // Validate that the response is a valid JSON schema
    if (typeof data !== 'object' || data === null) {
      throw {
        message: 'Invalid JSON schema: response is not an object',
      } as ApiError;
    }
    
    return data as JSONSchema7;
  } catch (error) {
    if (error && typeof error === 'object' && 'statusCode' in error) {
      throw error;
    }
    
    // Network or other errors
    throw {
      message: error instanceof Error ? error.message : 'Failed to fetch schema',
    } as ApiError;
  }
}
