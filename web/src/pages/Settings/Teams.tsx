/**
 * 团队管理页面
 */

import { TeamList } from '../../components/settings/TeamList';

export function TeamsPage() {
  return (
    <div className="container mx-auto py-6">
      <h1 className="mb-6 text-2xl font-bold">团队管理</h1>
      <TeamList />
    </div>
  );
}
