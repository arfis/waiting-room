import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';

@Component({
  selector: 'app-check-in-entry',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './check-in-entry.component.html',
  styleUrls: ['./check-in-entry.component.scss']
})
export class CheckInEntryComponent {
  constructor(private router: Router) {}

  goCardReader() {
    this.router.navigate(['/card']);
  }

  goManual() {
    this.router.navigate(['/manual']);
  }
}
