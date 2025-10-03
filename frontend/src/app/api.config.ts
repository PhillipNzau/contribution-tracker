import { InjectionToken } from '@angular/core';

// Define the shape of your API configuration object
export interface ApiConfig {
  baseUrl: string;
  endpoints: {
    loginUser: string;
    registerUser: string;
    refreshToken: string;
    verifyOtp: string;
    updateUser: string;

    // property
    propertyUrl: string;
  };
}

// Create the injection token
export const API_CONFIG = new InjectionToken<ApiConfig>('api.config');

// Define the actual configuration value
export const apiConfigValue: ApiConfig = {
  // baseUrl: 'http://localhost:8080',
  baseUrl: 'https://unit-wise.onrender.com',

  endpoints: {
    loginUser: '/auth/login',
    registerUser: '/auth/register',
    refreshToken: '/auth/refresh',
    verifyOtp: '/auth/verify-otp',
    updateUser: '/users',

    // property
    propertyUrl: '/properties',
  },
};
