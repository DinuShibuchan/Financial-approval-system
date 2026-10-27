import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';
import { AnalyticsService } from '../../core/services/analytics.service';
import { 
  CategoryBreakdown, 
  DashboardSummary, 
  RecentActivityItem, 
  TierBreakdown, 
  User 
} from '../../core/models';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './dashboard.component.html',
  styleUrl: './dashboard.component.scss'
})
export class DashboardComponent implements OnInit {
  currentUser: User | null = null;
  summary: DashboardSummary | null = null;
  categoryBreakdown: CategoryBreakdown[] = [];
  tierBreakdown: TierBreakdown[] = [];
  recentActivities: RecentActivityItem[] = [];
  isLoading = true;

  constructor(
    public authService: AuthService,
    private analyticsService: AnalyticsService
  ) {}

  ngOnInit(): void {
    this.currentUser = this.authService.currentUser;
    this.loadDashboardData();
  }

  loadDashboardData(): void {
    this.isLoading = true;

    this.analyticsService.getDashboardSummary().subscribe({
      next: (data) => {
        this.summary = data;
      },
      error: () => {}
    });

    this.analyticsService.getCategoryBreakdown().subscribe({
      next: (data) => {
        this.categoryBreakdown = data || [];
      },
      error: () => {}
    });

    this.analyticsService.getTierBreakdown().subscribe({
      next: (data) => {
        this.tierBreakdown = data || [];
      },
      error: () => {}
    });

    this.analyticsService.getRecentActivity(6).subscribe({
      next: (data) => {
        this.recentActivities = data || [];
        this.isLoading = false;
      },
      error: () => {
        this.isLoading = false;
      }
    });
  }

  logout(): void {
    this.authService.logout();
  }

  getRoleBadgeClass(role: string): string {
    switch (role) {
      case 'DIRECTOR': return 'badge-director';
      case 'FINANCE': return 'badge-finance';
      case 'MANAGER': return 'badge-manager';
      default: return 'badge-employee';
    }
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
