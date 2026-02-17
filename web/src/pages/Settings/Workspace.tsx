/**
 * 工作区设置页面
 */

import { WorkspaceSettings } from '../../components/settings/WorkspaceSettings';

export function WorkspaceSettingsPage() {
  return (
    <div className="container mx-auto py-6">
      <h1 className="mb-6 text-2xl font-bold">工作区设置</h1>
      <WorkspaceSettings />
    </div>
  );
}
