import { Routes } from '@angular/router';
import { LoginComponent } from './features/auth/login/login.component';
import { RegisterComponent } from './features/auth/register/register.component';
import { DashboardComponent } from './features/dashboard/dashboard.component';
import { RequestListComponent } from './features/requests/request-list/request-list.component';
import { RequestFormComponent } from './features/requests/request-form/request-form.component';
import { RequestDetailComponent } from './features/requests/request-detail/request-detail.component';
import { ApprovalQueueComponent } from './features/approvals/approval-queue/approval-queue.component';
import { ReportsComponent } from './features/reports/reports.component';
import { authGuard, loginGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  {
    path: 'login',
    component: LoginComponent,
    canActivate: [loginGuard]
  },
  {
    path: 'register',
    component: RegisterComponent,
    canActivate: [loginGuard]
  },
  {
    path: 'dashboard',
    component: DashboardComponent,
    canActivate: [authGuard]
  },
  {
    path: 'requests',
    component: RequestListComponent,
    canActivate: [authGuard]
  },
  {
    path: 'requests/new',
    component: RequestFormComponent,
    canActivate: [authGuard]
  },
  {
    path: 'requests/:id',
    component: RequestDetailComponent,
    canActivate: [authGuard]
  },
  {
    path: 'requests/:id/edit',
    component: RequestFormComponent,
    canActivate: [authGuard]
  },
  {
    path: 'approvals',
    component: ApprovalQueueComponent,
    canActivate: [authGuard]
  },
  {
    path: 'reports',
    component: ReportsComponent,
    canActivate: [authGuard]
  },
  {
    path: '',
    redirectTo: '/dashboard',
    pathMatch: 'full'
  },
  {
    path: '**',
    redirectTo: '/dashboard'
  }
];
