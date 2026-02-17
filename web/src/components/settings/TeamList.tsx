/**
 * 团队列表组件
 */

import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTeamStore } from '../../stores/teamStore';
import { useAuthStore } from '../../stores/authStore';
import { Button } from '../ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/card';
import { CreateTeamDialog } from './CreateTeamDialog';

export function TeamList() {
  const navigate = useNavigate();
  const { user } = useAuthStore();
  const { teams, isLoading, error, fetchTeams } = useTeamStore();
  const [showCreateDialog, setShowCreateDialog] = useState(false);

  useEffect(() => {
    if (user?.workspace_id) {
      fetchTeams(user.workspace_id);
    }
  }, [user?.workspace_id]);

  const handleTeamClick = (teamId: string) => {
    navigate(`/settings/teams/${teamId}`);
  };

  const handleCreateSuccess = () => {
    if (user?.workspace_id) {
      fetchTeams(user.workspace_id);
    }
    setShowCreateDialog(false);
  };

  return (
    <>
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle>团队管理</CardTitle>
            <CardDescription>管理工作区中的团队</CardDescription>
          </div>
          <Button onClick={() => setShowCreateDialog(true)}>创建团队</Button>
        </CardHeader>
        <CardContent>
          {error && (
            <div className="mb-4 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
              {error}
            </div>
          )}

          {isLoading ? (
            <div className="py-8 text-center text-muted-foreground">加载中...</div>
          ) : teams.length === 0 ? (
            <div className="py-8 text-center text-muted-foreground">
              暂无团队，点击"创建团队"添加第一个团队
            </div>
          ) : (
            <div className="space-y-2">
              {teams.map((team) => (
                <div
                  key={team.id}
                  onClick={() => handleTeamClick(team.id)}
                  className="flex cursor-pointer items-center justify-between rounded-lg border p-4 hover:bg-accent"
                >
                  <div>
                    <div className="font-medium">{team.name}</div>
                    <div className="text-sm text-muted-foreground">{team.key}</div>
                  </div>
                  <div className="text-muted-foreground">→</div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      <CreateTeamDialog
        open={showCreateDialog}
        onOpenChange={setShowCreateDialog}
        onSuccess={handleCreateSuccess}
      />
    </>
  );
}
