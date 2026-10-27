import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { RequestService } from '../../../core/services/request.service';
import { AuthService } from '../../../core/services/auth.service';
import { FinancialRequest, RequestStatus } from '../../../core/models';

@Component({
  selector: 'app-request-list',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './request-list.component.html',
  styleUrl: './request-list.component.scss'
})
export class RequestListComponent implements OnInit {
  requests: FinancialRequest[] = [];
  filteredRequests: FinancialRequest[] = [];

  isLoading = true;
  errorMessage = '';
  activeTab: 'my' | 'all' = 'my';
  selectedStatus: string = 'ALL';
  selectedType: string = 'ALL';
  searchQuery = '';

  constructor(
    private requestService: RequestService,
    public authService: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.loadRequests();
  }

  loadRequests(): void {
    this.isLoading = true;
    this.errorMessage = '';

    const req$ = this.activeTab === 'my'
      ? this.requestService.getMyRequests()
      : this.requestService.getAllRequests();

    req$.subscribe({
      next: (data) => {
        this.requests = data || [];
        this.applyFilters();
        this.isLoading = false;
      },
      error: (err) => {
        this.errorMessage = err.error?.error || 'Failed to load requests';
        this.isLoading = false;
      }
    });
  }

  switchTab(tab: 'my' | 'all'): void {
    if (this.activeTab !== tab) {
      this.activeTab = tab;
      this.loadRequests();
    }
  }

  filterByStatus(status: string): void {
    this.selectedStatus = status;
    this.applyFilters();
  }

  filterByType(type: string): void {
    this.selectedType = type;
    this.applyFilters();
  }

  applyFilters(): void {
    let result = [...this.requests];

    if (this.selectedStatus !== 'ALL') {
      result = result.filter(r => r.status === this.selectedStatus);
    }

    if (this.selectedType !== 'ALL') {
      result = result.filter(r => r.request_type === this.selectedType);
    }

    if (this.searchQuery.trim()) {
      const q = this.searchQuery.toLowerCase();
      result = result.filter(r =>
        r.request_number.toLowerCase().includes(q) ||
        r.description.toLowerCase().includes(q) ||
        (r.requested_by_name && r.requested_by_name.toLowerCase().includes(q))
      );
    }

    this.filteredRequests = result;
  }

  submitDraft(id: number): void {
    if (!confirm('Are you sure you want to submit this request for multi-level approval?')) return;

    this.requestService.submitRequest(id).subscribe({
      next: () => {
        this.loadRequests();
      },
      error: (err) => {
        alert(err.error?.error || 'Failed to submit request');
      }
    });
  }

  cancelRequest(id: number): void {
    if (!confirm('Are you sure you want to cancel this request?')) return;

    this.requestService.cancelRequest(id).subscribe({
      next: () => {
        this.loadRequests();
      },
      error: (err) => {
        alert(err.error?.error || 'Failed to cancel request');
      }
    });
  }

  deleteDraft(id: number): void {
    if (!confirm('Are you sure you want to permanently delete this draft?')) return;

    this.requestService.deleteRequest(id).subscribe({
      next: () => {
        this.loadRequests();
      },
      error: (err) => {
        alert(err.error?.error || 'Failed to delete request');
      }
    });
  }

  getStatusBadgeClass(status: RequestStatus): string {
    switch (status) {
      case 'DRAFT': return 'status-draft';
      case 'PENDING_APPROVAL': return 'status-pending';
      case 'APPROVED': return 'status-approved';
      case 'REJECTED': return 'status-rejected';
      case 'CANCELLED': return 'status-cancelled';
      default: return '';
    }
  }

  getStatusIcon(status: RequestStatus): string {
    switch (status) {
      case 'DRAFT': return '📝';
      case 'PENDING_APPROVAL': return '⏳';
      case 'APPROVED': return '✅';
      case 'REJECTED': return '❌';
      case 'CANCELLED': return '🚫';
      default: return '📄';
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

  isOwner(req: FinancialRequest): boolean {
    return this.authService.currentUser?.id === req.requested_by;
  }
}
