import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { RequestService } from '../../../core/services/request.service';
import { AuthService } from '../../../core/services/auth.service';
import { FinancialRequest } from '../../../core/models';

@Component({
  selector: 'app-approval-queue',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './approval-queue.component.html',
  styleUrl: './approval-queue.component.scss'
})
export class ApprovalQueueComponent implements OnInit {
  pendingRequests: FinancialRequest[] = [];
  isLoading = true;
  isActionLoading = false;
  errorMessage = '';

  // Decision Modal
  selectedRequest: FinancialRequest | null = null;
  showApproveModal = false;
  showRejectModal = false;
  actionNote = '';
  rejectionReason = '';

  constructor(
    private requestService: RequestService,
    public authService: AuthService
  ) {}

  ngOnInit(): void {
    this.loadQueue();
  }

  loadQueue(): void {
    this.isLoading = true;
    this.errorMessage = '';
    this.requestService.getPendingApprovals().subscribe({
      next: (data) => {
        this.pendingRequests = data || [];
        this.isLoading = false;
      },
      error: (err) => {
        this.errorMessage = err.error?.error || 'Failed to load approval queue';
        this.isLoading = false;
      }
    });
  }

  openApprove(req: FinancialRequest): void {
    this.selectedRequest = req;
    this.actionNote = '';
    this.showApproveModal = true;
  }

  confirmApprove(): void {
    if (!this.selectedRequest) return;
    this.isActionLoading = true;

    this.requestService.approveRequest(this.selectedRequest.id, this.actionNote).subscribe({
      next: () => {
        this.isActionLoading = false;
        this.showApproveModal = false;
        this.selectedRequest = null;
        this.loadQueue();
      },
      error: (err) => {
        this.isActionLoading = false;
        alert(err.error?.error || 'Failed to approve request');
      }
    });
  }

  openReject(req: FinancialRequest): void {
    this.selectedRequest = req;
    this.rejectionReason = '';
    this.showRejectModal = true;
  }

  confirmReject(): void {
    if (!this.selectedRequest) return;
    if (!this.rejectionReason.trim()) {
      alert('Please enter a mandatory rejection reason.');
      return;
    }

    this.isActionLoading = true;
    this.requestService.rejectRequest(this.selectedRequest.id, this.rejectionReason).subscribe({
      next: () => {
        this.isActionLoading = false;
        this.showRejectModal = false;
        this.selectedRequest = null;
        this.loadQueue();
      },
      error: (err) => {
        this.isActionLoading = false;
        alert(err.error?.error || 'Failed to reject request');
      }
    });
  }

  getLevelBadge(level: number): string {
    switch (level) {
      case 1: return 'badge-manager';
      case 2: return 'badge-finance';
      case 3: return 'badge-director';
      default: return 'badge-employee';
    }
  }

  getLevelRole(level: number): string {
    switch (level) {
      case 1: return 'Manager Review';
      case 2: return 'Finance Review';
      case 3: return 'Director Review';
      default: return 'Review';
    }
  }
}
