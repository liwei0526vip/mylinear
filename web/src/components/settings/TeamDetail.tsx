import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useTeamStore } from '../../stores/teamStore';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { TeamMemberList } from './TeamMemberList';
import { TeamWorkflowSettings } from './TeamWorkflowSettings';
import { TeamLabelSettings } from './TeamLabelSettings';
import { Settings, Users, GitGraph, Tag } from 'lucide-react';

type TabType = 'general' | 'members' | 'workflow' | 'labels';

export function TeamDetail() {
  const navigate = useNavigate();
  const { teamId } = useParams<{ teamId: string }>();
  const { currentTeam, isLoading, error, fetchTeam, updateTeam, deleteTeam } =
    useTeamStore();

  const [activeTab, setActiveTab] = useState<TabType>('general');
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

  const tabs = [
    { id: 'general', label: '基本信息', icon: Settings },
    { id: 'members', label: '成员管理', icon: Users },
    { id: 'workflow', label: '工作流', icon: GitGraph },
    { id: 'labels', label: '标签', icon: Tag },
  ];

  return (
    <div className="space-y-6">
      {/* Tab 导航 */}
      <div className="flex border-b">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id as TabType)}
            className={`flex items-center gap-2 px-4 py-2 text-sm font-medium transition-colors border-b-2 -mb-[2px] ${activeTab === tab.id
                ? 'border-primary text-foreground'
                : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
          >
            <tab.icon className="h-4 w-4" />
            {tab.label}
          </button>
        ))}
      </div>

      <div className="mt-4">
        {activeTab === 'general' && (
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

              <div className="flex justify-between pt-4">
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
        )}

        {activeTab === 'members' && teamId && <TeamMemberList teamId={teamId} />}

        {activeTab === 'workflow' && teamId && <TeamWorkflowSettings teamId={teamId} />}

        {activeTab === 'labels' && teamId && <TeamLabelSettings teamId={teamId} />}
      </div>
    </div>
  );
}
