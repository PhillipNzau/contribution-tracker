import { Component, EventEmitter, Input, Output } from '@angular/core';

@Component({
  selector: 'app-modal',
  standalone: true,
  templateUrl: './modal.html',
  styleUrl: './modal.css',
})
export class Modal {
  @Input() showModal = false;
  @Input() width: string = 'auto';
  @Input() borderColor: string = 'auto';
  @Input() justifyTop: string = 'top';

  @Output() backdropClick = new EventEmitter<void>();

  onBackdropClick(event: MouseEvent): void {
    if (event.target === event.currentTarget) {
      this.backdropClick.emit();
    }
  }
}
