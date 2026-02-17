/**
 * 团队详情组件
 */

import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useTeamStore } from '../../stores/teamStore';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { TeamMemberList } from './TeamMemberList';

export function TeamDetail() {
  const navigate = useNavigate();
  const { teamId } = useParams<{ teamId: string }>();
  const { currentTeam, isLoading, error, fetchTeam, updateTeam, deleteTeam } =
    useTeamStore();

  const [name, setName] = useState('');
  const [key, setKey] = useState('');
  const [description, setDescription] = useState('');
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  useEffect(() => {
    if (teamId) {
      fetchTeam(teamId);
    }
  }, [teamId]);

  useEffect(() => {
    if (currentTeam) {
      setName(currentTeam.name);
      setKey(currentTeam.key);
      setDescription(currentTeam.description || '');
    }
  }, [currentTeam]);

  const handleSave = async () => {
    if (!teamId) return;

    setSaving(true);
    try {
      await updateTeam(teamId, {
        name: name || undefined,
        key: key || undefined,
        description: description || undefined,
      });
    } catch {
      // 错误由 store 处理
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!teamId) return;

    if (!confirm('确定要删除此团队吗？此操作不可撤销。')) {
      return;
    }

    setDeleting(true);
    try {
      await deleteTeam(teamId);
      navigate('/settings/teams');
    } catch {
      // 错误由 store 处理
    } finally {
      setDeleting(false);
    }
  };

  if (isLoading && !currentTeam) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-muted-foreground">
          加载中...
        </CardContent>
      </Card>
    );
  }

  if (!currentTeam) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-muted-foreground">
          团队不存在或无权访问
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle>团队信息</CardTitle>
          <CardDescription>修改团队的基本信息</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && (
            <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {error}
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="name">团队名称</Label>
            <Input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="输入团队名称"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="key">团队标识符</Label>
            <Input
              id="key"
              value={key}
              onChange={(e) => setKey(e.target.value.toUpperCase())}
              placeholder="例如：PROD"
              maxLength={10}
            />
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

          <div className="flex justify-between">
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleting}
            >
              {deleting ? '删除中...' : '删除团队'}
            </Button>
            <Button onClick={handleSave} disabled={saving}>
              {saving ? '保存中...' : '保存更改'}
            </Button>
          </div>
        </CardContent>
      </Card>

      {teamId && <TeamMemberList teamId={teamId} />}
    </div>
  );
}
