import {
  Component,
  EventEmitter,
  inject,
  Input,
  Output,
  signal,
  Signal,
} from '@angular/core';
import { Router, RouterModule } from '@angular/router';
import { HotToastService } from '@ngneat/hot-toast';
import {
  AbstractControl,
  AsyncValidatorFn,
  FormBuilder,
  FormGroup,
  ReactiveFormsModule,
  ValidationErrors,
  ValidatorFn,
  Validators,
} from '@angular/forms';
import { CommonModule } from '@angular/common';
import { Authservice } from '../../../shared/services/authservice';

@Component({
  selector: 'app-login-form',
  imports: [CommonModule, ReactiveFormsModule, RouterModule],
  templateUrl: './login-form.html',
  styleUrl: './login-form.css',
})
export class LoginForm {
  @Input({ required: true }) authType!: Signal<string>;
  @Output() back = new EventEmitter<boolean>();

  handleBack() {
    // Emit a back
    this.back.emit(false);
  }

  fb = inject(FormBuilder);
  route = inject(Router);
  authService = inject(Authservice);
  toastService = inject(HotToastService);

  loginForm = this.fb.nonNullable.group({
    email: ['', [Validators.required, Validators.email]],
  });

  signUpForm = this.fb.nonNullable.group({
    name: ['', [Validators.required]],
    phone: ['', [Validators.required]],
    email: ['', [Validators.required, Validators.email]],
    role: ['host'],
  });

  verifyOtpForm = this.fb.nonNullable.group({
    email: ['', [Validators.required]],
    otp: ['', [Validators.required, Validators.min(6)]],
  });

  isSignup = signal<boolean>(false);
  isVerifyOtp = signal<boolean>(false);
  isSubmitting = signal<boolean>(false);

  email = signal<any>('');

  toggleSignup() {
    if (this.isVerifyOtp()) {
      this.isSignup.set(false);
    } else {
      this.isSignup.set(!this.isSignup());
    }
    this.isVerifyOtp.set(false);
  }

  // login user
  submitLoginForm() {
    const loadingToast = this.toastService.loading('Processing...');
    this.isSubmitting.set(true);

    this.authService.loginUser(this.loginForm.value).subscribe({
      next: (res) => {
        this.toastService.success(`OTP set successfully!`, {
          duration: 2000,
        });
        loadingToast.close();
        this.isSubmitting.set(false);
        this.isVerifyOtp.set(true);
        this.email.set(this.loginForm.value.email);
        this.verifyOtpForm.patchValue({
          email: this.loginForm.value.email,
        });
      },
      error: (err) => {
        this.toastService.error(
          `Something went wrong logging in! ${err.error.error}!!`,
          {
            duration: 2000,
          }
        );
        loadingToast.close();
        this.isSubmitting.set(false);
      },
    });
  }

  // register user
  submitSignupForm() {
    const loadingToast = this.toastService.loading('Processing...');
    this.isSubmitting.set(true);
    this.authService.registerUser(this.signUpForm.value).subscribe({
      next: (res) => {
        this.toastService.success(
          `User registration success, check email for otp!`,
          {
            duration: 2000,
          }
        );
        loadingToast.close();
        this.isSubmitting.set(false);
        this.isVerifyOtp.set(true);
        this.isSignup.set(false);
        this.verifyOtpForm.patchValue({
          email: this.signUpForm.value.email,
        });
      },
      error: (err) => {
        this.toastService.error(
          `Something went wrong registering! ${err.error.error}!`,
          {
            duration: 2000,
          }
        );
        loadingToast.close();
        this.isSubmitting.set(false);
      },
    });
  }

  // verify otp
  submitOtpForm() {
    const loadingToast = this.toastService.loading('Processing...');
    this.isSubmitting.set(true);
    this.authService.verifyOtp(this.verifyOtpForm.value).subscribe({
      next: (res) => {
        this.toastService.success(`User Verified success, Welcome!`, {
          duration: 2000,
        });
        this.route.navigate(['/properties']);
        loadingToast.close();
        this.isSubmitting.set(false);
      },
      error: (err) => {
        this.toastService.error(
          `Something went wrong verifying! ${err.error.error}!`,
          {
            duration: 2000,
          }
        );
        loadingToast.close();
        this.isSubmitting.set(false);
      },
    });
  }
}
