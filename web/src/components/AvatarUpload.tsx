/**
 * 头像上传组件
 */

import { useState, useRef } from 'react';
import { uploadAvatar } from '../api/user';
import { Button } from './ui/button';

interface AvatarUploadProps {
  currentAvatarUrl?: string;
  userName: string;
  onUploadSuccess: (avatarUrl: string) => void;
}

export function AvatarUpload({
  currentAvatarUrl,
  userName,
  onUploadSuccess,
}: AvatarUploadProps) {
  const [isUploading, setIsUploading] = useState(false);
  const [preview, setPreview] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const MAX_FILE_SIZE = 2 * 1024 * 1024; // 2MB
  const ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];

  const validateFile = (file: File): string | null => {
    if (!ALLOWED_TYPES.includes(file.type)) {
      return '不支持的文件类型，仅支持 JPG、PNG、GIF、WebP';
    }
    if (file.size > MAX_FILE_SIZE) {
      return '文件大小不能超过 2MB';
    }
    return null;
  };

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    setError(null);

    const validationError = validateFile(file);
    if (validationError) {
      setError(validationError);
      return;
    }

    // 创建预览
    const reader = new FileReader();
    reader.onload = (event) => {
      setPreview(event.target?.result as string);
    };
    reader.readAsDataURL(file);
  };

  const handleUpload = async () => {
    const file = fileInputRef.current?.files?.[0];
    if (!file) {
      setError('请先选择文件');
      return;
    }

    setIsUploading(true);
    setError(null);

    try {
      const result = await uploadAvatar(file);
      onUploadSuccess(result.avatar_url);
      setPreview(null);
      // 清空文件输入
      if (fileInputRef.current) {
        fileInputRef.current.value = '';
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : '上传失败';
      setError(message);
    } finally {
      setIsUploading(false);
    }
  };

  const handleCancel = () => {
    setPreview(null);
    setError(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  const displayAvatar = preview || currentAvatarUrl;

  return (
    <div className="space-y-4">
      {/* 头像预览 */}
      <div className="flex items-center gap-4">
        {displayAvatar ? (
          <img
            src={displayAvatar}
            alt={userName}
            className="h-20 w-20 rounded-full object-cover"
          />
        ) : (
          <div className="flex h-20 w-20 items-center justify-center rounded-full bg-muted text-2xl font-medium text-muted-foreground">
            {userName.charAt(0).toUpperCase()}
          </div>
        )}
        <div className="flex-1">
          <p className="text-sm font-medium">更换头像</p>
          <p className="text-xs text-muted-foreground">
            支持 JPG、PNG、GIF、WebP，最大 2MB
          </p>
        </div>
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}

      {/* 文件选择 */}
      <div className="flex items-center gap-3">
        <input
          ref={fileInputRef}
          type="file"
          accept="image/jpeg,image/png,image/gif,image/webp"
          onChange={handleFileSelect}
          className="hidden"
          id="avatar-upload"
        />
        <label htmlFor="avatar-upload">
          <Button variant="outline" asChild disabled={isUploading}>
            <span>选择图片</span>
          </Button>
        </label>

        {preview && (
          <>
            <Button onClick={handleUpload} disabled={isUploading}>
              {isUploading ? '上传中...' : '确认上传'}
            </Button>
            <Button
              variant="ghost"
              onClick={handleCancel}
              disabled={isUploading}
            >
              取消
            </Button>
          </>
        )}
      </div>
    </div>
  );
}
