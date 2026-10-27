import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { 
  CategoryBreakdown, 
  DashboardSummary, 
  FinancialRequest, 
  RecentActivityItem, 
  ReportFilter, 
  TierBreakdown 
} from '../models';

@Injectable({
  providedIn: 'root'
})
export class AnalyticsService {
  private readonly BASE_URL = 'http://localhost:8085/api';

  constructor(private http: HttpClient) {}

  getDashboardSummary(): Observable<DashboardSummary> {
    return this.http.get<DashboardSummary>(`${this.BASE_URL}/dashboard/summary`);
  }

  getCategoryBreakdown(): Observable<CategoryBreakdown[]> {
    return this.http.get<CategoryBreakdown[]>(`${this.BASE_URL}/dashboard/by-type`);
  }

  getTierBreakdown(): Observable<TierBreakdown[]> {
    return this.http.get<TierBreakdown[]>(`${this.BASE_URL}/dashboard/by-amount`);
  }

  getRecentActivity(limit = 8): Observable<RecentActivityItem[]> {
    return this.http.get<RecentActivityItem[]>(`${this.BASE_URL}/dashboard/recent?limit=${limit}`);
  }

  getReports(filter: ReportFilter): Observable<FinancialRequest[]> {
    let params = new HttpParams();
    if (filter.start_date) params = params.set('start_date', filter.start_date);
    if (filter.end_date) params = params.set('end_date', filter.end_date);
    if (filter.type && filter.type !== 'ALL') params = params.set('type', filter.type);
    if (filter.status && filter.status !== 'ALL') params = params.set('status', filter.status);
    if (filter.min_amount) params = params.set('min_amount', filter.min_amount.toString());
    if (filter.max_amount) params = params.set('max_amount', filter.max_amount.toString());

    return this.http.get<FinancialRequest[]>(`${this.BASE_URL}/reports`, { params });
  }

  /**
   * Generates and downloads a CSV file from an array of financial requests
   */
  exportToCsv(requests: FinancialRequest[], filename = 'financial_report.csv'): void {
    if (!requests || requests.length === 0) return;

    const headers = [
      'Request Number',
      'Category',
      'Requested By',
      'Amount (INR)',
      'Status',
      'Approval Level',
      'Created Date',
      'Description'
    ];

    const rows = requests.map(r => [
      `"${r.request_number}"`,
      `"${r.request_type}"`,
      `"${r.requested_by_name || 'User #' + r.requested_by}"`,
      r.amount.toFixed(2),
      `"${r.status}"`,
      r.current_approval_level,
      `"${new Date(r.created_at).toISOString().split('T')[0]}"`,
      `"${(r.description || '').replace(/"/g, '""')}"`
    ]);

    const csvContent = [headers.join(','), ...rows.map(row => row.join(','))].join('\r\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute('download', filename);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  }
}
