/**
 * IssueAssigneeSelector 组件 - 负责人选择器
 */

import { useEffect, useState } from 'react';
import { useTeamStore } from '@/stores/teamStore';

interface IssueAssigneeSelectorProps {
  teamId: string;
  value?: string;
  onChange: (assigneeId: string | undefined) => void;
  disabled?: boolean;
}

export function IssueAssigneeSelector({
  teamId,
  value,
  onChange,
  disabled = false,
}: IssueAssigneeSelectorProps) {
  const { members, fetchMembers } = useTeamStore();
  const [isOpen, setIsOpen] = useState(false);

  // 获取团队成员
  useEffect(() => {
    if (teamId) {
      fetchMembers(teamId);
    }
  }, [teamId, fetchMembers]);

  const selectedMember = members.find((m) => m.user_id === value);

  return (
    <div className="relative">
      <button
        type="button"
        onClick={() => !disabled && setIsOpen(!isOpen)}
        disabled={disabled}
        className="flex h-8 items-center gap-2 rounded-md border border-gray-300 bg-white px-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50"
      >
        {selectedMember ? (
          <>
            <div className="h-5 w-5 rounded-full bg-gray-300" />
            <span>{selectedMember.user?.name || '未知'}</span>
          </>
        ) : (
          <span className="text-gray-400">未分配</span>
        )}
      </button>

      {isOpen && (
        <div className="absolute left-0 top-full z-10 mt-1 w-48 rounded-md border border-gray-200 bg-white py-1 shadow-lg">
          <button
            type="button"
            onClick={() => {
              onChange(undefined);
              setIsOpen(false);
            }}
            className="w-full px-3 py-2 text-left text-sm text-gray-500 hover:bg-gray-50"
          >
            未分配
          </button>
          {members.map((member) => (
            <button
              key={member.user_id}
              type="button"
              onClick={() => {
                onChange(member.user_id);
                setIsOpen(false);
              }}
              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-gray-50"
            >
              <div className="h-5 w-5 rounded-full bg-gray-300" />
              <span>{member.user?.name || '未知'}</span>
            </button>
          ))}
        </div>
      )}
    </div>
  );
}
