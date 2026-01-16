import type { JSONSchema7 } from "json-schema";
import type { ApiError, ProtoFile } from "../types";

const API_BASE_URL = "http://localhost:8080/api";

export async function fetchProtoFiles(): Promise<ProtoFile[]> {
  const url = `${API_BASE_URL}/v1/proto-files`;

  try {
    const response = await fetch(url);

    if (!response.ok) {
      const errorText = await response.text().catch(() => "Unknown error");
      const error: ApiError = {
        message: errorText || `HTTP ${response.status}: ${response.statusText}`,
        statusCode: response.status,
      };
      throw error;
    }

    const data = await response.json();

    // Validate that the response is an array
    if (!Array.isArray(data)) {
      throw {
        message: "Invalid response: expected an array of proto files",
      } as ApiError;
    }

    // Validate each proto file has required fields
    for (const protoFile of data) {
      if (typeof protoFile !== "object" || protoFile === null) {
        throw {
          message: "Invalid response: proto file is not an object",
        } as ApiError;
      }
      if (typeof protoFile.fullyQualifiedName !== "string") {
        throw {
          message: "Invalid response: proto file missing fullyQualifiedName",
        } as ApiError;
      }
    }

    return data as ProtoFile[];
  } catch (error) {
    if (error && typeof error === "object" && "statusCode" in error) {
      throw error;
    }

    // Network or other errors
    throw {
      message:
        error instanceof Error ? error.message : "Failed to fetch proto files",
    } as ApiError;
  }
}

export async function fetchSchema(
  fullyQualifiedName: string
): Promise<JSONSchema7> {
  const url = `${API_BASE_URL}/v1/schema/${encodeURIComponent(
    fullyQualifiedName
  )}`;

  try {
    const response = await fetch(url);

    if (!response.ok) {
      const errorText = await response.text().catch(() => "Unknown error");
      const error: ApiError = {
        message: errorText || `HTTP ${response.status}: ${response.statusText}`,
        statusCode: response.status,
      };
      throw error;
    }

    const data = await response.json();

    // Validate that the response is a valid JSON schema
    if (typeof data !== "object" || data === null) {
      throw {
        message: "Invalid JSON schema: response is not an object",
      } as ApiError;
    }

    return data as JSONSchema7;
  } catch (error) {
    if (error && typeof error === "object" && "statusCode" in error) {
      throw error;
    }

    // Network or other errors
    throw {
      message:
        error instanceof Error ? error.message : "Failed to fetch schema",
    } as ApiError;
  }
}

export interface ValidationError {
  friendly: string; // Human-readable message
  technical: string; // Original technical error
}

export interface ValidateProtoResponse {
  success: boolean;
  errors: ValidationError[];
}

export async function validateProto(
  schemaName: string,
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  payload: any
): Promise<ValidateProtoResponse> {
  const url = `${API_BASE_URL}/v1/validate-proto`;

  try {
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        schemaName,
        payload,
      }),
    });

    if (!response.ok) {
      const errorText = await response.text().catch(() => "Unknown error");
      const error: ApiError = {
        message: errorText || `HTTP ${response.status}: ${response.statusText}`,
        statusCode: response.status,
      };
      throw error;
    }

    const data = await response.json();

    // Validate response structure
    if (typeof data !== "object" || data === null) {
      throw {
        message: "Invalid response: response is not an object",
      } as ApiError;
    }

    // Ensure response has expected structure
    if (typeof data.success !== "boolean") {
      throw {
        message: "Invalid response: missing or invalid success field",
      } as ApiError;
    }

    if (!Array.isArray(data.errors)) {
      throw {
        message: "Invalid response: errors field is not an array",
      } as ApiError;
    }

    return {
      success: data.success,
      errors: data.errors || [],
    };
  } catch (error) {
    if (error && typeof error === "object" && "statusCode" in error) {
      throw error;
    }

    // Network or other errors
    throw {
      message:
        error instanceof Error ? error.message : "Failed to validate proto",
    } as ApiError;
  }
}
