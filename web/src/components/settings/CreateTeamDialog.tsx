/**
 * 创建团队对话框
 */

import { useState } from 'react';
import { useTeamStore } from '../../stores/teamStore';
import { useAuthStore } from '../../stores/authStore';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Card, CardContent, CardHeader, CardTitle } from '../ui/card';

interface CreateTeamDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function CreateTeamDialog({ open, onOpenChange, onSuccess }: CreateTeamDialogProps) {
  const { user } = useAuthStore();
  const { createTeam, isLoading, error } = useTeamStore();

  const [name, setName] = useState('');
  const [key, setKey] = useState('');
  const [description, setDescription] = useState('');
  const [keyError, setKeyError] = useState('');
  const [localError, setLocalError] = useState('');

  const validateKey = (value: string): boolean => {
    if (value.length < 2 || value.length > 10) {
      setKeyError('团队标识符长度必须为 2-10 位');
      return false;
    }
    if (!/^[A-Z]/.test(value)) {
      setKeyError('团队标识符首字母必须为大写字母');
      return false;
    }
    if (!/^[A-Z][A-Z0-9]*$/.test(value)) {
      setKeyError('团队标识符只能包含大写字母和数字');
      return false;
    }
    setKeyError('');
    return true;
  };

  const handleKeyChange = (value: string) => {
    const upperValue = value.toUpperCase();
    setKey(upperValue);
    validateKey(upperValue);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLocalError('');

    if (!validateKey(key)) {
      return;
    }

    if (!user?.workspace_id) {
      setLocalError('无法获取工作区信息，请重新登录');
      return;
    }

    try {
      await createTeam({
        name,
        key,
        description,
        workspace_id: user.workspace_id,
      });
      setName('');
      setKey('');
      setDescription('');
      setKeyError('');
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
          <CardTitle>创建团队</CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {(error || localError) && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {localError || error}
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="teamName">团队名称</Label>
              <Input
                id="teamName"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="例如：产品团队"
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="teamKey">团队标识符</Label>
              <Input
                id="teamKey"
                value={key}
                onChange={(e) => handleKeyChange(e.target.value)}
                placeholder="例如：PROD"
                maxLength={10}
                required
              />
              {keyError && (
                <p className="text-sm text-destructive">{keyError}</p>
              )}
              <p className="text-xs text-muted-foreground">
                2-10 位大写字母和数字，首字母必须为大写字母
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">团队描述</Label>
              <textarea
                id="description"
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                rows={3}
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="关于此团队的简短描述"
              />
            </div>

            <div className="flex justify-end gap-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
              >
                取消
              </Button>
              <Button type="submit" disabled={isLoading || !name || !key || !!keyError}>
                {isLoading ? '创建中...' : '创建'}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
