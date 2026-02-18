/**
 * Activity API
 */

import { api } from './client';
import type {
  Activity,
  ActivityListResponse,
  ActivityListParams,
} from '../types/activity';

// =============================================================================
// Activity API
// =============================================================================

/**
 * 获取 Issue 的活动列表
 */
export async function fetchIssueActivities(
  issueId: string,
  params?: ActivityListParams
) {
  const searchParams = new URLSearchParams();
  if (params?.page) searchParams.append('page', String(params.page));
  if (params?.page_size) searchParams.append('page_size', String(params.page_size));
  if (params?.types && params.types.length > 0) {
    searchParams.append('types', params.types.join(','));
  }

  const query = searchParams.toString();
  const endpoint = query
    ? `/issues/${issueId}/activities?${query}`
    : `/issues/${issueId}/activities`;

  return api<ActivityListResponse>(endpoint);
}
