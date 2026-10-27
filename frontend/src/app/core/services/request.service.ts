import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { ApprovalHistory, CreateRequestPayload, FinancialRequest, RequestType, RequestStatus } from '../models';

@Injectable({
  providedIn: 'root'
})
export class RequestService {
  private readonly BASE_URL = 'http://localhost:8085/api';
  private readonly API_URL = 'http://localhost:8085/api/requests';

  constructor(private http: HttpClient) {}

  createRequest(payload: CreateRequestPayload): Observable<FinancialRequest> {
    return this.http.post<FinancialRequest>(this.API_URL, payload);
  }

  getMyRequests(): Observable<FinancialRequest[]> {
    return this.http.get<FinancialRequest[]>(`${this.API_URL}/my`);
  }

  getAllRequests(status?: string, type?: string): Observable<FinancialRequest[]> {
    let params = new HttpParams();
    if (status) params = params.set('status', status);
    if (type) params = params.set('type', type);
    return this.http.get<FinancialRequest[]>(this.API_URL, { params });
  }

  getRequestById(id: number): Observable<FinancialRequest> {
    return this.http.get<FinancialRequest>(`${this.API_URL}/${id}`);
  }

  updateRequest(id: number, payload: Partial<CreateRequestPayload>): Observable<FinancialRequest> {
    return this.http.put<FinancialRequest>(`${this.API_URL}/${id}`, payload);
  }

  submitRequest(id: number): Observable<FinancialRequest> {
    return this.http.post<FinancialRequest>(`${this.API_URL}/${id}/submit`, {});
  }

  cancelRequest(id: number): Observable<FinancialRequest> {
    return this.http.post<FinancialRequest>(`${this.API_URL}/${id}/cancel`, {});
  }

  deleteRequest(id: number): Observable<{ message: string }> {
    return this.http.delete<{ message: string }>(`${this.API_URL}/${id}`);
  }

  // ==========================================
  // Approval Workflow APIs
  // ==========================================
  getApprovalHistory(requestId: number): Observable<ApprovalHistory[]> {
    return this.http.get<ApprovalHistory[]>(`${this.API_URL}/${requestId}/history`);
  }

  getPendingApprovals(): Observable<FinancialRequest[]> {
    return this.http.get<FinancialRequest[]>(`${this.BASE_URL}/approvals/pending`);
  }

  approveRequest(requestId: number, reason?: string): Observable<FinancialRequest> {
    return this.http.post<FinancialRequest>(`${this.API_URL}/${requestId}/approve`, { reason });
  }

  rejectRequest(requestId: number, reason: string): Observable<FinancialRequest> {
    return this.http.post<FinancialRequest>(`${this.API_URL}/${requestId}/reject`, { reason });
  }

  /**
   * Evaluates if a user can act on a pending request:
   * 1. Request must be in PENDING_APPROVAL
   * 2. User cannot approve their own request
   * 3. User must have the role required for current_approval_level
   */
  canUserApprove(request: FinancialRequest | null, userRoles: string[], userId: number | undefined): boolean {
    if (!request || !userId || !userRoles) return false;
    if (request.status !== 'PENDING_APPROVAL') return false;
    if (request.requested_by === userId) return false; // Self-approval disallowed

    switch (request.current_approval_level) {
      case 1:
        return userRoles.includes('MANAGER');
      case 2:
        return userRoles.includes('FINANCE');
      case 3:
        return userRoles.includes('DIRECTOR');
      default:
        return false;
    }
  }

  /**
   * Helper to compute required approval chain based on business rules:
   * < 10,000 => Tier 1 (Manager)
   * 10,000 - 50,000 => Tier 2 (Manager, Finance)
   * > 50,000 => Tier 3 (Manager, Finance, Director)
   */
  getApprovalTierInfo(amount: number): {
    tier: number;
    title: string;
    description: string;
    approvers: string[];
  } {
    if (amount <= 0 || isNaN(amount)) {
      return {
        tier: 0,
        title: 'Specify Amount',
        description: 'Enter an amount to see required approval chain',
        approvers: []
      };
    }
    if (amount < 10000) {
      return {
        tier: 1,
        title: 'Tier 1 (< ₹10,000)',
        description: 'Requires single-level approval from your Reporting Manager.',
        approvers: ['MANAGER']
      };
    }
    if (amount <= 50000) {
      return {
        tier: 2,
        title: 'Tier 2 (₹10,000 – ₹50,000)',
        description: 'Requires sequential approval: Manager first, followed by Finance Team.',
        approvers: ['MANAGER', 'FINANCE']
      };
    }
    return {
      tier: 3,
      title: 'Tier 3 (> ₹50,000)',
      description: 'Requires executive approval: Manager, Finance, and final Director signoff.',
      approvers: ['MANAGER', 'FINANCE', 'DIRECTOR']
    };
  }
}
