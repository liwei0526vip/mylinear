/**
 * ProgressBar - 进度条组件
 */

import { cn } from '@/lib/utils';

interface ProgressBarProps {
  percent: number;
  size?: 'sm' | 'md' | 'lg';
  showLabel?: boolean;
  className?: string;
}

const sizeStyles = {
  sm: 'h-1.5',
  md: 'h-2',
  lg: 'h-3',
};

function getProgressColor(percent: number): string {
  if (percent >= 100) return '#22c55e'; // green
  if (percent >= 75) return '#84cc16'; // lime
  if (percent >= 50) return '#eab308'; // yellow
  if (percent >= 25) return '#f97316'; // orange
  return '#ef4444'; // red
}

export function ProgressBar({
  percent,
  size = 'md',
  showLabel = false,
  className,
}: ProgressBarProps) {
  const clampedPercent = Math.min(100, Math.max(0, percent));
  const color = getProgressColor(clampedPercent);

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <div
        className={cn(
          'w-full overflow-hidden rounded-full bg-gray-200 dark:bg-gray-700',
          sizeStyles[size]
        )}
      >
        <div
          className="h-full rounded-full transition-all duration-300"
          style={{
            width: `${clampedPercent}%`,
            backgroundColor: color,
          }}
        />
      </div>
      {showLabel && (
        <span className="min-w-[3rem] text-right text-sm text-gray-600 dark:text-gray-400">
          {clampedPercent.toFixed(0)}%
        </span>
      )}
    </div>
  );
}
