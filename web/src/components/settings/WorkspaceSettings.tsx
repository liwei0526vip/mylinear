/**
 * 工作区设置组件
 */

import { useState, useEffect } from 'react';
import { useWorkspaceStore } from '../../stores/workspaceStore';
import { useAuthStore } from '../../stores/authStore';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';

export function WorkspaceSettings() {
  const { user } = useAuthStore();
  const { workspace, isLoading, error, fetchWorkspace, updateWorkspace, clearError } =
    useWorkspaceStore();

  const [name, setName] = useState('');
  const [logoUrl, setLogoUrl] = useState('');
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (user?.workspace_id) {
      fetchWorkspace(user.workspace_id);
    }
  }, [user?.workspace_id]);

  useEffect(() => {
    if (workspace) {
      setName(workspace.name);
      setLogoUrl(workspace.logo_url || '');
    }
  }, [workspace]);

  const handleSave = async () => {
    if (!workspace) return;

    setSaving(true);
    try {
      await updateWorkspace(workspace.id, {
        name: name || undefined,
        logo_url: logoUrl || undefined,
      });
    } catch {
      // 错误由 store 处理
    } finally {
      setSaving(false);
    }
  };

  if (isLoading && !workspace) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-muted-foreground">
          加载中...
        </CardContent>
      </Card>
    );
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>工作区设置</CardTitle>
        <CardDescription>管理工作区的基本信息</CardDescription>
      </CardHeader>
      <CardContent className="space-y-4">
        {error && (
          <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
            {error}
          </div>
        )}

        <div className="space-y-2">
          <Label htmlFor="name">工作区名称</Label>
          <Input
            id="name"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="输入工作区名称"
          />
        </div>

        <div className="space-y-2">
          <Label htmlFor="logoUrl">Logo URL</Label>
          <Input
            id="logoUrl"
            value={logoUrl}
            onChange={(e) => setLogoUrl(e.target.value)}
            placeholder="https://example.com/logo.png"
          />
        </div>

        <div className="flex justify-end">
          <Button onClick={handleSave} disabled={saving || isLoading}>
            {saving ? '保存中...' : '保存更改'}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
