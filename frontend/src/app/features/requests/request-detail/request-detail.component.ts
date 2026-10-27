import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { RequestService } from '../../../core/services/request.service';
import { AuthService } from '../../../core/services/auth.service';
import { ApprovalHistory, FinancialRequest } from '../../../core/models';

@Component({
  selector: 'app-request-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './request-detail.component.html',
  styleUrl: './request-detail.component.scss'
})
export class RequestDetailComponent implements OnInit {
  request: FinancialRequest | null = null;
  history: ApprovalHistory[] = [];
  isLoading = true;
  isActionLoading = false;
  errorMessage = '';

  // Approver action modals/fields
  showApproveModal = false;
  showRejectModal = false;
  actionNote = '';
  rejectionReason = '';

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private requestService: RequestService,
    public authService: AuthService
  ) {}

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.loadRequest(Number(id));
      this.loadHistory(Number(id));
    } else {
      this.errorMessage = 'No request ID provided';
      this.isLoading = false;
    }
  }

  loadRequest(id: number): void {
    this.isLoading = true;
    this.requestService.getRequestById(id).subscribe({
      next: (req) => {
        this.request = req;
        this.isLoading = false;
      },
      error: (err) => {
        this.errorMessage = err.error?.error || 'Failed to load request details';
        this.isLoading = false;
      }
    });
  }

  loadHistory(id: number): void {
    this.requestService.getApprovalHistory(id).subscribe({
      next: (hist) => {
        this.history = hist || [];
      },
      error: () => {
        // history may be empty initially
        this.history = [];
      }
    });
  }

  get tierInfo() {
    return this.request ? this.requestService.getApprovalTierInfo(this.request.amount) : null;
  }

  get isOwner(): boolean {
    return this.authService.currentUser?.id === this.request?.requested_by;
  }

  get canCurrentUserApprove(): boolean {
    const roles = this.authService.currentUser?.roles || [];
    const userId = this.authService.currentUser?.id;
    return this.requestService.canUserApprove(this.request, roles, userId);
  }

  get currentLevelRoleName(): string {
    if (!this.request) return '';
    switch (this.request.current_approval_level) {
      case 1: return 'Manager';
      case 2: return 'Finance';
      case 3: return 'Director';
      default: return '';
    }
  }

  submitDraft(): void {
    if (!this.request) return;
    if (!confirm('Submit this request for approval?')) return;

    this.requestService.submitRequest(this.request.id).subscribe({
      next: (updated) => {
        this.request = updated;
        this.loadHistory(updated.id);
      },
      error: (err) => {
        alert(err.error?.error || 'Failed to submit request');
      }
    });
  }

  cancelRequest(): void {
    if (!this.request) return;
    if (!confirm('Cancel this financial request?')) return;

    this.requestService.cancelRequest(this.request.id).subscribe({
      next: (updated) => {
        this.request = updated;
        this.loadHistory(updated.id);
      },
      error: (err) => {
        alert(err.error?.error || 'Failed to cancel request');
      }
    });
  }

  openApproveModal(): void {
    this.actionNote = '';
    this.showApproveModal = true;
  }

  confirmApprove(): void {
    if (!this.request) return;
    this.isActionLoading = true;

    this.requestService.approveRequest(this.request.id, this.actionNote).subscribe({
      next: (updated) => {
        this.request = updated;
        this.isActionLoading = false;
        this.showApproveModal = false;
        this.loadHistory(updated.id);
      },
      error: (err) => {
        this.isActionLoading = false;
        alert(err.error?.error || 'Failed to approve request');
      }
    });
  }

  openRejectModal(): void {
    this.rejectionReason = '';
    this.showRejectModal = true;
  }

  confirmReject(): void {
    if (!this.request) return;
    if (!this.rejectionReason.trim()) {
      alert('Please provide a mandatory reason for rejection.');
      return;
    }

    this.isActionLoading = true;
    this.requestService.rejectRequest(this.request.id, this.rejectionReason).subscribe({
      next: (updated) => {
        this.request = updated;
        this.isActionLoading = false;
        this.showRejectModal = false;
        this.loadHistory(updated.id);
      },
      error: (err) => {
        this.isActionLoading = false;
        alert(err.error?.error || 'Failed to reject request');
      }
    });
  }

  getStepStatus(stepLevel: number): 'completed' | 'active' | 'waiting' | 'rejected' {
    if (!this.request) return 'waiting';
    if (this.request.status === 'APPROVED') return 'completed';
    if (this.request.status === 'REJECTED' && this.request.current_approval_level === stepLevel) return 'rejected';

    if (this.request.status === 'PENDING_APPROVAL') {
      if (this.request.current_approval_level > stepLevel) return 'completed';
      if (this.request.current_approval_level === stepLevel) return 'active';
    }
    return 'waiting';
  }
}
