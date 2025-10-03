import { Routes } from '@angular/router';
import { Auth } from './auth/auth';
import { AuthLayout } from './layouts/auth-layout/auth-layout';
import { inject } from '@angular/core';
import { Authservice } from './shared/services/authservice';

export const routes: Routes = [
  {
    path: 'auth',
    component: AuthLayout,
    children: [{ path: '', component: Auth }],
  },
  {
    path: '',
    loadChildren: () =>
      import('./layouts/main-layout/main-layout.routes').then(
        (mod) => mod.PAGES_ROUTES
      ),
    canActivate: [() => inject(Authservice).isLoggedIn],
  },
  // redirect empty path
  {
    path: '',
    pathMatch: 'full',
    redirectTo: 'properties',
  },
  // catch-all wildcard
  {
    path: '**',
    redirectTo: 'properties',
  },
];
