/**
 * 团队成员列表组件
 */

import { useState, useEffect } from 'react';
import { useTeamStore } from '../../stores/teamStore';
import { Button } from '../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { AddMemberDialog } from './AddMemberDialog';

interface TeamMemberListProps {
  teamId: string;
}

export function TeamMemberList({ teamId }: TeamMemberListProps) {
  const { members, isLoading, error, fetchMembers, removeMember, updateMemberRole } =
    useTeamStore();
  const [showAddDialog, setShowAddDialog] = useState(false);

  useEffect(() => {
    fetchMembers(teamId);
  }, [teamId]);

  const handleRemove = async (userId: string) => {
    if (!confirm('确定要移除此成员吗？')) {
      return;
    }

    try {
      await removeMember(teamId, userId);
    } catch {
      // 错误由 store 处理
    }
  };

  const handleRoleChange = async (userId: string, newRole: string) => {
    try {
      await updateMemberRole(teamId, userId, newRole as 'admin' | 'member');
    } catch {
      // 错误由 store 处理
    }
  };

  const handleAddSuccess = () => {
    fetchMembers(teamId);
    setShowAddDialog(false);
  };

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>团队成员</CardTitle>
            <CardDescription>管理团队的成员和权限</CardDescription>
          </div>
          <Button onClick={() => setShowAddDialog(true)}>添加成员</Button>
        </CardHeader>
        <CardContent>
          {error && (
            <div className="mb-4 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {error}
            </div>
          )}

          {isLoading ? (
            <div className="py-8 text-center text-muted-foreground">加载中...</div>
          ) : members.length === 0 ? (
            <div className="py-8 text-center text-muted-foreground">
              暂无成员
            </div>
          ) : (
            <div className="space-y-2">
              {members.map((member) => (
                <div
                  key={member.user_id}
                  className="flex items-center justify-between rounded-lg border p-4"
                >
                  <div>
                    <div className="font-medium">
                      {member.user?.name || member.id}
                    </div>
                    <div className="text-sm text-muted-foreground">
                      {member.role === 'admin' ? 'Owner' : 'Member'}
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <select
                      value={member.role}
                      onChange={(e) => handleRoleChange(member.user_id, e.target.value)}
                      className="rounded border px-2 py-1 text-sm"
                    >
                      <option value="admin">Owner</option>
                      <option value="member">Member</option>
                    </select>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleRemove(member.user_id)}
                    >
                      移除
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <AddMemberDialog
        teamId={teamId}
        open={showAddDialog}
        onOpenChange={setShowAddDialog}
        onSuccess={handleAddSuccess}
      />
    </>
  );
}
