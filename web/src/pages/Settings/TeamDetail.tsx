/**
 * 团队详情页面
 */

import { useNavigate } from 'react-router-dom';
import { TeamDetail } from '../../components/settings/TeamDetail';
import { Button } from '../../components/ui/button';

export function TeamDetailPage() {
  const navigate = useNavigate();

  return (
    <div className="container mx-auto py-6">
      <div className="mb-6 flex items-center gap-4">
        <Button variant="outline" onClick={() => navigate('/settings/teams')}>
          ← 返回
        </Button>
        <h1 className="text-2xl font-bold">团队详情</h1>
      </div>
      <TeamDetail />
    </div>
  );
}
