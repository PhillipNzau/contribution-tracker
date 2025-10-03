import { Component, inject, signal } from '@angular/core';
import { ReactiveFormsModule, FormBuilder, Validators } from '@angular/forms';
import { Authservice } from '../../shared/services/authservice';
import { HotToastService } from '@ngneat/hot-toast';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';

@Component({
  selector: 'app-settings',
  imports: [ReactiveFormsModule, CommonModule],
  templateUrl: './settings.html',
  styleUrl: './settings.css',
})
export class Settings {
  fb = inject(FormBuilder);
  authService = inject(Authservice);
  toastService = inject(HotToastService);
  route = inject(Router);

  isSubmitting = signal<boolean>(false);
  isToggled = signal<boolean>(false);

  usr = JSON.parse(localStorage.getItem('uWfUsr') || '');
  userDetails = signal<any>(this.usr);

  updateProfileForm = this.fb.nonNullable.group({
    name: [this.userDetails().name, [Validators.required]],
    phone: [this.userDetails().phone, [Validators.required]],
  });

  toggleChanged(event: Event): void {
    const checkbox = event.target as HTMLInputElement;
    this.isToggled.set(checkbox.checked);
  }

  submitUpdateForm() {
    const loadingToast = this.toastService.loading('Processing...');
    this.isSubmitting.set(true);

    this.authService
      .updateUser(this.updateProfileForm.value, this.userDetails()!.id)
      .subscribe({
        next: (res) => {
          loadingToast.close();
          this.isSubmitting.set(false);
          this.getUser();
          this.isToggled.set(!this.isToggled);
        },
        error: (err) => {
          this.toastService.error(
            `Something went wrong updating! ${err.error.error}!!`,
            {
              duration: 2000,
            }
          );
          loadingToast.close();
          this.isSubmitting.set(false);
        },
      });
  }

  getUser() {
    this.authService.getUser(this.userDetails()!.id).subscribe({
      next: (res) => {},
      error: (err) => {
        this.toastService.error(
          `Something went wrong updating! ${err.error.error}!!`,
          {
            duration: 2000,
          }
        );
      },
    });
  }

  logOut() {
    localStorage.clear();
    this.route.navigate(['/auth']);
  }
}
