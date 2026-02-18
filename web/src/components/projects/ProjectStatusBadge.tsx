/**
 * ProjectStatusBadge - 项目状态徽章组件
 */

import { cn } from '@/lib/utils';
import { getStatusConfig, type ProjectStatus } from '@/types/project';

interface ProjectStatusBadgeProps {
  status: ProjectStatus;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const sizeStyles = {
  sm: 'px-2 py-0.5 text-xs',
  md: 'px-2.5 py-1 text-sm',
  lg: 'px-3 py-1.5 text-base',
};

export function ProjectStatusBadge({
  status,
  size = 'md',
  className,
}: ProjectStatusBadgeProps) {
  const config = getStatusConfig(status);

  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full font-medium',
        sizeStyles[size],
        className
      )}
      style={{
        backgroundColor: `${config.color}20`,
        color: config.color,
      }}
    >
      {config.label}
    </span>
  );
}

/**
 * ProjectStatusIcon - 项目状态图标组件
 */
export function ProjectStatusIcon({
  status,
  size = 16,
  className,
}: {
  status: ProjectStatus;
  size?: number;
  className?: string;
}) {
  const config = getStatusConfig(status);

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 16 16"
      fill="none"
      className={className}
    >
      {status === 'planned' && (
        <circle
          cx="8"
          cy="8"
          r="6"
          stroke={config.color}
          strokeWidth="1.5"
          strokeDasharray="3 2"
        />
      )}
      {status === 'in_progress' && (
        <>
          <circle
            cx="8"
            cy="8"
            r="6"
            stroke={config.color}
            strokeWidth="1.5"
          />
          <path
            d="M8 4 L8 8 L11 10"
            stroke={config.color}
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </>
      )}
      {status === 'paused' && (
        <>
          <circle
            cx="8"
            cy="8"
            r="6"
            stroke={config.color}
            strokeWidth="1.5"
          />
          <rect x="5" y="5" width="2" height="6" rx="0.5" fill={config.color} />
          <rect x="9" y="5" width="2" height="6" rx="0.5" fill={config.color} />
        </>
      )}
      {status === 'completed' && (
        <>
          <circle cx="8" cy="8" r="6" fill={config.color} />
          <path
            d="M5 8 L7 10 L11 6"
            stroke="white"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </>
      )}
      {status === 'cancelled' && (
        <>
          <circle
            cx="8"
            cy="8"
            r="6"
            stroke={config.color}
            strokeWidth="1.5"
          />
          <path
            d="M5 5 L11 11 M11 5 L5 11"
            stroke={config.color}
            strokeWidth="1.5"
            strokeLinecap="round"
          />
        </>
      )}
    </svg>
  );
}
