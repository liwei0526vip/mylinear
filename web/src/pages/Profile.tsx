/**
 * Profile 设置页面
 */

import { useEffect, useState } from 'react';
import { useAuthStore } from '../stores/authStore';
import { AvatarUpload } from '../components/AvatarUpload';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  CardFooter,
} from '../components/ui/card';
import { Separator } from '../components/ui/separator';

export default function ProfilePage() {
  const { user, updateUser, refreshUser, isLoading, error, clearError } = useAuthStore();

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [successMessage, setSuccessMessage] = useState('');
  const [validationError, setValidationError] = useState('');

  // 初始化表单数据
  useEffect(() => {
    if (user) {
      setName(user.name);
      setEmail(user.email);
      setUsername(user.username);
    }
  }, [user]);

  const validateForm = (): boolean => {
    if (!name.trim()) {
      setValidationError('姓名不能为空');
      return false;
    }
    if (!email.trim()) {
      setValidationError('邮箱不能为空');
      return false;
    }
    if (!username.trim()) {
      setValidationError('用户名不能为空');
      return false;
    }
    if (username.length < 3) {
      setValidationError('用户名至少需要 3 个字符');
      return false;
    }
    setValidationError('');
    return true;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    clearError();
    setSuccessMessage('');

    if (!validateForm()) {
      return;
    }

    try {
      await updateUser({
        name,
        email,
        username,
      });
      setSuccessMessage('保存成功');
      setTimeout(() => setSuccessMessage(''), 3000);
    } catch {
      // 错误已在 store 中处理
    }
  };

  const handleReset = () => {
    if (user) {
      setName(user.name);
      setEmail(user.email);
      setUsername(user.username);
    }
    clearError();
    setSuccessMessage('');
    setValidationError('');
  };

  const handleAvatarUploadSuccess = async (_avatarUrl: string) => {
    setSuccessMessage('头像更新成功');
    await refreshUser();
    setTimeout(() => setSuccessMessage(''), 3000);
  };

  const displayError = validationError || error;

  if (!user) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-muted-foreground">加载中...</div>
      </div>
    );
  }

  return (
    <div className="container max-w-2xl py-8">
      <h1 className="mb-6 text-2xl font-bold">设置</h1>

      {/* 成功提示 */}
      {successMessage && (
        <div className="mb-4 rounded-md bg-green-500/10 p-3 text-sm text-green-600">
          {successMessage}
        </div>
      )}

      {/* 个人资料卡片 */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>个人资料</CardTitle>
          <CardDescription>管理您的账户信息</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* 错误提示 */}
            {displayError && (
              <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
                {displayError}
              </div>
            )}

            {/* 姓名 */}
            <div className="space-y-2">
              <Label htmlFor="name">姓名</Label>
              <Input
                id="name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                disabled={isLoading}
              />
            </div>

            {/* 邮箱 */}
            <div className="space-y-2">
              <Label htmlFor="email">邮箱</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={isLoading}
              />
            </div>

            {/* 用户名 */}
            <div className="space-y-2">
              <Label htmlFor="username">用户名</Label>
              <Input
                id="username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={isLoading}
              />
            </div>

            {/* 按钮组 */}
            <div className="flex gap-3">
              <Button type="submit" disabled={isLoading}>
                {isLoading ? '保存中...' : '保存更改'}
              </Button>
              <Button
                type="button"
                variant="outline"
                onClick={handleReset}
                disabled={isLoading}
              >
                重置
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>

      {/* 头像卡片 */}
      <Card className="mb-6">
        <CardHeader>
          <CardTitle>头像</CardTitle>
          <CardDescription>上传您的个人头像</CardDescription>
        </CardHeader>
        <CardContent>
          <AvatarUpload
            currentAvatarUrl={user.avatar_url}
            userName={user.name}
            onUploadSuccess={handleAvatarUploadSuccess}
          />
        </CardContent>
      </Card>

      {/* 账户信息卡片 */}
      <Card>
        <CardHeader>
          <CardTitle>账户信息</CardTitle>
          <CardDescription>您的账户详细信息</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">角色</p>
              <p className="text-xs text-muted-foreground">角色由系统管理员分配</p>
            </div>
            <span className="rounded-full bg-primary/10 px-3 py-1 text-sm font-medium text-primary">
              {user.role}
            </span>
          </div>
          <Separator />
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium">用户 ID</p>
              <p className="text-xs text-muted-foreground">您的唯一用户标识</p>
            </div>
            <code className="rounded bg-muted px-2 py-1 text-xs">{user.id}</code>
          </div>
        </CardContent>
        <CardFooter>
          <p className="text-xs text-muted-foreground">
            如需修改角色或其他账户信息，请联系系统管理员
          </p>
        </CardFooter>
      </Card>
    </div>
  );
}
