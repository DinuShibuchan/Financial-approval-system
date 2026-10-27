import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { RequestService } from '../../../core/services/request.service';
import { RequestType } from '../../../core/models';

@Component({
  selector: 'app-request-form',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  templateUrl: './request-form.component.html',
  styleUrl: './request-form.component.scss'
})
export class RequestFormComponent implements OnInit {
  isEditMode = false;
  requestId: number | null = null;

  requestType: RequestType = 'EXPENSE';
  amount: number | null = null;
  description = '';

  isLoading = false;
  errorMessage = '';
  successMessage = '';

  constructor(
    private requestService: RequestService,
    private route: ActivatedRoute,
    private router: Router
  ) {}

  ngOnInit(): void {
    const idParam = this.route.snapshot.paramMap.get('id');
    if (idParam) {
      this.requestId = Number(idParam);
      this.isEditMode = true;
      this.loadRequest(this.requestId);
    }
  }

  loadRequest(id: number): void {
    this.isLoading = true;
    this.requestService.getRequestById(id).subscribe({
      next: (req) => {
        this.isLoading = false;
        if (req.status !== 'DRAFT') {
          this.errorMessage = 'Only DRAFT requests can be edited.';
          return;
        }
        this.requestType = req.request_type;
        this.amount = req.amount;
        this.description = req.description;
      },
      error: () => {
        this.isLoading = false;
        this.errorMessage = 'Failed to load request.';
      }
    });
  }

  get tierInfo() {
    return this.requestService.getApprovalTierInfo(this.amount || 0);
  }

  save(submitNow: boolean): void {
    if (!this.amount || this.amount <= 0) {
      this.errorMessage = 'Please enter a valid amount greater than 0';
      return;
    }

    if (!this.description.trim()) {
      this.errorMessage = 'Please provide a justification / description';
      return;
    }

    this.isLoading = true;
    this.errorMessage = '';

    if (this.isEditMode && this.requestId) {
      this.requestService.updateRequest(this.requestId, {
        request_type: this.requestType,
        amount: this.amount,
        description: this.description
      }).subscribe({
        next: (updated) => {
          if (submitNow) {
            this.requestService.submitRequest(updated.id).subscribe({
              next: () => {
                this.router.navigate(['/requests']);
              },
              error: (err) => {
                this.isLoading = false;
                this.errorMessage = err.error?.error || 'Failed to submit request';
              }
            });
          } else {
            this.router.navigate(['/requests']);
          }
        },
        error: (err) => {
          this.isLoading = false;
          this.errorMessage = err.error?.error || 'Failed to update request';
        }
      });
    } else {
      this.requestService.createRequest({
        request_type: this.requestType,
        amount: this.amount,
        description: this.description,
        submit: submitNow
      }).subscribe({
        next: () => {
          this.router.navigate(['/requests']);
        },
        error: (err) => {
          this.isLoading = false;
          this.errorMessage = err.error?.error || 'Failed to create request';
        }
      });
    }
  }
}
