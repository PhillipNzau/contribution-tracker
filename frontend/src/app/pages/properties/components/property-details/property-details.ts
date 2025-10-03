import { CommonModule } from '@angular/common';
import { Component, inject, OnInit, signal } from '@angular/core';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { HotToastService } from '@ngneat/hot-toast';
import { PropertyService } from '../../../../shared/services/propertyService';
import { PropertyResponseModel } from '../../../../shared/models/properties-model';
import { FormBuilder, ReactiveFormsModule } from '@angular/forms';

@Component({
  selector: 'app-property-details',
  imports: [CommonModule, RouterModule, ReactiveFormsModule],
  templateUrl: './property-details.html',
  styleUrl: './property-details.css',
})
export class PropertyDetails implements OnInit {
  private propertyService = inject(PropertyService);
  private toastService = inject(HotToastService);
  private route = inject(ActivatedRoute);
  private fb = inject(FormBuilder);

  isToggled = signal<boolean>(false);

  property = signal<PropertyResponseModel>({});
  selectedImage = signal<string>('');

  usr = JSON.parse(localStorage.getItem('uWfUsr') || '');
  role = signal<string>('');

  isEditing = signal(false);
  propertyForm = this.fb.nonNullable.group({
    title: [this.property().title],
    location: [this.property().location],
    available: [this.property().available],
    description: [this.property().description],
  });

  ngOnInit(): void {
    this.getProperty();
    this.role.set(this.usr?.role);
  }
  getProperty() {
    const loadingToast = this.toastService.loading('Processing...');

    const id = this.route.snapshot.paramMap.get('id');
    if (!id) {
      this.toastService.error('No property ID found in route');
      return;
    }

    this.propertyService.getSingleProperty(id).subscribe({
      next: (res) => {
        loadingToast.close();
        this.property.set(res);
      },
      error: () => {
        loadingToast.close();
        this.toastService.error('Failed to fetch property details');
      },
    });
  }

  selectImage(image: string) {
    this.selectedImage.set(image);
  }

  toggleChanged(event: Event): void {
    const checkbox = event.target as HTMLInputElement;
    this.isToggled.set(checkbox.checked);

    // Reset form with latest property values when editing starts
    if (this.isToggled()) {
      this.propertyForm.patchValue(this.property());
    }
  }

  saveChanges() {
    const loadingToast = this.toastService.loading('Processing...');

    const id = this.route.snapshot.paramMap.get('id');
    if (!id) {
      this.toastService.error('No property ID found in route');
      return;
    }

    if (this.propertyForm.valid) {
      this.propertyService
        .updateProperty(this.propertyForm.value, id)
        .subscribe({
          next: (res) => {
            loadingToast.close();
            this.property.set(res);
          },
          error: () => {
            loadingToast.close();
            this.toastService.error('Failed to update property details');
          },
        });
      this.isToggled.set(false);
    }
  }
}
