import { CommonModule } from '@angular/common';
import { Component, inject } from '@angular/core';
import { RouterModule, Router } from '@angular/router';

@Component({
  selector: 'app-navbar',
  imports: [CommonModule, RouterModule],
  templateUrl: './nav-bar.html',
  styleUrl: './nav-bar.css',
})
export class NavBar {
  router = inject(Router);
  isRouteActive(routePath: string): boolean {
    return this.router.url.includes(routePath);
  }

  vibrateOnClick() {
    navigator.vibrate(30);
  }
}
