import { Inject, Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { map } from 'rxjs/operators';

import { API_CONFIG, ApiConfig } from '../../api.config';
import {
  CreatePropertyModel,
  PropertyResponseModel,
  PropertyUpdateResponseModel,
  UpdatePropertyModel,
} from '../models/properties-model';

@Injectable({
  providedIn: 'root',
})
export class PropertyService {
  private apiConfig = inject(API_CONFIG);

  router = inject(Router);
  http = inject(HttpClient);

  // get all property
  getAllProperty() {
    const url = `${this.apiConfig.baseUrl}${this.apiConfig.endpoints.propertyUrl}`;
    return this.http.get<PropertyResponseModel[]>(url).pipe(
      map((res) => {
        // if (res.status === 200) {
        //   return res;
        // }
        return res;
      })
    );
  }

  // get single property
  getSingleProperty(propertyId: string) {
    const url = `${this.apiConfig.baseUrl}${this.apiConfig.endpoints.propertyUrl}/${propertyId}`;
    return this.http.get<PropertyResponseModel>(url).pipe(
      map((res) => {
        // if (res.status === 200) {
        //   return res;
        // }
        return res;
      })
    );
  }

  // create  property
  createProperty(propertyData: any) {
    const url = `${this.apiConfig.baseUrl}${this.apiConfig.endpoints.propertyUrl}`;
    return this.http.post<PropertyResponseModel>(url, propertyData).pipe(
      map((res) => {
        // if (res.status === 200) {
        //   return res;
        // }
        return res;
      })
    );
  }

  // update  property
  updateProperty(propertyData: UpdatePropertyModel, propertyId: string) {
    const url = `${this.apiConfig.baseUrl}${this.apiConfig.endpoints.propertyUrl}/${propertyId}`;
    return this.http.patch<any>(url, propertyData).pipe(
      map((res) => {
        // if (res.status === 200) {
        //   return res;
        // }
        return res?.property;
      })
    );
  }
}
