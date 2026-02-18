/**
 * IssueStatusSelector 组件 - 状态选择下拉框
 */

import { useEffect } from 'react';
import { useWorkflowStore } from '@/stores/workflowStore';
import type { WorkflowState } from '@/types/workflow';

interface IssueStatusSelectorProps {
  teamId: string;
  value: string;
  onChange: (statusId: string) => void;
  disabled?: boolean;
}

export function IssueStatusSelector({
  teamId,
  value,
  onChange,
  disabled = false,
}: IssueStatusSelectorProps) {
  const { states, fetchStates } = useWorkflowStore();

  // 获取状态列表
  useEffect(() => {
    if (teamId && states.length === 0) {
      fetchStates(teamId);
    }
  }, [teamId, states.length, fetchStates]);

  return (
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      disabled={disabled}
      className="h-8 rounded-md border border-gray-300 bg-white px-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50"
    >
      {states.map((state) => (
        <option key={state.id} value={state.id}>
          {state.name}
        </option>
      ))}
    </select>
  );
}

/**
 * 状态图标组件
 */
interface StatusIconProps {
  type: string;
  color: string;
  size?: number;
}

export function StatusIcon({ type, color, size = 16 }: StatusIconProps) {
  // 根据状态类型返回不同的图标
  switch (type) {
    case 'backlog':
      return (
        <svg width={size} height={size} viewBox="0 0 16 16" fill="none">
          <circle cx="8" cy="8" r="6" stroke={color} strokeWidth="1.5" strokeDasharray="2 2" />
        </svg>
      );
    case 'unstarted':
      return (
        <svg width={size} height={size} viewBox="0 0 16 16" fill="none">
          <circle cx="8" cy="8" r="6" stroke={color} strokeWidth="1.5" />
        </svg>
      );
    case 'started':
      return (
        <svg width={size} height={size} viewBox="0 0 16 16" fill="none">
          <circle cx="8" cy="8" r="6" stroke={color} strokeWidth="1.5" />
          <path d="M8 4v4l2 2" stroke={color} strokeWidth="1.5" strokeLinecap="round" />
        </svg>
      );
    case 'completed':
      return (
        <svg width={size} height={size} viewBox="0 0 16 16" fill="none">
          <circle cx="8" cy="8" r="6" stroke={color} strokeWidth="1.5" />
          <path d="M5 8l2 2 4-4" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
        </svg>
      );
    case 'canceled':
      return (
        <svg width={size} height={size} viewBox="0 0 16 16" fill="none">
          <circle cx="8" cy="8" r="6" stroke={color} strokeWidth="1.5" />
          <path d="M5 5l6 6M11 5l-6 6" stroke={color} strokeWidth="1.5" strokeLinecap="round" />
        </svg>
      );
    default:
      return (
        <svg width={size} height={size} viewBox="0 0 16 16" fill="none">
          <circle cx="8" cy="8" r="6" stroke={color} strokeWidth="1.5" />
        </svg>
      );
  }
}
