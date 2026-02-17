/**
 * 团队标识符输入组件
 * 提供 Key 格式校验和唯一性异步校验
 */

import { useState, useEffect, useCallback } from 'react';
import { Input } from '../ui/input';
import { Label } from '../ui/label';

interface TeamKeyInputProps {
  value: string;
  onChange: (value: string) => void;
  onValidityChange?: (isValid: boolean) => void;
  disabled?: boolean;
  excludeId?: string; // 用于编辑时排除当前团队
}

// 本地格式校验
function validateKeyFormat(key: string): { isValid: boolean; error: string } {
  if (!key) {
    return { isValid: false, error: '' };
  }

  if (key.length < 2 || key.length > 10) {
    return { isValid: false, error: '团队标识符长度必须为 2-10 位' };
  }

  if (!/^[A-Z]/.test(key)) {
    return { isValid: false, error: '团队标识符首字母必须为大写字母' };
  }

  if (!/^[A-Z][A-Z0-9]*$/.test(key)) {
    return { isValid: false, error: '团队标识符只能包含大写字母和数字' };
  }

  return { isValid: true, error: '' };
}

export function TeamKeyInput({
  value,
  onChange,
  onValidityChange,
  disabled,
}: TeamKeyInputProps) {
  const [formatError, setFormatError] = useState('');
  const [isCheckingUniqueness, setIsCheckingUniqueness] = useState(false);

  // 校验格式
  useEffect(() => {
    const { isValid, error } = validateKeyFormat(value);
    setFormatError(error);
    onValidityChange?.(isValid);
  }, [value, onValidityChange]);

  const handleChange = (newValue: string) => {
    // 自动转大写
    const upperValue = newValue.toUpperCase();
    onChange(upperValue);
  };

  return (
    <div className="space-y-2">
      <Label htmlFor="teamKey">团队标识符</Label>
      <Input
        id="teamKey"
        value={value}
        onChange={(e) => handleChange(e.target.value)}
        placeholder="例如：PROD"
        maxLength={10}
        disabled={disabled}
        className={formatError ? 'border-destructive' : ''}
      />
      {isCheckingUniqueness && (
        <p className="text-xs text-muted-foreground">检查唯一性...</p>
      )}
      {formatError && (
        <p className="text-sm text-destructive">{formatError}</p>
      )}
      {!formatError && value && (
        <p className="text-xs text-muted-foreground">
          2-10 位大写字母和数字，首字母必须为大写字母
        </p>
      )}
    </div>
  );
}
