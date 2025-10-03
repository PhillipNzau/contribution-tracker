import { CommonModule } from '@angular/common';
import { Component, inject, OnInit, signal } from '@angular/core';
import { PropertyCard } from '../../shared/components/property-card/property-card';
import { PropertyService } from '../../shared/services/propertyService';
import {
  FormArray,
  FormBuilder,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import { HotToastService } from '@ngneat/hot-toast';
import { PropertyResponseModel } from '../../shared/models/properties-model';
import { Modal } from '../../shared/components/modal/modal';

@Component({
  selector: 'app-properties',
  imports: [ReactiveFormsModule, CommonModule, PropertyCard, Modal],
  templateUrl: './properties.html',
  styleUrl: './properties.css',
})
export class Properties implements OnInit {
  fb = inject(FormBuilder);
  propertyService = inject(PropertyService);
  toastService = inject(HotToastService);

  usr = JSON.parse(localStorage.getItem('uWfUsr') || '');
  userDetails = signal<any>(this.usr);

  isLayout = signal<boolean>(true);
  properties = signal<PropertyResponseModel[]>([]);

  createPropertyForm = this.fb.nonNullable.group({
    title: ['', [Validators.required]],
    location: ['', [Validators.required]],
    price: ['', [Validators.required]],
    description: ['', [Validators.required]],
    images: this.fb.nonNullable.array<File>([], [Validators.required]),
  });

  ngOnInit(): void {
    this.getAllProperties();
  }
  toggleLayout() {
    this.isLayout.set(!this.isLayout());
  }

  getAllProperties() {
    const loadingToast = this.toastService.loading('Processing...');

    this.propertyService.getAllProperty().subscribe({
      next: (res) => {
        loadingToast.close();
        this.properties.set(res);
      },
      error: (err) => {
        loadingToast.close();

        this.toastService.error(
          `Something went wrong feting properties! ${err.error.error}!!`,
          {
            duration: 2000,
          }
        );
      },
    });
  }

  isModal = signal<boolean>(false);
  toggleAddModal() {
    this.isModal.set(!this.isModal());
  }

  // Getter for easy access in template
  get images(): FormArray<any> {
    return this.createPropertyForm.get('images') as FormArray<any>;
  }
  // Function to handle file input and convert to base64
  imagePreviews: string[] = [];

  onImageSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (!input.files) return;

    Array.from(input.files).forEach((file) => {
      // Add the File to the FormArray
      this.images.push(this.fb.control(file));

      // Create a preview URL for the file
      const previewUrl = URL.createObjectURL(file);
      this.imagePreviews.push(previewUrl);
    });

    input.value = '';
  }

  removeImage(index: number): void {
    this.images.removeAt(index);
    this.imagePreviews.splice(index, 1);
  }

  submitCreateForm() {
    const loadingToast = this.toastService.loading('Processing...');

    if (this.createPropertyForm.invalid) return;

    const formValue = this.createPropertyForm.value;
    const formData = new FormData();

    // Append other fields
    formData.append('title', formValue.title ?? '');
    formData.append('location', formValue.location ?? '');
    formData.append('price', formValue.price ?? '');
    formData.append('description', formValue.description ?? '');

    // Append image files, safely
    (formValue.images ?? []).forEach((file: File) => {
      if (file) {
        formData.append('images', file);
      }
    });

    this.propertyService.createProperty(formData).subscribe({
      next: (res) => {
        this.getAllProperties();

        this.isModal.set(false);

        loadingToast.close();

        // Reset form and preview
        this.createPropertyForm.reset();
        this.imagePreviews = [];
        this.images.clear();
      },
      error: (err) => console.error('Error:', err),
    });
  }
}
