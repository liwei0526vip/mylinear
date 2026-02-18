/**
 * IssueLabelSelector 组件 - 标签多选
 */

import { useEffect, useState } from 'react';
import { useLabelStore } from '@/stores/labelStore';

interface IssueLabelSelectorProps {
  teamId: string;
  value: string[];
  onChange: (labelIds: string[]) => void;
  disabled?: boolean;
}

export function IssueLabelSelector({
  teamId,
  value,
  onChange,
  disabled = false,
}: IssueLabelSelectorProps) {
  const { labels, fetchLabels } = useLabelStore();
  const [isOpen, setIsOpen] = useState(false);

  // 获取标签列表
  useEffect(() => {
    if (teamId) {
      fetchLabels(teamId);
    }
  }, [teamId, fetchLabels]);

  const toggleLabel = (labelId: string) => {
    if (value.includes(labelId)) {
      onChange(value.filter((id) => id !== labelId));
    } else {
      onChange([...value, labelId]);
    }
  };

  const selectedLabels = labels.filter((l) => value.includes(l.id));

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className="flex h-8 min-w-[120px] flex-wrap items-center gap-1 rounded-md border border-gray-300 bg-white px-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50"
      >
        {selectedLabels.length === 0 ? (
          <span className="text-gray-400">选择标签</span>
        ) : (
          selectedLabels.map((label) => (
            <span
              key={label.id}
              className="inline-flex items-center rounded-full px-2 py-0.5 text-xs"
              style={{ backgroundColor: label.color + '20', color: label.color }}
            >
              {label.name}
            </span>
          ))
        )}
      </button>

      {isOpen && (
        <div className="absolute left-0 top-full z-10 mt-1 w-48 rounded-md border border-gray-200 bg-white py-1 shadow-lg">
          {labels.length === 0 ? (
            <div className="px-3 py-2 text-sm text-gray-500">暂无标签</div>
          ) : (
            labels.map((label) => (
              <button
                key={label.id}
                type="button"
                onClick={() => toggleLabel(label.id)}
                className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-gray-50"
              >
                <span
                  className="h-3 w-3 rounded-full"
                  style={{ backgroundColor: label.color }}
                />
                <span>{label.name}</span>
                {value.includes(label.id) && (
                  <svg
                    className="ml-auto h-4 w-4 text-indigo-500"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M5 13l4 4L19 7"
                    />
                  </svg>
                )}
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}
