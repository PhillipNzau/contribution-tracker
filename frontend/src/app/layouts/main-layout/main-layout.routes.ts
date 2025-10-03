import { Route } from '@angular/router';
import { MainLayout } from './main-layout';

export const PAGES_ROUTES: Route[] = [
  {
    path: '',
    redirectTo: 'properties',
    pathMatch: 'full',
  },
  {
    path: '',
    component: MainLayout,
    children: [
      {
        path: 'properties',
        loadComponent: () =>
          import('../../pages/properties/properties').then((m) => m.Properties),
      },
      {
        path: 'properties/:id',
        loadComponent: () =>
          import(
            '../../pages/properties/components/property-details/property-details'
          ).then((m) => m.PropertyDetails),
      },
      {
        path: 'reports',
        loadComponent: () =>
          import('../../pages/reports/reports').then((m) => m.Reports),
      },
      {
        path: 'notifications',
        loadComponent: () =>
          import('../../pages/notifications/notifications').then(
            (m) => m.Notifications
          ),
      },
      {
        path: 'profile',
        loadComponent: () =>
          import('../../pages/settings/settings').then((m) => m.Settings),
      },
    ],
  },
];
