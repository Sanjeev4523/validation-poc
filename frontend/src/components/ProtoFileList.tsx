import { useState, useMemo } from 'react';
import type { ProtoFile } from '../types';

interface ProtoFileListProps {
  protoFiles: ProtoFile[];
  selectedFullyQualifiedName?: string;
  onSelect: (protoFile: ProtoFile) => void;
}

export function ProtoFileList({
  protoFiles,
  selectedFullyQualifiedName,
  onSelect,
}: ProtoFileListProps) {
  const [searchQuery, setSearchQuery] = useState('');

  // Filter and sort proto files
  const filteredAndSortedFiles = useMemo(() => {
    let filtered = protoFiles;

    // Filter by search query (name and description)
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase().trim();
      filtered = protoFiles.filter(
        (file) =>
          file.name.toLowerCase().includes(query) ||
          file.description.toLowerCase().includes(query)
      );
    }

    // Sort alphabetically by name
    return [...filtered].sort((a, b) => a.name.localeCompare(b.name));
  }, [protoFiles, searchQuery]);

  return (
    <div className="space-y-3">
      <h2 className="text-xl font-semibold text-white mb-4">Proto Files</h2>
      
      {/* Search Input */}
      <div className="mb-4">
        <div className="relative">
          <input
            type="text"
            placeholder="Search by name or description..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full px-4 py-2 pl-10 bg-gray-800 border border-gray-700 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          <svg
            className="absolute left-3 top-2.5 h-5 w-5 text-gray-500"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
            />
          </svg>
        </div>
      </div>

      {protoFiles.length === 0 ? (
        <p className="text-gray-400 text-sm">No proto files available</p>
      ) : filteredAndSortedFiles.length === 0 ? (
        <p className="text-gray-400 text-sm">No proto files match your search</p>
      ) : (
        <div className="space-y-3">
          {filteredAndSortedFiles.map((protoFile) => {
            const isSelected = protoFile.fullyQualifiedName === selectedFullyQualifiedName;
            return (
              <button
                key={protoFile.fullyQualifiedName}
                type="button"
                onClick={() => onSelect(protoFile)}
                className={`w-full text-left p-4 rounded-lg border-2 transition-all ${
                  isSelected
                    ? 'border-blue-500 bg-blue-900/30 shadow-lg shadow-blue-500/20'
                    : 'border-gray-700 bg-gray-800 hover:border-gray-600 hover:bg-gray-750'
                }`}
              >
                <h3 className="font-medium text-white">{protoFile.name}</h3>
                <p className="text-sm text-gray-400 mt-1">{protoFile.description}</p>
                <p className="text-xs text-gray-500 mt-2 font-mono">
                  {protoFile.fullyQualifiedName}
                </p>
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}
