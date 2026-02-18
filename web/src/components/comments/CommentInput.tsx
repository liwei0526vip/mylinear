/**
 * CommentInput 组件 - 评论输入框
 */

import { useState, useRef, useEffect } from 'react';

interface CommentInputProps {
  onSubmit: (body: string) => Promise<void>;
  onCancel?: () => void;
  initialValue?: string;
  placeholder?: string;
  submitText?: string;
  isReply?: boolean;
  disabled?: boolean;
}

export function CommentInput({
  onSubmit,
  onCancel,
  initialValue = '',
  placeholder = '添加评论...（支持 @mention 提及团队成员）',
  submitText = '评论',
  isReply = false,
  disabled = false,
}: CommentInputProps) {
  const [body, setBody] = useState(initialValue);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isFocused, setIsFocused] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  // 自动聚焦
  useEffect(() => {
    if (isFocused && textareaRef.current) {
      textareaRef.current.focus();
    }
  }, [isFocused]);

  // 自动调整高度
  useEffect(() => {
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`;
    }
  }, [body]);

  // 提交评论
  const handleSubmit = async () => {
    if (!body.trim() || isSubmitting || disabled) return;

    setIsSubmitting(true);
    try {
      await onSubmit(body.trim());
      setBody('');
      setIsFocused(false);
    } catch {
      // 错误在调用方处理
    } finally {
      setIsSubmitting(false);
    }
  };

  // 键盘事件处理
  const handleKeyDown = (e: React.KeyboardEvent) => {
    // Cmd/Ctrl + Enter 提交
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
      e.preventDefault();
      handleSubmit();
    }
    // Escape 取消
    if (e.key === 'Escape' && onCancel) {
      e.preventDefault();
      onCancel();
    }
  };

  // 检测 @mention
  const handleInput = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const value = e.target.value;
    setBody(value);

    // TODO: 实现 @mention 下拉选择器
    // 当检测到 @ 符号时，显示团队成员列表
  };

  const isActive = isFocused || body.length > 0;

  return (
    <div className={`rounded-lg border ${isActive ? 'border-indigo-300 ring-1 ring-indigo-100' : 'border-gray-200'}`}>
      {/* 输入区域 */}
      <div className="p-3">
        <textarea
          ref={textareaRef}
          value={body}
          onChange={handleInput}
          onFocus={() => setIsFocused(true)}
          onBlur={() => {
            // 延迟失焦，让按钮点击有机会触发
            setTimeout(() => setIsFocused(false), 150);
          }}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          disabled={disabled || isSubmitting}
          rows={1}
          className="w-full resize-none border-0 bg-transparent text-sm text-gray-700 placeholder-gray-400 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
          style={{ minHeight: '24px' }}
        />
      </div>

      {/* 操作按钮 */}
      {isActive && (
        <div className="flex items-center justify-between border-t border-gray-100 px-3 py-2">
          <div className="flex items-center gap-2 text-xs text-gray-400">
            <span>⌘ + Enter 发送</span>
            {isReply && <span>Esc 取消</span>}
          </div>
          <div className="flex items-center gap-2">
            {onCancel && (
              <button
                onClick={onCancel}
                disabled={isSubmitting}
                className="rounded px-3 py-1 text-sm text-gray-500 hover:bg-gray-100 disabled:opacity-50"
              >
                取消
              </button>
            )}
            <button
              onClick={handleSubmit}
              disabled={!body.trim() || isSubmitting || disabled}
              className="rounded bg-indigo-500 px-3 py-1 text-sm text-white hover:bg-indigo-600 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {isSubmitting ? '发送中...' : submitText}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
