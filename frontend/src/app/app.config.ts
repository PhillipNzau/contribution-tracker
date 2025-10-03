import {
  ApplicationConfig,
  provideBrowserGlobalErrorListeners,
  provideZonelessChangeDetection,
  isDevMode,
} from '@angular/core';
import { provideRouter } from '@angular/router';

import { routes } from './app.routes';
import { provideServiceWorker } from '@angular/service-worker';
import { provideHotToastConfig } from '@ngneat/hot-toast';
import { API_CONFIG, apiConfigValue } from './api.config';
import {
  provideHttpClient,
  withFetch,
  withInterceptors,
} from '@angular/common/http';

import { apiInterceptor } from './shared/services/api-interceptor';

export const appConfig: ApplicationConfig = {
  providers: [
    provideBrowserGlobalErrorListeners(),
    provideZonelessChangeDetection(),
    provideHttpClient(withFetch(), withInterceptors([apiInterceptor])),
    provideRouter(routes),
    {
      provide: API_CONFIG,
      useValue: apiConfigValue,
    },
    provideHotToastConfig({
      visibleToasts: 1,
      duration: 700,
      position: 'bottom-center',
    }),
    provideServiceWorker('ngsw-worker.js', {
      enabled: true,
      registrationStrategy: 'registerWhenStable:30000',
    }),
  ],
};
