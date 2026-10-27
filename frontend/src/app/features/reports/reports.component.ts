import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { AnalyticsService } from '../../core/services/analytics.service';
import { FinancialRequest, ReportFilter } from '../../core/models';

@Component({
  selector: 'app-reports',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './reports.component.html',
  styleUrl: './reports.component.scss'
})
export class ReportsComponent implements OnInit {
  requests: FinancialRequest[] = [];
  isLoading = true;
  errorMessage = '';

  filter: ReportFilter = {
    start_date: '',
    end_date: '',
    type: 'ALL',
    status: 'ALL',
    min_amount: undefined,
    max_amount: undefined
  };

  constructor(private analyticsService: AnalyticsService) {}

  ngOnInit(): void {
    this.runReport();
  }

  runReport(): void {
    this.isLoading = true;
    this.errorMessage = '';

    this.analyticsService.getReports(this.filter).subscribe({
      next: (data) => {
        this.requests = data || [];
        this.isLoading = false;
      },
      error: (err) => {
        this.errorMessage = err.error?.error || 'Failed to generate report';
        this.isLoading = false;
      }
    });
  }

  resetFilters(): void {
    this.filter = {
      start_date: '',
      end_date: '',
      type: 'ALL',
      status: 'ALL',
      min_amount: undefined,
      max_amount: undefined
    };
    this.runReport();
  }

  exportCsv(): void {
    const timestamp = new Date().toISOString().split('T')[0];
    this.analyticsService.exportToCsv(this.requests, `financial_report_${timestamp}.csv`);
  }

  get totalAmount(): number {
    return this.requests.reduce((acc, r) => acc + r.amount, 0);
  }

  get averageAmount(): number {
    return this.requests.length > 0 ? this.totalAmount / this.requests.length : 0;
  }

  get approvedCount(): number {
    return this.requests.filter(r => r.status === 'APPROVED').length;
  }

  getStatusBadgeClass(status: string): string {
    switch (status) {
      case 'DRAFT': return 'status-draft';
      case 'PENDING_APPROVAL': return 'status-pending';
      case 'APPROVED': return 'status-approved';
      case 'REJECTED': return 'status-rejected';
      case 'CANCELLED': return 'status-cancelled';
      default: return '';
    }
  }

  getTypeIcon(type: string): string {
    switch (type) {
      case 'EXPENSE': return '💳';
      case 'PURCHASE': return '🛍️';
      case 'TRAVEL': return '✈️';
      default: return '💰';
    }
  }
}
