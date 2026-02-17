/**
 * 添加成员对话框
 */

import { useState } from 'react';
import { useTeamStore } from '../../stores/teamStore';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';

interface AddMemberDialogProps {
  teamId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function AddMemberDialog({
  teamId,
  open,
  onOpenChange,
  onSuccess,
}: AddMemberDialogProps) {
  const { addMember, isLoading, error } = useTeamStore();

  const [userId, setUserId] = useState('');
  const [role, setRole] = useState<'admin' | 'member'>('member');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      await addMember(teamId, {
        user_id: userId,
        role,
      });
      setUserId('');
      setRole('member');
      onSuccess();
    } catch {
      // 错误由 store 处理
    }
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>添加成员</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {error}
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="userId">用户 ID</Label>
              <Input
                id="userId"
                value={userId}
                onChange={(e) => setUserId(e.target.value)}
                placeholder="输入用户 ID"
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="role">角色</Label>
              <select
                id="role"
                value={role}
                onChange={(e) => setRole(e.target.value as 'admin' | 'member')}
                className="w-full rounded border px-3 py-2"
              >
                <option value="member">Member</option>
                <option value="admin">Owner</option>
              </select>
            </div>

            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
              >
                取消
              </Button>
              <Button type="submit" disabled={isLoading || !userId}>
                {isLoading ? '添加中...' : '添加'}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
