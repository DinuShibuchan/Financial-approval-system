export interface User {
  id: number;
  name: string;
  email: string;
  roles: string[];
  created_at: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

export interface Role {
  id: number;
  role_name: string;
}

export type RequestType = 'EXPENSE' | 'PURCHASE' | 'TRAVEL';
export type RequestStatus = 'DRAFT' | 'PENDING_APPROVAL' | 'APPROVED' | 'REJECTED' | 'CANCELLED';

export interface FinancialRequest {
  id: number;
  request_number: string;
  request_type: RequestType;
  requested_by: number;
  requested_by_name?: string;
  amount: number;
  description: string;
  status: RequestStatus;
  current_approval_level: number;
  created_at: string;
  updated_at: string;
}

export interface CreateRequestPayload {
  request_type: RequestType;
  amount: number;
  description: string;
  submit?: boolean;
}

export interface ApprovalHistory {
  id: number;
  request_id: number;
  approval_level: number;
  approved_by: number;
  approved_by_name?: string;
  action: 'APPROVED' | 'REJECTED';
  reason?: string;
  acted_at: string;
}

export interface DashboardSummary {
  total_requests: number;
  pending_requests: number;
  approved_requests: number;
  rejected_requests: number;
  draft_requests: number;
  total_amount: number;
  approved_amount: number;
  pending_amount: number;
}

export interface CategoryBreakdown {
  type: string;
  count: number;
  total_amount: number;
  percentage: number;
}

export interface TierBreakdown {
  tier: number;
  tier_name: string;
  range_label: string;
  count: number;
  total_amount: number;
  approver: string;
}

export interface RecentActivityItem {
  id: number;
  request_number: string;
  request_type: string;
  requested_by: string;
  amount: number;
  status: string;
  timestamp: string;
}

export interface ReportFilter {
  start_date?: string;
  end_date?: string;
  type?: string;
  status?: string;
  min_amount?: number;
  max_amount?: number;
}
