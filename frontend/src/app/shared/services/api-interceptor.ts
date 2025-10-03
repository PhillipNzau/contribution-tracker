import {
  HttpInterceptorFn,
  HttpResponse,
  HttpErrorResponse,
  HttpRequest,
} from '@angular/common/http';
import { of, throwError, switchMap, catchError, tap, startWith } from 'rxjs';
import { HttpClient } from '@angular/common/http';
import { inject } from '@angular/core';
import { Authservice } from './authservice';

type CacheEntry = { etag?: string; lastModified?: string; body: any };
const etagCache = new Map<string, CacheEntry>();

// store refresh in localStorage as well
const ACCESS_TOKEN_KEY = 'uWfTk';
const REFRESH_TOKEN_KEY = 'uWfRTk';

function getToken(): string | null {
  return (
    localStorage.getItem(ACCESS_TOKEN_KEY)?.replace(/^"(.*)"$/, '$1') || null
  );
}

function getRefreshToken(): string | null {
  return (
    localStorage.getItem(REFRESH_TOKEN_KEY)?.replace(/^"(.*)"$/, '$1') || null
  );
}

function setTokens(access: string, refresh: string) {
  localStorage.setItem(ACCESS_TOKEN_KEY, access);
  localStorage.setItem(REFRESH_TOKEN_KEY, refresh);
}

export const apiInterceptor: HttpInterceptorFn = (req, next) => {
  const authService = inject(Authservice);

  // 🚨 Skip adding token / refresh logic for refresh endpoint
  if (req.url.includes('/auth/refresh') || req.url.includes('/auth/verify')) {
    return next(req);
  }

  const token = getToken();
  let authReq = req;

  if (token) {
    authReq = authReq.clone({
      setHeaders: { Authorization: `Bearer ${token}` },
    });
  }

  const cacheKey = authReq.urlWithParams;
  const cached = etagCache.get(cacheKey);

  if (cached?.etag || cached?.lastModified) {
    let headers: Record<string, string> = {};
    if (cached.etag) headers['If-None-Match'] = cached.etag;
    if (cached.lastModified) headers['If-Modified-Since'] = cached.lastModified;
    authReq = authReq.clone({ setHeaders: headers });
  }

  const network$ = next(authReq).pipe(
    tap((event) => {
      if (event instanceof HttpResponse && event.status === 200) {
        const etag = event.headers.get('ETag') || undefined;
        const lastModified = event.headers.get('Last-Modified') || undefined;
        etagCache.set(cacheKey, { etag, lastModified, body: event.body });
      }
    }),
    catchError((err: HttpErrorResponse) => {
      if (err.status === 304 && cached) {
        return of(
          new HttpResponse({
            body: cached.body,
            status: 200,
            statusText: 'OK (from cache)',
            url: authReq.url,
          })
        );
      }

      // 🚨 Handle 401 safely
      if (err.status === 401) {
        const refresh = getRefreshToken();
        if (!refresh) return throwError(() => err);

        return authService.refreshToken(refresh).pipe(
          switchMap((res) => {
            const newAccess = res.access_token;
            const newRefresh = res.refresh_token;

            if (newAccess && newRefresh) {
              setTokens(newAccess, newRefresh);

              // retry original request with new token
              const retryReq = req.clone({
                setHeaders: { Authorization: `Bearer ${newAccess}` },
              });
              return next(retryReq);
            }

            return throwError(() => err);
          }),
          catchError((refreshErr) => {
            console.error('Refresh failed', refreshErr);
            return throwError(() => refreshErr);
          })
        );
      }

      return throwError(() => err);
    })
  );

  if (cached) {
    return network$.pipe(
      startWith(
        new HttpResponse({
          body: cached.body,
          status: 200,
          statusText: 'OK (stale cache)',
          url: authReq.url,
        })
      )
    );
  }

  return network$;
};
