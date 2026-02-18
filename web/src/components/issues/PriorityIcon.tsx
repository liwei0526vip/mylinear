/**
 * PriorityIcon 组件 - 显示 Issue 优先级图标
 * 参考 Linear 的优先级图标设计
 */

import { IssuePriority, getPriorityConfig } from '@/types/issue';

interface PriorityIconProps {
  priority: IssuePriority;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
  className?: string;
}

const sizeMap = {
  sm: 14,
  md: 16,
  lg: 20,
};

export function PriorityIcon({
  priority,
  size = 'md',
  showLabel = false,
  className = '',
}: PriorityIconProps) {
  const config = getPriorityConfig(priority);
  const iconSize = sizeMap[size];

  // 根据优先级返回不同的图标
  const renderIcon = () => {
    switch (priority) {
      case 0: // 无优先级 - 虚线圆
        return (
          <svg
            width={iconSize}
            height={iconSize}
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <circle
              cx="8"
              cy="8"
              r="6"
              stroke={config.color}
              strokeWidth="1.5"
              strokeDasharray="2 2"
            />
          </svg>
        );
      case 1: // 紧急 - 实心圆带边框
        return (
          <svg
            width={iconSize}
            height={iconSize}
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <circle cx="8" cy="8" r="6" stroke={config.color} strokeWidth="1.5" />
            <circle cx="8" cy="8" r="4" fill={config.color} />
          </svg>
        );
      case 2: // 高 - 两个实心竖线
        return (
          <svg
            width={iconSize}
            height={iconSize}
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <rect x="4" y="4" width="3" height="8" rx="1" fill={config.color} />
            <rect x="9" y="4" width="3" height="8" rx="1" fill={config.color} />
          </svg>
        );
      case 3: // 中 - 三个实心竖线（中间高）
        return (
          <svg
            width={iconSize}
            height={iconSize}
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <rect x="2" y="6" width="3" height="6" rx="1" fill={config.color} />
            <rect x="6.5" y="4" width="3" height="8" rx="1" fill={config.color} />
            <rect x="11" y="6" width="3" height="6" rx="1" fill={config.color} />
          </svg>
        );
      case 4: // 低 - 一个实心竖线
        return (
          <svg
            width={iconSize}
            height={iconSize}
            viewBox="0 0 16 16"
            fill="none"
            xmlns="http://www.w3.org/2000/svg"
          >
            <rect x="6.5" y="6" width="3" height="6" rx="1" fill={config.color} />
          </svg>
        );
      default:
        return null;
    }
  };

  return (
    <span className={`inline-flex items-center gap-1.5 ${className}`}>
      {renderIcon()}
      {showLabel && (
        <span style={{ color: config.color }} className="text-sm font-medium">
          {config.label}
        </span>
      )}
    </span>
  );
}

/**
 * PrioritySelector 组件 - 优先级选择器
 */
interface PrioritySelectorProps {
  value: IssuePriority;
  onChange: (priority: IssuePriority) => void;
  disabled?: boolean;
}

export function PrioritySelector({ value, onChange, disabled = false }: PrioritySelectorProps) {
  return (
    <select
      value={value}
      onChange={(e) => onChange(Number(e.target.value) as IssuePriority)}
      disabled={disabled}
      className="h-8 rounded-md border border-gray-300 bg-white px-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-500 disabled:opacity-50"
    >
      <option value={0}>无优先级</option>
      <option value={1}>紧急</option>
      <option value={2}>高</option>
      <option value={3}>中</option>
      <option value={4}>低</option>
    </select>
  );
}
