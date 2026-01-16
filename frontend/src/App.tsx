import { useState, useEffect } from 'react';
import { ProtoFileList } from './components/ProtoFileList';
import { SchemaForm } from './components/SchemaForm';
import { fetchProtoFiles } from './services/api';
import type { ProtoFile, ApiError } from './types';

function App() {
  const [selectedProto, setSelectedProto] = useState<ProtoFile | null>(null);
  const [protoFiles, setProtoFiles] = useState<ProtoFile[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadProtoFiles = async () => {
      try {
        setLoading(true);
        setError(null);
        const files = await fetchProtoFiles();
        setProtoFiles(files);
      } catch (err) {
        const apiError = err as ApiError;
        setError(apiError.message || 'Failed to load proto files');
        console.error('Error fetching proto files:', err);
      } finally {
        setLoading(false);
      }
    };

    loadProtoFiles();
  }, []);

  const handleSelectProto = (protoFile: ProtoFile) => {
    setSelectedProto(protoFile);
  };

  return (
    <div className="min-h-screen bg-gray-900">
      <div className="bg-gray-800 border-b border-gray-700">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <h1 className="text-3xl font-bold text-white">Validation Service</h1>
          <p className="text-sm text-gray-400 mt-2">Generate and validate forms from proto schemas</p>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <div className="lg:col-span-1">
            {loading ? (
              <div className="bg-gray-800 border border-gray-700 rounded-lg p-8 text-center">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto"></div>
                <p className="text-gray-400 text-sm mt-4">Loading proto files...</p>
              </div>
            ) : error ? (
              <div className="bg-gray-800 border border-red-700 rounded-lg p-8 text-center">
                <svg
                  className="mx-auto h-8 w-8 text-red-500"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
                <h3 className="mt-4 text-lg font-medium text-white">Error Loading Proto Files</h3>
                <p className="mt-2 text-sm text-red-400">{error}</p>
                <button
                  onClick={() => window.location.reload()}
                  className="mt-4 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
                >
                  Retry
                </button>
              </div>
            ) : (
              <ProtoFileList
                protoFiles={protoFiles}
                selectedFullyQualifiedName={selectedProto?.fullyQualifiedName}
                onSelect={handleSelectProto}
              />
            )}
          </div>

          <div className="lg:col-span-2">
            {selectedProto ? (
              <SchemaForm fullyQualifiedName={selectedProto.fullyQualifiedName} />
            ) : (
              <div className="bg-gray-800 border border-gray-700 rounded-lg p-12 text-center">
                <svg
                  className="mx-auto h-12 w-12 text-gray-500"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                  />
                </svg>
                <h3 className="mt-4 text-lg font-medium text-white">Select a Proto File</h3>
                <p className="mt-2 text-sm text-gray-400">
                  Choose a proto file from the list to view and interact with its schema form.
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}

export default App;
