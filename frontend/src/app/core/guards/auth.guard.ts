import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { AuthService } from '../services/auth.service';

export const authGuard: CanActivateFn = (route, state) => {
  const authService = inject(AuthService);
  const router = inject(Router);

  if (authService.isLoggedIn) {
    // Check for required roles if specified in route data
    const requiredRoles = route.data?.['roles'] as string[];
    if (requiredRoles && requiredRoles.length > 0) {
      if (!authService.hasAnyRole(requiredRoles)) {
        router.navigate(['/dashboard']);
        return false;
      }
    }
    return true;
  }

  router.navigate(['/login'], { queryParams: { returnUrl: state.url } });
  return false;
};

export const loginGuard: CanActivateFn = (route, state) => {
  const authService = inject(AuthService);
  const router = inject(Router);

  if (!authService.isLoggedIn) {
    return true;
  }

  router.navigate(['/dashboard']);
  return false;
};
