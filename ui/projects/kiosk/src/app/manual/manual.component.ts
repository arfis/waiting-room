import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-manual-id',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './manual.component.html',
  styleUrls: ['./manual.component.scss']
})
export class ManualIdComponent {
  nationalId: string = '';

  constructor(private router: Router) {}

  submit() {
    console.log('Submitted National ID:', this.nationalId);
    // TODO: wire this to backend service
  }

  back() {
    this.router.navigate(['/']);
  }
}
