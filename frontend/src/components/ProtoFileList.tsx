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
  return (
    <div className="space-y-3">
      <h2 className="text-xl font-semibold text-white mb-4">Proto Files</h2>
      {protoFiles.length === 0 ? (
        <p className="text-gray-400 text-sm">No proto files available</p>
      ) : (
        <div className="space-y-3">
          {protoFiles.map((protoFile) => {
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
